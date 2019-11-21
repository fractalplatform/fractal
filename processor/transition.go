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
	"errors"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/txpool"
	"github.com/fractalplatform/fractal/types"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

type StateTransition struct {
	from        string
	gp          *common.GasPool
	action      *types.Action
	gas         uint64
	initialGas  uint64
	gasPrice    *big.Int
	assetID     uint64
	pm          plugin.IPM
	evm         *vm.EVM
	chainConfig *params.ChainConfig
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(pm plugin.IPM, evm *vm.EVM,
	action *types.Action, gp *common.GasPool, gasPrice *big.Int, assetID uint64,
	config *params.ChainConfig) *StateTransition {
	return &StateTransition{
		from:        action.Sender(),
		gp:          gp,
		evm:         evm,
		action:      action,
		gasPrice:    gasPrice,
		assetID:     assetID,
		pm:          pm,
		chainConfig: config,
	}
}

// ApplyMessage computes the new state by applying the given message against the old state within the environment.
func ApplyMessage(pm plugin.IPM, evm *vm.EVM,
	action *types.Action, gp *common.GasPool, gasPrice *big.Int,
	assetID uint64, config *params.ChainConfig) ([]byte, uint64, bool, error, error) {
	return NewStateTransition(pm, evm, action, gp, gasPrice,
		assetID, config).TransitionDb()
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return vm.ErrOutOfGas
	}
	st.gas -= amount
	return nil
}

func (st *StateTransition) preCheck() error {
	return st.buyGas()
}

func (st *StateTransition) buyGas() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.action.Gas()), st.gasPrice)
	balance, err := st.pm.GetBalance(st.from, st.assetID)
	if err != nil {
		return err
	}
	if balance.Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}
	if err := st.gp.SubGas(st.action.Gas()); err != nil {
		return err
	}
	st.gas += st.action.Gas()
	st.initialGas = st.action.Gas()
	return st.pm.TransferAsset(st.from, string(st.chainConfig.FeeName), st.assetID, mgval)
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the the used gas. It returns an error if it
// failed. An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool,
	err error, vmerr error) {
	if err = st.preCheck(); err != nil {
		return
	}

	intrinsicGas, err := txpool.IntrinsicGas(st.pm, st.action)
	if err != nil {
		return nil, 0, true, err, vmerr
	}
	if err := st.useGas(intrinsicGas); err != nil {
		return nil, 0, true, err, vmerr
	}

	caller := vm.AccountRef(st.action.Sender())
	actionType := st.action.Type()
	switch {
	case actionType == types.CreateContract:
		ret, st.gas, vmerr = st.evm.Create(caller, st.action, st.gas)
	case actionType == types.CallContract:
		ret, st.gas, vmerr = st.evm.Call(caller, st.action, st.gas)
	default:
		ret, vmerr = st.pm.ExecTx(st.action)
	}

	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == plugin.ErrInsufficientBalance || vmerr == vm.ErrInsufficientBalance || vmerr == vm.ErrExecOverTime {
			return nil, 0, false, vmerr, vmerr
		}
	}

	nonce, err := st.pm.GetNonce(st.from)
	if err != nil {
		return nil, st.gasUsed(), true, err, vmerr
	}
	err = st.pm.SetNonce(st.from, nonce+1)
	if err != nil {
		return nil, st.gasUsed(), true, err, vmerr
	}

	st.refundGas()

	key := types.DistributeKey{ObjectName: st.evm.Coinbase,
		ObjectType: types.CoinbaseFeeType}
	st.evm.FounderGasMap[key] = types.DistributeGas{
		Value:  int64(intrinsicGas),
		TypeID: types.CoinbaseFeeType}
	// st.distributeGas(intrinsicGas)
	if err := st.pm.DistributeGas(st.chainConfig.FeeName, st.evm.FounderGasMap, st.assetID, st.gasPrice, st.pm); err != nil {
		return ret, st.gasUsed(), true, err, vmerr
	}
	return ret, st.gasUsed(), vmerr != nil, nil, vmerr
}

func (st *StateTransition) refundGas() {
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.pm.TransferAsset(string(st.chainConfig.FeeName), st.from, st.assetID, remaining)
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}
