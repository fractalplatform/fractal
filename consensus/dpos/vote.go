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

// RegProducer  register a producer
func (sys *System) RegProducer(producer string, url string, stake *big.Int) error {
	// parameter validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.ProducerMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, producer min %v)", stake, new(big.Int).Mul(sys.config.ProducerMinQuantity, sys.config.unitStake()))
	}

	if voter, err := sys.GetVoter(producer); err != nil {
		return err
	} else if voter != nil {
		return fmt.Errorf("invalid producer %v(alreay vote to %v)", producer, voter.Producer)
	}
	prod, err := sys.GetProducer(producer)
	if err != nil {
		return err
	}
	if prod != nil {
		return fmt.Errorf("invalid producer %v(already exist)", producer)
	}
	prod = &producerInfo{
		Name:          producer,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
	}
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)

	if err := sys.Delegate(producer, stake); err != nil {
		return fmt.Errorf("delegate (%v) failed(%v)", stake, err)
	}

	prod.URL = url
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	prod.Height = gstate.Height
	if err := sys.SetProducer(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UpdateProducer  update a producer
func (sys *System) UpdateProducer(producer string, url string, stake *big.Int) error {
	// parameter validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url %v(too long, max %v)", url, sys.config.MaxURLLen)
	}
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.ProducerMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, producer min %v)", stake, new(big.Int).Mul(sys.config.ProducerMinQuantity, sys.config.unitStake()))
	}

	prod, err := sys.GetProducer(producer)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid producer %v(not exist)", producer)
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
		if err := sys.Undelegate(producer, tstake); err != nil {
			return fmt.Errorf("undelegate %v failed(%v)", q, err)
		}
	} else {
		if err := sys.Delegate(producer, tstake); err != nil {
			return fmt.Errorf("delegate (%v) failed(%v)", q, err)
		}
	}

	if len(url) > 0 {
		prod.URL = url
	}
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	if err := sys.SetProducer(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// UnregProducer  unregister a producer
func (sys *System) UnregProducer(producer string) error {
	// parameter validity
	// modify or update
	prod, err := sys.GetProducer(producer)
	if err != nil {
		return err
	}
	if prod != nil {

		gstate, err := sys.GetState(LastBlockHeight)
		if err != nil {
			return err
		}
		if sys.isdpos(gstate) {
			if cnt, err := sys.ProducersSize(); err != nil {
				return err
			} else if uint64(cnt) <= sys.config.consensusSize() {
				return fmt.Errorf("insufficient actived producers")
			}
			if new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity).Cmp(sys.config.ActivatedMinQuantity) < 0 {
				return fmt.Errorf("insufficient actived stake")
			}
		}

		// voters, err := sys.GetDelegators(producer)
		// if err != nil {
		// 	return err
		// }
		// for _, voter := range voters {
		// 	if err := sys.unvoteProducer(voter); err != nil {
		// 		return err
		// 	}
		// }

		if prod.TotalQuantity.Cmp(prod.Quantity) > 0 {
			return fmt.Errorf("already has voter")
		}

		stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
		if err := sys.Undelegate(producer, stake); err != nil {
			return fmt.Errorf("undelegate %v failed(%v)", stake, err)
		}
		if err := sys.DelProducer(prod.Name); err != nil {
			return err
		}
		gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.Quantity)
		if err := sys.SetState(gstate); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalide producer %v", producer)
	}
	return nil
}

// VoteProducer vote a producer
func (sys *System) VoteProducer(voter string, producer string, stake *big.Int) error {
	// parameter validity
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", stake, sys.config.unitStake())
	}
	if q.Cmp(sys.config.VoterMinQuantity) < 0 {
		return fmt.Errorf("invalid stake %v(insufficient, voter min %v)", stake, new(big.Int).Mul(sys.config.VoterMinQuantity, sys.config.unitStake()))
	}

	if prod, err := sys.GetProducer(voter); err != nil {
		return err
	} else if prod != nil {
		return fmt.Errorf("invalid vote(alreay is producer)")
	}
	if vote, err := sys.GetVoter(voter); err != nil {
		return err
	} else if vote != nil {
		return fmt.Errorf("invalid vote(already voted to producer %v)", vote.Producer)
	}
	prod, err := sys.GetProducer(producer)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid vote(invalid producers %v)", producer)
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
		Producer: producer,
		Quantity: q,
		Height:   gstate.Height,
	}
	if err := sys.SetVoter(vote); err != nil {
		return err
	}
	if err := sys.SetProducer(prod); err != nil {
		return err
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// ChangeProducer change a producer
func (sys *System) ChangeProducer(voter string, producer string) error {
	// parameter validity
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("invalid voter %v", voter)
	}
	if strings.Compare(vote.Producer, producer) == 0 {
		return fmt.Errorf("same producer")
	}
	prod, err := sys.GetProducer(producer)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid producer %v", producer)
	}

	// modify or update
	oprod, err := sys.GetProducer(vote.Producer)
	if err != nil {
		return err
	}
	oprod.TotalQuantity = new(big.Int).Sub(oprod.TotalQuantity, vote.Quantity)
	if err := sys.SetProducer(oprod); err != nil {
		return err
	}
	if err := sys.DelVoter(vote.Name, vote.Producer); err != nil {
		return err
	}

	vote.Producer = prod.Name
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, vote.Quantity)
	if err := sys.SetVoter(vote); err != nil {
		return err
	}
	if err := sys.SetProducer(prod); err != nil {
		return err
	}
	return nil
}

