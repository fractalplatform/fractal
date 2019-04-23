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
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	ldb "github.com/fractalplatform/fractal/utils/fdb/leveldb"
)

type levelDB struct {
	fdb.Database
}

func (ldb *levelDB) Has(key string) (bool, error) {
	return ldb.Database.Has([]byte(key))
}
func (ldb *levelDB) Get(key string) ([]byte, error) {
	if has, err := ldb.Database.Has([]byte(key)); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return ldb.Database.Get([]byte(key))
}
func (ldb *levelDB) Put(key string, value []byte) error {
	return ldb.Database.Put([]byte(key), value)
}
func (ldb *levelDB) Delete(key string) error {
	if has, err := ldb.Database.Has([]byte(key)); err != nil {
		return err
	} else if !has {
		return nil
	}
	return ldb.Database.Delete([]byte(key))
}
func (ldb *levelDB) Delegate(string, *big.Int) error {
	return nil
}
func (ldb *levelDB) Undelegate(string, *big.Int) (*types.Action, error) {
	return nil, nil
}
func (ldb *levelDB) IncAsset2Acct(string, string, *big.Int) (*types.Action, error) {
	return nil, nil
}
func (ldb *levelDB) GetSnapshot(string, uint64) ([]byte, error) {
	return nil, nil
}
func newTestLDB() (*levelDB, func()) {
	dirname, err := ioutil.TempDir(os.TempDir(), "dpos_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := ldb.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return &levelDB{Database: db}, func() {
		db.Close()
		os.RemoveAll(dirname)
	}
}

var (
	candidates = []string{
		"candidate1",
		"candidate2",
		"candidate3",
	}
	voters = []string{
		"voter1",
		"voter2",
		"voter3",
	}
)

