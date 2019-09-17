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

package rawdb

import (
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/utils/fdb/leveldb"
	"github.com/fractalplatform/fractal/utils/fdb/memdb"
)

// NewMemoryDatabase creates an ephemeral in-memory key-value database .
func NewMemoryDatabase() fdb.Database {
	return memdb.NewMemDatabase()

}

// NewMemoryDatabaseWithCap creates an ephemeral in-memory key-value database
// with an initial starting capacity.
func NewMemoryDatabaseWithCap(size int) fdb.Database {
	return memdb.NewMemDatabaseWithCap(size)
}

// NewLevelDBDatabase creates a persistent key-value database.
func NewLevelDBDatabase(file string, cache int, handles int) (fdb.Database, error) {
	return leveldb.NewLDBDatabase(file, cache, handles)
}
