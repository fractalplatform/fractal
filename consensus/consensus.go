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

package consensus

import (
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// IAPI returns the RPC APIs this consensus engine provides.
type IAPI interface {
	APIs(chain IChainReader) []rpc.API
}

// IChainReader defines a small collection of methods needed to access the local blockchain.
type IChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block

	// StateAt retrieves a block state from the database by hash.
	StateAt(hash common.Hash) (*state.StateDB, error)

	// WriteBlockWithState writes the block and all associated state to the database.
	WriteBlockWithState(block *types.Block, receipts []*types.Receipt, state *state.StateDB) (bool, error)

	// HasBlockAndState checks if a block and associated state trie is fully present
	// in the database or not, caching it if present.
	HasBlockAndState(hash common.Hash, number uint64) bool

	// HasBlock checks if a block is fully present in the database or not.
	HasBlock(hash common.Hash, number uint64) bool

	// FillForkID fills the current and next forkID
	FillForkID(header *types.Header, statedb *state.StateDB) error

	// ForkUpdate checks and records the fork information
	ForkUpdate(block *types.Block, statedb *state.StateDB) error
}

// IValidator implements consensus validator.
type IValidator interface {
	// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
	// that a new block should have.
	CalcDifficulty(chain IChainReader, time uint64, parent *types.Header) *big.Int

	// VerifySeal checks whether the crypto seal on a header is valid according to the consensus rules of the given engine.
	VerifySeal(chain IChainReader, header *types.Header) error
}

// IEngine is an algorithm agnostic consensus engine.
type IEngine interface {
	// Author retrieves the name of the account that minted the given block
	Author(header *types.Header) (common.Name, error)

	// Prepare initializes the consensus fields of a block header according to the rules of a particular engine. The changes are executed inline.
	Prepare(chain IChainReader, header *types.Header, txs []*types.Transaction, receipts []*types.Receipt, state *state.StateDB) error

	// Finalize assembles the final block.
	Finalize(chain IChainReader, header *types.Header, txs []*types.Transaction, receipts []*types.Receipt, state *state.StateDB) (*types.Block, error)

	// Seal generates a new block for the given input block with the local miner's seal place on top.
	Seal(chain IChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error)

	Engine() IEngine

	ProcessAction(fid uint64, number uint64, chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) ([]*types.InternalAction, error)

	GetDelegatedByTime(state *state.StateDB, candidate string, timestamp uint64) (stake *big.Int, err error)

	//GetLatestEpoch(state *state.StateDB) (epoch uint64, err error)

	//GetPrevEpoch(state *state.StateDB, epoch uint64) (pecho uint64, err error)

	//GetNextEpoch(state *state.StateDB, epoch uint64) (necho uint64, err error)

	GetEpoch(state *state.StateDB, t uint64, curEpoch uint64) (epoch uint64, time uint64, err error)

	GetActivedCandidateSize(state *state.StateDB, epoch uint64) (size uint64, err error)

	GetActivedCandidate(state *state.StateDB, epoch uint64, index uint64) (name string, stake *big.Int, totalVote *big.Int, counter uint64, actualCounter uint64, replace uint64, isbad bool, err error)

	GetCandidateStake(state *state.StateDB, epoch uint64, candidate string) (stake *big.Int, err error)

	GetVoterStake(state *state.StateDB, epoch uint64, voter string, candidate string) (stake *big.Int, err error)

	IAPI

	IValidator
}

// ITxProcessor is an Processor.
type ITxProcessor interface {
	// ApplyTransaction attempts to apply a transaction.
	ApplyTransaction(coinbase *common.Name, gp *common.GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error)
}

// ITxPool contains all currently known transactions.
type ITxPool interface {
	Pending() (map[common.Name][]*types.Transaction, error)
}

// IConsensus defines a small collection of methods needed for miner.
type IConsensus interface {
	IChainReader
	IEngine
	ITxProcessor
	ITxPool
}
