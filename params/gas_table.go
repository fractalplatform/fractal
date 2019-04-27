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

import (
	"github.com/fractalplatform/fractal/types"
)

// GasTable organizes gas prices for different phases.
type GasTable struct {
	ExtcodeSize uint64
	ExtcodeCopy uint64
	Balance     uint64
	SLoad       uint64
	Calls       uint64
	Suicide     uint64

	ExpByte uint64

	// CreateBySuicide occurs when the
	// refunded account is one that does
	// not exist. This logic is similar
	// to call. May be left nil. Nil means
	// not charged.
	CreateBySuicide uint64
	SetOwner        uint64
	GetAccountTime  uint64
	GetSnapshotTime uint64
	GetAssetAmount  uint64
	SnapBalance     uint64
	IssueAsset      uint64
	DestroyAsset    uint64
	AddAsset        uint64
	GetAccountID    uint64
	GetDelegate     uint64
	CryptoCalc      uint64
	CryptoByte      uint64
	DeductGas       uint64
	WithdrawFee     uint64
}

// Variables containing gas prices for different phases.
var (
	// GasTable contain the gas re-prices
	GasTableInstanse = GasTable{
		ExtcodeSize: 700,
		ExtcodeCopy: 700,
		Balance:     400,
		SLoad:       200,
		Calls:       700,
		Suicide:     5000,
		ExpByte:     50,

		CreateBySuicide: 25000,
		SetOwner:        200,
		WithdrawFee:     700,
		GetAccountTime:  200,
		GetSnapshotTime: 200,
		GetAssetAmount:  200,
		SnapBalance:     200,
		IssueAsset:      200,
		DestroyAsset:    200,
		AddAsset:        200,
		GetAccountID:    200,
		GetDelegate:     200,
		CryptoCalc:      20000,
		CryptoByte:      1000,
		DeductGas:       200,
	}
)

// CalcGasLimit computes the gas limit of the next block after parent.
// This is miner strategy, not consensus protocol.
func CalcGasLimit(parent *types.Block) uint64 {
	// contrib = (parentGasUsed * 3 / 2) / 1024
	contrib := (parent.GasUsed() + parent.GasUsed()/2) / GasLimitBoundDivisor

	// decay = parentGasLimit / 1024 -1
	decay := parent.GasLimit()/GasLimitBoundDivisor - 1

	/*
		strategy: gasLimit of block-to-mine is set based on parent's
		gasUsed value.  if parentGasUsed > parentGasLimit * (2/3) then we
		increase it, otherwise lower it (or leave it unchanged if it's right
		at that usage) the amount increased/decreased depends on how far away
		from parentGasLimit * (2/3) parentGasUsed is.
	*/
	limit := parent.GasLimit() - decay + contrib
	if limit < MinGasLimit {
		limit = MinGasLimit
	}
	// however, if we're now below the target (GenesisGasLimit) we increase the
	// limit as much as we can (parentGasLimit / 1024 -1)
	if limit < GenesisGasLimit {
		limit = parent.GasLimit() + decay
		if limit > GenesisGasLimit {
			limit = GenesisGasLimit
		}
	}
	return limit
}
