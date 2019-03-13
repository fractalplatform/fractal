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

package blockchain

// import (
// 	"testing"

// 	"github.com/fractalplatform/fractal/rawdb"
// )

// func TestTheLastBlock(t *testing.T) {
// 	genesis, db, chain, st, err := newCanonical(t, tengine)
// 	if err != nil {
// 		t.Error("newCanonical err", err)
// 	}
// 	defer chain.Stop()

// 	prods, ht := makeProduceAndTime(st, 10)
// 	_, _, blocks, err := makeNewChain(t, genesis, chain, &db, len(prods), ht, prods, makeTransferTx)
// 	if err != nil {
// 		t.Error("makeNewChain err", err)
// 	}
// 	if blocks[len(blocks)-1].Hash() != rawdb.ReadHeadBlockHash(chain.db) {
// 		t.Fatalf("Write/Get HeadBlockHash failed")
// 	}
// }

// func TestForkChain(t *testing.T) {
// 	genesis, db, chain, st, err := newCanonical(t, tengine)
// 	if err != nil {
// 		t.Error("newCanonical err", err)
// 	}
// 	defer chain.Stop()

// 	prods, ht := makeProduceAndTime(st, 10)
// 	_, _, blocks, err := makeNewChain(t, genesis, chain, &db, len(prods), ht, prods, nil)
// 	if err != nil {
// 		t.Error("makeNewChain err", err)
// 	}

// 	prods = append(prods[0:3], prods[10:]...)
// 	ht = append(ht[0:3], ht[10:]...)
// 	genesis1, db1, chain1, _, err := newCanonical(t, tengine)
// 	if err != nil {
// 		t.Error("newCanonical err", err)
// 	}
// 	defer chain.Stop()

// 	_, _, _, err = makeNewChain(t, genesis1, chain1, &db1, len(prods), ht, prods, makeTransferTx)
// 	if err != nil {
// 		t.Error("makeNewChain err", err)
// 	}
// 	_, err = chain1.InsertChain(blocks)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if chain1.CurrentBlock().Hash() != blocks[len(blocks)-1].Hash() {
// 		t.Fatalf("fork chain err! actual hash %x,  want hash %x ", chain1.CurrentBlock().Hash(), blocks[len(blocks)-1].Hash())
// 	}
// }

// func TestFullTxChain(t *testing.T) {
// 	genesis, db, chain, st, err := newCanonical(t, tengine)
// 	if err != nil {
// 		t.Error("newCanonical err", err)
// 	}
// 	prods, ht := makeProduceAndTime(st, 100)
// 	_, _, _, err = makeNewChain(t, genesis, chain, &db, len(prods), ht, prods, makeTransferTx)
// 	if err != nil {
// 		t.Error("makeNewChain err", err)
// 	}
// }
