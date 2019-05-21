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
	"math/big"

	"github.com/fractalplatform/fractal/common"
)

// ContractRef is a reference to the contract's backing object
type ContractRef interface {
	Name() common.Name
}

// AccountRef implements ContractRef.
//
// Account references are used during EVM initialisation and
// it's primary use is to fetch names. Removing this object
// proves difficult because of the cached jump destinations which
// are fetched from the parent contract (i.e. the caller), which
// is a ContractRef.
type AccountRef common.Name

// Name casts AccountRef to a Name
func (ar AccountRef) Name() common.Name { return (common.Name)(ar) }

// Contract represents an contract in the state database. It contains
// the the contract code, calling arguments. Contract implements ContractRef
type Contract struct {
	// CallerName is the result of the caller which initialised this
	// contract. However when the "call method" is delegated this value
	// needs to be initialised to that of the caller's caller.
	CallerName common.Name
	caller     ContractRef
	self       ContractRef

	jumpdests destinations // result of JUMPDEST analysis.

	Code     []byte
	CodeHash common.Hash
	CodeName *common.Name
	Input    []byte

	Gas     uint64
	value   *big.Int
	AssetId uint64

	Args []byte

	DelegateCall bool
}

// NewContract returns a new contract environment for the execution of EVM.
func NewContract(caller ContractRef, object ContractRef, value *big.Int, gas uint64, assetId uint64) *Contract {
	c := &Contract{CallerName: caller.Name(), caller: caller, self: object, Args: nil}

	if parent, ok := caller.(*Contract); ok {
		// Reuse JUMPDEST analysis from parent context if available.
		c.jumpdests = parent.jumpdests
	} else {
		c.jumpdests = make(destinations)
	}

	// Gas should be a pointer so it can safely be reduced through the run
	// This pointer will be off the state transition
	c.Gas = gas
	// ensures a value is set
	c.value = value
	c.AssetId = assetId
	return c
}

// AsDelegate sets the contract to be a delegate call and returns the current
// contract (for chaining calls)
func (c *Contract) AsDelegate() *Contract {
	c.DelegateCall = true
	// NOTE: caller must, at all times be a contract. It should never happen
	// that caller is something other than a Contract.
	parent := c.caller.(*Contract)
	c.CallerName = parent.CallerName
	c.value = parent.value

	return c
}

// GetOp returns the n'th element in the contract's byte array
func (c *Contract) GetOp(n uint64) OpCode {
	return OpCode(c.GetByte(n))
}

// GetByte returns the n'th byte in the contract's byte array
func (c *Contract) GetByte(n uint64) byte {
	if n < uint64(len(c.Code)) {
		return c.Code[n]
	}

	return 0
}

// Caller returns the caller of the contract.
//
// Caller will recursively call caller when the contract is a delegate
// call, including that of caller's caller.
func (c *Contract) Caller() common.Name {
	return c.CallerName
}

// UseGas attempts the use gas and subtracts it and returns true on success
func (c *Contract) UseGas(gas uint64) (ok bool) {
	if c.Gas < gas {
		return false
	}
	c.Gas -= gas
	return true
}

// Name returns the contracts name
func (c *Contract) Name() common.Name {
	return c.self.Name()
}

// Value returns the contracts value (sent to it from it's caller)
func (c *Contract) Value() *big.Int {
	return c.value
}

// SetCallCode sets the code of the contract and name of the backing dataobject
func (c *Contract) SetCallCode(name *common.Name, hash common.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
	c.CodeName = name

}
