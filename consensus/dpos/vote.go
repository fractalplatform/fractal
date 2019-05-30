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
	"strings"

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// System dpos internal contract
type System struct {
	config          *Config
	internalActions []*types.InternalAction
	IDB
}

// NewSystem new object
func NewSystem(state *state.StateDB, config *Config) *System {
	return &System{
		config: config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:    config.AccountName,
				assetid: config.AssetID,
				state:   state,
			},
		},
	}
}

// RegCandidate  register a candidate
func (sys *System) RegCandidate(epcho uint64, candidate string, url string, stake *big.Int, number uint64) error {
	// url validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url (too long, max %v)", sys.config.MaxURLLen)
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

	// quantity validity
	quantity, err := sys.getAvailableQuantity(epcho, candidate)
	if err != nil {
		return err
	}

	sub := new(big.Int).Sub(quantity, q)
	if sub.Sign() == -1 {
		sub = big.NewInt(0)
	}
	if err := sys.SetAvailableQuantity(epcho, candidate, sub); err != nil {
		return err
	}

	// db
	prod = &CandidateInfo{
		Name:          candidate,
		URL:           url,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
		Number:        number,
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
func (sys *System) UpdateCandidate(epcho uint64, candidate string, url string, nstake *big.Int, number uint64) error {
	// url validity
	if uint64(len(url)) > sys.config.MaxURLLen {
		return fmt.Errorf("invalid url (too long, max %v)", sys.config.MaxURLLen)
	}

	// stake validity
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(nstake, sys.config.unitStake(), m)
	if m.Sign() != 0 {
		return fmt.Errorf("invalid stake %v(non divisibility, unit %v)", nstake, sys.config.unitStake())
	}
	// if q.Cmp(sys.config.CandidateMinQuantity) < 0 {
	// 	return fmt.Errorf("invalid stake %v(insufficient, candidate min %v)", nstake, new(big.Int).Mul(sys.config.CandidateMinQuantity, sys.config.unitStake()))
	// }

	// name validity
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}
	if prod.Type != Normal {
		return fmt.Errorf("not in normal %v", candidate)
	}

	// if q.Sign() != 0 && q.Cmp(prod.Quantity) == -1 {
	// 	return fmt.Errorf("not support reduce stake %v", candidate)
	// }

	// q = new(big.Int).Sub(q, prod.Quantity)
	// quantity validity
	if q.Sign() == 1 {
		quantity, err := sys.getAvailableQuantity(epcho, candidate)
		if err != nil {
			return err
		}
		sub := new(big.Int).Sub(quantity, q)
		if sub.Sign() == -1 {
			sub = big.NewInt(0)
		}
		if err := sys.SetAvailableQuantity(epcho, candidate, sub); err != nil {
			return err
		}
	}

	// db
	// stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	// action, err := sys.Undelegate(candidate, stake)
	// if err != nil {
	// 	return fmt.Errorf("undelegate %v failed(%v)", q, err)
	// }
	// if action != nil {
	// 	sys.internalActions = append(sys.internalActions, &types.InternalAction{
	// 		Action: action.NewRPCAction(0),
	// 	})
	// }

	prod.URL = url
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	prod.Number = number
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
func (sys *System) UnregCandidate(epcho uint64, candidate string, number uint64) error {
	// name validity
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}
	if prod.Type != Normal {
		return fmt.Errorf("not in normal %v", candidate)
	}

	// db
	prod.Type = Freeze
	prod.Number = number
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}

	// stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	// action, err := sys.Undelegate(candidate, stake)
	// if err != nil {
	// 	return fmt.Errorf("undelegate %v failed(%v)", stake, err)
	// }
	// sys.internalActions = append(sys.internalActions, &types.InternalAction{
	// 	Action: action.NewRPCAction(0),
	// })

	// voters, err := sys.GetVoters(epcho, prod.Name)
	// if err != nil {
	// 	return err
	// }
	// for _, voter := range voters {
	// 	if voterInfo, err := sys.GetVoter(epcho, voter, candidate); err != nil {
	// 		return err
	// 	} else if err := sys.DelVoter(voterInfo); err != nil {
	// 		return err
	// 	} else if quantity, err := sys.GetAvailableQuantity(epcho, voter); err != nil {
	// 		return err
	// 	} else if err := sys.SetAvailableQuantity(epcho, voter, new(big.Int).Add(quantity, voterInfo.Quantity)); err != nil {
	// 		return err
	// 	}
	// }
	// if err := sys.DelCandidate(prod.Name); err != nil {
	// 	return err
	// }

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

// RefundCandidate  refund a candidate
func (sys *System) RefundCandidate(epcho uint64, candidate string, number uint64) error {
	// name validity
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}
	if prod.Type != Freeze {
		return fmt.Errorf("not in freeze %v", candidate)
	}

	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}

	freeze := uint64(0)
	tepcho := gstate.PreEpcho
	for i := uint64(0); i < sys.config.FreezeEpchoSize+1; i++ {
		tstate, err := sys.GetState(tepcho)
		if err != nil {
			return err
		}
		if tstate == nil {
			break
		}
		if tstate.Number < prod.Number {
			break
		}
		freeze++
		tepcho = tstate.PreEpcho
	}
	if freeze < sys.config.FreezeEpchoSize {
		return fmt.Errorf("%v freeze period %v has not arrived %v", candidate, freeze, sys.config.FreezeEpchoSize)
	}

	// db
	stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	action, err := sys.Undelegate(candidate, stake)
	if err != nil {
		return fmt.Errorf("undelegate %v failed(%v)", stake, err)
	}
	sys.internalActions = append(sys.internalActions, &types.InternalAction{
		Action: action.NewRPCAction(0),
	})

	// voters, err := sys.GetVoters(epcho, prod.Name)
	// if err != nil {
	// 	return err
	// }
	// for _, voter := range voters {
	// 	if voterInfo, err := sys.GetVoter(epcho, voter, candidate); err != nil {
	// 		return err
	// 	} else if err := sys.DelVoter(voterInfo); err != nil {
	// 		return err
	// 	} else if quantity, err := sys.GetAvailableQuantity(epcho, voter); err != nil {
	// 		return err
	// 	} else if err := sys.SetAvailableQuantity(epcho, voter, new(big.Int).Add(quantity, voterInfo.Quantity)); err != nil {
	// 		return err
	// 	}
	// }
	if err := sys.DelCandidate(prod.Name); err != nil {
		return err
	}

	// gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity)
	// if err := sys.SetState(gstate); err != nil {
	// 	return err
	// }
	return nil
}

