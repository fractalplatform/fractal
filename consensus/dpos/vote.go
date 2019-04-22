// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dpos

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"

	"github.com/fractalplatform/fractal/state"
)

type System struct {
	config *Config
	IDB
}

func NewSystem(state *state.StateDB, config *Config) *System {
	return &System{
		config: config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:  config.AccountName,
				state: state,
			},
		},
	}
}

// RegCandidate  register a candidate
func (sys *System) RegCandidate(candidate string, url string, stake *big.Int) error {
	// parameter validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.CandidateMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, candidate min %v)", stake, new(big.Int).Mul(sys.config.CandidateMinQuantity, sys.config.unitStake()))
	}

	if voter, err := sys.GetVoter(candidate); err != nil {
		return err
	} else if voter != nil {
		return fmt.Errorf("invalid candidate %v(alreay vote to %v)", candidate, voter.Candidate)
	}
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod != nil {
		return fmt.Errorf("invalid candidate %v(already exist)", candidate)
	}
	prod = &candidateInfo{
		Name:          candidate,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
	}
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)

	if err := sys.Delegate(candidate, stake); err != nil {
		return fmt.Errorf("delegate (%v) failed(%v)", stake, err)
	}

	prod.URL = url
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	prod.Height = gstate.Height
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UpdateCandidate  update a candidate
func (sys *System) UpdateCandidate(candidate string, url string, stake *big.Int) error {
	// parameter validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.CandidateMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, candidate min %v)", stake, new(big.Int).Mul(sys.config.CandidateMinQuantity, sys.config.unitStake()))
	}

	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}

	q = new(big.Int).Sub(q, prod.Quantity)

	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}
	if sys.isdpos(gstate) && new(big.Int).Add(gstate.TotalQuantity, q).Cmp(sys.config.ActivatedMinQuantity) < 0 {
		return fmt.Errorf("insufficient actived stake")
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)

	tstake := new(big.Int).Mul(q, sys.config.unitStake())
	if q.Sign() < 0 {
		tstake = new(big.Int).Abs(tstake)
		if err := sys.Undelegate(candidate, tstake); err != nil {
			return fmt.Errorf("undelegate %v failed(%v)", q, err)
		}
	} else {
		if err := sys.Delegate(candidate, tstake); err != nil {
			return fmt.Errorf("delegate (%v) failed(%v)", q, err)
		}
	}

	if len(url) > 0 {
		prod.URL = url
	}
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UnregCandidate  unregister a candidate
func (sys *System) UnregCandidate(candidate string) (*big.Int, error) {
	// parameter validity
	// modify or update
	var stake *big.Int
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return nil, err
	}
	if prod != nil {

		gstate, err := sys.GetState(LastBlockHeight)
		if err != nil {
			return nil, err
		}
		if sys.isdpos(gstate) {
			if cnt, err := sys.CandidatesSize(); err != nil {
				return nil, err
			} else if uint64(cnt) <= sys.config.consensusSize() {
				return nil, fmt.Errorf("insufficient actived candidates")
			}
			if new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity).Cmp(sys.config.ActivatedMinQuantity) < 0 {
				return nil, fmt.Errorf("insufficient actived stake")
			}
		}

		// voters, err := sys.GetDelegators(candidate)
		// if err != nil {
		// 	return err
		// }
		// for _, voter := range voters {
		// 	if err := sys.unvoteCandidate(voter); err != nil {
		// 		return err
		// 	}
		// }

		if prod.TotalQuantity.Cmp(prod.Quantity) > 0 {
			return nil, fmt.Errorf("already has voter")
		}

		stake = new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
		if err := sys.Undelegate(candidate, stake); err != nil {
			return nil, fmt.Errorf("undelegate %v failed(%v)", stake, err)
		}
		if prod.InBlackList {
			if err := sys.SetCandidate(prod); err != nil {
				return nil, err
			}
		} else {
			if err := sys.DelCandidate(prod.Name); err != nil {
				return nil, err
			}
		}

		gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.Quantity)
		if err := sys.SetState(gstate); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalide candidate %v", candidate)
	}
	return stake, nil
}

