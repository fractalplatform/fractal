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
	"context"
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/common/hexutil"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
)

// PublicBlockChainAPI provides an API to access the blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{b}
}

// GetCurrentBlock returns current block.
func (s *PublicBlockChainAPI) GetCurrentBlock(fullTx bool) map[string]interface{} {
	return s.rpcOutputBlock(s.b.ChainConfig().ChainID, s.b.CurrentBlock(), true, fullTx)
}

// GetBlockByHash returns the requested block. When fullTx is true all transactions in the block are returned in full
// detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return s.rpcOutputBlock(s.b.ChainConfig().ChainID, block, true, fullTx), nil
	}
	return nil, err
}

// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) map[string]interface{} {
	block := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		response := s.rpcOutputBlock(s.b.ChainConfig().ChainID, block, true, fullTx)
		return response
	}
	return nil
}

// rpcOutputBlock uses the generalized output filler, then adds the total difficulty field, which requires
// a `PublicBlockchainAPI`.
func (s *PublicBlockChainAPI) rpcOutputBlock(chainID *big.Int, b *types.Block, inclTx bool, fullTx bool) map[string]interface{} {
	fields := RPCMarshalBlock(chainID, b, inclTx, fullTx)
	fields["totalDifficulty"] = s.b.GetTd(b.Hash())
	return fields
}

// GetTransactionByHash returns the transaction for the given hash
func (s *PublicBlockChainAPI) GetTransactionByHash(ctx context.Context, hash common.Hash) interface{} {
	// Try to return an already finalized transaction
	if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
		return tx.NewRPCTransaction(blockHash, blockNumber, index)
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.TxPool().Get(hash); tx != nil {
		return tx.NewRPCTransaction(common.Hash{}, 0, 0)
	}
	// Transaction unknown, return as such
	return nil
}

func (s *PublicBlockChainAPI) GetTransactions(ctx context.Context, hashes []common.Hash) []interface{} {
	var result []interface{}
	for i, hash := range hashes {
		if i > 2048 {
			break
		}
		if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
			result = append(result, tx.NewRPCTransaction(blockHash, blockNumber, index))
		}
	}
	return result
}

// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (s *PublicBlockChainAPI) GetTransactionReceipt(ctx context.Context, hash common.Hash) (*types.RPCReceipt, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}

	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}
	receipt := receipts[index]
	return receipt.NewRPCReceipt(blockHash, blockNumber, index, tx), nil
}

func (s *PublicBlockChainAPI) GetBlockAndResultByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.BlockAndResult {
	r := s.b.GetBlockDetailLog(ctx, blockNr)
	if r == nil {
		return nil
	}
	block := s.GetBlockByNumber(ctx, blockNr, true)
	r.Block = block
	return r
}

type GetRangeTxArgs struct {
	BlockNumber rpc.BlockNumber
	BackCount   uint64
}

func (a *GetRangeTxArgs) CheckArgs(s *PublicBlockChainAPI) error {
	cur := s.b.CurrentBlock().Number().Uint64()
	if a.BlockNumber == rpc.LatestBlockNumber {
		a.BlockNumber = rpc.BlockNumber(cur)
	}

	if uint64(a.BlockNumber.Int64()) > cur {
		return fmt.Errorf("Block Number %v bigger than current block %v", a.BlockNumber, cur)
	}

	if a.BackCount > uint64(a.BlockNumber.Int64())+1 {
		a.BackCount = uint64(a.BlockNumber.Int64()) + 1
	}

	if a.BackCount > 128 {
		a.BackCount = 128
	}

	return nil
}

type GetInternalTxByAccount struct {
	Account string
	*GetRangeTxArgs
}

type GetInternalTxByBloom struct {
	Bloom hexutil.Bytes
	*GetRangeTxArgs
}

// GetInternalTxByAccount return all logs of internal txs, sent from or received by a specific account
func (s *PublicBlockChainAPI) GetInternalTxByAccount(ctx context.Context, args *GetInternalTxByAccount) ([]*types.DetailTx, error) {
	// check input arguments
	if err := args.CheckArgs(s); err != nil {
		return nil, err
	}

	filterFn := func(name string) bool {
		return name == args.Account
	}
	return s.b.GetDetailTxByFilter(ctx, filterFn, args.GetRangeTxArgs), nil
}

// GetInternalTxByBloom return all logs of internal txs, filtered by a bloomByte
func (s *PublicBlockChainAPI) GetInternalTxByBloom(ctx context.Context, args *GetInternalTxByBloom) ([]*types.DetailTx, error) {
	// check input arguments
	if err := args.CheckArgs(s); err != nil {
		return nil, err
	}

	bloom := types.BytesToBloom(args.Bloom)
	filterFn := func(name string) bool {
		return types.BloomLookup(bloom, new(big.Int).SetBytes([]byte(name)))
	}
	return s.b.GetDetailTxByFilter(ctx, filterFn, args.GetRangeTxArgs), nil
}

// GetInternalTxByHash return logs of internal txs include by a transcastion
func (s *PublicBlockChainAPI) GetInternalTxByHash(ctx context.Context, hash common.Hash) (*types.DetailTx, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}

	detailTxs := rawdb.ReadDetailTxs(s.b.ChainDb(), blockHash, blockNumber)
	if len(detailTxs) <= int(index) {
		return nil, fmt.Errorf("")
	}

	return detailTxs[index], nil
}

func (s *PublicBlockChainAPI) GetBadBlocks(ctx context.Context, fullTx bool) ([]map[string]interface{}, error) {
	blocks, err := s.b.GetBadBlocks(ctx)
	if len(blocks) != 0 {
		badBlocks := make([]map[string]interface{}, len(blocks))
		for i, b := range blocks {
			badBlocks[i] = s.rpcOutputBlock(s.b.ChainConfig().ChainID, b, true, fullTx)
		}
		return badBlocks, nil
	}
	return nil, err
}

// PrivateBlockChainAPI provides an API to access the blockchain.
// It offers only methods that operate on private data that is freely available to anyone.
type PrivateBlockChainAPI struct {
	b Backend
}

// NewPrivateBlockChainAPI creates a new blockchain API.
func NewPrivateBlockChainAPI(b Backend) *PrivateBlockChainAPI {
	return &PrivateBlockChainAPI{b}
}

// SetStatePruning start blockchain state prune
func (s *PrivateBlockChainAPI) SetStatePruning(enable bool) types.BlockState {
	prestatus, number := s.b.SetStatePruning(enable)
	return types.BlockState{PreStatePruning: prestatus, CurrentNumber: number}
}
