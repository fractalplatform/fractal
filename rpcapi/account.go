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

package rpcapi

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
)

type AccountAPI struct {
	b Backend
}

func NewAccountAPI(b Backend) *AccountAPI {
	return &AccountAPI{b}
}

//AccountIsExist
func (aapi *AccountAPI) AccountIsExist(acctName common.Name) (bool, error) {
	acct, err := aapi.b.GetAccountManager()
	if err != nil {
		return false, err
	}
	return acct.AccountIsExist(acctName)
}

//GetAccountByID
func (aapi *AccountAPI) GetAccountByID(accountID uint64) (*accountmanager.Account, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}

	return am.GetAccountById(accountID)
}

//GetAccountByName
func (aapi *AccountAPI) GetAccountByName(accountName common.Name) (*accountmanager.Account, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}

	return am.GetAccountByName(accountName)
}

//GetAccountBalanceByID
func (aapi *AccountAPI) GetAccountBalanceByID(accountName common.Name, assetID uint64, typeID uint64) (*big.Int, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return am.GetAccountBalanceByID(accountName, assetID, typeID)
}

//GetCode
func (aapi *AccountAPI) GetCode(accountName common.Name) (hexutil.Bytes, error) {
	acct, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}

	result, err := acct.GetCode(accountName)
	if err != nil {
		return nil, err
	}
	return (hexutil.Bytes)(result), nil

}

//GetNonce
func (aapi *AccountAPI) GetNonce(accountName common.Name) (uint64, error) {
	acct, err := aapi.b.GetAccountManager()
	if err != nil {
		return 0, err
	}

	return acct.GetNonce(accountName)

}

//GetAssetInfoByName
func (aapi *AccountAPI) GetAssetInfoByName(ctx context.Context, assetName string) (*asset.AssetObject, error) {
	acct, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return acct.GetAssetInfoByName(assetName)
}

//GetAssetInfoByID
func (aapi *AccountAPI) GetAssetInfoByID(assetID uint64) (*asset.AssetObject, error) {
	acct, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return acct.GetAssetInfoByID(assetID)
}

//GetAssetAmountByTime
func (aapi *AccountAPI) GetAssetAmountByTime(assetID uint64, time uint64) (*big.Int, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return am.GetAssetAmountByTime(assetID, time)
}

//GetAccountBalanceByTime
func (aapi *AccountAPI) GetAccountBalanceByTime(accountName common.Name, assetID uint64, typeID uint64, time uint64) (*big.Int, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return am.GetBalanceByTime(accountName, assetID, typeID, time)
}

//GetSnapshotLast  get last snapshot time
func (aapi *AccountAPI) GetSnapshotLast() (uint64, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return 0, err
	}

	return am.GetSnapshotTime(0, 0)
}

//getSnapshottime  m: 1  preview time   2 next time
func (aapi *AccountAPI) GetSnapshotTime(ctx context.Context, m uint64, time uint64) (uint64, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return 0, err
	}
	return am.GetSnapshotTime(m, time)
}
