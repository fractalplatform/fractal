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
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/snapshot"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	assetRegExp       = regexp.MustCompile(`^([a-z][a-z0-9]{1,15})(?:\.([a-z0-9]{1,8})){0,1}$`)
	assetNameLength   = uint64(31)
	assetManagerName  = "assetAccount"
	assetCountPrefix  = "assetCount"
	assetNameIDPrefix = "assetNameId"
	assetObjectPrefix = "assetDefinitionObject"
)

type Asset struct {
	sdb *state.StateDB
}

func SetAssetNameConfig(config *Config) bool {
	if config.AssetNameLevel < 1 || config.AssetNameLength < config.MainAssetNameMinLength || config.MainAssetNameMinLength >= config.MainAssetNameMaxLength {
		panic("asset name level config error")
	}

	if config.AssetNameLevel > 1 && (config.SubAssetNameMinLength < 1 || config.SubAssetNameMinLength >= config.SubAssetNameMaxLength) {
		return false
	}

	regexpStr := fmt.Sprintf("([a-z][a-z0-9]{%v,%v})", config.MainAssetNameMinLength-1, config.MainAssetNameMaxLength-1)
	for i := 1; i < int(config.AssetNameLevel); i++ {
		regexpStr += fmt.Sprintf("(?:\\.([a-z0-9]{%v,%v})){0,1}", config.SubAssetNameMinLength, config.SubAssetNameMaxLength)
	}

	regexp, err := regexp.Compile(fmt.Sprintf("^%s$", regexpStr))
	if err != nil {
		panic(err)
	}
	assetRegExp = regexp
	assetNameLength = config.AssetNameLength
	return true
}

func GetAssetNameRegExp() *regexp.Regexp {
	return assetRegExp
}

func GetAssetNameLength() uint64 {
	return assetNameLength
}

//SetAssetMangerName  set the global asset manager name
func SetAssetMangerName(name common.Name) {
	assetManagerName = name.String()
}

//NewAsset New create Asset
func NewAsset(sdb *state.StateDB) *Asset {
	asset := Asset{
		sdb: sdb,
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
	snapshotManager := snapshot.NewSnapshotManager(a.sdb)
	b, err := snapshotManager.GetSnapshotMsg(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetID, 10), time)
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

//GetAssetIDByName get assset id by asset name
func (a *Asset) GetAssetIDByName(assetName string) (uint64, error) {
	if assetName == "" {
		return 0, ErrAssetNameEmpty
	}
	b, err := a.sdb.Get(assetManagerName, assetNameIDPrefix+assetName)
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

//GetAssetFounderByID get asset founder by ID
func (a *Asset) GetAssetFounderByID(ID uint64) (common.Name, error) {
	ao, err := a.GetAssetObjectByID(ID)
	if err != nil {
		return "", err
	}
	return ao.GetAssetFounder(), nil
}

//GetAssetObjectByID get asset by asset ID
func (a *Asset) GetAssetObjectByID(ID uint64) (*AssetObject, error) {
	b, err := a.sdb.Get(assetManagerName, assetObjectPrefix+strconv.FormatUint(ID, 10))
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
		var assetID uint64 //store assetCount
		b, err := rlp.EncodeToBytes(&assetID)
		if err != nil {
			panic(err)
		}
		a.sdb.Put(assetManagerName, assetCountPrefix, b)
	}
}

//GetAllAssetObject get all asset
// func (a *Asset) GetAllAssetObject() ([]*AssetObject, error) {
// 	assetCount, err := a.getAssetCount()
// 	if err != nil {
// 		return nil, err
// 	}
// 	assets := make([]*AssetObject, assetCount)
// 	//
// 	var i uint64
// 	for i = 1; i <= assetCount; i++ {
// 		asset, err := a.GetAssetObjectByID(i)
// 		if err != nil {
// 			return nil, err
// 		}
// 		assets[i-1] = asset
// 	}
// 	return assets, nil
// }

//GetAssetObjectByName get asset object by name
func (a *Asset) GetAssetObjectByName(assetName string) (*AssetObject, error) {
	assetID, err := a.GetAssetIDByName(assetName)
	if err != nil {
		return nil, err
	}
	return a.GetAssetObjectByID(assetID)
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

	ao.SetAssetID(assetCount)
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

	a.sdb.Put(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetCount, 10), object)
	a.sdb.Put(assetManagerName, assetNameIDPrefix+ao.GetAssetName(), assetID)
	//store assetCount
	a.sdb.Put(assetManagerName, assetCountPrefix, aid)

	return assetCount, nil
}