// VoteCandidate vote a candidate
func (sys *System) VoteCandidate(epcho uint64, voter string, candidate string, stake *big.Int, number uint64) error {
	// candidate validity
	prod, err := sys.GetCandidate(candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}
	if prod.Type != Normal {
		return fmt.Errorf("not in normal %v", candidate)
	}

	gstate, err := sys.GetState(epcho)
	if err != nil {
		return err
	}
	timestamp := sys.config.epochTimeStamp(epcho)
	if sys.config.epoch(sys.config.ReferenceTime) == gstate.PreEpcho {
		timestamp = sys.config.epochTimeStamp(gstate.PreEpcho)
	}
	bquantity, err := sys.GetBalanceByTime(candidate, timestamp)
	if err != nil {
		return err
	}
	if s := new(big.Int).Mul(sys.config.unitStake(), sys.config.CandidateAvailableMinQuantity); bquantity.Cmp(s) == -1 {
		return fmt.Errorf("invalid candidate %v,(insufficient available quantity %v < %v)", candidate, bquantity, s)
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

	// quantity validity
	quantity, err := sys.getAvailableQuantity(epcho, voter)
	if err != nil {
		return err
	}
	if sub := new(big.Int).Sub(quantity, q); sub.Sign() == -1 {
		return fmt.Errorf("invalid vote stake %v(insufficient) %v < %v", voter, new(big.Int).Mul(quantity, sys.config.unitStake()), new(big.Int).Mul(q, sys.config.unitStake()))
	} else if err := sys.SetAvailableQuantity(epcho, voter, sub); err != nil {
		return err
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
			Quantity:  big.NewInt(0),
		}
	}

	voterInfo.Number = number
	voterInfo.Quantity = new(big.Int).Add(voterInfo.Quantity, q)
	if err := sys.SetVoter(voterInfo); err != nil {
		return err
	}

	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}

	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	return nil
}

