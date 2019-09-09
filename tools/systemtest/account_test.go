package main

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	. "github.com/fractalplatform/systemtest/rpc"
	. "github.com/smartystreets/goconvey/convey"
)

func createNewAccount(fromAccount string, fromPriKey *ecdsa.PrivateKey, newAccountName string, amount *big.Int) error {
	result, _, _ := CreateNewAccountWithName(fromAccount, fromPriKey, newAccountName, amount)
	return result
}

func registerAccount(namePrefix string, suffixLen int) (string, *ecdsa.PrivateKey) {
	newAccountName, err := GenerateValidAccountName(namePrefix, suffixLen)
	So(err, ShouldBeNil)
	result, _, priKey := CreateNewAccountWithName(SystemAccount, SystemAccountPriKey, newAccountName, nil)
	So(result, ShouldBeNil)
	return newAccountName, priKey
}

func registerAccountAndTransfer(namePrefix string, suffixLen int, assetId uint64, amount *big.Int) (string, *ecdsa.PrivateKey) {
	newAccountName, priKey := registerAccount(namePrefix, suffixLen)
	So(TransferAsset(SystemAccount, newAccountName, assetId, amount, SystemAccountPriKey), ShouldBeNil)
	return newAccountName, priKey
}

func TestCreateNormalAccount_111(t *testing.T) {
	Convey("用系统账户创建新账户(Create new account by system account)", t, func() {
		accountName, err := GenerateValidAccountName("111", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName, nil), ShouldBeNil)
	})
}

func TestCreateDuplicateAccount_112(t *testing.T) {
	Convey("用系统账户重复创建账户失败(Fail to create a duplicate account by system account)", t, func() {
		accountName, err := GenerateValidAccountName("112", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName, nil), ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName, nil), ShouldNotBeNil)
	})
}

func TestCreateErrAccount_113(t *testing.T) {
	Convey("用系统账户创建新账户失败, 账户名不符合规范(Fail to create new account, which name is invalid)", t, func() {
		So(createNewAccount(SystemAccount, SystemAccountPriKey, "&！@123456", nil), ShouldNotBeNil)
	})
}

func TestCreateErrAccount_114(t *testing.T) {
	Convey("用系统账户创建新账户失败, 账户名不符合规范(Fail to create new account, which name is invalid)", t, func() {
		accountName, err := GenerateValidAccountName("114", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName, big.NewInt(100)), ShouldBeNil)
	})
}

func TestGetAccountInfo_131(t *testing.T) {
	Convey("获取系统账户信息(Get info of system account)", t, func() {
		So(GetAccountInfo(SystemAccount), ShouldBeTrue)
	})

	Convey("获取已创建的普通账户信息(Get info of the account which has been created)", t, func() {
		accountName, err := GenerateValidAccountName("131", 8)
		So(err, ShouldBeNil)
		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName, nil), ShouldBeNil)
		So(GetAccountInfo(accountName), ShouldBeTrue)
	})
}

func TestGetAccountInfo_132(t *testing.T) {
	Convey("获取不存在的账户信息失败(Get info of the account which is not exist)", t, func() {
		accountName := GenerateAccountName("132", 8)
		So(GetAccountInfo(accountName), ShouldBeFalse)
	})
}

func TestAccountExist_133(t *testing.T) {
	Convey("确认已创建的账户存在(Confirm that the account is exist)", t, func() {
		accountName, err := GenerateValidAccountName("133", 8)
		So(err, ShouldBeNil)

		So(createNewAccount(SystemAccount, SystemAccountPriKey, accountName, nil), ShouldBeNil)

		bExist, err := AccountIsExist(accountName)
		So(err, ShouldBeNil)
		So(bExist, ShouldBeTrue)
	})
}

func TestAccountExist_134(t *testing.T) {
	Convey("确认未创建的账户不存在(Confirm that the account which hasn't been created is not exist)", t, func() {
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
