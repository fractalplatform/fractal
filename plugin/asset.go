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
	"errors"
	"math/big"
	"regexp"
	"strconv"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	assetRegExp        = regexp.MustCompile(`^([a-z][a-z0-9]{1,31})$`)
	assetNameMaxLength = uint64(32)
	assetManagerName   = "assetAccount"
	assetObjectPrefix  = "assetDefinitionObject"
)

const SystemAssetID uint64 = 0

var UINT256_MAX *big.Int = big.NewInt(0).Sub((big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil)), big.NewInt(1))

type AssetManager struct {
	sdb *state.StateDB
}

type Asset struct {
	AssetID     uint64   `json:"assetId"`
	AssetName   string   `json:"assetName"`
	Symbol      string   `json:"symbol"`
	Amount      *big.Int `json:"amount"`
	Decimals    uint64   `json:"decimals"`
	Founder     string   `json:"founder"`
	Owner       string   `json:"owner"`
	AddIssue    *big.Int `json:"addIssue"`
	UpperLimit  *big.Int `json:"upperLimit"`
	Description string   `json:"description"`
}

// NewASM New a AssetManager
func NewASM(sdb *state.StateDB) (*AssetManager, error) {
	if sdb == nil {
		return nil, ErrNewAssetManagerErr
	}

	asset := AssetManager{
		sdb: sdb,
	}
	return &asset, nil
}

func (asm *AssetManager) AccountName() string {
	return "fractalasset"
}

func (asm *AssetManager) CallTx(tx *envelope.PluginTx, ctx *Context, pm IPM) ([]byte, error) {
	switch tx.PayloadType() {
	case IssueAsset:
		param := &IssueAssetAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return asm.IssueAsset(tx.Sender(), param.AssetName, param.Symbol, param.Amount, param.Decimals, param.Founder, param.Owner, param.UpperLimit, param.Description, pm)
	case IncreaseAsset:
		param := &IncreaseAssetAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return asm.IncreaseAsset(tx.Sender(), param.To, param.AssetID, param.Amount, pm)
	}
	return nil, ErrWrongTransaction
}

// IssueAsset Issue system asset
func (asm *AssetManager) IssueAsset(accountName string, assetName string, symbol string, amount *big.Int,
	decimals uint64, founder string, owner string, limit *big.Int, description string, am IAccount) ([]byte, error) {

	_, err := asm.getAssetObjectByID(SystemAssetID)
	if err == nil { // system asset has issued
		return nil, ErrIssueAsset
	}

	err = asm.checkIssueAssetParam(accountName, assetName, symbol, amount, decimals, owner, limit, description, am)
	if err != nil {
		return nil, err
	}
	// check owner and founder
	_, err = am.getAccount(owner)
	if err != nil {
		return nil, err
	}

	if len(founder) > 0 {
		_, err = am.getAccount(founder)
		if err != nil {
			return nil, err
		}
	} else {
		founder = owner
	}

	ao := Asset{
		AssetID:     SystemAssetID,
		AssetName:   assetName,
		Symbol:      symbol,
		Amount:      amount,
		Decimals:    decimals,
		Founder:     founder,
		Owner:       owner,
		AddIssue:    amount,
		UpperLimit:  limit,
		Description: description,
	}

	err = asm.setAsset(&ao)
	if err != nil {
		return nil, err
	}

	if err = am.addBalanceByID(owner, SystemAssetID, amount); err != nil {
		return nil, err
	}

	return nil, nil
}

