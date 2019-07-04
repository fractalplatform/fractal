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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// IDatabase level db
type IDatabase interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error

	Undelegate(string, *big.Int) (*types.Action, error)
	IncAsset2Acct(string, string, *big.Int) (*types.Action, error)
	GetBalanceByTime(name string, timestamp uint64) (*big.Int, error)

	GetSnapshot(key string, timestamp uint64) ([]byte, error)
}

var (
	// CandidateKeyPrefix candidateInfo
	CandidateKeyPrefix = "p"
	// CandidateHead all candidate key
	CandidateHead = "s"
	// ActivatedCandidateKeyPrefix candidateInfo
	ActivatedCandidateKeyPrefix = "ap"

	// VoterKeyPrefix voterInfo
	VoterKeyPrefix = "v"
	// VoterHead head
	VoterHead = "v"

	// TakeOver key
	TakeOver = "takeover"

	// StateKeyPrefix globalState
	StateKeyPrefix = "s"
	// LastestStateKey lastest
	LastestStateKey = "lastest"

	// Separator Split characters
	Separator = "_"
)

// LDB dpos level db
type LDB struct {
	IDatabase
}

var _ IDB = &LDB{}

// NewLDB new object
func NewLDB(db IDatabase) (*LDB, error) {
	ldb := &LDB{
		IDatabase: db,
	}
	return ldb, nil
}

// SetCandidate update candidate info
func (db *LDB) SetCandidate(candidate *CandidateInfo) error {
	if candidate.Name != CandidateHead && len(candidate.PrevKey) == 0 && len(candidate.NextKey) == 0 {
		head, err := db.GetCandidate(candidate.Epoch, CandidateHead)
		if err != nil {
			return err
		}
		if head == nil {
			head = &CandidateInfo{
				Epoch:   candidate.Epoch,
				Name:    CandidateHead,
				Counter: 0,
			}
		}
		candidate.NextKey = head.NextKey
		candidate.PrevKey = head.Name
		head.NextKey = candidate.Name
		head.Counter++
		if err := db.SetCandidate(head); err != nil {
			return err
		}

		if len(candidate.NextKey) != 0 {
			next, err := db.GetCandidate(candidate.Epoch, candidate.NextKey)
			if err != nil {
				return err
			}
			next.PrevKey = candidate.Name
			if err := db.SetCandidate(next); err != nil {
				return err
			}
		}
	}
	key := strings.Join([]string{CandidateKeyPrefix, fmt.Sprintf("0x%x_%s", candidate.Epoch, candidate.Name)}, Separator)
	if val, err := rlp.EncodeToBytes(candidate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}
	return nil
}

// DelCandidate del candidate info
func (db *LDB) DelCandidate(epoch uint64, name string) error {
	candidateInfo, err := db.GetCandidate(epoch, name)
	if err != nil {
		return err
	}
	if candidateInfo == nil {
		return nil
	}

	prev, err := db.GetCandidate(epoch, candidateInfo.PrevKey)
	if err != nil {
		return err
	}
	prev.NextKey = candidateInfo.NextKey
	if err := db.SetCandidate(prev); err != nil {
		return err
	}

	if len(candidateInfo.NextKey) > 0 {
		next, err := db.GetCandidate(epoch, candidateInfo.NextKey)
		if err != nil {
			return err
		}
		next.PrevKey = candidateInfo.PrevKey
		if err := db.SetCandidate(next); err != nil {
			return err
		}
	}
	key := strings.Join([]string{CandidateKeyPrefix, fmt.Sprintf("0x%x_%s", epoch, name)}, Separator)
	if err := db.Delete(key); err != nil {
		return err
	}
	head, err := db.GetCandidate(epoch, CandidateHead)
	if err != nil {
		return err
	}
	head.Counter--
	return db.SetCandidate(head)
}

