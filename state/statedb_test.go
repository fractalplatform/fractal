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
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/types"
	mdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

func TestNew(t *testing.T) {
	db := mdb.NewMemDatabase()
	cacheDb := NewDatabase(db)

	rootHash := common.BytesToHash([]byte("not exist hash"))
	_, err := New(rootHash, cacheDb)
	if err == nil {
		t.Error(fmt.Sprintf("new state error, %v", err))
	}

	rootNullHash := common.Hash{}
	_, err = New(rootNullHash, cacheDb)
	if err != nil {
		t.Error(fmt.Sprintf("new state error, %v", err))
	}
}

func TestReset(t *testing.T) {
	db := mdb.NewMemDatabase()
	cacheDb := NewDatabase(db)
	rootNullHash := common.Hash{}

	stateX, err := New(rootNullHash, cacheDb)
	if err != nil {
		t.Error(fmt.Sprintf("new state error, %v", err))
	}

	accountName := "A"
	key := "key"
	value := []byte("100")

	stateX.Put(accountName, key, value)
	valueRet, err := stateX.Get(accountName, key)
	if err != nil {
		t.Error(fmt.Sprintf("get value error, %v", err))
	}

	if !bytes.Equal(valueRet, value) {
		t.Error("value error")
	}

	err = stateX.Reset(rootNullHash)
	if err != nil {
		t.Error(fmt.Sprintf("state reset error, %v", err))
	}

	valueRet, err = stateX.Get(accountName, key)
	if err != nil {
		t.Error(fmt.Sprintf("get value error, %v", err))
	}

	if !bytes.Equal(valueRet, nil) {
		t.Error("value error")
	}

	err = stateX.Reset(common.BytesToHash([]byte("not exist hash")))
	if err == nil {
		t.Error(fmt.Sprintf("state reset error"))
	}
}

func TestRefund(t *testing.T) {
	db := mdb.NewMemDatabase()
	cacheDb := NewDatabase(db)
	rootNullHash := common.Hash{}

	stateX, err := New(rootNullHash, cacheDb)
	if err != nil {
		t.Error(fmt.Sprintf("new state error, %v", err))
	}

	stateX.AddRefund(1000)
	stateX.AddRefund(2000)

	refund := stateX.GetRefund()
	if refund != 3000 {
		t.Error(fmt.Sprintf("refund error, %v", refund))
	}
}

func TestPutAndGet(t *testing.T) {
	db := mdb.NewMemDatabase()
	batch := db.NewBatch()
	cachedb := NewDatabase(db)
	prevRoot := common.Hash{}
	currentBlockHash := common.Hash{}
	currentBlockNumber := uint64(0)

	accountName := "testtest"
	key := "testKey"
	value := []byte("1")

	state, err := New(prevRoot, cachedb)
	if err != nil {
		t.Error(fmt.Sprintf("new state error, %v", err))
	}

	state.Put(accountName, key, value)
	root, err := state.Commit(batch, currentBlockHash, currentBlockNumber)
	if err != nil {
		t.Error("commit trie err", err)
	}

	triedb := state.db.TrieDB()
	triedb.Reference(root, common.Hash{})
	if err := triedb.Commit(root, false); err != nil {
		t.Error("commit db err", err)
	}
	batch.Write()
	triedb.Dereference(root)

	// read
	state, err = New(root, cachedb)
	if err != nil {
		t.Error(fmt.Sprintf("new state error, %v", err))
	}

	value, err = state.Get(accountName, key)
	if err != nil {
		t.Error(fmt.Sprintf("get value error, %v", err))
	}

	if !bytes.Equal(value, []byte("1")) {
		t.Error("value error")
	}

	accountName01 := "testtest01"
	value, err = state.Get(accountName01, key)
	if err != nil {
		t.Error(fmt.Sprintf("get value error, %v", err))
	}

	if len(value) != 0 {
		t.Error("value error")
	}
}