func TestLDBCandidate(t *testing.T) {
	// SetCandidate(*CandidateInfo) error
	// DelCandidate(string) error
	// GetCandidate(string) (*CandidateInfo, error)
	// GetCandidates() ([]string, error)
	// CandidatesSize() (uint64, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

	for index, candidate := range candidates {
		candidateInfo := &CandidateInfo{
			Name: candidate,
			URL:  fmt.Sprintf("www.%v.com", candidate),
		}
		if err := db.SetCandidate(candidateInfo); err != nil {
			panic(fmt.Errorf("SetCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetCandidate(candidate); err != nil {
			panic(fmt.Errorf("GetCandidate --- %v", err))
		} else if !reflect.DeepEqual(candidateInfo, nCandidateInfo) {
			panic(fmt.Errorf("GetCandidate mismatch"))
		}
		if nCandidates, err := db.GetCandidates(); err != nil {
			panic(fmt.Errorf("GetCandidates --- %v", err))
		} else if len(nCandidates) != index+1 {
			panic(fmt.Errorf("GetCandidates mismatch"))
		}
		if size, err := db.CandidatesSize(); err != nil {
			panic(fmt.Errorf("CandidatesSize --- %v", err))
		} else if size != index+1 {
			panic(fmt.Errorf("CandidatesSize mismatch"))
		}
	}

	for index, candidate := range candidates {
		candidateInfo := &CandidateInfo{
			Name: candidate,
			URL:  fmt.Sprintf("www.%v.com", candidate),
		}
		if err := db.SetCandidate(candidateInfo); err != nil {
			panic(fmt.Errorf("Redo SetCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetCandidate(candidate); err != nil {
			panic(fmt.Errorf("Redo GetCandidate --- %v", err))
		} else if !reflect.DeepEqual(candidateInfo, nCandidateInfo) {
			panic(fmt.Errorf("Redo GetCandidate mismatch"))
		}
		if nCandidates, err := db.GetCandidates(); err != nil {
			panic(fmt.Errorf("Redo GetCandidates --- %v", err))
		} else if !reflect.DeepEqual(candidates, nCandidates) {
			panic(fmt.Errorf("Redo GetCandidates mismatch"))
		}
		if size, err := db.CandidatesSize(); err != nil {
			panic(fmt.Errorf("Redo CandidatesSize --- %v", err))
		} else if size != len(candidates) {
			panic(fmt.Errorf("Redo CandidatesSize mismatch"))
		}
	}

	for index, candidate := range candidates {
		if err := db.DelCandidate(candidate); err != nil {
			panic(fmt.Errorf("DelCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetCandidate(candidate); err != nil {
			panic(fmt.Errorf("Del GetCandidate --- %v", err))
		} else if nCandidateInfo != nil {
			panic(fmt.Errorf("Del GetCandidate mismatch"))
		}

		if nCandidates, err := db.GetCandidates(); err != nil {
			panic(fmt.Errorf("Del GetCandidates --- %v", err))
		} else if len(nCandidates) != len(candidates)-index+1 {
			panic(fmt.Errorf("Del GetCandidates mismatch"))
		}
		if size, err := db.CandidatesSize(); err != nil {
			panic(fmt.Errorf("Del CandidatesSize --- %v", err))
		} else if size != len(candidates)-index+1 {
			panic(fmt.Errorf("Del CandidatesSize mismatch"))
		}
	}
}

func TestLDBAvailableQuantity(t *testing.T) {
	// SetAvailableQuantity(uint64, string, *big.Int) error
	// GetAvailableQuantity(uint64, string) (*big.Int, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

	for index, voter := range voters {
		if err := db.SetAvailableQuantity(index, voter, big.NewInt(index+1)); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}

		if quantity, err := db.GetAvailableQuantity(index, voter); err != nil {
			panic(fmt.Errorf("GetAvailableQuantity --- %v", err))
		} else if quantity.Cmp(big.NewInt(index+1)) != 0 {
			panic(fmt.Errorf("GetAvailableQuantity mismatch"))
		}
	}

	for index, voter := range voters {
		if err := db.SetAvailableQuantity(index, voter, big.NewInt(index+2)); err != nil {
			panic(fmt.Errorf("Redo SetAvailableQuantity --- %v", err))
		}

		if quantity, err := db.GetAvailableQuantity(index, voter); err != nil {
			panic(fmt.Errorf("Redo GetAvailableQuantity --- %v", err))
		} else if quantity.Cmp(big.NewInt(index+2)) != 0 {
			panic(fmt.Errorf("Redo GetAvailableQuantity mismatch"))
		}
	}
}

func TestLDBVoter(t *testing.T) {
	// SetVoter(*VoterInfo) error
	// DelVoter(*VoterInfo) error
	// DelVoters(uint64, string) error
	// GetVoter(uint64, string, string) (*VoterInfo, error)
	// GetVoters(uint64, string) ([]string, error)
	// GetVoterCandidates(uint64, string) ([]string, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

}

func TestLDBGlobalState(t *testing.T) {
	// SetState(*GlobalState) error
	// GetState(uint64) (*GlobalState, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

	for index, candidate := range candidates {
		gstate := &GlobalState{
			Epcho:                      index,
			ActivatedCandidateSchedule: candidates[index:],
		}
		if err := db.SetState(gstate); err != nil {
			panic(fmt.Errorf("SetState --- %v", err))
		}
		if ngstate, err := db.GetState(index); err != nil {
			panic(fmt.Errorf("GetState --- %v", err))
		} else if !reflect.DeepEqual(gstate, ngstate) {
			panic(fmt.Errorf("GetState mismatch"))
		}
	}

	for index, candidate := range candidates {
		gstate := &GlobalState{
			Epcho:                      index,
			ActivatedCandidateSchedule: candidates[index:],
		}
		if err := db.SetState(gstate); err != nil {
			panic(fmt.Errorf("Redo SetState --- %v", err))
		}
		if ngstate, err := db.GetState(index); err != nil {
			panic(fmt.Errorf("Redo GetState --- %v", err))
		} else if !reflect.DeepEqual(gstate, ngstate) {
			panic(fmt.Errorf("Redo GetState mismatch"))
		}
	}
}
