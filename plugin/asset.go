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

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	assetRegExp        = regexp.MustCompile(`^([a-z][a-z0-9]{1,31})`)
	assetNameMaxLength = uint64(32)
	assetManagerName   = "assetAccount"
	assetCountPrefix   = "assetCount"
	assetNameIDPrefix  = "assetNameId"
	assetObjectPrefix  = "assetDefinitionObject"
)

const SystemAssetID uint64 = 0

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

type AddAsset struct {
	AssetID uint64
	Amount  *big.Int
	To      string
}

type IssueAsset struct {
	AssetName   string
	Symbol      string
	Amount      *big.Int
	Owner       string
	Founder     string
	Decimals    uint64
	UpperLimit  *big.Int
	Contract    string
	Description string
}

func NewASM(sdb *state.StateDB) (IAsset, error) {
	if sdb == nil {
		return nil, ErrNewAssetManagerErr
	}

	asset := AssetManager{
		sdb: sdb,
	}
	asset.initAssetCount()
	return &asset, nil
}

func (asm *AssetManager) IssueAsset(accountName string, assetName string, symbol string, amount *big.Int,
	decimals uint64, founder string, owner string, limit *big.Int, description string, am IAccount) ([]byte, error) {

	err := asm.checkIssueAssetParam(accountName, assetName, symbol, amount, decimals, owner, limit, description, am)
	if err != nil {
		return nil, err
	}
	// check owner and founder
	_, err = am.GetAccount(owner)
	if err != nil {
		return nil, err
	}

	if len(founder) > 0 {
		_, err = am.GetAccount(founder)
		if err != nil {
			return nil, err
		}
	} else {
		founder = owner
	}

	ao := Asset{
		AssetID:     0,
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

	assetID, err := asm.addNewAssetObject(&ao)
	if err != nil {
		return nil, err
	}
	if assetID != SystemAssetID {
		return nil, ErrIssueAsset
	}

	if err = am.AddBalanceByID(accountName, assetID, amount); err != nil {
		return nil, err
	}

	return nil, nil
}

func (asm *AssetManager) IncreaseAsset(from, to string, assetID uint64, amount *big.Int, am IAccount) ([]byte, error) {
	if from == "" || to == "" {
		return nil, ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrAmountValueInvalid
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

	total := new(big.Int).Add(assetObj.Amount, amount)
	if total.Cmp(assetObj.UpperLimit) > 0 {
		return nil, ErrUpperLimit
	}
	assetObj.Amount = total

	if err = asm.setAsset(assetObj); err != nil {
		return nil, err
	}

	if err = am.AddBalanceByID(to, assetID, amount); err != nil {
		return nil, err
	}

	return nil, nil
}

func (asm *AssetManager) DestroyAsset(accountName string, assetID uint64, amount *big.Int, am IAccount) ([]byte, error) {
	if accountName == "" {
		return nil, ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrAmountValueInvalid
	}

	if err := am.SubBalanceByID(accountName, assetID, amount); err != nil {
		return nil, err
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

	return nil, nil
}

func (asm *AssetManager) GetAssetID(assetName string) (uint64, error) {
	return asm.getAssetIDByName(assetName)
}

func (asm *AssetManager) GetAssetName(assetID uint64) (string, error) {
	asset, err := asm.getAssetByID(assetID)
	if err != nil {
		return "", err
	}
	return asset.AssetName, nil
}

func (asm *AssetManager) checkIssueAssetParam(accountName string, assetName string, symbol string, amount *big.Int,
	decimals uint64, owner string, limit *big.Int, description string, am IAccount) error {

	if accountName == "" || assetName == "" || symbol == "" || owner == "" {
		return ErrParamIsNil
	}

	if amount.Cmp(big.NewInt(0)) < 0 || limit.Cmp(big.NewInt(0)) < 0 {
		return ErrNewAssetObject
	}

	if limit.Cmp(big.NewInt(0)) > 0 {
		if amount.Cmp(limit) > 0 {
			return ErrNewAssetObject
		}
	}

	if uint64(len(description)) > MaxDescriptionLength {
		return ErrDetailTooLong
	}

	err := asm.checkAssetName(assetName)
	if err != nil {
		return err
	}
	err = asm.checkAssetName(symbol)
	if err != nil {
		return err
	}

	if _, err = am.GetAccount(assetName); err == nil {
		return ErrAssetNameEqualAccountName
	}

	_, err = asm.getAssetIDByName(assetName)
	if err == nil {
		return ErrAssetIsExist
	} else if err != ErrAssetNotExist {
		return err
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

func (asm *AssetManager) initAssetCount() {
	_, err := asm.getAssetCount()
	if err == ErrAssetCountNotExist {
		var assetID uint64
		b, err := rlp.EncodeToBytes(&assetID)
		if err != nil {
			panic(err)
		}
		asm.sdb.Put(assetManagerName, assetCountPrefix, b)
	}
}

func (asm *AssetManager) getAssetCount() (uint64, error) {
	b, err := asm.sdb.Get(assetManagerName, assetCountPrefix)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, ErrAssetCountNotExist
	}
	var assetCount uint64
	err = rlp.DecodeBytes(b, &assetCount)
	if err != nil {
		return 0, err
	}
	return assetCount, nil
}

func (asm *AssetManager) getAssetByID(assetID uint64) (*Asset, error) {
	b, err := asm.sdb.Get(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetID, 10))
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

func (asm *AssetManager) getAssetIDByName(assetName string) (uint64, error) {
	if assetName == "" {
		return 0, ErrAssetNameEmpty
	}
	b, err := asm.sdb.Get(assetManagerName, assetNameIDPrefix+assetName)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, ErrAssetNotExist
	}
	var assetID uint64
	if err := rlp.DecodeBytes(b, &assetID); err != nil {
		return 0, err
	}
	return assetID, nil
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

func (asm *AssetManager) addNewAssetObject(ao *Asset) (uint64, error) {
	if ao == nil {
		return 0, ErrAssetObjectEmpty
	}
	//get assetCount
	assetCount, err := asm.getAssetCount()
	if err != nil {
		return 0, err
	}

	ao.AssetID = assetCount
	//store asset object
	object, err := rlp.EncodeToBytes(ao)
	if err != nil {
		return 0, err
	}

	//store asset name with asset id
	assetID, err := rlp.EncodeToBytes(&assetCount)
	if err != nil {
		return 0, err
	}

	assetCount2 := assetCount + 1
	aid, err := rlp.EncodeToBytes(&assetCount2)
	if err != nil {
		return 0, err
	}

	asm.sdb.Put(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetCount, 10), object)
	asm.sdb.Put(assetManagerName, assetNameIDPrefix+ao.AssetName, assetID)
	//store assetCount
	asm.sdb.Put(assetManagerName, assetCountPrefix, aid)

	return assetCount, nil
}

var (
	ErrNewAssetManagerErr        = errors.New("new AssetManager error")
	ErrAssetIsExist              = errors.New("asset is exist")
	ErrAssetNotExist             = errors.New("asset not exist")
	ErrNewAssetObject            = errors.New("create asset object input invalid")
	ErrDetailTooLong             = errors.New("detail info exceed maximum")
	ErrAssetObjectEmpty          = errors.New("asset object is empty")
	ErrAssetCountNotExist        = errors.New("asset total count not exist")
	ErrAssetNameEmpty            = errors.New("asset name is empty")
	ErrOwnerMismatch             = errors.New("asset owner mismatch")
	ErrParamIsNil                = errors.New("param is nil")
	ErrUpperLimit                = errors.New("asset amount over the issuance limit")
	ErrDestroyLimit              = errors.New("asset destroy exceeding the lower limit")
	ErrAssetNameEqualAccountName = errors.New("asset name equal account name")
	ErrIssueAsset                = errors.New("issue asset err")
	ErrAssetNameinvalid          = errors.New("asset name invalid")
	ErrAssetNameLengthErr        = errors.New("asset name length err")
)
