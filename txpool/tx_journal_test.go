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
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/stretchr/testify/assert"
)

func TestTxJournal(t *testing.T) {
	// Create a temporary file for the journal
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create temporary journal: %v", err)
	}
	txj := newTxJournal(file.Name())
	defer os.Remove(file.Name())

	txsMap := make(map[common.Name][]*types.Transaction)
	tx := newTx(big.NewInt(200), types.NewAction(
		types.Transfer,
		common.Name("fromtest"),
		common.Name("tototest"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(1000),
		[]byte("test action"),
		[]byte("test remark"),
	))
	txsMap[common.Name("test")] = []*types.Transaction{
		tx,
	}
	if err := txj.rotate(txsMap); err != nil {
		t.Fatalf("Failed to rotate transaction journal: %v", err)
	}

	if err := txj.close(); err != nil {
		t.Fatalf("Failed to close transaction journal: %v", err)
	}

	txjLoad := newTxJournal(file.Name())

	if err := txjLoad.load(func(txs []*types.Transaction) []error {
		assert.Equal(t, tx.Hash(), txs[0].Hash())
		return nil
	}); err != nil {
		t.Fatalf("Failed to close transaction journal: %v", err)
	}
}
