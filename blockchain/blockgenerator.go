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

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// blockGenerator creates blocks for testing.
type blockGenerator struct {
	i       int
	parent  *types.Block
	header  *types.Header
	stateDB *state.StateDB
	am      *accountmanager.AccountManager

	gasPool  *common.GasPool
	txs      []*types.Transaction
	receipts []*types.Receipt

	config *params.ChainConfig
	engine *dpos.Dpos
	*BlockChain
}

// SetCoinbase sets the coinbase of the generated block.
func (bg *blockGenerator) SetCoinbase(name common.Name) {
	if bg.gasPool != nil {
		if len(bg.txs) > 0 {
			panic("coinbase must be set before adding transactions")
		}
		panic("coinbase can only be set once")
	}
	bg.header.Coinbase = name
	bg.gasPool = new(common.GasPool).AddGas(bg.header.GasLimit)
}

// TxNonce retrun nonce
func (bg *blockGenerator) TxNonce(name common.Name) uint64 {
	am, _ := accountmanager.NewAccountManager(bg.stateDB)
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
func (bg *blockGenerator) AddTxWithChain(tx *types.Transaction) {
	if bg.gasPool == nil {
		bg.SetCoinbase(bg.genesisBlock.Coinbase())
	}

	bg.stateDB.Prepare(tx.Hash(), common.Hash{}, len(bg.txs))

	receipt, _, err := bg.processor.ApplyTransaction(&bg.header.Coinbase, bg.gasPool, bg.stateDB, bg.header, tx, &bg.header.GasUsed, vm.Config{})
	if err != nil {
		panic(fmt.Sprintf(" apply transaction hash:%v ,err %v", tx.Hash().Hex(), err))
	}

	bg.txs = append(bg.txs, tx)
	bg.receipts = append(bg.receipts, receipt)
}

// CurrentHeader return current header
func (bg *blockGenerator) CurrentHeader() *types.Header {
	return bg.parent.Head
}
