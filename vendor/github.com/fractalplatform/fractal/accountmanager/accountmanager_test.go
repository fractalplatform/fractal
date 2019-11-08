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
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var sdb = getStateDB()
var accountManager = getAccountManager()
var ast = getAsset()
var sysName = "fractal.account"
var blockNumber = uint64(0)

func getStateDB() *state.StateDB {
	db := rawdb.NewMemoryDatabase()
	trieDB := state.NewDatabase(db)
	stateDB, err := state.New(common.Hash{}, trieDB)
	if err != nil {
		fmt.Printf("test getStateDB() failure %v", err)
		return nil
	}

	return stateDB
}
func getAsset() *asset.Asset {
	return asset.NewAsset(sdb)
}
func getAccountManager() *AccountManager {
	SetAcctMangerName("systestname")
	//SetSysName("systestname")
	am, err := NewAccountManager(sdb)
	if err != nil {
		fmt.Printf("test getAccountManager() failure %v", err)
	}
	pubkey := new(common.PubKey)
	pubkey.SetBytes([]byte("abcde123456789"))
	am.CreateAccount(common.Name("fractal.founder"), common.Name("systestname"), common.Name(""), 0, 0, *pubkey, "")
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
	if err := accountManager.CreateAccount(common.Name("fractal.founder"), common.Name("a123asdf2"), common.Name(""), 0, 0, *new(common.PubKey), ""); err != nil {
		t.Errorf("err create account\n")
	}
	_, err := accountManager.GetAccountBalanceByID(common.Name("a123asdf2"), 1, 0)
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
		{"newacct", args{sdb}, accountManager, nil},
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

func TestAccountManager_InitAccountCounter(t *testing.T) {
	//TODO
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
		fromName    common.Name
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
		{"createAccount", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("a111222332a"), common.Name(""), pubkey3}, false},
		{"createAccountWithEmptyKey", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("a123456789aeee"), common.Name(""), *pubkey2}, false},
		{"createAccountWithEmptyKey", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("a123456789aeed"), common.Name(""), *pubkey}, false},
		{"createAccountWithInvalidName", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("a12345678-aeee"), common.Name(""), *pubkey}, true},
		{"createAccountWithInvalidName", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("a123456789aeeefgp"), common.Name(""), *pubkey}, true},
		{"creategensisAccount", fields{sdb, ast}, args{common.Name("fractal"), common.Name("fractal.account"), common.Name(""), *pubkey}, false},
		{"creategensisAccount1", fields{sdb, ast}, args{common.Name("fractal"), common.Name("fractal.asset"), common.Name(""), *pubkey}, false},
		{"createinvalidAccount0", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("\ttesttestf1"), common.Name(""), *pubkey}, true},
		{"createinvalidAccount1", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("testtestf1.."), common.Name(""), *pubkey}, true},
		{"createinvalidAccount2", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("fractal.account"), common.Name(""), *pubkey}, true},
		{"createinvalidAccount3", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("fractal.founder.1"), common.Name(""), *pubkey}, true},
		{"createinvalidAccount4", fields{sdb, ast}, args{common.Name("fractal.founder"), common.Name("fractal.founder.12"), common.Name(""), *pubkey}, true},
		{"createinvalidAccount5", fields{sdb, ast}, args{common.Name("fractal"), common.Name("fractal.founder1234"), common.Name(""), *pubkey}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.CreateAccount(tt.args.fromName, tt.args.accountName, tt.args.founderName, 0, 0, tt.args.pubkey, ""); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.CreateAccount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	am1 := &AccountManager{
		sdb: sdb,
		ast: ast,
	}
	err := am1.CreateAccount(common.Name("fractal.founder"), common.Name("aaaadddd"), common.Name("a111222332a"), 0, 0, *pubkey, "")
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
		sdb *state.StateDB
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
		{"accountExist", fields{sdb, ast}, args{common.Name("a111222332a")}, true, false},
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
		sdb *state.StateDB
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
		{"accountNotEmpty", fields{sdb, ast}, args{common.Name("11122233")}, false, ErrAccountNotExist},
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
	privateKey, _ := crypto.GenerateKey()
	return common.BytesToPubKey(crypto.FromECDSAPub(&privateKey.PublicKey)), privateKey
}

