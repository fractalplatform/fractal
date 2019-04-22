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

	"github.com/fractalplatform/fractal/utils/rlp"
)

// IDatabase level db
type IDatabase interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error

	Delegate(string, *big.Int) error
	Undelegate(string, *big.Int) error
	IncAsset2Acct(string, string, *big.Int) error

	GetSnapshot(key string, timestamp uint64) ([]byte, error)
}

var (
	// CandidateKeyPrefix candidateInfo
	CandidateKeyPrefix = "p"
	// CandidatesKey all candidate key
	CandidatesKey = "s"
	// CandidatesSizeKey len of candidate
	CandidatesSizeKey = "sl"

	// VoterQuantityPrefix quantity
	VoterQuantityPrefix = "q"
	// VoterKeyPrefix voterInfo
	VoterKeyPrefix = "v"
	// DelegatorsKeyPrfix voters of candidate
	DelegatorsKeyPrfix = "vs"
	// VotersKeyPrefix candidates of voter
	VotersKeyPrefix = "ps"

	// StateKeyPrefix globalState
	StateKeyPrefix = "s"
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
	key := strings.Join([]string{CandidateKeyPrefix, candidate.Name}, Separator)
	if val, err := rlp.EncodeToBytes(candidate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}

	// candidates
	candidates := []string{}
	pkey := strings.Join([]string{CandidateKeyPrefix, CandidatesKey}, Separator)
	if pval, err := db.Get(pkey); err != nil {
		return err
	} else if pval == nil {

	} else if err := rlp.DecodeBytes(pval, &candidates); err != nil {
		return err
	}

	for _, name := range candidates {
		if strings.Compare(name, candidate.Name) == 0 {
			return nil
		}
	}

	candidates = append(candidates, candidate.Name)
	npval, err := rlp.EncodeToBytes(candidates)
	if err != nil {
		return err
	}
	if err := db.Put(pkey, npval); err != nil {
		return err
	}

	skey := strings.Join([]string{CandidateKeyPrefix, CandidatesSizeKey}, Separator)
	return db.Put(skey, uint64tobytes(uint64(len(candidates))))
}

// DelCandidate del candidate info
func (db *LDB) DelCandidate(name string) error {
	key := strings.Join([]string{CandidateKeyPrefix, name}, Separator)
	if err := db.Delete(key); err != nil {
		return err
	}

	// candidates
	candidates := []string{}
	pkey := strings.Join([]string{CandidateKeyPrefix, CandidatesKey}, Separator)
	if pval, err := db.Get(pkey); err != nil {
		return err
	} else if pval == nil {

	} else if err := rlp.DecodeBytes(pval, &candidates); err != nil {
		return err
	}
	for index, prod := range candidates {
		if strings.Compare(prod, name) == 0 {
			candidates = append(candidates[:index], candidates[index+1:]...)
			break
		}
	}

	skey := strings.Join([]string{CandidateKeyPrefix, CandidatesSizeKey}, Separator)
	if err := db.Put(skey, uint64tobytes(uint64(len(candidates)))); err != nil {
		return err
	}

	if len(candidates) == 0 {
		return db.Delete(pkey)
	}
	npval, err := rlp.EncodeToBytes(candidates)
	if err != nil {
		return err
	}
	return db.Put(pkey, npval)
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
func (db *LDB) GetCandidates() ([]string, error) {
	// candidates
	pkey := strings.Join([]string{CandidateKeyPrefix, CandidatesKey}, Separator)
	candidates := []string{}
	if pval, err := db.Get(pkey); err != nil {
		return nil, err
	} else if pval == nil {

	} else if err := rlp.DecodeBytes(pval, &candidates); err != nil {
		return nil, err
	}
	return candidates, nil
}

// CandidatesSize candidate len
func (db *LDB) CandidatesSize() (uint64, error) {
	size := uint64(0)
	skey := strings.Join([]string{CandidateKeyPrefix, CandidatesSizeKey}, Separator)
	if sval, err := db.Get(skey); err != nil {
		return 0, err
	} else if sval != nil {
		size = bytestouint64(sval)
	}
	return size, nil
}

// SetAvailableQuantity set quantity
func (db *LDB) SetAvailableQuantity(epcho uint64, voter string, quantity *big.Int) error {
	key := strings.Join([]string{VoterQuantityPrefix, fmt.Sprintf("0x%x_%s", epcho, voter)}, Separator)
	if val, err := rlp.EncodeToBytes(quantity); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}
	return nil
}

// GetAvailableQuantity get quantity
func (db *LDB) GetAvailableQuantity(epcho uint64, voter string) (*big.Int, error) {
	key := strings.Join([]string{VoterQuantityPrefix, fmt.Sprintf("0x%x_%s", epcho, voter)}, Separator)
	quantity := big.NewInt(0)
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
	} else if err := rlp.DecodeBytes(val, quantity); err != nil {
		return nil, err
	}
	return quantity, nil
}

