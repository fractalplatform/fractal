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
	"testing"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
	"github.com/stretchr/testify/assert"
)

type testBlockChain struct {
	blocks map[int]*types.Block
}

func newTestBlockChain(price *big.Int) *testBlockChain {
	blocks := make(map[int]*types.Block)
	for i := 0; i < 6; i++ {
		if i%2 == 0 {
			price = new(big.Int).Mul(price, big.NewInt(2))
		}
		block := &types.Block{Head: &types.Header{Number: big.NewInt(int64(i)), GasLimit: params.BlockGasLimit, GasUsed: uint64((i + 1) * 100000)}}
		if i < 5 { // blocks[5] no transaction
			action := types.NewAction(types.CreateContract, "gpotestname", "", 1,
				10, 10, nil, nil, nil)
			block.Txs = []*types.Transaction{types.NewTransaction(1, price, action)}
		}
		blocks[i] = block
	}

	return &testBlockChain{
		blocks: blocks,
	}
}
func (b *testBlockChain) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Header {
	if blockNr == rpc.LatestBlockNumber {
		return b.blocks[len(b.blocks)-1].Header()
	}
	return b.blocks[int(blockNr)].Header()
}

func (b *testBlockChain) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Block {
	return b.blocks[int(blockNr)]
}

func TestSuggestPrice(t *testing.T) {
	cfg := Config{
		Blocks:  5,
		Default: big.NewInt(1),
	}
	price := big.NewInt(1)
	gpo := NewOracle(newTestBlockChain(price), cfg)

	gasPrice, err := gpo.SuggestPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, price, gasPrice)

	// test the Minimum configuration

	cfg1 := Config{
		Blocks:  5,
		Default: big.NewInt(10),
	}
	gpo = NewOracle(newTestBlockChain(price), cfg1)

	gasPrice, err = gpo.SuggestPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, big.NewInt(10), gasPrice)

}
