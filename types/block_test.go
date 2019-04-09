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

package types

import (
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
	"github.com/stretchr/testify/assert"
)

var (
	testHeader = &Header{
		ParentHash: common.HexToHash("0a5843ac1cb04865017cb35a57b50b07084e5fcee39b5acadade33149f4fff9e"),
		Coinbase:   common.Name("cpinbase"),
		Root:       common.HexToHash("ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017"),
		Difficulty: big.NewInt(131072),
		Number:     big.NewInt(100),
		GasLimit:   uint64(3141592),
		GasUsed:    uint64(21000),
		Time:       big.NewInt(1426516743),
		Extra:      []byte("test Header"),
	}
	testBlock = &Block{
		Head: testHeader,
		Txs:  []*Transaction{testTx},
	}
)

func TestBlockEncodeRLPAndDecodeRLP(t *testing.T) {
	bytes, err := testBlock.EncodeRLP()
	if err != nil {
		t.Fatal(err)
	}

	newBlock := &Block{}
	if err := newBlock.DecodeRLP(bytes); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testBlock.Transactions()[0].GasAssetID(), newBlock.Transactions()[0].GasAssetID())
	assert.Equal(t, testBlock.Transactions()[0].GasPrice(), newBlock.Transactions()[0].GasPrice())
	assert.Equal(t, testBlock.Hash(), newBlock.Hash())
}

func TestBlockForkID(t *testing.T) {
	testHeader.WithForkID(1, 2)
	testBlock := NewBlockWithHeader(testHeader)
	bytes, err := testBlock.EncodeRLP()
	if err != nil {
		t.Fatal(err)
	}

	newBlock := &Block{}
	assert.NoError(t, newBlock.DecodeRLP(bytes))
	assert.Equal(t, testBlock.CurForkID(), newBlock.CurForkID())
	assert.Equal(t, testBlock.NextForkID(), newBlock.NextForkID())
}

func TestBlockHeaderEncodeRLPAndDecodeRLP(t *testing.T) {
	bytes, err := rlp.EncodeToBytes(testHeader)
	if err != nil {
		t.Fatal(err)
	}
	newHeader := &Header{}

	if err := rlp.DecodeBytes(bytes, newHeader); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testHeader.Hash(), newHeader.Hash())
}

func TestSortByNumber(t *testing.T) {

	block0 := NewBlock(&Header{Number: big.NewInt(0)}, nil, nil)
	block1 := NewBlock(&Header{Number: big.NewInt(1)}, nil, nil)
	block2 := NewBlock(&Header{Number: big.NewInt(2)}, nil, nil)

	blocks := []*Block{block1, block2, block0}

	BlockBy(Number).Sort(blocks)

	for i := 0; i < len(blocks); i++ {
		assert.Equal(t, uint64(i), blocks[i].NumberU64())
	}
}
