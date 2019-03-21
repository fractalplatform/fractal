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

package leveldb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/fractalplatform/fractal/utils/fdb"
)

func newTestLDB() (*LDBDatabase, func()) {
	dirname, err := ioutil.TempDir(os.TempDir(), "leveldb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dirname)
	}
}

func TestLDB_PutGet(t *testing.T) {
	db, remove := newTestLDB()
	defer remove()
	fdb.TestPutGet(db, t)
}

func TestLDB_ParallelPutGet(t *testing.T) {
	db, remove := newTestLDB()
	defer remove()
	fdb.TestParallelPutGet(db, t)
}
