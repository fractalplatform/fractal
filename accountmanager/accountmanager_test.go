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

package accountmanager

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var sdb = getStateDB()

var acctm = getAccountManager()
var ast = getAsset()

func getStateDB() *state.StateDB {
	db := fdb.NewMemDatabase()
	tridb := state.NewDatabase(db)
	statedb, err := state.New(common.Hash{}, tridb)
	if err != nil {
		fmt.Printf("test getStateDB() failure %v", err)
		return nil
	}

	return statedb
}
func getAsset() *asset.Asset {
	return asset.NewAsset(sdb)
}
func getAccountManager() *AccountManager {
	SetAcctMangerName("systestname")
	SetSysName("systestname")
	am, err := NewAccountManager(sdb)
	if err != nil {
		fmt.Printf("test getAccountManager() failure %v", err)
	}
	pubkey := new(common.PubKey)
	pubkey.SetBytes([]byte("abcde123456789"))
	am.CreateAccount(common.Name("systestname"), common.Name(""), 0, *pubkey)
	return am
}

func TestSDB(t *testing.T) {

	b, err := rlp.EncodeToBytes("aaaa")
	if err != nil {
		fmt.Printf("encode err = %v", err)
	}
	sdb.Put("test1", acctInfoPrefix, b)
	b1, err := sdb.Get("test1", acctInfoPrefix)
	if err != nil {
		fmt.Printf("err = %v", err)
	}
	if len(b1) == 0 {
		fmt.Printf("len = 0 err")
	}

	var s string
	if err := rlp.DecodeBytes(b, &s); err != nil {
		fmt.Printf("decode err = %v", err)
	}
	if s != "aaaa" {
		fmt.Printf("err str = %v\n", s)
	}

}
func TestNN(t *testing.T) {
	if err := acctm.CreateAccount(common.Name("123asdf2"), common.Name(""), 0, *new(common.PubKey)); err != nil {
		t.Errorf("err create account\n")
	}
	_, err := acctm.GetAccountBalanceByID(common.Name("123asdf2"), 1)
	if err == nil {
		t.Errorf("err get balance err %v\n", err)
	}
}
func TestNewAccountManager(t *testing.T) {
	type args struct {
		db *state.StateDB
	}

	tests := []struct {
		name    string
		args    args
		want    *AccountManager
		wantErr error
	}{

		//
		{"newnilacct", args{nil}, nil, ErrNewAccountErr},
		{"newacct", args{sdb}, acctm, nil},
		//{"newacctErr", args{getStateDB(t)}, nil, true},
	}

	for _, tt := range tests {
		got, err := NewAccountManager(tt.args.db)
		if err != tt.wantErr {
			t.Errorf("%q. NewAccountManager() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. NewAccountManager() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_CreateAccount(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	pubkey := new(common.PubKey)
	pubkey2 := new(common.PubKey)
	pubkey.SetBytes([]byte("abcde123456789"))
	//pubkey2.SetBytes([]byte("abcde123456789"))

	pubkey3, _ := GeneragePubKey()
	type args struct {
		accountName common.Name
		founderName common.Name
		pubkey      common.PubKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"createAccount", fields{sdb, ast}, args{common.Name("111222332a"), common.Name(""), pubkey3}, false},
		{"createAccountWithEmptyKey", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name(""), *pubkey2}, false},
		{"createAccountWithEmptyKey", fields{sdb, ast}, args{common.Name("a123456789aeed"), common.Name(""), *pubkey}, false},
		{"createAccountWithInvalidName", fields{sdb, ast}, args{common.Name("a12345678-aeee"), common.Name(""), *pubkey}, true},
		{"createAccountWithInvalidName", fields{sdb, ast}, args{common.Name("a123456789aeeefgp"), common.Name(""), *pubkey}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.CreateAccount(tt.args.accountName, tt.args.founderName, 1, tt.args.pubkey); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.CreateAccount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}

	am1 := &AccountManager{
		sdb: sdb,
		ast: ast,
	}
	err := am1.CreateAccount(common.Name("aaaadddd"), common.Name("111222332a"), 1, *pubkey)
	if err != nil {
		t.Errorf("create acct err:%v", err)
	}
	ret, _ := am1.AccountIsExist(common.Name("aaaadddd"))
	if ret != true {
		t.Errorf("create acct err")
	}
}

func TestAccountManager_AccountIsExist(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		//
		{"accountExist", fields{sdb, ast}, args{common.Name("111222332a")}, true, false},
		{"accountnotExist", fields{sdb, ast}, args{common.Name("a123456789aeee1")}, false, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.AccountIsExist(tt.args.accountName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.AccountIsExist() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.AccountIsExist() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_AccountIsEmpty(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		//
		{"accountEmpty", fields{sdb, ast}, args{common.Name("11122233")}, false, ErrAccountNotExist},
		{"accountEmpty", fields{sdb, ast}, args{common.Name("a123456789aeee")}, true, nil},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.AccountIsEmpty(tt.args.accountName)
		if err != tt.wantErr {
			t.Errorf("%q. AccountManager.AccountIsEmpty() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.AccountIsEmpty() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
func GeneragePubKey() (common.PubKey, *ecdsa.PrivateKey) {
	prikey, _ := crypto.GenerateKey()
	return common.BytesToPubKey(crypto.FromECDSAPub(&prikey.PublicKey)), prikey
}
func TestAccountManager_UpdateAccount(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	pubkey := new(common.PubKey)
	pubkey2 := new(common.PubKey)
	pubkey.SetBytes([]byte("abcde123456789"))

	pubkey3, _ := GeneragePubKey()
	//fmt.Printf("UpdateAccount key=%v\n", pubkey3.Bytes())
	type args struct {
		accountName common.Name
		founderName common.Name
		chargeRatio uint64
		pubkey      common.PubKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"updateAccount", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, pubkey3}, false},
		{"updateAccount", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("111222332a"), 1, pubkey3}, false},
		{"updateAccountNilFounder", fields{sdb, ast}, args{common.Name("111222332a"), common.Name("1112223-"), 1, *pubkey2}, true},
		{"updateAccountInvaidName", fields{sdb, ast}, args{common.Name("a123456789aeeegty"), common.Name("a123456789aeeegty"), 1, *pubkey2}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		//if err := am.CreateAccount(tt.args.accountName, *pubkey2); (err != nil) != tt.wantErr {
		//	t.Errorf("%q. AccountManager.CreateAccount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		//}
		if err := am.UpdateAccount(tt.args.accountName, tt.args.founderName, tt.args.chargeRatio, tt.args.pubkey); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.UpdateAccount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	acct, _ := acctm.GetAccountByName(common.Name("a123456789aeee"))
	if bytes.Compare(acct.PublicKey.Bytes(), pubkey3.Bytes()) != 0 {
		t.Errorf("AccountManager.UpdateAccount() error key not eque:%v\n =%v", acct.PublicKey.Bytes(), []byte("abcde123456789"))
	}
}

func TestAccountManager_GetAccountByName(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	//pubkey2 := new(common.PubKey)
	//am := &AccountManager{
	//	sdb: getStateDB(t),
	//	ast: asset.NewAsset(getStateDB(t)),
	//}
	//accountmanager.NewAccount(common.Name("a123456789aeee"),)
	//if err := am.CreateAccount(common.Name("a123456789aeee"), *pubkey2); err != nil {
	//	t.Errorf("%q. GetAccountByName.CreateAccount() error = %v, ", common.Name("a123456789aeee"), err)
	//}

	type args struct {
		accountName common.Name
	}
	a, _ := acctm.GetAccountByName(common.Name("a123456789aeee"))

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Account
		wantErr bool
	}{
		//
		{"GetAccountByName Exist", fields{sdb, ast}, args{common.Name("a123456789aeee")}, a, false},
		//{"GetAccountByName NotExist", fields{getStateDB(t), asset.NewAsset(getStateDB(t))}, args{common.Name("a123456789aeeedeg")}, ,false},
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}

		got, err := am.GetAccountByName(tt.args.accountName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAccountByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAccountByName() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_SetAccount(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		acct *Account
	}
	pubkey2 := new(common.PubKey)
	acctm.CreateAccount(common.Name("a123456789"), common.Name(""), 0, *pubkey2)
	ac, _ := acctm.GetAccountByName(common.Name("a123456789"))

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"acctnotexist", fields{sdb, ast}, args{nil}, true},
		{"acctexist", fields{sdb, ast}, args{ac}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.SetAccount(tt.args.acct); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.SetAccount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_SetNonce(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		nonce       uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"setnonce", fields{sdb, ast}, args{common.Name("a123456789"), 9}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.SetNonce(tt.args.accountName, tt.args.nonce); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.SetNonce() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_GetNonce(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		//
		{"getnonce", fields{sdb, ast}, args{common.Name("a123456789")}, 9, false},
		{"notexist", fields{sdb, ast}, args{common.Name("a1234567891")}, 0, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetNonce(tt.args.accountName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetNonce() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.GetNonce() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
func TestAccountManager_DeleteAccountByName(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"delnotexist", fields{sdb, ast}, args{common.Name("a1234567891")}, true},
		{"delexist", fields{sdb, ast}, args{common.Name("a123456789")}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.DeleteAccountByName(tt.args.accountName); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.DeleteAccountByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

//func TestAccountManager_GetBalancesList(t *testing.T) {
//	type fields struct {
//		sdb SdbIf
//		ast *asset.Asset
//	}
//	type args struct {
//		accountName common.Name
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    []*AssetBalance
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		am := &AccountManager{
//			sdb: tt.fields.sdb,
//			ast: tt.fields.ast,
//		}
//		got, err := am.GetBalancesList(tt.args.accountName)
//		if (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.GetBalancesList() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//			continue
//		}
//		if !reflect.DeepEqual(got, tt.want) {
//			t.Errorf("%q. AccountManager.GetBalancesList() = %v, want %v", tt.name, got, tt.want)
//		}
//	}
//}

//func TestAccountManager_GetAllAccountBalance(t *testing.T) {
//	type fields struct {
//		sdb SdbIf
//		ast *asset.Asset
//	}
//	type args struct {
//		accountName common.Name
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    map[uint64]*big.Int
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		am := &AccountManager{
//			sdb: tt.fields.sdb,
//			ast: tt.fields.ast,
//		}
//		got, err := am.GetAllAccountBalance(tt.args.accountName)
//		if (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.GetAllAccountBalance() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//			continue
//		}
//		if !reflect.DeepEqual(got, tt.want) {
//			t.Errorf("%q. AccountManager.GetAllAccountBalance() = %v, want %v", tt.name, got, tt.want)
//		}
//	}
//}

//func TestAccountManager_RecoverTx(t *testing.T) {
//	type fields struct {
//		sdb SdbIf
//		ast *asset.Asset
//	}
//	type args struct {
//		signer types.Signer
//		tx     *types.Transaction
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		am := &AccountManager{
//			sdb: tt.fields.sdb,
//			ast: tt.fields.ast,
//		}
//		if err := am.RecoverTx(tt.args.signer, tt.args.tx); (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.RecoverTx() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//		}
//	}
//}

func TestAccountManager_IsValidSign(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		aType       types.ActionType
		pub         common.PubKey
	}
	pubkey := new(common.PubKey)
	pubkey2 := new(common.PubKey)
	pubkey2.SetBytes([]byte("abcde123456789"))
	acctm.UpdateAccount(common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, *pubkey2)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"invalidsign", fields{sdb, ast}, args{common.Name("a123456789aeee"), types.CreateContract, *pubkey}, true},
		{"validsign", fields{sdb, ast}, args{common.Name("a123456789aeee"), types.CreateContract, *pubkey2}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.IsValidSign(tt.args.accountName, tt.args.aType, tt.args.pub); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.IsValidSign() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_GetAccountBalanceByID(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		assetID     uint64
	}
	//asset ID = 1
	acctm.ast.IssueAsset("ziz", "zz", big.NewInt(1000), 0, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(1000))
	id, _ := acctm.ast.GetAssetIdByName("ziz")
	t.Logf("GetAccountBalanceByID id=%v", id)
	if err := acctm.AddAccountBalanceByID(common.Name("a123456789aeee"), id, big.NewInt(800)); err != nil {
		t.Errorf("%q. GetAccountByName.AddBalanceByName() error = %v, ", common.Name("a123456789aeee"), err)
	}

	val, _ := acctm.GetAccountBalanceByID(common.Name("a123456789aeee"), id)
	t.Logf("a123456789aeee asset id=%v : val=%v\n", id, val)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *big.Int
		wantErr bool
	}{
		//
		{"getAcctByID", fields{sdb, ast}, args{common.Name("a123456789aeee"), id}, big.NewInt(800), false},
	}
	//acctm := getAccountManager(t)
	//pubkey2 := new(common.PubKey)

	//if err := acctm.CreateAccount(common.Name("a123456789aeee"), *pubkey2); err != nil {
	//	t.Errorf("%q. GetAccountByName.CreateAccount() error = %v, ", common.Name("a123456789aeee"), err)
	//}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAccountBalanceByID(tt.args.accountName, tt.args.assetID)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAccountBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAccountBalanceByID() = %v, want %v", tt.name, got, tt.want)
		}
	}
	val, _ = acctm.GetAccountBalanceByID(common.Name("a123456789aeee"), id)
	if val.Cmp(big.NewInt(800)) != 0 {
		t.Errorf("TestAccountManager_GetAccountBalanceByID = %v", val)
	}
}

func TestAccountManager_GetAssetInfoByName(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		name string
	}
	ast1, err := asset.NewAssetObject("ziz", "zz", big.NewInt(1000), 0, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(1000))
	if err != nil {
		t.Errorf("new asset object err")
	}
	ast1.SetAssetId(1)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *asset.AssetObject
		wantErr bool
	}{
		//
		{"assetnotexist", fields{sdb, ast}, args{"ziz1"}, nil, true},
		{"assetexist", fields{sdb, ast}, args{"ziz"}, ast1, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAssetInfoByName(tt.args.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAssetInfoByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAssetInfoByName() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_GetAssetInfoByID(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		assetID uint64
	}

	ast1, err := asset.NewAssetObject("ziz", "zz", big.NewInt(1000), 0, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(1000))
	if err != nil {
		t.Errorf("new asset object err")
	}
	ast1.SetAssetId(1)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *asset.AssetObject
		wantErr bool
	}{
		//
		{"assetnotexist", fields{sdb, ast}, args{0}, nil, true},
		{"asssetexist", fields{sdb, ast}, args{1}, ast1, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAssetInfoByID(tt.args.assetID)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAssetInfoByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAssetInfoByID() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

//func TestAccountManager_GetAccountBalanceByName(t *testing.T) {
//	type fields struct {
//		sdb SdbIf
//		ast *asset.Asset
//	}
//	type args struct {
//		accountName common.Name
//		assetName   string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *big.Int
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		am := &AccountManager{
//			sdb: tt.fields.sdb,
//			ast: tt.fields.ast,
//		}
//		got, err := am.GetAccountBalanceByName(tt.args.accountName, tt.args.assetName)
//		if (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.GetAccountBalanceByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//			continue
//		}
//		if !reflect.DeepEqual(got, tt.want) {
//			t.Errorf("%q. AccountManager.GetAccountBalanceByName() = %v, want %v", tt.name, got, tt.want)
//		}
//	}
//}

func TestAccountManager_GetAssetAmountByTime(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		assetID uint64
		time    uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *big.Int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAssetAmountByTime(tt.args.assetID, tt.args.time)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAssetAmountByTime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAssetAmountByTime() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_GetAccountLastChange(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAccountLastChange(tt.args.accountName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAccountLastChange() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.GetAccountLastChange() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_GetSnapshotTime(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		num  uint64
		time uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetSnapshotTime(tt.args.num, tt.args.time)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetSnapshotTime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.GetSnapshotTime() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_GetBalanceByTime(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		assetID     uint64
		time        uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *big.Int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetBalanceByTime(tt.args.accountName, tt.args.assetID, tt.args.time)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetBalanceByTime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetBalanceByTime() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_GetAssetFounder(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		assetID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    common.Name
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAssetFounder(tt.args.assetID)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAssetFounder() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAssetFounder() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_GetChargeRatio(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetChargeRatio(tt.args.accountName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetChargeRatio() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.GetChargeRatio() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_GetAssetChargeRatio(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		assetID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAssetChargeRatio(tt.args.assetID)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAssetChargeRatio() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.GetAssetChargeRatio() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_SubAccountBalanceByID(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		assetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"subAcctByID", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(200)}, false},
		{"subAcctByID", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(700)}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.SubAccountBalanceByID(tt.args.accountName, tt.args.assetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.SubAccountBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_AddAccountBalanceByID(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		assetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"subAcctByID", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(200)}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.AddAccountBalanceByID(tt.args.accountName, tt.args.assetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.AddAccountBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_AddAccountBalanceByName(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		assetName   string
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"addAcctByname", fields{sdb, ast}, args{common.Name("a123456789aeee"), "ziz", big.NewInt(200)}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.AddAccountBalanceByName(tt.args.accountName, tt.args.assetName, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.AddAccountBalanceByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_EnoughAccountBalance(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		assetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//amount = 1000
		{"enough", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(-2)}, true},
		{"enough", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(999)}, false},
		{"notenough", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(1001)}, true},
	}

	//val, _ := acctm.GetAccountBalanceByID(common.Name("a123456789aeee"), 1)
	//t.Logf("EnoughAccountBalance asset id=%v : val=%v\n", 1, val)

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.EnoughAccountBalance(tt.args.accountName, tt.args.assetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.EnoughAccountBalance() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	val, _ := acctm.GetAccountBalanceByID(common.Name("a123456789aeee"), 1)
	if val.Cmp(big.NewInt(1000)) != 0 {
		t.Logf("TestAccountManager_EnoughAccountBalance = %v", val)
	}
}

func TestAccountManager_GetCode(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	pubkey2 := new(common.PubKey)
	acct, _ := acctm.GetAccountByName(common.Name("a123456789aeee"))
	acctm.CreateAccount(common.Name("a123456789aeed"), common.Name("a123456789aeed"), 0, *pubkey2)
	acct.SetCode([]byte("abcde123456789"))
	acctm.SetAccount(acct)
	//t.Logf("EnoughAccountBalance asset id=%v : val=%v\n", 1, val)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		//
		{"haveCode", fields{sdb, ast}, args{common.Name("a123456789aeee")}, []byte("abcde123456789"), false},
		{"noCode", fields{sdb, ast}, args{common.Name("a123456789aeed")}, nil, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetCode(tt.args.accountName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetCode() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetCode() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

//func TestAccountManager_SetCode(t *testing.T) {
//	type fields struct {
//		sdb SdbIf
//		ast *asset.Asset
//	}
//	type args struct {
//		accountName common.Name
//		code        []byte
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    bool
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		am := &AccountManager{
//			sdb: tt.fields.sdb,
//			ast: tt.fields.ast,
//		}
//		got, err := am.SetCode(tt.args.accountName, tt.args.code)
//		if (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.SetCode() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//			continue
//		}
//		if got != tt.want {
//			t.Errorf("%q. AccountManager.SetCode() = %v, want %v", tt.name, got, tt.want)
//		}
//	}
//}

func TestAccountManager_GetCodeSize(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		//
		{"newacct", fields{sdb, ast}, args{common.Name("a123456789aeee")}, uint64(len([]byte("abcde123456789"))), false},
		{"newacct", fields{sdb, ast}, args{common.Name("a123456789aeed")}, 0, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetCodeSize(tt.args.accountName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetCodeSize() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.GetCodeSize() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

//func TestAccountManager_GetCodeHash(t *testing.T) {
//	type fields struct {
//		sdb SdbIf
//		ast *asset.Asset
//	}
//	type args struct {
//		accountName common.Name
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    common.Hash
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		am := &AccountManager{
//			sdb: tt.fields.sdb,
//			ast: tt.fields.ast,
//		}
//		got, err := am.GetCodeHash(tt.args.accountName)
//		if (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.GetCodeHash() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//			continue
//		}
//		if !reflect.DeepEqual(got, tt.want) {
//			t.Errorf("%q. AccountManager.GetCodeHash() = %v, want %v", tt.name, got, tt.want)
//		}
//	}
//}

func TestAccountManager_CanTransfer(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		assetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		//
		{"cantranfer", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(3)}, true, false},
		{"can'ttranfer", fields{sdb, ast}, args{common.Name("a123456789aeee"), 1, big.NewInt(30000)}, false, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.CanTransfer(tt.args.accountName, tt.args.assetID, tt.args.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.CanTransfer() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. AccountManager.CanTransfer() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccountManager_TransferAsset(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		fromAccount common.Name
		toAccount   common.Name
		assetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"tranferok", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeed"), 1, big.NewInt(3)}, false},
	}
	val, err := acctm.GetAccountBalanceByID(common.Name("a123456789aeee"), 1)
	if err != nil {
		t.Error("TransferAsset GetAccountBalanceByID err")
	}
	if val.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("TransferAsset GetAccountBalanceByID val=%v", val)
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.TransferAsset(tt.args.fromAccount, tt.args.toAccount, tt.args.assetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.TransferAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	val1, err := acctm.GetAccountBalanceByID(common.Name("a123456789aeee"), 1)
	if err != nil {
		t.Error("TransferAsset GetAccountBalanceByID err")
	}
	if val1.Cmp(big.NewInt(997)) != 0 {
		t.Errorf("TransferAsset1 GetAccountBalanceByID val=%v", val1)
	}

}

func TestAccountManager_IssueAsset(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		asset *asset.AssetObject
	}

	//am := getAccountManager(t)
	//
	//err := am.ast.IssueAsset("ziz0123456789ziz1", "ziz", big.NewInt(2), 18, common.Name("a0123456789ziz"))
	//if err != nil {
	//	t.Fatal("IssueAsset err", err)
	//}

	ast1, err := asset.NewAssetObject("ziz0123456789ziz", "ziz", big.NewInt(2), 18, common.Name("a123456789aeee"), common.Name("a0123456789ziz"), big.NewInt(100000))
	if err != nil {
		t.Fatal("IssueAsset err", err)
	}

	ast3, err := asset.NewAssetObject("ziz0123456789", "ziz", big.NewInt(2), 18, common.Name("a123456777777"), common.Name("a123456789aeee"), big.NewInt(2))
	if err != nil {
		t.Fatal("IssueAsset err", err)
	}
	//asset id =2
	ast2, err := asset.NewAssetObject("ziz0123456789zi", "ziz", big.NewInt(2), 18, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(12000))
	if err != nil {
		t.Fatal("IssueAsset err", err)
	}

	//asset,err := am.ast.GetAssetObjectByName("ziz")
	//if err != nil {
	//	t.Fatal("GetAssetObjectByName err", err)
	//}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"ownernotexist", fields{sdb, ast}, args{ast1}, true},
		{"foundernotexist", fields{sdb, ast}, args{ast3}, true},
		{"ownerexist", fields{sdb, ast}, args{ast2}, false},
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.IssueAsset(tt.args.asset); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.IssueAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_IncAsset2Acct(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		fromName common.Name
		toName   common.Name
		assetID  uint64
		amount   *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"over upperlimit", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeee"), 2, big.NewInt(11999)}, true},
		{"accountexist", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeee"), 2, big.NewInt(10)}, false},
		{"notexist", fields{sdb, ast}, args{common.Name("a0123456789ziz"), common.Name("a123456789aeef"), 2, big.NewInt(1)}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.IncAsset2Acct(tt.args.fromName, tt.args.toName, tt.args.assetID, tt.args.amount); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.IncAsset2Acct() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

//func TestAccountManager_AddBalanceByName(t *testing.T) {
//	type fields struct {
//		sdb SdbIf
//		ast *asset.Asset
//	}
//	type args struct {
//		accountName common.Name
//		assetID     uint64
//		amount      *big.Int
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		//
//	}
//	for _, tt := range tests {
//		am := &AccountManager{
//			sdb: tt.fields.sdb,
//			ast: tt.fields.ast,
//		}
//		if err := am.AddAccountBalanceByID(tt.args.accountName, tt.args.assetID, tt.args.amount); (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.AddBalanceByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//		}
//	}
//}

func TestAccountManager_Process(t *testing.T) {
	type fields struct {
		sdb SdbIf
		ast *asset.Asset
	}
	type args struct {
		action *types.Action
	}

	inc := &IncAsset{
		AssetId: 2,
		Amount:  big.NewInt(100),
		To:      common.Name("a123456789aeee"),
	}
	payload2, err := rlp.EncodeToBytes(inc)
	if err != nil {
		panic("rlp payload err")
	}

	ast0 := &asset.AssetObject{
		AssetId:    2,
		AssetName:  "abced99",
		Symbol:     "aaa",
		Amount:     big.NewInt(100000000),
		Decimals:   2,
		Founder:    common.Name("a123456789aeee"),
		Owner:      common.Name("a123456789aeee"),
		AddIssue:   big.NewInt(0),
		UpperLimit: big.NewInt(1000000000),
	}
	ast1 := &asset.AssetObject{
		AssetId:    2,
		AssetName:  "abced99",
		Symbol:     "aaa",
		Amount:     big.NewInt(100000000),
		Decimals:   2,
		Founder:    common.Name(sysName),
		Owner:      common.Name(sysName),
		AddIssue:   big.NewInt(0),
		UpperLimit: big.NewInt(1000000000),
	}
	payload, err := rlp.EncodeToBytes(ast0)
	if err != nil {
		panic("rlp payload err")
	}
	payload1, err := rlp.EncodeToBytes(ast1)
	if err != nil {
		panic("rlp payload err")
	}
	pubkey, _ := GeneragePubKey()
	pubkey1, _ := GeneragePubKey()
	aa := &AccountAction{
		AccountName: common.Name("a123456789addd"),
		Founder:     common.Name(""),
		ChargeRatio: 10,
		PublicKey:   pubkey,
	}
	payload3, err := rlp.EncodeToBytes(aa)
	if err != nil {
		panic("rlp payload err")
	}
	aa1 := &AccountAction{
		AccountName: common.Name("a123456789addd"),
		Founder:     common.Name(""),
		ChargeRatio: 99,
		PublicKey:   pubkey1,
	}
	payload4, err := rlp.EncodeToBytes(aa1)
	if err != nil {
		panic("rlp payload err")
	}

	action := types.NewAction(types.IssueAsset, common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, 1, 0, big.NewInt(0), payload)
	action1 := types.NewAction(types.IncreaseAsset, common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, 1, 2, big.NewInt(0), payload2)
	action2 := types.NewAction(types.UpdateAsset, common.Name("a123456789aeee"), common.Name("a123456789addd"), 1, 1, 2, big.NewInt(0), payload1)
	action3 := types.NewAction(types.CreateAccount, common.Name("a123456789aeee"), common.Name(sysName), 1, 1, 2, big.NewInt(10), payload3)
	action4 := types.NewAction(types.UpdateAccount, common.Name("a123456789addd"), common.Name("a123456789addd"), 1, 1, 2, big.NewInt(0), payload4)
	//action5 := types.NewAction(types.DeleteAccount, common.Name("123asdf2"), common.Name("123asdf2"), 1, 1, 2, big.NewInt(0), pubkey1[:])
	//action6 := types.NewAction(types.Transfer, common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, 1, 2, big.NewInt(1), pubkey1[:])
	//action7 := types.NewAction(types.Transfer, common.Name("a123456789addd"), common.Name("a123456789aeee"), 1, 1, 2, big.NewInt(1), payload)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"issue", fields{sdb, ast}, args{action}, false},
		{"increase", fields{sdb, ast}, args{action1}, false},
		{"createaccount", fields{sdb, ast}, args{action3}, false},
		{"updateasset", fields{sdb, ast}, args{action2}, false},
		{"updateaccount", fields{sdb, ast}, args{action4}, false},
		//{"deleteaccount", fields{sdb, ast}, args{action5}, false},
		//{"transfer2self", fields{sdb, ast}, args{action6}, false},
		//{"transfer", fields{sdb, ast}, args{action7}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.Process(tt.args.action); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.Process() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}

	asset2, err := acctm.GetAssetInfoByName("abced99")
	if asset2 == nil {
		t.Error("Process issue asset failure")
	}
	//t.Logf("issue ok id=%v", asset2.AssetId)
	if asset2.Amount.Cmp(big.NewInt(100000000)) != 0 {
		t.Errorf("Process increase asset failure amount=%v", asset2.Amount)
	}
	if asset2.GetAssetOwner() != "a123456789aeee" {
		t.Errorf("Process set asset owner failure =%v", asset2.GetAssetName())
	}
	val, err := acctm.GetAccountBalanceByID(common.Name("a123456789aeee"), 1)
	if err != nil {
		t.Error("Process GetAccountBalanceByID err")
	}
	//t.Logf("Process GetAccountBalanceByID val=%v", val)

	ac, err := acctm.GetAccountByName(common.Name("123asdf2"))
	if err != nil {
		t.Error("Process GetAccountByName err")
	}
	if !ac.IsDestroyed() {
		//t.Error("Process delete account failure")
	}

	ac1, err := acctm.GetAccountByName(common.Name("a123456789addd"))
	if err != nil {
		t.Error("Process GetAccountByName err")
	}
	if ac1 == nil {
		t.Error("Process create account err")
	}
	if bytes.Compare(ac1.PublicKey.Bytes(), pubkey1[:]) != 0 {
		t.Error("Process update account failure")
	}
	val, err = ac1.GetBalanceByID(1)
	if val.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Process transfer  failure=%v", val)
	}

}
