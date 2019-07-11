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

	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	mdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

const (
	NodeNum = 1024
	TxNum   = 1024
)

func TxsGen(num int) []*types.Transaction {
	txs := make([]*types.Transaction, num)
	action := types.NewAction(types.CallContract, "yanprogram", "lixiaopeng", 0, 0, 0, nil, nil, nil)
	for i := 0; i < num; i++ {
		txs[i] = types.NewTransaction(uint64(i), nil, action)
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
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(mdb.NewMemDatabase()))
		manager, _ = am.NewAccountManager(statedb)
		fname      = common.Name("fromname")
		tname      = common.Name("totestname")
		fkey       = generateAccount(t, fname, manager)
		_          = generateAccount(t, tname, manager)
		asset      = asset.NewAsset(statedb)
		trigger    = false
	)
	// issue asset
	if _, err := asset.IssueAsset("ft", 0, 0, "zz", new(big.Int).SetUint64(params.Fractal), 10, common.Name(""), fname, new(big.Int).SetUint64(params.Fractal), common.Name(""), ""); err != nil {
		t.Fatal(err)
	}
	// add balance
	if err := manager.AddAccountBalanceByName(fname, "ft", new(big.Int).SetUint64(params.Fractal)); err != nil {
		t.Fatal(err)
	}
	params.DefaultChainconfig.SysTokenID = 0
	blockchain := &testChain{&testBlockChain{statedb, 1000000000, new(event.Feed)}, fname, &trigger}
	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	nonce, err := pool.State().GetNonce(fname)
	if err != nil {
		t.Fatal("Invalid getNonce ", err)
	}
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

	nonce, err = pool.State().GetNonce(fname)
	if err != nil {
		t.Fatal("Invalid getNonce ", err)
	}
	if nonce != 2 {
		t.Fatalf("Invalid nonce, want 2, got %d", nonce)
	}

	// trigger state change in the background
	trigger = true
	pool.requestReset(nil, nil)

	_, err = pool.Pending()
	if err != nil {
		t.Fatalf("Could not fetch pending transactions: %v", err)
	}
	nonce, err = pool.State().GetNonce(fname)
	if err != nil {
		t.Fatal("Invalid getNonce ", err)
	}
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
