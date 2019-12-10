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

package txpool

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/params"
	pm "github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
)

const (
	NodeNum = 1024
	TxNum   = 1024
)

func TxsGen(num int) []*types.Transaction {
	txs := make([]*types.Transaction, num)

	for i := 0; i < num; i++ {
		ctx, _ := envelope.NewContractTx(envelope.CreateContract, "sender", "receipt", uint64(i), 0, 0, 100000,
			big.NewInt(100), big.NewInt(1000), []byte("payload"), []byte("remark"))
		txs[i] = types.NewTransaction(ctx)
	}
	return txs
}

func NodesGen(num int) []string {
	nodes := make([]string, num)
	for i := 0; i < num; i++ {
		nodes[i] = fmt.Sprintf("Node%04x", i)
	}
	return nodes
}

func TestBloom(t *testing.T) {
	var (
		txs     = TxsGen(TxNum)
		nodeIDs = NodesGen(NodeNum)
		cache   = &txsCache{}
	)

	for _, tx := range txs {
		for _, node := range nodeIDs[:len(nodeIDs)/2] {
			cache.addTx(tx, nil, node)
		}
	}

	var hadTx *types.Transaction
	for _, tx := range txs {
		if cache.hadTx(tx) {
			hadTx = tx
			break
		}
	}

	if hadTx == nil {
		t.Fatalf("no transaction cached!\n")
	}

	var unkownNode string
	for i, node := range nodeIDs {
		if !cache.txHadPath(hadTx, node) {
			unkownNode = node
			if i < len(nodeIDs)/2 {
				t.Fatalf("%d %x %x\n", i, hadTx.Hash(), *cache.getTxBloom(hadTx))
			}
		}
	}

	if len(unkownNode) == 0 {
		t.Fatalf("didn't find unkown node")
	}

	target := cache.getTarget(hadTx.Hash())
	oldBloom := target.bloom
	refBloom := cache.getTxBloom(hadTx)
	cpyBloom := cache.copyTxBloom(hadTx, &types.Bloom{})
	if (*refBloom != *cpyBloom) || (oldBloom != refBloom) {
		t.Fatalf("Bloom not equal")
	}
	target.reset(hadTx.Hash(), nil)

	cache.addTx(hadTx, oldBloom, unkownNode)

	for i, node := range nodeIDs {
		if !cache.txHadPath(hadTx, node) {
			if i < len(nodeIDs)/2 || node == unkownNode {
				t.Fatalf("%d %x %x\n", i, hadTx.Hash(), *cache.getTxBloom(hadTx))
			}
		}
	}
}

func TestP2PTxMsg(t *testing.T) {
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
		pm         = pm.NewPM(statedb)
		fname      = "fromname"
		tname      = "totestname"
		fkey       = generateAccount(t, fname, pm)
		_          = generateAccount(t, tname, pm)
	)
	// issue asset
	if _, err := pm.IssueAsset(fname, "ft", "zz", new(big.Int).SetUint64(params.Fractal),
		10, "", fname, new(big.Int).SetUint64(params.Fractal), "", pm); err != nil {
		t.Fatal(err)
	}

	// add balance
	if err := pm.TransferAsset(fname, tname, 0, new(big.Int).SetUint64(params.Fractal)); err != nil {
		t.Fatal(err)
	}
	params.DefaultChainconfig.SysTokenID = 0
	blockchain := &testBlockChain{statedb, 1000000000, new(event.Feed)}
	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	nonce, _ := pool.State().GetNonce(fname)
	if nonce != 0 {
		t.Fatalf("Invalid nonce, want 0, got %d", nonce)
	}

	txs := []*TransactionWithPath{
		&TransactionWithPath{
			Tx:    transaction(0, fname, tname, 109000, fkey),
			Bloom: &types.Bloom{},
		},
		&TransactionWithPath{
			Tx:    transaction(1, fname, tname, 109000, fkey),
			Bloom: &types.Bloom{},
		},
	}

	event.SendTo(event.NewLocalStation("test", nil), nil, event.P2PTxMsg, txs)
	for {
		if pending, _ := pool.Stats(); pending > 0 {
			break
		}
	}

	nonce, _ = pool.State().GetNonce(fname)
	if nonce != 2 {
		t.Fatalf("Invalid nonce, want 2, got %d", nonce)
	}

	pool.requestReset(nil, nil)

	_, err := pool.Pending()
	if err != nil {
		t.Fatalf("Could not fetch pending transactions: %v", err)
	}
	nonce, _ = pool.State().GetNonce(fname)
	if nonce != 2 {
		t.Fatalf("Invalid nonce, want 2, got %d", nonce)
	}
}

func TestBloombits(t *testing.T) {
	bp := &bloomPath{
		hash:  common.Hash{},
		bloom: &types.Bloom{},
	}
	for i := 0; i < len(bp.bloom)*8; i++ {
		b := big.NewInt(1)
		b.Lsh(b, uint(i))
		b.Sub(b, big.NewInt(1))
		bp.bloom.SetBytes(b.Bytes())
		if bp.getBloomSetBits() != i {
			t.Fatalf("error:%d", i)
		}
	}
}

func TestTTL(t *testing.T) {
	var (
		tx    = TxsGen(1)[0]
		cache = &txsCache{}
	)
	i := 0
	NewName := func() string {
		i++
		return fmt.Sprintf("Node%x", i)
	}
	for {
		cache.addTx(tx, nil, NewName())
		target := cache.getTarget(tx.Hash())
		if target.getBloomSetBits() > len(*target.bloom)*3 {
			break
		}
	}
	cache.ttlCheck(tx)
	target := cache.getTarget(tx.Hash())
	if target.getBloomSetBits() != 0 {
		t.Fatal("ttl check failed!")
	}
}
