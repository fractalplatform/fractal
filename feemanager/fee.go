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

package feemanager

import (
	"fmt"
	"math/big"
	"strconv"

	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	feeCounterKey     = "feeCounter"
	objectFeeIDPrefix = "feeIdPrefix"
	objectFeePrefix   = "feePrefix"
)

type feeManagerConfig struct {
	feeName string
}

//FeeManager account fee manager
type FeeManager struct {
	name      string
	accountDB *am.AccountManager
	stateDB   *state.StateDB
}

//AssetFee asset fee
type AssetFee struct {
	AssetID   uint64   `json:"assetID"`
	TotalFee  *big.Int `json:"totalFee"`
	RemainFee *big.Int `json:"remainFee"`
}

//ObjectFee object's fee
type ObjectFee struct {
	ObjectFeeID uint64      `json:"objectFeeID"`
	ObjectType  uint64      `json:"objectType"`
	ObjectName  string      `json:"objectName"`
	AssetFees   []*AssetFee `json:"assetFee"`
}

//WithdrawInfo record withdraw info
type WithdrawInfo struct {
	ObjectName common.Name
	ObjectType uint64
	Founder    common.Name
	AssetID    uint64
	Amount     *big.Int
}

var feeConfig feeManagerConfig

//NewFeeManager new fee manager
func NewFeeManager(state *state.StateDB, accountDB *am.AccountManager) *FeeManager {
	return &FeeManager{name: feeConfig.feeName,
		stateDB:   state,
		accountDB: accountDB}
}

//SetFeeManagerName set fee manager name
func SetFeeManagerName(name common.Name) bool {
	if common.IsValidAccountName(name.String()) {
		feeConfig.feeName = name.String()
		return true
	}
	return false
}

func newAssetFee(assetID uint64, value *big.Int) *AssetFee {
	return &AssetFee{AssetID: assetID,
		TotalFee:  new(big.Int).Set(value),
		RemainFee: new(big.Int).Set(value)}
}

func (fm *FeeManager) getFeeCounter() (uint64, error) {
	countEnc, err := fm.stateDB.Get(fm.name, feeCounterKey)
	if err != nil {
		errInfo := fmt.Errorf("get fee counter failed, err %v", err)
		return 0, errInfo
	}

	//fee counter start from zero
	if len(countEnc) == 0 {
		return 0, nil
	}

	var objectFeeCounter uint64
	err = rlp.DecodeBytes(countEnc, &objectFeeCounter)
	if err != nil {
		errInfo := fmt.Errorf("decode fee counter failed, err %v", err)
		return 0, errInfo
	}
	return objectFeeCounter, nil
}

//GetObjectFeeByName get object fee by name
func (fm *FeeManager) GetObjectFeeByName(objectName common.Name) (*ObjectFee, error) {
	objectFeeID, err := fm.getObjectFeeIDByName(objectName)

	if err != nil || objectFeeID == 0 {
		return nil, err
	}

	return fm.getObjectFeeByID(objectFeeID)
}

func (fm *FeeManager) getObjectFeeIDByName(objectName common.Name) (uint64, error) {
	feeIDEnc, err := fm.stateDB.Get(fm.name, objectFeeIDPrefix+objectName.String())

	if err != nil || len(feeIDEnc) == 0 {
		return 0, err
	}
	var objectFeeID uint64
	if err = rlp.DecodeBytes(feeIDEnc, &objectFeeID); err != nil {
		return 0, err
	}
	return objectFeeID, nil
}

func (fm *FeeManager) getObjectFeeByID(objectFeeID uint64) (*ObjectFee, error) {
	key := objectFeePrefix + strconv.FormatUint(objectFeeID, 10)
	objectFeeEnc, err := fm.stateDB.Get(fm.name, key)

	if err != nil || len(objectFeeEnc) == 0 {
		return nil, err
	}

	var objectFee ObjectFee
	if err = rlp.DecodeBytes(objectFeeEnc, &objectFee); err != nil {
		return nil, err
	}

	return &objectFee, nil
}

func (fm *FeeManager) setObjectFee(objectFee *ObjectFee) error {
	value, err := rlp.EncodeToBytes(objectFee)
	if err != nil {
		return err
	}

	key := objectFeePrefix + strconv.FormatUint(objectFee.ObjectFeeID, 10)
	fm.stateDB.Put(fm.name, key, value)
	return nil
}

func (fm *FeeManager) createObjectFee(objectName common.Name, objectType uint64) (*ObjectFee, error) {
	//get object fee id
	feeCounter, err := fm.getFeeCounter()
	if err != nil {
		return nil, err
	}

	feeCounter = feeCounter + 1
	objectFee := &ObjectFee{ObjectFeeID: feeCounter,
		ObjectName: objectName.String(),
		ObjectType: objectType,
		AssetFees:  make([]*AssetFee, 0)}

	value, err := rlp.EncodeToBytes(&feeCounter)
	if err != nil {
		return nil, err
	}
	fm.stateDB.Put(fm.name, objectFeeIDPrefix+objectName.String(), value)
	fm.stateDB.Put(fm.name, feeCounterKey, value)

	return objectFee, nil
}