// IncreaseAsset increase system asset
func (asm *AssetManager) IncreaseAsset(from, to string, assetID uint64, amount *big.Int, am IAccount) ([]byte, error) {
	if from == "" || to == "" {
		return nil, ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrAmountValueInvalid
	}

	if assetID != SystemAssetID {
		return nil, ErrAssetNotExist
	}

	assetObj, err := asm.getAssetObjectByID(assetID)
	if err != nil {
		return nil, err
	}

	// check asset owner
	if assetObj.Owner != from {
		return nil, ErrOwnerMismatch
	}

	addissue := new(big.Int).Add(assetObj.AddIssue, amount)
	if assetObj.UpperLimit.Cmp(big.NewInt(0)) > 0 && addissue.Cmp(assetObj.UpperLimit) > 0 {
		return nil, ErrUpperLimit
	}
	assetObj.AddIssue = addissue
	assetObj.Amount = new(big.Int).Add(assetObj.Amount, amount)
	if assetObj.Amount.Cmp(UINT256_MAX) > 0 {
		return nil, ErrAssetTotalExceedLimitErr
	}

	if err = asm.setAsset(assetObj); err != nil {
		return nil, err
	}

	if err = am.addBalanceByID(to, assetID, amount); err != nil {
		return nil, err
	}

	return nil, nil
}

// DestroyAsset destroy system asset
func (asm *AssetManager) DestroyAsset(accountName string, assetID uint64, amount *big.Int, am IAccount) ([]byte, error) {
	if accountName == "" {
		return nil, ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrAmountValueInvalid
	}

	if assetID != SystemAssetID {
		return nil, ErrAssetNotExist
	}

	assetObj, err := asm.getAssetObjectByID(assetID)
	if err != nil {
		return nil, err
	}

	total := new(big.Int).Sub(assetObj.Amount, amount)
	if total.Cmp(big.NewInt(0)) < 0 {
		return nil, ErrDestroyLimit
	}
	assetObj.Amount = total

	if err = asm.setAsset(assetObj); err != nil {
		return nil, err
	}

	if err := am.subBalanceByID(accountName, assetID, amount); err != nil {
		return nil, err
	}

	return nil, nil
}

// GetAssetID Get asset id
func (asm *AssetManager) GetAssetID(assetName string) (uint64, error) {
	if assetName == "" {
		return 0, ErrAssetNotExist
	}

	obj, err := asm.getAssetObjectByID(SystemAssetID)
	if err != nil {
		return 0, err
	}

	if obj.AssetName != assetName {
		return 0, ErrAssetNotExist
	}

	return obj.AssetID, nil
}

// GetAssetName Get asset name
func (asm *AssetManager) GetAssetName(assetID uint64) (string, error) {
	if assetID != SystemAssetID {
		return "", ErrAssetNotExist
	}

	obj, err := asm.getAssetObjectByID(SystemAssetID)
	if err != nil {
		return "", err
	}
	return obj.AssetName, nil
}

func (asm *AssetManager) GetAssetInfoByName(assetName string) (*Asset, error) {
	if assetName == "" {
		return nil, ErrAssetNotExist
	}

	obj, err := asm.getAssetObjectByID(SystemAssetID)
	if err != nil {
		return nil, ErrAssetNotExist
	}

	if obj.AssetName != assetName {
		return nil, ErrAssetNotExist
	}
	return obj, nil
}

func (asm *AssetManager) GetAssetInfoByID(assetID uint64) (*Asset, error) {
	if assetID != SystemAssetID {
		return nil, ErrAssetNotExist
	}

	obj, err := asm.getAssetObjectByID(SystemAssetID)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (asm *AssetManager) checkIssueAssetParam(accountName string, assetName string, symbol string, amount *big.Int,
	decimals uint64, owner string, limit *big.Int, description string, am IAccount) error {

	if accountName == "" || assetName == "" || symbol == "" || owner == "" {
		return ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) < 0 || limit.Cmp(big.NewInt(0)) < 0 || amount.Cmp(UINT256_MAX) > 0 || limit.Cmp(UINT256_MAX) > 0 {
		return ErrAmountValueInvalid
	}

	if limit.Cmp(big.NewInt(0)) > 0 {
		if amount.Cmp(limit) > 0 {
			return ErrAmountValueInvalid
		}
	}

	if uint64(len(description)) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}

	err := asm.checkAssetName(assetName)
	if err != nil {
		return err
	}
	err = asm.checkAssetName(symbol)
	if err != nil {
		return err
	}

	if _, err = am.getAccount(assetName); err == nil {
		return ErrAssetNameEqualAccountName
	}

	return nil
}

