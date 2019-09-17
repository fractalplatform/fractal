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
	"testing"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
)

var sdb = getStateDB()
var acctm = getAccountManager()
var fm = getFeeManager()
var ast = getAsset()

func getAsset() *asset.Asset {
	return asset.NewAsset(sdb)
}

func getStateDB() *state.StateDB {
	db := rawdb.NewMemoryDatabase()
	tridb := state.NewDatabase(db)
	statedb, err := state.New(common.Hash{}, tridb)
	if err != nil {
		fmt.Printf("test getStateDB() failure %v", err)
		return nil
	}

	return statedb
}

func getAccountManager() *accountmanager.AccountManager {
	accountmanager.SetAcctMangerName("systestname")
	am, err := accountmanager.NewAccountManager(sdb)
	if err != nil {
		fmt.Printf("test getAccountManager() failure %v", err)
	}
	pubkey := new(common.PubKey)
	pubkey.SetBytes([]byte("abcde123456789"))
	am.CreateAccount(common.Name("fractal.founder"), common.Name("systestname"), common.Name(""), 0, 0, *pubkey, "")
	am.CreateAccount(common.Name("fractal"), common.Name("fractal.fee"), common.Name(""), 0, 0, *pubkey, "")
	return am
}

func getFeeManager() *FeeManager {
	SetFeeManagerName(common.Name("fractal.fee"))
	return NewFeeManager(sdb, acctm)
}

func TestRecordFeeInSystem(t *testing.T) {
	type testFee struct {
		assetIndex int
		objectName string
		objectType uint64
		assetID    uint64
		value      *big.Int
		totalValue *big.Int
	}

	testFeeInfo := []*testFee{
		{1, "testtest.tt", uint64(params.ContractFeeType), uint64(2), big.NewInt(200), big.NewInt(700)},
		{0, "testtest.tt", uint64(params.ContractFeeType), uint64(1), big.NewInt(100), big.NewInt(100)},
		{3, "testtest.tt", uint64(params.ContractFeeType), uint64(4), big.NewInt(400), big.NewInt(400)},
		{2, "testtest.tt", uint64(params.ContractFeeType), uint64(3), big.NewInt(300), big.NewInt(300)},
		{1, "testtest.tt", uint64(params.ContractFeeType), uint64(2), big.NewInt(500), big.NewInt(700)},
		{0, "testtest.tt1", uint64(params.AssetFeeType), uint64(1), big.NewInt(600), big.NewInt(600)},
	}

	for _, tf := range testFeeInfo {
		objectName := tf.objectName
		objectType := uint64(tf.objectType)
		assetID := tf.assetID
		value := tf.value

		//NewFeeManager
		err := fm.RecordFeeInSystem(objectName, objectType, assetID, value)

		if err != nil {
			t.Errorf("record fee in system failed, err:%v", err)
			return
		}
	}

	//check
	for _, tf := range testFeeInfo {
		index := tf.assetIndex
		objectName := tf.objectName
		objectType := uint64(tf.objectType)
		assetID := tf.assetID
		totalValue := tf.totalValue

		objectID, err := fm.getObjectFeeIDByName(objectName, objectType)
		if err != nil {
			t.Errorf("get object id failed by id, err:%v", err)
			return
		}

		objectFee, err := fm.GetObjectFeeByID(objectID)

		if err != nil || objectFee == nil {
			t.Errorf("get objectfee failed by id, err:%v", err)
			return
		}

		if objectFee.ObjectFeeID != objectID || objectFee.ObjectType != objectType {
			t.Errorf("check object id and type failed")
		}

		/*test asset info*/
		assetFee := objectFee.AssetFees[index]
		if assetFee.AssetID != assetID || assetFee.TotalFee.Cmp(totalValue) != 0 ||
			assetFee.RemainFee.Cmp(totalValue) != 0 {
			t.Errorf("check asset fee failed, objectName:%v, assetId:%d", objectName, assetID)
		}
	}
}

func addAssetAndAccount() error {
	type args struct {
		assetName string
		symbol    string
		amount    *big.Int
		dec       uint64
		founder   common.Name
		owner     common.Name
	}

	var (
		tname  = common.Name("testtest.testact1")
		pubKey = new(common.PubKey)
	)

	tests := []args{
		// TODO: Add test cases.
		{"assettest.asset0", "s0", big.NewInt(0), 2, tname, tname},
		{"assettest.asset1", "s1", big.NewInt(0), 2, tname, tname},
		{"assettest.asset2", "s2", big.NewInt(0), 2, tname, tname},
		{"assettest.asset3", "s3", big.NewInt(0), 2, tname, tname},
		{"assettest.asset4", "s4", big.NewInt(0), 2, tname, tname},
	}

	if err := acctm.CreateAccount(common.Name("testtest"), tname, tname, 0, 0, *pubKey, ""); err != nil {
		return err
	}

	for _, tt := range tests {
		_, err := ast.IssueAsset(tt.assetName, 0, 0, tt.symbol, tt.amount, tt.dec, tt.founder, tt.owner, big.NewInt(9999999999), common.Name(""), "desv")
		if err != nil {
			return err
		}
	}
	return nil
}

