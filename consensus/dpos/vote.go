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

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
)

type System struct {
	config *Config
	IDB
}

func NewSystem(state *state.StateDB, chainConfig *params.ChainConfig, config *Config) *System {
	return &System{
		config: config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:    config.AccountName,
				assetid: chainConfig.SysTokenID,
				state:   state,
			},
		},
	}
}

// RegCandidate  register a candidate
func (sys *System) RegCandidate(epcho uint64, candidate string, url string, stake *big.Int, height uint64) error {
	// url validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}

	// stake validity
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.CandidateMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, candidate min %v)", stake, new(big.Int).Mul(sys.config.CandidateMinQuantity, sys.config.unitStake()))
	}

	// name validity
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod != nil {
		return fmt.Errorf("invalid candidate %v(already exist)", candidate)
	}

	// db
	prod = &CandidateInfo{
		Name:          candidate,
		URL:           url,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
		Height:        height,
	}
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UpdateCandidate  update a candidate
func (sys *System) UpdateCandidate(epcho uint64, candidate string, url string, nstake *big.Int, height uint64) error {
	// url validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}

	// stake validity
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(nstake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", nstake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.CandidateMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, candidate min %v)", nstake, new(big.Int).Mul(sys.config.CandidateMinQuantity, sys.config.unitStake()))
	}

	// name validity
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}

	// db
	stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	if err := sys.Undelegate(candidate, stake); err != nil {
		return fmt.Errorf("undelegate %v failed(%v)", q, err)
	}

	q = new(big.Int).Sub(q, prod.Quantity)
	if len(url) > 0 {
		prod.URL = url
	}
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	prod.Height = height
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}

	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UnregCandidate  unregister a candidate
func (sys *System) UnregCandidate(epcho uint64, candidate string) error {
	// name validity
	var stake *big.Int
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalide candidate %v", candidate)
	}
	if prod.InBlackList {
		return fmt.Errorf("in backlist %v", candidate)
	}

	// db
	// TODO
	stake = new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	if err := sys.Undelegate(candidate, stake); err != nil {
		return fmt.Errorf("undelegate %v failed(%v)", stake, err)
	}
	if err := sys.DelCandidate(prod.Name); err != nil {
		return err
	}

	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity)
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// VoteCandidate vote a candidate
func (sys *System) VoteCandidate(epcho uint64, voter string, candidate string, stake *big.Int) error {
	// candidate validity
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidates %v", candidate)
	}
	// stake validity
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.VoterMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, voter min %v)", stake, new(big.Int).Mul(sys.config.VoterMinQuantity, sys.config.unitStake()))
	}

	// db
	voterInfo, err := sys.GetVoter(epcho, voter, candidate)
	if err != nil {
		return err
	}
	if voterInfo == nil {
		voterInfo = &VoterInfo{
			Epcho:     epcho,
			Name:      voter,
			Candidate: candidate,
		}
	}

	//db
	voterInfo.Quantity = new(big.Int).Add(voterInfo.Quantity, q)
	if err := sys.SetVoter(vote); err != nil {
		return err
	}
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// GetDelegatedByTime candidate delegated
func (sys *System) GetDelegatedByTime(candidate string, timestamp uint64) (*big.Int, *big.Int, uint64, error) {
	q, tq, c, err := sys.IDB.GetDelegatedByTime(candidate, timestamp)
	if err != nil {
		return big.NewInt(0), big.NewInt(0), 0, err
	}
	return new(big.Int).Mul(q, sys.config.unitStake()), new(big.Int).Mul(tq, sys.config.unitStake()), c, nil
}

// KickedCandidate kicked
func (sys *System) KickedCandidate(epcho uint64, candidate string) error {
	// name validity
	prod, err := sys.GetCandidate(candidate)
	if prod == nil || err != nil {
		return err
	}

	// db
	// TODO
	stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	if err := sys.Undelegate(sys.config.SystemName, stake); err != nil {
		return err
	}
	prod.TotalQuantity = big.NewInt(0)
	prod.Quantity = big.NewInt(0)
	prod.InBlackList = true
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}

	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity)
	return sys.SetState(state)
}

// ExitTakeOver system exit take over
func (sys *System) ExitTakeOver(epcho uint64) error {
	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}
	gstate.TakeOver = false
	return sys.SetState(gstate)
}

// func (sys *System) onblock(height uint64) error {
// 	gstate, err := sys.GetState(height)
// 	if err != nil {
// 		return err
// 	}
// 	ngstate := &globalState{
// 		Height:                           height + 1,
// 		ActivatedCandidateSchedule:       gstate.ActivatedCandidateSchedule,
// 		ActivatedCandidateScheduleUpdate: gstate.ActivatedCandidateScheduleUpdate,
// 		ActivatedTotalQuantity:           gstate.ActivatedTotalQuantity,
// 		TotalQuantity:                    new(big.Int).SetBytes(gstate.TotalQuantity.Bytes()),
// 		TakeOver:                         gstate.TakeOver,
// 	}
// 	sys.SetState(ngstate)
// 	return nil
// }

// func (sys *System) updateElectedCandidates(height uint64) error {
// 	gstate, err := sys.GetState(LastBlockHeight)
// 	if err != nil {
// 		return err
// 	}

// 	size, _ := sys.CandidatesSize()
// 	if gstate.TotalQuantity.Cmp(sys.config.ActivatedMinQuantity) < 0 || uint64(size) < sys.config.consensusSize() {
// 		activatedCandidateSchedule := []string{}
// 		activeTotalQuantity := big.NewInt(0)
// 		candidate, _ := sys.GetCandidate(sys.config.SystemName)
// 		for i := uint64(0); i < sys.config.CandidateScheduleSize; i++ {
// 			activatedCandidateSchedule = append(activatedCandidateSchedule, sys.config.SystemName)
// 			activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, candidate.TotalQuantity)
// 		}
// 		gstate.ActivatedCandidateSchedule = activatedCandidateSchedule
// 		gstate.ActivatedCandidateScheduleUpdate = timestamp
// 		gstate.ActivatedTotalQuantity = activeTotalQuantity
// 		return sys.SetState(gstate)
// 	}

// 	candidates, err := sys.Candidates()
// 	if err != nil {
// 		return err
// 	}
// 	activatedCandidateSchedule := []string{}
// 	activeTotalQuantity := big.NewInt(0)
// 	for _, candidate := range candidates {
// 		if candidate.InBlackList || strings.Compare(candidate.Name, sys.config.SystemName) == 0 {
// 			continue
// 		}
// 		activatedCandidateSchedule = append(activatedCandidateSchedule, candidate.Name)
// 		activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, candidate.TotalQuantity)
// 		if uint64(len(activatedCandidateSchedule)) == sys.config.CandidateScheduleSize {
// 			break
// 		}
// 	}

// 	seed := int64(timestamp)
// 	r := rand.New(rand.NewSource(seed))
// 	for i := len(activatedCandidateSchedule) - 1; i > 0; i-- {
// 		j := int(r.Int31n(int32(i + 1)))
// 		activatedCandidateSchedule[i], activatedCandidateSchedule[j] = activatedCandidateSchedule[j], activatedCandidateSchedule[i]
// 	}

// 	gstate.ActivatedCandidateSchedule = activatedCandidateSchedule
// 	gstate.ActivatedCandidateScheduleUpdate = timestamp
// 	gstate.ActivatedTotalQuantity = activeTotalQuantity
// 	return sys.SetState(gstate)
// }
