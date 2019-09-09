package rpc

import (
	"crypto/ecdsa"
	"errors"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
	"math/big"
)

func createAccount(fromAccount string, fromPriKey *ecdsa.PrivateKey, newAccountName string, amount *big.Int) (common.Hash, common.PubKey, *ecdsa.PrivateKey, common.Name, error) {
	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, common.PubKey{}, nil, "", err
	}
	createdAccountName := common.Name(newAccountName) //common.Name(newAccountName + strconv.FormatInt(int64(nonce),10))
	pubkey, priKey := GeneratePubKey()
	gc := NewGeAction(types.CreateAccount, accountName, createdAccountName, nonce, 1, Gaslimit, amount, pubkey[:], fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, pubkey, priKey, createdAccountName, err
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

func CreateNewAccountWithName(fromAccount string, fromPriKey *ecdsa.PrivateKey, newAccountName string, amount *big.Int) (error, common.PubKey, *ecdsa.PrivateKey) {
	txHash, pubKey, priKey, createdAccountName, err := createAccount(fromAccount, fromPriKey, newAccountName, amount)
	if err != nil {
		return errors.New("创建账户交易失败：" + err.Error() + ", hash=" + txHash.Hex()), pubKey, priKey
	}
	err = checkReceipt(fromAccount, txHash, 60)
	if err != nil {
		return errors.New("无法获取创建账号后的receipt：" + err.Error()), pubKey, priKey
	}

	bExist, err := AccountIsExist(createdAccountName.String())
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
