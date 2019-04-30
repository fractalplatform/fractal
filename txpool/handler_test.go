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
	if err := asset.IssueAsset("ft", 0, "zz", new(big.Int).SetUint64(params.Fractal), 10, common.Name(""), fname, new(big.Int).SetUint64(params.Fractal), common.Name(""), ""); err != nil {
		t.Fatal(err)
	}

	// add balance
	if err := manager.AddAccountBalanceByName(fname, "ft", new(big.Int).SetUint64(params.Fractal)); err != nil {
		t.Fatal(err)
	}
	params.DefaultChainconfig.SysTokenID = 1
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
			Tx:    transaction(0, fname, tname, 100000, fkey),
			Bloom: &types.Bloom{},
		},
		&TransactionWithPath{
			Tx:    transaction(1, fname, tname, 100000, fkey),
			Bloom: &types.Bloom{},
		},
	}
	event.SendTo(event.NewLocalStation("test", nil), nil, event.P2PTxMsg, txs)
	for {
		if pending, quened := pool.Stats(); pending > 0 || quened > 0 {
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
	pool.lockedReset(nil, nil)

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
