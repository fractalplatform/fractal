package snapshot

import (
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	mdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

func TestSnapshot(t *testing.T) {
	db := mdb.NewMemDatabase()
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
		t.Error("set snapshot err", err)
	}

	if timestamp != 100000000 {
		t.Error("set snapshot err", err)
	}

	timestamp, err = snapshotManager.GetPrevSnapshotTime(100000000)
	if err != nil {
		t.Error("set snapshot err", err)
	}

	_, _, err = snapshotManager.GetCurrentSnapshotHash()
	if err != nil {
		t.Error("set snapshot err", err)
	}

	_, err = snapshotManager.GetSnapshotMsg(addr, key, 100000000)
	if err != nil {
		t.Error("set snapshot err", err)
	}
}
