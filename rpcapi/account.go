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

import "math/big"

type AccountAPI struct {
	b Backend
}

func NewAccountAPI(b Backend) *AccountAPI {
	return &AccountAPI{b}
}

//AccountIsExist
func (api *AccountAPI) AccountIsExist(accountName string) (bool, error) {
	return api.b.GetPM().AccountIsExist(accountName)
}

//GetAccountByName
func (api *AccountAPI) GetAccountByName(accountName string) (interface{}, error) {
	return api.b.GetPM().GetAccountByName(accountName)
}

//GetAccountBalanceByID
func (api *AccountAPI) GetAccountBalanceByID(accountName string, assetID uint64) (*big.Int, error) {
	return api.b.GetPM().GetBalance(accountName, assetID)
}

//GetCode
func (api *AccountAPI) GetCode(accountName string) ([]byte, error) {
	return api.b.GetPM().GetCode(accountName)
}

//GetNonce
func (api *AccountAPI) GetNonce(accountName string) (uint64, error) {
	return api.b.GetPM().GetNonce(accountName)
}

//GetAssetInfoByName
func (api *AccountAPI) GetAssetInfoByName(assetName string) (interface{}, error) {
	return api.b.GetPM().GetAssetInfoByName(assetName)
}

//GetAssetInfoByID
func (api *AccountAPI) GetAssetInfoByID(assetID uint64) (interface{}, error) {
	return api.b.GetPM().GetAssetInfoByID(assetID)
}
