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
package asset

import (
	"math/big"
	"regexp"

	"github.com/fractalplatform/fractal/common"
)

type AssetObject struct {
	AssetId   uint64      `json:"assetid,omitempty"`
	AssetName string      `json:"assetname,omitempty"`
	Symbol    string      `json:"symbol,omitempty"`
	Amount    *big.Int    `json:"amount,omitempty"`
	Decimals  uint64      `json:"decimals,omitempty"`
	Owner     common.Name `json:"owner,omitempty"`
}

func NewAssetObject(assetName string, symbol string, amount *big.Int, dec uint64, owner common.Name) (*AssetObject, error) {
	if assetName == "" || symbol == "" || owner == "" {
		return nil, ErrNewAssetObject
	}

	if amount.Cmp(big.NewInt(0)) < 0 {
		return nil, ErrNewAssetObject
	}

	reg := regexp.MustCompile("^[a-z0-9]{2,16}$")
	if reg.MatchString(assetName) == false {
		return nil, ErrNewAssetObject
	}

	if reg.MatchString(symbol) == false {
		return nil, ErrNewAssetObject
	}

	ao := AssetObject{
		AssetId:   0,
		AssetName: assetName,
		Symbol:    symbol,
		Amount:    amount,
		Decimals:  dec,
		Owner:     owner,
	}
	return &ao, nil
}

func (ao *AssetObject) GetAssetId() uint64 {
	return ao.AssetId
}

func (ao *AssetObject) SetAssetId(assetId uint64) {
	ao.AssetId = assetId
}

func (ao *AssetObject) GetSymbol() string {
	return ao.Symbol
}
func (ao *AssetObject) SetSymbol(sym string) {
	ao.Symbol = sym
}

func (ao *AssetObject) GetDecimals() uint64 {
	return ao.Decimals
}

func (ao *AssetObject) SetDecimals(dec uint64) {
	ao.Decimals = dec
}

func (ao *AssetObject) GetAssetName() string {
	return ao.AssetName
}
func (ao *AssetObject) SetAssetName(assetName string) {
	ao.AssetName = assetName
}

func (ao *AssetObject) GetAssetAmount() *big.Int {
	return ao.Amount
}

func (ao *AssetObject) SetAssetAmount(amount *big.Int) {
	ao.Amount = amount
}

func (ao *AssetObject) GetAssetOwner() common.Name {
	return ao.Owner
}

func (ao *AssetObject) SetAssetOwner(owner common.Name) {
	ao.Owner = owner
}
