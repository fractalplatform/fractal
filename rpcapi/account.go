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

type RPCAccount struct {
	AcctName              common.Name                    `json:"accountName"`
	Founder               common.Name                    `json:"founder"`
	AccountID             uint64                         `json:"accountID"`
	Number                uint64                         `json:"number"`
	Nonce                 uint64                         `json:"nonce"`
	Code                  hexutil.Bytes                  `json:"code"`
	CodeHash              common.Hash                    `json:"codeHash"`
	CodeSize              uint64                         `json:"codeSize"`
	Threshold             uint64                         `json:"threshold"`
	UpdateAuthorThreshold uint64                         `json:"updateAuthorThreshold"`
	AuthorVersion         common.Hash                    `json:"authorVersion"`
	Balances              []*accountmanager.AssetBalance `json:"balances"`
	Authors               []*common.Author               `json:"authors"`
	Suicide               bool                           `json:"suicide"`
	Destroy               bool                           `json:"destroy"`
	Description           string                         `json:"description"`
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

type AccountAPI struct {
	b Backend
}

func NewAccountAPI(b Backend) *AccountAPI {
	return &AccountAPI{b}
}

//AccountIsExist
func (api *AccountAPI) AccountIsExist(acctName common.Name) (bool, error) {
	acct, err := api.b.GetAccountManager()
	if err != nil {
		return false, err
	}
	return acct.AccountIsExist(acctName)
}

func (api *AccountAPI) GetAccountExByID(accountID uint64) (*RPCAccount, error) {
	am, err := api.b.GetAccountManager()
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
func (api *AccountAPI) GetAccountByID(accountID uint64) (*RPCAccount, error) {

	am, err := api.b.GetAccountManager()
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

func (api *AccountAPI) GetAccountExByName(accountName common.Name) (*RPCAccount, error) {
	am, err := api.b.GetAccountManager()
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
func (api *AccountAPI) GetAccountByName(accountName common.Name) (*RPCAccount, error) {
	am, err := api.b.GetAccountManager()
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
func (api *AccountAPI) GetAccountBalanceByID(accountName common.Name, assetID uint64, typeID uint64) (*big.Int, error) {
	am, err := api.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return am.GetAccountBalanceByID(accountName, assetID, typeID)
}

//GetCode
func (api *AccountAPI) GetCode(accountName common.Name) (hexutil.Bytes, error) {
	acct, err := api.b.GetAccountManager()
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
func (api *AccountAPI) GetNonce(accountName common.Name) (uint64, error) {
	acct, err := api.b.GetAccountManager()
	if err != nil {
		return 0, err
	}
	return acct.GetNonce(accountName)
}

//GetAssetInfoByName
func (api *AccountAPI) GetAssetInfoByName(ctx context.Context, assetName string) (*asset.AssetObject, error) {
	acct, err := api.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return acct.GetAssetInfoByName(assetName)
}

//GetAssetInfoByID
func (api *AccountAPI) GetAssetInfoByID(assetID uint64) (*asset.AssetObject, error) {
	acct, err := api.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return acct.GetAssetInfoByID(assetID)
}

//GetAssetAmountByTime
func (api *AccountAPI) GetAssetAmountByTime(assetID uint64, time uint64) (*big.Int, error) {
	am, err := api.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return am.GetAssetAmountByTime(assetID, time)
}

//GetAccountBalanceByTime
func (api *AccountAPI) GetAccountBalanceByTime(accountName common.Name, assetID uint64, typeID uint64, time uint64) (*big.Int, error) {
	am, err := api.b.GetAccountManager()
	if err != nil {
		return nil, err
	}
	return am.GetBalanceByTime(accountName, assetID, typeID, time)
}

//GetSnapshotLast  get last snapshot time
func (api *AccountAPI) GetSnapshotLast() (uint64, error) {
	am, err := api.b.GetAccountManager()
	if err != nil {
		return 0, err
	}

	return am.GetSnapshotTime(0, 0)
}

//GetSnapshotTime  m: 1  preview time   2 next time
func (api *AccountAPI) GetSnapshotTime(ctx context.Context, m uint64, time uint64) (uint64, error) {
	am, err := api.b.GetAccountManager()
	if err != nil {
		return 0, err
	}
	return am.GetSnapshotTime(m, time)
}
