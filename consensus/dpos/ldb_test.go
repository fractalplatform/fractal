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
	"io/ioutil"
	"math/big"
	"os"
	"testing"

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
	db, err := ldb.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return &levelDB{Database: db}, func() {
		db.Close()
		os.RemoveAll(dirname)
	}
}
func TestLDB(t *testing.T) {
	// SetVoter(*voterInfo) error
	// DelVoter(string, string) error
	// GetVoter(string) (*voterInfo, error)
	// GetDelegators(string) ([]string, error)
	ldb, function := newTestLDB()
	db, _ := NewLDB(ldb)
	defer function()

}
