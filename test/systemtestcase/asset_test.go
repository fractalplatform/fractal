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

package main

import (
	"math/big"
	"testing"

	. "github.com/fractalplatform/fractal/test/systemtestcase/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIssueAsset_211_1(t *testing.T) {
	amount := new(big.Int).SetUint64(1000000000)
	decimals := uint64(18)

	Convey("系统账户创建资产，无收款方", t, func() {
		assetName := GenerateAssetName("211", 8)
		symbol := assetName
		fromAccount := SystemAccount
		ownerAccount := SystemAccount
		fromPriKey := SystemAccountPriKey
		So(IssueAsset(fromAccount, ownerAccount, fromPriKey, "", assetName, symbol, amount, decimals), ShouldBeNil)
	})
}

func TestIssueAsset_211_2(t *testing.T) {

	Convey("普通账户创建资产，无收款方", t, func() {
		newAccount, err := GenerateValidAccountName("211", 8)
		So(err, ShouldBeNil)

		assetName := GenerateAssetName("211", 8)
		symbol := assetName
		amount := new(big.Int).SetUint64(1000000000)
		decimals := uint64(18)
		_, err = IssueAssetWithValidAccount(newAccount, newAccount, "", assetName, symbol, amount, decimals)
		So(err, ShouldBeNil)
	})
}

func TestIssueAsset_212(t *testing.T) {
	Convey("系统账户创建资产，带收款方", t, func() {
		assetName := GenerateAssetName("211", 8)
		symbol := assetName
		amount := new(big.Int).SetUint64(1000000000)
		decimals := uint64(18)
		So(IssueAsset(SystemAccount, SystemAccount, SystemAccountPriKey, "toaccount", assetName, symbol, amount, decimals), ShouldNotBeNil)
	})
}

func TestIssueAsset_213(t *testing.T) {
	Convey("系统账户重复创建资产，无收款方", t, func() {
		assetName := GenerateAssetName("213", 8)
		symbol := assetName
		amount := new(big.Int).SetUint64(1000000000)
		decimals := uint64(18)
		So(IssueAsset(SystemAccount, SystemAccount, SystemAccountPriKey, "", assetName, symbol, amount, decimals), ShouldBeNil)
		So(IssueAsset(SystemAccount, SystemAccount, SystemAccountPriKey, "", assetName, symbol, amount, decimals), ShouldNotBeNil)
	})
}

func TestIncreaseAsset_221(t *testing.T) {
	Convey("普通账户创建资产后，增发一定数量的资产", t, func() {
		newAccount, err := GenerateValidAccountName("221", 8)
		So(err, ShouldBeNil)

		assetName := GenerateAssetName("221", 8)
		symbol := assetName
		amount := new(big.Int).SetUint64(1000000000)
		decimals := uint64(18)
		newPriKey, err := IssueAssetWithValidAccount(newAccount, newAccount, "", assetName, symbol, amount, decimals)
		So(err, ShouldBeNil)
		So(IncreaseAsset(newAccount, newPriKey, assetName, amount), ShouldBeNil)
	})
}

func TestIncreaseAsset_222(t *testing.T) {

}

func TestModifyAssetOwner_231(t *testing.T) {
	Convey("普通账户创建资产后，将资产Owner设置为新的账户", t, func() {
		accountName, err := GenerateValidAccountName("231", 8)
		So(err, ShouldBeNil)

		assetName := GenerateAssetName("231", 8)
		symbol := assetName
		amount := new(big.Int).SetUint64(1000000000)
		decimals := uint64(18)
		newPriKey, err := IssueAssetWithValidAccount(accountName, accountName, "", assetName, symbol, amount, decimals)
		So(err, ShouldBeNil)

		newAccountName, err := GenerateValidAccountName("231", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, newAccountName), ShouldBeNil)

		So(SetAssetNewOwner(accountName, assetName, newAccountName, newPriKey), ShouldBeNil)
	})
}

func TestModifyAssetOwner_232(t *testing.T) {
	Convey("普通账户创建资产后，将资产Owner设置为新的账户", t, func() {
		accountName, err := GenerateValidAccountName("231", 8)
		So(err, ShouldBeNil)

		assetName := GenerateAssetName("231", 8)
		symbol := assetName
		amount := new(big.Int).SetUint64(1000000000)
		decimals := uint64(18)
		newPriKey, err := IssueAssetWithValidAccount(accountName, accountName, "", assetName, symbol, amount, decimals)
		So(err, ShouldBeNil)

		newAccountName := GenerateAccountName("231", 8)

		So(SetAssetNewOwner(accountName, assetName, newAccountName, newPriKey), ShouldNotBeNil)
	})
}

