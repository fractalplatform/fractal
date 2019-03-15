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

package processor

import (
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// ChainContext supports retrieving headers and consensus parameters from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block

	// HasBlockAndState checks if a block and associated state trie is fully present
	// in the database or not, caching it if present.
	HasBlockAndState(hash common.Hash, number uint64) bool

	// HasBlock checks if a block is fully present in the database or not.
	HasBlock(hash common.Hash, number uint64) bool

	// CalcGasLimit computes the gas limit of the next block after parent.
	CalcGasLimit(parent *types.Block) uint64

	// StateAt retrieves a block state from the database by hash.
	StateAt(hash common.Hash) (*state.StateDB, error)

	// WriteBlockWithState writes the block and all associated state to the database.
	WriteBlockWithState(block *types.Block, receipts []*types.Receipt, state *state.StateDB) error
}

type EgnineContext interface {
	// Author retrieves the Ethereum address of the account that minted the given
	// block, which may be different from the header's coinbase if a consensus
	// engine is based on signatures.
	Author(header *types.Header) (common.Name, error)

	ProcessAction(chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) error

	GetDelegatedByTime(name string, timestamp uint64, state *state.StateDB) (*big.Int, error)
}

type EvmContext struct {
	ChainContext
	EgnineContext
}

// NewEVMContext creates a new context for use in the EVM.
func NewEVMContext(sender common.Name, assetID uint64, gasPrice *big.Int, header *types.Header, chain *EvmContext, author *common.Name) vm.Context {
	// If we don't have an explicit author (i.e. not mining), extract from the header
	var beneficiary common.Name
	if author == nil {
		beneficiary, _ = chain.Author(header) // Ignore error, we're past header validation
	} else {
		beneficiary = *author
	}
	return vm.Context{
		GetHash:            GetHashFn(header, chain),
		GetDelegatedByTime: chain.GetDelegatedByTime,
		Origin:             sender,
		AssetID:            assetID,
		Coinbase:           beneficiary,
		BlockNumber:        new(big.Int).Set(header.Number),
		Time:               new(big.Int).Set(header.Time),
		Difficulty:         new(big.Int).Set(header.Difficulty),
		GasLimit:           header.GasLimit,
		GasPrice:           new(big.Int).Set(gasPrice),
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.Header, chain ChainContext) func(n uint64) common.Hash {
	var cache map[uint64]common.Hash

	return func(n uint64) common.Hash {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64]common.Hash{
				ref.Number.Uint64() - 1: ref.ParentHash,
			}
		}
		// Try to fulfill the request from the cache
		if hash, ok := cache[n]; ok {
			return hash
		}
		// Not cached, iterate the blocks and cache the hashes
		for header := chain.GetHeader(ref.ParentHash, ref.Number.Uint64()-1); header != nil; header = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1) {
			cache[header.Number.Uint64()-1] = header.ParentHash
			if n == header.Number.Uint64()-1 {
				return header.ParentHash
			}
		}
		return common.Hash{}
	}
}
