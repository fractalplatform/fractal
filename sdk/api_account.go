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

package sdk

import (
	"math/big"

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/rpcapi"
)

// AccountIsExist account exist
func (api *API) AccountIsExist(name string) (bool, error) {
	isExist := false
	err := api.client.Call(&isExist, "account_accountIsExist", name)
	return isExist, err
}

// AccountInfo get account by name
func (api *API) AccountInfo(name string) (*rpcapi.RPCAccount, error) {
	account := &rpcapi.RPCAccount{}
	err := api.client.Call(account, "account_getAccountExByName", name)
	return account, err
}

// AccountInfoByID get account by id
func (api *API) AccountInfoByID(id uint64) (*rpcapi.RPCAccount, error) {
	account := &rpcapi.RPCAccount{}
	err := api.client.Call(account, "account_getAccountExByID", id)
	return account, err
}

// AccountCode get account code
func (api *API) AccountCode(name string) ([]byte, error) {
	code := []byte{}
	err := api.client.Call(code, "account_getCode", name)
	return code, err
}

// AccountNonce get account nonce
func (api *API) AccountNonce(name string) (uint64, error) {
	nonce := uint64(0)
	err := api.client.Call(&nonce, "account_getNonce", name)
	return nonce, err
}

// AssetInfoByName get asset info
func (api *API) AssetInfoByName(name string) (*asset.AssetObject, error) {
	assetInfo := &asset.AssetObject{}
	err := api.client.Call(assetInfo, "account_getAssetInfoByName", name)
	return assetInfo, err
}

// AssetInfoByID get asset info
func (api *API) AssetInfoByID(id uint64) (*asset.AssetObject, error) {
	assetInfo := &asset.AssetObject{}
	err := api.client.Call(assetInfo, "account_getAssetInfoByID", id)
	return assetInfo, err
}

// BalanceByAssetID get asset balance
func (api *API) BalanceByAssetID(name string, id uint64, typeID uint64) (*big.Int, error) {
	balance := big.NewInt(0)
	err := api.client.Call(balance, "account_getAccountBalanceByID", name, id, typeID)
	return balance, err
}
