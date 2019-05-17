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

package params

import "math/big"

const (
	// GenesisGasLimit Gas limit of the Genesis block.
	GenesisGasLimit uint64 = 30000000
	// MinGasLimit Minimum the gas limit may ever be.
	MinGasLimit uint64 = 5000
	// GasLimitBoundDivisor The bound divisor of the gas limit, used in update calculations.
	GasLimitBoundDivisor uint64 = 1024
	// MaximumExtraDataSize Maximum size extra data may be after Genesis.
	MaximumExtraDataSize uint64 = 32 + 65
	// ActionGas Per action not creating a contract. NOTE: Not payable on data of calls between transactions.
	ActionGas uint64 = 21000
	// ActionGasContractCreation Per action that creates a contract. NOTE: Not payable on data of calls between transactions.
	ActionGasContractCreation uint64 = 53000
	// TxDataNonZeroGas Per byte of data attached to a transaction that is not equal to zero. NOTE:Not payable on data of calls between transactions.
	TxDataNonZeroGas uint64 = 68
	// TxDataZeroGas Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	TxDataZeroGas uint64 = 4

	//VM const

	CallValueTransferGas uint64 = 9000  // Paid for CALL when the value transfer is non-zero.
	QuadCoeffDiv         uint64 = 512   // Divisor for the quadratic particle of the memory cost equation.
	SstoreSetGas         uint64 = 20000 // Once per SLOAD operation.
	LogDataGas           uint64 = 8     // Per byte in a LOG* operation's data.
	CallStipend          uint64 = 2300  // Free gas given at beginning of call.

	Sha3Gas         uint64 = 30    // Once per SHA3 operation.
	Sha3WordGas     uint64 = 6     // Once per word of the SHA3 operation's data.
	SstoreResetGas  uint64 = 5000  // Once per SSTORE operation if the zeroness changes from zero.
	JumpdestGas     uint64 = 1     // Refunded gas, once per SSTORE operation if the zeroness changes to zero.
	EpochDuration   uint64 = 30000 // Duration between proof-of-work epochs.
	CreateDataGas   uint64 = 200   //
	CallCreateDepth uint64 = 1024  // Maximum depth of call/create stack.
	LogGas          uint64 = 375   // Per LOG* operation.
	CopyGas         uint64 = 3     //
	StackLimit      uint64 = 1024  // Maximum size of VM stack allowed.
	LogTopicGas     uint64 = 375   // Multiplied by the * of the LOG*, per LOG transaction. e.g. LOG0 incurs 0 * c_txLogTopicGas, LOG4 incurs 4 * c_txLogTopicGas.
	CreateGas       uint64 = 32000 // Once per CREATE operation & contract-creation transaction.
	MemoryGas       uint64 = 3     // Times the address of the (highest referenced byte in memory + 1). NOTE: referencing happens on read, write and in instructions such as RETURN and CALL.

	MaxCodeSize uint64 = 24576     // Maximum bytecode to permit for a contract
	MaxTxSize   uint64 = 32 * 1024 // Heuristic limit, reject transactions over 32KB to prfeed DOS attacks

	// Precompiled contract gas prices

	EcrecoverGas            uint64 = 3000   // Elliptic curve sender recovery gas price
	Sha256BaseGas           uint64 = 60     // Base price for a SHA256 operation
	Sha256PerWordGas        uint64 = 12     // Per-word price for a SHA256 operation
	Ripemd160BaseGas        uint64 = 600    // Base price for a RIPEMD160 operation
	Ripemd160PerWordGas     uint64 = 120    // Per-word price for a RIPEMD160 operation
	IdentityBaseGas         uint64 = 15     // Base price for a data copy operation
	IdentityPerWordGas      uint64 = 3      // Per-work price for a data copy operation
	ModExpQuadCoeffDiv      uint64 = 20     // Divisor for the quadratic particle of the big int modular exponentiation
	Bn256AddGas             uint64 = 500    // Gas needed for an elliptic curve addition
	Bn256ScalarMulGas       uint64 = 40000  // Gas needed for an elliptic curve scalar multiplication
	Bn256PairingBaseGas     uint64 = 100000 // Base price for an elliptic curve pairing check
	Bn256PairingPerPointGas uint64 = 80000  // Per-point price for an elliptic curve pairing check
)

var (
	DifficultyBoundDivisor = big.NewInt(2048)   // The bound divisor of the difficulty, used in the update calculations.
	GenesisDifficulty      = big.NewInt(131072) // Difficulty of the Genesis block.
	MinimumDifficulty      = big.NewInt(131072) // The minimum that the difficulty may ever be.
	DurationLimit          = big.NewInt(13)     // The decision boundary on the blocktime duration used to determine whether difficulty should go up or not.
)

const (
	MaxSignDepth  = uint64(10)
	MaxSignLength = uint64(50)
)

//type for fee
const (
	AssetFeeType    = uint64(0)
	ContractFeeType = uint64(1)
	CoinbaseFeeType = uint64(2)
)

//rpc max fee result count
const (
	MaxFeeResultCount = uint64(1000)
)
