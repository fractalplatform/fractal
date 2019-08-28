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
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

func TestLookup(t *testing.T) {
	txs := make([]*types.Transaction, 1024)
	txsmap := make(map[common.Hash]*types.Transaction, 1024)
	lk := newTxLookup()
	for i := 0; i < len(txs); i++ {
		tx := types.NewTransaction(uint64(i), big.NewInt(int64(i)), types.NewAction(types.CreateAccount, common.StrToName(fmt.Sprintf("from%d", i)), common.StrToName(fmt.Sprintf("to%d", i)), uint64(i), uint64(i), uint64(i), big.NewInt(int64(i)), []byte(fmt.Sprintf("from%d", i)), []byte(fmt.Sprintf("from%d", i))))
		lk.Add(tx)
		txs[i] = tx
		txsmap[tx.Hash()] = tx
	}
	assert.Equal(t, len(txs), lk.Count())
	lk.Range(func(hash common.Hash, tx *types.Transaction) bool {
		assert.Equal(t, txsmap[hash], tx)
		return true
	})
	for _, tx := range txs {
		assert.Equal(t, tx, lk.Get(tx.Hash()))
		lk.Remove(tx.Hash())
	}
	assert.Equal(t, 0, lk.Count())
}
