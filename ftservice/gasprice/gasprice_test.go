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

package gasprice

import (
	"context"
	"math/big"
	"math/rand"
	"sort"
	"testing"

	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
	"github.com/stretchr/testify/assert"
)

type testBlockChain struct {
	blocks map[int]*types.Block
}

func newTestBlockChain(prices *big.Int) *testBlockChain {
	blocks := make(map[int]*types.Block)
	for i := 0; i < 100; i++ {
		block := &types.Block{Head: &types.Header{Number: big.NewInt(int64(i))}}
		action := types.NewAction(types.CreateContract, "gpotestname", "", 1,
			10, 10, nil, nil, nil)
		block.Txs = []*types.Transaction{types.NewTransaction(1, prices, action)}
		blocks[i] = block
	}
	return &testBlockChain{
		blocks: blocks,
	}
}
func (b *testBlockChain) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber {
		return b.blocks[len(b.blocks)-1].Header(), nil
	}
	return b.blocks[int(blockNr)].Header(), nil
}

func (b *testBlockChain) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	return b.blocks[int(blockNr)], nil
}

func TestSuggestPrice(t *testing.T) {
	cfg := Config{
		Blocks:     100,
		Percentile: 60,
		Default:    big.NewInt(1),
	}
	price := big.NewInt(1)
	gpo := NewOracle(newTestBlockChain(price), cfg)

	gasPrice, err := gpo.SuggestPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, price, gasPrice)

}

func TestSortPrice(t *testing.T) {
	pricesLen := 10
	prices := make([]*big.Int, pricesLen)
	for i := 0; i < pricesLen; i++ {
		prices[i] = big.NewInt(rand.Int63())
	}
	sort.Sort(bigIntArray(prices))
	for k, v := range prices {
		if k == (pricesLen - 1) {
			if v.Cmp(prices[k-1]) < 0 {
				t.Errorf("prices[%d]=%v must > prices[%d]=%v", k, v, pricesLen-1, prices[pricesLen-1])
			}
			break
		}
		if v.Cmp(prices[k+1]) > 0 {
			t.Errorf("prices[%d]=%v must > prices[%d]=%v", k+1, prices[k+1], k, v)
		}
	}
}
