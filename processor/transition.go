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
	"sort"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/feemanager"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/txpool"
	"github.com/fractalplatform/fractal/types"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

type StateTransition struct {
	engine      EngineContext
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
func NewStateTransition(accountDB *accountmanager.AccountManager, evm *vm.EVM,
	action *types.Action, gp *common.GasPool, gasPrice *big.Int, assetID uint64,
	config *params.ChainConfig, engine EngineContext) *StateTransition {
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
func ApplyMessage(accountDB *accountmanager.AccountManager, evm *vm.EVM,
	action *types.Action, gp *common.GasPool, gasPrice *big.Int,
	assetID uint64, config *params.ChainConfig, engine EngineContext) ([]byte, uint64, bool, error, error) {
	return NewStateTransition(accountDB, evm, action, gp, gasPrice,
		assetID, config, engine).TransitionDb()
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
	balance, err := st.account.GetAccountBalanceByID(st.from, st.assetID, 0)
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
	return st.account.TransferAsset(st.from, common.Name(st.chainConfig.FeeName), st.assetID, mgval)
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the the used gas. It returns an error if it
// failed. An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool,
	err error, vmerr error) {
	if err = st.preCheck(); err != nil {
		return
	}

	intrinsicGas, err := txpool.IntrinsicGas(st.account, st.action)
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
	case actionType == types.RegCandidate:
		fallthrough
	case actionType == types.UpdateCandidate:
		fallthrough
	case actionType == types.UnregCandidate:
		fallthrough
	case actionType == types.VoteCandidate:
		fallthrough
	case actionType == types.RefundCandidate:
		fallthrough
	case actionType == types.KickedCandidate:
		fallthrough
	case actionType == types.RemoveKickedCandidate:
		fallthrough
	case actionType == types.ExitTakeOver:
		internalLogs, err := st.engine.ProcessAction(st.evm.Context.ForkID, st.evm.Context.BlockNumber.Uint64(),
			st.evm.ChainConfig(), st.evm.StateDB, st.action)
		vmerr = err
		evm.InternalTxs = append(evm.InternalTxs, internalLogs...)
	default:
		internalLogs, err := st.account.Process(&types.AccountManagerContext{
			Action:      st.action,
			Number:      st.evm.Context.BlockNumber.Uint64(),
			CurForkID:   st.evm.Context.ForkID,
			ChainConfig: st.chainConfig,
		})
		vmerr = err
		evm.InternalTxs = append(evm.InternalTxs, internalLogs...)
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance || vmerr == vm.ErrExecOverTime {
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

	st.distributeGas(intrinsicGas)

	if err := st.distributeFee(); err != nil {
		return ret, st.gasUsed(), true, err, vmerr
	}
	return ret, st.gasUsed(), vmerr != nil, nil, vmerr
}

func (st *StateTransition) distributeGas(intrinsicGas uint64) {
	switch st.action.Type() {
	case types.Transfer:
		assetInfo, _ := st.evm.AccountDB.GetAssetInfoByID(st.action.AssetID())
		assetName := common.Name(assetInfo.GetAssetName())
		assetFounderRatio := st.chainConfig.ChargeCfg.AssetRatio

		key := vm.DistributeKey{ObjectName: assetName,
			ObjectType: params.AssetFeeType}
		assetGas := int64(st.gasUsed() * assetFounderRatio / 100)
		dGas := vm.DistributeGas{
			Value:  assetGas,
			TypeID: params.AssetFeeType}
		st.evm.FounderGasMap[key] = dGas

		key = vm.DistributeKey{ObjectName: st.evm.Coinbase,
			ObjectType: params.CoinbaseFeeType}
		st.evm.FounderGasMap[key] = vm.DistributeGas{
			Value:  int64(st.gasUsed()) - assetGas,
			TypeID: params.CoinbaseFeeType}

	case types.CreateContract:
		fallthrough
	case types.CallContract:
		st.distributeToContract(st.action.Recipient(), intrinsicGas)
		return
	case types.CreateAccount:
		fallthrough
	case types.UpdateAccount:
		fallthrough
	case types.DeleteAccount:
		fallthrough
	case types.UpdateAccountAuthor:
		st.distributeToSystemAccount(common.Name(st.chainConfig.AccountName))
		return
	case types.IncreaseAsset:
		fallthrough
	case types.IssueAsset:
		fallthrough
	case types.DestroyAsset:
		fallthrough
	case types.SetAssetOwner:
		fallthrough
	case types.UpdateAssetContract:
		fallthrough
	case types.UpdateAsset:
		st.distributeToSystemAccount(common.Name(st.chainConfig.AssetName))
		return
	case types.RegCandidate:
		fallthrough
	case types.UpdateCandidate:
		fallthrough
	case types.UnregCandidate:
		fallthrough
	case types.VoteCandidate:
		fallthrough
	case types.RefundCandidate:
		fallthrough
	case types.KickedCandidate:
		fallthrough
	case types.RemoveKickedCandidate:
		fallthrough
	case types.ExitTakeOver:
		st.distributeToSystemAccount(common.Name(st.chainConfig.DposName))
		return
	}
}

func (st *StateTransition) distributeToContract(name common.Name, intrinsicGas uint64) {
	contractFounderRation := st.chainConfig.ChargeCfg.ContractRatio
	key := vm.DistributeKey{ObjectName: name,
		ObjectType: params.ContractFeeType}
	contractGas := int64(intrinsicGas * contractFounderRation / 100)

	if _, ok := st.evm.FounderGasMap[key]; !ok {
		st.evm.FounderGasMap[key] = vm.DistributeGas{
			Value:  contractGas,
			TypeID: params.ContractFeeType}
	} else {
		dGas := vm.DistributeGas{
			Value:  contractGas,
			TypeID: params.ContractFeeType}
		dGas.Value = st.evm.FounderGasMap[key].Value + dGas.Value
		st.evm.FounderGasMap[key] = dGas
	}

	var totalGas int64
	for _, gas := range st.evm.FounderGasMap {
		totalGas += gas.Value
	}

	key = vm.DistributeKey{ObjectName: st.evm.Coinbase,
		ObjectType: params.CoinbaseFeeType}
	st.evm.FounderGasMap[key] = vm.DistributeGas{
		Value:  int64(st.gasUsed()) - totalGas,
		TypeID: params.CoinbaseFeeType}
}

func (st *StateTransition) distributeToSystemAccount(name common.Name) {
	contractFounderRation := st.chainConfig.ChargeCfg.ContractRatio
	key := vm.DistributeKey{ObjectName: name,
		ObjectType: params.ContractFeeType}
	contractGas := int64(st.gasUsed() * contractFounderRation / 100)
	dGas := vm.DistributeGas{
		Value:  contractGas,
		TypeID: params.ContractFeeType}
	st.evm.FounderGasMap[key] = dGas

	key = vm.DistributeKey{ObjectName: st.evm.Coinbase,
		ObjectType: params.CoinbaseFeeType}
	st.evm.FounderGasMap[key] = vm.DistributeGas{
		Value:  int64(st.gasUsed()) - contractGas,
		TypeID: params.CoinbaseFeeType}

}

func (st *StateTransition) distributeFee() error {
	fm := feemanager.NewFeeManager(st.evm.StateDB, st.evm.AccountDB)

	var keys vm.DistributeKeys
	for key := range st.evm.FounderGasMap {
		keys = append(keys, key)
	}
	sort.Sort(keys)

	for _, key := range keys {
		gas := st.evm.FounderGasMap[key]
		if gas.Value > 0 {
			value := new(big.Int).Mul(st.gasPrice, big.NewInt(gas.Value))
			err := fm.RecordFeeInSystem(key.ObjectName.String(), gas.TypeID, st.assetID, value)
			if err != nil {
				return fmt.Errorf("record fee err(%v), key:%v,assetID:%d", err, key, st.assetID)
			}
		}
	}

	return nil
}

func (st *StateTransition) refundGas() {
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.account.TransferAsset(common.Name(st.chainConfig.FeeName), st.from, st.assetID, remaining)
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}
