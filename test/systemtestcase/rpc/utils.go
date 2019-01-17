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
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

import (
	"github.com/fractalplatform/fractal/rpc"
)

var (
	once           sync.Once
	clientInstance *rpc.Client
	hostIp         = "192.168.2.13" // "127.0.0.1" //
	port           = 8090           // 8545 //
	gasPrice       = big.NewInt(2000000)
)

type GenAction struct {
	*types.Action
	PrivateKey *ecdsa.PrivateKey
}

// DefultURL default rpc url
func DefultURL() string {
	return fmt.Sprintf("http://%s:%d", hostIp, port)
}

func GeneratePubKey() (common.PubKey, *ecdsa.PrivateKey) {
	prikey, _ := crypto.GenerateKey()
	return common.BytesToPubKey(crypto.FromECDSAPub(&prikey.PublicKey)), prikey
}

func NewGeAction(at types.ActionType, from, to common.Name, nonce uint64, assetid uint64, gaslimit uint64, amount *big.Int, payload []byte, prikey *ecdsa.PrivateKey) *GenAction {
	action := types.NewAction(at, from, to, nonce, assetid, gaslimit, amount, payload)
	return &GenAction{
		Action:     action,
		PrivateKey: prikey,
	}
}
func SendTxTest(gcs []*GenAction) (common.Hash, error) {
	//nonce := GetNonce(sendaddr, "latest")
	signer := types.NewSigner(params.DefaultChainconfig.ChainID)
	var actions []*types.Action
	for _, v := range gcs {
		actions = append(actions, v.Action)
	}
	tx := types.NewTransaction(uint64(1), gasPrice, actions...)
	for _, v := range gcs {
		err := types.SignAction(v.Action, tx, signer, v.PrivateKey)
		if err != nil {
			return common.Hash{}, err
		}
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	hash, err := SendRawTx(rawtx)
	return hash, err
}

//SendRawTx send raw transaction
func SendRawTx(rawTx []byte) (common.Hash, error) {
	hash := new(common.Hash)
	err := ClientCall("ft_sendRawTransaction", hash, hexutil.Bytes(rawTx))
	return *hash, err
}

// MustRPCClient Wraper rpc's client
func MustRPCClient() (*rpc.Client, error) {
	once.Do(func() {
		client, err := rpc.DialHTTP(DefultURL())
		if err != nil {
			return
		}
		clientInstance = client
	})

	return clientInstance, nil
}

// ClientCall Wrapper rpc call api.
func ClientCall(method string, result interface{}, args ...interface{}) error {
	client, err := MustRPCClient()
	if err != nil {
		return err
	}
	err = client.CallContext(context.Background(), result, method, args...)
	if err != nil {
		return err
	}
	return nil
}

// GasPrice suggest gas price
func GasPrice() (*big.Int, error) {
	gp := big.NewInt(0)
	err := ClientCall("ft_gasPrice", gp)
	return gp, err
}

// GetNonce get nonce by address and block number.
func GetNonce(accountname common.Name) (uint64, error) {
	nonce := new(uint64)
	err := ClientCall("account_getNonce", nonce, accountname)
	return *nonce, err
}

// defaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := HomeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "pi_ledger")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "pi_ledger")
		} else {
			return filepath.Join(home, ".pi_ledger")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func GenerateRandomName(namePrefix string, addStrLen int) string {
	newRandomName := namePrefix
	var str string = "abcdefghijklmnopqrstuvwxyz0123456789"
	size := len(str)
	rand.Seed(time.Now().Unix())
	for i := 0; i < addStrLen; i++ {
		index := rand.Intn(10000) % size
		newRandomName += string(str[index])
	}
	return newRandomName
}
