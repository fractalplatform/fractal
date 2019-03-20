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
	"testing"
	"time"

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
func (ldb *levelDB) Undelegate(string, *big.Int) error {
	return nil
}
func (ldb *levelDB) IncAsset2Acct(string, string, *big.Int) error {
	return nil
}
func (ldb *levelDB) GetSnapshot(string, uint64) ([]byte, error) {
	return nil, nil
}
func newTestLDB() (*levelDB, func()) {
	dirname, err := ioutil.TempDir(os.TempDir(), "dpos_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := fdb.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return &levelDB{Database: db}, func() {
		db.Close()
		os.RemoveAll(dirname)
	}
}
func TestLDBVoter(t *testing.T) {
	// SetVoter(*voterInfo) error
	// DelVoter(string, string) error
	// GetVoter(string) (*voterInfo, error)
	// GetDelegators(string) ([]string, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()
	voter := &voterInfo{
		Name:     "vos",
		Cadidate: "fos",
		Quantity: big.NewInt(1000),
		Height:   uint64(time.Now().UnixNano()),
	}
	if err := db.SetVoter(voter); err != nil {
		panic(fmt.Errorf("setvoter --- %v", err))
	}

	nvoter, err := db.GetVoter(voter.Name)
	if err != nil {
		panic(fmt.Errorf("getvoter --- %v", err))
	} else if nvoter == nil {
		panic(fmt.Errorf("getvoter --- not nil"))
	}

	// if delegators, err := db.GetDelegators(nvoter.Cadidate); err != nil {
	// 	panic(fmt.Errorf("getdelegators --- %v", err))
	// } else if len(delegators) != 1 {
	// 	panic(fmt.Errorf("getdelegators --- not mismatch"))
	// }

	if err := db.DelVoter(nvoter.Name, nvoter.Cadidate); err != nil {
		panic(fmt.Errorf("delvoter --- %v", err))
	}

	// if delegators, err := db.GetDelegators(nvoter.Cadidate); err != nil {
	// 	panic(fmt.Errorf("getdelegators after del --- %v", err))
	// } else if len(delegators) != 0 {
	// 	t.Log(len(delegators))
	// 	panic(fmt.Errorf("getdelegators after del --- not mismatch"))
	// }

	if nvoter, err := db.GetVoter(voter.Name); err != nil {
		panic(fmt.Errorf("getvoter after del --- %v", err))
	} else if nvoter != nil {
		panic(fmt.Errorf("getvoter after del --- should nil"))
	}
}

func TestLDBCadidate(t *testing.T) {
	// SetCadidate(*cadidateInfo) error
	// DelCadidate(string) error
	// GetCadidate(string) (*cadidateInfo, error)
	// Cadidates() ([]*cadidateInfo, error)
	// CadidatesSize() (uint64, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()
	prod := &cadidateInfo{
		Name:          "fos",
		URL:           "www.fractalproject.com",
		Quantity:      big.NewInt(1000),
		TotalQuantity: big.NewInt(1000),
		Height:        uint64(time.Now().UnixNano()),
	}
	if err := db.SetCadidate(prod); err != nil {
		panic(fmt.Errorf("setprod --- %v", err))
	}

	nprod, err := db.GetCadidate(prod.Name)
	if err != nil {
		panic(fmt.Errorf("getprod --- %v", err))
	} else if nprod == nil {
		panic(fmt.Errorf("getprod --- not nil"))
	}

	if size, err := db.CadidatesSize(); err != nil {
		panic(fmt.Errorf("prodsize --- %v", err))
	} else if size != 1 {
		panic(fmt.Errorf("prodsize --- mismatch"))
	}

	if err := db.DelCadidate(nprod.Name); err != nil {
		panic(fmt.Errorf("delprod --- %v", err))
	}

	if size, err := db.CadidatesSize(); err != nil {
		panic(fmt.Errorf("prodsize --- %v", err))
	} else if size != 0 {
		panic(fmt.Errorf("prodsize --- mismatch"))
	}

	if nprod, err := db.GetCadidate(nprod.Name); err != nil {
		panic(fmt.Errorf("getprod --- %v", err))
	} else if nprod != nil {
		panic(fmt.Errorf("getprod --- should nil"))
	}
}

func TestLDBState(t *testing.T) {
	// SetState(*globalState) error
	// DelState(uint64) error
	// GetState(uint64) (*globalState, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()
	gstate := &globalState{
		Height:                          10,
		ActivatedCadidateScheduleUpdate: uint64(time.Now().UnixNano()),
		ActivatedCadidateSchedule:       []string{},
		ActivatedTotalQuantity:          big.NewInt(1000),
		TotalQuantity:                   big.NewInt(100000),
	}

	if err := db.SetState(gstate); err != nil {
		panic(fmt.Errorf("setstate --- %v", err))
	}

	ngstate, err := db.GetState(gstate.Height)
	if err != nil {
		panic(fmt.Errorf("getstate --- %v", err))
	} else if ngstate == nil {
		panic(fmt.Errorf("getstate --- not nil"))
	}

	if err := db.DelState(ngstate.Height); err != nil {
		panic(fmt.Errorf("delstate --- %v", err))
	}

	if ngstate, err := db.GetState(ngstate.Height); err != nil {
		panic(fmt.Errorf("getstate --- %v", err))
	} else if ngstate != nil {
		panic(fmt.Errorf("getstate --- should nil"))
	}

}