// GetCandidate get candidate info by name
func (db *LDB) GetCandidate(epoch uint64, name string) (*CandidateInfo, error) {
	key := strings.Join([]string{CandidateKeyPrefix, fmt.Sprintf("0x%x_%s", epoch, name)}, Separator)
	candidateInfo := &CandidateInfo{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(val, candidateInfo); err != nil {
		return nil, err
	}
	return candidateInfo, nil
}

// GetCandidates get all candidate info & sort
func (db *LDB) GetCandidates(epoch uint64) (CandidateInfoArray, error) {
	// candidates
	head, err := db.GetCandidate(epoch, CandidateHead)
	if err != nil {
		return nil, err
	}
	if head == nil {
		return nil, nil
	}

	candidateInfos := CandidateInfoArray{}
	nextKey := head.NextKey
	for len(nextKey) != 0 {
		candidateInfo, err := db.GetCandidate(epoch, nextKey)
		if err != nil {
			return nil, err
		}
		nextKey = candidateInfo.NextKey
		candidateInfos = append(candidateInfos, candidateInfo)
	}
	// sort.Sort(candidateInfos)
	return candidateInfos, nil
}

// CandidatesSize candidate len
func (db *LDB) CandidatesSize(epoch uint64) (uint64, error) {
	head, err := db.GetCandidate(epoch, CandidateHead)
	if err != nil {
		return 0, err
	}
	if head == nil {
		return 0, nil
	}
	return head.Counter, nil
}

// SetActivatedCandidate update activated candidate info
func (db *LDB) SetActivatedCandidate(index uint64, candidate *CandidateInfo) error {
	key := strings.Join([]string{ActivatedCandidateKeyPrefix, hex.EncodeToString(uint64tobytes(index))}, Separator)
	if val, err := rlp.EncodeToBytes(candidate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}
	return nil
}

// GetActivatedCandidate get activated candidate info
func (db *LDB) GetActivatedCandidate(index uint64) (*CandidateInfo, error) {
	key := strings.Join([]string{ActivatedCandidateKeyPrefix, hex.EncodeToString(uint64tobytes(index))}, Separator)
	candidateInfo := &CandidateInfo{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(val, candidateInfo); err != nil {
		return nil, err
	}
	return candidateInfo, nil
}

// SetAvailableQuantity set quantity
func (db *LDB) SetAvailableQuantity(epoch uint64, voter string, quantity *big.Int) error {
	head, err := db.GetVoter(epoch, voter, VoterHead)
	if err != nil {
		return err
	}
	if head == nil {
		head = &VoterInfo{
			Epoch:           epoch,
			Name:            voter,
			Candidate:       VoterHead,
			NextKeyForVoter: "EOF",
		}
	}
	head.Quantity = quantity
	return db.SetVoter(head)
}

// GetAvailableQuantity get quantity
func (db *LDB) GetAvailableQuantity(epoch uint64, voter string) (*big.Int, error) {
	head, err := db.GetVoter(epoch, voter, VoterHead)
	if err != nil {
		return nil, err
	}
	if head == nil {
		return nil, nil
	}
	return head.Quantity, nil
}

// SetVoter update voter info
func (db *LDB) SetVoter(voter *VoterInfo) error {
	if len(voter.NextKeyForVoter) == 0 && len(voter.NextKeyForCandidate) == 0 {
		if voter.Candidate != VoterHead {
			head, err := db.GetVoter(voter.Epoch, voter.Name, VoterHead)
			if err != nil {
				return err
			}
			if head == nil {
				head = &VoterInfo{
					Epoch:           voter.Epoch,
					Name:            voter.Name,
					Candidate:       VoterHead,
					NextKeyForVoter: "EOF",
				}
			}
			voter.NextKeyForVoter = head.NextKeyForVoter
			head.NextKeyForVoter = voter.key()
			if err := db.SetVoter(head); err != nil {
				return err
			}
		}
		if voter.Name != CandidateHead {
			head, err := db.GetVoter(voter.Epoch, CandidateHead, voter.Candidate)
			if err != nil {
				return err
			}
			if head == nil {
				head = &VoterInfo{
					Epoch:               voter.Epoch,
					Name:                CandidateHead,
					Candidate:           voter.Candidate,
					NextKeyForCandidate: "EOF",
				}
			}
			voter.NextKeyForCandidate = head.NextKeyForCandidate
			head.NextKeyForCandidate = voter.key()
			if err := db.SetVoter(head); err != nil {
				return err
			}
		}
	}
	key := strings.Join([]string{VoterKeyPrefix, voter.key()}, Separator)
	if val, err := rlp.EncodeToBytes(voter); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}

	return nil
}

// GetVoter voter info by name
func (db *LDB) GetVoter(epoch uint64, voter string, candidate string) (*VoterInfo, error) {
	vf := &VoterInfo{
		Epoch:     epoch,
		Name:      voter,
		Candidate: candidate,
	}
	key := strings.Join([]string{VoterKeyPrefix, vf.key()}, Separator)
	voterInfo := &VoterInfo{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(val, voterInfo); err != nil {
		return nil, err
	}
	return voterInfo, nil
}

// GetVotersByCandidate voters info by candidate
func (db *LDB) GetVotersByCandidate(epoch uint64, candidate string) ([]*VoterInfo, error) {
	head, err := db.GetVoter(epoch, CandidateHead, candidate)
	if err != nil {
		return nil, err
	}
	if head == nil || len(head.NextKeyForCandidate) == 0 {
		return nil, nil
	}

	voterInfos := []*VoterInfo{}
	nextKey := head.NextKeyForCandidate
	for nextKey != "EOF" {
		next := &VoterInfo{}
		if val, err := db.Get(strings.Join([]string{VoterKeyPrefix, nextKey}, Separator)); err != nil {
			return nil, err
		} else if val == nil {
		} else if err := rlp.DecodeBytes(val, next); err != nil {
			return nil, err
		}
		voterInfos = append(voterInfos, next)
		nextKey = next.NextKeyForCandidate
	}
	return voterInfos, nil
}

// GetVotersByVoter voters info by voter
func (db *LDB) GetVotersByVoter(epoch uint64, voter string) ([]*VoterInfo, error) {
	head, err := db.GetVoter(epoch, voter, VoterHead)
	if err != nil {
		return nil, err
	}
	if head == nil || len(head.NextKeyForVoter) == 0 {
		return nil, nil
	}

	voterInfos := []*VoterInfo{}
	nextKey := head.NextKeyForVoter
	for nextKey != "EOF" {
		next := &VoterInfo{}
		if val, err := db.Get(strings.Join([]string{VoterKeyPrefix, nextKey}, Separator)); err != nil {
			return nil, err
		} else if val == nil {
		} else if err := rlp.DecodeBytes(val, next); err != nil {
			return nil, err
		}
		voterInfos = append(voterInfos, next)
		nextKey = next.NextKeyForVoter
	}
	return voterInfos, nil
}

// SetTakeOver update activated candidate info
func (db *LDB) SetTakeOver(epoch uint64) error {
	if val, err := rlp.EncodeToBytes(epoch); err != nil {
		return err
	} else if err := db.Put(TakeOver, val); err != nil {
		return err
	}
	return nil
}

// GetTakeOver get activated candidate info
func (db *LDB) GetTakeOver() (uint64, error) {
	epoch := uint64(0)
	if val, err := db.Get(TakeOver); err != nil {
		return epoch, err
	} else if val == nil {
		return epoch, nil
	} else if err := rlp.DecodeBytes(val, &epoch); err != nil {
		return epoch, err
	}
	return epoch, nil
}

// SetState set global state info
func (db *LDB) SetState(gstate *GlobalState) error {
	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(gstate.Epoch))}, Separator)
	if val, err := rlp.EncodeToBytes(gstate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}
	return nil
}