func TestAccountManager_GetAccountByName(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
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
	a, _ := accountManager.GetAccountByName(common.Name("a123456789aeee"))

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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		acct *Account
	}
	pubkey2 := new(common.PubKey)
	accountManager.CreateAccount(common.Name("fractal.founder"), common.Name("a123456789"), common.Name(""), 0, 0, *pubkey2, "")
	ac, _ := accountManager.GetAccountByName(common.Name("a123456789"))

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
		sdb *state.StateDB
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
		sdb *state.StateDB
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
		sdb *state.StateDB
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		pub         common.PubKey
	}
	pubkey := new(common.PubKey)
	pubkey2 := new(common.PubKey)
	pubkey2.SetBytes([]byte("abcde123456789"))
	//accountManager.UpdateAccount(common.Name("a123456789aeee"), &AccountAction{common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, *pubkey2})
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"invalidsign", fields{sdb, ast}, args{common.Name("a123456789aeee"), *pubkey}, false},
		{"validsign", fields{sdb, ast}, args{common.Name("a123456789aeee"), *pubkey2}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.IsValidSign(tt.args.accountName, tt.args.pub); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.IsValidSign() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_GetAccountBalanceByID(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		AssetID     uint64
	}
	//asset ID = 0
	accountManager.ast.IssueAsset("ziz", blockNumber, 0, "zz", big.NewInt(1000), 0, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(1000), common.Name(""), "")
	id, err := accountManager.ast.GetAssetIDByName("ziz")
	if err != nil {
		t.Errorf("GetAssetIDByName ziz err")
	}
	t.Logf("GetAccountBalanceByID asset id=%v", id)
	if err := accountManager.AddAccountBalanceByID(common.Name("a123456789aeee"), id, big.NewInt(800)); err != nil {
		t.Errorf("%q. GetAccountByName.AddBalanceByName() error = %v, ", common.Name("a123456789aeee"), err)
	}

	val, _ := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), id, 0)
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
	//accountManager := getAccountManager(t)
	//pubkey2 := new(common.PubKey)

	//if err := accountManager.CreateAccount(common.Name("a123456789aeee"), *pubkey2); err != nil {
	//	t.Errorf("%q. GetAccountByName.CreateAccount() error = %v, ", common.Name("a123456789aeee"), err)
	//}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAccountBalanceByID(tt.args.accountName, tt.args.AssetID, 0)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAccountBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAccountBalanceByID() = %v, want %v", tt.name, got, tt.want)
		}
	}
	val, _ = accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), id, 0)
	if val.Cmp(big.NewInt(800)) != 0 {
		t.Errorf("TestAccountManager_GetAccountBalanceByID = %v", val)
	}
}

func TestAccountManager_GetAssetInfoByName(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		name string
	}
	ast1, err := asset.NewAssetObject("ziz", 0, "zz", big.NewInt(1000), 0, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(1000), common.Name(""), "")
	if err != nil {
		t.Errorf("new asset object err")
	}
	ast1.SetAssetID(0)
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		AssetID uint64
	}

	ast1, err := asset.NewAssetObject("ziz", 0, "zz", big.NewInt(1000), 0, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(1000), common.Name(""), "")
	if err != nil {
		t.Errorf("new asset object err")
	}
	ast1.SetAssetID(0)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *asset.AssetObject
		wantErr bool
	}{
		//
		//{"assetnotexist", fields{sdb, ast}, args{0}, nil, false},
		{"asssetexist", fields{sdb, ast}, args{0}, ast1, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.GetAssetInfoByID(tt.args.AssetID)
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		AssetID uint64
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
		got, err := am.GetAssetAmountByTime(tt.args.AssetID, tt.args.time)
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
		sdb *state.StateDB
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
		sdb *state.StateDB
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		AssetID     uint64
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
		got, err := am.GetBalanceByTime(tt.args.accountName, tt.args.AssetID, 0, tt.args.time)
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		AssetID uint64
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
		got, err := am.GetAssetFounder(tt.args.AssetID)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.GetAssetFounder() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AccountManager.GetAssetFounder() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// func TestAccountManager_GetChargeRatio(t *testing.T) {
// 	type fields struct {
// 		sdb *state.StateDB
// 		ast *asset.Asset
// 	}
// 	type args struct {
// 		accountName common.Name
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    uint64
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		am := &AccountManager{
// 			sdb: tt.fields.sdb,
// 			ast: tt.fields.ast,
// 		}
// 		got, err := am.GetChargeRatio(tt.args.accountName)
// 		if (err != nil) != tt.wantErr {
// 			t.Errorf("%q. AccountManager.GetChargeRatio() error = %v, wantErr %v", tt.name, err, tt.wantErr)
// 			continue
// 		}
// 		if got != tt.want {
// 			t.Errorf("%q. AccountManager.GetChargeRatio() = %v, want %v", tt.name, got, tt.want)
// 		}
// 	}
// }

// func TestAccountManager_GetAssetChargeRatio(t *testing.T) {
// 	type fields struct {
// 		sdb *state.StateDB
// 		ast *asset.Asset
// 	}
// 	type args struct {
// 		AssetID uint64
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    uint64
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		am := &AccountManager{
// 			sdb: tt.fields.sdb,
// 			ast: tt.fields.ast,
// 		}
// 		got, err := am.GetAssetChargeRatio(tt.args.AssetID)
// 		if (err != nil) != tt.wantErr {
// 			t.Errorf("%q. AccountManager.GetAssetChargeRatio() error = %v, wantErr %v", tt.name, err, tt.wantErr)
// 			continue
// 		}
// 		if got != tt.want {
// 			t.Errorf("%q. AccountManager.GetAssetChargeRatio() = %v, want %v", tt.name, got, tt.want)
// 		}
// 	}
// }

func TestAccountManager_SubAccountBalanceByID(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		AssetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"subAcctByID", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(200)}, false},
		{"subAcctByID", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(700)}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.SubAccountBalanceByID(tt.args.accountName, tt.args.AssetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.SubAccountBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_AddAccountBalanceByID(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		AssetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"subAcctByID", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(200)}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.AddAccountBalanceByID(tt.args.accountName, tt.args.AssetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.AddAccountBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_AddAccountBalanceByName(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
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
		{"addAcctByName", fields{sdb, ast}, args{common.Name("a123456789aeee"), "ziz", big.NewInt(200)}, false},
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		AssetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//amount = 1000
		{"enough", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(-2)}, true},
		{"enough2", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(999)}, false},
		{"notenough", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(1001)}, true},
	}

	//val, _ := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), 1)
	//t.Logf("EnoughAccountBalance asset id=%v : val=%v\n", 1, val)

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.EnoughAccountBalance(tt.args.accountName, tt.args.AssetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.EnoughAccountBalance() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	val, _ := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), 0, 0)
	if val.Cmp(big.NewInt(1000)) != 0 {
		t.Logf("TestAccountManager_EnoughAccountBalance = %v", val)
	}
}

