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

import (
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// BlockGenerator creates blocks for testing.
type BlockGenerator struct {
	i       int
	parent  *types.Block
	header  *types.Header
	statedb *state.StateDB
	am      *accountmanager.AccountManager

	gasPool  *common.GasPool
	txs      []*types.Transaction
	receipts []*types.Receipt

	config *params.ChainConfig
	engine consensus.IEngine
	*BlockChain
}

// SetCoinbase sets the coinbase of the generated block.
func (bg *BlockGenerator) SetCoinbase(name common.Name) {
	if bg.gasPool != nil {
		if len(bg.txs) > 0 {
			panic("coinbase must be set before adding transactions")
		}
		panic("coinbase can only be set once")
	}
	bg.header.Coinbase = name
	bg.gasPool = new(common.GasPool).AddGas(bg.header.GasLimit)
}

// OffsetTime modifies the time instance of a block
func (bg *BlockGenerator) OffsetTime(seconds int64) {
	bg.header.Time.Add(bg.header.Time, new(big.Int).SetInt64(seconds))
	if bg.header.Time.Cmp(bg.parent.Header().Time) <= 0 {
		panic(fmt.Sprintf("header time %d less than parent header time %v ", bg.header.Time.Uint64(), bg.parent.Time().Uint64()))
	}
	bg.header.Difficulty = bg.engine.CalcDifficulty(bg, bg.header.Time.Uint64(), bg.parent.Header())
}

// AddTx adds a transaction to the generated block.
func (bg *BlockGenerator) AddTx(tx *types.Transaction) {
	bg.AddTxWithChain(tx)
}

// TxNonce retrun nonce
func (bg *BlockGenerator) TxNonce(name common.Name) uint64 {
	am, _ := accountmanager.NewAccountManager(bg.statedb)
	a, err := am.GetAccountByName(name)
	if err != nil {
		panic(fmt.Sprintf("name: %v, GetTxNonce failed: %v", name, err))
	}
	if a == nil {
		panic("Account Not exist")
	}
	return a.GetNonce()
}

// AddTxWithChain adds a transaction to the generated block.
func (bg *BlockGenerator) AddTxWithChain(tx *types.Transaction) {
	if bg.gasPool == nil {
		bg.SetCoinbase(bg.genesisBlock.Coinbase())
	}

	bg.statedb.Prepare(tx.Hash(), common.Hash{}, len(bg.txs))

	receipt, _, err := bg.processor.ApplyTransaction(&bg.header.Coinbase, bg.gasPool, bg.statedb, bg.header, tx, &bg.header.GasUsed, vm.Config{})
	if err != nil {
		panic(fmt.Sprintf(" apply transaction hash:%v ,err %v", tx.Hash().Hex(), err))
	}

	bg.txs = append(bg.txs, tx)
	bg.receipts = append(bg.receipts, receipt)
}

// CurrentHeader return current header
func (bg *BlockGenerator) CurrentHeader() *types.Header {
	return bg.parent.Head
}
