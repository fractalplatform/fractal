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
	"container/heap"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/stretchr/testify/assert"
)

func getPriceTx(price *big.Int, nonce uint64) *types.Transaction {
	return types.NewTransaction(0, price, types.NewAction(
		types.Transfer,
		common.Name("fromtest"),
		common.Name("tototest"),
		nonce,
		uint64(3),
		uint64(2000),
		big.NewInt(1000),
		[]byte("test action"),
		[]byte("test remark"),
	))
}

func TestPriceHeap(t *testing.T) {
	var ph = new(priceHeap)
	heap.Init(ph)

	txs := []*types.Transaction{
		getPriceTx(big.NewInt(200), 2),
		getPriceTx(big.NewInt(200), 1),
		getPriceTx(big.NewInt(500), 4),
		getPriceTx(big.NewInt(100), 3),
		getPriceTx(big.NewInt(500), 5),
		getPriceTx(big.NewInt(100), 6),
	}

	//test sort,first sort by price,if the price is equal,sort by nonce,high nonce is worse.
	sortTxs := []*types.Transaction{
		getPriceTx(big.NewInt(100), 6),
		getPriceTx(big.NewInt(100), 3),
		getPriceTx(big.NewInt(200), 2),
		getPriceTx(big.NewInt(200), 1),
		getPriceTx(big.NewInt(500), 5),
		getPriceTx(big.NewInt(500), 4),
	}

	for _, v := range txs {
		heap.Push(ph, v)
	}

	for i := 0; i < len(txs); i++ {
		tx := heap.Pop(ph).(*types.Transaction)
		assert.Equal(t, sortTxs[i].Hash(), tx.Hash())
	}
}