func TestAccountManager_GetCode(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
	}
	pubkey2 := new(common.PubKey)
	acct, _ := accountManager.GetAccountByName(common.Name("a123456789aeee"))
	accountManager.CreateAccount(common.Name("fractal.founder"), common.Name("a123456789aeed"), common.Name("a123456789aeed"), 0, 0, *pubkey2, "")
	acct.SetCode([]byte("abcde123456789"))
	accountManager.SetAccount(acct)
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
		sdb *state.StateDB
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		accountName common.Name
		AssetID     uint64
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
		{"cantranfer", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(3)}, true, false},
		{"can'ttranfer", fields{sdb, ast}, args{common.Name("a123456789aeee"), 0, big.NewInt(30000)}, false, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		got, err := am.CanTransfer(tt.args.accountName, tt.args.AssetID, tt.args.value)
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
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		fromAccount common.Name
		toAccount   common.Name
		AssetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"tranferOk", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeed"), 0, big.NewInt(3)}, false},
		{"tranferFail", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeed"), 0, big.NewInt(-3)}, true},
	}
	val, err := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), 0, 0)
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
		if err := am.TransferAsset(tt.args.fromAccount, tt.args.toAccount, tt.args.AssetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.TransferAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	val1, err := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), 0, 0)
	if err != nil {
		t.Error("TransferAsset GetAccountBalanceByID err")
	}
	if val1.Cmp(big.NewInt(997)) != 0 {
		t.Errorf("TransferAsset1 GetAccountBalanceByID val=%v", val1)
	}

}