func TestTransferAsset_241(t *testing.T) {
	Convey("系统账户向普通账户转其已有的资产，", t, func() {
		newAccountName, err := GenerateValidAccountName("241", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, newAccountName), ShouldBeNil)

		So(TransferAsset(SystemAccount, newAccountName, 1, 10000000, SystemAccountPriKey), ShouldBeNil)
		So(TransferAsset(SystemAccount, newAccountName, 1, 10000000, SystemAccountPriKey), ShouldBeNil)
	})
}

func TestTransferAsset_242(t *testing.T) {
	Convey("系统账户向普通账户转其未曾有过的的资产，", t, func() {
		newAccountName, err := GenerateValidAccountName("241", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, newAccountName), ShouldBeNil)

		So(TransferAsset(SystemAccount, newAccountName, 1, 10000000, SystemAccountPriKey), ShouldBeNil)
	})
}

func TestTransferAsset_243(t *testing.T) {
	Convey("系统账户向不存在的账户转资产，", t, func() {
		newAccountName, err := GenerateValidAccountName("242", 8)
		So(err, ShouldBeNil)

		So(TransferAsset(SystemAccount, newAccountName, 1, 10000000, SystemAccountPriKey), ShouldNotBeNil)
	})
}

func TestTransferAsset_244(t *testing.T) {
	Convey("系统账户向普通账户转未曾创建过的的资产，", t, func() {
		newAccountName, err := GenerateValidAccountName("241", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, newAccountName), ShouldBeNil)

		notExistAssetId := GetNextAssetIdFrom(1)
		So(TransferAsset(SystemAccount, newAccountName, notExistAssetId, 10000000, SystemAccountPriKey), ShouldNotBeNil)
	})
}

func TestTransferAsset_245(t *testing.T) {
	Convey("系统账户将其未拥有过的资产转给其它账号", t, func() {
		newAccountName, err := GenerateValidAccountName("241", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, newAccountName), ShouldBeNil)
		notOwnedAssetId := GetNextAssetIdFrom(1) - 1

		for notOwnedAssetId > 0 {
			balance, _ := GetAssetBalanceByID(SystemAccount, notOwnedAssetId)
			if balance.Cmp(big.NewInt(0)) == 0 {
				break
			}
			notOwnedAssetId--
		}

		So(TransferAsset(SystemAccount, newAccountName, notOwnedAssetId, 1, SystemAccountPriKey), ShouldNotBeNil)
	})
}

func TestTransferAsset_246(t *testing.T) {
	Convey("系统账户向普通账户转账金额超过其余额", t, func() {
		newAccountName, err := GenerateValidAccountName("241", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, newAccountName), ShouldBeNil)

		balance, _ := GetAssetBalanceByID(SystemAccount, 1)
		So(TransferAsset(SystemAccount, newAccountName, 1, balance.Int64()+1, SystemAccountPriKey), ShouldNotBeNil)
	})
}

func TestTransferAsset_247(t *testing.T) {
	Convey("给自己转账", t, func() {

	})
}

func TestTransferAsset_25_1to5(t *testing.T) {

	Convey("查询资产相关信息", t, func() {

		amount := new(big.Int).SetUint64(1000000000)
		decimals := uint64(18)
		assetName := GenerateAssetName("211", 8)
		symbol := assetName
		assetId := uint64(0)
		oldAccount := GetAccountByName(SystemAccount)
		So(len(oldAccount.Balances) > 0, ShouldBeTrue)

		So(IssueAsset(SystemAccount, SystemAccount, SystemAccountPriKey, "", assetName, symbol, amount, decimals), ShouldBeNil)
		Convey("1:通过资产名称查询资产id", func() {
			assetInfo, err := GetAssetInfoByName(assetName)
			So(err, ShouldBeNil)
			assetId = assetInfo.AssetId
			So(assetId > 0, ShouldBeTrue)

			Convey("2:通过资产ID查询资产对象", func() {
				assetInfo, err := GetAssetInfoById(assetId)
				So(err, ShouldBeNil)
				So(assetInfo.AssetId == assetId, ShouldBeTrue)
				So(assetInfo.AssetName == assetName, ShouldBeTrue)
			})
			Convey("3:查询某账户拥有的资产数量", func() {
				newAccount := GetAccountByName(SystemAccount)
				So(len(newAccount.Balances)-len(oldAccount.Balances) == 1, ShouldBeTrue)
			})
			Convey("4:查询所有资产数量", func() {
				maxAssetId := GetNextAssetIdFrom(1) - 1
				So(assetId == maxAssetId, ShouldBeTrue)
			})
			Convey("5:通过资产名称查询资产对象", func() {
				assetInfo, err := GetAssetInfoByName(assetName)
				So(err, ShouldBeNil)
				So(assetInfo.AssetId == assetId, ShouldBeTrue)
				So(assetInfo.AssetName == assetName, ShouldBeTrue)
			})
		})
	})
}
