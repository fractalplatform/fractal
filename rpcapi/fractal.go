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
	"math"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/common/hexutil"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/params"
	pm "github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// PublicFractalAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicFractalAPI struct {
	b Backend
}

// NewPublicFractalAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicFractalAPI(b Backend) *PublicFractalAPI {
	return &PublicFractalAPI{b}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicFractalAPI) GasPrice(ctx context.Context) (*big.Int, error) {
	return s.b.SuggestPrice(ctx)
}

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicFractalAPI) SendRawTransaction(ctx context.Context, encodedTx hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, tx)
}

type CallArgs struct {
	Type     envelope.Type `json:"txType"`
	From     string        `json:"from"`
	To       string        `json:"to"`
	AssetID  uint64        `json:"assetID"`
	Gas      uint64        `json:"gas"`
	GasPrice *big.Int      `json:"gasPrice"`
	Value    *big.Int      `json:"value"`
	Data     hexutil.Bytes `json:"data"`
	Remark   hexutil.Bytes `json:"remark"`
}

func (s *PublicFractalAPI) doCall(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber, vmCfg vm.Config, timeout time.Duration) ([]byte, uint64, bool, error) {
	defer func(start time.Time) { log.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())

	state, header, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, 0, false, err
	}
	account := pm.NewPM(state)

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
	evm, vmError, err := s.b.GetEVM(ctx, account, state, args.From, args.To, assetID, gasPrice, header, vmCfg)
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
	action, err := envelope.NewContractTx(args.Type, args.From, args.To, 0, assetID, 0, gas, big.NewInt(0), value, args.Data, args.Remark)
	if err != nil {
		return nil, 0, false, err
	}
	res, gas, _, failed, err, _ := processor.ApplyMessage(account, evm, types.NewTransaction(action), gp, gasPrice, assetID, s.b.ChainConfig())
	if err := vmError(); err != nil {
		return nil, 0, false, err
	}

	return res, gas, failed, err
}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (s *PublicFractalAPI) Call(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	result, _, _, err := s.doCall(ctx, args, blockNr, vm.Config{}, 5*time.Second)
	return (hexutil.Bytes)(result), err
}

// EstimateGas returns an estimate of the amount of gas needed to execute the
// given transaction against the current pending block.
func (s *PublicFractalAPI) EstimateGas(ctx context.Context, args CallArgs) (uint64, error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = params.GasTableInstance.ActionGas - 1
		hi  uint64
		cap uint64
	)
	if uint64(args.Gas) >= params.GasTableInstance.ActionGas {
		hi = uint64(args.Gas)
	} else {
		// Retrieve the current pending block to act as the gas ceiling
		block := s.b.BlockByNumber(ctx, rpc.LatestBlockNumber)
		hi = block.GasLimit()
	}
	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) bool {
		args.Gas = gas
		_, _, failed, err := s.doCall(ctx, args, rpc.LatestBlockNumber, vm.Config{}, 0)
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
	return hi, nil
}

// GetChainConfig returns chain config.
func (s *PublicFractalAPI) GetChainConfig(ctx context.Context) *params.ChainConfig {
	g := s.b.BlockByNumber(ctx, 0)
	return rawdb.ReadChainConfig(s.b.ChainDb(), g.Hash())
}