// VoteCandidate vote a candidate
func (sys *System) VoteCandidate(voter string, candidate string, stake *big.Int) error {
	// parameter validity
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.VoterMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, voter min %v)", stake, new(big.Int).Mul(sys.config.VoterMinQuantity, sys.config.unitStake()))
	}

	if prod, err := sys.GetCandidate(voter); err != nil {
		return err
	} else if prod != nil {
		return fmt.Errorf("invalid vote(alreay is candidate)")
	}
	if vote, err := sys.GetVoter(voter); err != nil {
		return err
	} else if vote != nil {
		return fmt.Errorf("invalid vote(already voted to candidate %v)", vote.Candidate)
	}
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid vote(invalid candidates %v)", candidate)
	}

	// modify or update
	if err := sys.Delegate(voter, stake); err != nil {
		return fmt.Errorf("delegate %v failed(%v)", stake, err)
	}

	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	vote := &voterInfo{
		Name:      voter,
		Candidate: candidate,
		Quantity:  q,
		Height:    gstate.Height,
	}
	if err := sys.SetVoter(vote); err != nil {
		return err
	}
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// ChangeCandidate change a candidate
func (sys *System) ChangeCandidate(voter string, candidate string) error {
	// parameter validity
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("invalid voter %v", voter)
	}
	if strings.Compare(vote.Candidate, candidate) == 0 {
		return fmt.Errorf("same candidate")
	}
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v", candidate)
	}

	// modify or update
	oprod, err := sys.GetCandidate(vote.Candidate)
	if err != nil {
		return err
	}
	oprod.TotalQuantity = new(big.Int).Sub(oprod.TotalQuantity, vote.Quantity)
	if err := sys.SetCandidate(oprod); err != nil {
		return err
	}
	if err := sys.DelVoter(vote.Name, vote.Candidate); err != nil {
		return err
	}

	vote.Candidate = prod.Name
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, vote.Quantity)
	if err := sys.SetVoter(vote); err != nil {
		return err
	}
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	return nil
}

// UnvoteCandidate cancel vote
func (sys *System) UnvoteCandidate(voter string) (*big.Int, error) {
	// parameter validity
	return sys.unvoteCandidate(voter)
}

// UnvoteVoter cancel voter
func (sys *System) UnvoteVoter(candidate string, voter string) (*big.Int, error) {
	// parameter validity
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return nil, err
	}
	if vote == nil {
		return nil, fmt.Errorf("invalid voter %v", voter)
	}
	if strings.Compare(candidate, vote.Candidate) != 0 {
		return nil, fmt.Errorf("invalid candidate %v", candidate)
	}
	return sys.unvoteCandidate(voter)
}

func (sys *System) GetDelegatedByTime(name string, timestamp uint64) (*big.Int, *big.Int, uint64, error) {
	q, tq, c, err := sys.IDB.GetDelegatedByTime(name, timestamp)
	if err != nil {
		return big.NewInt(0), big.NewInt(0), 0, err
	}
	return new(big.Int).Mul(q, sys.config.unitStake()), new(big.Int).Mul(tq, sys.config.unitStake()), c, nil
}

func (sys *System) KickedCandidate(candidate string) error {
	prod, err := sys.GetCandidate(candidate)
	if prod != nil {
		if err := sys.Undelegate(sys.config.SystemName, new(big.Int).Mul(prod.Quantity, sys.config.unitStake())); err != nil {
			return err
		}
		state, err := sys.GetState(LastBlockHeight)
		if err != nil {
			return err
		}
		state.TotalQuantity = new(big.Int).Sub(state.TotalQuantity, prod.Quantity)
		prod.TotalQuantity = new(big.Int).Sub(prod.TotalQuantity, prod.Quantity)
		prod.Quantity = big.NewInt(0)
		prod.InBlackList = true
		if err := sys.SetState(state); err != nil {
			return err
		}
		return sys.SetCandidate(prod)
	}
	return err
}

func (sys *System) ExitTakeOver() error {
	latest, err := sys.GetState(LastBlockHeight)
	if latest != nil {
		latest.TakeOver = false
		return sys.SetState(latest)
	}
	return err
}

