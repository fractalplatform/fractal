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
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/stretchr/testify/assert"
)

type headerMap map[uint64]*types.Header

func (h headerMap) getHeader(num uint64) *types.Header {
	return h[num]
}
func (h headerMap) setHeader(num uint64, header *types.Header) {
	h[num] = header
}

func TestForkController1(t *testing.T) {
	var (
		testcfg    = &ForkConfig{ForkBlockNum: 10, Forkpercentage: 80}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db))
		hm         = make(headerMap)
	)
	if err := initForkController(params.DefaultChainconfig.ChainName, statedb, 0); err != nil {
		t.Fatal(err)
	}
	fc := NewForkController(testcfg, params.DefaultChainconfig)

	var number int64
	for j := 0; j < 1; j++ {
		for i := 0; i < 8; i++ {
			block := &types.Block{Head: &types.Header{Number: big.NewInt(number)}}
			block.Head.WithForkID(uint64(j), uint64(j))
			hm.setHeader(uint64(number), block.Head)
			assert.NoError(t, fc.checkForkID(block.Header(), statedb))
			assert.NoError(t, fc.update(block, statedb, hm.getHeader))
			number++
		}

		for i := 0; i < 8; i++ {
			block := &types.Block{Head: &types.Header{Number: big.NewInt(number)}}
			block.Head.WithForkID(uint64(j), uint64(j+1))
			hm.setHeader(uint64(number), block.Head)
			assert.NoError(t, fc.checkForkID(block.Header(), statedb))
			assert.NoError(t, fc.update(block, statedb, hm.getHeader))
			number++
		}

		id, _, err := fc.currentForkID(statedb)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint64(j+1), id)
	}
}

func TestForkController2(t *testing.T) {
	var (
		testcfg    = &ForkConfig{ForkBlockNum: 10, Forkpercentage: 80}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db))
		hm         = make(headerMap)
	)
	if err := initForkController(params.DefaultChainconfig.ChainName, statedb, 0); err != nil {
		t.Fatal(err)
	}
	fc := NewForkController(testcfg, params.DefaultChainconfig)

	var number int64
	for j := 0; j < 1; j++ {
		for i := 0; i < 16; i++ {
			block := &types.Block{Head: &types.Header{Number: big.NewInt(number)}}
			if i%2 == 0 {
				block.Head.WithForkID(uint64(j), uint64(j))
			} else {
				block.Head.WithForkID(uint64(j), uint64(j+2))
			}
			hm.setHeader(uint64(number), block.Head)
			assert.NoError(t, fc.checkForkID(block.Header(), statedb))
			assert.NoError(t, fc.update(block, statedb, hm.getHeader))
			number++
		}

		for i := 0; i < 2; i++ {
			block := &types.Block{Head: &types.Header{Number: big.NewInt(number)}}
			block.Head.WithForkID(uint64(j), uint64(j+1))
			hm.setHeader(uint64(number), block.Head)
			assert.NoError(t, fc.checkForkID(block.Header(), statedb))
			assert.NoError(t, fc.update(block, statedb, hm.getHeader))
			number++
		}

		for i := 0; i < 8; i++ {
			block := &types.Block{Head: &types.Header{Number: big.NewInt(number)}}
			block.Head.WithForkID(uint64(j), uint64(j+2))
			hm.setHeader(uint64(number), block.Head)
			assert.NoError(t, fc.checkForkID(block.Header(), statedb))
			assert.NoError(t, fc.update(block, statedb, hm.getHeader))
			number++
		}

		// after fork success
		for i := 0; i < 6; i++ {
			block := &types.Block{Head: &types.Header{Number: big.NewInt(number)}}
			block.Head.WithForkID(uint64(j+2), uint64(j+2))
			hm.setHeader(uint64(number), block.Head)
			assert.NoError(t, fc.checkForkID(block.Header(), statedb))
			assert.NoError(t, fc.update(block, statedb, hm.getHeader))
			number++
		}

		id, _, err := fc.currentForkID(statedb)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint64(j+2), id)
	}
}

func TestUpdateDifferentForkBlock(t *testing.T) {
	var (
		testcfg    = &ForkConfig{ForkBlockNum: 10, Forkpercentage: 80}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db))
		hm         = make(headerMap)
	)
	if err := initForkController(params.DefaultChainconfig.ChainName, statedb, 0); err != nil {
		t.Fatal(err)
	}

	fc := NewForkController(testcfg, params.DefaultChainconfig)
	var number int64
	for j := 0; j < 2; j++ {
		for i := 0; i < 7; i++ {
			block := &types.Block{Head: &types.Header{Number: big.NewInt(number)}}
			block.Head.WithForkID(uint64(0), uint64(j+1))
			hm.setHeader(uint64(number), block.Head)
			assert.NoError(t, fc.checkForkID(block.Header(), statedb))
			assert.NoError(t, fc.update(block, statedb, hm.getHeader))
			number++

			info, err := fc.getForkInfo(statedb)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, uint64(i+1), info.NextForkIDBlockNum)
		}
	}

}

func TestFillForkID(t *testing.T) {
	var (
		testcfg    = &ForkConfig{ForkBlockNum: 10, Forkpercentage: 80}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db))
	)
	if err := initForkController(params.DefaultChainconfig.ChainName, statedb, 0); err != nil {
		t.Fatal(err)
	}

	fc := NewForkController(testcfg, params.DefaultChainconfig)

	header := &types.Header{Number: big.NewInt(0)}

	assert.NoError(t, fc.fillForkID(header, statedb))

	curForkID, nextForkID, err := fc.currentForkID(statedb)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, curForkID, header.CurForkID())
	assert.Equal(t, nextForkID, header.NextForkID())
}
