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
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// txList is a "list" of transactions belonging to an account, sorted by account
// nonce. The same type can be used both for storing contiguous transactions for
// the executable/pending queue; and for storing gapped transactions for the non-
// executable/future queue, with minor behavioral changes.
type txList struct {
	strict     bool         // Whether nonces are strictly continuous or not
	txs        *txSortedMap // Heap indexed sorted hash map of the transactions
	gascostcap *big.Int     // Price of the highest gas costing transaction (reset only if exceeds balance)
	gascap     uint64       // Gas limit of the highest spending transaction (reset only if exceeds block limit)
}

// newTxList create a new transaction list for maintaining nonce-indexable fast,
// gapped, sortable transaction lists.
func newTxList(strict bool) *txList {
	return &txList{
		strict:     strict,
		txs:        newTxSortedMap(),
		gascostcap: new(big.Int),
	}
}

// Overlaps returns whether the transaction specified has the same nonce as one
// already contained within the list.
func (l *txList) Overlaps(tx *types.Transaction) bool {
	// todo change action
	return l.txs.Get(tx.GetActions()[0].Nonce()) != nil
}

// Add tries to insert a new transaction into the list, returning whether the
// transaction was accepted, and if yes, any previous transaction it replaced.
func (l *txList) Add(tx *types.Transaction, priceBump uint64) (bool, *types.Transaction) {
	// If there's an older better transaction, abort
	// todo change action nonce
	old := l.txs.Get(tx.GetActions()[0].Nonce())
	if old != nil {
		threshold := new(big.Int).Div(new(big.Int).Mul(old.GasPrice(), big.NewInt(100+int64(priceBump))), big.NewInt(100))
		// Have to ensure that the new gas price is higher than the old gas
		// price as well as checking the percentage threshold to ensure that
		// this is accurate for low (Wei-level) gas price replacements
		if old.GasPrice().Cmp(tx.GasPrice()) >= 0 || threshold.Cmp(tx.GasPrice()) > 0 {
			return false, nil
		}
	}

	// Otherwise overwrite the old transaction with the current one
	l.txs.Put(tx)

	if cost := tx.Cost(); l.gascostcap.Cmp(cost) < 0 {
		l.gascostcap = cost
	}

	// todo change action
	if gas := tx.GetActions()[0].Gas(); l.gascap < gas {
		l.gascap = gas
	}
	return true, old
}

// Forward removes all transactions from the list with a nonce lower than the
// provided threshold. Every removed transaction is returned for any post-removal
// maintenance.
func (l *txList) Forward(threshold uint64) []*types.Transaction {
	return l.txs.Forward(threshold)
}

// Filter removes all transactions from the list with a cost or gas limit or no permissions higher
// than the provided thresholds. Every removed transaction is returned for any
// post-removal maintenance. Strict-mode invalidated transactions are also
// returned.
func (l *txList) Filter(costLimit *big.Int, gasLimit uint64, signer types.Signer,
	getBalance func(name common.Name, assetID uint64, typeID uint64) (*big.Int, error),
	recoverTx func(signer types.Signer, tx *types.Transaction) error) ([]*types.Transaction, []*types.Transaction) {
	// If all transactions are below the threshold, short circuit
	if l.gascostcap.Cmp(costLimit) > 0 {
		l.gascostcap = new(big.Int).Set(costLimit) // Lower the caps to the thresholds
	}
	if l.gascap > gasLimit {
		l.gascap = gasLimit
	}

	// Filter out all the transactions above the account's funds
	removed := l.txs.Filter(func(tx *types.Transaction) bool {
		act := tx.GetActions()[0]
		balance, err := getBalance(act.Sender(), act.AssetID(), 0)
		if err != nil {
			log.Warn("txpool filter get balance failed", "err", err)
			return true
		}

		if err := recoverTx(signer, tx); err != nil {
			log.Warn("txpool filter recover transaction failed", "err", err)
			return true
		}

		// todo change action
		return act.Value().Cmp(balance) > 0 || tx.Cost().Cmp(costLimit) > 0 || act.Gas() > gasLimit
	})

	// If the list was strict, filter anything above the lowest nonce
	var invalids []*types.Transaction

	if l.strict && len(removed) > 0 {
		lowest := uint64(math.MaxUint64)
		for _, tx := range removed {
			if nonce := tx.GetActions()[0].Nonce(); lowest > nonce {
				lowest = nonce
			}
		}
		// todo change action
		invalids = l.txs.Filter(func(tx *types.Transaction) bool { return tx.GetActions()[0].Nonce() > lowest })
	}
	return removed, invalids
}

// Cap places a hard limit on the number of items, returning all transactions
// exceeding that limit.
func (l *txList) Cap(threshold int) []*types.Transaction {
	return l.txs.Cap(threshold)
}

// Remove deletes a transaction from the maintained list, returning whether the
// transaction was found, and also returning any transaction invalidated due to
// the deletion (strict mode only).
func (l *txList) Remove(tx *types.Transaction) (bool, []*types.Transaction) {
	// Remove the transaction from the set
	// todo change action
	nonce := tx.GetActions()[0].Nonce()
	if removed := l.txs.Remove(nonce); !removed {
		return false, nil
	}
	// In strict mode, filter out non-executable transactions
	if l.strict {
		// todo change action
		return true, l.txs.Filter(func(tx *types.Transaction) bool { return tx.GetActions()[0].Nonce() > nonce })
	}
	return true, nil
}

// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
func (l *txList) Ready(start uint64) []*types.Transaction {
	return l.txs.Ready(start)
}

// Len returns the length of the transaction list.
func (l *txList) Len() int {
	return l.txs.Len()
}

// Empty returns whether the list of transactions is empty or not.
func (l *txList) Empty() bool {
	return l.Len() == 0
}

// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifications are made to the contents.
func (l *txList) Flatten() []*types.Transaction {
	return l.txs.Flatten()
}
