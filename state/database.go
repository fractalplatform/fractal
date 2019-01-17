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
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/hashicorp/golang-lru"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

const (
	kvCacheSize = 100000
)

//Database cache db exported
type Database interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
	PutCache(key string, value []byte) error
	DeleteCache(key string) error
	GetDB() fdb.Database
	GetHash() common.Hash
	SetHash(hash common.Hash)
	Lock()
	UnLock()
	RLock()
	RUnLock()
	Purge()
}

// NewDatabase creates a backing store for state.
func NewDatabase(db fdb.Database) Database {
	kvch, _ := lru.New(kvCacheSize)
	//get cache hash from db
	curHash := rawdb.ReadOptBlockHash(db)

	return &cachingDB{db: db,
		kvCache: kvch,
		hash:    curHash}
}

type cachingDB struct {
	db      fdb.Database
	lock    sync.RWMutex
	kvCache *lru.Cache
	hash    common.Hash
}

func (db *cachingDB) Lock() {
	db.lock.Lock()
}

func (db *cachingDB) UnLock() {
	db.lock.Unlock()
}

func (db *cachingDB) RLock() {
	db.lock.RLock()
}

func (db *cachingDB) RUnLock() {
	db.lock.RUnlock()
}

func (db *cachingDB) GetDB() fdb.Database {
	return db.db
}

func (db *cachingDB) GetHash() common.Hash {
	return db.hash
}

func (db *cachingDB) SetHash(hash common.Hash) {
	db.hash = hash
}

func (db *cachingDB) Purge() {
	db.kvCache.Purge()
}

func (db *cachingDB) Get(key string) ([]byte, error) {
	if cached, ok := db.kvCache.Get(key); ok {
		return cached.([]byte), nil
	}

	value, err := db.db.Get([]byte(key))
	if err != nil {
		if err != errors.ErrNotFound && err != fdb.ErrNotFound {
			return nil, err
		}
		//not found return nil
	}

	db.kvCache.Add(key, common.CopyBytes(value))

	return value, nil
}

func (db *cachingDB) Put(key string, value []byte) error {
	err := db.db.Put([]byte(key), value)
	if err != nil {
		return err
	}
	db.kvCache.Add(key, common.CopyBytes(value))

	return nil
}

//only put value to cache
func (db *cachingDB) PutCache(key string, value []byte) error {
	db.kvCache.Add(key, common.CopyBytes(value))

	return nil
}

func (db *cachingDB) Delete(key string) error {
	db.kvCache.Remove(key)
	return db.db.Delete([]byte(key))
}

func (db *cachingDB) DeleteCache(key string) error {
	db.kvCache.Remove(key)
	return nil
}
