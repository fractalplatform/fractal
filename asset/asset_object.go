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

	"github.com/fractalplatform/fractal/common"
)

type AssetObject struct {
	AssetId     uint64      `json:"assetId"`
	Number      uint64      `json:"number"`
	AssetName   string      `json:"assetName"`
	Symbol      string      `json:"symbol"`
	Amount      *big.Int    `json:"amount"`
	Decimals    uint64      `json:"decimals"`
	Founder     common.Name `json:"founder"`
	Owner       common.Name `json:"owner"`
	AddIssue    *big.Int    `json:"addIssue"`
	UpperLimit  *big.Int    `json:"upperLimit"`
	Contract    common.Name `json:"contract"`
	Description string      `json:"description"`
}

func NewAssetObject(assetName string, number uint64, symbol string, amount *big.Int, dec uint64, founder common.Name, owner common.Name, limit *big.Int, contract common.Name, description string) (*AssetObject, error) {

	if assetName == "" || symbol == "" || owner == "" {
		return nil, ErrNewAssetObject
	}

	if amount.Cmp(big.NewInt(0)) < 0 || limit.Cmp(big.NewInt(0)) < 0 {
		return nil, ErrNewAssetObject
	}

	if limit.Cmp(big.NewInt(0)) > 0 {
		if amount.Cmp(limit) > 0 {
			return nil, ErrNewAssetObject
		}
	}
	if !common.StrToName(assetName).IsValid(assetRegExp) {
		return nil, ErrNewAssetObject
	}
	if !common.StrToName(symbol).IsValid(assetRegExp) {
		return nil, ErrNewAssetObject
	}
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDetailTooLong
	}

	ao := AssetObject{
		AssetId:     0,
		Number:      number,
		AssetName:   assetName,
		Symbol:      symbol,
		Amount:      amount,
		Decimals:    dec,
		Founder:     founder,
		Owner:       owner,
		AddIssue:    amount,
		UpperLimit:  limit,
		Contract:    contract,
		Description: description,
	}
	return &ao, nil
}

func (ao *AssetObject) GetAssetId() uint64 {
	return ao.AssetId
}

func (ao *AssetObject) SetAssetId(assetId uint64) {
	ao.AssetId = assetId
}

func (ao *AssetObject) GetAssetNumber() uint64 {
	return ao.Number
}

func (ao *AssetObject) SetAssetNumber(number uint64) {
	ao.Number = number
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

func (ao *AssetObject) SetAssetAddIssue(amount *big.Int) {
	ao.AddIssue = amount
}

func (ao *AssetObject) GetAssetAddIssue() *big.Int {
	return ao.AddIssue
}

func (ao *AssetObject) GetUpperLimit() *big.Int {
	return ao.UpperLimit
}

func (ao *AssetObject) GetContract() common.Name {
	return ao.Contract
}

func (ao *AssetObject) SetAssetAmount(amount *big.Int) {
	ao.Amount = amount
}

func (ao *AssetObject) GetAssetFounder() common.Name {
	return ao.Founder
}

func (ao *AssetObject) SetAssetFounder(f common.Name) {
	ao.Founder = f
}

func (ao *AssetObject) GetAssetOwner() common.Name {
	return ao.Owner
}

func (ao *AssetObject) SetAssetOwner(owner common.Name) {
	ao.Owner = owner
}

func (ao *AssetObject) GetAssetContract() common.Name {
	return ao.Contract
}

func (ao *AssetObject) SetAssetContract(contract common.Name) {
	ao.Contract = contract
}

func (ao *AssetObject) GetAssetDescription() string {
	return ao.Description
}

func (ao *AssetObject) SetAssetDescription(description string) {
	ao.Description = description
}
