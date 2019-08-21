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
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
)

type RPCAccount struct {
	//LastTime *big.Int
	AcctName              common.Name   `json:"accountName"`
	Founder               common.Name   `json:"founder"`
	AccountID             uint64        `json:"accountID"`
	Number                uint64        `json:"number"`
	Nonce                 uint64        `json:"nonce"`
	Code                  hexutil.Bytes `json:"code"`
	CodeHash              common.Hash   `json:"codeHash"`
	CodeSize              uint64        `json:"codeSize"`
	Threshold             uint64        `json:"threshold"`
	UpdateAuthorThreshold uint64        `json:"updateAuthorThreshold"`
	AuthorVersion         common.Hash   `json:"authorVersion"`
	//sort by asset id asc
	Balances []*accountmanager.AssetBalance `json:"balances"`
	//realated account, pubkey and address
	Authors []*common.Author `json:"authors"`
	//code Suicide
	Suicide bool `json:"suicide"`
	//account destroy
	Destroy     bool   `json:"destroy"`
	Description string `json:"description"`
}

func NewRPCAccount(account *accountmanager.Account) *RPCAccount {
	acctObject := RPCAccount{
		AcctName:              account.AcctName,
		Founder:               account.Founder,
		AccountID:             account.AccountID,
		Number:                account.Number,
		Nonce:                 account.Nonce,
		Code:                  hexutil.Bytes(account.Code),
		CodeHash:              account.CodeHash,
		CodeSize:              account.CodeSize,
		Threshold:             account.Threshold,
		UpdateAuthorThreshold: account.UpdateAuthorThreshold,
		AuthorVersion:         account.AuthorVersion,
		Balances:              account.Balances,
		Authors:               account.Authors,
		Suicide:               account.Suicide,
		Destroy:               account.Destroy,
		Description:           account.Description,
	}
	return &acctObject
}

func (a *RPCAccount) GetBalanceByID(assetID uint64) (*big.Int, error) {
	p, find := a.binarySearch(assetID)
	if find {
		return a.Balances[p].Balance, nil
	}
	return big.NewInt(0), errors.New("account asset not exist")
}

// BinarySearch binary search
func (a *RPCAccount) binarySearch(assetID uint64) (int64, bool) {

	low := int64(0)
	high := int64(len(a.Balances)) - 1
	for low <= high {
		mid := (low + high) / 2
		if a.Balances[mid].AssetID < assetID {
			low = mid + 1
		} else if a.Balances[mid].AssetID > assetID {
			high = mid - 1
		} else if a.Balances[mid].AssetID == assetID {
			return mid, true
		}
	}
	if high < 0 {
		high = 0
	}
	return high, false
}

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

func (aapi *AccountAPI) GetAccountExByID(accountID uint64) (*RPCAccount, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}

	accountObj, err := am.GetAccountById(accountID)
	if err != nil {
		return nil, err
	}

	var rpcAccountObj *RPCAccount
	if accountObj != nil {
		rpcAccountObj = NewRPCAccount(accountObj)
	}

	return rpcAccountObj, nil
}

//GetAccountByID
func (aapi *AccountAPI) GetAccountByID(accountID uint64) (*RPCAccount, error) {

	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}

	accountObj, err := am.GetAccountById(accountID)
	if err != nil {
		return nil, err
	}

	var rpcAccountObj *RPCAccount
	if accountObj != nil {
		rpcAccountObj = NewRPCAccount(accountObj)
		balances := make([]*accountmanager.AssetBalance, 0, len(rpcAccountObj.Balances))
		zero := big.NewInt(0)
		for _, balance := range rpcAccountObj.Balances {
			if balance.Balance.Cmp(zero) > 0 {
				balances = append(balances, &accountmanager.AssetBalance{AssetID: balance.AssetID, Balance: balance.Balance})
			}
		}
		rpcAccountObj.Balances = balances
	}
	return rpcAccountObj, nil
}

func (aapi *AccountAPI) GetAccountExByName(accountName common.Name) (*RPCAccount, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}

	accountObj, err := am.GetAccountByName(accountName)
	if err != nil {
		return nil, err
	}

	var rpcAccountObj *RPCAccount
	if accountObj != nil {
		rpcAccountObj = NewRPCAccount(accountObj)
	}

	return rpcAccountObj, nil
}

//GetAccountByName
func (aapi *AccountAPI) GetAccountByName(accountName common.Name) (*RPCAccount, error) {
	am, err := aapi.b.GetAccountManager()
	if err != nil {
		return nil, err
	}

	accountObj, err := am.GetAccountByName(accountName)
	if err != nil {
		return nil, err
	}

	var rpcAccountObj *RPCAccount
	if accountObj != nil {
		rpcAccountObj = NewRPCAccount(accountObj)
		balances := make([]*accountmanager.AssetBalance, 0, len(rpcAccountObj.Balances))
		zero := big.NewInt(0)
		for _, balance := range rpcAccountObj.Balances {
			if balance.Balance.Cmp(zero) > 0 {
				balances = append(balances, &accountmanager.AssetBalance{AssetID: balance.AssetID, Balance: balance.Balance})
			}
		}
		rpcAccountObj.Balances = balances
	}

	return rpcAccountObj, nil
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
