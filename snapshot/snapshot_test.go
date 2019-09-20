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

package snapshot

import (
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

func TestSnapshot(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	batch := db.NewBatch()
	cachedb := state.NewDatabase(db)
	prevHash := common.Hash{}
	state1, _ := state.New(prevHash, cachedb)

	addr := "snapshot01"
	key := "aaaaaa"
	value := []byte("1")
	state1.Put(addr, key, value)

	root, err := state1.Commit(batch, prevHash, 0)
	if err != nil {
		t.Error("commit trie err", err)
	}

	triedb := state1.Database().TrieDB()
	triedb.Reference(root, common.Hash{})
	if err := triedb.Commit(root, false); err != nil {
		t.Error("commit db err", err)
	}
	triedb.Dereference(root)

	state2, _ := state.New(root, cachedb)
	snapshotManager := NewSnapshotManager(state2)
	err = snapshotManager.SetSnapshot(uint64(100000000), BlockInfo{Number: 0, BlockHash: prevHash, Timestamp: 0})
	if err != nil {
		t.Error("set snapshot err", err)
	}
	snapshotInfo := types.SnapshotInfo{
		Root: root,
	}
	key1 := types.SnapshotBlock{
		Number:    0,
		BlockHash: prevHash,
	}
	rawdb.WriteSnapshot(db, key1, snapshotInfo)

	timestamp, err := snapshotManager.GetLastSnapshotTime()
	if err != nil {
		t.Error("get snapshot err", err)
	}

	if timestamp != 100000000 {
		t.Error("get snapshot time err", err)
	}

	timestamp, err = snapshotManager.GetPrevSnapshotTime(100000000)
	if err != nil {
		t.Error("get prev snapshot err", err)
	}

	_, _, err = snapshotManager.GetCurrentSnapshotHash()
	if err != nil {
		t.Error("get current snapshot err", err)
	}

	_, err = snapshotManager.GetSnapshotMsg(addr, key, 100000000)
	if err != nil {
		t.Error("get snapshot msg err", err)
	}

	_, err = snapshotManager.GetSnapshotState(100000000)
	if err != nil {
		t.Error("get snapshot state err", err)
	}
}

func TestSnapshotError(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	batch := db.NewBatch()
	cachedb := state.NewDatabase(db)
	prevHash := common.Hash{}
	state1, _ := state.New(prevHash, cachedb)

	addr := "snapshot01"
	key := "aaaaaa"
	value := []byte("1")
	state1.Put(addr, key, value)

	root, err := state1.Commit(batch, prevHash, 0)
	if err != nil {
		t.Error("commit trie err", err)
	}

	triedb := state1.Database().TrieDB()
	triedb.Reference(root, common.Hash{})
	if err := triedb.Commit(root, false); err != nil {
		t.Error("commit db err", err)
	}
	triedb.Dereference(root)

	state2, _ := state.New(root, cachedb)
	snapshotManager := NewSnapshotManager(state2)

	_, err = snapshotManager.GetLastSnapshotTime()
	if err == nil {
		t.Error("get snapshot err", err)
	}

	_, err = snapshotManager.GetPrevSnapshotTime(100000000)
	if err == nil {
		t.Error("get prev snapshot err", err)
	}

	_, _, err = snapshotManager.GetCurrentSnapshotHash()
	if err == nil {
		t.Error("get current snapshot err", err)
	}

	_, err = snapshotManager.GetSnapshotMsg(addr, key, 100000000)
	if err == nil {
		t.Error("get snapshot msg err", err)
	}

	_, err = snapshotManager.GetSnapshotState(100000000)
	if err == nil {
		t.Error("get snapshot state err", err)
	}
}
