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

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
)

type (
	// GetHashFunc returns the nth block hash in the blockchain and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) common.Hash
	// CurrentHeaderFunc retrieves the current header from the local chain.
	CurrentHeaderFunc func() *types.Header
	// GetHeaderByNumberFunc retrieves a block header from the database by number.
	GetHeaderByNumberFunc func(uint64) *types.Header
	// GetHeaderByHashFunc retrieves a block header from the database by its hash.
	GetHeaderByHashFunc func(common.Hash) *types.Header
)

// Context provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context struct {
	GetHash           GetHashFunc
	CurrentHeader     CurrentHeaderFunc
	GetHeaderByNumber GetHeaderByNumberFunc
	GetHeaderByHash   GetHeaderByHashFunc
	// Message information
	Origin    string // Provides information for ORIGIN
	Recipient string
	From      string   // Provides information for ORIGIN
	AssetID   uint64   // provides information for ASSETID
	GasPrice  *big.Int // Provides information for GASPRICE

	// Block information
	Coinbase    string   // Provides information for COINBASE
	GasLimit    uint64   // Provides information for GASLIMIT
	BlockNumber *big.Int // Provides information for NUMBER
	ForkID      uint64   // Provides information for FORKID
	Time        *big.Int // Provides information for TIME
	Difficulty  *big.Int // Provides information for DIFFICULTY
}

type EVM struct {
	// Context provides auxiliary blockchain related information
	Context
	// Asset operation func
	PM plugin.IPM
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

	FounderGasMap map[types.DistributeKey]types.DistributeGas

	InternalTxs []*types.InternalTx
}

