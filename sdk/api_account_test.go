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
		existed, err := api.AccountIsExist(chainCfg.SysName)
		So(err, ShouldBeNil)
		So(existed, ShouldBeTrue)
	})
}
func TestAccountInfo(t *testing.T) {
	Convey("account_getAccountByName", t, func() {
		acct, err := api.AccountInfo(chainCfg.SysName)
		So(err, ShouldBeNil)
		So(acct, ShouldNotBeNil)
	})
}
func TestAccountCode(t *testing.T) {
	Convey("account_getCode", t, func() {
		code, err := api.AccountCode(chainCfg.SysName)
		So(err, ShouldNotBeNil)
		So(code, ShouldBeEmpty)
	})
}
func TestAccountNonce(t *testing.T) {
	Convey("account_getNonce", t, func() {
		nonce, err := api.AccountNonce(chainCfg.SysName)
		So(err, ShouldBeNil)
		So(nonce, ShouldNotBeNil)
	})
}
func TestAssetInfoByName(t *testing.T) {
	Convey("account_getAssetInfoByName", t, func() {
		asset, err := api.AssetInfoByName(chainCfg.SysToken)
		So(err, ShouldBeNil)
		So(asset, ShouldNotBeNil)
	})
}
func TestAssetInfoByID(t *testing.T) {
	Convey("account_getAssetInfoByID", t, func() {
		asset, err := api.AssetInfoByID(chainCfg.SysTokenID)
		So(err, ShouldBeNil)
		So(asset, ShouldNotBeNil)
	})
}
func TestBalanceByAssetID(t *testing.T) {
	Convey("account_getAccountBalanceByID", t, func() {
		balance, err := api.BalanceByAssetID(chainCfg.SysName, chainCfg.SysTokenID, 0)
		So(err, ShouldBeNil)
		So(balance, ShouldNotBeNil)
	})
}
