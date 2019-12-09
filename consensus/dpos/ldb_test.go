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

	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
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
func (ldb *levelDB) IncAsset2Acct(string, string, *big.Int, uint64) (*types.Action, error) {
	return nil, nil
}
func (ldb *levelDB) GetSnapshot(string, uint64) ([]byte, error) {
	return nil, nil
}
func (ldb *levelDB) GetBalanceByTime(name string, timestamp uint64) (*big.Int, error) {
	return new(big.Int).Mul(big.NewInt(1000000000), DefaultConfig.decimals()), nil
}

func (ldb *levelDB) IsValidSign(name string, pubkey []byte) error {
	return nil
}

func newTestLDB() (*levelDB, func()) {
	dirname, err := ioutil.TempDir(os.TempDir(), "dpos_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := rawdb.NewLevelDBDatabase(dirname, 0, 0)
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

	// SetActivatedCandidate(uint64, *CandidateInfo) error
	// GetActivatedCandidate(uint64) (*CandidateInfo, error)

	// SetAvailableQuantity(uint64, string, *big.Int) error
	// GetAvailableQuantity(uint64, string) (*big.Int, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

	for index, candidate := range candidates {
		candidateInfo := &CandidateInfo{
			Epoch:         uint64(index),
			Name:          candidate,
			Info:          fmt.Sprintf("www.%v.com", candidate),
			Quantity:      big.NewInt(0),
			TotalQuantity: big.NewInt(0),
		}
		if err := db.SetCandidate(candidateInfo); err != nil {
			panic(fmt.Errorf("SetCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetCandidate(uint64(index), candidate); err != nil {
			panic(fmt.Errorf("GetCandidate --- %v", err))
		} else if !reflect.DeepEqual(candidateInfo, nCandidateInfo) {
			panic(fmt.Errorf("GetCandidate mismatch"))
		}
		if nCandidates, err := db.GetCandidates(uint64(index)); err != nil {
			panic(fmt.Errorf("GetCandidates --- %v", err))
		} else if len(nCandidates) != 1 {
			panic(fmt.Errorf("GetCandidates mismatch"))
		}
		if size, err := db.CandidatesSize(uint64(index)); err != nil {
			panic(fmt.Errorf("CandidatesSize --- %v", err))
		} else if size != 1 {
			panic(fmt.Errorf("CandidatesSize mismatch"))
		}

		if err := db.SetActivatedCandidate(uint64(index), candidateInfo); err != nil {
			panic(fmt.Errorf("SetActivatedCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetActivatedCandidate(uint64(index)); err != nil {
			panic(fmt.Errorf("GetActivatedCandidate --- %v", err))
		} else if !reflect.DeepEqual(candidateInfo, nCandidateInfo) {
			panic(fmt.Errorf("GetActivatedCandidate mismatch"))
		}

		if err := db.SetAvailableQuantity(uint64(index), candidate, big.NewInt(int64(index))); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if stake, err := db.GetAvailableQuantity(uint64(index), candidate); err != nil {
			panic(fmt.Errorf("GetAvailableQuantity --- %v", err))
		} else if stake.Cmp(big.NewInt(int64(index))) != 0 {
			panic(fmt.Errorf("GetAvailableQuantity mismatch"))
		}
	}

	for index, candidate := range candidates {
		candidateInfo, _ := db.GetCandidate(uint64(index), candidate)
		if err := db.SetCandidate(candidateInfo); err != nil {
			panic(fmt.Errorf("Redo SetCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetCandidate(uint64(index), candidate); err != nil {
			panic(fmt.Errorf("Redo GetCandidate --- %v", err))
		} else if !reflect.DeepEqual(candidateInfo, nCandidateInfo) {
			panic(fmt.Errorf("Redo GetCandidate mismatch"))
		}
		if nCandidates, err := db.GetCandidates(uint64(index)); err != nil {
			panic(fmt.Errorf("Redo GetCandidates --- %v", err))
		} else if len(nCandidates) != 1 {
			panic(fmt.Errorf("Redo GetCandidates mismatch"))
		}
		if size, err := db.CandidatesSize(uint64(index)); err != nil {
			panic(fmt.Errorf("Redo CandidatesSize --- %v", err))
		} else if size != 1 {
			panic(fmt.Errorf("Redo CandidatesSize mismatch"))
		}

		if err := db.SetActivatedCandidate(uint64(index), candidateInfo); err != nil {
			panic(fmt.Errorf("Redo SetActivatedCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetActivatedCandidate(uint64(index)); err != nil {
			panic(fmt.Errorf("Redo GetActivatedCandidate --- %v", err))
		} else if !reflect.DeepEqual(candidateInfo, nCandidateInfo) {
			panic(fmt.Errorf("Redo GetActivatedCandidate mismatch"))
		}

		if err := db.SetAvailableQuantity(uint64(index), candidate, big.NewInt(int64(index))); err != nil {
			panic(fmt.Errorf("Redo SetAvailableQuantity --- %v", err))
		}
		if stake, err := db.GetAvailableQuantity(uint64(index), candidate); err != nil {
			panic(fmt.Errorf("Redo GetAvailableQuantity --- %v", err))
		} else if stake.Cmp(big.NewInt(int64(index))) != 0 {
			panic(fmt.Errorf("Redo GetAvailableQuantity mismatch"))
		}
	}

	for index, candidate := range candidates {
		if err := db.DelCandidate(uint64(index), candidate); err != nil {
			panic(fmt.Errorf("DelCandidate --- %v", err))
		}
		if nCandidateInfo, err := db.GetCandidate(uint64(index), candidate); err != nil {
			panic(fmt.Errorf("Del GetCandidate --- %v", err))
		} else if nCandidateInfo != nil {
			panic(fmt.Errorf("Del GetCandidate mismatch"))
		}

		if nCandidates, err := db.GetCandidates(uint64(index)); err != nil {
			panic(fmt.Errorf("Del GetCandidates --- %v", err))
		} else if len(nCandidates) != 0 {
			panic(fmt.Errorf("Del GetCandidates mismatch"))
		}
		if size, err := db.CandidatesSize(uint64(index)); err != nil {
			panic(fmt.Errorf("Del CandidatesSize --- %v", err))
		} else if size != 0 {
			panic(fmt.Errorf("Del CandidatesSize mismatch"))
		}

		if nCandidateInfo, err := db.GetActivatedCandidate(uint64(index)); err != nil {
			panic(fmt.Errorf("Del GetActivatedCandidate --- %v", err))
		} else if nCandidateInfo.Name != candidate {
			panic(fmt.Errorf("Del GetActivatedCandidate mismatch"))
		}

		if stake, err := db.GetAvailableQuantity(uint64(index), candidate); err != nil {
			panic(fmt.Errorf("Del GetAvailableQuantity --- %v", err))
		} else if stake.Cmp(big.NewInt(int64(index))) != 0 {
			panic(fmt.Errorf("Del GetAvailableQuantity mismatch"))
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
		if err := db.SetAvailableQuantity(uint64(index), voter, big.NewInt(int64(index))); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}

		if quantity, err := db.GetAvailableQuantity(uint64(index), voter); err != nil {
			panic(fmt.Errorf("GetAvailableQuantity --- %v", err))
		} else if quantity.Cmp(big.NewInt(int64(index))) != 0 {
			panic(fmt.Errorf("GetAvailableQuantity mismatch"))
		}
	}

	for index, voter := range voters {
		if err := db.SetAvailableQuantity(uint64(index), voter, big.NewInt(int64(index+1))); err != nil {
			panic(fmt.Errorf("Redo SetAvailableQuantity --- %v", err))
		}

		if quantity, err := db.GetAvailableQuantity(uint64(index), voter); err != nil {
			panic(fmt.Errorf("Redo GetAvailableQuantity --- %v", err))
		} else if quantity.Cmp(big.NewInt(int64(index+1))) != 0 {
			panic(fmt.Errorf("Redo GetAvailableQuantity mismatch"))
		}
	}
}

func TestLDBVoter(t *testing.T) {
	// SetVoter(*VoterInfo) error
	// GetVoter(uint64, string, string) (*VoterInfo, error)
	// GetVotersByVoter(uint64, string) ([]*VoterInfo, error)
	// GetVotersByCandidate(uint64, string) ([]*VoterInfo, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

	for _, candidate := range candidates {
		candidateInfo := &CandidateInfo{
			Epoch:         0,
			Name:          candidate,
			Info:          fmt.Sprintf("www.%v.com", candidate),
			Quantity:      big.NewInt(0),
			TotalQuantity: big.NewInt(0),
		}
		if err := db.SetCandidate(candidateInfo); err != nil {
			panic(fmt.Errorf("SetCandidate --- %v", err))
		}
	}

	epoch := uint64(0)
	for index, voter := range voters {
		voterInfo := &VoterInfo{
			Epoch:     epoch,
			Name:      voter,
			Candidate: candidates[0],
			Quantity:  big.NewInt(int64(index)),
			Number:    uint64(index),
		}
		if err := db.SetAvailableQuantity(epoch, voter, big.NewInt(int64(index))); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := db.SetVoter(voterInfo); err != nil {
			panic(fmt.Errorf("SetVoter --- %v", err))
		}
		if nvoterInfo, err := db.GetVoter(epoch, voter, candidates[0]); err != nil {
			panic(fmt.Errorf("GetVoter --- %v", err))
		} else if !reflect.DeepEqual(voterInfo, nvoterInfo) {
			panic(fmt.Errorf("GetVoter mismatch"))
		}
		if nCandidates, err := db.GetVotersByVoter(epoch, voter); err != nil {
			panic(fmt.Errorf("GetVotersByVoter --- %v", err))
		} else if len(nCandidates) != 1 {
			panic(fmt.Errorf("GetVotersByVoter mismatch %v %v", len(nCandidates), index+1))
		}
		if nVoters, err := db.GetVotersByCandidate(epoch, candidates[0]); err != nil {
			panic(fmt.Errorf("GetVotersByCandidate --- %v", err))
		} else if len(nVoters) != index+1 {
			panic(fmt.Errorf("GetVotersByCandidate mismatch"))
		}
	}

	epoch++
	for index, candidate := range candidates {
		voterInfo := &VoterInfo{
			Epoch:     epoch,
			Name:      voters[0],
			Candidate: candidate,
			Quantity:  big.NewInt(int64(index)),
			Number:    uint64(index),
		}
		if err := db.SetAvailableQuantity(epoch, voters[0], big.NewInt(int64(index))); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := db.SetVoter(voterInfo); err != nil {
			panic(fmt.Errorf("SetVoter --- %v", err))
		}
		if nvoterInfo, err := db.GetVoter(epoch, voters[0], candidate); err != nil {
			panic(fmt.Errorf("GetVoter --- %v", err))
		} else if !reflect.DeepEqual(voterInfo, nvoterInfo) {
			panic(fmt.Errorf("GetVoter mismatch"))
		}
		if nCandidates, err := db.GetVotersByVoter(epoch, voters[0]); err != nil {
			panic(fmt.Errorf("GetVotersByVoter --- %v", err))
		} else if len(nCandidates) != index+1 {
			panic(fmt.Errorf("GetVotersByVoter mismatch %v %v", len(nCandidates), index+1))
		}
		if nVoters, err := db.GetVotersByCandidate(epoch, candidate); err != nil {
			panic(fmt.Errorf("GetVotersByCandidate --- %v", err))
		} else if len(nVoters) != 1 {
			panic(fmt.Errorf("GetVotersByCandidate mismatch"))
		}
	}

	epoch++
	for index, candidate := range candidates {
		voterInfo := &VoterInfo{
			Epoch:     epoch,
			Name:      voters[index],
			Candidate: candidate,
			Quantity:  big.NewInt(int64(index)),
			Number:    uint64(index),
		}
		if err := db.SetAvailableQuantity(epoch, voters[index], big.NewInt(int64(index))); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := db.SetVoter(voterInfo); err != nil {
			panic(fmt.Errorf("SetVoter --- %v", err))
		}
		if nvoterInfo, err := db.GetVoter(epoch, voters[index], candidate); err != nil {
			panic(fmt.Errorf("GetVoter --- %v", err))
		} else if !reflect.DeepEqual(voterInfo, nvoterInfo) {
			panic(fmt.Errorf("GetVoter mismatch"))
		}
		if nCandidates, err := db.GetVotersByVoter(epoch, voters[index]); err != nil {
			panic(fmt.Errorf("GetVotersByVoter --- %v", err))
		} else if len(nCandidates) != 1 {
			panic(fmt.Errorf("GetVotersByVoter mismatch"))
		}
		if nVoters, err := db.GetVotersByCandidate(epoch, candidate); err != nil {
			panic(fmt.Errorf("GetVotersByCandidate --- %v", err))
		} else if len(nVoters) != 1 {
			panic(fmt.Errorf("GetVotersByCandidate mismatch"))
		}
	}

}

func TestLDBGlobalState(t *testing.T) {
	// SetState(*GlobalState) error
	// GetState(uint64) (*GlobalState, error)
	// SetLastestEpoch(uint64) error
	// GetLastestEpoch() (uint64, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

	for index := range candidates {
		gstate := &GlobalState{
			Epoch:                       uint64(index + 1),
			PreEpoch:                    uint64(index),
			ActivatedTotalQuantity:      big.NewInt(0),
			ActivatedCandidateSchedule:  candidates[index:],
			UsingCandidateIndexSchedule: []uint64{},
			BadCandidateIndexSchedule:   []uint64{},
			TotalQuantity:               big.NewInt(0),
		}
		if err := db.SetState(gstate); err != nil {
			panic(fmt.Errorf("SetState --- %v", err))
		}
		if ngstate, err := db.GetState(uint64(index + 1)); err != nil {
			panic(fmt.Errorf("GetState --- %v", err))
		} else if !reflect.DeepEqual(gstate, ngstate) {
			panic(fmt.Errorf("GetState mismatch %v %v", gstate, ngstate))
		}
		if err := db.SetLastestEpoch(gstate.Epoch); err != nil {
			panic(fmt.Errorf("GetLastestEpoch --- %v", err))
		} else if epoch, err := db.GetLastestEpoch(); err != nil {
			panic(fmt.Errorf("GetLastestEpoch --- %v", err))
		} else if epoch != uint64(index+1) {
			panic(fmt.Errorf("GetLastestEpoch mismatch"))
		}
	}

	for index := range candidates {
		gstate := &GlobalState{
			Epoch:                       uint64(index + 1),
			PreEpoch:                    uint64(index),
			ActivatedTotalQuantity:      big.NewInt(0),
			ActivatedCandidateSchedule:  candidates[index:],
			TotalQuantity:               big.NewInt(0),
			UsingCandidateIndexSchedule: []uint64{},
			BadCandidateIndexSchedule:   []uint64{},
		}
		if err := db.SetState(gstate); err != nil {
			panic(fmt.Errorf("Redo SetState --- %v", err))
		}
		if ngstate, err := db.GetState(uint64(index + 1)); err != nil {
			panic(fmt.Errorf("Redo GetState --- %v", err))
		} else if !reflect.DeepEqual(gstate, ngstate) {
			panic(fmt.Errorf("Redo GetState mismatch"))
		}
		if err := db.SetLastestEpoch(gstate.Epoch); err != nil {
			panic(fmt.Errorf("GetLastestEpoch --- %v", err))
		} else if epoch, err := db.GetLastestEpoch(); err != nil {
			panic(fmt.Errorf("GetLastestEpoch --- %v", err))
		} else if epoch != uint64(index+1) {
			panic(fmt.Errorf("GetLastestEpoch mismatch"))
		}
	}
}

func TestLDBTakeOver(t *testing.T) {
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

	epoch := uint64(2)
	if err := db.SetTakeOver(epoch); err != nil {
		panic(fmt.Errorf("SetTakeOver --- %v", err))
	} else if tepoch, err := db.GetTakeOver(); err != nil {
		panic(fmt.Errorf("GetTakeOver --- %v", err))
	} else if tepoch != epoch {
		panic(fmt.Errorf("GetTakeOver mismatch"))
	}
	// return 0 when not set
	if zepoch, err := db.GetTakeOver(); err != nil {
		panic(fmt.Errorf("Zero GetTakeOver --- %v", err))
	} else if zepoch != epoch {
		panic(fmt.Errorf("Zero GetTakeOver mismatch"))
	}

}
