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
)

type System struct {
	config *Config
	IDB
}

// RegCadidate  register a cadidate
func (sys *System) RegCadidate(cadidate string, url string, stake *big.Int) error {
	// parameter validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.CadidateMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, cadidate min %v)", stake, new(big.Int).Mul(sys.config.CadidateMinQuantity, sys.config.unitStake()))
	}

	if voter, err := sys.GetVoter(cadidate); err != nil {
		return err
	} else if voter != nil {
		return fmt.Errorf("invalid cadidate %v(alreay vote to %v)", cadidate, voter.Cadidate)
	}
	prod, err := sys.GetCadidate(cadidate)
	if err != nil {
		return err
	}
	if prod != nil {
		return fmt.Errorf("invalid cadidate %v(already exist)", cadidate)
	}
	prod = &cadidateInfo{
		Name:          cadidate,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
	}
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)

	if err := sys.Delegate(cadidate, stake); err != nil {
		return fmt.Errorf("delegate (%v) failed(%v)", stake, err)
	}

	prod.URL = url
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	prod.Height = gstate.Height
	if err := sys.SetCadidate(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UpdateCadidate  update a cadidate
func (sys *System) UpdateCadidate(cadidate string, url string, stake *big.Int) error {
	// parameter validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.CadidateMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, cadidate min %v)", stake, new(big.Int).Mul(sys.config.CadidateMinQuantity, sys.config.unitStake()))
	}

	prod, err := sys.GetCadidate(cadidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid cadidate %v(not exist)", cadidate)
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
		if err := sys.Undelegate(cadidate, tstake); err != nil {
			return fmt.Errorf("undelegate %v failed(%v)", q, err)
		}
	} else {
		if err := sys.Delegate(cadidate, tstake); err != nil {
			return fmt.Errorf("delegate (%v) failed(%v)", q, err)
		}
	}

	if len(url) > 0 {
		prod.URL = url
	}
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	if err := sys.SetCadidate(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UnregCadidate  unregister a cadidate
func (sys *System) UnregCadidate(cadidate string) error {
	// parameter validity
	// modify or update
	prod, err := sys.GetCadidate(cadidate)
	if err != nil {
		return err
	}
	if prod != nil {

		gstate, err := sys.GetState(LastBlockHeight)
		if err != nil {
			return err
		}
		if sys.isdpos(gstate) {
			if cnt, err := sys.CadidatesSize(); err != nil {
				return err
			} else if uint64(cnt) <= sys.config.consensusSize() {
				return fmt.Errorf("insufficient actived cadidates")
			}
			if new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity).Cmp(sys.config.ActivatedMinQuantity) < 0 {
				return fmt.Errorf("insufficient actived stake")
			}
		}

		// voters, err := sys.GetDelegators(cadidate)
		// if err != nil {
		// 	return err
		// }
		// for _, voter := range voters {
		// 	if err := sys.unvoteCadidate(voter); err != nil {
		// 		return err
		// 	}
		// }

		if prod.TotalQuantity.Cmp(prod.Quantity) > 0 {
			return fmt.Errorf("already has voter")
		}

		stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
		if err := sys.Undelegate(cadidate, stake); err != nil {
			return fmt.Errorf("undelegate %v failed(%v)", stake, err)
		}
		if err := sys.DelCadidate(prod.Name); err != nil {
			return err
		}
		gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.Quantity)
		if err := sys.SetState(gstate); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalide cadidate %v", cadidate)
	}
	return nil
}

// VoteCadidate vote a cadidate
func (sys *System) VoteCadidate(voter string, cadidate string, stake *big.Int) error {
	// parameter validity
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.VoterMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, voter min %v)", stake, new(big.Int).Mul(sys.config.VoterMinQuantity, sys.config.unitStake()))
	}

	if prod, err := sys.GetCadidate(voter); err != nil {
		return err
	} else if prod != nil {
		return fmt.Errorf("invalid vote(alreay is cadidate)")
	}
	if vote, err := sys.GetVoter(voter); err != nil {
		return err
	} else if vote != nil {
		return fmt.Errorf("invalid vote(already voted to cadidate %v)", vote.Cadidate)
	}
	prod, err := sys.GetCadidate(cadidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid vote(invalid cadidates %v)", cadidate)
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
		Name:     voter,
		Cadidate: cadidate,
		Quantity: q,
		Height:   gstate.Height,
	}
	if err := sys.SetVoter(vote); err != nil {
		return err
	}
	if err := sys.SetCadidate(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// ChangeCadidate change a cadidate
func (sys *System) ChangeCadidate(voter string, cadidate string) error {
	// parameter validity
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("invalid voter %v", voter)
	}
	if strings.Compare(vote.Cadidate, cadidate) == 0 {
		return fmt.Errorf("same cadidate")
	}
	prod, err := sys.GetCadidate(cadidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid cadidate %v", cadidate)
	}

	// modify or update
	oprod, err := sys.GetCadidate(vote.Cadidate)
	if err != nil {
		return err
	}
	oprod.TotalQuantity = new(big.Int).Sub(oprod.TotalQuantity, vote.Quantity)
	if err := sys.SetCadidate(oprod); err != nil {
		return err
	}
	if err := sys.DelVoter(vote.Name, vote.Cadidate); err != nil {
		return err
	}

	vote.Cadidate = prod.Name
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, vote.Quantity)
	if err := sys.SetVoter(vote); err != nil {
		return err
	}
	if err := sys.SetCadidate(prod); err != nil {
		return err
	}
	return nil
}

