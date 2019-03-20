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
	"math/rand"
	"testing"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/stretchr/testify/assert"
)

func TestStrictTxListAdd(t *testing.T) {
	// Generate a list of transactions to insert
	key, _ := crypto.GenerateKey()

	txs := make([]*types.Transaction, 1024)
	for i := 0; i < len(txs); i++ {
		txs[i] = transaction(uint64(i), "fromtest", "tototest", 0, key)
	}

	// Insert the transactions in a random order
	list := newTxList(true)
	for _, v := range rand.Perm(len(txs)) {
		list.Add(txs[v], 10)
	}

	// Verify internal state
	assert.Equal(t, len(list.txs.items), len(txs))

	for _, tx := range txs {
		assert.Equal(t, list.txs.items[tx.GetActions()[0].Nonce()], tx)
	}
}
