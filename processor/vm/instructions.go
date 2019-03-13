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
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	bigZero                  = new(big.Int)
	tt255                    = math.BigPow(2, 255)
	errWriteProtection       = errors.New("evm: write protection")
	errReturnDataOutOfBounds = errors.New("evm: return data out of bounds")
	errExecutionReverted     = errors.New("evm: execution reverted")
	errMaxCodeSizeExceeded   = errors.New("evm: max code size exceeded")
)

func opAdd(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	math.U256(y.Add(x, y))

	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSub(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	math.U256(y.Sub(x, y))

	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opMul(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	stack.push(math.U256(x.Mul(x, y)))

	evm.interpreter.intPool.put(y)

	return nil, nil
}

func opDiv(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if y.Sign() != 0 {
		math.U256(y.Div(x, y))
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSdiv(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := evm.interpreter.intPool.getZero()

	if y.Sign() == 0 || x.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() != y.Sign() {
			res.Div(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Div(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	evm.interpreter.intPool.put(x, y)
	return nil, nil
}

func opMod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	if y.Sign() == 0 {
		stack.push(x.SetUint64(0))
	} else {
		stack.push(math.U256(x.Mod(x, y)))
	}
	evm.interpreter.intPool.put(y)
	return nil, nil
}

func opSmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := evm.interpreter.intPool.getZero()

	if y.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() < 0 {
			res.Mod(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Mod(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	evm.interpreter.intPool.put(x, y)
	return nil, nil
}

func opExp(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	base, exponent := stack.pop(), stack.pop()
	stack.push(math.Exp(base, exponent))

	evm.interpreter.intPool.put(base, exponent)

	return nil, nil
}

func opSignExtend(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	back := stack.pop()
	if back.Cmp(big.NewInt(31)) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := stack.pop()
		mask := back.Lsh(common.Big1, bit)
		mask.Sub(mask, common.Big1)
		if num.Bit(int(bit)) > 0 {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}

		stack.push(math.U256(num))
	}

	evm.interpreter.intPool.put(back)
	return nil, nil
}

func opNot(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	math.U256(x.Not(x))
	return nil, nil
}

func opLt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) < 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opGt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) > 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSlt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(1)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(0)

	default:
		if x.Cmp(y) < 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSgt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(0)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(1)

	default:
		if x.Cmp(y) > 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opEq(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) == 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opIszero(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if x.Sign() > 0 {
		x.SetUint64(0)
	} else {
		x.SetUint64(1)
	}
	return nil, nil
}

func opAnd(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	stack.push(x.And(x, y))

	evm.interpreter.intPool.put(y)
	return nil, nil
}

func opOr(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Or(x, y)

	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opXor(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Xor(x, y)

	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opByte(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	th, val := stack.pop(), stack.peek()
	if th.Cmp(common.Big32) < 0 {
		b := math.Byte(val, 32, int(th.Int64()))
		val.SetUint64(uint64(b))
	} else {
		val.SetUint64(0)
	}
	evm.interpreter.intPool.put(th)
	return nil, nil
}

func opAddmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Add(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}
	evm.interpreter.intPool.put(y, z)
	return nil, nil
}

func opMulmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Mul(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}
	evm.interpreter.intPool.put(y, z)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	defer evm.interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Lsh(value, n))

	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	defer evm.interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Rsh(value, n))

	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, S256 returns (potentially) a new bigint, so we're popping, not peeking this one
	shift, value := math.U256(stack.pop()), math.S256(stack.pop())
	defer evm.interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		if value.Sign() > 0 {
			value.SetUint64(0)
		} else {
			value.SetInt64(-1)
		}
		stack.push(math.U256(value))
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.Rsh(value, n)
	stack.push(math.U256(value))

	return nil, nil
}

func opSha3(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	data := memory.Get(offset.Int64(), size.Int64())
	hash := crypto.Keccak256(data)

	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(common.BytesToHash(hash), data)
	}
	stack.push(evm.interpreter.intPool.get().SetBytes(hash))

	evm.interpreter.intPool.put(offset, size)
	return nil, nil
}

func opAddress(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(contract.Name().Big())
	return nil, nil
}

func opGetSnapshotTime(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	time, num := stack.pop(), stack.pop()
	index := num.Uint64()
	t := time.Uint64()
	tt, err := evm.AccountDB.GetSnapshotTime(index, t)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(tt))
	}
	evm.interpreter.intPool.put(num, time)
	return nil, nil
}

