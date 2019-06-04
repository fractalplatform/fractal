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
	if err := account.CreateAccount(common.Name("fractal"), senderName, "", 0, senderPubkey, ""); err != nil {
		fmt.Println("create sender account error\n", err)
		return
	}

	if err := account.CreateAccount(common.Name("fractal"), receiverName, "", 0, receiverPubkey, ""); err != nil {
		fmt.Println("create receiver account error\n", err)
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

	var code = common.Hex2Bytes("608060405260043610610057576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806342d6c3351461005c57806359b5d50e14610073578063f1db541614610209575b600080fd5b34801561006857600080fd5b50610071610234565b005b34801561007f57600080fd5b506100a860048036038101908080359060200190929190803590602001909291905050506102ab565b60405180806020018060200180602001868152602001858152602001848103845289818151815260200191508051906020019080838360005b838110156100fc5780820151818401526020810190506100e1565b50505050905090810190601f1680156101295780820380516001836020036101000a031916815260200191505b50848103835288818151815260200191508051906020019080838360005b83811015610162578082015181840152602081019050610147565b50505050905090810190601f16801561018f5780820380516001836020036101000a031916815260200191505b50848103825287818151815260200191508051906020019080838360005b838110156101c85780820151818401526020810190506101ad565b50505050905090810190601f1680156101f55780820380516001836020036101000a031916815260200191505b509850505050505050505060405180910390f35b34801561021557600080fd5b5061021e6102bd565b6040518082815260200191505060405180910390f35b6060600080600080600060146040519080825280601f01601f1916602001820160405280156102725781602001602082028038833980820191505090505b50955060018087805190602001d280601f01601f1916604051016040528095508196508297508398508499505050505050505050505050565b60608060606000809295509295909350565b600060608060606040805190810160405280600d81526020017f48656c6c6f2c20776f726c642e00000000000000000000000000000000000000815250925060c06040519081016040528060828152602001610382608291399150600a6040519080825280601f01601f19166020018201604052801561034c5781602001602082028038833980820191505090505b5090508280519060200183805190602001848051906020016000cc80601f01601f191660405101604052935083935050505090560030343764623232376437303934636532313563336130663537653162636337333235353166653335316639343234393437313933343536376530663564633162663739353936326238636363623837613265623536623239666265333764363134653266346333633435623738396165346631663531663463623231393732666664a165627a7a72305820e715352e1ce3ebf330fe94e519c17c38007bbc520dd6bd33f8387a541d13540a0029")
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

	myInput, err := input(myAbifile, "mydecode")
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
