// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// IPM plugin manager interface.
type IPM interface {
	IAccount
	IAsset
	IConsensus
	IContract
	IFee
	ISinger
	ExecTx(arg interface{}) ([]byte, error)
}

// IAccount account manager interface.
type IAccount interface {
	GetNonce(accountAddress common.Address) (uint64, error)
	SetNonce(accountAddress common.Address, nonce uint64) error
	GetAccount(accountAddress common.Address) (*Account, error)
	DeleteAccount(accountAddress common.Address) error
	AccountHaveCode(accountAddress common.Address) (bool, error)
	GetCode(accountAddress common.Address) ([]byte, error)
	SetCode(accountAddress common.Address, code []byte) (bool, error)
	GetBalanceByID(accountAddress common.Address, assetID uint64) (*big.Int, error)
	CreateAccount(pubKey common.PubKey, description string) ([]byte, error)
	IssueAsset(accountAddress common.Address, assetName string, symbol string, amount *big.Int, dec uint64, founder common.Address, owner common.Address, limit *big.Int, description string, asm IAsset) ([]byte, error)
	CanTransfer(accountAddress common.Address, assetID uint64, value *big.Int) (bool, error)
	TransferAsset(fromAccount, toAccount common.Address, assetID uint64, value *big.Int, asm IAsset, fromAccountExtra ...common.Address) error
	RecoverTx(signer types.Signer, tx *types.Transaction) error
}

type IAsset interface {
	IncStats(assetID uint64) error
	CheckIssueAssetInfo(account common.Address, assetInfo *IssueAsset) error
	IssueAssetForAccount(assetName string, symbol string, amount *big.Int, dec uint64, founder common.Address, owner common.Address, limit *big.Int, description string) (uint64, error)
}

type IConsensus interface {
}

type IContract interface {
}

type IFee interface {
}

type ISinger interface {
}
