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
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
)

type revision struct {
	id           int
	journalIndex int
}

var (
	statePrefix    = "ST"
	acctDataPrefix = "AD"
	linkSymbol     = "*"
)

// StateDB store block operate info
type StateDB struct {
	db   Database
	trie Trie

	readSet  map[string][]byte   // save old/unmodified data
	writeSet map[string][]byte   // last modify data
	dirtySet map[string]struct{} // writeSet which key is modified

	dbErr  error
	refund uint64 // unuse gas

	thash, bhash common.Hash // current transaction hash and current block hash
	txIndex      int         // transaction index in block

	logs    map[common.Hash][]*types.Log
	logSize uint

	preimages map[common.Hash][]byte

	journal        *journal
	validRevisions []revision
	nextRevisionID int

	stateTrace bool // replay transaction, true is replayed , false is not replayed

	lock sync.Mutex
}

//New func generate a statedb object
//parentHash: block's parent hash, db: cachedb
func New(root common.Hash, db Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}

	return &StateDB{
		db:         db,
		trie:       tr,
		readSet:    make(map[string][]byte),
		writeSet:   make(map[string][]byte),
		dirtySet:   make(map[string]struct{}),
		logs:       make(map[common.Hash][]*types.Log),
		preimages:  make(map[common.Hash][]byte),
		journal:    newJournal(),
		stateTrace: false}, nil
}

//setError only save first err
func (s *StateDB) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

// Error return error info
func (s *StateDB) Error() error {
	return s.dbErr
}

// Reset clear StateDB
func (s *StateDB) Reset(root common.Hash) error {
	tr, err := s.db.OpenTrie(root)
	if err != nil {
		return err
	}

	s.trie = tr
	s.readSet = make(map[string][]byte)
	s.writeSet = make(map[string][]byte)
	s.dirtySet = make(map[string]struct{})
	s.thash = common.Hash{}
	s.bhash = common.Hash{}
	s.txIndex = 0
	s.logs = make(map[common.Hash][]*types.Log)
	s.logSize = 0
	s.preimages = make(map[common.Hash][]byte)
	s.dbErr = nil
	s.clearJournalAndRefund()
	return nil
}

// AddLog save transaction log
func (s *StateDB) AddLog(log *types.Log) {
	s.journal.append(addLogChange{txhash: s.thash})

	log.TxHash = s.thash
	log.BlockHash = s.bhash
	log.TxIndex = uint(s.txIndex)
	log.Index = s.logSize
	s.logs[s.thash] = append(s.logs[s.thash], log)
	s.logSize++
}

// GetLogs get a strip of transaction log
func (s *StateDB) GetLogs(hash common.Hash) []*types.Log {
	return s.logs[hash]
}

// Logs get all transaction log
func (s *StateDB) Logs() []*types.Log {
	var logs []*types.Log
	for _, lgs := range s.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// AddPreimage hash is preimageHash
func (s *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	if _, ok := s.preimages[hash]; !ok {
		s.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		s.preimages[hash] = pi
	}
}

// Preimages return preimages
func (s *StateDB) Preimages() map[common.Hash][]byte {
	return s.preimages
}

// save unuse gas
func (s *StateDB) AddRefund(gas uint64) {
	s.journal.append(refundChange{prev: s.refund})
	s.refund += gas
}

func (s *StateDB) GetRefund() uint64 {
	return s.refund
}

func (s *StateDB) GetState(account string, key common.Hash) common.Hash {
	optKey := statePrefix + linkSymbol + account + linkSymbol + key.String()
	value, _ := s.get(optKey)
	if (value == nil) || (len(value) != common.HashLength) {
		return common.Hash{}
	}
	return common.BytesToHash(value)
}

// set contract variable key value
func (s *StateDB) SetState(account string, key, value common.Hash) {
	optKey := statePrefix + linkSymbol + account + linkSymbol + key.String()
	s.put(optKey, value[:])
}

// set writeSet
func (s *StateDB) set(key string, value []byte) {
	if value == nil {
		s.writeSet[key] = nil
	} else {
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)
		s.writeSet[key] = valueCopy
	}
}