func TestSetAndGetState(t *testing.T) {
	db := mdb.NewMemDatabase()
	batch := db.NewBatch()
	cachedb := NewDatabase(db)
	prevRoot := common.Hash{}
	currentBlockHash := common.Hash{}
	currentBlockNumber := uint64(0)

	state, _ := New(prevRoot, cachedb)
	for i := 0; i < 4; i++ {
		addr := string([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := []byte("sk" + strconv.Itoa(i) + strconv.Itoa(j))
			value := []byte("sv" + strconv.Itoa(i) + strconv.Itoa(j))
			state.SetState(addr, common.BytesToHash(key), common.BytesToHash(value))
		}
	}

	root, err := state.Commit(batch, currentBlockHash, currentBlockNumber)
	if err != nil {
		t.Error("commit trie err", err)
	}

	triedb := state.db.TrieDB()
	triedb.Reference(root, common.Hash{})
	if err := triedb.Commit(root, false); err != nil {
		t.Error("commit db err", err)
	}
	triedb.Dereference(root)
	batch.Write()

	//get from db
	cachedb01 := NewDatabase(db)
	state01, _ := New(root, cachedb01)
	for i := 0; i < 4; i++ {
		addr := string([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := []byte("sk" + strconv.Itoa(i) + strconv.Itoa(j))
			value := []byte("sv" + strconv.Itoa(i) + strconv.Itoa(j))
			s := state01.GetState(addr, common.BytesToHash(key))
			if common.BytesToHash(value) != s {
				t.Error("get from cachedb failed")
			}
		}
	}

}

func TestLog(t *testing.T) {
	db := mdb.NewMemDatabase()
	cachedb := NewDatabase(db)
	prevRoot := common.Hash{}
	currentBlockHash := common.BytesToHash([]byte("01"))
	currentBlockNumber := uint64(0)
	currentTxHash := common.BytesToHash([]byte("02"))
	currentTxIndex := 0

	currentLog01 := types.Log{
		Data:        []byte("01"),
		BlockNumber: currentBlockNumber,
	}

	currentLog02 := types.Log{
		Data:        []byte("02"),
		BlockNumber: currentBlockNumber,
	}

	type args struct {
		txLog types.Log
	}

	currentLog := []args{
		args{txLog: currentLog01},
		args{txLog: currentLog02},
	}

	state, _ := New(prevRoot, cachedb)
	state.Prepare(currentTxHash, currentBlockHash, currentTxIndex)
	state.AddLog(&currentLog01)
	state.AddLog(&currentLog02)

	getLogs := state.GetLogs(currentTxHash)
	for i, l := range getLogs {
		if !bytes.Equal(l.Data, currentLog[i].txLog.Data) {
			t.Error(fmt.Sprintf("log error get %v, want %v", l.Data, currentLog[i].txLog.Data))
		}
	}
}

func TestRevertSnap(t *testing.T) {
	db := mdb.NewMemDatabase()
	cachedb := NewDatabase(db)
	prevHash := common.Hash{}
	state, _ := New(prevHash, cachedb)

	addr := "addr01"
	key1 := []byte("sk01")
	value1 := []byte("sv01")
	state.SetState(addr, common.BytesToHash(key1), common.BytesToHash(value1))

	snapInx := state.Snapshot()

	key2 := []byte("sk02")
	value2 := []byte("sv02")
	state.SetState(addr, common.BytesToHash(key2), common.BytesToHash(value2))

	testValue1 := state.GetState(addr, common.BytesToHash(key1))
	testValue2 := state.GetState(addr, common.BytesToHash(key2))

	if testValue1 != common.BytesToHash(value1) {
		t.Error("test value1 before revert failed")
	}

	if testValue2 != common.BytesToHash(value2) {
		t.Error("test value2 before revert failed")
	}

	state.RevertToSnapshot(snapInx)

	testValue1 = state.GetState(addr, common.BytesToHash(key1))
	testValue2 = state.GetState(addr, common.BytesToHash(key2))

	if testValue1 != common.BytesToHash(value1) {
		t.Error("test value1 after revert failed")
	}

	if (testValue2 != common.Hash{}) {
		t.Error("test value2 after revert failed ", testValue2)
	}
}

//element : 1->2->3
func TestTransToSpecBlock1(t *testing.T) {
	db := mdb.NewMemDatabase()
	batch := db.NewBatch()
	cachedb := NewDatabase(db)
	addr := "addr01"
	var curHash common.Hash
	key1 := []byte("sk")
	root := common.Hash{}
	var roothash [12]common.Hash

	for i := 0; i < 12; i++ {
		state, _ := New(root, cachedb)
		value1 := []byte("sv" + strconv.Itoa(i))
		state.SetState(addr, common.BytesToHash(key1), common.BytesToHash(value1))
		curHash = common.BytesToHash([]byte("hash" + strconv.Itoa(i)))

		root, err := state.Commit(batch, curHash, uint64(i))
		if err != nil {
			t.Error("commit trie err", err)
		}
		triedb := state.db.TrieDB()
		if err := triedb.Commit(root, false); err != nil {
			t.Error("commit db err", err)
		}
		rawdb.WriteCanonicalHash(batch, curHash, uint64(i))
		batch.Write()
		roothash[i] = root
	}

	from := curHash
	to := common.BytesToHash([]byte("hash" + strconv.Itoa(1)))
	err := TransToSpecBlock(db, cachedb, from, to)

	if err != nil {
		t.Error("TransToSpecBlock return fail")
	}

	state, _ := New(roothash[1], cachedb)
	hash := state.GetState(addr, common.BytesToHash(key1))

	value := []byte("sv" + strconv.Itoa(1))
	if hash != common.BytesToHash(value) {
		t.Error("TestTransToSpecBlock, to block 1 failed")
	}
}

func TestStateDB_IntermediateRoot(t *testing.T) {
	state, err := New(common.Hash{}, NewDatabase(mdb.NewMemDatabase()))
	if err != nil {
		t.Error("New err")
	}
	vv := "asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklzasdfghjklz" +
		"asdfssssssssssssssssssss"
	v := []byte(vv)
	k := "ahhshhhhhhhhhhhhhhhhddddddddhj"
	st := time.Now()

	addr := "addr01"
	for j := 0; j < 680; j++ {
		tk := k + strconv.Itoa(j)
		tv := append(v, byte(j))
		state.Put(addr, tk, tv)
		state.ReceiptRoot()
	}
	state.IntermediateRoot()
	fmt.Println("time: ", time.Since(st))
}
