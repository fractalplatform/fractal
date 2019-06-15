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

package rpcapi

import (
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// PrivateTxPoolAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PrivateTxPoolAPI struct {
	b Backend
}

// NewPrivateTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPrivateTxPoolAPI(b Backend) *PrivateTxPoolAPI {
	return &PrivateTxPoolAPI{b}
}

// Status returns the number of pending and queued transaction in the pool.
func (s *PrivateTxPoolAPI) Status() map[string]int {
	pending, queue := s.b.Stats()
	return map[string]int{
		"pending": pending,
		"queued":  queue,
	}
}

// Content returns the transactions contained within the transaction pool.
func (s *PrivateTxPoolAPI) Content() map[string]map[string]map[string]*types.RPCTransaction {
	content := map[string]map[string]map[string]*types.RPCTransaction{
		"pending": make(map[string]map[string]*types.RPCTransaction),
		"queued":  make(map[string]map[string]*types.RPCTransaction),
	}
	pending, queue := s.b.TxPoolContent()
	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]*types.RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.GetActions()[0].Nonce())] = tx.NewRPCTransaction(common.Hash{}, 0, 0)
		}
		content["pending"][account.String()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]*types.RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.GetActions()[0].Nonce())] = tx.NewRPCTransaction(common.Hash{}, 0, 0)
		}
		content["queued"][account.String()] = dump
	}
	return content
}

// SetGasPrice set gas price
func (s *PrivateTxPoolAPI) SetGasPrice(gasprice *big.Int) bool {
	return s.b.SetGasPrice(gasprice)
}
