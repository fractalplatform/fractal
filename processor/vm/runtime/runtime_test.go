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

package runtime

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/abi"
	"github.com/fractalplatform/fractal/utils/rlp"
	"github.com/stretchr/testify/assert"
)

func input(abifile string, method string, params ...interface{}) ([]byte, error) {
	var abicode string

	hexcode, err := ioutil.ReadFile(abifile)
	if err != nil {
		fmt.Printf("Could not load code from file: %v\n", err)
		return nil, err
	}
	abicode = string(bytes.TrimRight(hexcode, "\n"))

	parsed, err := abi.JSON(strings.NewReader(abicode))
	if err != nil {
		fmt.Println("abi.json error ", err)
		return nil, err
	}

	input, err := parsed.Pack(method, params...)
	if err != nil {
		fmt.Println("parsed.pack error ", err)
		return nil, err
	}
	return input, nil
}

func createContract(abifile string, binfile string, contractName common.Name, runtimeConfig Config) error {
	hexcode, err := ioutil.ReadFile(binfile)
	if err != nil {
		fmt.Printf("Could not load code from file: %v\n", err)
		os.Exit(1)
	}
	code := common.Hex2Bytes(string(bytes.TrimRight(hexcode, "\n")))

	createInput, err := input(abifile, "")
	if err != nil {
		fmt.Println("createInput error ", err)
		return err
	}

	createCode := append(code, createInput...)
	action := types.NewAction(types.CreateContract, runtimeConfig.Origin, contractName, 0, 1, runtimeConfig.GasLimit, runtimeConfig.Value, createCode, nil)
	_, _, err = Create(action, &runtimeConfig)
	if err != nil {
		fmt.Println("create error ", err)
		return err
	}
	return nil
}

func createAccount(account *accountmanager.AccountManager, name string) error {
	if err := account.CreateAccount(common.Name("fractal"), common.Name(name), "", 0, 2, common.HexToPubKey("12345"), ""); err != nil {
		fmt.Printf("create account %s err %s", name, err)
		return fmt.Errorf("create account %s err %s", name, err)
	}
	return nil
}

func issueAssetAction(ownerName, toName common.Name) *types.Action {
	asset := accountmanager.IssueAsset{
		AssetName:  "bitcoin",
		Symbol:     "btc",
		Amount:     big.NewInt(1000000000000000000),
		Decimals:   10,
		Owner:      ownerName,
		UpperLimit: big.NewInt(2000000000000000000),
		Founder:    ownerName,
	}

	b, err := rlp.EncodeToBytes(asset)
	if err != nil {
		panic(err)
	}

	action := types.NewAction(types.IssueAsset, ownerName, common.Name("fractal.asset"), 0, 0, 0, big.NewInt(0), b, nil)
	return action
}

