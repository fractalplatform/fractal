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
	"strconv"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/utils/rlp"
)

//AssetManager is used to access asset
var assetManagerName = "sysAccount"

var (
	assetCountPrefix  = "assetCount"
	assetNameIdPrefix = "assetNameId"
	assetObjectPrefix = "assetDefinitionObject"
)

type Asset struct {
	sdb *state.StateDB
}

//SetAssetMangerName  set the global asset manager name
func SetAssetMangerName(name common.Name) bool {
	if common.IsValidName(name.String()) {
		assetManagerName = name.String()
		return true
	}
	return false
}

//NewAsset New create Asset
func NewAsset(sdb *state.StateDB) *Asset {
	asset := Asset{
		sdb: sdb,
	}

	if len(assetManagerName) == 0 {
		log.Error("NewAsset error", "name", ErrAssetManagerNotExist, assetManagerName)
		return nil
	}

	asset.InitAssetCount()
	return &asset
}

//GetAssetAmountByTime get asset amount by time
func (a *Asset) GetAssetAmountByTime(assetID uint64, time uint64) (*big.Int, error) {
	ao, err := a.GetAssetObjectByTime(assetID, time)
	if err != nil {
		return big.NewInt(0), err
	}
	return ao.GetAssetAmount(), nil
}

//GetAssetObjectByTime  get asset object by time
func (a *Asset) GetAssetObjectByTime(assetID uint64, time uint64) (*AssetObject, error) {
	if assetID == 0 {
		return nil, ErrAssetIdInvalid
	}
	b, err := a.sdb.GetSnapshot(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetID, 10), time)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrAssetNotExist
	}
	var asset AssetObject
	if err := rlp.DecodeBytes(b, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

//GetAssetIdByName get assset id by asset name
func (a *Asset) GetAssetIdByName(assetName string) (uint64, error) {
	if assetName == "" {
		return 0, ErrAssetNameEmpty
	}
	b, err := a.sdb.Get(assetManagerName, assetNameIdPrefix+assetName)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	var assetID uint64
	if err := rlp.DecodeBytes(b, &assetID); err != nil {
		return 0, err
	}
	return assetID, nil
}

//GetAssetFounderById get asset founder by id
func (a *Asset) GetAssetFounderById(id uint64) (common.Name, error) {
	ao, err := a.GetAssetObjectById(id)
	if err != nil {
		return "", err
	}
	return ao.GetAssetFounder(), nil
}

//GetAssetObjectById get asset by asset id
func (a *Asset) GetAssetObjectById(id uint64) (*AssetObject, error) {
	if id == 0 {
		return nil, ErrAssetIdInvalid
	}
	b, err := a.sdb.Get(assetManagerName, assetObjectPrefix+strconv.FormatUint(id, 10))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	var asset AssetObject
	if err := rlp.DecodeBytes(b, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

//get asset total count
func (a *Asset) getAssetCount() (uint64, error) {
	b, err := a.sdb.Get(assetManagerName, assetCountPrefix)
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

//InitAssetCount init asset count
func (a *Asset) InitAssetCount() {
	_, err := a.getAssetCount()
	if err == ErrAssetCountNotExist {
		var assetID uint64
		assetID = 0
		//store assetCount
		b, err := rlp.EncodeToBytes(&assetID)
		if err != nil {
			panic(err)
		}
		a.sdb.Put(assetManagerName, assetCountPrefix, b)
	}
	return
}

//GetAllAssetObject get all asset
func (a *Asset) GetAllAssetObject() ([]*AssetObject, error) {
	assetCount, err := a.getAssetCount()
	if err != nil {
		return nil, err
	}
	assets := make([]*AssetObject, assetCount)
	//
	var i uint64
	for i = 1; i <= assetCount; i++ {
		asset, err := a.GetAssetObjectById(i)
		if err != nil {
			return nil, err
		}
		assets[i] = asset
	}
	return assets, nil
}

//GetAssetObjectByName get asset object by name
func (a *Asset) GetAssetObjectByName(assetName string) (*AssetObject, error) {
	assetID, err := a.GetAssetIdByName(assetName)
	if err != nil {
		return nil, err
	}
	return a.GetAssetObjectById(assetID)
}

//addNewAssetObject add new asset object and store into database
func (a *Asset) addNewAssetObject(ao *AssetObject) (uint64, error) {
	if ao == nil {
		return 0, ErrAssetObjectEmpty
	}
	//get assetCount
	assetCount, err := a.getAssetCount()
	if err != nil {
		return 0, err
	}
	assetCount = assetCount + 1
	ao.SetAssetId(assetCount)
	//store asset object
	aobject, err := rlp.EncodeToBytes(ao)
	if err != nil {
		return 0, err
	}
	//store asset name with asset id
	aid, err := rlp.EncodeToBytes(&assetCount)
	if err != nil {
		return 0, err
	}
	//store assetCount
	b, err := rlp.EncodeToBytes(&assetCount)
	if err != nil {
		return 0, err
	}

	a.sdb.Put(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetCount, 10), aobject)
	a.sdb.Put(assetManagerName, assetNameIdPrefix+ao.GetAssetName(), aid)
	a.sdb.Put(assetManagerName, assetCountPrefix, b)
	return assetCount, nil
}

//SetAssetObject store an asset into database
func (a *Asset) SetAssetObject(ao *AssetObject) error {
	if ao == nil {
		return ErrAssetObjectEmpty
	}
	assetId := ao.GetAssetId()
	if assetId == 0 {
		return ErrAssetIdInvalid
	}
	b, err := rlp.EncodeToBytes(ao)
	if err != nil {
		return err
	}
	a.sdb.Put(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetId, 10), b)
	return nil
}

//IssueAssetObject Issue Asset Object
func (a *Asset) IssueAssetObject(ao *AssetObject) (uint64, error) {
	if ao == nil {
		return 0, ErrAssetObjectEmpty
	}
	assetId, err := a.GetAssetIdByName(ao.GetAssetName())
	if err != nil {
		return 0, err
	}
	if assetId > 0 {
		return 0, ErrAssetIsExist
	}
	assetID, err := a.addNewAssetObject(ao)
	if err != nil {
		return 0, err
	}
	return assetID, nil
}

//IssueAsset issue asset
func (a *Asset) IssueAsset(assetName string, symbol string, amount *big.Int, dec uint64, founder common.Name, owner common.Name, limit *big.Int) error {
	assetId, err := a.GetAssetIdByName(assetName)
	if err != nil {
		return err
	}
	if assetId > 0 {
		return ErrAssetIsExist
	}
	ao, err := NewAssetObject(assetName, symbol, amount, dec, founder, owner, limit)
	if err != nil {
		return err
	}
	assetId, err = a.addNewAssetObject(ao)
	if err != nil {
		return err
	}
	return nil
}

//DestroyAsset destroy asset
func (a *Asset) DestroyAsset(accountName common.Name, assetId uint64, amount *big.Int) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	if assetId == 0 {
		return ErrAssetIdInvalid
	}
	if amount.Sign() <= 0 {
		return ErrAssetAmountZero
	}
	asset, err := a.GetAssetObjectById(assetId)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotExist
	}

	//everyone can destory asset
	// if asset.GetAssetOwner() != accountName {
	// 	return ErrOwnerMismatch
	// }

	var total *big.Int
	if total = new(big.Int).Sub(asset.GetAssetAmount(), amount); total.Cmp(big.NewInt(0)) < 0 {
		return ErrDestroyLimit
	}
	asset.SetAssetAmount(total)
	err = a.SetAssetObject(asset)
	if err != nil {
		return err
	}
	return nil
}

