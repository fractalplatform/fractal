package sdk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccountIsExist(t *testing.T) {
	Convey("account_accountIsExist", t, func() {
		api := NewAPI(rpchost)
		existed, err := api.AccountIsExist(systemaccount)
		So(err, ShouldBeNil)
		So(existed, ShouldBeTrue)
	})
}
func TestAccountInfo(t *testing.T) {
	Convey("account_getAccountByName", t, func() {
		api := NewAPI(rpchost)
		acct, err := api.AccountInfo(systemaccount)
		So(err, ShouldBeNil)
		So(acct, ShouldNotBeNil)
	})
}
func TestAccountCode(t *testing.T) {
	Convey("account_getCode", t, func() {
		api := NewAPI(rpchost)
		code, err := api.AccountCode(systemaccount)
		So(err, ShouldNotBeNil)
		So(code, ShouldBeEmpty)
	})
}
func TestAccountNonce(t *testing.T) {
	Convey("account_getNonce", t, func() {
		api := NewAPI(rpchost)
		nonce, err := api.AccountNonce(systemaccount)
		So(err, ShouldBeNil)
		So(nonce, ShouldNotBeNil)
	})
}
func TestAssetInfoByName(t *testing.T) {
	Convey("account_getAssetInfoByName", t, func() {
		api := NewAPI(rpchost)
		asset, err := api.AssetInfoByName(systemassetname)
		So(err, ShouldBeNil)
		So(asset, ShouldNotBeNil)
	})
}
func TestAssetInfoByID(t *testing.T) {
	Convey("account_getAssetInfoByID", t, func() {
		api := NewAPI(rpchost)
		asset, err := api.AssetInfoByID(systemassetid)
		So(err, ShouldBeNil)
		So(asset, ShouldNotBeNil)
	})
}
func TestBalanceByAssetID(t *testing.T) {
	Convey("account_getAccountBalanceByID", t, func() {
		api := NewAPI(rpchost)
		balance, err := api.BalanceByAssetID(systemaccount, systemassetid)
		So(err, ShouldBeNil)
		So(balance, ShouldNotBeNil)
	})
}
