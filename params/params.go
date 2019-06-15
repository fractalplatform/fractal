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
	// BlockGasLimit Gas limit of the  block.
	BlockGasLimit uint64 = 30000000
	// MinGasLimit Minimum the gas limit may ever be.
	MinGasLimit uint64 = 5000
	// GasLimitBoundDivisor The bound divisor of the gas limit, used in update calculations.
	GasLimitBoundDivisor uint64 = 1024
	// MaximumExtraDataSize Maximum size extra data may be after Genesis.
	MaximumExtraDataSize uint64 = 32 + 65

	EpochDuration   uint64 = 30000 // Duration between proof-of-work epochs.
	CallCreateDepth uint64 = 1024  // Maximum depth of call/create stack.
	StackLimit      uint64 = 1024  // Maximum size of VM stack allowed.

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
	MaxAuthorNum  = uint64(10)
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
