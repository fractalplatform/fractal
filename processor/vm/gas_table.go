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

package vm

import (
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
)

// memoryGasCosts calculates the quadratic gas for memory expansion. It does so
// only for the memory region that is expanded, not the total memory.
func memoryGasCost(gt params.GasTable, mem *Memory, newMemSize uint64) (uint64, error) {

	if newMemSize == 0 {
		return 0, nil
	}
	// The maximum that will fit in a uint64 is max_word_count - 1
	// anything above that will result in an overflow.
	// Additionally, a newMemSize which results in a
	// newMemSizeWords larger than 0x7ffffffff will cause the square operation
	// to overflow.
	// The constant 0xffffffffe0 is the highest number that can be used without
	// overflowing the gas calculation
	if newMemSize > 0xffffffffe0 {
		return 0, errGasUintOverflow
	}

	newMemSizeWords := toWordSize(newMemSize)
	newMemSize = newMemSizeWords * 32

	if newMemSize > uint64(mem.Len()) {
		square := newMemSizeWords * newMemSizeWords
		linCoef := newMemSizeWords * gt.MemoryGas
		quadCoef := square / gt.QuadCoeffDiv
		newTotalFee := linCoef + quadCoef

		fee := newTotalFee - mem.lastGasCost
		mem.lastGasCost = newTotalFee

		return fee, nil
	}
	return 0, nil
}

func constGasFunc(gas uint64) gasFunc {
	return func(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		return gas, nil
	}
}

func gasCallDataCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	words, overflow := bigUint64(stack.Back(2))
	if overflow {
		return 0, errGasUintOverflow
	}

	if words, overflow = math.SafeMul(toWordSize(words), gt.CopyGas); overflow {
		return 0, errGasUintOverflow
	}

	if gas, overflow = math.SafeAdd(gas, words); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasReturnDataCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	words, overflow := bigUint64(stack.Back(2))
	if overflow {
		return 0, errGasUintOverflow
	}

	if words, overflow = math.SafeMul(toWordSize(words), gt.CopyGas); overflow {
		return 0, errGasUintOverflow
	}

	if gas, overflow = math.SafeAdd(gas, words); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasSStore(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		y, x = stack.Back(1), stack.Back(0)
		val  = evm.StateDB.GetState(contract.Name().String(), common.BigToHash(x))
	)
	// This checks for 3 scenario's and calculates gas accordingly
	// 1. From a zero-value address to a non-zero value         (NEW VALUE)
	// 2. From a non-zero value address to a zero-value address (DELETE)
	// 3. From a non-zero to a non-zero                         (CHANGE)
	if val == (common.Hash{}) && y.Sign() != 0 {
		// 0 => non 0
		return gt.SstoreSetGas, nil
	} else if val != (common.Hash{}) && y.Sign() == 0 {
		// non 0 => 0
		//evm.StateDB.AddRefund(params.SstoreRefundGas)
		return 0, nil
	} else {
		// non 0 => non 0 (or 0 => 0)
		return gt.SstoreResetGas, nil
	}
}

func makeGasLog(n uint64) gasFunc {
	return func(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		requestedSize, overflow := bigUint64(stack.Back(1))
		if overflow {
			return 0, errGasUintOverflow
		}

		gas, err := memoryGasCost(gt, mem, memorySize)
		if err != nil {
			return 0, err
		}

		if gas, overflow = math.SafeAdd(gas, gt.LogGas); overflow {
			return 0, errGasUintOverflow
		}
		if gas, overflow = math.SafeAdd(gas, n*gt.LogTopicGas); overflow {
			return 0, errGasUintOverflow
		}

		var memorySizeGas uint64
		if memorySizeGas, overflow = math.SafeMul(requestedSize, gt.LogDataGas); overflow {
			return 0, errGasUintOverflow
		}
		if gas, overflow = math.SafeAdd(gas, memorySizeGas); overflow {
			return 0, errGasUintOverflow
		}
		return gas, nil
	}
}

func gasSha3(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}

	if gas, overflow = math.SafeAdd(gas, gt.Sha3Gas); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.Back(1))
	if overflow {
		return 0, errGasUintOverflow
	}
	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), gt.Sha3WordGas); overflow {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCodeCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.Back(2))
	if overflow {
		return 0, errGasUintOverflow
	}
	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), gt.CopyGas); overflow {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasExtCodeCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, gt.ExtcodeCopy); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.Back(3))
	if overflow {
		return 0, errGasUintOverflow
	}

	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), gt.CopyGas); overflow {
		return 0, errGasUintOverflow
	}

	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMLoad(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMStore8(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMStore(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCreate(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, gt.CreateGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasBalance(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.Balance, nil
}

func gasExtCodeSize(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.ExtcodeSize, nil
}

func gasSLoad(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.SLoad, nil
}

func gasExp(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	expByteLen := uint64((stack.data[stack.len()-2].BitLen() + 7) / 8)

	var (
		gas      = expByteLen * gt.ExpByte // no overflow check required. Max is 256 * ExpByte gas
		overflow bool
	)
	if gas, overflow = math.SafeAdd(gas, GasSlowStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCall(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		gas            = gt.Calls
		transfersValue = stack.Back(2).Sign() != 0
	)
	if evm.ForkID >= params.ForkID4 {
		if p := PrecompiledContracts[stack.Back(1).Uint64()]; p != nil {
			return gas, nil
		}
	}
	if transfersValue {
		gas += gt.CallValueTransferGas
	}
	memoryGas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, memoryGas); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCallWithPay(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		gas            = gt.Calls
		transfersValue = stack.Back(3).Sign() != 0
	)
	if transfersValue {
		gas += gt.CallValueTransferGas
	}
	memoryGas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, memoryGas); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCallCode(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas := gt.Calls
	if stack.Back(2).Sign() != 0 {
		gas += gt.CallValueTransferGas
	}
	memoryGas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, memoryGas); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasReturn(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(gt, mem, memorySize)
}

func gasRevert(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(gt, mem, memorySize)
}

func gasInvalid(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return 0, nil
}

func gasSuicide(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var gas uint64
	return gas, nil
}

func gasDelegateCall(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, gt.Calls); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasGetEpoch(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.GetEpoch, nil
}

func gasGetCandidateNum(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.GetCandidateNum, nil
}

func gasGetCandidate(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.GetCandidate, nil
}

func gasGetVoterStake(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.GetVoterStake, nil
}

func gasIssueAsset(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.IssueAsset, nil
}

func gasBalanceex(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.Balance, nil
}

func gasGetAssetInfo(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.GetAssetInfo, nil
}

func gasSnapBalance(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.SnapBalance, nil
}

func gasGetAccountTime(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.GetAccountTime, nil
}

func gasGetSnapshotTime(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.GetSnapshotTime, nil
}

func gasCryptoCalc(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.CryptoCalc, nil
}

func gasDeductGas(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.DeductGas, nil
}

func gasAddAsset(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {

	return gt.AddAsset, nil
}

func gasGetAccountID(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {

	return gt.GetAccountID, nil
}

func gasGetAssetID(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {

	return gt.GetAssetID, nil
}

func gasDestroyAsset(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {

	return gt.DestroyAsset, nil
}

func gasSetAssetOwner(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.SetOwner, nil
}

func gasWithdrawFee(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.WithdrawFee, nil
}

func gasCallEx(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.CallValueTransferGas + gt.Calls, nil
}

func gasStaticCall(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(gt, mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, gt.Calls); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasPush(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

func gasJumpDest(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.JumpdestGas, nil
}
func gasSwap(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

func gasDup(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}