func opGetAssetAmount(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	time, assetId := stack.pop(), stack.pop()
	assetID := assetId.Uint64()
	t := time.Uint64()
	amount, err := evm.AccountDB.GetAssetAmountByTime(assetID, t)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(amount)
	}
	evm.interpreter.intPool.put(time, assetId)
	return nil, nil
}

func opSnapBalance(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	opt, time, assetId, account := stack.pop(), stack.pop(), stack.pop(), stack.pop()
	o := opt.Uint64()
	assetID := assetId.Uint64()
	t := time.Uint64()
	var rerr error
	var rbalance = big.NewInt(0)

	if name, err := common.BigToName(account); err == nil {
		if balance, err := evm.AccountDB.GetBalanceByTime(name, assetID, t); err == nil {
			if (o == 1) && (assetID == evm.chainConfig.SysTokenID) {
				if dbalance, err := evm.Context.GetDelegatedByTime(name.String(), t, evm.StateDB); err == nil {
					rbalance = new(big.Int).Add(balance, dbalance)
				} else {
					rerr = err
				}
			}
		} else {
			rerr = err
		}

	} else {
		rerr = err
	}

	if rerr != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(rbalance)
	}
	evm.interpreter.intPool.put(time, assetId)
	return nil, nil
}
func opBalanceex(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	assetId := stack.pop()
	slot := stack.peek()
	name, err := common.BigToName(slot)
	if err != nil {
		slot.Set(big.NewInt(0))
		return nil, nil
	}
	account, _ := evm.AccountDB.GetAccountByName(name)
	if account == nil {
		slot.Set(big.NewInt(0))
		return nil, nil
	}
	balance, _ := account.GetBalanceByID(assetId.Uint64())
	slot.Set(balance)
	return nil, nil
}

func opBalance(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	name, err := common.BigToName(slot)
	if err != nil {
		slot.Set(big.NewInt(0))
		return nil, nil
	}
	account, _ := evm.AccountDB.GetAccountByName(name)
	if account == nil {
		slot.Set(big.NewInt(0))
		return nil, nil
	}
	balance, _ := account.GetBalanceByID(contract.AssetId)
	slot.Set(balance)
	return nil, nil
}

func opOrigin(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.Origin.Big())
	return nil, nil
}

func opCaller(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(contract.Caller().Big())
	return nil, nil
}

func opCallValue(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().Set(contract.value))
	return nil, nil
}

func opCallDataLoad(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetBytes(getDataBig(contract.Input, stack.pop(), big32)))
	return nil, nil
}

func opCallDataSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetInt64(int64(len(contract.Input))))
	return nil, nil
}

func opCallDataCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)
	memory.Set(memOffset.Uint64(), length.Uint64(), getDataBig(contract.Input, dataOffset, length))

	evm.interpreter.intPool.put(memOffset, dataOffset, length)
	return nil, nil
}

func opReturnDataSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(uint64(len(evm.interpreter.returnData))))
	return nil, nil
}

func opReturnDataCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()

		end = evm.interpreter.intPool.get().Add(dataOffset, length)
	)
	defer evm.interpreter.intPool.put(memOffset, dataOffset, length, end)

	if end.BitLen() > 64 || uint64(len(evm.interpreter.returnData)) < end.Uint64() {
		return nil, errReturnDataOutOfBounds
	}
	memory.Set(memOffset.Uint64(), length.Uint64(), evm.interpreter.returnData[dataOffset.Uint64():end.Uint64()])

	return nil, nil
}

func opExtCodeSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	name, err := common.BigToName(slot)
	if err != nil {
		slot.SetUint64(0)
		return nil, nil
	}
	account, _ := evm.AccountDB.GetAccountByName(name)
	if account == nil {
		slot.SetUint64(0)
		return nil, nil
	}
	codeSize := account.GetCodeSize()
	slot.SetUint64(uint64(codeSize))
	return nil, nil
}

func opCodeSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	l := evm.interpreter.intPool.get().SetInt64(int64(len(contract.Code)))
	stack.push(l)

	return nil, nil
}

func opCodeCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	codeCopy := getDataBig(contract.Code, codeOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	evm.interpreter.intPool.put(memOffset, codeOffset, length)
	return nil, nil
}

func opExtCodeCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	name, err := common.BigToName(stack.pop())
	if err != nil {
		memory.Set(memOffset.Uint64(), length.Uint64(), nil)
		evm.interpreter.intPool.put(memOffset, codeOffset, length)
		return nil, nil
	}
	account, _ := evm.AccountDB.GetAccountByName(name)
	if account == nil {
		memory.Set(memOffset.Uint64(), length.Uint64(), nil)
		evm.interpreter.intPool.put(memOffset, codeOffset, length)
	} else {
		code, _ := account.GetCode()
		codeCopy := getDataBig(code, codeOffset, length)
		memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)
		evm.interpreter.intPool.put(memOffset, codeOffset, length)
	}
	return nil, nil
}

func opGasprice(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().Set(evm.GasPrice))
	return nil, nil
}

func opBlockhash(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	num := stack.pop()

	n := evm.interpreter.intPool.get().Sub(evm.BlockNumber, common.Big257)
	if num.Cmp(n) > 0 && num.Cmp(evm.BlockNumber) < 0 {
		stack.push(evm.GetHash(num.Uint64()).Big())
	} else {
		stack.push(evm.interpreter.intPool.getZero())
	}
	evm.interpreter.intPool.put(num, n)
	return nil, nil
}

func opCoinbase(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.Coinbase.Big())
	return nil, nil
}

func opTimestamp(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().Set(evm.Time)))
	return nil, nil
}

func opNumber(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().Set(evm.BlockNumber)))
	return nil, nil
}

func opDifficulty(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().Set(evm.Difficulty)))
	return nil, nil
}

func opGasLimit(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().SetUint64(evm.GasLimit)))
	return nil, nil
}

func opCallAssetId(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(contract.AssetId))
	return nil, nil
}

func opPop(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	evm.interpreter.intPool.put(stack.pop())
	return nil, nil
}

func opMload(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset := stack.pop()
	val := evm.interpreter.intPool.get().SetBytes(memory.Get(offset.Int64(), 32))
	stack.push(val)

	evm.interpreter.intPool.put(offset)
	return nil, nil
}

func opMstore(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.Set(mStart.Uint64(), 32, math.PaddedBigBytes(val, 32))

	evm.interpreter.intPool.put(mStart, val)
	return nil, nil
}

func opMstore8(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	off, val := stack.pop().Int64(), stack.pop().Int64()
	memory.store[off] = byte(val & 0xff)

	return nil, nil
}

func opSload(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := common.BigToHash(stack.pop())
	val := evm.StateDB.GetState(contract.Name().String(), loc).Big()
	stack.push(val)
	return nil, nil
}

func opSstore(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := common.BigToHash(stack.pop())
	val := stack.pop()
	evm.StateDB.SetState(contract.Name().String(), loc, common.BigToHash(val))

	evm.interpreter.intPool.put(val)
	return nil, nil
}

func opJump(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos := stack.pop()
	if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
		nop := contract.GetOp(pos.Uint64())
		return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
	}
	*pc = pos.Uint64()

	evm.interpreter.intPool.put(pos)
	return nil, nil
}

func opJumpi(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos, cond := stack.pop(), stack.pop()
	if cond.Sign() != 0 {
		if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
			nop := contract.GetOp(pos.Uint64())
			return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}

	evm.interpreter.intPool.put(pos, cond)
	return nil, nil
}

func opJumpdest(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opPc(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(*pc))
	return nil, nil
}

func opMsize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetInt64(int64(memory.Len())))
	return nil, nil
}

func opGas(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(contract.Gas))
	return nil, nil
}

func opCreate(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// var (
	// 	value        = stack.pop()
	// 	offset, size = stack.pop(), stack.pop()
	// 	input        = memory.Get(offset.Int64(), size.Int64())
	// 	gas          = contract.Gas
	// 	assetAddr    = contract.AssetId
	// )
	// //if evm.ChainConfig().IsEIP150(evm.BlockNumber) {
	// //	gas -= gas / 64
	// //}

	// contract.UseGas(gas)
	// // todo
	// action := types.NewAction(types.CreateContract, contract.Name(), "", 0, gas, value, input)

	// res, name, returnGas, suberr := evm.Create(contract, action, gas)
	// // Push item on the stack based on the returned error. If the ruleset is
	// // homestead we must check for CodeStoreOutOfGasError (homestead only
	// // rule) and treat as an error, if the ruleset is frontier we must
	// // ignore this error and pretend the operation was successful.
	// //if evm.ChainConfig().IsHomestead(evm.BlockNumber) && suberr == ErrCodeStoreOutOfGas {
	// //	stack.push(evm.interpreter.intPool.getZero())
	// if suberr != nil && suberr != ErrCodeStoreOutOfGas {
	// 	stack.push(evm.interpreter.intPool.getZero())
	// } else {
	// 	stack.push(name.Big())
	// }
	// contract.Gas += returnGas
	// evm.interpreter.intPool.put(value, offset, size)

	// if suberr == errExecutionReverted {
	// 	return res, nil
	// }
	return nil, nil
}

func opCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	name, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toName, _ := common.BigToName(name)
	value = math.U256(value)
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	action := types.NewAction(types.CallContract, contract.Name(), toName, 0, evm.AssetID, gas, value, args)

	ret, returnGas, err := evm.Call(contract, action, gas)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(name, value, inOffset, inSize, retOffset, retSize)
	if evm.vmConfig.ContractLogFlag {
		internalLog := &types.InternalLog{Action: action.NewRPCAction(0), ActionType: "call", GasUsed: gas - returnGas, GasLimit: gas, Error: err.Error()}
		evm.InternalTxs = append(evm.InternalTxs, internalLog)
	}
	return ret, nil
}

func opCallCode(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	//addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	name, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	//addr, assetId,value, inOffset, inSize, retOffset, retSize := stack.pop(),stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toName, _ := common.BigToName(name)
	value = math.U256(value)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}
	// todo
	action := types.NewAction(types.CallContract, contract.Name(), toName, 0, evm.AssetID, gas, value, args)

	ret, returnGas, err := evm.CallCode(contract, action, gas)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(name, value, inOffset, inSize, retOffset, retSize)
	if evm.vmConfig.ContractLogFlag {
		internalLog := &types.InternalLog{Action: action.NewRPCAction(0), ActionType: "callcode", GasUsed: gas - returnGas, GasLimit: gas, Error: err.Error()}
		evm.InternalTxs = append(evm.InternalTxs, internalLog)
	}
	return ret, nil
}

func opDelegateCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	name, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toName, _ := common.BigToName(name)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := evm.DelegateCall(contract, toName, args, gas)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(name, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

//multi-asset
//Increase asset already exist
func opAddAsset(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	value, to, assetId := stack.pop(), stack.pop(), stack.pop()
	assetID := assetId.Uint64()
	toName, _ := common.BigToName(to)
	value = math.U256(value)

	err := execAddAsset(evm, contract, assetID, toName, value)

	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	return nil, nil
}

func execAddAsset(evm *EVM, contract *Contract, assetID uint64, toName common.Name, value *big.Int) error {
	asset := &accountmanager.IncAsset{AssetId: assetID, Amount: value, To: toName}
	b, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return err
	}
	action := types.NewAction(types.IncreaseAsset, contract.CallerName, "", 0, 0, 0, big.NewInt(0), b)

	err = evm.AccountDB.Process(action)
	if evm.vmConfig.ContractLogFlag {
		internalLog := &types.InternalLog{Action: action.NewRPCAction(0), ActionType: "addasset", GasUsed: 0, GasLimit: contract.Gas, Error: err.Error()}
		evm.InternalTxs = append(evm.InternalTxs, internalLog)
	}
	return err
}

//issue an asset for multi-asset
func opIssueAsset(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.Get(offset.Int64(), size.Int64())
	ret = bytes.TrimRight(ret, "\x00")
	desc := string(ret)

	assetId, err := executeIssuseAsset(evm, contract, desc)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(assetId))
	}
	evm.interpreter.intPool.put(offset, size)
	return nil, nil
}

func executeIssuseAsset(evm *EVM, contract *Contract, desc string) (uint64, error) {
	input := strings.Split(desc, ",")
	if len(input) != 7 {
		return 0, fmt.Errorf("invalid desc string")
	}
	name := input[0]
	symbol := input[1]
	total, ifOK := new(big.Int).SetString(input[2], 10)
	if !ifOK {
		return 0, fmt.Errorf("amount not correct")
	}
	decimal, err := strconv.ParseUint(input[3], 10, 64)
	if err != nil {
		return 0, err
	}
	owner := common.Name(input[4])
	limit, ifOK := new(big.Int).SetString(input[5], 10)
	if !ifOK {
		return 0, fmt.Errorf("amount not correct")
	}
	founder := common.Name(input[6])
	asset := &asset.AssetObject{AssetName: name, Symbol: symbol, Amount: total, Owner: owner, Founder: founder, Decimals: decimal, UpperLimit: limit}

	b, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return 0, err
	}
	action := types.NewAction(types.IssueAsset, contract.CallerName, "", 0, 0, 0, big.NewInt(0), b)

	err = evm.AccountDB.Process(action)
	if evm.vmConfig.ContractLogFlag {
		internalLog := &types.InternalLog{Action: action.NewRPCAction(0), ActionType: "issueasset", GasUsed: 0, GasLimit: contract.Gas, Error: err.Error()}
		evm.InternalTxs = append(evm.InternalTxs, internalLog)
	}
	if err != nil {
		return 0, err
	} else {
		assetInfo, err := evm.AccountDB.GetAssetInfoByName(name)
		if err != nil {
			return 0, err
		} else {
			return assetInfo.AssetId, nil
		}
	}
}

