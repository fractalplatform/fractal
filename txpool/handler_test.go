package txpool

import (
	"fmt"
	"testing"

	"github.com/fractalplatform/fractal/types"
)

const (
	NodeNum = 1024
	TxNum   = 1024
)

func TxsGen(num int) []*types.Transaction {
	txs := make([]*types.Transaction, num)
	action := types.NewAction(types.CallContract, "yanprogram", "lixiaopeng", 0, 0, 0, nil, nil)
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