func (asm *AssetManager) checkAssetName(assetName string) error {
	if uint64(len(assetName)) > assetNameMaxLength {
		return ErrAssetNameLengthErr
	}

	if assetRegExp.MatchString(assetName) != true {
		return ErrAssetNameinvalid
	}
	return nil
}

func (asm *AssetManager) setAsset(asset *Asset) error {
	if asset == nil {
		return ErrAssetObjectEmpty
	}
	assetID := asset.AssetID

	b, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return err
	}
	asm.sdb.Put(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetID, 10), b)
	return nil
}

func (asm *AssetManager) getAssetObjectByID(ID uint64) (*Asset, error) {
	b, err := asm.sdb.Get(assetManagerName, assetObjectPrefix+strconv.FormatUint(ID, 10))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrAssetNotExist
	}
	var asset Asset
	if err := rlp.DecodeBytes(b, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (asm *AssetManager) checkIssueAsset(accountName string, assetName string, symbol string, amount *big.Int,
	decimals uint64, founder string, owner string, limit *big.Int, description string, am IAccount) error {

	if accountName == "" || assetName == "" || symbol == "" || owner == "" {
		return ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) < 0 || limit.Cmp(big.NewInt(0)) < 0 || amount.Cmp(UINT256_MAX) > 0 || limit.Cmp(UINT256_MAX) > 0 {
		return ErrAmountValueInvalid
	}

	if limit.Cmp(big.NewInt(0)) > 0 {
		if amount.Cmp(limit) > 0 {
			return ErrAmountValueInvalid
		}
	}

	if uint64(len(description)) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}

	err := asm.checkAssetName(assetName)
	if err != nil {
		return err
	}
	err = asm.checkAssetName(symbol)
	if err != nil {
		return err
	}
	return nil
}

func (asm *AssetManager) checkIncreaseAsset(from, to string, assetID uint64, amount *big.Int, am IAccount) error {
	if from == "" || to == "" {
		return ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return ErrAmountValueInvalid
	}

	if assetID != SystemAssetID {
		return ErrAssetNotExist
	}
	return nil
}

func (asm *AssetManager) Sol_IssueAsset(context *ContextSol, name string, symbol string, amount *big.Int, decimals uint64, founder string, owner string, limit *big.Int, desc string) error {
	_, err := asm.IssueAsset(context.tx.Sender(), name, symbol, amount, decimals, founder, owner, limit, desc, context.pm)
	return err
}

func (asm *AssetManager) Sol_IncreaseAsset(context *ContextSol, to common.Address, assetID uint64, amount *big.Int) error {
	_, err := asm.IncreaseAsset(context.tx.Sender(), to.AccountName(), assetID, amount, context.pm)
	return err
}

var (
	ErrNewAssetManagerErr        = errors.New("new AssetManager error")
	ErrAssetIsExist              = errors.New("asset is exist")
	ErrAssetNotExist             = errors.New("asset not exist")
	ErrDescriptionTooLong        = errors.New("description exceed max length")
	ErrAssetObjectEmpty          = errors.New("asset object is empty")
	ErrOwnerMismatch             = errors.New("asset owner mismatch")
	ErrParamIsNil                = errors.New("param is nil")
	ErrUpperLimit                = errors.New("asset amount over the issuance limit")
	ErrDestroyLimit              = errors.New("asset destroy exceeding the lower limit")
	ErrAssetNameEqualAccountName = errors.New("asset name equal account name")
	ErrIssueAsset                = errors.New("system asset has issued")
	ErrAssetNameinvalid          = errors.New("asset name invalid")
	ErrAssetNameLengthErr        = errors.New("asset name length err")
	ErrAssetTotalExceedLimitErr  = errors.New("asset total exceed uint256 err")
)
