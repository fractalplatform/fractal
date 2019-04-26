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
	"math/big"
	"sort"
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

	// VoterKeyPrefix voterInfo
	VoterKeyPrefix = "v"
	// VoterHead head
	VoterHead = "v"

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
		head, err := db.GetCandidate(CandidateHead)
		if err != nil {
			return err
		}
		if head == nil {
			head = &CandidateInfo{
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
			next, err := db.GetCandidate(candidate.NextKey)
			if err != nil {
				return err
			}
			next.PrevKey = candidate.Name
			if err := db.SetCandidate(next); err != nil {
				return err
			}
		}
	}
	key := strings.Join([]string{CandidateKeyPrefix, candidate.Name}, Separator)
	if val, err := rlp.EncodeToBytes(candidate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}
	return nil
}

// DelCandidate del candidate info
func (db *LDB) DelCandidate(name string) error {
	candidateInfo, err := db.GetCandidate(name)
	if err != nil {
		return err
	}
	if candidateInfo == nil {
		return nil
	}

	prev, err := db.GetCandidate(candidateInfo.PrevKey)
	if err != nil {
		return err
	}
	prev.NextKey = candidateInfo.NextKey
	if err := db.SetCandidate(prev); err != nil {
		return err
	}

	if len(candidateInfo.NextKey) > 0 {
		next, err := db.GetCandidate(candidateInfo.NextKey)
		if err != nil {
			return err
		}
		next.PrevKey = candidateInfo.PrevKey
		if err := db.SetCandidate(next); err != nil {
			return err
		}
	}
	key := strings.Join([]string{CandidateKeyPrefix, name}, Separator)
	if err := db.Delete(key); err != nil {
		return err
	}
	head, err := db.GetCandidate(CandidateHead)
	if err != nil {
		return err
	}
	head.Counter--
	return db.SetCandidate(head)
}

// GetCandidate get candidate info by name
func (db *LDB) GetCandidate(name string) (*CandidateInfo, error) {
	key := strings.Join([]string{CandidateKeyPrefix, name}, Separator)
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
func (db *LDB) GetCandidates() ([]*CandidateInfo, error) {
	// candidates
	head, err := db.GetCandidate(CandidateHead)
	if err != nil {
		return nil, err
	}
	if head == nil {
		return nil, nil
	}

	candidateInfos := CandidateInfoArray{}
	nextKey := head.NextKey
	for len(nextKey) != 0 {
		candidateInfo, err := db.GetCandidate(nextKey)
		if err != nil {
			return nil, err
		}
		nextKey = candidateInfo.NextKey
		candidateInfos = append(candidateInfos, candidateInfo)
	}
	sort.Sort(candidateInfos)
	return candidateInfos, nil
}

// CandidatesSize candidate len
func (db *LDB) CandidatesSize() (uint64, error) {
	head, err := db.GetCandidate(CandidateHead)
	if err != nil {
		return 0, err
	}
	if head == nil {
		return 0, nil
	}
	return head.Counter, nil
}

// SetAvailableQuantity set quantity
func (db *LDB) SetAvailableQuantity(epcho uint64, voter string, quantity *big.Int) error {
	head, err := db.GetVoter(epcho, voter, VoterHead)
	if err != nil {
		return err
	}
	if head == nil {
		head = &VoterInfo{
			Epcho:     epcho,
			Name:      voter,
			Candidate: VoterHead,
		}
	}
	head.Quantity = quantity
	return db.SetVoter(head)
}

// GetAvailableQuantity get quantity
func (db *LDB) GetAvailableQuantity(epcho uint64, voter string) (*big.Int, error) {
	head, err := db.GetVoter(epcho, voter, VoterHead)
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
			head, err := db.GetVoter(voter.Epcho, voter.Name, VoterHead)
			if err != nil {
				return err
			}
			voter.NextKeyForVoter = head.NextKeyForVoter
			head.NextKeyForVoter = voter.key()
			if err := db.SetVoter(head); err != nil {
				return err
			}
		}
		if voter.Name != CandidateHead {
			head, err := db.GetVoter(voter.Epcho, CandidateHead, voter.Candidate)
			if err != nil {
				return err
			}
			if head == nil {
				head = &VoterInfo{
					Epcho:     voter.Epcho,
					Name:      CandidateHead,
					Candidate: voter.Candidate,
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
func (db *LDB) GetVoter(epcho uint64, voter string, candidate string) (*VoterInfo, error) {
	vf := &VoterInfo{
		Epcho:     epcho,
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
func (db *LDB) GetVotersByCandidate(epcho uint64, candidate string) ([]*VoterInfo, error) {
	head, err := db.GetVoter(epcho, CandidateHead, candidate)
	if err != nil {
		return nil, err
	}
	if head == nil || len(head.NextKeyForCandidate) == 0 {
		return nil, nil
	}

	voterInfos := []*VoterInfo{}
	nextKey := head.NextKeyForCandidate
	for len(nextKey) != 0 {
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
func (db *LDB) GetVotersByVoter(epcho uint64, voter string) ([]*VoterInfo, error) {
	head, err := db.GetVoter(epcho, voter, VoterHead)
	if err != nil {
		return nil, err
	}
	if head == nil || len(head.NextKeyForVoter) == 0 {
		return nil, nil
	}

	voterInfos := []*VoterInfo{}
	nextKey := head.NextKeyForVoter
	for len(nextKey) != 0 {
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

// SetState set global state info
func (db *LDB) SetState(gstate *GlobalState) error {
	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(gstate.Epcho))}, Separator)
	if val, err := rlp.EncodeToBytes(gstate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}

	lkey := strings.Join([]string{StateKeyPrefix, LastestStateKey}, Separator)
	if err := db.Put(lkey, uint64tobytes(gstate.Epcho)); err != nil {
		return err
	}
	return nil
}

// GetState get state info
func (db *LDB) GetState(epcho uint64) (*GlobalState, error) {
	if epcho == LastEpcho {
		var err error
		epcho, err = db.GetLastestEpcho()
		if err != nil {
			return nil, err
		}
	}

	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(epcho))}, Separator)
	gstate := &GlobalState{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(val, gstate); err != nil {
		return nil, err
	}
	return gstate, nil
}

// GetDelegatedByTime candidate delegate
func (db *LDB) GetDelegatedByTime(candidate string, timestamp uint64) (*big.Int, *big.Int, uint64, error) {
	key := strings.Join([]string{CandidateKeyPrefix, candidate}, Separator)
	val, err := db.GetSnapshot(key, timestamp)
	if val == nil || err != nil {
		return big.NewInt(0), big.NewInt(0), 0, err
	}
	candidateInfo := &CandidateInfo{}
	if err := rlp.DecodeBytes(val, candidateInfo); err != nil {
		return big.NewInt(0), big.NewInt(0), 0, err
	}
	if candidateInfo.Type == Black {
		return big.NewInt(0), candidateInfo.TotalQuantity, candidateInfo.Counter, nil
	}
	return candidateInfo.Quantity, candidateInfo.TotalQuantity, candidateInfo.Counter, nil
}

// GetLastestEpcho get latest epcho
func (db *LDB) GetLastestEpcho() (uint64, error) {
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