func TestWithdrawFeeFromSystem(t *testing.T) {
	type testFee struct {
		assetIndex int
		objectName string
		objectType uint64
		assetID    uint64
		value      *big.Int
		totalValue *big.Int
	}

	testFeeInfo := []*testFee{
		{1, "assettest.asset1", uint64(params.AssetFeeType), uint64(2), big.NewInt(200), big.NewInt(700)},
		{0, "assettest.asset1", uint64(params.AssetFeeType), uint64(1), big.NewInt(100), big.NewInt(100)},
		{3, "assettest.asset1", uint64(params.AssetFeeType), uint64(4), big.NewInt(400), big.NewInt(400)},
		{2, "assettest.asset1", uint64(params.AssetFeeType), uint64(3), big.NewInt(300), big.NewInt(300)},
		{1, "assettest.asset1", uint64(params.AssetFeeType), uint64(2), big.NewInt(500), big.NewInt(700)},
		{0, "testtest.testact1", uint64(params.CoinbaseFeeType), uint64(1), big.NewInt(600), big.NewInt(600)},
	}

	for _, tf := range testFeeInfo {
		objectName := tf.objectName
		objectType := uint64(tf.objectType)
		assetID := tf.assetID
		value := tf.value

		//NewFeeManager
		err := fm.RecordFeeInSystem(objectName, objectType, assetID, value)

		if err != nil {
			t.Errorf("record fee in system failed, err:%v", err)
			return
		}
		fm.accountDB.AddAccountBalanceByID(common.Name(feeConfig.feeName), assetID, value)
	}

	err := addAssetAndAccount()
	if err != nil {
		t.Errorf("add asset and account failed, err:%v", err)
	}

	//withdraw fee from system
	withdrawInfo, err := fm.WithdrawFeeFromSystem(testFeeInfo[0].objectName, testFeeInfo[0].objectType)

	if err != nil || withdrawInfo == nil {
		t.Errorf("withdraw fee from system failed, err:%v", err)
	}

	//check
	objectFee, err := fm.GetObjectFeeByName(testFeeInfo[0].objectName, testFeeInfo[0].objectType)

	if err != nil || objectFee == nil {
		t.Errorf("check withdraw fee from system failed, err:%v", err)
	}

	for _, tf := range testFeeInfo {

		index := tf.assetIndex
		objectName := tf.objectName
		objectType := uint64(tf.objectType)
		assetID := tf.assetID
		totalValue := tf.totalValue

		objectID, err := fm.getObjectFeeIDByName(objectName, objectType)
		if err != nil {
			t.Errorf("get object id failed by id, err:%v", err)
			return
		}

		objectFee, err := fm.GetObjectFeeByID(objectID)

		if err != nil || objectFee == nil {
			t.Errorf("get objectfee failed by id, err:%v", err)
			return
		}

		if tf.objectName == testFeeInfo[0].objectName {
			//check withdraw account balance
			value, err := fm.accountDB.GetAccountBalanceByID(common.Name("testtest.testact1"), tf.assetID, 0)
			if err != nil || value.Cmp(tf.totalValue) != 0 {
				t.Errorf("check account balances failed, name:%v, value:%v, err:%v", tf.objectName, value, err)
			}
		}

		if objectFee.ObjectFeeID != objectID || objectFee.ObjectType != objectType {
			t.Errorf("check object id and type failed")
		}

		/*test asset info*/
		assetFee := objectFee.AssetFees[index]
		if assetFee.AssetID != assetID || assetFee.TotalFee.Cmp(totalValue) != 0 {
			t.Errorf("check asset fee failed, objectName:%v, assetId:%d", objectName, assetID)
		}

		if tf.objectName == testFeeInfo[0].objectName {
			if assetFee.RemainFee.Cmp(big.NewInt(0)) != 0 {
				t.Errorf("check asset remain failed, objectName:%v, assetId:%d", objectName, assetID)
			}
		} else {
			if assetFee.RemainFee.Cmp(totalValue) != 0 {
				t.Errorf("check asset remain failed, objectName:%v, assetId:%d", objectName, assetID)
			}
		}
	}
}

func TestWithdrawFeeFromSystemNotExsit(t *testing.T) {
	type testFee struct {
		assetIndex int
		objectName string
		objectType uint64
		assetID    uint64
		value      *big.Int
		totalValue *big.Int
	}

	testFeeInfo := []*testFee{
		{1, "testnotexsit.tt", uint64(params.ContractFeeType), uint64(2), big.NewInt(200), big.NewInt(700)},
	}

	//withdraw fee from system
	_, err := fm.WithdrawFeeFromSystem(testFeeInfo[0].objectName, testFeeInfo[0].objectType)

	if err == nil {
		t.Errorf("withdraw not exsit fee from system case failed")
	}
}
