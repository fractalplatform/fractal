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

	"github.com/fractalplatform/fractal/params"

	"github.com/fractalplatform/fractal/accountmanager"
	//"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/abi"
	mdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// func TestEVM(t *testing.T) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Fatalf("crashed with: %v", r)
// 		}
// 	}()

// 	Execute([]byte{
// 		byte(vm.DIFFICULTY),
// 		byte(vm.TIMESTAMP),
// 		byte(vm.GASLIMIT),
// 		byte(vm.PUSH1),
// 		byte(vm.ORIGIN),
// 		byte(vm.BLOCKHASH),
// 		byte(vm.COINBASE),
// 	}, nil, nil)
// }

// func TestExecute(t *testing.T) {
// 	ret, _, err := Execute([]byte{
// 		byte(vm.PUSH1), 10,
// 		byte(vm.PUSH1), 0,
// 		byte(vm.MSTORE),
// 		byte(vm.PUSH1), 32,
// 		byte(vm.PUSH1), 0,
// 		byte(vm.RETURN),
// 	}, nil, nil)
// 	if err != nil {
// 		t.Fatal("didn't expect error", err)
// 	}

// 	num := new(big.Int).SetBytes(ret)
// 	if num.Cmp(big.NewInt(10)) != 0 {
// 		t.Error("Expected 10, got", num)
// 	}
// }

// func TestCall(t *testing.T) {
// 	state, _ := statedb.New(common.Hash{})
// 	assetdb := asset.NewAsset(state)
// 	address := common.HexToAddress("0x0a")
// 	state.SetCode(address, []byte{
// 		byte(vm.PUSH1), 10,
// 		byte(vm.PUSH1), 0,
// 		byte(vm.MSTORE),
// 		byte(vm.PUSH1), 32,
// 		byte(vm.PUSH1), 0,
// 		byte(vm.RETURN),
// 	})

// 	ret, _, err := Call(address, nil, &Config{State: state, Asset: assetdb})
// 	if err != nil {
// 		t.Fatal("didn't expect error", err)
// 	}

// 	num := new(big.Int).SetBytes(ret)
// 	fmt.Println("num ", num)
// 	if num.Cmp(big.NewInt(10)) != 0 {
// 		t.Error("Expected 10, got", num)
// 	}
// }

// func BenchmarkCall(b *testing.B) {
// 	var definition = `[{"constant":true,"inputs":[],"name":"seller","outputs":[{"name":"","type":"address"}],"type":"function"},{"constant":false,"inputs":[],"name":"abort","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"value","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":false,"inputs":[],"name":"refund","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"buyer","outputs":[{"name":"","type":"address"}],"type":"function"},{"constant":false,"inputs":[],"name":"confirmReceived","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"state","outputs":[{"name":"","type":"uint8"}],"type":"function"},{"constant":false,"inputs":[],"name":"confirmPurchase","outputs":[],"type":"function"},{"inputs":[],"type":"constructor"},{"anonymous":false,"inputs":[],"name":"Aborted","type":"event"},{"anonymous":false,"inputs":[],"name":"PurchaseConfirmed","type":"event"},{"anonymous":false,"inputs":[],"name":"ItemReceived","type":"event"},{"anonymous":false,"inputs":[],"name":"Refunded","type":"event"}]`

// 	var code = common.Hex2Bytes("6060604052361561006c5760e060020a600035046308551a53811461007457806335a063b4146100865780633fa4f245146100a6578063590e1ae3146100af5780637150d8ae146100cf57806373fac6f0146100e1578063c19d93fb146100fe578063d696069714610112575b610131610002565b610133600154600160a060020a031681565b610131600154600160a060020a0390811633919091161461015057610002565b61014660005481565b610131600154600160a060020a039081163391909116146102d557610002565b610133600254600160a060020a031681565b610131600254600160a060020a0333811691161461023757610002565b61014660025460ff60a060020a9091041681565b61013160025460009060ff60a060020a9091041681146101cc57610002565b005b600160a060020a03166060908152602090f35b6060908152602090f35b60025460009060a060020a900460ff16811461016b57610002565b600154600160a060020a03908116908290301631606082818181858883f150506002805460a060020a60ff02191660a160020a179055506040517f72c874aeff0b183a56e2b79c71b46e1aed4dee5e09862134b8821ba2fddbf8bf9250a150565b80546002023414806101dd57610002565b6002805460a060020a60ff021973ffffffffffffffffffffffffffffffffffffffff1990911633171660a060020a1790557fd5d55c8a68912e9a110618df8d5e2e83b8d83211c57a8ddd1203df92885dc881826060a15050565b60025460019060a060020a900460ff16811461025257610002565b60025460008054600160a060020a0390921691606082818181858883f150508354604051600160a060020a0391821694503090911631915082818181858883f150506002805460a060020a60ff02191660a160020a179055506040517fe89152acd703c9d8c7d28829d443260b411454d45394e7995815140c8cbcbcf79250a150565b60025460019060a060020a900460ff1681146102f057610002565b6002805460008054600160a060020a0390921692909102606082818181858883f150508354604051600160a060020a0391821694503090911631915082818181858883f150506002805460a060020a60ff02191660a160020a179055506040517f8616bbbbad963e4e65b1366f1d75dfb63f9e9704bbbf91fb01bec70849906cf79250a15056")

