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

// GasTable organizes gas prices for different phases.
type GasTable struct {
	ActionGas             uint64
	ActionGasCallContract uint64
	ActionGasCreation     uint64
	ActionGasIssueAsset   uint64
	SignGas               uint64
	TxDataNonZeroGas      uint64
	TxDataZeroGas         uint64

	ExtcodeSize          uint64
	ExtcodeCopy          uint64
	Balance              uint64
	SLoad                uint64
	Calls                uint64
	ExpByte              uint64
	CallValueTransferGas uint64
	QuadCoeffDiv         uint64
	SstoreSetGas         uint64
	LogDataGas           uint64
	CallStipend          uint64

	SetOwner        uint64
	GetAccountTime  uint64
	GetSnapshotTime uint64
	GetAssetInfo    uint64
	SnapBalance     uint64
	IssueAsset      uint64
	DestroyAsset    uint64
	AddAsset        uint64
	GetAccountID    uint64
	GetAssetID      uint64
	CryptoCalc      uint64
	CryptoByte      uint64
	DeductGas       uint64
	WithdrawFee     uint64
	GetEpoch        uint64
	GetCandidateNum uint64
	GetCandidate    uint64
	GetVoterStake   uint64

	Sha3Gas        uint64
	Sha3WordGas    uint64
	SstoreResetGas uint64
	JumpdestGas    uint64
	CreateDataGas  uint64
	LogGas         uint64
	CopyGas        uint64
	LogTopicGas    uint64
	CreateGas      uint64
	MemoryGas      uint64
}

// Variables containing gas prices for different phases.
var (
	// GasTable contain the gas re-prices
	GasTableInstance = GasTable{
		ActionGas:             100000,
		ActionGasCallContract: 200000,
		ActionGasCreation:     500000,
		ActionGasIssueAsset:   10000000,
		SignGas:               50000,

		ExtcodeSize: 700,
		ExtcodeCopy: 700,
		Balance:     400,
		SLoad:       200,
		Calls:       700,
		ExpByte:     50,

		SetOwner:        200,
		WithdrawFee:     700,
		GetAccountTime:  200,
		GetSnapshotTime: 200,
		GetAssetInfo:    200,
		SnapBalance:     200,
		IssueAsset:      10000000,
		DestroyAsset:    200,
		AddAsset:        200,
		GetAccountID:    200,
		GetAssetID:      200,
		CryptoCalc:      20000,
		CryptoByte:      1000,
		DeductGas:       200,
		GetEpoch:        200,
		GetCandidateNum: 200,
		GetCandidate:    200,
		GetVoterStake:   200,

		TxDataNonZeroGas: 68,
		TxDataZeroGas:    4,

		CallValueTransferGas: 9000,
		QuadCoeffDiv:         512,
		SstoreSetGas:         20000,
		LogDataGas:           8,
		CallStipend:          0,

		Sha3Gas:        30,
		Sha3WordGas:    6,
		SstoreResetGas: 5000,
		JumpdestGas:    1,
		CreateDataGas:  200,
		LogGas:         375,
		CopyGas:        3,
		LogTopicGas:    375,
		CreateGas:      32000,
		MemoryGas:      3,
	}
)
