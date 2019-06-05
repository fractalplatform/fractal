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
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	testcommon "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/abi"
	"github.com/fractalplatform/fractal/utils/rlp"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	privateKey, _  = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	founderAccount = common.Name("fractal.founder")
	accAccount     = common.Name("fractal.account")
	feeAccount     = common.Name("fractal.fee")

	normal_a   = common.Name("normalaccta")
	normal_b   = common.Name("normalacctb")
	contract_a = common.Name("contracta")
	contract_b = common.Name("contractb")

	multiAssetAbi = "./MultiAsset.abi"
	multiAssetBin = "./MultiAsset.bin"
	withDrawAbi   = "./WithdrawFee.abi"
	withDrawBin   = "./WithdrawFee.bin"

	a_author_0_priv *ecdsa.PrivateKey
	b_author_0_priv *ecdsa.PrivateKey
	c_author_0_priv *ecdsa.PrivateKey

	newPrivateKey_a *ecdsa.PrivateKey
	newPrivateKey_b *ecdsa.PrivateKey
	newPrivateKey_c *ecdsa.PrivateKey
	pubKey_a        common.PubKey
	pubKey_b        common.PubKey
	pubKey_c        common.PubKey

	aNonce = uint64(0)
	bNonce = uint64(0)
	cNonce = uint64(0)

	assetID        = uint64(0)
	issueAssetName = "ether" + contract_a.String()
	issueAssetID   = int64(1)
	contract_a_ID  = int64(4105)
	nonce          = uint64(0)
	gasLimit       = uint64(20000000)
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

