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
	"crypto/ecdsa"
	"testing"

	. "github.com/fractalplatform/fractal/test/systemtestcase/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func createNewAccount(fromAccount string, fromPriKey *ecdsa.PrivateKey, newAccountName string) error {
	result, _, _ := CreateNewAccountWithName(fromAccount, fromPriKey, newAccountName)
	return result
}

func TestCreateNormalAccount_111(t *testing.T) {
	Convey("用系统账户创建新账户", t, func() {
		accountName, err := GenerateValidAccountName("111", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName), ShouldBeNil)
	})
}

func TestCreateDuplicateAccount_112(t *testing.T) {
	Convey("用系统账户重复创建账户失败", t, func() {
		accountName, err := GenerateValidAccountName("112", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName), ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName), ShouldNotBeNil)
	})
}

func TestCreateErrAccount_113(t *testing.T) {
	Convey("用系统账户创建新账户失败（账户名不符合规范）", t, func() {
		So(createNewAccount(SystemAccount, SystemAccountPriKey, "&！@123456"), ShouldNotBeNil)
	})
}

func TestGetAccountInfo_131(t *testing.T) {
	Convey("获取系统账户信息", t, func() {
		So(GetAccountInfo(SystemAccount), ShouldBeTrue)
	})

	Convey("获取已创建的普通账户信息", t, func() {
		accountName, err := GenerateValidAccountName("131", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName), ShouldBeNil)
		So(GetAccountInfo(accountName), ShouldBeTrue)
	})
}

func TestGetAccountInfo_132(t *testing.T) {
	Convey("获取不存在的账户信息失败", t, func() {
		accountName := GenerateAccountName("132", 8)
		So(GetAccountInfo(accountName), ShouldBeFalse)
	})
}

func TestAccountExist_133(t *testing.T) {
	Convey("确认已创建的账户存在", t, func() {
		accountName, err := GenerateValidAccountName("133", 8)
		So(err, ShouldBeNil)

		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName), ShouldBeNil)

		bExist, err := AccountIsExist(accountName)
		So(err, ShouldBeNil)
		So(bExist, ShouldBeTrue)
	})
}

func TestAccountExist_134(t *testing.T) {
	Convey("确认未创建的账户不存在", t, func() {
		accountName := GenerateAccountName("134", 8)
		bExist, err := AccountIsExist(accountName)
		So(err, ShouldBeNil)
		So(bExist, ShouldBeFalse)
	})
}

//func TestDeleteAccount_121(t *testing.T) {
//	Convey("删除已存在账号", t, func() {
//		accountName := GenerateAccountName("121", 8)
//		So(deleteAccount(SystemAccount, SystemAccountPriKey, accountName), ShouldBeTrue)
//	})
//}
//
//func TestDeleteAccount_122(t *testing.T) {
//	Convey("删除不存在账号", t, func() {
//		accountName := GenerateAccountName("122", 8)
//		_, prikey := GeneratePubKey()
//		So(deleteNotExistAcount(accountName, prikey), ShouldBeFalse)
//	})
//}

//func deleteAccount(fromname string, fromprikey *ecdsa.PrivateKey, newname string) bool {
//	result, _, prikey := CreateNewAccountWithName(fromname, fromprikey, newname)
//	if !result {
//		return false
//	}
//	return DeleteAcountWithResult(common.Name(newname), prikey)
//}
//
//func deleteNotExistAcount(fromname string, fromprikey *ecdsa.PrivateKey) bool {
//	return DeleteAcountWithResult(common.Name(fromname), fromprikey)
//}
