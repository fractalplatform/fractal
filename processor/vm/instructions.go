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
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/crypto/ecies"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
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

	// if evm.vmConfig.EnablePreimageRecording {
	// 	evm.StateDB.AddPreimage(common.BytesToHash(hash), data)
	// }
	stack.push(evm.interpreter.intPool.get().SetBytes(hash))

	evm.interpreter.intPool.put(offset, size)
	return nil, nil
}

func opAddress(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	address, _ := common.StringToAddress(contract.Name())
	stack.push(evm.interpreter.intPool.get().SetBytes(address.Bytes()))
	return nil, nil
}

func opBalanceex(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	assetId := stack.pop()
	slot := stack.peek()
	accountName := common.BigToAddress(slot).AccountName()

	balance, err := evm.PM.GetBalance(accountName, assetId.Uint64())
	if err != nil {
		slot.Set(big.NewInt(0))
	} else {
		slot.Set(balance)
	}
	return nil, nil
}

func opBalance(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	accountName := common.BigToAddress(slot).AccountName()

	balance, err := evm.PM.GetBalance(accountName, contract.AssetID)
	if err != nil {
		slot.Set(big.NewInt(0))
	} else {
		slot.Set(balance)
	}
	return nil, nil
}

func opOrigin(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	address, _ := common.StringToAddress(evm.Origin)
	stack.push(evm.interpreter.intPool.get().SetBytes(address.Bytes()))
	return nil, nil
}

func opRecipient(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	address, _ := common.StringToAddress(evm.Recipient)
	stack.push(evm.interpreter.intPool.get().SetBytes(address.Bytes()))
	return nil, nil
}

