// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/snapshot"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

type Consensus struct {
}

// NewConsensus create new consenus.
func NewConsensus(state *state.StateDB) (IConsensus, error) {
	return &Consensus{}, nil
}

func (c *Consensus) Seal(block *types.Block) (*types.Block, error) {
	return block, nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to the consensus rules of the given engine.
func (c *Consensus) VerifySeal(header *types.Header) error {
	return nil
}

// Prepare initializes the consensus fields of a block header according to the rules of a particular engine. The changes are executed inline.
func (c *Consensus) Prepare(header *types.Header, txs []*types.Transaction, receipts []*types.Receipt, state *state.StateDB) error {
	return nil
}

// Finalize assembles the final block.
func (c *Consensus) Finalize(parent, header *types.Header, txs []*types.Transaction, receipts []*types.Receipt, state *state.StateDB) (*types.Block, error) {

	//snapshot
	snapshotInterval := 3600000 * uint64(time.Millisecond)
	parentTimeFormat := parent.Time.Uint64() / snapshotInterval * snapshotInterval
	currentTimeFormat := header.Time.Uint64() / snapshotInterval * snapshotInterval
	if parentTimeFormat != currentTimeFormat {
		snapshotManager := snapshot.NewSnapshotManager(state)
		if err := snapshotManager.SetSnapshot(currentTimeFormat, snapshot.BlockInfo{Number: header.Number.Uint64(), BlockHash: header.ParentHash, Timestamp: parentTimeFormat}); err != nil {
			return nil, err
		}
	}

	header.Difficulty = big.NewInt(1)
	header.Root = state.IntermediateRoot()

	return types.NewBlock(header, txs, receipts), nil
}