//SetAssetObject store an asset into database
func (a *Asset) SetAssetObject(ao *AssetObject) error {
	if ao == nil {
		return ErrAssetObjectEmpty
	}
	assetID := ao.GetAssetID()

	b, err := rlp.EncodeToBytes(ao)
	if err != nil {
		return err
	}
	a.sdb.Put(assetManagerName, assetObjectPrefix+strconv.FormatUint(assetID, 10), b)
	return nil
}

//IssueAssetObject Issue Asset Object
func (a *Asset) IssueAssetObject(ao *AssetObject) (uint64, error) {
	if ao == nil {
		return 0, ErrAssetObjectEmpty
	}

	_, err := a.GetAssetIDByName(ao.GetAssetName())
	if err != nil && err != ErrAssetNotExist {
		return 0, err
	}

	if err == nil {
		return 0, ErrAssetIsExist
	}

	assetID, err := a.addNewAssetObject(ao)
	if err != nil {
		return 0, err
	}
	return assetID, nil
}

//IssueAsset issue asset
func (a *Asset) IssueAsset(assetName string, number uint64, forkID uint64, symbol string, amount *big.Int, dec uint64, founder common.Name, owner common.Name, limit *big.Int, contract common.Name, description string) (uint64, error) {
	if forkID >= params.ForkID4 {
		if amount.Cmp(math.MaxBig256) > 0 {
			return 0, ErrAmountOverMax256
		}
	}
	_, err := a.GetAssetIDByName(assetName)
	if err != nil && err != ErrAssetNotExist {
		return 0, err
	}

	if err == nil {
		return 0, ErrAssetIsExist
	}

	var ao *AssetObject

	if forkID >= params.ForkID1 {
		ao = NewAssetObjectNoCheck(assetName, number, symbol, amount, dec, founder, owner, limit, contract, description)
	} else {
		ao, err = NewAssetObject(assetName, number, symbol, amount, dec, founder, owner, limit, contract, description)
		if err != nil {
			return 0, err
		}
	}

	return a.addNewAssetObject(ao)
}

//DestroyAsset destroy asset
func (a *Asset) DestroyAsset(accountName common.Name, assetID uint64, amount *big.Int) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	if amount.Sign() < 0 {
		return ErrNegativeAmount
	}
	if amount.Sign() == 0 {
		return nil
	}

	asset, err := a.GetAssetObjectByID(assetID)
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
func (a *Asset) IncreaseAsset(accountName common.Name, assetID uint64, amount *big.Int, forkID uint64) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	if amount.Sign() < 0 {
		return ErrNegativeAmount
	}
	if amount.Sign() == 0 {
		return nil
	}
	asset, err := a.GetAssetObjectByID(assetID)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotExist
	}
	if forkID >= params.ForkID4 {
		if (new(big.Int).Add(asset.GetAssetAmount(), amount)).Cmp(math.MaxBig256) > 0 {
			return ErrAmountOverMax256
		}
	}
	// if asset.GetAssetOwner() != accountName {
	// 	return ErrOwnerMismatch
	// }

	//check AddIssue > UpperLimit
	AddIssue := new(big.Int).Add(asset.GetAssetAddIssue(), amount)
	if asset.GetUpperLimit().Cmp(big.NewInt(0)) > 0 && AddIssue.Cmp(asset.GetUpperLimit()) > 0 {
		return ErrUpperLimit
	}
	asset.SetAssetAddIssue(AddIssue)

	//check Amount > UpperLimit
	total := new(big.Int).Add(asset.GetAssetAmount(), amount)
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

//UpdateAsset change asset info
func (a *Asset) UpdateAsset(accountName common.Name, assetID uint64, founderName common.Name, curForkID uint64) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	asset, err := a.GetAssetObjectByID(assetID)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotExist
	}
	// if asset.GetAssetOwner() != accountName {
	// 	return ErrOwnerMismatch
	// }
	if curForkID >= params.ForkID4 {
		if len(founderName.String()) == 0 {
			assetOwner := asset.GetAssetOwner()
			asset.SetAssetFounder(assetOwner)
		} else {
			asset.SetAssetFounder(founderName)
		}
	} else {
		asset.SetAssetFounder(founderName)
	}
	return a.SetAssetObject(asset)
}

