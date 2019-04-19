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
		balance, err := api.BalanceByAssetID(systemaccount, systemassetid, 0)
		So(err, ShouldBeNil)
		So(balance, ShouldNotBeNil)
	})
}