// KickedCandidate kicked
func (sys *System) KickedCandidate(epcho uint64, candidate string, number uint64) error {
	// name validity
	prod, err := sys.GetCandidate(candidate)
	if prod == nil || err != nil {
		return err
	}
	if prod.Type == Black {
		return nil
	}

	// db
	stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	action, err := sys.Undelegate(sys.config.SystemName, stake)
	if err != nil {
		return fmt.Errorf("undelegate %v failed(%v)", stake, err)
	}
	sys.internalActions = append(sys.internalActions, &types.InternalAction{
		Action: action.NewRPCAction(0),
	})

	// voters, err := sys.GetVoters(epcho, prod.Name)
	// if err != nil {
	// 	return err
	// }
	// for _, voter := range voters {
	// 	if voterInfo, err := sys.GetVoter(epcho, voter, candidate); err != nil {
	// 		return err
	// 	} else if err := sys.DelVoter(voterInfo); err != nil {
	// 		return err
	// 	} else if quantity, err := sys.GetAvailableQuantity(epcho, voter); err != nil {
	// 		return err
	// 	} else if err := sys.SetAvailableQuantity(epcho, voter, new(big.Int).Add(quantity, voterInfo.Quantity)); err != nil {
	// 		return err
	// 	}
	// }

	if !prod.invalid() {
		gstate, err := sys.GetState(epcho)
		if err != nil {
			return err
		}
		gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity)
		if err := sys.SetState(gstate); err != nil {
			return err
		}
	}

	prod.Number = number
	prod.Type = Black
	return sys.SetCandidate(prod)
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

func (sys *System) onblock(epcho uint64, number uint64) error {
	pepcho, err := sys.GetLastestEpcho()
	if err != nil {
		return err
	}
	if pepcho == epcho {
		return nil
	}

	if pepcho > epcho {
		panic(err)
	}

	pState, err := sys.GetState(pepcho)
	if err != nil {
		return err
	}
	gstate := &GlobalState{
		Epcho:                  epcho,
		PreEpcho:               pepcho,
		ActivatedTotalQuantity: big.NewInt(0),
		TotalQuantity:          new(big.Int).SetBytes(pState.TotalQuantity.Bytes()),
		TakeOver:               pState.TakeOver,
		Dpos:                   pState.Dpos,
		Number:                 number,
	}
	return sys.SetState(gstate)
}