// UnvoteCadidate cancel vote
func (sys *System) UnvoteCadidate(voter string) error {
	// parameter validity
	return sys.unvoteCadidate(voter)
}

// UnvoteVoter cancel voter
func (sys *System) UnvoteVoter(cadidate string, voter string) error {
	// parameter validity
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("invalid voter %v", voter)
	}
	if strings.Compare(cadidate, vote.Cadidate) != 0 {
		return fmt.Errorf("invalid cadidate %v", cadidate)
	}
	return sys.unvoteCadidate(voter)
}

func (sys *System) GetDelegatedByTime(name string, timestamp uint64) (*big.Int, *big.Int, uint64, error) {
	q, tq, c, err := sys.IDB.GetDelegatedByTime(name, timestamp)
	if err != nil {
		return big.NewInt(0), big.NewInt(0), 0, err
	}
	return new(big.Int).Mul(q, sys.config.unitStake()), new(big.Int).Mul(tq, sys.config.unitStake()), c, nil
}

func (sys *System) KickedCadidate(name string, cadidates []string, invalid bool) error {
	if strings.Compare(name, sys.config.SystemName) == 0 {
		for _, cadidate := range cadidates {
			if prod, _ := sys.GetCadidate(cadidate); prod != nil {
				prod.Invalid = invalid
				sys.SetCadidate(prod)
			}
		}
	}
	return nil
}

func (sys *System) unvoteCadidate(voter string) error {
	// modify or update
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("invalid voter %v", voter)
	}
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}
	if sys.isdpos(gstate) && new(big.Int).Sub(gstate.TotalQuantity, vote.Quantity).Cmp(sys.config.ActivatedMinQuantity) < 0 {
		return fmt.Errorf("insufficient actived stake")
	}
	stake := new(big.Int).Mul(vote.Quantity, sys.config.unitStake())
	if err := sys.Undelegate(voter, stake); err != nil {
		return fmt.Errorf("undelegate %v failed(%v)", stake, err)
	}
	gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, vote.Quantity)
	prod, err := sys.GetCadidate(vote.Cadidate)
	if err != nil {
		return err
	}
	prod.TotalQuantity = new(big.Int).Sub(prod.TotalQuantity, vote.Quantity)
	if err := sys.SetCadidate(prod); err != nil {
		return err
	}
	if err := sys.DelVoter(vote.Name, vote.Cadidate); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

func (sys *System) onblock(height uint64) error {
	gstate, err := sys.GetState(height)
	if err != nil {
		return err
	}
	ngstate := &globalState{
		Height:                          height + 1,
		ActivatedCadidateSchedule:       gstate.ActivatedCadidateSchedule,
		ActivatedCadidateScheduleUpdate: gstate.ActivatedCadidateScheduleUpdate,
		ActivatedTotalQuantity:          gstate.ActivatedTotalQuantity,
		TotalQuantity:                   new(big.Int).SetBytes(gstate.TotalQuantity.Bytes()),
	}
	sys.SetState(ngstate)
	return nil
}

func (sys *System) updateElectedCadidates(timestamp uint64) error {
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}

	size, _ := sys.CadidatesSize()
	if gstate.TotalQuantity.Cmp(sys.config.ActivatedMinQuantity) < 0 || uint64(size) < sys.config.consensusSize() {
		activatedCadidateSchedule := []string{}
		activeTotalQuantity := big.NewInt(0)
		cadidate, _ := sys.GetCadidate(sys.config.SystemName)
		for i := uint64(0); i < sys.config.CadidateScheduleSize; i++ {
			activatedCadidateSchedule = append(activatedCadidateSchedule, sys.config.SystemName)
			activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, cadidate.TotalQuantity)
		}
		gstate.ActivatedCadidateSchedule = activatedCadidateSchedule
		gstate.ActivatedCadidateScheduleUpdate = timestamp
		gstate.ActivatedTotalQuantity = activeTotalQuantity
		return sys.SetState(gstate)
	}

	cadidates, err := sys.Cadidates()
	if err != nil {
		return err
	}
	activatedCadidateSchedule := []string{}
	activeTotalQuantity := big.NewInt(0)
	for _, cadidate := range cadidates {
		if cadidate.Invalid {
			continue
		}
		activatedCadidateSchedule = append(activatedCadidateSchedule, cadidate.Name)
		activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, cadidate.TotalQuantity)
		if uint64(len(activatedCadidateSchedule)) == sys.config.CadidateScheduleSize {
			break
		}
	}

	for uint64(len(activatedCadidateSchedule)) != sys.config.consensusSize() {
		activatedCadidateSchedule = append(activatedCadidateSchedule, sys.config.SystemName)
	}

	seed := int64(timestamp)
	r := rand.New(rand.NewSource(seed))
	for i := len(activatedCadidateSchedule) - 1; i > 0; i-- {
		j := int(r.Int31n(int32(i + 1)))
		activatedCadidateSchedule[i], activatedCadidateSchedule[j] = activatedCadidateSchedule[j], activatedCadidateSchedule[i]
	}

	gstate.ActivatedCadidateSchedule = activatedCadidateSchedule
	gstate.ActivatedCadidateScheduleUpdate = timestamp
	gstate.ActivatedTotalQuantity = activeTotalQuantity
	return sys.SetState(gstate)
}

func (sys *System) isdpos(gstate *globalState) bool {
	for _, cadidate := range gstate.ActivatedCadidateSchedule {
		if strings.Compare(cadidate, sys.config.SystemName) != 0 {
			return true
		}
	}
	return false
}
