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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/txpool"
	"github.com/fractalplatform/fractal/types"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

type StateTransition struct {
	engine      EgnineContext
	from        common.Name
	gp          *common.GasPool
	action      *types.Action
	gas         uint64
	initialGas  uint64
	gasPrice    *big.Int
	assetID     uint64
	account     *accountmanager.AccountManager
	evm         *vm.EVM
	chainConfig *params.ChainConfig
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(accountDB *accountmanager.AccountManager, evm *vm.EVM, action *types.Action, gp *common.GasPool, gasPrice *big.Int, assetID uint64, config *params.ChainConfig, engine EgnineContext) *StateTransition {
	return &StateTransition{
		engine:      engine,
		from:        action.Sender(),
		gp:          gp,
		evm:         evm,
		action:      action,
		gasPrice:    gasPrice,
		assetID:     assetID,
		account:     accountDB,
		chainConfig: config,
	}
}

// ApplyMessage computes the new state by applying the given message against the old state within the environment.
func ApplyMessage(accountDB *accountmanager.AccountManager, evm *vm.EVM, action *types.Action, gp *common.GasPool, gasPrice *big.Int, assetID uint64, config *params.ChainConfig, engine EgnineContext) ([]byte, uint64, bool, error, error) {
	return NewStateTransition(accountDB, evm, action, gp, gasPrice, assetID, config, engine).TransitionDb()
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
	balance, err := st.account.GetAccountBalanceByID(st.from, st.assetID)
	//balance, err := st.account.GetAccountBalanceByID(st.from, st.assetID)
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
	err = st.account.SubAccountBalanceByID(st.from, st.assetID, mgval)
	if err != nil {
		return err
	}
	//st.account.SubAccountBalanceByID(st.from, st.assetID, mgval)
	return nil
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the the used gas. It returns an error if it
// failed. An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error, vmerr error) {
	if err = st.preCheck(); err != nil {
		return
	}

	intrinsicGas, err := txpool.IntrinsicGas(st.action)
	if err != nil {
		return nil, 0, true, err, vmerr
	}
	if err := st.useGas(intrinsicGas); err != nil {
		return nil, 0, true, err, vmerr
	}

	sender := vm.AccountRef(st.from)

	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
	)
	actionType := st.action.Type()
	switch {
	case actionType == types.CreateContract:
		ret, st.gas, vmerr = evm.Create(sender, st.action, st.gas)
	case actionType == types.CallContract:
		ret, st.gas, vmerr = evm.Call(sender, st.action, st.gas)
	case actionType == types.RegProducer:
		fallthrough
	case actionType == types.UpdateProducer:
		fallthrough
	case actionType == types.UnregProducer:
		fallthrough
	case actionType == types.RemoveVoter:
		fallthrough
	case actionType == types.VoteProducer:
		fallthrough
	case actionType == types.ChangeProducer:
		fallthrough
	case actionType == types.UnvoteProducer:
		vmerr = st.engine.ProcessAction(st.evm.ChainConfig(), st.evm.StateDB, st.action)
	default:
		vmerr = st.account.Process(st.action)
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr, vmerr
		}
	}
	nonce, err := st.account.GetNonce(st.from)
	if err != nil {
		return nil, st.gasUsed(), true, err, vmerr
	}
	err = st.account.SetNonce(st.from, nonce+1)
	if err != nil {
		return nil, st.gasUsed(), true, err, vmerr
	}
	st.refundGas()

	if st.action.Value().Sign() != 0 {
		assetFounder, _ := st.account.GetAssetFounder(st.action.AssetID())
		assetFounderRatio := st.chainConfig.AssetChargeRatio
		if len(assetFounder.String()) > 0 {
			if _, ok := evm.FounderGasMap[assetFounder]; !ok {
				evm.FounderGasMap[assetFounder] = int64(params.ActionGas * assetFounderRatio / 100)
			} else {
				evm.FounderGasMap[assetFounder] += int64(params.ActionGas * assetFounderRatio / 100)
			}
		}
	}
	if err := st.distributeGas(); err != nil {
		return ret, st.gasUsed(), true, err, vmerr
	}
	return ret, st.gasUsed(), vmerr != nil, nil, vmerr
}

func (st *StateTransition) distributeGas() error {
	var totalGas int64
	for founder, gas := range st.evm.FounderGasMap {
		st.account.AddAccountBalanceByID(founder, st.assetID, new(big.Int).Mul(st.gasPrice, big.NewInt(gas)))
		totalGas += gas
	}
	if totalGas > int64(st.gasUsed()) {
		return fmt.Errorf("calc wrong gas used")
	}
	st.account.AddAccountBalanceByID(st.evm.Coinbase, st.assetID, new(big.Int).Mul(st.gasPrice, new(big.Int).SetUint64(st.gasUsed()-uint64(totalGas))))
	return nil
}

func (st *StateTransition) refundGas() {
	//st.gas += st.evm.StateDB.GetRefund()

	// Return remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.account.AddAccountBalanceByID(st.from, st.assetID, remaining)
	//st.account.AddAccountBalanceByID(st.from, st.assetID, remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next message.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}
