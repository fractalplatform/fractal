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
	"fmt"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/params"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	mdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

//TestRunCode run runtime code directly
func TestRunCode(t *testing.T) {
	//fmt.Println("in TestRunCode ...")
	state, _ := state.New(common.Hash{}, state.NewDatabase(mdb.NewMemDatabase()))
	account, _ := accountmanager.NewAccountManager(state)
	//fmt.Println("in TestRunCode2 ...")
	//sender
	senderName := common.Name("jacobwolf")
	senderPubkey := common.HexToPubKey("12345")
	//
	receiverName := common.Name("denverfolk")
	receiverPubkey := common.HexToPubKey("12345")
	//
	toName := common.Name("fractal.asset")

	//fmt.Println("in TestRunCode3 ...")
	if err := account.CreateAccount(common.Name("fractal"), senderName, "", 0, 0, senderPubkey, ""); err != nil {
		fmt.Println("create sender account error\n", err)
		return
	}

	if err := account.CreateAccount(common.Name("fractal"), receiverName, "", 0, 0, receiverPubkey, ""); err != nil {
		fmt.Println("create receiver account error\n", err)
		return
	}

	if err := account.CreateAccount(common.Name("fractal"), toName, "", 0, 0, receiverPubkey, ""); err != nil {
		fmt.Println("create toName  error\n", err)
		return
	}

	action := issueAssetAction(senderName, toName)
	if _, err := account.Process(&types.AccountManagerContext{
		Action:      action,
		Number:      0,
		ChainConfig: params.DefaultChainconfig,
	}); err != nil {
		fmt.Println("issue asset error\n", err)
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

	var code = common.Hex2Bytes("60806040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063796a017a14610051578063f1db5416146100a4575b600080fd5b34801561005d57600080fd5b506100a26004803603810190808035906020019082018035906020019190919293919293908035906020019082018035906020019190919293919293905050506100cf565b005b3480156100b057600080fd5b506100b9610152565b6040518082815260200191505060405180910390f35b60008083838080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050805190602001c0915085858080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050805190602001c09050505050505050565b600060608060606040805190810160405280600d81526020017f48656c6c6f2c20776f726c642e00000000000000000000000000000000000000815250925060c06040519081016040528060828152602001610217608291399150600a6040519080825280601f01601f1916602001820160405280156101e15781602001602082028038833980820191505090505b5090508280519060200183805190602001848051906020016000cc80601f01601f191660405101604052935083935050505090560030343764623232376437303934636532313563336130663537653162636337333235353166653335316639343234393437313933343536376530663564633162663739353936326238636363623837613265623536623239666265333764363134653266346333633435623738396165346631663531663463623231393732666664a165627a7a723058208678bec8702f4e60d4763e8607e8a78d0dc259dcd70853d1105fda8a3bdfea1a0029")
	// myBinfile := "./contract/Ven/VEN.bin"
	myAbifile := "./contract/crypto/testcrypto.abi"

	// myContractName := common.Name("mycontract")

	// err = createContract(myAbifile, myBinfile, venContractName, runtimeConfig)
	// if err != nil {
	// 	fmt.Println("create venContractAddress error")
	// 	return
	// }

	receiverAcct, err := account.GetAccountByName(receiverName)
	if err != nil {
		fmt.Println("GetAccountByName receiverAcct error")
		return
	}
	fmt.Println("GetAccountByName receiverAcct id=", receiverAcct.GetAccountID())

	if receiverAcct != nil {
		receiverAcct.SetCode(code)
		account.SetAccount(receiverAcct)
	}

	myInput, err := input(myAbifile, "mydecode", "fffffffe", "asdfee")
	//myInput, err := input(myAbifile, "myencode")
	if err != nil {
		fmt.Println("initialize myInput error ", err)
		return
	}

	action = types.NewAction(types.CallContract, runtimeConfig.Origin, receiverName, 0, runtimeConfig.AssetID, runtimeConfig.GasLimit, runtimeConfig.Value, myInput, nil)

	ret, _, err := Call(action, &runtimeConfig)
	if err != nil {
		fmt.Println("call error ", err)
		return
	}

	fmt.Println("ret =", ret)
	//go test -v -test.run TestRunCode
}
