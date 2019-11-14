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
	pm, err := api.b.GetPM()
	if err != nil {
		return false, err
	}
	return pm.AccountIsExist(accountName)
}

//GetAccountByName
func (api *AccountAPI) GetAccountByName(accountName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return nil, err
	}
	return pm.GetAccountByName(accountName)
}

//GetAccountBalanceByID
func (api *AccountAPI) GetAccountBalanceByID(accountName string, assetID uint64) (*big.Int, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return big.NewInt(0), err
	}
	return pm.GetBalance(accountName, assetID)
}

//GetCode
func (api *AccountAPI) GetCode(accountName string) ([]byte, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return nil, err
	}
	return pm.GetCode(accountName)
}

//GetNonce
func (api *AccountAPI) GetNonce(accountName string) (uint64, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetNonce(accountName)
}

//GetAssetInfoByName
func (api *AccountAPI) GetAssetInfoByName(assetName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return nil, err
	}

	return pm.GetAssetInfoByName(assetName)
}

//GetAssetInfoByID
func (api *AccountAPI) GetAssetInfoByID(assetID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return nil, err
	}
	return pm.GetAssetInfoByID(assetID)
}