func formIssueAssetInput(abifile string, desc string) ([]byte, error) {
	issueAssetInput, err := input(abifile, "reg", desc)
	if err != nil {
		jww.INFO.Println("formIssueAssetInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(issueAssetInput), nil
}

func formTransferAssetInput(abifile string, toAddr common.Address, value *big.Int) ([]byte, error) {
	transferAssetInput, err := input(abifile, "transAsset", toAddr, value)
	if err != nil {
		jww.INFO.Println("transferAssetInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(transferAssetInput), nil
}

func formWithdrawAssetFeeInput(abifile string, assetId *big.Int) ([]byte, error) {
	transferAssetInput, err := input(abifile, "withdrawAssetFee", assetId)
	if err != nil {
		jww.INFO.Println("transferAssetInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(transferAssetInput), nil
}

func formWithdrawContractFeeInput(abifile string, userId *big.Int) ([]byte, error) {
	transferAssetInput, err := input(abifile, "withdrawContractFee", userId)
	if err != nil {
		jww.INFO.Println("transferAssetInput error ", err)
		return nil, err
	}
	return common.Hex2Bytes(transferAssetInput), nil
}

func generateAccount() {
	nonce, _ = testcommon.GetNonce(founderAccount)
	issueAssetID = int64(nonce/4 + 1)
	contract_a_ID = int64(4105 + 4*nonce/4)

	newPrivateKey_a, _ = crypto.GenerateKey()
	pubKey_a = common.BytesToPubKey(crypto.FromECDSAPub(&newPrivateKey_a.PublicKey))
	a_author_0_priv = newPrivateKey_a
	fmt.Println("priv_a ", hex.EncodeToString(crypto.FromECDSA(newPrivateKey_a)), " pubKey_a ", pubKey_a.String())

	newPrivateKey_b, _ = crypto.GenerateKey()
	pubKey_b = common.BytesToPubKey(crypto.FromECDSAPub(&newPrivateKey_b.PublicKey))
	b_author_0_priv = newPrivateKey_b
	fmt.Println("priv_b ", hex.EncodeToString(crypto.FromECDSA(newPrivateKey_b)), " pubKey_b ", pubKey_b.String())

	newPrivateKey_c, _ = crypto.GenerateKey()
	pubKey_c = common.BytesToPubKey(crypto.FromECDSAPub(&newPrivateKey_c.PublicKey))
	c_author_0_priv = newPrivateKey_c
	fmt.Println("priv_c ", hex.EncodeToString(crypto.FromECDSA(newPrivateKey_c)), " pubKey_c ", pubKey_c.String())

	balance, _ := testcommon.GetAccountBalanceByID(founderAccount, assetID)
	balance.Div(balance, big.NewInt(10))

	normal_a = common.Name(fmt.Sprintf("normalaccta%d", nonce))
	normal_b = common.Name(fmt.Sprintf("normalacctb%d", nonce))
	contract_a = common.Name(fmt.Sprintf("contracta%d", nonce))
	contract_b = common.Name(fmt.Sprintf("contractb%d", nonce))

	key := types.MakeKeyPair(privateKey, []uint64{0})
	acct := &accountmanager.CreateAccountAction{
		AccountName: normal_a,
		Founder:     normal_a,
		PublicKey:   pubKey_a,
	}
	b, _ := rlp.EncodeToBytes(acct)
	sendTransferTx(types.CreateAccount, founderAccount, accAccount, nonce, assetID, balance, b, []*types.KeyPair{key})

	acct = &accountmanager.CreateAccountAction{
		AccountName: normal_b,
		Founder:     normal_b,
		PublicKey:   pubKey_b,
	}
	b, _ = rlp.EncodeToBytes(acct)
	sendTransferTx(types.CreateAccount, founderAccount, accAccount, nonce+1, assetID, balance, b, []*types.KeyPair{key})

	acct = &accountmanager.CreateAccountAction{
		AccountName: contract_a,
		Founder:     contract_a,
		PublicKey:   pubKey_c,
	}
	b, _ = rlp.EncodeToBytes(acct)
	sendTransferTx(types.CreateAccount, founderAccount, accAccount, nonce+2, assetID, big.NewInt(1000000000000), b, []*types.KeyPair{key})

	acct = &accountmanager.CreateAccountAction{
		AccountName: contract_b,
		Founder:     contract_b,
		PublicKey:   pubKey_c,
	}
	b, _ = rlp.EncodeToBytes(acct)
	sendTransferTx(types.CreateAccount, founderAccount, accAccount, nonce+3, assetID, balance, b, []*types.KeyPair{key})

	for {
		time.Sleep(10 * time.Second)
		aexist, _ := testcommon.CheckAccountIsExist(normal_a)
		bexist, _ := testcommon.CheckAccountIsExist(normal_b)
		cexist, _ := testcommon.CheckAccountIsExist(contract_a)
		dexist, _ := testcommon.CheckAccountIsExist(contract_b)

		if aexist && bexist && cexist && dexist {
			break
		}
	}

	issueAssetName = "ether" + contract_a.String()
	fmt.Println("normal_a ", normal_a, " normal_b ", normal_b, " contract_a ", contract_a, " contract_b ", contract_b)
}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)

}

func transferFromA2B() {
	jww.INFO.Println("transferFromA2B ")

	key_0 := types.MakeKeyPair(a_author_0_priv, []uint64{0})
	sendTransferTx(types.Transfer, normal_a, normal_b, aNonce, assetID, big.NewInt(1), nil, []*types.KeyPair{key_0})
}

func deployMultiAssetContract() {
	jww.INFO.Println("deployMultiAssetContract ")

	input, err := formCreateContractInput(multiAssetAbi, multiAssetBin)
	if err != nil {
		jww.INFO.Println("sendDeployContractTransaction formCreateContractInput error ... ", err)
		return
	}

	key_0 := types.MakeKeyPair(c_author_0_priv, []uint64{0})

	sendTransferTx(types.CreateContract, contract_a, contract_a, 0, assetID, big.NewInt(0), input, []*types.KeyPair{key_0})
}

func issueAssetForA() {
	jww.INFO.Println("issueAssetForA ")

	input, err := formIssueAssetInput(multiAssetAbi, issueAssetName+","+issueAssetName+",10000000000,10,"+contract_a.String()+",20000000000,"+contract_a.String()+",,this is contracgt asset")
	if err != nil {
		jww.INFO.Println("sendDeployContractTransaction issueAssetForA error ... ", err)
		return
	}

	key_0 := types.MakeKeyPair(a_author_0_priv, []uint64{0})
	aNonce++
	sendTransferTx(types.CallContract, normal_a, contract_a, aNonce, assetID, big.NewInt(0), input, []*types.KeyPair{key_0})
}

func transferAssetByContractFromA2B() {
	jww.INFO.Println("transferAssetByContractFromA2B ")

	input, err := formTransferAssetInput(multiAssetAbi, common.BigToAddress(big.NewInt(4099)), big.NewInt(1))
	if err != nil {
		jww.INFO.Println("transferAssetByContractFromA2B formTransferAssetInput error ... ", err)
		return
	}

	key_0 := types.MakeKeyPair(a_author_0_priv, []uint64{0})
	aNonce++
	sendTransferTx(types.CallContract, normal_a, contract_a, aNonce, assetID, big.NewInt(0), input, []*types.KeyPair{key_0})
}

func deployWithDrawContract() {
	jww.INFO.Println("deployWithDrawContract ")

	input, err := formCreateContractInput(withDrawAbi, withDrawBin)
	if err != nil {
		jww.INFO.Println("deployWithDrawContract formCreateContractInput error ... ", err)
		return
	}

	key_0 := types.MakeKeyPair(c_author_0_priv, []uint64{0})

	sendTransferTx(types.CreateContract, contract_b, contract_b, 0, assetID, big.NewInt(0), input, []*types.KeyPair{key_0})
}

func withdrawFee() {
	jww.INFO.Println("withdrawFee asset ")

	// as, err := testcommon.GetAssetInfoByName(issueAssetName)
	// if err != nil {
	// 	jww.INFO.Println("withdrawFee GetAssetInfoByName error ... ", err)
	// 	return
	// }
	fmt.Println("withdraw assetID ", issueAssetID)
	input, err := formWithdrawAssetFeeInput(withDrawAbi, big.NewInt(issueAssetID))
	if err != nil {
		jww.INFO.Println("withdrawFee formTransferAssetInput error ... ", err)
		return
	}

	key_0 := types.MakeKeyPair(a_author_0_priv, []uint64{0})
	aNonce++
	sendTransferTx(types.CallContract, normal_a, contract_b, aNonce, assetID, big.NewInt(0), input, []*types.KeyPair{key_0})

	jww.INFO.Println("withdrawFee contract ")

	fmt.Println("withdraw contract_a_ID ", contract_a_ID)
	input, err = formWithdrawContractFeeInput(withDrawAbi, big.NewInt(contract_a_ID))
	if err != nil {
		jww.INFO.Println("withdrawFee formTransferAssetInput error ... ", err)
		return
	}

	key_0 = types.MakeKeyPair(a_author_0_priv, []uint64{0})
	aNonce++
	sendTransferTx(types.CallContract, normal_a, contract_b, aNonce, assetID, big.NewInt(0), input, []*types.KeyPair{key_0})
}

func main() {
	jww.INFO.Println("test send sundry transaction...")

	generateAccount()
	transferFromA2B()
	deployMultiAssetContract()

	time.Sleep(10 * time.Second)

	issueAssetForA()
	transferAssetByContractFromA2B()
	deployWithDrawContract()

	time.Sleep(10 * time.Second)

	b, _ := testcommon.GetAccountBalanceByID(contract_a, 0)
	fmt.Println("balance ", b)
	//	withdrawFee()

	time.Sleep(10 * time.Second)

	b, _ = testcommon.GetAccountBalanceByID(contract_a, 0)
	fmt.Println("balance after withdraw ", b) //shoud be 1000000028786
}

func sendTransferTx(txType types.ActionType, from, to common.Name, nonce, assetID uint64, value *big.Int, input []byte, keys []*types.KeyPair) {
	action := types.NewAction(txType, from, to, nonce, assetID, gasLimit, value, input, nil)
	gasprice := big.NewInt(1)
	tx := types.NewTransaction(0, gasprice, action)

	signer := types.MakeSigner(big.NewInt(1))
	err := types.SignActionWithMultiKey(action, tx, signer, 0, keys)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	hash, err := testcommon.SendRawTx(rawtx)
	if err != nil {
		jww.INFO.Println("result err: ", err)

	}
	jww.INFO.Println("result hash: ", hash.Hex())
}
