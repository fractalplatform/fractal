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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/fdb"
)

func TestSetState(t *testing.T) {
	db := fdb.NewMemDatabase()
	batch := db.NewBatch()
	cachedb := NewDatabase(db)
	prevHash := common.Hash{}
	curHash := common.BytesToHash([]byte("2222222"))
	state, _ := New(prevHash, cachedb)
	for i := 0; i < 4; i++ {
		addr := string([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := []byte("sk" + strconv.Itoa(i) + strconv.Itoa(j))
			value := []byte("sv" + strconv.Itoa(i) + strconv.Itoa(j))
			state.SetState(addr, common.BytesToHash(key), common.BytesToHash(value))
		}
	}
	_, err := state.Commit(batch, curHash, 1)
	if err != nil {
		t.Error("commit trie err", err)
	}

	batch.Write()
	state.CommitCache(curHash)

	//get from db
	cachedb1 := NewDatabase(db)
	state3, _ := New(curHash, cachedb1)
	for i := 0; i < 4; i++ {
		addr := string([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := []byte("sk" + strconv.Itoa(i) + strconv.Itoa(j))
			value := []byte("sv" + strconv.Itoa(i) + strconv.Itoa(j))
			s := state3.GetState(addr, common.BytesToHash(key))
			if common.BytesToHash(value) != s {
				t.Error("get from cachedb failed")
			}
		}
	}

}

func TestRevertSnap(t *testing.T) {
	db := fdb.NewMemDatabase()
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
	db := fdb.NewMemDatabase()
	batch := db.NewBatch()
	cachedb := NewDatabase(db)
	addr := "addr01"
	var curHash common.Hash
	key1 := []byte("sk")

	prevHash := common.Hash{}

	for i := 0; i < 12; i++ {
		state, _ := New(prevHash, cachedb)
		value1 := []byte("sv" + strconv.Itoa(i))
		state.SetState(addr, common.BytesToHash(key1), common.BytesToHash(value1))
		curHash = common.BytesToHash([]byte("hash" + strconv.Itoa(i)))
		_, err := state.Commit(batch, curHash, uint64(i))

		if err != nil {
			t.Error("Commit hash1 content failed")
		}
		batch.Write()
		state.CommitCache(curHash)
		prevHash = curHash
	}

	from := curHash
	to := common.BytesToHash([]byte("hash" + strconv.Itoa(1)))
	err := TransToSpecBlock(db, cachedb, from, to)

	if err != nil {
		t.Error("TransToSpecBlock return fail")
	}

	state, _ := New(to, cachedb)
	hash := state.GetState(addr, common.BytesToHash(key1))

	value := []byte("sv" + strconv.Itoa(1))
	if hash != common.BytesToHash(value) {
		t.Error("TestTransToSpecBlock, to block 1 failed")
	}
}

//element : 0->1->2
//           ->3
func TestTransToSpecBlock2(t *testing.T) {
	db := fdb.NewMemDatabase()
	batch := db.NewBatch()
	cachedb := NewDatabase(db)
	addr := "addr01"
	var curHash common.Hash
	key1 := []byte("sk")

	//other branch
	prevHash := common.Hash{}
	state, _ := New(prevHash, cachedb)
	value1 := []byte("sv" + strconv.Itoa(0))
	state.SetState(addr, common.BytesToHash(key1), common.BytesToHash(value1))
	curHash = common.BytesToHash([]byte("hash" + strconv.Itoa(0)))
	_, err := state.Commit(batch, curHash, 0)

	if err != nil {
		t.Error("commit value 0 failed")
	}
	batch.Write()
	state.CommitCache(curHash)

	prevHash = curHash
	state1, _ := New(prevHash, cachedb)
	value3 := []byte("sv" + strconv.Itoa(3))
	state1.SetState(addr, common.BytesToHash(key1), common.BytesToHash(value3))
	curHash1 := common.BytesToHash([]byte("hash" + strconv.Itoa(3)))
	_, err = state1.Commit(batch, curHash1, 1)
	if err != nil {
		t.Error("commit value 3 failed")
	}
	batch.Write()
	state1.CommitCache(curHash1)

	err = TransToSpecBlock(db, cachedb, curHash1, prevHash)

	if err != nil {
		t.Error("Trans to spec block failed ")
	}

	prevHash = curHash
	for i := 1; i < 3; i++ {
		state, _ := New(prevHash, cachedb)
		value1 := []byte("sv" + strconv.Itoa(i))
		state.SetState(addr, common.BytesToHash(key1), common.BytesToHash(value1))
		curHash = common.BytesToHash([]byte("hash" + strconv.Itoa(i)))
		_, err := state.Commit(batch, curHash, uint64(i))

		if err != nil {
			t.Error("Commit hash1 content failed")
		}
		batch.Write()
		state.CommitCache(curHash)
		prevHash = curHash
	}

	from := curHash
	to := common.BytesToHash([]byte("hash" + strconv.Itoa(3)))
	err = TransToSpecBlock(db, cachedb, from, to)

	if err != nil {
		t.Error("TransToSpecBlock return fail")
	}

	state, _ = New(to, cachedb)
	hash := state.GetState(addr, common.BytesToHash(key1))

	value := []byte("sv" + strconv.Itoa(3))

	if hash != common.BytesToHash(value) {
		t.Error("TestTransToSpecBlock, to block 3 failed")
	}
}

func TestStateDB_IntermediateRoot(t *testing.T) {
	state, err := New(common.Hash{}, NewDatabase(fdb.NewMemDatabase()))
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
