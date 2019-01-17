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
	"container/list"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/utils/rlp"
	"golang.org/x/crypto/sha3"
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

const (
	optAdd = 1 // Reverts/Changes record add key value
	optUpd = 2 // Reverts/Changes record update key value
	optDel = 3 // Reverts/Changes record delete key value
)

// StateDB store block operate info
type StateDB struct {
	db Database

	readSet  map[string][]byte   // save old/unmodified data
	writeSet map[string][]byte   // last modify data
	dirtySet map[string]struct{} // writeSet which key is modified

	parentHash common.Hash // save previous block hash

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

	dirtyHash map[string]common.Hash

	stateTrace bool // replay transaction, true is replayed , false is not replayed

	lock sync.Mutex
}

type transferInfo struct {
	// list save state info
	rollBack list.List
	forworad list.List
}

//New func generate a statedb object
//parentHash: block's parent hash, db: cachedb
func New(parentHash common.Hash, db Database) (*StateDB, error) {
	//current cache hash
	db.RLock()
	hash := db.GetHash()
	db.RUnLock()
	if hash != parentHash {
		err := fmt.Errorf("stateNew error, hash:%x,phash:%x", hash, parentHash)
		return nil, err
	}
	return &StateDB{
		db:         db,
		parentHash: parentHash,
		readSet:    make(map[string][]byte),
		writeSet:   make(map[string][]byte),
		dirtySet:   make(map[string]struct{}),
		logs:       make(map[common.Hash][]*types.Log),
		preimages:  make(map[common.Hash][]byte),
		dirtyHash:  make(map[string]common.Hash),
		journal:    newJournal(),
		stateTrace: false}, nil
}

// only save first err
func (s *StateDB) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *StateDB) Error() error {
	return s.dbErr
}

