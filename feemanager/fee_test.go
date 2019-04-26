package feemanager

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	memdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

var sdb = getStateDB()
var acctm = getAccountManager()
var fm = getFeeManager()
var ast = getAsset()

func getAsset() *asset.Asset {
	return asset.NewAsset(sdb)
}

func getStateDB() *state.StateDB {
	db := memdb.NewMemDatabase()
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
	accountmanager.SetSysName("systestname")
	am, err := accountmanager.NewAccountManager(sdb)
	if err != nil {
		fmt.Printf("test getAccountManager() failure %v", err)
	}
	pubkey := new(common.PubKey)
	pubkey.SetBytes([]byte("abcde123456789"))
	am.CreateAccount(common.Name("systestname"), common.Name(""), 0, 0, *pubkey)
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
		{1, "test.tt", uint64(common.ContractName), uint64(2), big.NewInt(200), big.NewInt(700)},
		{0, "test.tt", uint64(common.ContractName), uint64(1), big.NewInt(100), big.NewInt(100)},
		{3, "test.tt", uint64(common.ContractName), uint64(4), big.NewInt(400), big.NewInt(400)},
		{2, "test.tt", uint64(common.ContractName), uint64(3), big.NewInt(300), big.NewInt(300)},
		{1, "test.tt", uint64(common.ContractName), uint64(2), big.NewInt(500), big.NewInt(700)},
		{0, "test.tt1", uint64(common.AssetName), uint64(1), big.NewInt(600), big.NewInt(600)},
	}

	for _, tf := range testFeeInfo {
		objectName := common.Name(tf.objectName)
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
		objectName := common.Name(tf.objectName)
		objectType := uint64(tf.objectType)
		assetID := tf.assetID
		totalValue := tf.totalValue

		objectID, err := fm.getObjectFeeIDByName(objectName)
		if err != nil {
			t.Errorf("get object id failed by id, err:%v", err)
			return
		}

		objectFee, err := fm.getObjectFeeByID(objectID)

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
		tname  = common.Name("testact1")
		pubKey = new(common.PubKey)
	)

	tests := []args{
		// TODO: Add test cases.
		{"asset1", "s1", big.NewInt(0), 2, tname, tname},
		{"asset2", "s2", big.NewInt(0), 2, tname, tname},
		{"asset3", "s3", big.NewInt(0), 2, tname, tname},
		{"asset4", "s4", big.NewInt(0), 2, tname, tname},
	}

	if err := acctm.CreateAccount(tname, tname, 0, 0, *pubKey); err != nil {
		return err
	}

	for _, tt := range tests {
		err := ast.IssueAsset(tt.assetName, 0, tt.symbol, tt.amount, tt.dec, tt.founder, tt.owner, big.NewInt(9999999999), common.Name(""))
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
		{1, "asset1", uint64(common.AssetName), uint64(2), big.NewInt(200), big.NewInt(700)},
		{0, "asset1", uint64(common.AssetName), uint64(1), big.NewInt(100), big.NewInt(100)},
		{3, "asset1", uint64(common.AssetName), uint64(4), big.NewInt(400), big.NewInt(400)},
		{2, "asset1", uint64(common.AssetName), uint64(3), big.NewInt(300), big.NewInt(300)},
		{1, "asset1", uint64(common.AssetName), uint64(2), big.NewInt(500), big.NewInt(700)},
		{0, "testact1", uint64(common.CoinbaseName), uint64(1), big.NewInt(600), big.NewInt(600)},
	}

	for _, tf := range testFeeInfo {
		objectName := common.Name(tf.objectName)
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

	err := addAssetAndAccount()
	if err != nil {
		t.Errorf("add asset and account failed, err:%v", err)
	}

	//withdraw fee from system
	withdrawInfo, err := fm.WithdrawFeeFromSystem(common.Name(testFeeInfo[0].objectName))

	if err != nil || len(withdrawInfo) == 0 {
		t.Errorf("withdraw fee from system failed, err:%v", err)
	}

	//check
	objectFee, err := fm.getObjectFeeByName(common.Name(testFeeInfo[0].objectName))

	if err != nil || len(objectFee.AssetFees) != 0 {
		t.Errorf("check withdraw fee from system failed, err:%v", err)
	}

	for _, tf := range testFeeInfo {
		if tf.objectName == testFeeInfo[0].objectName {
			//check account balance
			value, err := fm.accountDB.GetAccountBalanceByID(common.Name("testact1"), tf.assetID, 0)
			if err != nil || value.Cmp(tf.totalValue) != 0 {
				t.Errorf("check account balances failed, name:%v, value:%v, err:%v", tf.objectName, value, err)
			}
			continue
		}

		index := tf.assetIndex
		objectName := common.Name(tf.objectName)
		objectType := uint64(tf.objectType)
		assetID := tf.assetID
		totalValue := tf.totalValue

		objectID, err := fm.getObjectFeeIDByName(objectName)
		if err != nil {
			t.Errorf("get object id failed by id, err:%v", err)
			return
		}

		objectFee, err := fm.getObjectFeeByID(objectID)

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