// NewEVM retutrns a new EVM . The returned EVM is not thread safe and should
// only ever be used *once*.
func NewEVM(ctx Context, pm plugin.IPM, statedb *state.StateDB, chainCfg *params.ChainConfig, vmConfig Config) *EVM {
	evm := &EVM{
		Context:     ctx,
		PM:          pm,
		StateDB:     statedb,
		chainConfig: chainCfg,
		vmConfig:    vmConfig,
	}
	evm.interpreter = NewInterpreter(evm, vmConfig)
	evm.FounderGasMap = map[types.DistributeKey]types.DistributeGas{}
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

//
func (evm *EVM) OverTimeAbort() {
	atomic.StoreInt32(&evm.abort, 2)
}

// IsOverTime Check vm is overtime abort
func (evm *EVM) IsOverTime() bool {
	if atomic.LoadInt32(&evm.abort) == 2 {
		return true
	}
	return false
}

func (evm *EVM) GetCurrentGasTable() params.GasTable {
	return evm.interpreter.GetGasTable()
}

func (evm *EVM) CheckReceipt(action *envelope.ContractTx) uint64 {
	gasTable := evm.GetCurrentGasTable()
	if _, err := evm.PM.GetBalance(action.Recipient(), action.GetAssetID()); err != nil {
		return gasTable.CallValueTransferGas
	} else {
		return 0
	}
}

func (evm *EVM) distributeContractGas(runGas uint64, contractName string, callerName string) {
	if runGas > 0 && len(contractName) > 0 {
		key := types.DistributeKey{ObjectName: contractName,
			ObjectType: types.ContractFeeType}
		if _, ok := evm.FounderGasMap[key]; !ok {
			dGas := types.DistributeGas{int64(runGas), types.ContractFeeType}
			evm.FounderGasMap[key] = dGas
		} else {
			dGas := types.DistributeGas{int64(runGas), types.ContractFeeType}
			dGas.Value = evm.FounderGasMap[key].Value + dGas.Value
			evm.FounderGasMap[key] = dGas
		}
		if evm.depth != 0 {
			key = types.DistributeKey{ObjectName: callerName,
				ObjectType: types.ContractFeeType}
			if _, ok := evm.FounderGasMap[key]; !ok {
				dGas := types.DistributeGas{-int64(runGas), types.ContractFeeType}
				evm.FounderGasMap[key] = dGas
			} else {
				dGas := types.DistributeGas{-int64(runGas), types.ContractFeeType}
				dGas.Value = evm.FounderGasMap[key].Value + dGas.Value
				evm.FounderGasMap[key] = dGas
			}
		}
	}
}

func (evm *EVM) distributeAssetGas(callValueGas int64, assetName string, callerName string) {
	if evm.depth != 0 {
		key := types.DistributeKey{ObjectName: assetName,
			ObjectType: types.AssetFeeType}
		if _, ok := evm.FounderGasMap[key]; !ok {
			dGas := types.DistributeGas{int64(callValueGas), types.AssetFeeType}
			evm.FounderGasMap[key] = dGas
		} else {
			dGas := types.DistributeGas{int64(callValueGas), types.AssetFeeType}
			dGas.Value = evm.FounderGasMap[key].Value + dGas.Value
			evm.FounderGasMap[key] = dGas
		}
		if len(callerName) > 0 {
			key = types.DistributeKey{ObjectName: callerName,
				ObjectType: types.ContractFeeType}
			if _, ok := evm.FounderGasMap[key]; !ok {
				dGas := types.DistributeGas{-int64(callValueGas), types.ContractFeeType}
				evm.FounderGasMap[key] = dGas
			} else {
				dGas := types.DistributeGas{int64(callValueGas), types.ContractFeeType}
				dGas.Value = evm.FounderGasMap[key].Value - dGas.Value
				evm.FounderGasMap[key] = dGas
			}
		}
	}
}

func (evm *EVM) CallPlugin(tx *types.Transaction, ctx *plugin.Context, fromSol bool) ([]byte, error) {
	ctx.InternalTxs = make([]*types.InternalTx, 0)
	ret, err := evm.PM.ExecTx(tx, ctx, fromSol)
	if evm.vmConfig.ContractLogFlag {
		evm.InternalTxs = append(evm.InternalTxs, ctx.InternalTxs...)
	}
	return ret, err
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (evm *EVM) Call(caller ContractRef, action *envelope.ContractTx, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if err := evm.PM.CanTransfer(caller.Name(), action.GetAssetID(), action.Value()); err != nil {
		return nil, gas, ErrInsufficientBalance
	}

	toName := action.Recipient()

	var (
		// to       = AccountRef(toName)
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

	if err := evm.PM.TransferAsset(action.Sender(), action.Recipient(), action.GetAssetID(), action.Value()); err != nil {
		return nil, gas, err
	}

	contractName := toName

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.

	contract := NewContract(caller, toName, action.Value(), gas, action.GetAssetID())

	codeHash, err := evm.PM.GetCodeHash(toName)
	if err != nil {
		return nil, gas, err
	}
	code, _ := evm.PM.GetCode(toName)
	contract.SetCallCode(codeHash, code)

	ret, err = run(evm, contract, action.GetPayload())
	runGas := gas - contract.Gas

	evm.distributeContractGas(runGas, contractName, caller.Name())

	gasTable := evm.GetCurrentGasTable()
	callValueGas := int64(gasTable.CallValueTransferGas - gasTable.CallStipend)
	if action.Value().Sign() != 0 && callValueGas > 0 {
		assetName, err := evm.PM.GetAssetName(action.GetAssetID())
		if err == nil {
			evm.distributeAssetGas(callValueGas, assetName, caller.Name())
		}
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

	return ret, contract.Gas, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (evm *EVM) CallCode(caller ContractRef, action *envelope.ContractTx, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if err := evm.PM.CanTransfer(caller.Name(), evm.AssetID, action.Value()); err != nil {
		return nil, gas, ErrInsufficientBalance
	}

	toName := action.Recipient()

	var (
		snapshot = evm.StateDB.Snapshot()
		// to       = AccountRef(caller.Name())
	)
	// initialise a new contract and set the code that is to be used by the
	// E The contract is a scoped evmironment for this execution context
	// only.
	contract := NewContract(caller, caller.Name(), action.Value(), gas, evm.AssetID)

	codeHash, err := evm.PM.GetCodeHash(toName)
	if err != nil {
		return nil, gas, err
	}
	code, _ := evm.PM.GetCode(toName)
	contract.SetCallCode(codeHash, code)

	ret, err = run(evm, contract, action.GetPayload())
	runGas := gas - contract.Gas

	contractName := toName

	evm.distributeContractGas(runGas, contractName, caller.Name())

	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}

	return ret, contract.Gas, err
}

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (evm *EVM) DelegateCall(caller ContractRef, name string, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	var (
		snapshot = evm.StateDB.Snapshot()
		// to       = AccountRef(caller.Name())
	)

	// Initialise a new contract and make initialise the delegate values
	contract := NewContract(caller, caller.Name(), nil, gas, evm.AssetID).AsDelegate()

	codeHash, err := evm.PM.GetCodeHash(name)
	if err != nil {
		return nil, gas, err
	}
	code, _ := evm.PM.GetCode(name)
	contract.SetCallCode(codeHash, code)

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
	return ret, contract.Gas, err
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (evm *EVM) StaticCall(caller ContractRef, name string, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
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
		//to       = AccountRef(name)
		snapshot = evm.StateDB.Snapshot()
	)
	// Initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := NewContract(caller, name, new(big.Int), gas, evm.AssetID)
	codeHash, err := evm.PM.GetCodeHash(name)
	if err != nil {
		return nil, gas, err
	}
	code, _ := evm.PM.GetCode(name)
	contract.SetCallCode(codeHash, code)

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in Homestead this also counts for code storage gas errors.
	ret, err = run(evm, contract, input)
	// runGas := gas - contract.Gas

	// contractName := to.Name()

	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (evm *EVM) Create(caller ContractRef, action *envelope.ContractTx, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	if err := evm.PM.CanTransfer(caller.Name(), evm.AssetID, action.Value()); err != nil {
		return nil, gas, ErrInsufficientBalance
	}

	contractName := action.Recipient()
	snapshot := evm.StateDB.Snapshot()

	if b, err := evm.PM.GetCode(contractName); err != nil {
		return nil, gas, err
	} else if len(b) != 0 {
		return nil, gas, ErrContractCodeCollision
	}

	if err := evm.PM.TransferAsset(action.Sender(), action.Recipient(), evm.AssetID, action.Value()); err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		return nil, gas, err
	}

	// initialise a new contract and set the code that is to be used by the
	// E The contract is a scoped evmironment for this execution context
	// only.
	contract := NewContract(caller, contractName, action.Value(), gas, evm.AssetID)
	contract.SetCallCode(crypto.Keccak256Hash(action.GetPayload()), action.GetPayload())

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
			if err = evm.PM.SetCode(contractName, ret); err != nil {
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

	evm.distributeContractGas(gas-contract.Gas, contractName, contractName)
	return ret, contract.Gas, err
}

// ChainConfig returns the environment's chain configuration
func (evm *EVM) ChainConfig() *params.ChainConfig { return evm.chainConfig }

// Interpreter returns the EVM interpreter
func (evm *EVM) Interpreter() *Interpreter { return evm.interpreter }