// UnvoteProducer cancel vote
func (sys *System) UnvoteProducer(voter string) error {
	// parameter validity
	return sys.unvoteProducer(voter)
}

// UnvoteVoter cancel voter
func (sys *System) UnvoteVoter(producer string, voter string) error {
	// parameter validity
	vote, err := sys.GetVoter(voter)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("invalid voter %v", voter)
	}
	if strings.Compare(producer, vote.Producer) != 0 {
		return fmt.Errorf("invalid producer %v", producer)
	}
	return sys.unvoteProducer(voter)
}

func (sys *System) GetDelegatedByTime(name string, timestamp uint64) (*big.Int, error) {
	q, err := sys.IDB.GetDelegatedByTime(name, timestamp)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Mul(q, sys.config.unitStake()), nil
}

func (sys *System) unvoteProducer(voter string) error {
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
	prod, err := sys.GetProducer(vote.Producer)
	if err != nil {
		return err
	}
	prod.TotalQuantity = new(big.Int).Sub(prod.TotalQuantity, vote.Quantity)
	if err := sys.SetProducer(prod); err != nil {
		return err
	}
	if err := sys.DelVoter(vote.Name, vote.Producer); err != nil {
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
		ActivatedProducerSchedule:       gstate.ActivatedProducerSchedule,
		ActivatedProducerScheduleUpdate: gstate.ActivatedProducerScheduleUpdate,
		ActivatedTotalQuantity:          gstate.ActivatedTotalQuantity,
		TotalQuantity:                   new(big.Int).SetBytes(gstate.TotalQuantity.Bytes()),
	}
	sys.SetState(ngstate)
	return nil
}

func (sys *System) updateElectedProducers(timestamp uint64) error {
	gstate, err := sys.GetState(LastBlockHeight)
	if err != nil {
		return err
	}

	size, _ := sys.ProducersSize()
	if gstate.TotalQuantity.Cmp(sys.config.ActivatedMinQuantity) < 0 || uint64(size) < sys.config.consensusSize() {
		activatedProducerSchedule := []string{}
		activeTotalQuantity := big.NewInt(0)
		producer, _ := sys.GetProducer(sys.config.SystemName)
		for i := uint64(0); i < sys.config.ProducerScheduleSize; i++ {
			activatedProducerSchedule = append(activatedProducerSchedule, sys.config.SystemName)
			activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, producer.TotalQuantity)
		}
		gstate.ActivatedProducerSchedule = activatedProducerSchedule
		gstate.ActivatedProducerScheduleUpdate = timestamp
		gstate.ActivatedTotalQuantity = activeTotalQuantity
		return sys.SetState(gstate)
	}

	producers, err := sys.Producers()
	if err != nil {
		return err
	}
	activatedProducerSchedule := []string{}
	activeTotalQuantity := big.NewInt(0)
	for index, producer := range producers {
		if uint64(index) >= sys.config.ProducerScheduleSize {
			break
		}
		activatedProducerSchedule = append(activatedProducerSchedule, producer.Name)
		activeTotalQuantity = new(big.Int).Add(activeTotalQuantity, producer.TotalQuantity)
	}

	seed := int64(timestamp)
	r := rand.New(rand.NewSource(seed))
	for i := len(activatedProducerSchedule) - 1; i > 0; i-- {
		j := int(r.Int31n(int32(i + 1)))
		activatedProducerSchedule[i], activatedProducerSchedule[j] = activatedProducerSchedule[j], activatedProducerSchedule[i]
	}

	gstate.ActivatedProducerSchedule = activatedProducerSchedule
	gstate.ActivatedProducerScheduleUpdate = timestamp
	gstate.ActivatedTotalQuantity = activeTotalQuantity
	return sys.SetState(gstate)
}

func (sys *System) isdpos(gstate *globalState) bool {
	for _, producer := range gstate.ActivatedProducerSchedule {
		if strings.Compare(producer, sys.config.SystemName) != 0 {
			return true
		}
	}
	return false
}
