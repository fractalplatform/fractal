package plugin

import (
	"math/big"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
)

func CreateAccount(context *Context) ([]byte, error) {
	return nil, nil
}

func GetAccountBalanceByID(account *accountmanager.AccountManager, accountName common.Name, assetID uint64, typeID uint64) (*big.Int, error) {
	return account.GetAccountBalanceByID(accountName, assetID, typeID)
}

func GetAccountNonce(account *accountmanager.AccountManager, accountName common.Name) (uint64, error) {
	return account.GetNonce(accountName)
}
