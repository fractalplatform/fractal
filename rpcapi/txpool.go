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
	pending, queue := s.b.TxPool().Stats()
	return map[string]int{
		"pending": pending,
		"queued":  queue,
	}
}

// Content returns the transactions contained within the transaction pool.
func (s *PrivateTxPoolAPI) Content(fullTx bool) interface{} {
	content := map[string]map[string]map[string]interface{}{
		"pending": make(map[string]map[string]interface{}),
		"queued":  make(map[string]map[string]interface{}),
	}

	pending, queue := s.b.TxPool().Content()
	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]interface{})
		for _, tx := range txs {
			if fullTx {
				dump[fmt.Sprintf("%d", tx.GetActions()[0].Nonce())] = tx.NewRPCTransaction(common.Hash{}, 0, 0)
			} else {
				dump[fmt.Sprintf("%d", tx.GetActions()[0].Nonce())] = tx.Hash()
			}

		}
		content["pending"][account.String()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]interface{})
		for _, tx := range txs {
			if fullTx {
				dump[fmt.Sprintf("%d", tx.GetActions()[0].Nonce())] = tx.NewRPCTransaction(common.Hash{}, 0, 0)
			} else {
				dump[fmt.Sprintf("%d", tx.GetActions()[0].Nonce())] = tx.Hash()
			}
		}
		content["queued"][account.String()] = dump
	}
	return content
}

// PendingTransactions returns the pending transactions that are in the transaction pool
func (s *PrivateTxPoolAPI) PendingTransactions(fullTx bool) (interface{}, error) {
	pending, err := s.b.TxPool().Pending()
	if err != nil {
		return nil, err
	}

	var (
		txs       []*types.RPCTransaction
		txsHashes []common.Hash
	)

	for _, batch := range pending {
		for _, tx := range batch {
			if fullTx {
				txs = append(txs, tx.NewRPCTransaction(common.Hash{}, 0, 0))
			} else {
				txsHashes = append(txsHashes, tx.Hash())
			}
		}
	}
	if fullTx {
		return txs, nil
	}
	return txsHashes, nil
}

// GetPoolTransactions txpool returns the transaction for the given hash
func (s *PrivateTxPoolAPI) GetPoolTransactions(hashes []common.Hash) []*types.RPCTransaction {
	var txs []*types.RPCTransaction
	for _, hash := range hashes {
		if tx := s.b.TxPool().Get(hash); tx != nil {
			txs = append(txs, tx.NewRPCTransaction(common.Hash{}, 0, 0))
		}
	}
	return txs
}

// SetGasPrice set txpool gas price
func (s *PrivateTxPoolAPI) SetGasPrice(gasprice *big.Int) bool {
	return s.b.SetGasPrice(gasprice)
}