func (s *StateDB) put(key string, value []byte) {
	oldValue, _ := s.get(key)
	s.journal.append(stateChange{key: &key,
		prevalue: oldValue})
	s.set(key, value)
}

//get return nil when key not exsit
func (s *StateDB) get(key string) ([]byte, error) {
	if value, exsit := s.writeSet[key]; exsit {
		return common.CopyBytes(value), nil
	}

	// replay transaction
	if s.stateTrace {
		errInfo := fmt.Sprintf("No value when trace.")
		err := errors.New(errInfo)
		s.setError(err)
		return nil, err
	}

	value, err := s.trie.TryGet([]byte(key))
	if len(value) == 0 {
		s.setError(err)
		return nil, err
	}

	s.readSet[key] = common.CopyBytes(value)
	s.writeSet[key] = common.CopyBytes(value)

	return common.CopyBytes(value), nil
}

//RpcGetState provide get value of the key to rpc
//when called please RLock cachedb
func (s *StateDB) RpcGetState(account string, key common.Hash) common.Hash {
	optKey := statePrefix + linkSymbol + account + linkSymbol + key.String()

	value, err := s.trie.TryGet([]byte(optKey))
	if len(value) == 0 {
		s.setError(err)
		return common.Hash{}
	}

	return common.BytesToHash(value)
}

//RpcGet provide get value of the key to rpc
//when called please RLock cachedb
func (s *StateDB) RpcGet(account string, key string) ([]byte, error) {
	optKey := acctDataPrefix + linkSymbol + account + linkSymbol + key

	value, err := s.trie.TryGet([]byte(optKey))
	if len(value) == 0 {
		s.setError(err)
		return nil, err
	}

	return common.CopyBytes(value), nil
}

func (s *StateDB) Database() Database {
	return s.db
}

func (s *StateDB) Copy() *StateDB {
	s.lock.Lock()
	defer s.lock.Unlock()

	state := &StateDB{
		db:        s.db,
		trie:      s.trie,
		readSet:   make(map[string][]byte, len(s.writeSet)),
		writeSet:  make(map[string][]byte, len(s.writeSet)),
		dirtySet:  make(map[string]struct{}, len(s.dirtySet)),
		refund:    s.refund,
		logs:      make(map[common.Hash][]*types.Log, len(s.logs)),
		logSize:   s.logSize,
		preimages: make(map[common.Hash][]byte),
		journal:   newJournal()}

	for key := range s.journal.dirties {
		value := s.writeSet[key]
		state.readSet[key] = common.CopyBytes(value)
		state.writeSet[key] = common.CopyBytes(value)
	}

	for hash, logs := range s.logs {
		state.logs[hash] = make([]*types.Log, len(logs))
		copy(state.logs[hash], logs)
	}
	for hash, preimage := range s.preimages {
		state.preimages[hash] = preimage
	}
	return state
}

func (s *StateDB) Snapshot() int {
	id := s.nextRevisionID
	s.nextRevisionID++
	s.validRevisions = append(s.validRevisions, revision{id, s.journal.length()})
	return id
}

func (s *StateDB) RevertToSnapshot(revid int) {
	idx := sort.Search(len(s.validRevisions), func(i int) bool {
		return s.validRevisions[i].id >= revid
	})
	if idx == len(s.validRevisions) || s.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := s.validRevisions[idx].journalIndex

	s.journal.revert(s, snapshot)
	s.validRevisions = s.validRevisions[:idx]
}

//Put account's data to db
func (s *StateDB) Put(account string, key string, value []byte) {
	optKey := acctDataPrefix + linkSymbol + account + linkSymbol + key
	s.put(optKey, value)
}

//Get account's data from db
func (s *StateDB) Get(account string, key string) ([]byte, error) {
	optKey := acctDataPrefix + linkSymbol + account + linkSymbol + key
	return s.get(optKey)
}

//Delete account's data from db
func (s *StateDB) Delete(account string, key string) {
	optKey := acctDataPrefix + linkSymbol + account + linkSymbol + key
	s.put(optKey, nil)
}

// ReceiptRoot compute one tx‘ receipt hash
func (s *StateDB) ReceiptRoot() common.Hash {
	s.Finalise()
	return s.trie.Hash()
}