// UpdateElectedCandidates update
func (sys *System) UpdateElectedCandidates(pepcho uint64, epcho uint64, number uint64, miner string) error {
	if pepcho > epcho {
		panic(fmt.Errorf("UpdateElectedCandidates unreached"))
	}
	pstate, err := sys.GetState(pepcho)
	if err != nil {
		return err
	}

	// not is first & no changes
	if pstate.Epcho != pstate.PreEpcho && pepcho == epcho {
		return nil
	}

	candidateInfoArray, err := sys.GetCandidates()
	if err != nil {
		return err
	}
	n := sys.config.BackupScheduleSize + sys.config.CandidateScheduleSize
	activatedCandidateSchedule := []string{}
	activatedTotalQuantity := big.NewInt(0)
	totalQuantity := big.NewInt(0)
	quantity := big.NewInt(0)
	cnt := uint64(0)
	ntotalQuantity := big.NewInt(0)
	candidates := []*CandidateInfo{}
	for _, candidateInfo := range candidateInfoArray {
		if err := sys.SetCandidateByEpcho(pepcho, candidateInfo); err != nil {
			return err
		}

		if !candidateInfo.invalid() {
			if pstate.Dpos {
				if candidateInfo.Quantity.Sign() == 0 || strings.Compare(candidateInfo.Name, sys.config.SystemName) == 0 {
					continue
				}
			} else if candidateInfo.Quantity.Sign() == 0 || strings.Compare(candidateInfo.Name, sys.config.SystemName) == 0 {
				candidates = append(candidates, candidateInfo)
				continue
			}
			if uint64(len(activatedCandidateSchedule)) < n {
				activatedCandidateSchedule = append(activatedCandidateSchedule, candidateInfo.Name)
				activatedTotalQuantity = new(big.Int).Add(activatedTotalQuantity, candidateInfo.TotalQuantity)
			}
			totalQuantity = new(big.Int).Add(totalQuantity, candidateInfo.TotalQuantity)
			quantity = new(big.Int).Add(quantity, candidateInfo.Quantity)
			cnt++
		}

		// clear vote quantity
		candidateInfo.TotalQuantity = candidateInfo.Quantity
		if !candidateInfo.invalid() {
			ntotalQuantity = new(big.Int).Add(ntotalQuantity, candidateInfo.TotalQuantity)
		}
		if err := sys.SetCandidate(candidateInfo); err != nil {
			return err
		}
	}

	if !pstate.Dpos && totalQuantity.Cmp(sys.config.ActivatedMinQuantity) >= 0 &&
		cnt >= n && cnt >= sys.config.ActivatedMinCandidate {
		pstate.Dpos = true
		candidateInfo, err := sys.GetCandidate(miner)
		if err != nil {
			return err
		}
		candidateInfo.Counter++
		if err := sys.SetCandidateByEpcho(pepcho, candidateInfo); err != nil {
			return err
		}
		if err := sys.SetCandidate(candidateInfo); err != nil {
			return err
		}
	}

	if !pstate.Dpos {
		activatedTotalQuantity = big.NewInt(0)
		activatedCandidateSchedule = []string{}
		for index, candidateInfo := range candidates {
			if uint64(index) >= n {
				break
			}
			activatedCandidateSchedule = append(activatedCandidateSchedule, candidateInfo.Name)
			activatedTotalQuantity = new(big.Int).Add(activatedTotalQuantity, candidateInfo.TotalQuantity)
		}
		if init := len(activatedCandidateSchedule); init > 0 {
			index := 0
			for uint64(len(activatedCandidateSchedule)) < sys.config.CandidateScheduleSize {
				activatedCandidateSchedule = append(activatedCandidateSchedule, activatedCandidateSchedule[index/init])
				index++
			}
		}
	}
	pstate.ActivatedCandidateSchedule = activatedCandidateSchedule
	pstate.ActivatedTotalQuantity = activatedTotalQuantity
	pstate.Number = number
	if err := sys.SetState(pstate); err != nil {
		return err
	}

	if pepcho != epcho {
		gstate := &GlobalState{
			Epcho:                  epcho,
			PreEpcho:               pstate.Epcho,
			ActivatedTotalQuantity: big.NewInt(0),
			TotalQuantity:          new(big.Int).SetBytes(ntotalQuantity.Bytes()),
			OffCandidateNumber:     []uint64{},
			OffCandidateSchedule:   []uint64{},
			TakeOver:               pstate.TakeOver,
			Dpos:                   pstate.Dpos,
		}
		if err := sys.SetLastestEpcho(epcho); err != nil {
			return err
		}
		return sys.SetState(gstate)
	}
	return nil
}

func (sys *System) getAvailableQuantity(epcho uint64, voter string) (*big.Int, error) {
	q, err := sys.GetAvailableQuantity(epcho, voter)
	if err != nil {
		return nil, err
	}
	if q == nil {
		timestamp := sys.config.epochTimeStamp(epcho)
		gstate, err := sys.GetState(epcho)
		if err != nil {
			return nil, err
		}
		if sys.config.epoch(sys.config.ReferenceTime) == gstate.PreEpcho {
			timestamp = sys.config.epochTimeStamp(gstate.PreEpcho)
		}
		bquantity, err := sys.GetBalanceByTime(voter, timestamp)
		if err != nil {
			return nil, err
		}
		m := new(big.Int)
		quantity, _ := new(big.Int).DivMod(bquantity, sys.config.unitStake(), m)
		q = quantity
	}
	return q, nil
}
