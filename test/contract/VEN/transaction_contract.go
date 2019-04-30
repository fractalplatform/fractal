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

	testcommon "github.com/fractalplatform/fractal/test/common"
	jww "github.com/spf13/jwalterweatherman"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/abi"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	privateKey, _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	from          = common.Name("ftsystemio")

	pubKey = common.HexToPubKey("12345").Bytes()

	ethVault = common.Name("ethvault")
	venVault = common.Name("venvault")

	venContractAddr     = common.Name("vencontract")
	venSaleContractAddr = common.Name("vensalecontract")
	assetID             = 1

	nonce = uint64(0)

	gasLimit = uint64(20000000)
)

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

func formSetOwnerInput(abifile string, addr common.Address) ([]byte, error) {
	getSetOwnerInput, err := input(abifile, "setOwner", addr)
	if err != nil {
		jww.INFO.Println("getSetOwnerInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(getSetOwnerInput), nil
}

func formInitializeInput(abifile string) ([]byte, error) {
	getInitializeInput, err := input(abifile, "initialize", common.BytesToAddress([]byte(venContractAddr.String())), common.BytesToAddress([]byte(ethVault.String())), common.BytesToAddress([]byte(venVault.String())))
	if err != nil {
		jww.INFO.Println("getInitializeInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(getInitializeInput), nil
}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)
}

func main() {
	jww.INFO.Println("test send sundry transaction...")
	initNames()
	sendDeployVenTransaction()
	sendDeployVenSaleTransaction()
	setVenOwner()
	initializeVen()
	buyVen()
}

func initNames() {
	jww.INFO.Println("test initFrom... ")
	nonce, _ = testcommon.GetNonce(from)
	sendTransferTx(types.CreateAccount, from, ethVault, nonce, 1, big.NewInt(0), pubKey, nil)
	nonce++
	sendTransferTx(types.CreateAccount, from, venVault, nonce, 1, big.NewInt(0), pubKey, nil)
}

func sendDeployVenTransaction() {
	jww.INFO.Println("test sendDeployVenTransaction... ")
	input, err := formCreateContractInput("VEN.abi", "VEN.bin")
	if err != nil {
		jww.INFO.Println("sendDeployVenTransaction formCreateContractInput error ... ", err)
		return
	}
	nonce++
	sendTransferTx(types.CreateContract, from, venContractAddr, nonce, 1, big.NewInt(0), input, nil)
}

func sendDeployVenSaleTransaction() {
	jww.INFO.Println("test sendDeployVenSaleTransaction... ")
	input, err := formCreateContractInput("VENSale.abi", "VENSale.bin")
	if err != nil {
		jww.INFO.Println("sendDeployVenSaleTransaction formCreateContractInput error ... ", err)
		return
	}
	nonce++
	sendTransferTx(types.CreateContract, from, venSaleContractAddr, nonce, 1, big.NewInt(0), input, nil)
}

func setVenOwner() {
	jww.INFO.Println("test setVenOwner... ")
	input, err := formSetOwnerInput("VEN.abi", common.BytesToAddress([]byte(venSaleContractAddr.String())))
	if err != nil {
		jww.INFO.Println("setVenOwner formCreateContractInput error ... ", err)
		return
	}
	nonce++
	sendTransferTx(types.Transfer, from, venContractAddr, nonce, 1, big.NewInt(0), input, nil)
}

func initializeVen() {
	jww.INFO.Println("test initializeVen... ")
	input, err := formInitializeInput("VENSale.abi")
	if err != nil {
		jww.INFO.Println("sendDeployContractTransaction formCreateContractInput error ... ", err)
		return
	}
	nonce++
	sendTransferTx(types.Transfer, from, venSaleContractAddr, nonce, 1, big.NewInt(0), input, nil)
}

func buyVen() {
	jww.INFO.Println("test buyVen... ")
	nonce++
	sendTransferTx(types.Transfer, from, venSaleContractAddr, nonce, 1, big.NewInt(100000000000000000), nil, nil)
}

func sendTransferTx(txType types.ActionType, from, to common.Name, nonce, assetID uint64, value *big.Int, input, remark []byte) {
	action := types.NewAction(txType, from, to, nonce, assetID, gasLimit, value, input, remark)
	gasprice, _ := testcommon.GasPrice()
	tx := types.NewTransaction(assetID, gasprice, action)

	signer := types.MakeSigner(big.NewInt(1))
	keypair := types.MakeKeyPair(privateKey, []uint64{0})
	err := types.SignActionWithMultiKey(action, tx, signer, []*types.KeyPair{keypair})
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
