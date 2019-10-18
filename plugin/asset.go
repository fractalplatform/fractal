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
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/utils/rlp"
	"github.com/meitu/go-ethereum/log"
)

var (
	assetRegExp       = regexp.MustCompile(`^([a-z][a-z0-9]{1,15})(?:\.([a-z0-9]{1,8})){0,1}$`)
	assetNameLength   = uint64(31)
	assetManagerName  = "assetAccount"
	assetCountPrefix  = "assetCount"
	assetNameIDPrefix = "assetNameId"
	assetObjectPrefix = "assetDefinitionObject"
)

type AssetManager struct {
	sdb *state.StateDB
}

type Asset struct {
	AssetID     uint64         `json:"assetId"`
	Stats       uint64         `json:"stats"`
	AssetName   string         `json:"assetName"`
	Symbol      string         `json:"symbol"`
	Amount      *big.Int       `json:"amount"`
	Decimals    uint64         `json:"decimals"`
	Founder     common.Address `json:"founder"`
	Owner       common.Address `json:"owner"`
	AddIssue    *big.Int       `json:"addIssue"`
	UpperLimit  *big.Int       `json:"upperLimit"`
	Description string         `json:"description"`
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

func (asm *AssetManager) IncStats(assetID uint64) error {
	asset, err := asm.getAssetByID(assetID)
	if err != nil {
		return err
	}

	asset.Stats += 1

	err = asm.setAsset(asset)
	if err != nil {
		return err
	}

	return nil
}

func (asm *AssetManager) CheckIssueAssetInfo(account common.Address, assetInfo *IssueAsset) error {
	assetNames := common.FindStringSubmatch(assetRegExp, assetInfo.AssetName)
	if len(assetNames) < 2 {
		return nil
	}

	parentAssetID, isValid := asm.isValidSubAssetBeforeFork(account, assetInfo.AssetName)
	if !isValid {
		return fmt.Errorf("account %s can not create %s", account.String(), assetInfo.AssetName)
	}
	assetObj, _ := asm.getAssetObjectByID(parentAssetID)
	assetInfo.Decimals = assetObj.Decimals

	return nil
}

func (asm *AssetManager) IssueAssetForAccount(assetName string, symbol string, amount *big.Int, dec uint64, founder common.Address, owner common.Address, limit *big.Int, description string) (uint64, error) {
	_, err := asm.getAssetIDByName(assetName)
	if err != nil && err != ErrAssetNotExist {
		return 0, err
	}

	if err == nil {
		return 0, ErrAssetIsExist
	}

	var ao *Asset

	ao, err = asm.newAssetObject(assetName, symbol, amount, dec, founder, owner, limit, description)
	if err != nil {
		return 0, err
	}

	return asm.addNewAssetObject(ao)
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

func (asm *AssetManager) isValidSubAssetBeforeFork(from common.Address, assetName string) (uint64, bool) {
	assetNames := common.FindStringSubmatch(assetRegExp, assetName)
	if len(assetNames) < 2 {
		return 0, false
	}

	if !common.StrToName(assetName).IsValid(assetRegExp, assetNameLength) {
		return 0, false
	}

	var an string
	for i := 0; i < len(assetNames)-1; i++ {
		if i == 0 {
			an = assetNames[i]
		} else {
			an = an + "." + assetNames[i]
		}

		assetID, err := asm.getAssetIDByName(an)
		if err != nil {
			continue
		}

		assetObj, err := asm.getAssetObjectByID(assetID)
		if err != nil {
			continue
		}

		if assetObj == nil {
			continue
		}

		if assetObj.Owner.Compare(from) == 0 {
			log.Debug("Asset create", "name", an, "owner", assetObj.Owner, "fromAddress", from, "newName", assetName)
			return assetID, true
		}
	}
	log.Debug("Asset create failed", "account", from, "name", assetName)
	return 0, false
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

func (asm *AssetManager) newAssetObject(assetName string, symbol string, amount *big.Int, dec uint64, founder common.Address, owner common.Address,
	limit *big.Int, description string) (*Asset, error) {
	if assetName == "" || symbol == "" {
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
	if !common.StrToName(assetName).IsValid(assetRegExp, assetNameLength) {
		return nil, ErrNewAssetObject
	}
	if !common.StrToName(symbol).IsValid(assetRegExp, assetNameLength) {
		return nil, ErrNewAssetObject
	}
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDetailTooLong
	}

	ao := Asset{
		AssetID:     0,
		Stats:       0,
		AssetName:   assetName,
		Symbol:      symbol,
		Amount:      amount,
		Decimals:    dec,
		Founder:     founder,
		Owner:       owner,
		AddIssue:    amount,
		UpperLimit:  limit,
		Description: description,
	}
	return &ao, nil
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
	ErrAccountNameNull    = errors.New("account name is null")
	ErrAssetIsExist       = errors.New("asset is exist")
	ErrAssetNotExist      = errors.New("asset not exist")
	ErrOwnerMismatch      = errors.New("asset owner mismatch")
	ErrAssetNameEmpty     = errors.New("asset name is empty")
	ErrAssetObjectEmpty   = errors.New("asset object is empty")
	ErrNewAssetObject     = errors.New("create asset object input invalid")
	ErrAssetAmountZero    = errors.New("asset amount is zero")
	ErrUpperLimit         = errors.New("asset amount over the issuance limit")
	ErrDestroyLimit       = errors.New("asset destroy exceeding the lower limit")
	ErrAssetCountNotExist = errors.New("asset total count not exist")
	//ErrAssetIDInvalid       = errors.New("asset id invalid")
	ErrAssetManagerNotExist = errors.New("asset manager name not exist")
	ErrDetailTooLong        = errors.New("detail info exceed maximum")
	//ErrNegativeAmount       = errors.New("negative amount")
	ErrNewAssetManagerErr = errors.New("new AssetManager error")
)
