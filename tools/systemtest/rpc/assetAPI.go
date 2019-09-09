package rpc

import (
	"crypto/ecdsa"
	"errors"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
	"math/big"
)

func transfer(from, to string, assetId uint64, amount *big.Int, prikey *ecdsa.PrivateKey) (common.Hash, error) {
	nonce, err := GetNonce(common.Name(from))
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.Transfer, common.Name(from), common.Name(to), nonce, assetId, Gaslimit, amount, nil, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return SendTxTest(gcs)
}

// GetAccountBalanceByID get balance by address ,assetID and number.
func GetAssetBalanceByID(accountName string, assetID uint64) (*big.Int, error) {
	balance := big.NewInt(0)
	err := ClientCall("account_getAccountBalanceByID", balance, common.Name(accountName), assetID)
	return balance, err
}

//GetAssetInfoByName get assetINfo by accountName
func GetAssetInfoByName(assetName string) (*asset.AssetObject, error) {
	assetInfo := &asset.AssetObject{}
	err := ClientCall("account_getAssetInfoByName", assetInfo, assetName)
	return assetInfo, err
}

func GetAssetInfoById(assetID uint64) (*asset.AssetObject, error) {
	assetInfo := &asset.AssetObject{}
	err := ClientCall("account_getAssetInfoByID", assetInfo, assetID)
	return assetInfo, err
}

func IsAssetExist(assetName string) (bool, error) {
	assetObj, err := GetAssetInfoByName(assetName)
	return assetObj.AssetName == assetName, err
}

func TransferAsset(fromAccountName, toAccountName string, assetId uint64, amount *big.Int, prikey *ecdsa.PrivateKey) error {
	oldAssetAmount, _ := GetAssetBalanceByID(toAccountName, assetId)

	err := runTxAndCheckReceipt(func() (common.Hash, error) {
		txHash, err := transfer(fromAccountName, toAccountName, assetId, amount, prikey)
		if err != nil {
			return common.Hash{}, errors.New("转账交易失败(未进txpool)：" + err.Error())
		}
		return txHash, nil
	}, fromAccountName)

	if err != nil {
		return err
	}

	newAssetAmount, err := GetAssetBalanceByID(toAccountName, assetId)
	if err != nil {
		return errors.New("转账后无法获取目标账户的资产余额：" + err.Error())
	}

	if newAssetAmount.Sub(newAssetAmount, oldAssetAmount).Cmp(amount) != 0 {
		return errors.New("转账后资产余额差不等于转账金额")
	}
	return nil
}

func IssueAssetWithValidAccount(fromAccount string, owner string, toAccount string, assetName string, symbol string, amount *big.Int, decimals uint64) (*ecdsa.PrivateKey, error) {
	err, _, priKey := CreateNewAccountWithName(SystemAccount, SystemAccountPriKey, fromAccount, nil)
	if err != nil {
		return nil, errors.New("创建新账号失败：" + err.Error())
	}
	err = TransferAsset(SystemAccount, fromAccount, 1, big.NewInt(1000000000000000), SystemAccountPriKey)
	if err != nil {
		return nil, errors.New("创建新账号成功，但用系统账户给新账号转账出错：" + err.Error())
	}

	err = IssueAsset(fromAccount, owner, priKey, toAccount, assetName, symbol, amount, decimals)
	if err != nil {
		return nil, errors.New("有足够余额的新账号无法创建资产：" + err.Error())
	}
	return priKey, nil
}

func IssueAsset(fromAccount string, owner string, fromPrikey *ecdsa.PrivateKey, toAccount string, assetName string, symbol string, amount *big.Int, decimals uint64) error {
	nonce, err := GetNonce(common.Name(fromAccount))
	if err != nil {
		return errors.New("获取nonce失败：" + err.Error())
	}

	err = runTxAndCheckReceipt(func() (common.Hash, error) {
		txHash, err := issueAsset(common.Name(fromAccount), common.Name(owner), amount, assetName, symbol, nonce, decimals, fromPrikey)
		if err != nil {
			return common.Hash{}, errors.New("发布资产的交易失败：" + err.Error())
		}
		return txHash, nil
	}, fromAccount)

	if err != nil {
		return err
	}

	bExist, err := IsAssetExist(assetName)
	if err != nil {
		return errors.New("判断资产是否存在的RPC接口调用失败：" + err.Error())
	}
	if bExist {
		return nil
	} else {
		return errors.New("无法查到新发行的资产")
	}
}

