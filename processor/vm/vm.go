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

//VM is a Virtual Machine based on Ethereum Virtual Machine
package vm

import (
	"math/big"
	"sync/atomic"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

type (
	// GetHashFunc returns the nth block hash in the blockchain and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) common.Hash
	// GetDelegatedByTimeFunc returns the delegated balance
	GetDelegatedByTimeFunc func(*state.StateDB, string, uint64) (stake *big.Int, err error)
	//GetEpoch
	GetEpochFunc func(state *state.StateDB, t uint64, curEpoch uint64) (epoch uint64, time uint64, err error)
	//GetActivedCandidateSize
	GetActivedCandidateSizeFunc func(state *state.StateDB, epoch uint64) (size uint64, err error)
	//GetActivedCandidate
	GetActivedCandidateFunc func(state *state.StateDB, epoch uint64, index uint64) (name string, stake *big.Int, votes *big.Int, counter uint64, actualCounter uint64, replace uint64, isbad bool, err error)
	//GetVoterStake
	GetVoterStakeFunc func(state *state.StateDB, epoch uint64, voter string, candidate string) (stake *big.Int, err error)
	// GetHeaderByNumberFunc
	GetHeaderByNumberFunc func(number uint64) *types.Header
)

// Context provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context struct {
	GetHash            GetHashFunc
	GetDelegatedByTime GetDelegatedByTimeFunc
	//GetLatestEpoch          GetLatestEpochFunc
	//GetPrevEpoch            GetPrevEpochFunc
	GetEpoch                GetEpochFunc
	GetActivedCandidateSize GetActivedCandidateSizeFunc
	GetActivedCandidate     GetActivedCandidateFunc
	GetVoterStake           GetVoterStakeFunc
	GetHeaderByNumber       GetHeaderByNumberFunc

	// Message information
	Origin    common.Name // Provides information for ORIGIN
	Recipient common.Name
	From      common.Name // Provides information for ORIGIN
	AssetID   uint64      // provides assetId
	GasPrice  *big.Int    // Provides information for GASPRICE

	// Block information
	Coinbase    common.Name // Provides information for COINBASE
	GasLimit    uint64      // Provides information for GASLIMIT
	BlockNumber *big.Int    // Provides information for NUMBER
	ForkID      uint64      // Provides information for FORKID
	Time        *big.Int    // Provides information for TIME
	Difficulty  *big.Int    // Provides information for DIFFICULTY
}

type FounderGas struct {
	Founder common.Name
	Gas     uint64
}

type EVM struct {
	// Context provides auxiliary blockchain related information
	Context
	// Asset operation func
	AccountDB *accountmanager.AccountManager
	// StateDB gives access to the underlying state
	StateDB *state.StateDB
	// Depth is the current call stack
	depth int

	// chainConfig contains information about the current chain
	chainConfig *params.ChainConfig
	// chain rules contains the chain rules for the current epoch
	//chainRules params.Rules
	// virtual machine configuration options used to initialise the
	// evm.
	vmConfig Config
	// global (to this context) ethereum virtual machine
	// used throughout the execution of the tx.
	interpreter *Interpreter
	// abort is used to abort the EVM calling operations
	// NOTE: must be set atomically
	abort int32
	// callGasTemp holds the gas available for the current call. This is needed because the
	// available gas is calculated in gasCall* according to the 63/64 rule and later
	// applied in opCall*.
	callGasTemp uint64

	FounderGasMap map[DistributeKey]DistributeGas

	InternalTxs []*types.InternalAction
}

type DistributeGas struct {
	Value  int64
	TypeID uint64
}

type DistributeKey struct {
	ObjectName common.Name
	ObjectType uint64
}
type DistributeKeys []DistributeKey

func (keys DistributeKeys) Len() int {
	return len(keys)
}
func (keys DistributeKeys) Less(i, j int) bool {
	if keys[i].ObjectName == keys[j].ObjectName {
		return keys[i].ObjectType < keys[j].ObjectType
	}
	return keys[i].ObjectName < keys[j].ObjectName
}
func (keys DistributeKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}

// NewEVM retutrns a new EVM . The returned EVM is not thread safe and should
// only ever be used *once*.
func NewEVM(ctx Context, accountdb *accountmanager.AccountManager, statedb *state.StateDB, chainCfg *params.ChainConfig, vmConfig Config) *EVM {
	evm := &EVM{
		Context:     ctx,
		AccountDB:   accountdb,
		StateDB:     statedb,
		chainConfig: chainCfg,
		vmConfig:    vmConfig,
	}
	evm.interpreter = NewInterpreter(evm, vmConfig)
	evm.FounderGasMap = map[DistributeKey]DistributeGas{}
	return evm
}