// 	abi, err := abi.JSON(strings.NewReader(definition))
// 	if err != nil {
// 		b.Fatal(err)
// 	}

// 	cpurchase, err := abi.Pack("confirmPurchase")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	creceived, err := abi.Pack("confirmReceived")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	refund, err := abi.Pack("refund")
// 	if err != nil {
// 		b.Fatal(err)
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < 1; i++ {
// 		for j := 0; j < 1; j++ {
// 			Execute(code, cpurchase, nil)
// 			Execute(code, creceived, nil)
// 			Execute(code, refund, nil)
// 		}
// 	}
// }

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

	action := types.NewAction(types.IssueAsset, ownerName, "fractal.asset", 0, 0, 0, big.NewInt(0), b, nil)
	return action
}

func TestAsset(t *testing.T) {
	state, _ := state.New(common.Hash{}, state.NewDatabase(mdb.NewMemDatabase()))
	account, _ := accountmanager.NewAccountManager(state)

	senderName := common.Name("jacobwolf")
	senderPubkey := common.HexToPubKey("12345")

	receiverName := common.Name("denverfolk")
	receiverPubkey := common.HexToPubKey("12345")

	if err := account.CreateAccount(common.Name("fractal"), senderName, "", 0, 0, senderPubkey, ""); err != nil {
		fmt.Println("create sender account error", err)
		return
	}

	if err := account.CreateAccount(common.Name("fractal"), receiverName, "", 0, 0, receiverPubkey, ""); err != nil {
		fmt.Println("create receiver account error", err)
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
		AssetID:     1,
		GasLimit:    10000000000,
		GasPrice:    big.NewInt(0),
		Value:       big.NewInt(0),
		BlockNumber: new(big.Int).SetUint64(0),
	}

	binfile := "./contract/Asset/Asset.bin"
	abifile := "./contract/Asset/Asset.abi"
	contractName := common.Name("assetcontract")
	if err := account.CreateAccount(common.Name("fractal"), contractName, "", 0, 0, receiverPubkey, ""); err != nil {
		fmt.Println("create contract account error", err)
		return
	}

	err := createContract(abifile, binfile, contractName, runtimeConfig)
	if err != nil {
		fmt.Println("create calledcontract error", err)
		return
	}

	issuseAssetInput, err := input(abifile, "reg", "ethnewfromname2,ethereum,10000000000,10,assetcontract,20000000000,jacobwolf,assetcontract,this is contracgt asset")
	if err != nil {
		fmt.Println("issuseAssetInput error ", err)
		return
	}
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, contractName, 0, 1, runtimeConfig.GasLimit, runtimeConfig.Value, issuseAssetInput, nil)

	ret, _, err := Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}
	num := new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(2)) != 0 {
		t.Error("getBalance fail, want 2, get ", num)
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
	for _, b := range result {
		fmt.Println("asset result ", b)
	}

	addAssetInput, err := input(abifile, "add", big.NewInt(2), common.BigToAddress(big.NewInt(4097)), big.NewInt(210000))
	if err != nil {
		fmt.Println("addAssetInput error ", err)
		return
	}
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, contractName, 0, 1, runtimeConfig.GasLimit, runtimeConfig.Value, addAssetInput, nil)

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

	result = senderAcc.GetBalancesList()
	for _, b := range result {
		fmt.Println("asset result ", b)
	}

	transferExAssetInput, err := input(abifile, "transAsset", common.BigToAddress(big.NewInt(4098)), big.NewInt(2), big.NewInt(10000))
	if err != nil {
		fmt.Println("transferExAssetInput error ", err)
		return
	}
	runtimeConfig.Value = big.NewInt(100000)
	runtimeConfig.AssetID = 2
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, transferExAssetInput, nil)

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
	for _, b := range result {
		fmt.Println("asset sender result ", b)
	}

	result = receiverAcc.GetBalancesList()
	for _, b := range result {
		fmt.Println("asset receiver result ", b)
	}

	setOwnerInput, err := input(abifile, "setname", common.BigToAddress(big.NewInt(4098)), big.NewInt(2))
	if err != nil {
		fmt.Println("setOwnerInput error ", err)
		return
	}
	runtimeConfig.Value = big.NewInt(0)
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, setOwnerInput, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}

	getBalanceInput, err := input(abifile, "getbalance", common.BigToAddress(big.NewInt(4098)), big.NewInt(2))
	if err != nil {
		fmt.Println("getBalanceInput error ", err)
		return
	}
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, getBalanceInput, nil)

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
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, contractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, getAssetIDInput, nil)

	ret, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}
	num = new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(2)) != 0 {
		t.Error("getBalance fail, want 10000, get ", num)
	}
}
func TestBNB(t *testing.T) {
	state, _ := state.New(common.Hash{}, state.NewDatabase(mdb.NewMemDatabase()))
	account, _ := accountmanager.NewAccountManager(state)

	senderName := common.Name("jacobwolf")
	senderPubkey := common.HexToPubKey("12345")

	receiverName := common.Name("denverfolk")
	receiverPubkey := common.HexToPubKey("12345")

	if err := account.CreateAccount(common.Name("fractal"), senderName, common.Name(""), 0, 0, senderPubkey, ""); err != nil {
		fmt.Println("create sender account error", err)
		return
	}

	if err := account.CreateAccount(common.Name("fractal"), receiverName, common.Name(""), 0, 0, receiverPubkey, ""); err != nil {
		fmt.Println("create receiver account error", err)
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
		AssetID:     1,
		GasLimit:    10000000000,
		GasPrice:    big.NewInt(0),
		Value:       big.NewInt(0),
		BlockNumber: new(big.Int).SetUint64(0),
	}

	VenBinfile := "./contract/Ven/VEN.bin"
	VenAbifile := "./contract/Ven/VEN.abi"
	VenSaleBinfile := "./contract/Ven/VENSale.bin"
	VenSaleAbifile := "./contract/Ven/VENSale.abi"
	venContractName := common.Name("vencontract")
	venSaleContractName := common.Name("vensalevontract")
	ethvaultName := common.Name("ethvault")
	venvaultName := common.Name("venvault")

	if err := account.CreateAccount(common.Name("fractal"), venContractName, common.Name(""), 0, 0, receiverPubkey, ""); err != nil {
		fmt.Println("create venContractName account error", err)
		return
	}
	if err := account.CreateAccount(common.Name("fractal"), venSaleContractName, common.Name(""), 0, 0, receiverPubkey, ""); err != nil {
		fmt.Println("create venSaleContractName account error", err)
		return
	}
	account.CreateAccount(common.Name("fractal"), ethvaultName, common.Name(""), 0, 0, senderPubkey, "")
	account.CreateAccount(common.Name("fractal"), venvaultName, common.Name(""), 0, 0, senderPubkey, "")

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

	venAcct.AddBalanceByID(1, big.NewInt(1))
	venSaleAcct.AddBalanceByID(1, big.NewInt(1))

	setVenOwnerInput, err := input(VenAbifile, "setOwner", common.BytesToAddress([]byte(venSaleContractName.String())))
	if err != nil {
		fmt.Println("initializeVenSaleInput error ", err)
		return
	}

	action = types.NewAction(types.Transfer, runtimeConfig.Origin, venContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, setVenOwnerInput, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call set ven owner error ", err)
		return
	}

	initializeVenSaleInput, err := input(VenSaleAbifile, "initialize", common.BigToAddress(big.NewInt(4099)), common.BigToAddress(big.NewInt(4101)), common.BigToAddress(big.NewInt(4102)))
	if err != nil {
		fmt.Println("initializeVenSaleInput error ", err)
		return
	}
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, venSaleContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, initializeVenSaleInput, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call initialize vensale error ", err)
		return
	}

	runtimeConfig.Value = big.NewInt(100000000000000000)
	runtimeConfig.Time = big.NewInt(1503057700)
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, venSaleContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, nil, nil)

	_, _, err = Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call buy ven sale error ", err)
		return
	}

	runtimeConfig.Value = big.NewInt(0)
	getBalanceInput, err := input(VenAbifile, "balanceOf", common.BytesToAddress([]byte(senderName.String())))
	if err != nil {
		fmt.Println("getBalanceInput error ", err)
		return
	}
	action = types.NewAction(types.Transfer, runtimeConfig.Origin, venContractName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, getBalanceInput, nil)

	ret, _, err := Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call get balance error ", err)
		return
	}

	num := new(big.Int).SetBytes(ret)
	fmt.Println("num ", num)
}
