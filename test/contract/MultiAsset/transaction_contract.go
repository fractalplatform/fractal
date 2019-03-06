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
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	testcommon "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/abi"
	"github.com/fractalplatform/fractal/utils/rlp"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	abifile       = "MultiAsset.abi"
	binfile       = "MultiAsset.bin"
	privateKey, _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	from          = common.Name("ftsystemio")
	to            = common.Name("toaddrname")
	newFrom       = common.Name("newfromname")

	contractAddr = common.Name("multiasset")
	assetID      = uint64(1)

	nonce = uint64(0)

	gasLimit = uint64(2000000)
)

func generateAccount() {
	nonce, _ = testcommon.GetNonce(from)

	newPrivateKey, _ := crypto.GenerateKey()
	pubKey := common.BytesToPubKey(crypto.FromECDSAPub(&newPrivateKey.PublicKey))

	balance, _ := testcommon.GetAccountBalanceByID(from, assetID)
	balance.Div(balance, big.NewInt(10))

	newFrom = common.Name(fmt.Sprintf("newfromname%d", nonce))
	contractAddr = common.Name(fmt.Sprintf("multiasset%d", nonce))

	sendTransferTx(types.CreateAccount, from, newFrom, nonce, assetID, balance, pubKey.Bytes())
	sendTransferTx(types.CreateAccount, from, to, nonce+1, assetID, big.NewInt(0), pubKey.Bytes())
	sendTransferTx(types.CreateAccount, from, contractAddr, nonce+2, assetID, big.NewInt(0), pubKey.Bytes())

	for {
		time.Sleep(10 * time.Second)
		fromexist, _ := testcommon.CheckAccountIsExist(newFrom)
		toexist, _ := testcommon.CheckAccountIsExist(to)
		if fromexist && toexist {
			break
		}
	}

	from = newFrom
	fmt.Println("from ", from)
	privateKey = newPrivateKey
}

func input(abifile string, method string, params ...interface{}) (string, error) {
	var abicode string

	hexcode, err := ioutil.ReadFile(abifile)
	if err != nil {
		fmt.Printf("Could not load code from file: %v\n", err)
		return "", err
	}
	abicode = string(bytes.TrimRight(hexcode, "\n"))

	parsed, err := abi.JSON(strings.NewReader(abicode))
	if err != nil {
		fmt.Println("abi.json error ", err)
		return "", err
	}

	input, err := parsed.Pack(method, params...)
	if err != nil {
		fmt.Println("parsed.pack error ", err)
		return "", err
	}
	return common.Bytes2Hex(input), nil
}

func formCreateContractInput(abifile string, binfile string) ([]byte, error) {
	hexcode, err := ioutil.ReadFile(binfile)
	if err != nil {
		jww.INFO.Printf("Could not load code from file: %v\n", err)
		return nil, err
	}
	code := common.Hex2Bytes(string(bytes.TrimRight(hexcode, "\n")))

	createInput, err := input(abifile, "")
	if err != nil {
		jww.INFO.Println("createInput error ", err)
		return nil, err
	}

	createCode := append(code, common.Hex2Bytes(createInput)...)
	return createCode, nil
}

func formIssueAssetInput(abifile string, desc string) ([]byte, error) {
	issueAssetInput, err := input(abifile, "reg", desc)
	if err != nil {
		jww.INFO.Println("createInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(issueAssetInput), nil
}

func formTransferAssetInput(abifile string, toAddr common.Address, assetId common.Address, value *big.Int) ([]byte, error) {
	transferAssetInput, err := input(abifile, "transAsset", toAddr, assetId, value)
	if err != nil {
		jww.INFO.Println("transferAssetInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(transferAssetInput), nil
}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)

	generateAccount()
}

func main() {
	jww.INFO.Println("test send sundry transaction...")
	sendDeployContractTransaction()
	sendIssueTransaction()
	for {
		time.Sleep(10 * time.Second)
		assetName := "eth" + from.String()
		asset, _ := testcommon.GetAssetInfoByName(assetName)
		if asset.AssetId != 0 {
			assetID = asset.AssetId
			fmt.Println("assetID ", assetID)
			break
		}
	}
	sendFulfillContractTransaction()
	sendTransferTransaction()
}

func sendDeployContractTransaction() {
	jww.INFO.Println("test sendDeployContractTransaction... ")
	input, err := formCreateContractInput(abifile, binfile)
	if err != nil {
		jww.INFO.Println("sendDeployContractTransaction formCreateContractInput error ... ", err)
		return
	}
	nonce, _ = testcommon.GetNonce(from)
	sendTransferTx(types.CreateContract, from, contractAddr, nonce, assetID, big.NewInt(0), input)
}

func sendIssueTransaction() {
	jww.INFO.Println("test sendIssueTransaction... ")
	issueStr := "eth" + from.String() + ",ethereum,10000000000,10," + from.String() + ",10000000000," + from.String()
	input, err := formIssueAssetInput(abifile, issueStr)
	if err != nil {
		jww.INFO.Println("sendIssueTransaction formIssueAssetInput error ... ", err)
		return
	}
	nonce++
	sendTransferTx(types.Transfer, from, contractAddr, nonce, assetID, big.NewInt(0), input)
}

func sendFulfillContractTransaction() {
	jww.INFO.Println("test sendFulfillContractTransaction... ")
	nonce++
	sendTransferTx(types.Transfer, from, contractAddr, nonce, assetID, big.NewInt(1000000000), nil)
}

func sendTransferTransaction() {
	jww.INFO.Println("test sendTransferTransaction... ")
	input, err := formTransferAssetInput(abifile, common.BytesToAddress([]byte(to.String())), common.BigToAddress(big.NewInt(int64(assetID))), big.NewInt(1))
	if err != nil {
		jww.INFO.Println("sendDeployContractTransaction formCreateContractInput error ... ", err)
		return
	}

	for {
		nonce++
		sendTransferTx(types.Transfer, from, contractAddr, nonce, assetID, big.NewInt(0), input)
	}
}

func sendTransferTx(txType types.ActionType, from, to common.Name, nonce, assetID uint64, value *big.Int, input []byte) {
	action := types.NewAction(txType, from, to, nonce, assetID, gasLimit, value, input)
	gasprice, _ := testcommon.GasPrice()
	tx := types.NewTransaction(1, gasprice, action)

	signer := types.MakeSigner(big.NewInt(1))
	err := types.SignAction(action, tx, signer, privateKey)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	hash, _ := testcommon.SendRawTx(rawtx)
	jww.INFO.Println("result hash: ", hash.Hex())
}