// GetState get state info
func (db *LDB) GetState(epoch uint64) (*GlobalState, error) {
	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(epoch))}, Separator)
	gstate := &GlobalState{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, fmt.Errorf("epoch not found")
	} else if err := rlp.DecodeBytes(val, gstate); err != nil {
		return nil, err
	}
	return gstate, nil
}

// GetCandidateInfoByTime candidate info
func (db *LDB) GetCandidateInfoByTime(epoch uint64, candidate string, timestamp uint64) (*CandidateInfo, error) {
	key := strings.Join([]string{CandidateKeyPrefix, fmt.Sprintf("0x%x_%s", epoch, candidate)}, Separator)
	val, err := db.GetSnapshot(key, timestamp)
	if err != nil || val == nil {
		return nil, err
	}
	// if len(val) == 0 {
	// 	return nil, nil
	// }
	candidateInfo := &CandidateInfo{}
	if err := rlp.DecodeBytes(val, candidateInfo); err != nil {
		return nil, err
	}
	return candidateInfo, nil
}

// SetLastestEpoch set latest epoch
func (db *LDB) SetLastestEpoch(epoch uint64) error {
	lkey := strings.Join([]string{StateKeyPrefix, LastestStateKey}, Separator)
	return db.Put(lkey, uint64tobytes(epoch))
}

// GetLastestEpoch get latest epoch
func (db *LDB) GetLastestEpoch() (uint64, error) {
	lkey := strings.Join([]string{StateKeyPrefix, LastestStateKey}, Separator)
	if val, err := db.Get(lkey); err != nil {
		return 0, err
	} else if val == nil {
		return 0, nil
	} else {
		return bytestouint64(val), nil
	}
}

func uint64tobytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}

func bytestouint64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}