func IncreaseAsset(fromAccount string, fromPrikey *ecdsa.PrivateKey, assetName string, increasedAmount *big.Int) error {
	assetObj, err := GetAssetInfoByName(assetName)
	if err != nil {
		return errors.New("通过资产名获取资产信息失败(增发前)：" + err.Error())
	}

	nonce, err := GetNonce(common.Name(fromAccount))
	if err != nil {
		return errors.New("获取账户nonce失败：" + err.Error())
	}

	err = runTxAndCheckReceipt(func() (common.Hash, error) {
		txHash, err := increaseAsset(common.Name(fromAccount), "", assetObj.AssetId, increasedAmount, nonce, fromPrikey)
		if err != nil {
			return common.Hash{}, errors.New("发送增发资产的交易失败：" + err.Error())
		}
		return txHash, nil
	}, fromAccount)

	if err != nil {
		return err
	}

	newAssetObj, err := GetAssetInfoByName(assetName)
	if err != nil {
		return errors.New("通过资产名获取资产信息失败(增发成功后)：" + err.Error())
	}
	if new(big.Int).Sub(newAssetObj.Amount, assetObj.Amount).Cmp(increasedAmount) != 0 {
		return errors.New("增发的资产数额不对")
	}
	return nil
}

// just send a tx
func issueAsset(from, owner common.Name, amount *big.Int, assetName string, symbol string, nonce uint64, decimals uint64, prikey *ecdsa.PrivateKey) (common.Hash, error) {
	asset := &asset.AssetObject{
		AssetName: assetName,
		Symbol:    symbol,
		Amount:    amount,
		Decimals:  decimals,
		Owner:     owner,
	}
	payload, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.IssueAsset, from, "", nonce, 1, Gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return SendTxTest(gcs)
}

// just send a tx
func increaseAsset(from common.Name, to common.Name, assetId uint64, increasedAmount *big.Int, nonce uint64, prikey *ecdsa.PrivateKey) (common.Hash, error) {
	ast := &asset.AssetObject{
		AssetId:   assetId,
		AssetName: "",
		Symbol:    "",
		Amount:    increasedAmount,
	}
	payload, err := rlp.EncodeToBytes(ast)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.IncreaseAsset, from, to, nonce, assetId, Gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return SendTxTest(gcs)
}

func SetAssetNewOwner(fromAccount string, assetName string, newOwner string, priKey *ecdsa.PrivateKey) error {
	assetObj, err := GetAssetInfoByName(assetName)
	if err != nil {
		return errors.New("无法通过资产名获取资产信息：" + err.Error())
	}

	err = runTxAndCheckReceipt(func() (common.Hash, error) {
		txHash, err := setAssetOwner(common.Name(fromAccount), common.Name(newOwner), assetObj.AssetId, priKey)
		if err != nil {
			return common.Hash{}, errors.New("发送修改资产Owner的交易失败：" + err.Error())
		}
		return txHash, nil
	}, fromAccount)

	if err != nil {
		return err
	}

	newAssetObj, err := GetAssetInfoByName(assetName)
	if err != nil {
		return errors.New("通过资产名获取资产信息失败(修改Owner成功后)：" + err.Error())
	}
	if newAssetObj.Owner.String() != newOwner {
		return errors.New("资产Owner跟预期设置的Owner不一致")
	}
	return nil
}

func setAssetOwner(from, newOwner common.Name, assetId uint64, prikey *ecdsa.PrivateKey) (common.Hash, error) {
	ast := &asset.AssetObject{
		AssetId: assetId,
		Owner:   newOwner,
	}
	payload, err := rlp.EncodeToBytes(ast)
	if err != nil {
		return common.Hash{}, err
	}
	nonce, err := GetNonce(common.Name(from))
	if err != nil {
		return common.Hash{}, err
	}

	gc := NewGeAction(types.SetAssetOwner, from, "", nonce, assetId, Gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return SendTxTest(gcs)
}

func GenerateAssetName(namePrefix string, addStrLen int) string {
	return GenerateRandomName(namePrefix, addStrLen)
}

func GetNextAssetIdFrom(fromAssetId uint64) uint64 {
	if fromAssetId == 0 {
		fromAssetId = 1
	}
	for {
		assetInfo, err := GetAssetInfoById(fromAssetId)
		if err != nil {
			return 0
		}
		if assetInfo.AssetId > 0 {
			fromAssetId++
		} else {
			return fromAssetId
		}
	}
}

func GenerateValidAssetName(namePrefix string, suffixStrLen int) (string, error) {
	maxTime := 10
	for maxTime > 0 {
		newAssetName := GenerateAssetName(namePrefix, suffixStrLen)
		bExist, err := IsAssetExist(newAssetName)
		if err != nil {
			return "", errors.New("判断资产是否存在的RPC接口调用失败：" + err.Error())
		}
		if !bExist {
			return newAssetName, nil
		}
		maxTime--
	}
	return "", errors.New("难以获得有效的资产名")

}