func TestAccountManager_IssueAsset(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
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

	ast1, err := asset.NewAssetObject("ziz0123456789ziz", 0, "ziz", big.NewInt(2), 18, common.Name("a123456789aeee"), common.Name("a0123456789ziz"), big.NewInt(100000), common.Name(""), "")
	if err != nil {
		t.Fatal("IssueAsset err", err)
	}

	ast3, err := asset.NewAssetObject("ziz0123456789", 0, "ziz", big.NewInt(2), 18, common.Name("a123456777777"), common.Name("a123456789aeee"), big.NewInt(2), common.Name(""), "")
	if err != nil {
		t.Fatal("IssueAsset err", err)
	}
	//asset id =2
	ast2, err := asset.NewAssetObject("ziz0123456789zi", 0, "ziz", big.NewInt(2), 18, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(12000), common.Name(""), "")
	if err != nil {
		t.Fatal("IssueAsset err", err)
	}

	ast4, err := asset.NewAssetObject("ziz0123456789zi", 0, "ziz", big.NewInt(2), 18, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(12000), common.Name(""), "this")
	if err != nil {
		t.Fatal("IssueAsset err", err)
	}

	detail := "aaa"
	for i := 0; i < 200; i++ {
		detail += "aaa"
	}

	_, err = asset.NewAssetObject("ziz0123456789zi", 0, "ziz", big.NewInt(2), 18, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(12000), common.Name(""), detail)
	if err != asset.ErrDetailTooLong {
		t.Fatal("IssueAsset err", err)
	}

	//asset,err := am.ast.GetAssetObjectByName("ziz")
	//if err != nil {
	//	t.Fatal("GetAssetObjectByName err", err)
	//}

	tests := []struct {
		name    string
		from    common.Name
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"ownernotexist", common.Name(""), fields{sdb, ast}, args{ast1}, true},
		{"foundernotexist", common.Name(""), fields{sdb, ast}, args{ast3}, true},
		{"ownerexist", common.Name(""), fields{sdb, ast}, args{ast2}, false},
		{"detaillegal", common.Name(""), fields{sdb, ast}, args{ast4}, true},
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}

		var asset = IssueAsset{
			AssetName:   tt.args.asset.GetAssetName(),
			Symbol:      tt.args.asset.GetSymbol(),
			Amount:      tt.args.asset.GetAssetAmount(),
			Decimals:    tt.args.asset.GetDecimals(),
			Founder:     tt.args.asset.GetAssetFounder(),
			Owner:       tt.args.asset.GetAssetOwner(),
			UpperLimit:  tt.args.asset.GetUpperLimit(),
			Contract:    tt.args.asset.GetContract(),
			Description: tt.args.asset.GetAssetDescription(),
		}
		if _, err := am.IssueAsset(tt.from, asset, blockNumber, 0); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.IssueAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccountManager_IncAsset2Acct(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		fromName common.Name
		toName   common.Name
		AssetID  uint64
		amount   *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"over upperlimit", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, big.NewInt(11999)}, true},
		{"accountexist", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeee"), 1, big.NewInt(10)}, false},
		{"notexist", fields{sdb, ast}, args{common.Name("a0123456789ziz"), common.Name("a123456789aeef"), 1, big.NewInt(1)}, true},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.IncAsset2Acct(tt.args.fromName, tt.args.toName, tt.args.AssetID, tt.args.amount); (err != nil) != tt.wantErr {
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
//		AssetID     uint64
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
//		if err := am.AddAccountBalanceByID(tt.args.accountName, tt.args.AssetID, tt.args.amount); (err != nil) != tt.wantErr {
//			t.Errorf("%q. AccountManager.AddBalanceByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
//		}
//	}
//}

func TestAccountManager_Process(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		action *types.Action
	}

	inc := &IncAsset{
		AssetID: 1,
		Amount:  big.NewInt(100),
		To:      common.Name("a123456789aeee"),
	}
	payload2, err := rlp.EncodeToBytes(inc)
	if err != nil {
		panic("rlp payload err")
	}

	ast0 := &asset.AssetObject{
		AssetID:    1,
		AssetName:  "abced99",
		Symbol:     "aaa",
		Amount:     big.NewInt(100000000),
		Decimals:   2,
		Founder:    common.Name("a123456789aeee"),
		Owner:      common.Name("a123456789aeee"),
		AddIssue:   big.NewInt(0),
		UpperLimit: big.NewInt(1000000000),
	}
	var asset0 = IssueAsset{
		AssetName:   ast0.GetAssetName(),
		Symbol:      ast0.GetSymbol(),
		Amount:      ast0.GetAssetAmount(),
		Decimals:    ast0.GetDecimals(),
		Founder:     ast0.GetAssetFounder(),
		Owner:       ast0.GetAssetOwner(),
		UpperLimit:  ast0.GetUpperLimit(),
		Contract:    ast0.GetContract(),
		Description: ast0.GetAssetDescription(),
	}

	ast1 := &asset.AssetObject{
		AssetID:    1,
		AssetName:  "abced99",
		Symbol:     "aaa",
		Amount:     big.NewInt(100000000),
		Decimals:   2,
		Founder:    common.Name(sysName),
		Owner:      common.Name(sysName),
		AddIssue:   big.NewInt(0),
		UpperLimit: big.NewInt(1000000000),
	}
	payload, err := rlp.EncodeToBytes(asset0)
	if err != nil {
		panic("rlp payload err")
	}

	var asset1 = UpdateAsset{
		AssetID: ast1.GetAssetID(),
		Founder: ast1.GetAssetFounder(),
	}
	payload1, err := rlp.EncodeToBytes(asset1)
	if err != nil {
		panic("rlp payload err")
	}
	pubkey, _ := GeneragePubKey()
	// pubkey1, _ := GeneragePubKey()
	aa := &CreateAccountAction{
		AccountName: common.Name("a123456789addd"),
		Founder:     common.Name(""),
		//ChargeRatio: 10,
		PublicKey:   pubkey,
		Description: "",
	}
	payload3, err := rlp.EncodeToBytes(aa)
	if err != nil {
		panic("rlp payload err")
	}
	aa1 := &UpdataAccountAction{
		Founder: common.Name(""),
	}

	payload4, err := rlp.EncodeToBytes(aa1)
	if err != nil {
		panic("rlp payload err")
	}
	ca1 := common.NewAuthor(common.Name("a123456789aeee"), 1)
	author1 := &AuthorAction{ActionType: 0, Author: ca1}
	ca2 := common.NewAuthor(common.Name("a123456789aeee"), 2)
	author2 := &AuthorAction{ActionType: 1, Author: ca2}

	aaa1 := &AccountAuthorAction{Threshold: 10, AuthorActions: []*AuthorAction{author1, author2}}
	payload5, err := rlp.EncodeToBytes(aaa1)
	if err != nil {
		panic("rlp payload err")
	}

	action := types.NewAction(types.IssueAsset, common.Name("a123456789aeee"), common.Name(params.DefaultChainconfig.AssetName), 1, 1, 0, big.NewInt(0), payload, nil)
	action1 := types.NewAction(types.IncreaseAsset, common.Name("a123456789aeee"), common.Name(params.DefaultChainconfig.AssetName), 1, 1, 2, big.NewInt(0), payload2, nil)
	action2 := types.NewAction(types.UpdateAsset, common.Name("a123456789aeee"), common.Name(params.DefaultChainconfig.AssetName), 1, 1, 2, big.NewInt(0), payload1, nil)
	action3 := types.NewAction(types.CreateAccount, common.Name("a123456789aeee"), common.Name(params.DefaultChainconfig.AccountName), 1, 1, 2, big.NewInt(10), payload3, nil)
	action4 := types.NewAction(types.UpdateAccount, common.Name("a123456789addd"), common.Name(params.DefaultChainconfig.AccountName), 1, 1, 2, big.NewInt(0), payload4, nil)
	action5 := types.NewAction(types.UpdateAccountAuthor, common.Name("a123456789addd"), common.Name(params.DefaultChainconfig.AccountName), 1, 1, 2, big.NewInt(0), payload5, nil)

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
		{"updateaccountauthor", fields{sdb, ast}, args{action5}, false},

		//{"deleteaccount", fields{sdb, ast}, args{action5}, false},
		//{"transfer2self", fields{sdb, ast}, args{action6}, false},
		//{"transfer", fields{sdb, ast}, args{action7}, false},
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if _, err := am.Process(&types.AccountManagerContext{Action: tt.args.action, ChainConfig: params.DefaultChainconfig, Number: blockNumber}); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.Process() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}

	asset2, err := accountManager.GetAssetInfoByName("abced99")
	if asset2 == nil {
		t.Error("Process issue asset failure")
	}
	//t.Logf("issue ok id=%v", asset2.AssetID)
	if asset2.Amount.Cmp(big.NewInt(100000000)) != 0 {
		t.Errorf("Process increase asset failure amount=%v", asset2.Amount)
	}
	if asset2.GetAssetOwner() != "a123456789aeee" {
		t.Errorf("Process set asset owner failure =%v", asset2.GetAssetName())
	}
	val, err := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), 0, 0)
	if err != nil {
		t.Error("Process GetAccountBalanceByID err")
	}
	//t.Logf("Process GetAccountBalanceByID val=%v", val)

	ac, err := accountManager.GetAccountByName(common.Name("a123asdf2"))
	if err != nil {
		t.Error("Process GetAccountByName err")
	}
	if !ac.IsDestroyed() {
		//t.Error("Process delete account failure")
	}

	ac1, err := accountManager.GetAccountByName(common.Name("a123456789addd"))
	if err != nil {
		t.Error("Process GetAccountByName err")
	}
	if ac1 == nil {
		t.Error("Process create account err")
	}
	if ac1.Authors[1].Weight != 2 {
		t.Errorf("Process update accountauthor fail")
	}
	val, err = ac1.GetBalanceByID(1)
	if val.Cmp(big.NewInt(10)) != 0 {
		t.Errorf("Process transfer  failure=%v", val)
	}
	ca3 := common.NewAuthor(common.Name("a123456789aeee"), 1)
	author3 := &AuthorAction{ActionType: 2, Author: ca3}
	aaa2 := &AccountAuthorAction{AuthorActions: []*AuthorAction{author3}}

	payload6, err := rlp.EncodeToBytes(aaa2)
	if err != nil {
		panic("rlp payload err")
	}
	action6 := types.NewAction(types.UpdateAccountAuthor, common.Name("a123456789addd"), common.Name(sysName), 1, 0, 2, big.NewInt(0), payload6, nil)

	tests = []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"updateaccountauthor", fields{sdb, ast}, args{action6}, false},
	}
	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if _, err := am.Process(&types.AccountManagerContext{Action: tt.args.action, ChainConfig: params.DefaultChainconfig, Number: 0}); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.Process() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	// ac2, err := accountManager.GetAccountByName(common.Name("a123456789addd"))
	// if err != nil {
	// 	t.Error("Process GetAccountByName err")
	// }
	// if ac2 == nil {
	// 	t.Error("Process create account err")
	// }
	// if len(ac2.Authors) != 1 && ac2.Threshold != 10 {
	// 	t.Errorf("Process delete accountauthor fail")
	// }

}