// SetVoter update voter info
func (db *LDB) SetVoter(voter *VoterInfo) error {
	key := strings.Join([]string{VoterKeyPrefix, voter.key()}, Separator)
	if val, err := rlp.EncodeToBytes(voter); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}

	// delegators
	delegators := []string{}
	ckey := strings.Join([]string{DelegatorsKeyPrfix, voter.ckey()}, Separator)

	if cval, err := db.Get(ckey); err != nil {
		return err
	} else if cval == nil {

	} else if err := rlp.DecodeBytes(cval, &delegators); err != nil {
		return err
	}
	hasc := false
	for _, name := range delegators {
		if strings.Compare(name, voter.Name) == 0 {
			hasc = true
			break
		}
	}
	if !hasc {
		ncval, err := rlp.EncodeToBytes(append(delegators, voter.Name))
		if err != nil {
			return err
		}
		if err := db.Put(ckey, ncval); err != nil {
			return err
		}
	}

	// candidates
	candidates := []string{}
	vkey := strings.Join([]string{VotersKeyPrefix, voter.vkey()}, Separator)

	if vval, err := db.Get(vkey); err != nil {
		return err
	} else if vval == nil {

	} else if err := rlp.DecodeBytes(vval, &candidates); err != nil {
		return err
	}
	for _, name := range candidates {
		if strings.Compare(name, voter.Candidate) == 0 {
			return
		}
	}
	nvval, err := rlp.EncodeToBytes(append(candidates, voter.Candidate))
	if err != nil {
		return err
	}
	return db.Put(vkey, nvval)
}

// DelVoter delete voter info
func (db *LDB) DelVoter(voter *VoterInfo) error {
	key := strings.Join([]string{VoterKeyPrefix, voter.key()}, Separator)
	if err := db.Delete(key); err != nil {
		return err
	}

	// delegators
	delegators := []string{}
	ckey := strings.Join([]string{DelegatorKeyPrfix, voter.ckey()}, Separator)
	if cval, err := db.Get(ckey); err != nil {
		return err
	} else if cval == nil {

	} else if err := rlp.DecodeBytes(cval, &delegators); err != nil {
		return err
	}
	for index, delegator := range delegators {
		if strings.Compare(delegator, voter.Name) == 0 {
			delegators = append(delegators[:index], delegators[index+1:]...)
			break
		}
	}
	if len(delegators) == 0 {
		if err := db.Delete(ckey); err != nil {
			return err
		}
	} else {
		ncval, err := rlp.EncodeToBytes(delegators)
		if err != nil {
			return err
		}
		if err := db.Put(ckey, ncval); err != nil {
			return err
		}
	}

	// candidates
	candidates := []string{}
	vkey := strings.Join([]string{VotersKeyPrefix, voter.vkey()}, Separator)

	if vval, err := db.Get(vkey); err != nil {
		return err
	} else if vval == nil {

	} else if err := rlp.DecodeBytes(vval, &candidates); err != nil {
		return err
	}
	for index, name := range candidates {
		if strings.Compare(name, voter.Candidate) == 0 {
			candidates = append(candidates[:index], candidates[index+1:]...)
			break
		}
	}
	if len(candidates) == 0 {
		if err := db.Delete(vkey); err != nil {
			return err
		}
	} else {
		nvval, err := rlp.EncodeToBytes(candidates)
		if err != nil {
			return err
		}
		if err := db.Put(vkey, nvval); err != nil {
			return err
		}
	}

	return nil
}

// DelVoters delete voters info by candidate
func (db *LDB) DelVoters(epcho uint64, candidate string) error {
	voters, err := db.GetVoters(epcho, candidate)
	if err != nil {
		return err
	}
	for _, voter := range voters {
		if err := db.GetVoter(epcho, voter, candidate); err != nil {
			return err
		}
		if err != db.DelVoter(voterInfo); err != nil {
			return err
		}
	}

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

// GetVoters voters info
func (db *LDB) GetVoters(epcho uint64, candidate string) ([]string, error) {
	// delegators
	vf := &VoterInfo{
		Epcho:     epcho,
		Candidate: candidate,
	}
	delegators := []string{}
	ckey := strings.Join([]string{DelegatorsKeyPrfix, vf.ckey()}, Separator)

	if cval, err := db.Get(ckey); err != nil {
		return nil, err
	} else if cval == nil {

	} else if err := rlp.DecodeBytes(cval, &delegators); err != nil {
		return nil, err
	}

	return delegators, nil
}

// GetVoterCandidates candidates info
func (db *LDB) GetVoterCandidates(epcho uint64, voter string) ([]string, error) {
	// candidates
	vf := &VoterInfo{
		Epcho: epcho,
		Name:  voter,
	}
	candidates := []string{}
	vkey := strings.Join([]string{VotersKeyPrefix, vf.vkey()}, Separator)

	if vval, err := db.Get(vkey); err != nil {
		return nil, err
	} else if vval == nil {

	} else if err := rlp.DecodeBytes(vval, &candidates); err != nil {
		return nil, err
	}

	return candidates, nil
}

// SetState set global state info
func (db *LDB) SetState(gstate *GlobalState) error {
	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(gstate.Epcho))}, Separator)
	if val, err := rlp.EncodeToBytes(gstate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}
	return nil
}

// GetState get state info
func (db *LDB) GetState(epcho uint64) (*GlobalState, error) {
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
	return candidateInfo.Quantity, candidateInfo.TotalQuantity, candidateInfo.Counter, nil
}

func uint64tobytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}

func bytestouint64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}