//SetAssetNewOwner change asset owner
func (a *Asset) SetAssetNewOwner(accountName common.Name, assetID uint64, newOwner common.Name) error {
	if accountName == "" {
		return ErrAccountNameNull
	}
	asset, err := a.GetAssetObjectByID(assetID)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotExist
	}
	// if asset.GetAssetOwner() != accountName {
	// 	return ErrOwnerMismatch
	// }
	asset.SetAssetOwner(newOwner)
	return a.SetAssetObject(asset)
}

func (a *Asset) SetAssetNewContract(assetID uint64, contract common.Name) error {
	assetObj, err := a.GetAssetObjectByID(assetID)
	if err != nil {
		return err
	}
	if assetObj == nil {
		return ErrAssetNotExist
	}

	assetObj.SetAssetContract(contract)
	return a.SetAssetObject(assetObj)
}

//SetAssetFounder asset founder
// func (a *Asset) SetAssetFounder(accountName common.Name, assetId uint64, founderName common.Name) error {
// 	if accountName == "" {
// 		return ErrAccountNameNull
// 	}
// 	if assetId == 0 {
// 		return ErrAssetIDInvalid
// 	}
// 	asset, err := a.GetAssetObjectByID(assetId)
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

func (a *Asset) IsValidMainAssetBeforeFork(assetName string) bool {
	assetNames := common.FindStringSubmatch(assetRegExp, assetName)
	if len(assetNames) < 2 {
		return true
	}
	return false
}

// IsValidSubAssetBeforeFork check parent owner valid
func (a *Asset) IsValidSubAssetBeforeFork(fromName common.Name, assetName string) (uint64, bool) {
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

		assetID, err := a.GetAssetIDByName(an)
		if err != nil {
			continue
		}

		assetObj, err := a.GetAssetObjectByID(assetID)
		if err != nil {
			continue
		}

		if assetObj == nil {
			continue
		}

		if assetObj.GetAssetOwner() == fromName {
			log.Debug("Asset create", "name", an, "owner", assetObj.GetAssetOwner(), "fromName", fromName, "newName", assetName)
			return assetID, true
		}
	}
	log.Debug("Asset create failed", "account", fromName, "name", assetName)
	return 0, false
}

func (a *Asset) IsValidAssetOwner(fromName common.Name, assetPrefix string, assetNames []string) (uint64, bool) {
	var an string
	for i := 0; i < len(assetNames)-1; i++ {
		if i == 0 {
			an = assetPrefix + assetNames[i]
		} else {
			an = an + "." + assetNames[i]
		}

		assetID, err := a.GetAssetIDByName(an)
		if err != nil {
			continue
		}
		assetObj, err := a.GetAssetObjectByID(assetID)
		if err != nil {
			continue
		}

		if assetObj == nil {
			continue
		}

		if assetObj.GetAssetOwner() == fromName {
			return assetID, true
		}
	}
	return 0, false
}

// HasAccess contract asset access
func (a *Asset) HasAccess(assetID uint64, names ...common.Name) bool {
	ast, _ := a.GetAssetObjectByID(assetID)
	if ast != nil && len(ast.Contract.String()) != 0 {
		for _, name := range names {
			if name == ast.Contract {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

func (a *Asset) IncStats(assetID uint64) error {
	assetObj, err := a.GetAssetObjectByID(assetID)
	if err != nil {
		return err
	}
	count := assetObj.GetAssetStats()
	count = count + 1
	assetObj.SetAssetStats(count)

	err = a.SetAssetObject(assetObj)
	if err != nil {
		return err
	}
	return nil
}

func (a *Asset) CheckOwner(fromName common.Name, assetID uint64) error {
	assetObj, err := a.GetAssetObjectByID(assetID)
	if err != nil {
		return err
	}

	if assetObj.GetAssetOwner() != fromName {
		var assetNames []string
		var assetPrefix string
		names := strings.Split(assetObj.GetAssetName(), ":")
		if len(names) == 2 {
			assetNames = strings.Split(names[1], ".")
			assetPrefix = names[0] + ":"
		} else {
			assetNames = strings.Split(assetObj.GetAssetName(), ".")
			assetPrefix = ""
		}

		if len(assetNames) == 1 {
			return ErrOwnerMismatch
		}

		if _, isValid := a.IsValidAssetOwner(fromName, assetPrefix, assetNames); !isValid {
			return ErrOwnerMismatch
		}
	}
	return nil
}
