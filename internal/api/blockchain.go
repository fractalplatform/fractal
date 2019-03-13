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

package api

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
)

// PublicBlockChainAPI provides an API to access the Ethereum blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
const (
	defaultGasPrice = params.GWei
)

type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new Ethereum blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{b}
}

// GetCurrentBlock returns cureent block.
func (s *PublicBlockChainAPI) GetCurrentBlock(fullTx bool) map[string]interface{} {
	block := s.b.CurrentBlock()
	response := s.rpcOutputBlock(s.b.ChainConfig().ChainID, block, true, fullTx)
	return response
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
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		response := s.rpcOutputBlock(s.b.ChainConfig().ChainID, block, true, fullTx)
		if blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}
func (s *PublicBlockChainAPI) GetTxNumByBlockHash(ctx context.Context, blockHash common.Hash) (int, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return len(block.Transactions()), nil
	}
	return 0, err
}
func (s *PublicBlockChainAPI) GetTxNumByBlockNum(ctx context.Context, blockNr rpc.BlockNumber) (int, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		return len(block.Transactions()), nil
	}
	return 0, err
}
func (s *PublicBlockChainAPI) GetTotalTxNumByBlockHash(ctx context.Context, blockHash common.Hash, lookbackNum uint64) (*big.Int, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		txNum := len(block.Transactions())
		totalTxNum := big.NewInt(int64(txNum))
		height := block.Number().Uint64()
		for i := height - 1; i >= 0 && i > height-lookbackNum; i-- {
			block, err := s.b.BlockByNumber(ctx, rpc.BlockNumber(i))
			if block != nil {
				totalTxNum = totalTxNum.Add(totalTxNum, big.NewInt(int64(len(block.Transactions()))))
			} else {
				return nil, err
			}
		}
		return totalTxNum, nil
	}

	return nil, err
}
func (s *PublicBlockChainAPI) GetTotalTxNumByBlockNum(ctx context.Context, blockNr rpc.BlockNumber, lookbackNum uint64) (*big.Int, error) {
	totalTxNum := big.NewInt(0)
	for i := blockNr; i >= 0 && i > blockNr-rpc.BlockNumber(lookbackNum); i-- {
		block, err := s.b.BlockByNumber(ctx, i)
		if block != nil {
			totalTxNum = totalTxNum.Add(totalTxNum, big.NewInt(int64(len(block.Transactions()))))
		} else {
			return nil, err
		}
	}
	return totalTxNum, nil
}

// rpcOutputBlock uses the generalized output filler, then adds the total difficulty field, which requires
// a `PublicBlockchainAPI`.
func (s *PublicBlockChainAPI) rpcOutputBlock(chainID *big.Int, b *types.Block, inclTx bool, fullTx bool) map[string]interface{} {
	fields := RPCMarshalBlock(chainID, b, inclTx, fullTx)
	fields["totalDifficulty"] = s.b.GetTd(b.Hash())
	return fields
}

// GetTransactionByHash returns the transaction for the given hash
func (s *PublicBlockChainAPI) GetTransactionByHash(ctx context.Context, hash common.Hash) *types.RPCTransaction {
	// Try to return an already finalized transaction
	if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
		return tx.NewRPCTransaction(blockHash, blockNumber, index)
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.GetPoolTransaction(hash); tx != nil {
		return tx.NewRPCTransaction(common.Hash{}, 0, 0)
	}
	// Transaction unknown, return as such
	return nil
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

func (s *PublicBlockChainAPI) GetBlockAndResultByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.BlockAndResult, error) {
	return s.b.GetBlockAndResult(ctx, blockNr), nil
}

type CallArgs struct {
	ActionType types.ActionType `json:"actionType"`
	From       common.Name      `json:"from"`
	To         common.Name      `json:"to"`
	AssetID    uint64           `json:"assetId"`
	Gas        uint64           `json:"gas"`
	GasPrice   *big.Int         `json:"gasPrice"`
	Value      *big.Int         `json:"value"`
	Data       hexutil.Bytes    `json:"data"`
}

func (s *PublicBlockChainAPI) doCall(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber, vmCfg vm.Config, timeout time.Duration) ([]byte, uint64, bool, error) {
	defer func(start time.Time) { log.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())

	state, header, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, 0, false, err
	}
	account, err := accountmanager.NewAccountManager(state)
	if err != nil {
		return nil, 0, false, err
	}

	gasPrice := args.GasPrice
	value := args.Value
	assetID := uint64(args.AssetID)
	gas := uint64(args.Gas)

	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	// Get a new instance of the EVM.
	evm, vmError, err := s.b.GetEVM(ctx, account, state, args.From, assetID, gasPrice, header, vmCfg)
	if err != nil {
		return nil, 0, false, err
	}
	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	gp := new(common.GasPool).AddGas(math.MaxUint64)
	action := types.NewAction(args.ActionType, args.From, args.To, 0, assetID, gas, value, args.Data)
	res, gas, failed, err, _ := processor.ApplyMessage(account, evm, action, gp, gasPrice, assetID, s.b.ChainConfig(), s.b.Engine())
	if err := vmError(); err != nil {
		return nil, 0, false, err
	}
	return res, gas, failed, err
}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (s *PublicBlockChainAPI) Call(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	result, _, _, err := s.doCall(ctx, args, blockNr, vm.Config{}, 5*time.Second)
	return (hexutil.Bytes)(result), err
}

// EstimateGas returns an estimate of the amount of gas needed to execute the
// given transaction against the current pending block.
func (s *PublicBlockChainAPI) EstimateGas(ctx context.Context, args CallArgs) (hexutil.Uint64, error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = params.ActionGas - 1
		hi  uint64
		cap uint64
	)
	if uint64(args.Gas) >= params.ActionGas {
		hi = uint64(args.Gas)
	} else {
		// Retrieve the current pending block to act as the gas ceiling
		block, err := s.b.BlockByNumber(ctx, rpc.PendingBlockNumber)
		if err != nil {
			return 0, err
		}
		hi = block.GasLimit()
	}
	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) bool {
		args.Gas = gas

		_, _, failed, err := s.doCall(ctx, args, rpc.PendingBlockNumber, vm.Config{}, 0)
		if err != nil || failed {
			return false
		}
		return true
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		if !executable(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		if !executable(hi) {
			return 0, fmt.Errorf("gas required exceeds allowance or always failing transaction")
		}
	}
	return hexutil.Uint64(hi), nil
}