func opCaller(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	address, _ := common.StringToAddress(contract.Caller())
	stack.push(evm.interpreter.intPool.get().SetBytes(address.Bytes()))
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
	accountName := common.BigToAddress(slot).AccountName()
	if evm.PM.IsPlugin(accountName) {
		slot.SetUint64(1)
		return nil, nil
	}
	code, err := evm.PM.GetCode(accountName)
	if err != nil {
		slot.SetUint64(0)
	} else {
		slot.SetUint64(uint64(len(code)))
	}

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
		addr       = stack.pop()
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	accountName := common.BigToAddress(addr).AccountName()
	code, err := evm.PM.GetCode(accountName)
	if err != nil {
		memory.Set(memOffset.Uint64(), length.Uint64(), nil)
		evm.interpreter.intPool.put(memOffset, codeOffset, length)
	} else {
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
	address, _ := common.StringToAddress(evm.Coinbase)
	stack.push(evm.interpreter.intPool.get().SetBytes(address.Bytes()))
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
	stack.push(evm.interpreter.intPool.get().SetUint64(contract.AssetID))
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
	val := evm.StateDB.GetState(contract.Name(), loc).Big()
	stack.push(val)
	return nil, nil
}

func opSstore(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := common.BigToHash(stack.pop())
	val := stack.pop()
	evm.StateDB.SetState(contract.Name(), loc, common.BigToHash(val))
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
	_, _, _ = stack.pop(), stack.pop(), stack.pop()
	stack.push(evm.interpreter.intPool.getZero())
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

func opCallPluginWeak(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	gasUsed := evm.interpreter.gasTable.PluginCall
	if gas < gasUsed {
		stack.push(evm.interpreter.intPool.getZero())
		return nil, ErrOutOfGas
	}
	// Pop other call parameters.
	name, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	value = math.U256(value)
	accountName := common.BigToAddress(name).AccountName()
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += evm.interpreter.gasTable.CallStipend
	}

	action, _ := envelope.NewPluginTx(envelope.PayloadType(0xff), contract.Name(), accountName, 0, evm.ChainConfig().SysTokenID, evm.ChainConfig().SysTokenID, gas, evm.GasPrice, value, args, nil)

	var ret []byte
	var err error

	ctx := &plugin.Context{
		ChainContext: NewPContext(evm),
		Coinbase:     evm.Context.Coinbase,
		GasLimit:     evm.Context.GasLimit,
		BlockNumber:  evm.Context.BlockNumber.Uint64(),
		Time:         evm.Context.Time.Uint64(),
		Difficulty:   evm.Context.Difficulty.Uint64(),
		InternalTxs:  make([]*types.InternalTx, 0), // evm.InternalTxs?
	}

	ret, err = evm.PM.ExecTx(types.NewTransaction(action), ctx, true)
	returnGas := gas - gasUsed
	contract.Gas += returnGas

	if evm.vmConfig.ContractLogFlag {
		evm.InternalTxs = append(evm.InternalTxs, ctx.InternalTxs...) // the order of the logs is chaotic
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		internalAction := &types.InternalTx{
			From:       action.Sender(),
			To:         action.Recipient(),
			Amount:     action.Value(),
			Data:       action.Payload,
			ReturnData: ret,
			Type:       types.PluginCall,
			GasUsed:    gasUsed,
			GasLimit:   gas,
			Depth:      uint64(evm.depth),
			Error:      errmsg}
		evm.InternalTxs = append(evm.InternalTxs, internalAction)
	}

	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	evm.interpreter.intPool.put(name, value, inOffset, inSize, retOffset, retSize)

	return ret, nil
}

/*
func opCallPlugin(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	pluginGas := evm.interpreter.gasTable.PluginCall
	if !contract.UseGas(pluginGas) {
		stack.push(evm.interpreter.intPool.getZero())
		return nil, ErrOutOfGas
	}

	// Pop other call parameters.
	name, assetId, actype, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	value = math.U256(value)
	accountName := common.BigToAddress(name).AccountName()
	assetID := assetId.Uint64()
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += evm.interpreter.gasTable.CallStipend
	}

	action, _ := envelope.NewPluginTx(envelope.PayloadType(actype.Uint64()), contract.Name(), accountName, 0, assetID, evm.ChainConfig().SysTokenID, gas, evm.GasPrice, value, args, nil)

	var ret []byte
	var err error

	ret, err = evm.PM.ExecTx(types.NewTransaction(action), true)
	if evm.vmConfig.ContractLogFlag {
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		internalAction := &types.InternalTx{
			Type:     types.PluginCall,
			GasUsed:  pluginGas,
			GasLimit: gas,
			Depth:    uint64(evm.depth),
			Error:    errmsg}
		evm.InternalTxs = append(evm.InternalTxs, internalAction)
	}

	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	evm.interpreter.intPool.put(name, value, inOffset, inSize, retOffset, retSize)

	return ret, nil
}
*/
func opCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	bypassPlugin := common.BigToAddress(stack.Back(1)).AccountName()
	if evm.PM.IsPlugin(bypassPlugin) {
		return opCallPluginWeak(pc, evm, contract, memory, stack)
	}
	// Pop gas. The actual gas in in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	name, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	value = math.U256(value)

	accountName := common.BigToAddress(name).AccountName()

	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += evm.interpreter.gasTable.CallStipend
	}

	var ret []byte
	var err error
	if p := PrecompiledContracts[name.String()]; p != nil {
		ret, err = RunPrecompiledContract(p, args, contract)
	} else {
		action, _ := envelope.NewContractTx(envelope.CallContract, contract.Name(), accountName, 0, evm.AssetID, evm.ChainConfig().SysTokenID, gas, big.NewInt(0), value, args, nil)
		var returnGas uint64
		ret, returnGas, err = evm.Call(contract, action, gas)
		contract.Gas += returnGas

		if evm.vmConfig.ContractLogFlag {
			errmsg := ""
			if err != nil {
				errmsg = err.Error()
			}
			internalAction := &types.InternalTx{
				From:     action.Sender(),
				To:       action.Recipient(),
				Amount:   action.Value(),
				Type:     types.Call,
				GasUsed:  gas - returnGas,
				GasLimit: gas,
				Depth:    uint64(evm.depth),
				Error:    errmsg}
			evm.InternalTxs = append(evm.InternalTxs, internalAction)
		}
	}

	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	evm.interpreter.intPool.put(name, value, inOffset, inSize, retOffset, retSize)

	return ret, nil
}

func opCallWithPay(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	name, assetId, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	value = math.U256(value)
	accountName := common.BigToAddress(name).AccountName()
	assetID := assetId.Uint64()
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += evm.interpreter.gasTable.CallStipend
	}

	var ret []byte
	var err error
	if p := PrecompiledContracts[name.String()]; p != nil {
		ret, err = RunPrecompiledContract(p, args, contract)
	} else {
		action, _ := envelope.NewContractTx(envelope.CallContract, contract.Name(), accountName, 0, assetID, evm.ChainConfig().SysTokenID, gas, big.NewInt(0), value, args, nil)
		var returnGas uint64
		ret, returnGas, err = evm.Call(contract, action, gas)
		contract.Gas += returnGas

		if evm.vmConfig.ContractLogFlag {
			errmsg := ""
			if err != nil {
				errmsg = err.Error()
			}
			internalAction := &types.InternalTx{
				From:     action.Sender(),
				To:       action.Recipient(),
				Amount:   action.Value(),
				Type:     types.CallWithPay,
				GasUsed:  gas - returnGas,
				GasLimit: gas,
				Depth:    uint64(evm.depth),
				Error:    errmsg}
			evm.InternalTxs = append(evm.InternalTxs, internalAction)
		}
	}

	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}

	evm.interpreter.intPool.put(name, value, inOffset, inSize, retOffset, retSize)

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
	//toName, _ := common.BigToName(name)

	accountName := common.BigToAddress(name).AccountName()

	value = math.U256(value)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += evm.interpreter.gasTable.CallStipend
	}

	action, _ := envelope.NewContractTx(envelope.CallContract, contract.Name(), accountName, 0, evm.AssetID, evm.ChainConfig().SysTokenID, gas, big.NewInt(0), value, args, nil)

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
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		internalAction := &types.InternalTx{
			From:     action.Sender(),
			To:       action.Recipient(),
			Amount:   action.Value(),
			Type:     types.CallCode,
			GasUsed:  gas - returnGas,
			GasLimit: gas,
			Depth:    uint64(evm.depth),
			Error:    errmsg}
		evm.InternalTxs = append(evm.InternalTxs, internalAction)
	}
	return ret, nil
}

func opDelegateCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	name, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()

	accountName := common.BigToAddress(name).AccountName()

	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := evm.DelegateCall(contract, accountName, args, gas)
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

// opDeductGas use to deduct gas
func opDeductGas(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	gasAmount := stack.pop()
	amount := gasAmount.Uint64()

	if contract.Gas >= amount {
		contract.Gas = contract.Gas - amount
		stack.push(evm.interpreter.intPool.get().SetUint64(contract.Gas))
	} else {
		contract.Gas = 0
		stack.push(evm.interpreter.intPool.getZero())
	}

	evm.interpreter.intPool.put(gasAmount)
	return nil, nil
}

// opCryptoCalc to encrypt or decrypt bytes
func opCryptoCalc(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	typeID, retOffset, retSize, keyOffset, keySize, dataOffset, dataSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	//
	data := memory.Get(dataOffset.Int64(), dataSize.Int64())
	key := memory.Get(keyOffset.Int64(), keySize.Int64())
	i := typeID.Uint64()
	//
	var ret = make([]byte, retSize.Int64()*32)
	var datalen int
	var ecdsapubkey *ecdsa.PublicKey
	var ecdsaprikey *ecdsa.PrivateKey
	var err error

	//consume gas per byte
	if contract.Gas >= uint64(dataSize.Int64())*params.GasTableInstance.CryptoByte {
		contract.Gas = contract.Gas - uint64(dataSize.Int64())*params.GasTableInstance.CryptoByte
	} else {
		contract.Gas = 0
		stack.push(evm.interpreter.intPool.getZero())
		return nil, nil
	}

	if i == 0 {
		//Encrypt
		ecdsapubkey, err = crypto.UnmarshalPubkey(key)
		if err == nil {
			eciespubkey := ecies.ImportECDSAPublic(ecdsapubkey)
			ret, err = ecies.Encrypt(rand.Reader, eciespubkey, data, nil, nil)
			if err == nil {
				datalen = len(ret)
				if uint64(datalen) > retSize.Uint64()*32 {
					err = errors.New("Encrypt error")
				}
			}
		}
	} else if i == 1 {
		ecdsaprikey, err = crypto.ToECDSA(key)
		if err == nil {
			eciesprikey := ecies.ImportECDSA(ecdsaprikey)
			//ret, err = prv1.Decrypt(data, nil, nil)
			ret, err = eciesprikey.Decrypt(data, nil, nil)
			if err == nil {
				datalen = len(ret)
				if uint64(datalen) > retSize.Uint64()*32 {
					err = errors.New("Decrypt error")
				}
			}
		}
	}

	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
	} else {
		//write datalen data real length
		stack.push(evm.interpreter.intPool.get().SetUint64(uint64(datalen)))
		//write data
		memory.Set(retOffset.Uint64(), uint64(datalen), ret)
	}

	evm.interpreter.intPool.put(dataOffset, dataSize, keyOffset, keySize, typeID)
	return nil, nil
}

func opStaticCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	gas := evm.callGasTemp
	// Pop other call parameters.
	name, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()

	accountName := common.BigToAddress(name).AccountName()

	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := evm.StaticCall(contract, accountName, args, gas)
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
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		internalAction := &types.InternalTx{
			Type:     types.StaticCall,
			GasUsed:  gas - returnGas,
			GasLimit: gas,
			Depth:    uint64(evm.depth),
			Error:    errmsg}
		evm.InternalTxs = append(evm.InternalTxs, internalAction)
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

func opInvalid(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	log.Error("invalid opcode ")
	return nil, nil
}

func opSuicide(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	//todo
	// contractCreater := common.BigToAddress(stack.pop())

	// assets, err := evm.PM.GetUserAssets(contract.Name())
	// if err != nil {
	// 	return nil, nil
	// }
	// for _, asset := range assets {
	// 	balance := evm.PM.GetBalance(contract.Name(), asset.AssetID)
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