// emptyCodeHash is used by create to ensure deployment is disallowed to already
// deployed contract addresses (relevant after the account abstraction).
var emptyCodeHash = crypto.Keccak256Hash(nil)

// run runs the given contract and takes care of running precompiles with a fallback to the byte code interpreter.
func run(evm *EVM, contract *Contract, input []byte) ([]byte, error) {
	return evm.interpreter.Run(contract, input)
}

// Cancel cancels any running EVM operation. This may be called concurrently and
// it's safe to be called multiple times.
func (evm *EVM) Cancel() {
	atomic.StoreInt32(&evm.abort, 1)
}

func (evm *EVM) GetCurrentGasTable() params.GasTable {
	return evm.interpreter.GetGasTable()
}

func (evm *EVM) CheckReceipt(action *types.Action) uint64 {
	gasTable := evm.GetCurrentGasTable()
	if action.Value().Sign() == 0 {
		return 0
	}
	toAcct, err := evm.AccountDB.GetAccountByName(action.Recipient())
	if err != nil {
		return 0
	}
	if toAcct == nil {
		return 0
	}
	if toAcct.IsDestroyed() {
		return 0
	}
	_, err = toAcct.GetBalanceByID(action.AssetID())
	if err == accountmanager.ErrAccountAssetNotExist {
		return gasTable.CallValueTransferGas
	}
	return 0
}

func (evm *EVM) distributeContractGas(runGas uint64, contractName common.Name, callerName common.Name) {
	if runGas > 0 && len(contractName.String()) > 0 {
		contratFounderRatio := evm.chainConfig.ChargeCfg.ContractRatio

		key := DistributeKey{ObjectName: contractName,
			ObjectType: params.ContractFeeType}
		if _, ok := evm.FounderGasMap[key]; !ok {
			dGas := DistributeGas{int64(runGas * contratFounderRatio / 100), params.ContractFeeType}
			evm.FounderGasMap[key] = dGas
		} else {
			dGas := DistributeGas{int64(runGas * contratFounderRatio / 100), params.ContractFeeType}
			dGas.Value = evm.FounderGasMap[key].Value + dGas.Value
			evm.FounderGasMap[key] = dGas
		}
		if evm.depth != 0 {
			key = DistributeKey{ObjectName: callerName,
				ObjectType: params.ContractFeeType}
			if _, ok := evm.FounderGasMap[key]; !ok {
				dGas := DistributeGas{-int64(runGas * contratFounderRatio / 100), params.ContractFeeType}
				evm.FounderGasMap[key] = dGas
			} else {
				dGas := DistributeGas{-int64(runGas * contratFounderRatio / 100), params.ContractFeeType}
				dGas.Value = evm.FounderGasMap[key].Value + dGas.Value
				evm.FounderGasMap[key] = dGas
			}
		}
	}
}

func (evm *EVM) distributeAssetGas(callValueGas int64, assetName common.Name, callerName common.Name) {
	if evm.depth != 0 {
		assetFounderRatio := evm.chainConfig.ChargeCfg.AssetRatio //get asset founder charge ratio

		key := DistributeKey{ObjectName: assetName,
			ObjectType: params.AssetFeeType}
		if len(assetName.String()) > 0 {
			if _, ok := evm.FounderGasMap[key]; !ok {
				dGas := DistributeGas{int64(callValueGas * int64(assetFounderRatio) / 100), params.AssetFeeType}
				evm.FounderGasMap[key] = dGas
			} else {
				dGas := DistributeGas{int64(callValueGas * int64(assetFounderRatio) / 100), params.AssetFeeType}
				dGas.Value = evm.FounderGasMap[key].Value + dGas.Value
				evm.FounderGasMap[key] = dGas
			}
		}
		if len(callerName.String()) > 0 {
			key = DistributeKey{ObjectName: callerName,
				ObjectType: params.ContractFeeType}
			if _, ok := evm.FounderGasMap[key]; !ok {
				dGas := DistributeGas{-int64(callValueGas * int64(assetFounderRatio) / 100), params.ContractFeeType}
				evm.FounderGasMap[key] = dGas
			} else {
				dGas := DistributeGas{int64(callValueGas * int64(assetFounderRatio) / 100), params.ContractFeeType}
				dGas.Value = evm.FounderGasMap[key].Value - dGas.Value
				evm.FounderGasMap[key] = dGas
			}
		}
	}
}

