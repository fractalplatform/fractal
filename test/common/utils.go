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

package common

import (
	"context"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/feemanager"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	once           sync.Once
	clientInstance *rpc.Client
	defaultRPCPath = "ft.ipc"
)

// DefultURL default rpc url
func DefultURL() string {
	if strings.HasPrefix(defaultRPCPath, "http://") {
		return defaultRPCPath
	}
	if runtime.GOOS == "windows" {
		return `\\.\pipe\` + defaultRPCPath
	}
	return filepath.Join(defaultDataDir(), defaultRPCPath)
}

func SetDefultURL(rpchost string) {
	defaultRPCPath = rpchost
}

// MustRPCClient Wraper rpc's client
func MustRPCClient() *rpc.Client {
	once.Do(func() {
		client, err := rpc.Dial(DefultURL())
		if err != nil {
			jww.ERROR.Fatalln(err)
			os.Exit(1)
		}
		clientInstance = client
	})

	return clientInstance
}

// ClientCall Wrapper rpc call api.
func ClientCall(method string, result interface{}, args ...interface{}) error {
	client := MustRPCClient()
	err := client.CallContext(context.Background(), result, method, args...)
	return err
}

// GetNonce get nonce by name and block number.
func GetNonce(accountname common.Name) (uint64, error) {
	nonce := new(uint64)
	err := ClientCall("account_getNonce", nonce, accountname)
	return *nonce, err
}

// GetAccountBalanceByID get balance by name ,assetID and number.
func GetAccountBalanceByID(accountName common.Name, assetID uint64) (*big.Int, error) {
	balance := big.NewInt(0)
	err := ClientCall("account_getAccountBalanceByID", balance, accountName, assetID, 1)
	return balance, err
}

//GetAssetInfoByName get assetINfo by accountName
func GetAssetInfoByName(assetName string) (*asset.AssetObject, error) {
	assetInfo := &asset.AssetObject{}
	err := ClientCall("account_getAssetInfoByName", assetInfo, assetName)
	return assetInfo, err
}

// GasPrice suggest gas price
func GasPrice() (*big.Int, error) {
	gp := big.NewInt(0)
	err := ClientCall("ft_gasPrice", gp)
	return gp, err
}

//SendRawTx send raw transaction
func SendRawTx(rawTx []byte) (common.Hash, error) {
	hash := new(common.Hash)
	err := ClientCall("ft_sendRawTransaction", hash, hexutil.Bytes(rawTx))
	return *hash, err
}

// CheckAccountIsExist check Account is exist.
func CheckAccountIsExist(name common.Name) (bool, error) {
	isExist := false
	err := ClientCall("account_accountIsExist", &isExist, name)
	return isExist, err
}

func GetAccountByName(name common.Name) (*accountmanager.Account, error) {
	account := &accountmanager.Account{}
	err := ClientCall("account_getAccountByName", account, name)
	return account, err
}

func GetObjectFeeByName(objectName common.Name, objectType uint64) (*feemanager.ObjectFee, error) {
	objectFee := &feemanager.ObjectFee{}
	err := ClientCall("fee_getObjectFeeByName", objectFee, objectName, objectType)
	return objectFee, err
}

func GetObjectFeeResult(startObjectFeeID uint64, count uint64, time uint64) (*feemanager.ObjectFeeResult, error) {
	objectFeeResult := &feemanager.ObjectFeeResult{}
	err := ClientCall("fee_getObjectFeeResult", objectFeeResult, startObjectFeeID, count, time)
	return objectFeeResult, err
}

func GetDposAccount(name common.Name) (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	err := ClientCall("dpos_account", fields, name.String())
	return fields, err
}

func GetBlockAndResult(blockNr rpc.BlockNumber) (*types.BlockAndResult, error) {
	result := &types.BlockAndResult{}
	err := ClientCall("ft_getBlockAndResultByNumber", result, blockNr)
	return result, err
}

// defaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func defaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "ft_ledger")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "ft_ledger")
		} else {
			return filepath.Join(home, ".ft_ledger")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