//IncreaseAsset increase asset, upperlimit == 0 means no upper limit
func (a *Asset) IncreaseAsset(accountName common.Name, assetId uint64, amount *big.Int) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	if assetId == 0 {
		return ErrAssetIdInvalid
	}
	if amount.Sign() <= 0 {
		return ErrAssetAmountZero
	}
	asset, err := a.GetAssetObjectById(assetId)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotExist
	}
	if asset.GetAssetOwner() != accountName {
		return ErrOwnerMismatch
	}

	//check AddIssue > UpperLimit
	var AddIssue *big.Int
	AddIssue = new(big.Int).Add(asset.GetAssetAddIssue(), amount)
	if asset.GetUpperLimit().Cmp(big.NewInt(0)) > 0 && AddIssue.Cmp(asset.GetUpperLimit()) > 0 {
		return ErrUpperLimit
	}
	asset.SetAssetAddIssue(AddIssue)

	//check Amount > UpperLimit
	var total *big.Int
	total = new(big.Int).Add(asset.GetAssetAmount(), amount)
	if asset.GetUpperLimit().Cmp(big.NewInt(0)) > 0 && total.Cmp(asset.GetUpperLimit()) > 0 {
		return ErrUpperLimit
	}
	asset.SetAssetAmount(total)
	//save
	err = a.SetAssetObject(asset)
	if err != nil {
		return err
	}
	return nil
}

//SetAssetNewOwner change asset owner
func (a *Asset) UpdateAsset(accountName common.Name, assetId uint64, Owner common.Name, founderName common.Name) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	if assetId == 0 {
		return ErrAssetIdInvalid
	}
	asset, err := a.GetAssetObjectById(assetId)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotExist
	}
	if asset.GetAssetOwner() != accountName {
		return ErrOwnerMismatch
	}
	asset.SetAssetOwner(Owner)
	asset.SetAssetFounder(founderName)
	return a.SetAssetObject(asset)
}

//SetAssetNewOwner change asset owner
func (a *Asset) SetAssetNewOwner(accountName common.Name, assetId uint64, newOwner common.Name) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	if assetId == 0 {
		return ErrAssetIdInvalid
	}
	asset, err := a.GetAssetObjectById(assetId)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotExist
	}
	if asset.GetAssetOwner() != accountName {
		return ErrOwnerMismatch
	}
	asset.SetAssetOwner(newOwner)
	return a.SetAssetObject(asset)
}

//SetAssetFounder asset founder
// func (a *Asset) SetAssetFounder(accountName common.Name, assetId uint64, founderName common.Name) error {
// 	if accountName == "" {
// 		return ErrAccountNameNull
// 	}
// 	if assetId == 0 {
// 		return ErrAssetIdInvalid
// 	}
// 	asset, err := a.GetAssetObjectById(assetId)
// 	if err != nil {
// 		return err
// 	}
// 	if asset == nil {
// 		return ErrAssetNotExist
// 	}
// 	if asset.GetAssetOwner() != accountName {
// 		return ErrOwnerMismatch
// 	}
// 	asset.SetAssetFounder(founderName)
// 	return a.SetAssetObject(asset)
// }