func TestAsset(t *testing.T) {
	state, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	account, _ := accountmanager.NewAccountManager(state)

	senderName := common.Name("jacobwolf12345")
	senderPubkey := common.HexToPubKey("12345")

	receiverName := common.Name("denverfolk12345")

	if err := createAccount(account, "jacobwolf12345"); err != nil {
		return
	}

	if err := createAccount(account, "denverfolk12345"); err != nil {
		return
	}

	if err := createAccount(account, "fractal.asset"); err != nil {
		return
	}

	action := issueAssetAction(senderName, receiverName)
	if _, err := account.Process(&types.AccountManagerContext{
		Action:      action,
		Number:      0,
		ChainConfig: params.DefaultChainconfig,
	}); err != nil {
		fmt.Println("issue asset error", err)
		return
	}

	runtimeConfig := Config{
		Origin:      senderName,
		FromPubkey:  senderPubkey,
		State:       state,
		Account:     account,
		AssetID:     0,
		GasLimit:    10000000000,
		GasPrice:    big.NewInt(0),
		Value:       big.NewInt(0),
		BlockNumber: new(big.Int).SetUint64(0),
	}

	binfile := "./contract/Asset/Asset.bin"
	abifile := "./contract/Asset/Asset.abi"
	contractName := common.Name("assetcontract")
	if err := createAccount(account, "assetcontract"); err != nil {
		return
	}

	err := createContract(abifile, binfile, contractName, runtimeConfig)
	if err != nil {
		fmt.Println("create calledcontract error", err)
		return
	}

	issuseAssetInput, err := input(abifile, "reg", "ethnewfromname2,ethereum,10000000000,0,jacobwolf12345,20000000000,jacobwolf12345,assetcontract,this is contracgt asset")
	if err != nil {
		fmt.Println("issuseAssetInput error ", err)
		return
	}
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, contractName, 0, 0, runtimeConfig.GasLimit, runtimeConfig.Value, issuseAssetInput, nil)

	ret, _, err := Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}
	num := new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(1)) != 0 {
		t.Error("getBalance fail, want 1, get ", num)
	}

	senderAcc, err := account.GetAccountByName(senderName)
	if err != nil {
		fmt.Println("GetAccountByName sender account error", err)
		return
	}

	result := senderAcc.GetBalancesList()
	if err != nil {
		fmt.Println("GetAllAccountBalancesset error", err)
		return
	}
	assert.Equal(t, result[0], &accountmanager.AssetBalance{AssetID: 0, Balance: big.NewInt(1000000000000000000)})
	assert.Equal(t, result[1], &accountmanager.AssetBalance{AssetID: 1, Balance: big.NewInt(10000000000)})

	addAssetInput, err := input(abifile, "add", big.NewInt(1), common.BigToAddress(big.NewInt(4097)), big.NewInt(210000))
	if err != nil {
		fmt.Println("addAssetInput error ", err)
		return
	}
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, contractName, 0, 0, runtimeConfig.GasLimit, runtimeConfig.Value, addAssetInput, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("add call error ", err)
		return
	}

	senderAcc, err = account.GetAccountByName(senderName)
	if err != nil {
		fmt.Println("GetAccountByName sender account error", err)
		return
	}

	result = senderAcc.GetBalancesList()
	for _, b := range result {
		fmt.Println("asset result ", b)
	}

	transferExAssetInput, err := input(abifile, "transAsset", common.BigToAddress(big.NewInt(4098)), big.NewInt(1), big.NewInt(10000))
	if err != nil {
		fmt.Println("transferExAssetInput error ", err)
		return
	}
	runtimeConfig.Value = big.NewInt(100000)
	runtimeConfig.AssetID = 1
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, transferExAssetInput, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}

	senderAcc, err = account.GetAccountByName(senderName)
	if err != nil {
		fmt.Println("GetAccountByName sender account error", err)
		return
	}

	receiverAcc, err := account.GetAccountByName(receiverName)
	if err != nil {
		fmt.Println("GetAccountByName receiver account error", err)
		return
	}

	result = senderAcc.GetBalancesList()
	assert.Equal(t, result[0], &accountmanager.AssetBalance{AssetID: 0, Balance: big.NewInt(1000000000000000000)})
	assert.Equal(t, result[1], &accountmanager.AssetBalance{AssetID: 1, Balance: big.NewInt(9999900000)})

	result = receiverAcc.GetBalancesList()
	assert.Equal(t, result[0], &accountmanager.AssetBalance{AssetID: 1, Balance: big.NewInt(10000)})

	setOwnerInput, err := input(abifile, "setname", common.BigToAddress(big.NewInt(4098)), big.NewInt(1))
	if err != nil {
		fmt.Println("setOwnerInput error ", err)
		return
	}
	runtimeConfig.Value = big.NewInt(0)
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, setOwnerInput, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}

	getBalanceInput, err := input(abifile, "getbalance", common.BigToAddress(big.NewInt(4098)), big.NewInt(1))
	if err != nil {
		fmt.Println("getBalanceInput error ", err)
		return
	}
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, getBalanceInput, nil)

	ret, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}
	num = new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(10000)) != 0 {
		t.Error("getBalance fail, want 10000, get ", num)
	}

	getAssetIDInput, err := input(abifile, "getAssetId")
	if err != nil {
		fmt.Println("getBalanceInput error ", err)
		return
	}
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, getAssetIDInput, nil)

	ret, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}
	num = new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(1)) != 0 {
		t.Error("getBalance fail, want 1, get ", num)
	}
}
func TestVEN(t *testing.T) {
	state, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	account, _ := accountmanager.NewAccountManager(state)

	senderName := common.Name("jacobwolf12345")
	senderPubkey := common.HexToPubKey("12345")

	receiverName := common.Name("denverfolk12345")

	if err := createAccount(account, "jacobwolf12345"); err != nil {
		return
	}

	if err := createAccount(account, "denverfolk12345"); err != nil {
		return
	}

	if err := createAccount(account, "fractal.asset"); err != nil {
		return
	}

	action := issueAssetAction(senderName, receiverName)
	if _, err := account.Process(&types.AccountManagerContext{
		Action:      action,
		Number:      0,
		ChainConfig: params.DefaultChainconfig,
	}); err != nil {
		fmt.Println("issue asset error", err)
		return
	}

	runtimeConfig := Config{
		Origin:      senderName,
		FromPubkey:  senderPubkey,
		State:       state,
		Account:     account,
		AssetID:     0,
		GasLimit:    10000000000,
		GasPrice:    big.NewInt(0),
		Value:       big.NewInt(0),
		BlockNumber: new(big.Int).SetUint64(0),
	}

	VenBinfile := "./contract/Ven/VEN.bin"
	VenAbifile := "./contract/Ven/VEN.abi"
	VenSaleBinfile := "./contract/Ven/VENSale.bin"
	VenSaleAbifile := "./contract/Ven/VENSale.abi"
	venContractName := common.Name("vencontract12345")
	venSaleContractName := common.Name("vensalevontract")

	if err := createAccount(account, "vencontract12345"); err != nil {
		return
	}

	if err := createAccount(account, "vensalevontract"); err != nil {
		return
	}

	if err := createAccount(account, "ethvault12345"); err != nil {
		return
	}

	if err := createAccount(account, "venvault12345"); err != nil {
		return
	}

	err := createContract(VenSaleAbifile, VenSaleBinfile, venSaleContractName, runtimeConfig)
	if err != nil {
		fmt.Println("create venSaleContractAddress error")
		return
	}

	err = createContract(VenAbifile, VenBinfile, venContractName, runtimeConfig)
	if err != nil {
		fmt.Println("create venContractAddress error")
		return
	}

	venAcct, err := account.GetAccountByName(venContractName)
	if err != nil {
		fmt.Println("GetAccountByName venContractAddress error")
		return
	}

	venSaleAcct, err := account.GetAccountByName(venSaleContractName)
	if err != nil {
		fmt.Println("GetAccountByName venSaleContractAddress error")
		return
	}

	venAcct.AddBalanceByID(0, big.NewInt(1))
	venSaleAcct.AddBalanceByID(0, big.NewInt(1))
	account.SetAccount(venAcct)
	account.SetAccount(venSaleAcct)

	setVenOwnerInput, err := input(VenAbifile, "setOwner", common.BigToAddress(big.NewInt(4101)))
	if err != nil {
		fmt.Println("initializeVenSaleInput error ", err)
		return
	}

	action = types.NewAction(types.CallContract, runtimeConfig.Origin, venContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, setVenOwnerInput, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call set ven owner error ", err)
		return
	}

	initializeVenSaleInput, err := input(VenSaleAbifile, "initialize", common.BigToAddress(big.NewInt(4100)), common.BigToAddress(big.NewInt(4102)), common.BigToAddress(big.NewInt(4103)))
	if err != nil {
		fmt.Println("initializeVenSaleInput error ", err)
		return
	}
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, venSaleContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, initializeVenSaleInput, nil)
	runtimeConfig.Time = big.NewInt(1504180700)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call initialize vensale error ", err)
		return
	}

	runtimeConfig.Value = big.NewInt(100000000000000000)
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, venSaleContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, nil, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call buy ven sale error ", err)
		return
	}

	runtimeConfig.Value = big.NewInt(0)
	getBalanceInput, err := input(VenAbifile, "balanceOf", common.BigToAddress(big.NewInt(4097)))
	if err != nil {
		fmt.Println("getBalanceInput error ", err)
		return
	}
	action = types.NewAction(types.CallContract, runtimeConfig.Origin, venContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, getBalanceInput, nil)

	ret, _, err := Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call get balance error ", err)
		return
	}

	num := new(big.Int).SetBytes(ret)
	assert.Equal(t, num, new(big.Int).Mul(big.NewInt(3500000000), big.NewInt(100000000000)))
}