func (s *StateDB) Reset() error {
	s.readSet = make(map[string][]byte)
	s.writeSet = make(map[string][]byte)
	s.dirtySet = make(map[string]struct{})
	s.dirtyHash = make(map[string]common.Hash)
	s.parentHash = common.Hash{}
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

// save transaction log
func (s *StateDB) AddLog(log *types.Log) {
	s.journal.append(addLogChange{txhash: s.thash})

	log.TxHash = s.thash
	log.BlockHash = s.bhash
	log.TxIndex = uint(s.txIndex)
	log.Index = s.logSize
	s.logs[s.thash] = append(s.logs[s.thash], log)
	s.logSize++
}

// get a strip of transaction log
func (s *StateDB) GetLogs(hash common.Hash) []*types.Log {
	return s.logs[hash]
}

// get all transaction log
func (s *StateDB) Logs() []*types.Log {
	var logs []*types.Log
	for _, lgs := range s.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// hash is preimageHash
func (s *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	if _, ok := s.preimages[hash]; !ok {
		s.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		s.preimages[hash] = pi
	}
}

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
	if s.stateTrace == true {
		errInfo := fmt.Sprintf("No value when trace,phash:%x", s.parentHash)
		err := errors.New(errInfo)
		s.setError(err)
		return nil, err
	}

	s.db.RLock()
	hash := s.db.GetHash()

	if hash != s.parentHash {
		errInfo := fmt.Sprintf("Inconsistent hash:%x phash:%x", hash, s.parentHash)
		err := errors.New(errInfo)
		s.setError(err)
		s.db.RUnLock()
		return nil, err
	}

	value, err := s.db.Get(key)
	s.db.RUnLock()

	if err != nil {
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

	value, err := s.db.Get(optKey)

	if err != nil {
		return common.Hash{}
	}

	return common.BytesToHash(value)
}

//RpcGet provide get value of the key to rpc
//when called please RLock cachedb
func (s *StateDB) RpcGet(account string, key string) ([]byte, error) {
	optKey := acctDataPrefix + linkSymbol + account + linkSymbol + key
	value, err := s.db.Get(optKey)

	if err != nil {
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

	state := &StateDB{db: s.db,
		readSet:    make(map[string][]byte, len(s.writeSet)),
		writeSet:   make(map[string][]byte, len(s.writeSet)),
		dirtySet:   make(map[string]struct{}, len(s.dirtySet)),
		dirtyHash:  make(map[string]common.Hash),
		parentHash: s.parentHash,
		refund:     s.refund,
		logs:       make(map[common.Hash][]*types.Log, len(s.logs)),
		logSize:    s.logSize,
		preimages:  make(map[common.Hash][]byte),
		journal:    newJournal()}

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

func kvRlpHash(kvNode *types.KvNode) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, kvNode)
	hw.Sum(h[:0])
	return h
}

// ReceiptRoot compute one tx‘ receipt hash
func (s *StateDB) ReceiptRoot() common.Hash {
	defer s.Finalise()

	keys := make([]string, 0, len(s.journal.dirties))
	for key := range s.journal.dirties {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	dirtyHash := make([]common.Hash, 0, len(keys))

	for _, key := range keys {
		value := s.writeSet[key]
		node := &types.KvNode{Key: key, Value: value}
		hash := kvRlpHash(node)
		dirtyHash = append(dirtyHash, hash)
		s.dirtyHash[key] = hash
	}

	return common.MerkleRoot(dirtyHash)
}

func (s *StateDB) IntermediateRoot() common.Hash {
	defer s.Finalise()

	if len(s.journal.dirties) != 0 {
		s.ReceiptRoot()
	}

	keys := make([]string, 0, len(s.dirtyHash))
	for key := range s.dirtyHash {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	dirtyHash := make([]common.Hash, 0, len(keys))
	for _, key := range keys {
		hash := s.dirtyHash[key]
		dirtyHash = append(dirtyHash, hash)
	}

	return common.MerkleRoot(dirtyHash)
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
	s.clearJournalAndRefund()
}

// commit call, save state change record
func (s *StateDB) genBlockStateOut(parentHash, blockHash common.Hash, blockNum uint64) *types.StateOut {
	stateOut := &types.StateOut{ParentHash: parentHash,
		Number:  blockNum,
		Hash:    blockHash,
		ReadSet: make([]*types.KvNode, 0, len(s.readSet)),
		Reverts: make([]*types.OptInfo, 0, len(s.dirtySet)),
		Changes: make([]*types.OptInfo, 0, len(s.dirtySet))}

	for key := range s.dirtySet {
		readValue := s.readSet[key]
		writeValue := s.writeSet[key]

		if readValue != nil && writeValue != nil {
			stateOut.Reverts = append(stateOut.Reverts,
				&types.OptInfo{Key: key, Value: common.CopyBytes(readValue), Opt: optUpd})
			stateOut.Changes = append(stateOut.Changes,
				&types.OptInfo{Key: key, Value: common.CopyBytes(writeValue), Opt: optUpd})
		} else if readValue != nil && writeValue == nil {
			stateOut.Reverts = append(stateOut.Reverts,
				&types.OptInfo{Key: key, Value: common.CopyBytes(readValue), Opt: optAdd})
			stateOut.Changes = append(stateOut.Changes,
				&types.OptInfo{Key: key, Value: nil, Opt: optDel})
		} else {
			stateOut.Reverts = append(stateOut.Reverts,
				&types.OptInfo{Key: key, Value: common.CopyBytes(readValue), Opt: optDel})
			stateOut.Changes = append(stateOut.Changes,
				&types.OptInfo{Key: key, Value: common.CopyBytes(writeValue), Opt: optAdd})
		}
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

	if s.Error() != nil {
		return common.Hash{}, errors.New("DB error when commit")
	}

	for key := range s.journal.dirties {
		s.dirtySet[key] = struct{}{}
	}

	// check parentHash
	curHash := s.db.GetHash()

	if curHash != s.parentHash {
		return common.Hash{}, fmt.Errorf("Error hash when commit, cache: %x,parent: %x", curHash, s.parentHash)
	}

	//generate revert and write to db
	stateOut := s.genBlockStateOut(s.parentHash, blockHash, blockNum)
	rawdb.WriteBlockStateOut(batch, blockHash, stateOut)

	//scan dirtyset, commit to db
	for key := range s.dirtySet {
		value, exsit := s.writeSet[key]
		if exsit == false {
			panic("WriteSet is invalid when commit")
		}
		//update the value to db
		var err error
		if value != nil {
			err = batch.Put([]byte(key), value)
		} else {
			err = batch.Delete([]byte(key))
		}

		if err != nil {
			return common.Hash{}, err
		}
		//call commitcache write cache after
	}

	rawdb.WriteOptBlockHash(batch, blockHash)
	hash := s.IntermediateRoot()
	return hash, nil
}

//CommitCache commit the block state to cache
//call after state commit to db success
func (s *StateDB) CommitCache(blockHash common.Hash) {
	//scan dirtyset, commit to cache
	for key := range s.dirtySet {
		value, exsit := s.writeSet[key]
		if exsit == false {
			panic("WriteSet is invalid when commitcache")
		}
		s.db.PutCache(key, value)
	}
	s.db.SetHash(blockHash)
}

func recoverDbByOptInfos(batch fdb.Batch, optInfos []*types.OptInfo) error {
	var err error
	for _, optinfo := range optInfos {
		if optinfo.Opt == optAdd {
			err = batch.Put([]byte(optinfo.Key), optinfo.Value)
		} else if optinfo.Opt == optUpd {
			err = batch.Put([]byte(optinfo.Key), optinfo.Value)
		} else if optinfo.Opt == optDel {
			err = batch.Delete([]byte(optinfo.Key))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func recoverCacheByOptInfos(cache Database, optInfos []*types.OptInfo) {
	for _, optinfo := range optInfos {
		if optinfo.Opt == optAdd {
			cache.PutCache(optinfo.Key, optinfo.Value)
		} else if optinfo.Opt == optUpd {
			cache.PutCache(optinfo.Key, optinfo.Value)
		} else if optinfo.Opt == optDel {
			cache.DeleteCache(optinfo.Key)
		}
	}
}

func writeTransferToDb(db fdb.Database, transInfo *transferInfo, to common.Hash) error {
	var err error
	var state *types.StateOut
	rollList := &transInfo.rollBack
	fwdList := &transInfo.forworad
	batch := db.NewBatch()

	for node := rollList.Front(); node != nil; node = node.Next() {
		state = node.Value.(*types.StateOut)
		err = recoverDbByOptInfos(batch, state.Reverts)
		if err != nil {
			return err
		}

		//delete rollback block state
		rawdb.DeleteBlockStateOut(batch, state.Hash)
	}

	for node := fwdList.Front(); node != nil; node = node.Next() {
		state = node.Value.(*types.StateOut)
		err = recoverDbByOptInfos(batch, state.Changes)
		if err != nil {
			return err
		}
	}
	rawdb.WriteOptBlockHash(batch, to)
	err = batch.Write()
	return err
}

func writeTransferToCache(cache Database, transInfo *transferInfo, to common.Hash) {
	var state *types.StateOut
	rollList := &transInfo.rollBack
	fwdList := &transInfo.forworad

	for node := rollList.Front(); node != nil; node = node.Next() {
		state = node.Value.(*types.StateOut)
		recoverCacheByOptInfos(cache, state.Reverts)
	}

	for node := fwdList.Front(); node != nil; node = node.Next() {
		state = node.Value.(*types.StateOut)
		recoverCacheByOptInfos(cache, state.Changes)
	}

	cache.SetHash(to)
}

func fetchBranch(db fdb.Database, from common.Hash, to common.Hash) (*transferInfo, error) {
	var transInfo transferInfo

	rollState := rawdb.ReadBlockStateOut(db, from)
	fwdState := rawdb.ReadBlockStateOut(db, to)

	if rollState == nil || fwdState == nil {
		err := fmt.Errorf("from or to's stateout not exsit, from:%x to:%x", from, to)
		return nil, err
	}

	for rollState.Number > fwdState.Number {
		transInfo.rollBack.PushBack(rollState)
		rollState = rawdb.ReadBlockStateOut(db, rollState.ParentHash)
		if rollState == nil {
			err := fmt.Errorf("fetch branch failed, rollBack state not exsit, from:%x to:%x", from, to)
			return nil, err
		}
	}

	for fwdState.Number > rollState.Number {
		transInfo.forworad.PushFront(fwdState)
		fwdState = rawdb.ReadBlockStateOut(db, fwdState.ParentHash)
		if fwdState == nil {
			err := fmt.Errorf("fetch branch failed, forward state not exsit, from:%x to:%x", from, to)
			return nil, err
		}
	}

	for rollState.ParentHash != fwdState.ParentHash {
		transInfo.rollBack.PushBack(rollState)
		transInfo.forworad.PushFront(fwdState)
		rollState = rawdb.ReadBlockStateOut(db, rollState.ParentHash)
		fwdState = rawdb.ReadBlockStateOut(db, fwdState.ParentHash)

		if rollState == nil || fwdState == nil {
			err := fmt.Errorf("fetch branch failed when rollback and forward, from:%x to:%x", from, to)
			return nil, err
		}
	}

	if rollState != nil && fwdState != nil {
		//same node not push
		if rollState.Hash != fwdState.Hash {
			transInfo.rollBack.PushBack(rollState)
			transInfo.forworad.PushFront(fwdState)
		}
	}

	return &transInfo, nil
}

//TransToSpecBlock change block state (from->to)
func TransToSpecBlock(db fdb.Database, cache Database, from common.Hash, to common.Hash) error {
	//get near parent hash of from and to
	transInfo, err := fetchBranch(db, from, to)
	if err != nil {
		return err
	}

	cache.Lock()
	defer cache.UnLock()
	optHash := cache.GetHash()
	if optHash != from {
		errInfo := fmt.Sprintf("Invalid current hash, from:%x cur:%x", from, optHash)
		return errors.New(errInfo)
	}
	//exe rollback and forward
	err = writeTransferToDb(db, transInfo, to)
	if err != nil {
		return err
	}

	writeTransferToCache(cache, transInfo, to)

	return nil
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
		parentHash: stateOut.ParentHash,
		readSet:    make(map[string][]byte),
		writeSet:   make(map[string][]byte),
		dirtySet:   make(map[string]struct{}),
		dirtyHash:  make(map[string]common.Hash),
		logs:       make(map[common.Hash][]*types.Log),
		preimages:  make(map[common.Hash][]byte),
		journal:    newJournal(),
		stateTrace: true}

	for _, node := range stateOut.ReadSet {
		stateDb.writeSet[node.Key] = common.CopyBytes(node.Value)
	}

	return stateDb, nil
}