func (evm *EVM) distributeGasByScale(actualUsedGas uint64, runGas uint64) {
	if evm.depth == 0 && actualUsedGas != runGas {
		for key, gas := range evm.FounderGasMap {
			mulGas := new(big.Int).Mul(big.NewInt(gas.Value), big.NewInt(int64(actualUsedGas)))
			divgas := new(big.Int).Div(mulGas, big.NewInt(int64(runGas)))
			v := DistributeGas{divgas.Int64(), gas.TypeID}
			evm.FounderGasMap[key] = v
		}
	}
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (evm *EVM) Call(caller ContractRef, action *types.Action, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance

	if ok, err := evm.AccountDB.CanTransfer(caller.Name(), action.AssetID(), action.Value()); !ok || err != nil {
		return nil, gas, ErrInsufficientBalance
	}

	toName := action.Recipient()

	var (
		to       = AccountRef(toName)
		snapshot = evm.StateDB.Snapshot()
	)

	if evm.depth != 0 {
		receiptGas := evm.CheckReceipt(action)
		if gas < receiptGas {
			return nil, gas, ErrInsufficientBalance
		} else {
			gas -= receiptGas
		}
	}

	if err := evm.AccountDB.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value()); err != nil {
		return nil, gas, err
	}

	var assetName common.Name
	assetFounder, _ := evm.AccountDB.GetAssetFounder(action.AssetID()) //get asset founder name

	if len(assetFounder.String()) > 0 {
		assetInfo, _ := evm.AccountDB.GetAssetInfoByID(action.AssetID())
		assetName = common.Name(assetInfo.GetAssetName())
	}

	contractName := toName

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.

	contract := NewContract(caller, to, action.Value(), gas, action.AssetID())
	acct, err := evm.AccountDB.GetAccountByName(toName)
	if err != nil {
		return nil, gas, err
	}
	if acct == nil {
		return nil, gas, ErrAccountNotExist
	}
	codeHash, err := acct.GetCodeHash()
	if err != nil {
		return nil, gas, err
	}
	code, _ := acct.GetCode()
	contract.SetCallCode(&toName, codeHash, code)

	ret, err = run(evm, contract, action.Data())
	runGas := gas - contract.Gas

	evm.distributeContractGas(runGas, contractName, caller.Name())

	gasTable := evm.GetCurrentGasTable()
	callValueGas := int64(gasTable.CallValueTransferGas - gasTable.CallStipend)
	if action.Value().Sign() != 0 && callValueGas > 0 {
		evm.distributeAssetGas(callValueGas, assetName, caller.Name())
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	actualUsedGas := gas - contract.Gas
	evm.distributeGasByScale(actualUsedGas, runGas)
	return ret, contract.Gas, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (evm *EVM) CallCode(caller ContractRef, action *types.Action, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}

	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if ok, err := evm.AccountDB.CanTransfer(caller.Name(), evm.AssetID, action.Value()); !ok || err != nil {
		return nil, gas, ErrInsufficientBalance
	}

	toName := action.Recipient()

	var (
		snapshot = evm.StateDB.Snapshot()
		to       = AccountRef(caller.Name())
	)
	// initialise a new contract and set the code that is to be used by the
	// E The contract is a scoped evmironment for this execution context
	// only.
	contract := NewContract(caller, to, action.Value(), gas, evm.AssetID)
	acct, err := evm.AccountDB.GetAccountByName(toName)
	if err != nil {
		return nil, gas, err
	}
	codeHash, err := acct.GetCodeHash()
	if err != nil {
		return nil, gas, err
	}
	code, _ := acct.GetCode()
	//codeHash, _ := evm.AccountDB.GetCodeHash(toName)
	//code, _ := evm.AccountDB.GetCode(toName)
	contract.SetCallCode(&toName, codeHash, code)

	ret, err = run(evm, contract, action.Data())
	runGas := gas - contract.Gas

	contractName := toName

	evm.distributeContractGas(runGas, contractName, caller.Name())

	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}

	actualUsedGas := gas - contract.Gas
	evm.distributeGasByScale(actualUsedGas, runGas)
	return ret, contract.Gas, err
}

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (evm *EVM) DelegateCall(caller ContractRef, name common.Name, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	var (
		snapshot = evm.StateDB.Snapshot()
		to       = AccountRef(caller.Name())
	)

	// Initialise a new contract and make initialise the delegate values
	contract := NewContract(caller, to, nil, gas, evm.AssetID).AsDelegate()
	acct, err := evm.AccountDB.GetAccountByName(name)
	if err != nil {
		return nil, gas, err
	}
	codeHash, err := acct.GetCodeHash()
	if err != nil {
		return nil, gas, err
	}
	code, _ := acct.GetCode()
	//codeHash, _ := evm.AccountDB.GetCodeHash(name)
	//code, _ := evm.AccountDB.GetCode(name)
	contract.SetCallCode(&name, codeHash, code)

	ret, err = run(evm, contract, input)
	runGas := gas - contract.Gas

	contractName := name

	evm.distributeContractGas(runGas, contractName, caller.Name())

	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}

	actualUsedGas := gas - contract.Gas
	evm.distributeGasByScale(actualUsedGas, runGas)
	return ret, contract.Gas, err
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (evm *EVM) StaticCall(caller ContractRef, name common.Name, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Make sure the readonly is only set if we aren't in readonly yet
	// this makes also sure that the readonly flag isn't removed for
	// child calls.
	if !evm.interpreter.readOnly {
		evm.interpreter.readOnly = true
		defer func() { evm.interpreter.readOnly = false }()
	}

	var (
		to       = AccountRef(name)
		snapshot = evm.StateDB.Snapshot()
	)
	// Initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := NewContract(caller, to, new(big.Int), gas, evm.AssetID)
	acct, err := evm.AccountDB.GetAccountByName(name)
	if err != nil {
		return nil, gas, err
	}
	codeHash, err := acct.GetCodeHash()
	if err != nil {
		return nil, gas, err
	}
	code, _ := acct.GetCode()
	//codeHash, _ := evm.AccountDB.GetCodeHash(name)
	//code, _ := evm.AccountDB.GetCode(name)
	contract.SetCallCode(&name, codeHash, code)

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in Homestead this also counts for code storage gas errors.
	ret, err = run(evm, contract, input)
	runGas := gas - contract.Gas

	contractName := to.Name()

	evm.distributeContractGas(runGas, contractName, caller.Name())

	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}

	actualUsedGas := gas - contract.Gas
	evm.distributeGasByScale(actualUsedGas, runGas)
	return ret, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (evm *EVM) Create(caller ContractRef, action *types.Action, gas uint64) (ret []byte, leftOverGas uint64, err error) {

	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	if ok, err := evm.AccountDB.CanTransfer(caller.Name(), evm.AssetID, action.Value()); !ok || err != nil {
		return nil, gas, ErrInsufficientBalance
	}

	contractName := action.Recipient()
	snapshot := evm.StateDB.Snapshot()

	if b, err := evm.AccountDB.AccountHaveCode(contractName); err != nil {
		return nil, 0, err
	} else if b {
		return nil, 0, ErrContractCodeCollision
	}

	if err := evm.AccountDB.TransferAsset(action.Sender(), action.Recipient(), evm.AssetID, action.Value()); err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		return nil, gas, err
	}

	// initialise a new contract and set the code that is to be used by the
	// E The contract is a scoped evmironment for this execution context
	// only.
	contract := NewContract(caller, AccountRef(contractName), action.Value(), gas, evm.AssetID)
	contract.SetCallCode(&contractName, crypto.Keccak256Hash(action.Data()), action.Data())

	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}

	if evm.vmConfig.Debug && evm.depth == 0 {
		evm.vmConfig.Tracer.CaptureStart(caller.Name(), contractName, true, action.Data(), gas, action.Value())
	}
	start := time.Now()

	ret, err = run(evm, contract, nil)

	// check whether the max code size has been exceeded
	maxCodeSizeExceeded := len(ret) > int(params.MaxCodeSize)
	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	if err == nil && !maxCodeSizeExceeded {
		createDataGas := uint64(len(ret)) * evm.GetCurrentGasTable().CreateDataGas
		if contract.UseGas(createDataGas) {
			if _, err = evm.AccountDB.SetCode(contractName, ret); err != nil {
				return nil, gas, err
			}
		} else {
			err = ErrCodeStoreOutOfGas
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if maxCodeSizeExceeded || (err != nil && err != ErrCodeStoreOutOfGas) {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	// Assign err if contract code size exceeds the max while the err is still empty.
	if maxCodeSizeExceeded && err == nil {
		err = errMaxCodeSizeExceeded
	}
	if evm.vmConfig.Debug && evm.depth == 0 {
		evm.vmConfig.Tracer.CaptureEnd(ret, gas-contract.Gas, time.Since(start), err)
	}

	evm.distributeContractGas(gas-contract.Gas, contractName, contractName)
	return ret, contract.Gas, err
}

// ChainConfig returns the environment's chain configuration
func (evm *EVM) ChainConfig() *params.ChainConfig { return evm.chainConfig }

// Interpreter returns the EVM interpreter
func (evm *EVM) Interpreter() *Interpreter { return evm.interpreter }
