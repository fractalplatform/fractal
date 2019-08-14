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
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
)

type backend interface {
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Header
	BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Block
}

//Config gas price oracle config
type Config struct {
	Blocks  int `mapstructure:"blocks"`
	Default *big.Int
}

// Oracle recommends gas prices based on the content of recent
// blocks.
type Oracle struct {
	backend      backend
	defaultPrice *big.Int
	lastHead     common.Hash
	lastPrice    *big.Int
	cacheLock    sync.RWMutex
	fetchLock    sync.Mutex

	checkBlocks int
}

// NewOracle returns a new oracle.
func NewOracle(backend backend, params Config) *Oracle {
	blocks := params.Blocks
	if blocks < 1 {
		blocks = 1
	}
	return &Oracle{
		defaultPrice: params.Default,
		backend:      backend,
		lastPrice:    params.Default,
		checkBlocks:  blocks,
	}
}

// SuggestPrice returns the recommended gas price.
func (gpo *Oracle) SuggestPrice(ctx context.Context) (*big.Int, error) {
	gpo.cacheLock.RLock()
	lastHead := gpo.lastHead
	lastPrice := gpo.lastPrice
	gpo.cacheLock.RUnlock()

	head := gpo.backend.HeaderByNumber(ctx, rpc.LatestBlockNumber)

	headHash := head.Hash()
	if headHash == lastHead {
		return lastPrice, nil
	}

	gpo.fetchLock.Lock()
	defer gpo.fetchLock.Unlock()

	// try checking the cache again, maybe the last fetch fetched what we need
	gpo.cacheLock.RLock()
	lastHead = gpo.lastHead
	lastPrice = gpo.lastPrice
	gpo.cacheLock.RUnlock()
	if headHash == lastHead {
		return lastPrice, nil
	}

	blockNum := head.Number.Uint64()
	ch := make(chan getBlockPricesResult, gpo.checkBlocks)
	sent := 0
	exp := 0
	prices := new(big.Int)
	weights := new(big.Int)

	for sent < gpo.checkBlocks && blockNum > 0 {
		go gpo.getBlockPrices(ctx, blockNum, ch)
		sent++
		exp++
		blockNum--
	}
	for exp > 0 {
		res := <-ch
		if res.err != nil {
			return lastPrice, res.err
		}
		exp--
		prices = new(big.Int).Add(prices, new(big.Int).Mul(res.price, res.weight))
		weights = new(big.Int).Add(weights, res.weight)
	}

	price := lastPrice
	if prices.Sign() > 0 {
		price = new(big.Int).Div(prices, weights)
	}

	gpo.cacheLock.Lock()
	gpo.lastHead = headHash
	gpo.lastPrice = price
	gpo.cacheLock.Unlock()
	return price, nil
}

type getBlockPricesResult struct {
	weight *big.Int
	price  *big.Int
	err    error
}

type transactionsByGasPrice []*types.Transaction

func (t transactionsByGasPrice) Len() int           { return len(t) }
func (t transactionsByGasPrice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t transactionsByGasPrice) Less(i, j int) bool { return t[i].GasPrice().Cmp(t[j].GasPrice()) < 0 }

// getBlockPrices calculates the lowest transaction gas price in a given block
// and sends it to the result channel. If the block is empty, price is nil.
func (gpo *Oracle) getBlockPrices(ctx context.Context, blockNum uint64, ch chan getBlockPricesResult) {
	block := gpo.backend.BlockByNumber(ctx, rpc.BlockNumber(blockNum))
	if block == nil {
		ch <- getBlockPricesResult{nil, nil, fmt.Errorf("not found block %v", blockNum)}
		return
	}

	blockTxs := block.Transactions()
	txs := make([]*types.Transaction, len(blockTxs))
	copy(txs, blockTxs)
	sort.Sort(transactionsByGasPrice(txs))
	for _, tx := range txs {
		sender := tx.GetActions()[0].Sender()
		if sender != block.Coinbase() {
			ch <- getBlockPricesResult{
				new(big.Int).Div(big.NewInt(int64(block.GasUsed()*1000)),
					big.NewInt(int64(block.GasLimit()))),
				tx.GasPrice(), nil}
			return
		}
	}
	// if block no transaction  the weight is the biggest，price is default price
	ch <- getBlockPricesResult{big.NewInt(1 * 1000), gpo.defaultPrice, nil}
}
