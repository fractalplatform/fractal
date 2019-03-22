package sdk

import (
	"math/big"

	"github.com/fractalplatform/fractal/asset"
)

// AccountIsExist account exist
func (api *API) AccountIsExist(name string) (bool, error) {
	isExist := false
	err := api.client.Call(&isExist, "account_accountIsExist", name)
	return isExist, err
}

// AccountInfo get account by name
func (api *API) AccountInfo(name string) (map[string]interface{}, error) {
	account := map[string]interface{}{}
	err := api.client.Call(account, "account_getAccountByName", name)
	return account, err
}

// AccountCode get account code
func (api *API) AccountCode(name string) ([]byte, error) {
	code := []byte{}
	err := api.client.Call(code, "account_getCode", name)
	return code, err
}

// AccountNonce get account nonce
func (api *API) AccountNonce(name string) (uint64, error) {
	nonce := uint64(0)
	err := api.client.Call(&nonce, "account_getNonce", name)
	return nonce, err
}

// AssetInfoByName get asset info
func (api *API) AssetInfoByName(name string) (*asset.AssetObject, error) {
	assetInfo := &asset.AssetObject{}
	err := api.client.Call(assetInfo, "account_getAssetInfoByName", name)
	return assetInfo, err
}

// AssetInfoByID get asset info
func (api *API) AssetInfoByID(id uint64) (*asset.AssetObject, error) {
	assetInfo := &asset.AssetObject{}
	err := api.client.Call(assetInfo, "account_getAssetInfoByID", id)
	return assetInfo, err
}

// BalanceByAssetID get asset balance
func (api *API) BalanceByAssetID(name string, id uint64) (*big.Int, error) {
	balance := big.NewInt(0)
	err := api.client.Call(balance, "account_getAccountBalanceByID", name, id)
	return balance, err
}
