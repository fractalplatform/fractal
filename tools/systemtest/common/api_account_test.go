package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// AccountIsExist account exist
func TestAccountIsExist(t *testing.T) {
	Convey("account_accountIsExist", t, func() {
		api := NewAPI(rpchost)
		existed, err := api.AccountIsExist(systemaccount)
		So(existed, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
}

// AccountInfo get account by name
func TestAccountInfo(t *testing.T) {
	Convey("account_getAccountByName", t, func() {
		api := NewAPI(rpchost)
		acct, err := api.AccountInfo(systemaccount)
		_ = acct
		So(err, ShouldBeNil)
	})
}

// AccountNonce get account nonce
func TestAccountNonce(t *testing.T) {
	Convey("account_getNonce", t, func() {
		api := NewAPI(rpchost)
		nonce, err := api.AccountNonce(systemaccount)
		_ = nonce
		So(err, ShouldBeNil)
	})
}

// AccountCode get account code
func TestAccountCode(t *testing.T) {
	Convey("account_getCode", t, func() {
		api := NewAPI(rpchost)
		code, err := api.AccountCode(systemaccount)
		_ = code
		So(err, ShouldBeNil)
	})
}

// BalanceByAssetID get asset balance
func TestBalanceByAssetID(t *testing.T) {
	Convey("account_getAccountBalanceByID", t, func() {
		api := NewAPI(rpchost)
		balance, err := api.BalanceByAssetID(systemaccount, systemassetid)
		_ = balance
		So(err, ShouldBeNil)
	})
}

// AssetInfoByName get asset info
func TestAssetInfoByName(t *testing.T) {
	Convey("account_getAssetInfoByName", t, func() {
		api := NewAPI(rpchost)
		asset, err := api.AssetInfoByName(systemassetname)
		_ = asset
		So(err, ShouldBeNil)
	})
}

// AssetInfoByID get asset info
func TestAssetInfoByID(t *testing.T) {
	Convey("account_getAssetInfoByID", t, func() {
		api := NewAPI(rpchost)
		asset, err := api.AssetInfoByID(systemassetid)
		_ = asset
		So(err, ShouldBeNil)
	})
}
