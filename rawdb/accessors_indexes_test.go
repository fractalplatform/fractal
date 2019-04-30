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
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
	mdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

// Tests that positional lookup metadata can be stored and retrieved.
func TestLookupStorage(t *testing.T) {
	db := mdb.NewMemDatabase()

	action1 := types.NewAction(types.Transfer, common.Name("fromtest"), common.Name("tototest"), uint64(3), uint64(3), uint64(2000), big.NewInt(1000), []byte("test action1"), []byte("test remark1"))
	action2 := types.NewAction(types.Transfer, common.Name("fromtest"), common.Name("tototest"), uint64(3), uint64(3), uint64(2000), big.NewInt(1000), []byte("test action2"), []byte("test remark2"))
	action3 := types.NewAction(types.Transfer, common.Name("fromtest"), common.Name("tototest"), uint64(3), uint64(3), uint64(2000), big.NewInt(1000), []byte("test action3"), []byte("test remark3"))

	tx1 := types.NewTransaction(uint64(1), big.NewInt(1), action1)
	tx2 := types.NewTransaction(uint64(2), big.NewInt(2), action2)
	tx3 := types.NewTransaction(uint64(2), big.NewInt(2), action3)

	txs := []*types.Transaction{tx1, tx2, tx3}

	block := &types.Block{
		Head: &types.Header{Number: big.NewInt(314), Coinbase: "coinbase"},
		Txs:  txs,
	}

	// Check that no transactions entries are in a pristine database
	for i, tx := range txs {
		if txn, _, _, _ := ReadTransaction(db, tx.Hash()); txn != nil {
			t.Fatalf("tx #%d [%x]: non existent transaction returned: %v", i, tx.Hash(), txn)
		}
	}
	// Insert all the transactions into the database, and verify contents
	WriteBlock(db, block)
	WriteTxLookupEntries(db, block)

	for i, tx := range txs {
		if txn, hash, number, index := ReadTransaction(db, tx.Hash()); txn == nil {
			t.Fatalf("tx #%d [%x]: transaction not found", i, tx.Hash())
		} else {
			if hash != block.Hash() || number != block.NumberU64() || index != uint64(i) {
				t.Fatalf("tx #%d [%x]: positional metadata mismatch: have %x/%d/%d, want %x/%v/%v", i, tx.Hash(), hash, number, index, block.Hash(), block.NumberU64(), i)
			}
			if tx.Hash() != txn.Hash() {
				t.Fatalf("tx #%d [%x]: transaction mismatch: have %v, want %v", i, tx.Hash(), txn, tx)
			}
		}
	}
	// Delete the transactions and check purge
	for i, tx := range txs {
		DeleteTxLookupEntry(db, tx.Hash())
		if txn, _, _, _ := ReadTransaction(db, tx.Hash()); txn != nil {
			t.Fatalf("tx #%d [%x]: deleted transaction returned: %v", i, tx.Hash(), txn)
		}
	}
}