func (sys *System) unvoteCandidate(voter string) (*big.Int, error) {
	// modify or update
	var stake *big.Int
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return nil, err
	}
	if vote == nil {
		return nil, fmt.Errorf("invalid voter %v", voter)
	}
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return nil, err
	}
	if sys.isdpos(gstate) && new(big.Int).Sub(gstate.TotalQuantity, vote.Quantity).Cmp(sys.config.ActivatedMinQuantity) < 0 {
		return nil, fmt.Errorf("insufficient actived stake")
	}
	stake = new(big.Int).Mul(vote.Quantity, sys.config.unitStake())
	if err := sys.Undelegate(voter, stake); err != nil {
		return nil, fmt.Errorf("undelegate %v failed(%v)", stake, err)
	}
	gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, vote.Quantity)
	prod, err := sys.GetCandidate(vote.Candidate)
	if err != nil {
		return nil, err
	}
	prod.TotalQuantity = new(big.Int).Sub(prod.TotalQuantity, vote.Quantity)
	if err := sys.SetCandidate(prod); err != nil {
		return nil, err
	}
	if err := sys.DelVoter(vote.Name, vote.Candidate); err != nil {
		return nil, err
	}
	if err := sys.SetState(gstate); err != nil {
		return nil, err
	}
	return stake, nil
}

func (sys *System) onblock(height uint64) error {
	gstate, err := sys.GetState(height)
	if err != nil {
		return err
	}
	ngstate := &globalState{
		Height:                           height + 1,
		ActivatedCandidateSchedule:       gstate.ActivatedCandidateSchedule,
		ActivatedCandidateScheduleUpdate: gstate.ActivatedCandidateScheduleUpdate,
		ActivatedTotalQuantity:           gstate.ActivatedTotalQuantity,
		TotalQuantity:                    new(big.Int).SetBytes(gstate.TotalQuantity.Bytes()),
		TakeOver:                         gstate.TakeOver,
	}
	sys.SetState(ngstate)
	return nil
}

func (sys *System) updateElectedCandidates(timestamp uint64) error {
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}

	size, _ := sys.CandidatesSize()
	if gstate.TotalQuantity.Cmp(sys.config.ActivatedMinQuantity) < 0 || uint64(size) < sys.config.consensusSize() {
		activatedCandidateSchedule := []string{}
		activeTotalQuantity := big.NewInt(0)
		candidate, _ := sys.GetCandidate(sys.config.SystemName)
		for i := uint64(0); i < sys.config.CandidateScheduleSize; i++ {
			activatedCandidateSchedule = append(activatedCandidateSchedule, sys.config.SystemName)
			activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, candidate.TotalQuantity)
		}
		gstate.ActivatedCandidateSchedule = activatedCandidateSchedule
		gstate.ActivatedCandidateScheduleUpdate = timestamp
		gstate.ActivatedTotalQuantity = activeTotalQuantity
		return sys.SetState(gstate)
	}

	candidates, err := sys.Candidates()
	if err != nil {
		return err
	}
	activatedCandidateSchedule := []string{}
	activeTotalQuantity := big.NewInt(0)
	for _, candidate := range candidates {
		if candidate.InBlackList || strings.Compare(candidate.Name, sys.config.SystemName) == 0 {
			continue
		}
		activatedCandidateSchedule = append(activatedCandidateSchedule, candidate.Name)
		activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, candidate.TotalQuantity)
		if uint64(len(activatedCandidateSchedule)) == sys.config.CandidateScheduleSize {
			break
		}
	}

	seed := int64(timestamp)
	r := rand.New(rand.NewSource(seed))
	for i := len(activatedCandidateSchedule) - 1; i > 0; i-- {
		j := int(r.Int31n(int32(i + 1)))
		activatedCandidateSchedule[i], activatedCandidateSchedule[j] = activatedCandidateSchedule[j], activatedCandidateSchedule[i]
	}

	gstate.ActivatedCandidateSchedule = activatedCandidateSchedule
	gstate.ActivatedCandidateScheduleUpdate = timestamp
	gstate.ActivatedTotalQuantity = activeTotalQuantity
	return sys.SetState(gstate)
}

func (sys *System) isdpos(gstate *globalState) bool {
	for _, candidate := range gstate.ActivatedCandidateSchedule {
		if strings.Compare(candidate, sys.config.SystemName) != 0 {
			return true
		}
	}
	return false
}