//issue an asset for multi-asset
func opSetAssetOwner(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	newOwner, assetId := stack.pop(), stack.pop()
	newOwnerName, _ := common.BigToName(newOwner)
	assetID := assetId.Uint64()

	err := execSetAssetOwner(evm, contract, assetID, newOwnerName)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	evm.interpreter.intPool.put(newOwner, assetId)
	return nil, nil
}

func execSetAssetOwner(evm *EVM, contract *Contract, assetID uint64, owner common.Name) error {
	asset := &asset.AssetObject{AssetId: assetID, Owner: owner}
	b, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return err
	}

	action := types.NewAction(types.SetAssetOwner, contract.CallerName, "", 0, 0, 0, big.NewInt(0), b)
	if evm.vmConfig.ContractLogFlag {
		internalLog := &types.InternalLog{Action: action.NewRPCAction(0), ActionType: "setassetowner", GasUsed: 0, GasLimit: contract.Gas, Error: err.Error()}
		evm.InternalTxs = append(evm.InternalTxs, internalLog)
	}
	return evm.AccountDB.Process(action)

}

func opCallEx(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var ret []byte
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	name, assetId, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toName, _ := common.BigToName(name)
	assetID := assetId.Uint64()
	value = math.U256(value)
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	action := types.NewAction(types.CallContract, contract.Name(), toName, 0, assetID, gas, value, args)

	ret, returnGas, err := evm.Call(contract, action, gas)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(name, value, inOffset, inSize, retOffset, retSize)
	if evm.vmConfig.ContractLogFlag {
		internalLog := &types.InternalLog{Action: action.NewRPCAction(0), ActionType: "transferex", GasUsed: 0, GasLimit: gas, Error: err.Error()}
		evm.InternalTxs = append(evm.InternalTxs, internalLog)
	}
	return ret, nil
}

func opStaticCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	name, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toName, _ := common.BigToName(name)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := evm.StaticCall(contract, toName, args, gas)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(name, inOffset, inSize, retOffset, retSize)
	if evm.vmConfig.ContractLogFlag {
		internalLog := &types.InternalLog{ActionType: "staticcall", GasUsed: gas - returnGas, GasLimit: gas, Error: err.Error()}
		evm.InternalTxs = append(evm.InternalTxs, internalLog)
	}
	return ret, nil
}

func opReturn(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	evm.interpreter.intPool.put(offset, size)
	return ret, nil
}

func opRevert(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	evm.interpreter.intPool.put(offset, size)
	return ret, nil
}

func opStop(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	//todo
	// contractCreater := common.BigToAddress(stack.pop())

	// assets, err := evm.AccountDB.GetUserAssets(contract.Name())
	// if err != nil {
	// 	return nil, nil
	// }
	// for _, asset := range assets {
	// 	balance := evm.AccountDB.GetBalance(contract.Name(), asset.AssetID)
	// 	evm.Asset.AddBalance(contractCreater, balance)
	// }

	//todo
	//evm.StateDB.Suicide(contract.Address())
	return nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		topics := make([]common.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			topics[i] = common.BigToHash(stack.pop())
		}

		d := memory.Get(mStart.Int64(), mSize.Int64())
		evm.StateDB.AddLog(&types.Log{
			Name:   contract.Name(),
			Topics: topics,
			Data:   d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: evm.BlockNumber.Uint64(),
		})

		evm.interpreter.intPool.put(mStart, mSize)
		return nil, nil
	}
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		codeLen := len(contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := evm.interpreter.intPool.get()
		stack.push(integer.SetBytes(common.RightPadBytes(contract.Code[startMin:endMin], pushByteSize)))

		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.dup(evm.interpreter.intPool, int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size += 1
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.swap(int(size))
		return nil, nil
	}
}