func (s *StateDB) IntermediateRoot() common.Hash {
	s.Finalise()
	return s.trie.Hash()
}

// execute transaction called
func (s *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	s.thash = thash
	s.bhash = bhash
	s.txIndex = ti
}

func (s *StateDB) clearJournalAndRefund() {
	s.journal = newJournal()
	s.validRevisions = s.validRevisions[:0]
	s.refund = 0
}

func (s *StateDB) Finalise() {
	for key := range s.journal.dirties {
		s.dirtySet[key] = struct{}{}
	}

	for key := range s.dirtySet {
		value, exsit := s.writeSet[key]
		if !exsit {
			panic("WriteSet is invalid when commit")
		}
		//update the value to trie
		if value != nil {
			s.trie.TryUpdate([]byte(key), value)
		} else {
			s.trie.TryDelete([]byte(key))
		}
		delete(s.dirtySet, key)
	}

	s.clearJournalAndRefund()
}

// commit call, save state change record
func (s *StateDB) genBlockStateOut(parentHash, blockHash common.Hash, blockNum uint64) *types.StateOut {
	stateOut := &types.StateOut{
		ParentHash: parentHash,
		Number:     blockNum,
		Hash:       blockHash,
		ReadSet:    make([]*types.KvNode, 0, len(s.readSet)),
	}

	// replay
	for key, value := range s.readSet {
		stateOut.ReadSet = append(stateOut.ReadSet,
			&types.KvNode{Key: key, Value: common.CopyBytes(value)})
	}

	return stateOut
}

//Commit the block state to db. after success please call commitcache
//batch: batch to db
//blockHash: the hash of commit block
func (s *StateDB) Commit(batch fdb.Batch, blockHash common.Hash, blockNum uint64) (common.Hash, error) {
	defer s.clearJournalAndRefund()

	var parentHash common.Hash
	s.Finalise()

	if s.Error() != nil {
		return common.Hash{}, errors.New("DB error when commit")
	}

	db := s.db.GetDB()
	if blockNum == 0 {
		parentHash = common.Hash{}
	} else {
		parentHash = rawdb.ReadCanonicalHash(db, blockNum-1)
	}

	//generate revert and write to db
	stateOut := s.genBlockStateOut(parentHash, blockHash, blockNum)
	rawdb.WriteBlockStateOut(batch, blockHash, stateOut)
	rawdb.WriteOptBlockHash(batch, blockHash)

	root, err := s.trie.Commit(nil)
	return root, err
}

//TransToSpecBlock change block state (from->to)
func TransToSpecBlock(db fdb.Database, cache Database, from common.Hash, to common.Hash) error {
	batch := db.NewBatch()

	rollState := rawdb.ReadBlockStateOut(db, from)
	fwdState := rawdb.ReadBlockStateOut(db, to)
	if rollState == nil || fwdState == nil {
		err := fmt.Errorf("from or to's stateout not exsit, from:%x to:%x", from, to)
		return err
	}

	for rollState.Number > fwdState.Number {
		rawdb.DeleteBlockStateOut(batch, rollState.Hash)
		rollState = rawdb.ReadBlockStateOut(db, rollState.ParentHash)
	}
	err := batch.Write()
	return err
}

//TraceNew get state of special block hash for trace
//blockHash: the hash of block
func TraceNew(blockHash common.Hash, cache Database) (*StateDB, error) {
	db := cache.GetDB()
	stateOut := rawdb.ReadBlockStateOut(db, blockHash)

	if stateOut == nil {
		err := fmt.Errorf("TraceNew blockHash error, hash:%x", blockHash)
		return nil, err
	}

	stateDb := &StateDB{
		db:         cache,
		readSet:    make(map[string][]byte),
		writeSet:   make(map[string][]byte),
		dirtySet:   make(map[string]struct{}),
		logs:       make(map[common.Hash][]*types.Log),
		preimages:  make(map[common.Hash][]byte),
		journal:    newJournal(),
		stateTrace: true}

	for _, node := range stateOut.ReadSet {
		stateDb.writeSet[node.Key] = common.CopyBytes(node.Value)
	}

	return stateDb, nil
}