func (assetFee *AssetFee) addFee(value *big.Int) {
	assetFee.TotalFee.Add(assetFee.TotalFee, value)
	assetFee.RemainFee.Add(assetFee.RemainFee, value)
}

// BinarySearch binary search, find insert position
func (of *ObjectFee) binarySearch(assetID uint64) (int64, bool) {
	if len(of.AssetFees) == 0 {
		return 0, false
	}
	low := int64(0)
	high := int64(len(of.AssetFees)) - 1

	for low <= high {
		mid := (low + high) / 2
		if of.AssetFees[mid].AssetID < assetID {
			low = mid + 1
		} else if of.AssetFees[mid].AssetID > assetID {
			high = mid - 1
		} else if of.AssetFees[mid].AssetID == assetID {
			return mid, true
		}
	}
	high = high + 1
	return high, false
}

func (of *ObjectFee) addAssetFee(assetID uint64, value *big.Int) {
	index, find := of.binarySearch(assetID)

	if find == true {
		assetFee := of.AssetFees[index]
		assetFee.addFee(value)
		return
	}

	assetFee := newAssetFee(assetID, value)

	//insert index pos
	if index == (int64)(len(of.AssetFees)) {
		of.AssetFees = append(of.AssetFees, assetFee)
		return
	}

	tmp := append([]*AssetFee{}, of.AssetFees[index:]...)
	of.AssetFees = append(of.AssetFees[:index], assetFee)
	of.AssetFees = append(of.AssetFees, tmp...)
}

//RecordFeeInSystem record object fee in system
func (fm *FeeManager) RecordFeeInSystem(objectName common.Name, objectType uint64, assetID uint64, value *big.Int) error {
	//get object fee in system
	objectFee, err := fm.GetObjectFeeByName(objectName)

	if err != nil {
		return err
	}

	if objectFee == nil {
		objectFee, err = fm.createObjectFee(objectName, objectType)
		if err != nil {
			return err
		}
	}

	//modify object's asset fee
	objectFee.addAssetFee(assetID, value)

	//store object fee
	err = fm.setObjectFee(objectFee)
	if err != nil {
		return fmt.Errorf("set object(%s) fee failed, err:%v", objectName, err)
	}

	return nil
}

func (fm *FeeManager) getObjectFounder(objectName common.Name, objectType uint64) (common.Name, error) {
	var founder common.Name
	var err error

	if common.AssetName == objectType {
		var assetInfo *asset.AssetObject
		assetInfo, err = fm.accountDB.GetAssetInfoByName(objectName.String())
		if assetInfo != nil {
			founder = assetInfo.GetAssetFounder()
		}
	} else if common.ContractName == objectType {
		founder, err = fm.accountDB.GetFounder(objectName)
	} else if common.CoinbaseName == objectType {
		founder = objectName
	} else {
		err = fmt.Errorf("get founder failed, name:%s, type:%d", objectName, objectType)
	}
	return founder, err
}

//WithdrawFeeFromSystem withdraw object fee in system, return withdraw info
func (fm *FeeManager) WithdrawFeeFromSystem(objectName common.Name) ([]*WithdrawInfo, error) {
	var withdrawInfos []*WithdrawInfo

	//get fee info from system
	objectFee, err := fm.GetObjectFeeByName(objectName)

	if err != nil || objectFee == nil {
		return withdrawInfos, fmt.Errorf("object(%s) fee not exsit, err:%v", objectName, err)
	}

	founder, err1 := fm.getObjectFounder(objectName, objectFee.ObjectType)
	if err1 != nil || len(founder) == 0 {
		return withdrawInfos, fmt.Errorf("get object(%s) founder failed, err:%v", objectName, err1)
	}

	//store fee to object, scan all asset
	for _, assetFee := range objectFee.AssetFees {
		if assetFee.RemainFee.Cmp(big.NewInt(0)) > 0 {
			err = fm.accountDB.AddAccountBalanceByID(founder, assetFee.AssetID, assetFee.RemainFee)
			if err != nil {
				return withdrawInfos, fmt.Errorf("withdraw asset(%d) fee to founder(%s) err:%v", assetFee.AssetID, founder, err)
			}

			withdraw := &WithdrawInfo{ObjectName: objectName,
				ObjectType: objectFee.ObjectType,
				Founder:    founder,
				AssetID:    assetFee.AssetID,
				Amount:     new(big.Int).Set(assetFee.RemainFee)}
			withdrawInfos = append(withdrawInfos, withdraw)

			//clear remain fee
			assetFee.RemainFee = big.NewInt(0)
		}
	}

	//save fee modify info to db
	err = fm.setObjectFee(objectFee)
	return withdrawInfos, err
}
