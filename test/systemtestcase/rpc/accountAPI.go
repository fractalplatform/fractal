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

package rpc

import (
	"crypto/ecdsa"
	"errors"
	"strconv"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

func createAccount(fromAccount string, fromPriKey *ecdsa.PrivateKey, newAccountName string) (common.Hash, common.PubKey, *ecdsa.PrivateKey, error) {
	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, common.PubKey{}, nil, err
	}
	createdAccountName := common.Name(newAccountName)
	pubkey, priKey := GeneratePubKey()
	gc := NewGeAction(types.CreateAccount, accountName, createdAccountName, nonce, 1, Gaslimit, nil, pubkey[:], fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, pubkey, priKey, err
}

func AccountIsExist(accountName string) (bool, error) {
	isExist := new(bool)
	err := ClientCall("account_accountIsExist", isExist, accountName)
	if err != nil {
		return false, err
	}
	return *isExist, nil
}

func GetAccountInfo(accountName string) bool {
	account := &accountmanager.Account{}
	ClientCall("account_getAccountByName", account, accountName)
	return len(account.AcctName.String()) > 0
}

func GetAccountByName(accountName string) accountmanager.Account {
	account := &accountmanager.Account{}
	ClientCall("account_getAccountByName", account, accountName)
	return *account
}

func CreateNewAccountWithName(fromAccount string, fromPriKey *ecdsa.PrivateKey, newAccountName string) (error, common.PubKey, *ecdsa.PrivateKey) {
	txHash, pubKey, priKey, err := createAccount(fromAccount, fromPriKey, newAccountName)
	if err != nil {
		return errors.New("创建账户交易失败：" + err.Error()), pubKey, priKey
	}
	maxTime := uint(60)
	receipt, outOfTime, err := DelayGetReceiptByTxHash(txHash, maxTime)
	if err != nil {
		return errors.New("获取交易receipt失败：" + err.Error()), pubKey, priKey
	}
	if outOfTime {
		return errors.New("无法在" + strconv.Itoa(int(maxTime)) + "秒内获取交易receipt"), pubKey, priKey
	}
	if len(receipt.ActionResults) == 0 || receipt.ActionResults[0].Status == 0 {
		return errors.New("创建账户失败"), pubKey, priKey
	}

	bExist, err := AccountIsExist(newAccountName)
	if err != nil {
		return errors.New("无法查询到创建的账户：" + err.Error()), pubKey, priKey
	}
	if !bExist {
		return errors.New("无法查询到创建的账户"), pubKey, priKey
	}

	return nil, pubKey, priKey

}

//
//func deleteAcount(fromName common.Name, fromprikey *ecdsa.PrivateKey) (common.Hash, error) {
//	nonce, err := GetNonce(fromName)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	gc := NewGeAction(types.DeleteAccount, fromName, "", nonce, 1, Gaslimit, nil, nil, fromprikey)
//	var gcs []*GenAction
//	gcs = append(gcs, gc)
//	hash, err := SendTxTest(gcs)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	return hash, nil
//}

//func DeleteAcountWithResult(fromName common.Name, fromprikey *ecdsa.PrivateKey) bool {
//	txHash, err := deleteAcount(fromName, fromprikey)
//	if err != nil {
//		return false
//	}
//	receipt, outOfTime := DelayGetReceiptByTxHash(txHash, 60)
//
//	if outOfTime || len(receipt.ActionResults) == 0 || receipt.ActionResults[0].Status == 0 {
//		return false
//	}
//	return AccountIsExist(string(fromName))
//}

func GenerateAccountName(namePrefix string, addStrLen int) string {
	return GenerateRandomName(namePrefix, addStrLen)
}

func GenerateValidAccountName(namePrefix string, suffixStrLen int) (string, error) {
	maxTime := 10
	for maxTime > 0 {
		newAccountName := GenerateAccountName(namePrefix, suffixStrLen)
		bExist, err := AccountIsExist(newAccountName)
		if err != nil {
			return "", errors.New("判断账号是否存在的RPC接口调用失败：" + err.Error())
		}
		if !bExist {
			return newAccountName, nil
		}
		maxTime--
	}
	return "", errors.New("难以获得有效的账户名")
}

func GenerateValidAccountNameAndKey(namePrefix string, suffixStrLen int) (string, common.PubKey, *ecdsa.PrivateKey, error) {
	pubKey, priKey := GeneratePubKey()
	accountName, err := GenerateValidAccountName(namePrefix, suffixStrLen)
	return accountName, pubKey, priKey, err
}
