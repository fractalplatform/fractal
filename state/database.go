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

package state

import (
	"sync"

	"github.com/fractalplatform/fractal/common"
	trie "github.com/fractalplatform/fractal/state/mtp"
	"github.com/fractalplatform/fractal/utils/fdb"
)

var MaxTrieCacheGen = uint16(120)

//Database cache db exported
type Database interface {
	GetDB() fdb.Database
	OpenTrie(root common.Hash) (Trie, error)
	TrieDB() *trie.Database
	Lock()
	UnLock()
}

type Trie interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	Commit(onleaf trie.LeafCallback) (common.Hash, error)
	Hash() common.Hash
	NodeIterator(startKey []byte) trie.NodeIterator
	GetKey([]byte) []byte // TODO(fjl): remove this when SecureTrie is removed
	Prove(key []byte, fromLevel uint, proofDb fdb.Putter) error
}

// NewDatabase creates a backing store for state.
func NewDatabase(db fdb.Database) Database {
	return &cachingDB{
		db:     db,
		triedb: trie.NewDatabase(db),
	}
}

type cachingDB struct {
	db     fdb.Database
	lock   sync.Mutex
	triedb *trie.Database
}

// cachedTrie inserts its trie into a cachingDB on commit.
type cachedTrie struct {
	*trie.SecureTrie
	db *cachingDB
}

func (db *cachingDB) OpenTrie(root common.Hash) (Trie, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	tr, err := trie.NewSecure(root, db.triedb, MaxTrieCacheGen)
	if err != nil {
		return nil, err
	}

	return cachedTrie{tr, db}, nil
}

// TrieDB retrieves any intermediate trie-node caching layer.
func (db *cachingDB) TrieDB() *trie.Database {
	return db.triedb
}

func (db *cachingDB) Lock() {
	db.lock.Lock()
}

func (db *cachingDB) UnLock() {
	db.lock.Unlock()
}

func (db *cachingDB) GetDB() fdb.Database {
	return db.db
}