func TestAccountManager_SubAccount(t *testing.T) {

	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		action *types.Action
	}

	pubkey, _ := GeneragePubKey()
	a := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload, err := rlp.EncodeToBytes(a)
	if err != nil {
		panic("rlp payload err")
	}

	a1 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb.cc"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload1, err := rlp.EncodeToBytes(a1)
	if err != nil {
		panic("rlp payload err")
	}

	a2 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb.dd"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload2, err := rlp.EncodeToBytes(a2)
	if err != nil {
		panic("rlp payload err")
	}

	a3 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb.ccc"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload3, err := rlp.EncodeToBytes(a3)
	if err != nil {
		panic("rlp payload err")
	}

	a4 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb.cc.dd"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload4, err := rlp.EncodeToBytes(a4)
	if err != nil {
		panic("rlp payload err")
	}

	a5 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb.cc.ee"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload5, err := rlp.EncodeToBytes(a5)
	if err != nil {
		panic("rlp payload err")
	}

	a6 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb.cc.ff"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload6, err := rlp.EncodeToBytes(a6)
	if err != nil {
		panic("rlp payload err")
	}

	a7 := &CreateAccountAction{
		AccountName: common.Name("cccccccc"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload7, err := rlp.EncodeToBytes(a7)
	if err != nil {
		panic("rlp payload err")
	}

	a8 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb.ee"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	payload8, err := rlp.EncodeToBytes(a8)
	if err != nil {
		panic("rlp payload err")
	}

	a9 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbb."),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	_, err = rlp.EncodeToBytes(a9)
	if err != nil {
		panic("rlp payload err")
	}

	a10 := &CreateAccountAction{
		AccountName: common.Name("bbbbbb"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	_, err = rlp.EncodeToBytes(a10)
	if err != nil {
		panic("rlp payload err")
	}

	a11 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbbbbbbbbbbb"),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	_, err = rlp.EncodeToBytes(a11)
	if err != nil {
		panic("rlp payload err")
	}

	a12 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbbbbbbbb.."),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	_, err = rlp.EncodeToBytes(a12)
	if err != nil {
		panic("rlp payload err")
	}

	a13 := &CreateAccountAction{
		AccountName: common.Name("bbbbbbbbbbbbbbbb.aaa.."),
		Founder:     common.Name(""),
		PublicKey:   pubkey,
	}
	_, err = rlp.EncodeToBytes(a13)
	if err != nil {
		panic("rlp payload err")
	}

	action := types.NewAction(types.CreateAccount, common.Name("a123456789aeee"), common.Name(sysName), 1, 0, 2, big.NewInt(40), payload, nil)
	action1 := types.NewAction(types.CreateAccount, common.Name("bbbbbbbb"), common.Name(sysName), 1, 0, 2, big.NewInt(10), payload1, nil)
	action2 := types.NewAction(types.CreateAccount, common.Name("bbbbbbbb"), common.Name(sysName), 1, 0, 2, big.NewInt(10), payload2, nil)
	action3 := types.NewAction(types.CreateAccount, common.Name("bbbbbbbb"), common.Name(sysName), 1, 0, 2, big.NewInt(10), payload3, nil)
	action4 := types.NewAction(types.CreateAccount, common.Name("bbbbbbbb"), common.Name(sysName), 1, 0, 2, big.NewInt(10), payload4, nil)
	action5 := types.NewAction(types.CreateAccount, common.Name("bbbbbbbb.cc"), common.Name(sysName), 1, 0, 2, big.NewInt(10), payload5, nil)
	action6 := types.NewAction(types.CreateAccount, common.Name("bbbbbbbb.ccc"), common.Name(sysName), 1, 0, 2, big.NewInt(10), payload6, nil)
	action7 := types.NewAction(types.CreateAccount, common.Name("a123456789aeee"), common.Name(sysName), 1, 0, 2, big.NewInt(30), payload7, nil)
	action8 := types.NewAction(types.CreateAccount, common.Name("cccccccc"), common.Name(sysName), 1, 0, 2, big.NewInt(30), payload8, nil)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"createaccount", fields{sdb, ast}, args{action}, false},
		{"createsubaccount1", fields{sdb, ast}, args{action1}, false},
		{"createsubaccount2", fields{sdb, ast}, args{action2}, false},
		{"createsubaccount3", fields{sdb, ast}, args{action3}, false},
		{"createsubaccount4", fields{sdb, ast}, args{action4}, true},
		{"createsubaccount5", fields{sdb, ast}, args{action5}, true},
		{"createsubaccount6", fields{sdb, ast}, args{action6}, true},
		{"createsubaccount7", fields{sdb, ast}, args{action7}, false},
		{"createsubaccount8", fields{sdb, ast}, args{action8}, true},
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if _, err := am.Process(&types.AccountManagerContext{Action: tt.args.action, ChainConfig: params.DefaultChainconfig, Number: 0}); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.Process() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}

}

func TestAccountManager_TransferContractAsset(t *testing.T) {
	ast1, err := asset.NewAssetObject("zi0123456789zi2", 0, "zi2", big.NewInt(1000), 18, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(100000), common.Name("a123456789aeee"), "")
	if err != nil {
		t.Fatal("NewAssetObject err", err)
	}
	am := &AccountManager{
		sdb: sdb,
		ast: ast,
	}

	var asset1 = IssueAsset{
		AssetName:   ast1.GetAssetName(),
		Symbol:      ast1.GetSymbol(),
		Amount:      ast1.GetAssetAmount(),
		Decimals:    ast1.GetDecimals(),
		Founder:     ast1.GetAssetFounder(),
		Owner:       ast1.GetAssetOwner(),
		UpperLimit:  ast1.GetUpperLimit(),
		Contract:    ast1.GetContract(),
		Description: ast1.GetAssetDescription(),
	}

	if _, err := am.IssueAsset(common.Name(""), asset1, blockNumber, 0); err != nil {
		t.Errorf("%q. AccountManager.IssueAsset() error = %v", ast1.AssetName, err)
	}
	am.AddAccountBalanceByName(asset1.Owner, asset1.AssetName, asset1.Amount)
	ast1, _ = am.GetAssetInfoByName(ast1.GetAssetName())

	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		fromAccount common.Name
		toAccount   common.Name
		AssetID     uint64
		value       *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"tranferOk", fields{sdb, ast}, args{common.Name("a123456789aeee"), common.Name("a123456789aeed"), ast1.AssetID, big.NewInt(3)}, false},
		{"tranferFail", fields{sdb, ast}, args{common.Name("a123456789aeed"), common.Name("a123456789aeed"), ast1.AssetID, big.NewInt(3)}, true},
	}
	val, err := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), ast1.AssetID, 0)
	if err != nil {
		t.Errorf("TransferAsset GetAccountBalanceByID err = %v", err)
	}
	if val.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("TransferAsset GetAccountBalanceByID val=%v", val)
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if err := am.TransferAsset(tt.args.fromAccount, tt.args.toAccount, tt.args.AssetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.TransferAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
	val1, err := accountManager.GetAccountBalanceByID(common.Name("a123456789aeee"), ast1.AssetID, 0)
	if err != nil {
		t.Error("TransferAsset GetAccountBalanceByID err")
	}
	if val1.Cmp(big.NewInt(997)) != 0 {
		t.Errorf("TransferAsset1 GetAccountBalanceByID val=%v", val1)
	}
}
func TestAccountManager_ProcessContractAsset(t *testing.T) {
	ast1, err := asset.NewAssetObject("zi0123456789zi3", 0, "zi3", big.NewInt(1000), 18, common.Name("a123456789aeee"), common.Name("a123456789aeee"), big.NewInt(100000), common.Name("a123456789aeee"), "")
	if err != nil {
		t.Fatal("NewAssetObject err", err)
	}
	am := &AccountManager{
		sdb: sdb,
		ast: ast,
	}

	var asset1 = IssueAsset{
		AssetName:   ast1.GetAssetName(),
		Symbol:      ast1.GetSymbol(),
		Amount:      ast1.GetAssetAmount(),
		Decimals:    ast1.GetDecimals(),
		Founder:     ast1.GetAssetFounder(),
		Owner:       ast1.GetAssetOwner(),
		UpperLimit:  ast1.GetUpperLimit(),
		Contract:    ast1.GetContract(),
		Description: ast1.GetAssetDescription(),
	}

	if _, err := am.IssueAsset(common.Name(""), asset1, blockNumber, 0); err != nil {
		t.Errorf("%q. AccountManager.IssueAsset() error = %v", ast1.AssetName, err)
	}
	ast1, _ = am.GetAssetInfoByName(ast1.GetAssetName())

	type fields struct {
		sdb *state.StateDB
		ast *asset.Asset
	}
	type args struct {
		action *types.Action
	}

	inc := &IncAsset{
		AssetID: ast1.AssetID,
		Amount:  big.NewInt(100),
		To:      common.Name("a123456789aeee"),
	}
	payload2, err := rlp.EncodeToBytes(inc)
	if err != nil {
		panic("rlp payload err")
	}

	ast01 := &asset.AssetObject{
		AssetID:    ast1.AssetID,
		AssetName:  "abced99",
		Symbol:     "aaa",
		Amount:     big.NewInt(100000000),
		Decimals:   2,
		Founder:    common.Name(sysName),
		Owner:      common.Name(sysName),
		AddIssue:   big.NewInt(0),
		UpperLimit: big.NewInt(1000000000),
	}
	payload1, err := rlp.EncodeToBytes(ast01)
	if err != nil {
		panic("rlp payload err")
	}

	action1 := types.NewAction(types.IncreaseAsset, common.Name("a123456789aeed"), common.Name(params.DefaultChainconfig.AssetName), 1, 1, 2, big.NewInt(0), payload2, nil)
	action2 := types.NewAction(types.UpdateAsset, common.Name("a123456789aeed"), common.Name(params.DefaultChainconfig.AssetName), 1, 1, 2, big.NewInt(0), payload1, nil)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//
		{"increase", fields{sdb, ast}, args{action1}, true},
		{"updateasset", fields{sdb, ast}, args{action2}, true},
	}

	for _, tt := range tests {
		am := &AccountManager{
			sdb: tt.fields.sdb,
			ast: tt.fields.ast,
		}
		if _, err := am.Process(&types.AccountManagerContext{Action: tt.args.action, ChainConfig: params.DefaultChainconfig, Number: 0}); (err != nil) != tt.wantErr {
			t.Errorf("%q. AccountManager.Process() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func Test_IssueAssetForkID1(t *testing.T) {
	type args struct {
		assetName  string
		symbol     string
		amount     *big.Int
		dec        uint64
		founder    common.Name
		owner      common.Name
		UpperLimit *big.Int
		contract   common.Name
		desc       string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"normal", args{"ft", "ft", big.NewInt(2), 18, common.Name(""), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, true},
		{"shortname", args{"z", "z", big.NewInt(2), 18, common.Name("a0123456789abc"), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, false},
		{"longname", args{"ftt0123456789ftt12", "zz", big.NewInt(2), 18, common.Name("a0123456789abc"), common.Name("a123"), big.NewInt(999999), "", ""}, false},
		{"emptyname", args{"", "z", big.NewInt(2), 18, common.Name("a0123456789abc"), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, false},
		{"symbolempty", args{"ft", "", big.NewInt(2), 18, common.Name("a0123456789abc"), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, false},
		{"amount==0", args{"ft", "z", big.NewInt(-1), 18, common.Name("a0123456789abc"), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, false},
		{"ownerempty", args{"ft", "z", big.NewInt(2), 18, common.Name(""), common.Name(""), big.NewInt(999999), "", ""}, false},
		{"shortsymbol", args{"ft", "z", big.NewInt(2), 18, common.Name("a0123456789abc"), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, false},
		{"longsymbol", args{"ft", "ftt0123456789ftt1", big.NewInt(2), 18, common.Name("a0123456789abc"), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, false},
		{"emptyname", args{"ft", "#ip0123456789ft", big.NewInt(2), 18, common.Name("a0123456789abc"), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}, false},
	}

	am := &AccountManager{
		sdb: sdb,
		ast: ast,
	}

	newAccount := common.Name("a0123456789abc")
	pubkey := new(common.PubKey)
	pubkey.SetBytes([]byte("abcde123456789"))

	//create account
	err := am.CreateAccount(common.Name("fractal.founder"), newAccount, "", 0, 0, *pubkey, "")

	if err != nil {
		t.Errorf("Test_IssueAssetForkID1 create account error = %v", err)
	}

	argBeforeFork := args{"main", "ft", big.NewInt(2), 18, common.Name(""), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}
	argSubFork := args{"main.sub", "ft", big.NewInt(2), 18, common.Name(""), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}

	assetBefore := IssueAsset{
		AssetName:   (argBeforeFork.assetName),
		Symbol:      argBeforeFork.symbol,
		Amount:      argBeforeFork.amount,
		Decimals:    argBeforeFork.dec,
		Founder:     argBeforeFork.founder,
		Owner:       argBeforeFork.owner,
		UpperLimit:  argBeforeFork.UpperLimit,
		Contract:    argBeforeFork.contract,
		Description: argBeforeFork.desc,
	}

	_, err = am.IssueAsset(newAccount, assetBefore, blockNumber, 0)

	if err != nil {
		t.Errorf("Test_IssueAssetForkID1 main error = %v", err)
	}

	subFork := IssueAsset{
		AssetName:   (argSubFork.assetName),
		Symbol:      argSubFork.symbol,
		Amount:      argSubFork.amount,
		Decimals:    argSubFork.dec,
		Founder:     argSubFork.founder,
		Owner:       argSubFork.owner,
		UpperLimit:  argSubFork.UpperLimit,
		Contract:    argSubFork.contract,
		Description: argSubFork.desc,
	}

	_, err = am.IssueAsset(newAccount, subFork, blockNumber, 1)
	if err != nil {
		t.Errorf("Test_IssueAssetForkID1 error = %v", err)
	}

	for _, tt := range tests {
		var asset = IssueAsset{
			AssetName:   (newAccount.String() + ":" + tt.args.assetName),
			Symbol:      tt.args.symbol,
			Amount:      tt.args.amount,
			Decimals:    tt.args.dec,
			Founder:     tt.args.founder,
			Owner:       tt.args.owner,
			UpperLimit:  tt.args.UpperLimit,
			Contract:    tt.args.contract,
			Description: tt.args.desc,
		}
		_, err := am.IssueAsset(newAccount, asset, blockNumber, 1)

		if (err != nil && tt.wantErr == true) || (err == nil && tt.wantErr == false) {
			t.Errorf("%q. Test_IssueAssetForkID1 error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}

	argAfterFork := args{"test", "ft", big.NewInt(2), 18, common.Name(""), common.Name("a0123456789abc"), big.NewInt(999999), "", ""}

	assetAfter := IssueAsset{
		AssetName:   (argAfterFork.assetName),
		Symbol:      argAfterFork.symbol,
		Amount:      argAfterFork.amount,
		Decimals:    argAfterFork.dec,
		Founder:     argAfterFork.founder,
		Owner:       argAfterFork.owner,
		UpperLimit:  argAfterFork.UpperLimit,
		Contract:    argAfterFork.contract,
		Description: argAfterFork.desc,
	}

	_, err = am.IssueAsset(newAccount, assetAfter, blockNumber, 1)

	if err == nil {
		t.Errorf("Test_IssueAssetForkID1 error, create invalid asset name failed")
	}

}

func Test_GetAccountNameLevel(t *testing.T) {
	level, err := GetAccountNameLevel("fractal")

	if level != 1 || err != nil {
		t.Errorf("account level 1 test failed, level:%v, err:%v", level, err)
	}

	level, err = GetAccountNameLevel("fractal.test")

	if level != 2 || err != nil {
		t.Errorf("account level 1 test failed, level:%v, err:%v", level, err)
	}
}

func Test_CheckAssetContract(t *testing.T) {
	type args struct {
		contractName string
		owner        string
		from         string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"assetContract", args{"fractal.contract", "fractal.contract", "fractal.test"}, true},
		{"assetContract", args{"fractal.contract", "fractal.test1", "fractal.test"}, false},
	}

	am := &AccountManager{
		sdb: sdb,
		ast: ast,
	}

	for _, tt := range tests {
		assetContract := am.CheckAssetContract(common.Name(tt.args.contractName),
			common.Name(tt.args.owner), common.Name(tt.args.from))

		if assetContract != tt.wantErr {
			t.Errorf("asset contract test failed")
		}
	}
}
