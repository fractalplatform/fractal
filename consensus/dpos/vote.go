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
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
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
func (sys *System) RegCandidate(epoch uint64, candidate string, url string, stake *big.Int, number uint64, fid uint64) error {
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
	prod, err := sys.GetCandidate(epoch, candidate)
	if err != nil {
		return err
	}
	if prod != nil {
		return fmt.Errorf("invalid candidate %v(already exist)", candidate)
	}

	// quantity validity
	quantity, err := sys.getAvailableQuantity(epoch, candidate)
	if err != nil {
		return err
	}

	sub := new(big.Int).Sub(quantity, q)
	if sub.Sign() == -1 {
		sub = big.NewInt(0)
	}
	if err := sys.SetAvailableQuantity(epoch, candidate, sub); err != nil {
		return err
	}

	// db
	prod = &CandidateInfo{
		Epoch:         epoch,
		Name:          candidate,
		Info:          url,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
		Number:        number,
	}
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)

	gstate, err := sys.GetState(epoch)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	if fid >= params.ForkID2 {
		if err := sys.updateState(gstate, prod); err != nil {
			return err
		}
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}

	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	return nil
}

// UpdateCandidate  update a candidate
func (sys *System) UpdateCandidate(epoch uint64, candidate string, info string, nstake *big.Int, number uint64, fid uint64) error {
	// url validity
	if uint64(len(info)) > sys.config.MaxURLLen {
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
	prod, err := sys.GetCandidate(epoch, candidate)
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
		quantity, err := sys.getAvailableQuantity(epoch, candidate)
		if err != nil {
			return err
		}
		sub := new(big.Int).Sub(quantity, q)
		if sub.Sign() == -1 {
			sub = big.NewInt(0)
		}
		if err := sys.SetAvailableQuantity(epoch, candidate, sub); err != nil {
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

	prod.Info = info
	prod.Quantity = new(big.Int).Add(prod.Quantity, q)
	prod.TotalQuantity = new(big.Int).Add(prod.TotalQuantity, q)
	prod.Number = number

	gstate, err := sys.GetState(epoch)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	if fid >= params.ForkID2 {
		if err := sys.updateState(gstate, prod); err != nil {
			return err
		}
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	return nil
}

// UnregCandidate  unregister a candidate
func (sys *System) UnregCandidate(epoch uint64, candidate string, number uint64, fid uint64) error {
	// name validity
	prod, err := sys.GetCandidate(epoch, candidate)
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

	// stake := new(big.Int).Mul(prod.Quantity, sys.config.unitStake())
	// action, err := sys.Undelegate(candidate, stake)
	// if err != nil {
	// 	return fmt.Errorf("undelegate %v failed(%v)", stake, err)
	// }
	// sys.internalActions = append(sys.internalActions, &types.InternalAction{
	// 	Action: action.NewRPCAction(0),
	// })

	// voters, err := sys.GetVoters(epoch, prod.Name)
	// if err != nil {
	// 	return err
	// }
	// for _, voter := range voters {
	// 	if voterInfo, err := sys.GetVoter(epoch, voter, candidate); err != nil {
	// 		return err
	// 	} else if err := sys.DelVoter(voterInfo); err != nil {
	// 		return err
	// 	} else if quantity, err := sys.GetAvailableQuantity(epoch, voter); err != nil {
	// 		return err
	// 	} else if err := sys.SetAvailableQuantity(epoch, voter, new(big.Int).Add(quantity, voterInfo.Quantity)); err != nil {
	// 		return err
	// 	}
	// }
	// if err := sys.DelCandidate(prod.Name); err != nil {
	// 	return err
	// }

	gstate, err := sys.GetState(epoch)
	if err != nil {
		return err
	}
	gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity)
	if fid >= params.ForkID2 {
		if err := sys.updateState(gstate, prod); err != nil {
			return err
		}
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	return nil
}

// RefundCandidate  refund a candidate
func (sys *System) RefundCandidate(epoch uint64, candidate string, number uint64, fid uint64) error {
	// name validity
	prod, err := sys.GetCandidate(epoch, candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}
	if prod.Type != Freeze {
		return fmt.Errorf("not in freeze %v", candidate)
	}

	gstate, err := sys.GetState(epoch)
	if err != nil {
		return err
	}

	freeze := uint64(0)
	tepoch := gstate.PreEpoch
	for i := uint64(0); i < sys.config.FreezeEpochSize; i++ {
		tstate, err := sys.GetState(tepoch)
		if err != nil && strings.Compare(err.Error(), "epoch not found") != 0 {
			return err
		}
		if tstate == nil {
			break
		}
		if tstate.Number < prod.Number {
			break
		}
		freeze++
		if tstate.Epoch == tstate.PreEpoch {
			break
		}
		tepoch = tstate.PreEpoch
	}
	if freeze < sys.config.FreezeEpochSize {
		return fmt.Errorf("%v freeze period %v has not arrived %v", candidate, freeze, sys.config.FreezeEpochSize)
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

	// voters, err := sys.GetVoters(epoch, prod.Name)
	// if err != nil {
	// 	return err
	// }
	// for _, voter := range voters {
	// 	if voterInfo, err := sys.GetVoter(epoch, voter, candidate); err != nil {
	// 		return err
	// 	} else if err := sys.DelVoter(voterInfo); err != nil {
	// 		return err
	// 	} else if quantity, err := sys.GetAvailableQuantity(epoch, voter); err != nil {
	// 		return err
	// 	} else if err := sys.SetAvailableQuantity(epoch, voter, new(big.Int).Add(quantity, voterInfo.Quantity)); err != nil {
	// 		return err
	// 	}
	// }
	if err := sys.DelCandidate(epoch, prod.Name); err != nil {
		return err
	}

	// gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity)
	// if err := sys.SetState(gstate); err != nil {
	// 	return err
	// }
	return nil
}

// VoteCandidate vote a candidate
func (sys *System) VoteCandidate(epoch uint64, voter string, candidate string, stake *big.Int, number uint64, fid uint64) error {
	// candidate validity
	prod, err := sys.GetCandidate(epoch, candidate)
	if err != nil {
		return err
	}
	if prod == nil {
		return fmt.Errorf("invalid candidate %v(not exist)", candidate)
	}
	if prod.Type != Normal {
		return fmt.Errorf("not in normal %v", candidate)
	}

	gstate, err := sys.GetState(epoch)
	if err != nil {
		return err
	}
	timestamp := sys.config.epochTimeStamp(epoch)
	if sys.config.epoch(sys.config.ReferenceTime) == gstate.PreEpoch {
		timestamp = sys.config.epochTimeStamp(gstate.PreEpoch)
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
	quantity, err := sys.getAvailableQuantity(epoch, voter)
	if err != nil {
		return err
	}
	if sub := new(big.Int).Sub(quantity, q); sub.Sign() == -1 {
		return fmt.Errorf("invalid vote stake %v(insufficient) %v < %v", voter, new(big.Int).Mul(quantity, sys.config.unitStake()), new(big.Int).Mul(q, sys.config.unitStake()))
	} else if err := sys.SetAvailableQuantity(epoch, voter, sub); err != nil {
		return err
	}

	// db
	voterInfo, err := sys.GetVoter(epoch, voter, candidate)
	if err != nil {
		return err
	}
	if voterInfo == nil {
		voterInfo = &VoterInfo{
			Epoch:     epoch,
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

	gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, q)
	if fid >= params.ForkID2 {
		if err := sys.updateState(gstate, prod); err != nil {
			return err
		}
	}
	if err := sys.SetState(gstate); err != nil {
		return err
	}
	if err := sys.SetCandidate(prod); err != nil {
		return err
	}
	return nil
}

// KickedCandidate kicked
func (sys *System) KickedCandidate(epoch uint64, candidate string, number uint64, fid uint64) error {
	// name validity
	prod, err := sys.GetCandidate(epoch, candidate)
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

	// voters, err := sys.GetVoters(epoch, prod.Name)
	// if err != nil {
	// 	return err
	// }
	// for _, voter := range voters {
	// 	if voterInfo, err := sys.GetVoter(epoch, voter, candidate); err != nil {
	// 		return err
	// 	} else if err := sys.DelVoter(voterInfo); err != nil {
	// 		return err
	// 	} else if quantity, err := sys.GetAvailableQuantity(epoch, voter); err != nil {
	// 		return err
	// 	} else if err := sys.SetAvailableQuantity(epoch, voter, new(big.Int).Add(quantity, voterInfo.Quantity)); err != nil {
	// 		return err
	// 	}
	// }

	if !prod.invalid() {
		gstate, err := sys.GetState(epoch)
		if err != nil {
			return err
		}
		gstate.TotalQuantity = new(big.Int).Sub(gstate.TotalQuantity, prod.TotalQuantity)
		if fid >= params.ForkID2 {
			prod.Type = Black
			if err := sys.updateState(gstate, prod); err != nil {
				return err
			}
		}
		if err := sys.SetState(gstate); err != nil {
			return err
		}
	}

	prod.Number = number
	prod.Type = Black
	return sys.SetCandidate(prod)
}

// RemoveKickedCandidate remove
func (sys *System) RemoveKickedCandidate(epoch uint64, candidate string, number uint64, fid uint64) error {
	// name validity
	prod, err := sys.GetCandidate(epoch, candidate)
	if prod == nil || err != nil {
		return err
	}
	if prod.Type != Black {
		return nil
	}

	if err := sys.DelCandidate(epoch, prod.Name); err != nil {
		return err
	}
	return nil
}

// ExitTakeOver system exit take over
func (sys *System) ExitTakeOver(epoch uint64, number uint64, fid uint64) error {
	gstate, err := sys.GetState(epoch)
	if err != nil {
		return err
	}
	if fid >= params.ForkID2 {
		epoch, err := sys.GetTakeOver()
		if err != nil {
			return err
		}
		if gstate.Epoch == epoch {
			return fmt.Errorf("take over must in diff epoch")
		}
	}
	gstate.TakeOver = false
	return sys.SetState(gstate)
}

// UpdateElectedCandidates0 update
func (sys *System) UpdateElectedCandidates0(pepoch uint64, epoch uint64, number uint64, miner string) error {
	if pepoch > epoch {
		panic(fmt.Errorf("UpdateElectedCandidates unreached"))
	}
	pstate, err := sys.GetState(pepoch)
	if err != nil {
		return err
	}

	// not is first & no changes
	if pstate.Epoch != pstate.PreEpoch && pepoch == epoch {
		return nil
	}

	candidateInfoArray, err := sys.GetCandidates(pepoch)
	if err != nil {
		return err
	}
	sort.Sort(candidateInfoArray)
	n := sys.config.BackupScheduleSize + sys.config.CandidateScheduleSize
	activatedCandidateSchedule := []string{}
	activatedTotalQuantity := big.NewInt(0)
	totalQuantity := big.NewInt(0)
	quantity := big.NewInt(0)
	cnt := uint64(0)
	ntotalQuantity := big.NewInt(0)
	candidates := []*CandidateInfo{}
	var sysCandidate *CandidateInfo
	for _, candidateInfo := range candidateInfoArray {
		if pepoch != epoch {
			// clear vote quantity
			tcandidateInfo := candidateInfo.copy()
			tcandidateInfo.Epoch = epoch
			tcandidateInfo.TotalQuantity = tcandidateInfo.Quantity
			if !tcandidateInfo.invalid() {
				ntotalQuantity = new(big.Int).Add(ntotalQuantity, tcandidateInfo.TotalQuantity)
			}
			if err := sys.SetCandidate(tcandidateInfo); err != nil {
				return err
			}
		}

		if !candidateInfo.invalid() {
			if pstate.Dpos {
				if candidateInfo.Quantity.Sign() == 0 || strings.Compare(candidateInfo.Name, sys.config.SystemName) == 0 {
					continue
				}
			} else if candidateInfo.Quantity.Sign() == 0 || strings.Compare(candidateInfo.Name, sys.config.SystemName) == 0 {
				if strings.Compare(candidateInfo.Name, sys.config.SystemName) == 0 {
					sysCandidate = candidateInfo
				} else {
					candidates = append(candidates, candidateInfo)
				}
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
	}

	if !pstate.Dpos && totalQuantity.Cmp(sys.config.ActivatedMinQuantity) >= 0 &&
		cnt >= n && cnt >= sys.config.ActivatedMinCandidate {
		pstate.Dpos = true
	}

	if !pstate.Dpos {
		activatedTotalQuantity = big.NewInt(0)
		activatedCandidateSchedule = []string{}
		activatedCandidateSchedule = append(activatedCandidateSchedule, sysCandidate.Name)
		activatedTotalQuantity = new(big.Int).Add(activatedTotalQuantity, sysCandidate.TotalQuantity)
		for index, candidateInfo := range candidates {
			if uint64(index) >= n-1 {
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

	if pepoch != epoch {
		gstate := &GlobalState{
			Epoch:                       epoch,
			PreEpoch:                    pstate.Epoch,
			ActivatedTotalQuantity:      big.NewInt(0),
			TotalQuantity:               new(big.Int).SetBytes(ntotalQuantity.Bytes()),
			UsingCandidateIndexSchedule: []uint64{},
			BadCandidateIndexSchedule:   []uint64{},
			TakeOver:                    pstate.TakeOver,
			Dpos:                        pstate.Dpos,
		}
		if err := sys.SetLastestEpoch(epoch); err != nil {
			return err
		}
		return sys.SetState(gstate)
	}
	return nil
}

// UpdateElectedCandidates1 update
func (sys *System) UpdateElectedCandidates1(pepoch uint64, epoch uint64, number uint64, miner string) error {
	if pepoch > epoch {
		panic(fmt.Errorf("UpdateElectedCandidates unreached"))
	}
	pstate, err := sys.GetState(pepoch)
	if err != nil {
		return err
	}
	if pepoch == epoch &&
		len(pstate.ActivatedCandidateSchedule) != 0 {
		return nil
	}

	t := time.Now()
	defer func() {
		log.Debug("UpdateElectedCandidates1", "pepoch", pepoch, "epoch", epoch, "number", number, "elapsed", common.PrettyDuration(time.Now().Sub(t)))
	}()
	n := sys.config.BackupScheduleSize + sys.config.CandidateScheduleSize
	initActivatedCandidateSchedule := func(gstate *GlobalState, candidateInfoArray CandidateInfoArray) error {
		activatedCandidateSchedule := []string{}
		activatedTotalQuantity := big.NewInt(0)
		sort.Sort(candidateInfoArray)
		if gstate.Dpos {
			for _, candidateInfo := range candidateInfoArray {
				if !candidateInfo.invalid() {
					if candidateInfo.Quantity.Sign() == 0 || strings.Compare(candidateInfo.Name, sys.config.SystemName) == 0 {
						continue
					}
					if uint64(len(activatedCandidateSchedule)) >= n {
						break
					}
					activatedCandidateSchedule = append(activatedCandidateSchedule, candidateInfo.Name)
					activatedTotalQuantity = new(big.Int).Add(activatedTotalQuantity, candidateInfo.TotalQuantity)
				}
			}
			gstate.ActivatedCandidateSchedule = activatedCandidateSchedule
			gstate.ActivatedTotalQuantity = activatedTotalQuantity
		} else {
			tstate := &GlobalState{
				Epoch:                       math.MaxUint64,
				PreEpoch:                    math.MaxUint64,
				ActivatedTotalQuantity:      big.NewInt(0),
				TotalQuantity:               big.NewInt(0),
				UsingCandidateIndexSchedule: []uint64{},
				BadCandidateIndexSchedule:   []uint64{},
				Number:                      0,
			}
			for _, candidateInfo := range candidateInfoArray {
				if !candidateInfo.invalid() {
					if candidateInfo.Quantity.Sign() != 0 && strings.Compare(candidateInfo.Name, sys.config.SystemName) != 0 {
						tstate.Number++
						tstate.TotalQuantity = new(big.Int).Add(tstate.TotalQuantity, candidateInfo.TotalQuantity)
						if uint64(len(tstate.ActivatedCandidateSchedule)) < n {
							tstate.ActivatedCandidateSchedule = append(tstate.ActivatedCandidateSchedule, candidateInfo.Name)
							tstate.ActivatedTotalQuantity = new(big.Int).Add(tstate.ActivatedTotalQuantity, candidateInfo.TotalQuantity)
						}
						continue
					}
					if uint64(len(activatedCandidateSchedule)) < n {
						activatedCandidateSchedule = append(activatedCandidateSchedule, candidateInfo.Name)
						activatedTotalQuantity = new(big.Int).Add(activatedTotalQuantity, candidateInfo.TotalQuantity)
					}
				}
			}

			if tstate.TotalQuantity.Cmp(sys.config.ActivatedMinQuantity) >= 0 &&
				tstate.Number >= n &&
				tstate.Number >= sys.config.ActivatedMinCandidate {
				gstate.Dpos = true
				gstate.ActivatedTotalQuantity = tstate.ActivatedTotalQuantity
				gstate.ActivatedCandidateSchedule = tstate.ActivatedCandidateSchedule
			} else {
				if err := sys.SetState(tstate); err != nil {
					return err
				}
				if init := len(activatedCandidateSchedule); init > 0 {
					index := 0
					for uint64(len(activatedCandidateSchedule)) < sys.config.CandidateScheduleSize {
						activatedCandidateSchedule = append(activatedCandidateSchedule, activatedCandidateSchedule[index%init])
						index++
					}
				}
				gstate.ActivatedCandidateSchedule = activatedCandidateSchedule
				gstate.ActivatedTotalQuantity = activatedTotalQuantity
			}
		}

		if err := sys.SetState(gstate); err != nil {
			return err
		}
		return nil
	}

	candidateInfoArray, err := sys.GetCandidates(pstate.Epoch)
	if err != nil {
		return err
	}
	if len(pstate.ActivatedCandidateSchedule) == 0 {
		ppstate, err := sys.GetState(pstate.PreEpoch)
		if err != nil {
			return err
		}
		usingCandidateIndexSchedule := []uint64{}
		for index := range ppstate.ActivatedCandidateSchedule {
			if uint64(index) >= sys.config.CandidateScheduleSize {
				break
			}
			usingCandidateIndexSchedule = append(usingCandidateIndexSchedule, uint64(index))
		}
		for index, offset := range ppstate.BadCandidateIndexSchedule {
			usingCandidateIndexSchedule[int(offset)] = sys.config.CandidateScheduleSize + uint64(index)
		}
		ppstate.UsingCandidateIndexSchedule = usingCandidateIndexSchedule
		if err := sys.SetState(ppstate); err != nil {
			return err
		}

		if err := initActivatedCandidateSchedule(pstate, candidateInfoArray); err != nil {
			return err
		}
	}

	if pepoch != epoch {
		usingCandidateIndexSchedule := []uint64{}
		for index := range pstate.ActivatedCandidateSchedule {
			if uint64(index) >= sys.config.CandidateScheduleSize {
				break
			}
			usingCandidateIndexSchedule = append(usingCandidateIndexSchedule, uint64(index))
		}
		for index, offset := range pstate.BadCandidateIndexSchedule {
			usingCandidateIndexSchedule[int(offset)] = sys.config.CandidateScheduleSize + uint64(index)
		}
		pstate.UsingCandidateIndexSchedule = usingCandidateIndexSchedule
		if err := sys.SetState(pstate); err != nil {
			return err
		}

		tcandidateInfoArray := CandidateInfoArray{}
		gstate := &GlobalState{
			Epoch:                       epoch,
			PreEpoch:                    pepoch,
			ActivatedTotalQuantity:      big.NewInt(0),
			TotalQuantity:               big.NewInt(0),
			UsingCandidateIndexSchedule: []uint64{},
			BadCandidateIndexSchedule:   []uint64{},
			TakeOver:                    pstate.TakeOver,
			Dpos:                        pstate.Dpos,
			Number:                      number,
		}
		for _, candidateInfo := range candidateInfoArray {
			tcandidateInfo := candidateInfo.copy()
			tcandidateInfo.Epoch = epoch
			tcandidateInfo.TotalQuantity = tcandidateInfo.Quantity
			if !tcandidateInfo.invalid() {
				gstate.TotalQuantity = new(big.Int).Add(gstate.TotalQuantity, tcandidateInfo.TotalQuantity)
			}
			if err := sys.SetCandidate(tcandidateInfo); err != nil {
				return err
			}
			tcandidateInfoArray = append(tcandidateInfoArray, tcandidateInfo)
		}
		if err := initActivatedCandidateSchedule(gstate, tcandidateInfoArray); err != nil {
			return err
		}
		if err := sys.SetLastestEpoch(gstate.Epoch); err != nil {
			return err
		}
	}
	return nil
}
func (sys *System) getAvailableQuantity(epoch uint64, voter string) (*big.Int, error) {
	q, err := sys.GetAvailableQuantity(epoch, voter)
	if err != nil {
		return nil, err
	}
	if q == nil {
		timestamp := sys.config.epochTimeStamp(epoch)
		gstate, err := sys.GetState(epoch)
		if err != nil {
			return nil, err
		}
		if sys.config.epoch(sys.config.ReferenceTime) == gstate.PreEpoch {
			timestamp = sys.config.epochTimeStamp(gstate.PreEpoch)
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

func (sys *System) usingCandiate(gstate *GlobalState, offset uint64) string {
	size := uint64(len(gstate.UsingCandidateIndexSchedule))
	if offset >= size {
		return ""
	}
	index := gstate.UsingCandidateIndexSchedule[offset]
	if index == InvalidIndex {
		return ""
	}
	return gstate.ActivatedCandidateSchedule[index]
}

func (sys *System) updateState(gstate *GlobalState, prod *CandidateInfo) error {
	if prod.Quantity.Sign() == 0 ||
		strings.Compare(prod.Name, sys.config.SystemName) == 0 {
		return nil
	}
	// timestamp := sys.config.epochTimeStamp(gstate.Epoch)
	// if bquantity, err := sys.GetBalanceByTime(prod.Name, timestamp); err != nil {
	// 	log.Debug("insert", "candidate", prod.Name, "ignore", err)
	// 	return nil
	// } else if s := new(big.Int).Mul(sys.config.unitStake(), sys.config.CandidateAvailableMinQuantity); bquantity.Cmp(s) == -1 {
	// 	log.Debug("insert", "candidate", prod.Name, "ignore", "insufficient available quantity")
	// 	return nil
	// }

	insert := func(gstate *GlobalState, prod *CandidateInfo) error {
		n := sys.config.CandidateScheduleSize + sys.config.BackupScheduleSize
		var low *CandidateInfo
		if cnt := len(gstate.ActivatedCandidateSchedule); uint64(cnt) == n {
			lowprod, err := sys.GetCandidate(prod.Epoch, gstate.ActivatedCandidateSchedule[cnt-1])
			if err != nil {
				return err
			}
			if cmp := lowprod.TotalQuantity.Cmp(prod.TotalQuantity); cmp == 1 {
				return nil
			}
			low = lowprod
		}

		if prod.invalid() {
			has := false
			findex := 0
			names := map[string]bool{}
			for index, name := range gstate.ActivatedCandidateSchedule {
				names[name] = true
				if strings.Compare(name, prod.Name) == 0 {
					findex = index
					has = true
				}
			}
			if !has {
				return nil
			}
			gstate.ActivatedTotalQuantity = new(big.Int).Sub(gstate.ActivatedTotalQuantity, prod.TotalQuantity)
			gstate.ActivatedCandidateSchedule = append(gstate.ActivatedCandidateSchedule[:findex], gstate.ActivatedCandidateSchedule[findex+1:]...)

			candidateInfoArray, err := sys.GetCandidates(prod.Epoch)
			if err != nil {
				return err
			}
			var sprod *CandidateInfo
			for _, tprod := range candidateInfoArray {
				if !tprod.invalid() {
					if _, ok := names[tprod.Name]; ok {
						continue
					}
					if tprod.Quantity.Sign() == 0 ||
						strings.Compare(tprod.Name, sys.config.SystemName) == 0 {
						continue
					}
					if sprod == nil || more(tprod, sprod) {
						sprod = tprod
						log.Debug("updateState", "candiate invalid", prod.Name, "replace", sprod.Name)
					}
				}
			}
			if sprod == nil {
				log.Debug("updateState", "candiate invalid", prod.Name)
				return nil
			}
			log.Debug("updateState", "candiate invalid", prod.Name, "replaced", sprod.Name)
			prod = sprod
		}

		activatedCandidateSchedule := []string{}
		has := false
		for _, name := range gstate.ActivatedCandidateSchedule {
			tprod, err := sys.GetCandidate(prod.Epoch, name)
			if err != nil {
				return err
			}
			if strings.Compare(prod.Name, name) == 0 {
				gstate.ActivatedTotalQuantity = new(big.Int).Sub(gstate.ActivatedTotalQuantity, tprod.TotalQuantity)
				continue
			}
			if !has && more(prod, tprod) {
				has = true
				gstate.ActivatedTotalQuantity = new(big.Int).Add(gstate.ActivatedTotalQuantity, prod.TotalQuantity)
				activatedCandidateSchedule = append(activatedCandidateSchedule, prod.Name)
			}
			activatedCandidateSchedule = append(activatedCandidateSchedule, name)
		}
		if cnt := len(activatedCandidateSchedule); uint64(cnt) > n {
			activatedCandidateSchedule = activatedCandidateSchedule[:n]
			gstate.ActivatedTotalQuantity = new(big.Int).Sub(gstate.ActivatedTotalQuantity, low.TotalQuantity)
		} else if !has && uint64(cnt) < n {
			gstate.ActivatedTotalQuantity = new(big.Int).Add(gstate.ActivatedTotalQuantity, prod.TotalQuantity)
			activatedCandidateSchedule = append(activatedCandidateSchedule, prod.Name)
		}
		gstate.ActivatedCandidateSchedule = activatedCandidateSchedule
		return nil
	}

	if !gstate.Dpos {
		epoch := uint64(math.MaxUint64)
		tstate, err := sys.GetState(epoch)
		if err != nil {
			return err
		}

		if prod.invalid() {
			tstate.TotalQuantity = new(big.Int).Sub(tstate.TotalQuantity, prod.TotalQuantity)
			tstate.Number--
		} else {
			tprod, err := sys.GetCandidate(prod.Epoch, prod.Name)
			if err != nil {
				return err
			}
			if tprod == nil {
				tstate.TotalQuantity = new(big.Int).Add(tstate.TotalQuantity, prod.TotalQuantity)
				tstate.Number++
			} else {
				tstate.TotalQuantity = new(big.Int).Add(tstate.TotalQuantity, new(big.Int).Sub(prod.TotalQuantity, tprod.TotalQuantity))
			}
		}

		if err := insert(tstate, prod); err != nil {
			return err
		}

		if tstate.TotalQuantity.Cmp(sys.config.ActivatedMinQuantity) >= 0 &&
			tstate.Number >= sys.config.BackupScheduleSize+sys.config.CandidateScheduleSize &&
			tstate.Number >= sys.config.ActivatedMinCandidate {
			gstate.Dpos = true
			gstate.ActivatedTotalQuantity = tstate.ActivatedTotalQuantity
			gstate.ActivatedCandidateSchedule = tstate.ActivatedCandidateSchedule
		}

		if err := sys.SetState(tstate); err != nil {
			return err
		}
		return nil
	}
	if err := insert(gstate, prod); err != nil {
		return err
	}
	return nil
}
