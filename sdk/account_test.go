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

package sdk

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	// 	tAssetName   = "testasset"
	// 	tAssetSymbol = "tat"
	// 	tAmount      = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e8))
	// 	tDecimals    = uint64(8)
	// 	tAssetID     uint64
	rpchost         = "http://127.0.0.1:8545"
	systemaccount   = params.DefaultChainconfig.SysName
	accountaccount  = params.DefaultChainconfig.AccountName
	dposaccount     = params.DefaultChainconfig.DposName
	assetaccount    = params.DefaultChainconfig.AssetName
	systemprivkey   = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
	systemassetname = params.DefaultChainconfig.SysToken
	systemassetid   = uint64(0)
	chainid         = big.NewInt(1)
	tValue          = new(big.Int).Mul(big.NewInt(300000), big.NewInt(1e18))
	tGas            = uint64(20000000)

	AssetAbi = "./test/Asset.abi"
	AssetBin = "./test/Asset.bin"
)

func TestAccount(t *testing.T) {
	Convey("Account", t, func() {
		api := NewAPI(rpchost)
		var systempriv, _ = crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)
		// CreateAccount
		priv, pub := GenerateKey()
		accountName := common.StrToName(GenerateAccountName("test", 8))
		hash, err := sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.CreateAccountAction{
			AccountName: accountName,
			PublicKey:   pub,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// Transfer
		hash, err = sysAcct.Transfer(accountName, tValue, systemassetid, tGas)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// UpdateAccount
		acct := NewAccount(api, accountName, priv, systemassetid, math.MaxUint64, true, chainid)
		// _, npub := GenerateKey()
		hash, err = acct.UpdateAccount(common.StrToName(accountaccount), new(big.Int).Mul(tValue, big.NewInt(0)), systemassetid, tGas, &accountmanager.UpdataAccountAction{
			Founder: accountName,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// DestroyAccount
	})
}

func TestAsset(t *testing.T) {
	Convey("Asset", t, func() {
		api := NewAPI(rpchost)
		var systempriv, _ = crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)
		// CreateAccount
		priv, pub := GenerateKey()
		accountName := common.StrToName(GenerateAccountName("test", 8))
		hash, err := sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.CreateAccountAction{
			AccountName: accountName,
			PublicKey:   pub,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		acct := NewAccount(api, accountName, priv, systemassetid, math.MaxUint64, true, chainid)
		assetname := common.StrToName(GenerateAccountName("asset", 2)).String()
		// IssueAsset
		ast1 := accountmanager.IssueAsset{
			AssetName: assetname,
			Symbol:    assetname[len(assetname)-4:],
			Amount:    new(big.Int).Mul(big.NewInt(10000000), big.NewInt(1e18)),
			Decimals:  18,
			Owner:     accountName,
			Founder:   accountName,
			//AddIssue:   big.NewInt(0),
			UpperLimit: big.NewInt(0),
		}

		hash, err = acct.IssueAsset(common.StrToName(assetaccount), big.NewInt(0), systemassetid, tGas, &ast1)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		ast, _ := api.AssetInfoByName(assetname)

		ast2 := accountmanager.UpdateAsset{
			AssetID: ast.AssetId,
			Founder: accountName,

			//UpperLimit: big.NewInt(0),
		}

		// acct.UpdateAsset()
		hash, err = acct.UpdateAsset(common.StrToName(assetaccount), big.NewInt(0), systemassetid, tGas, &ast2)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		hash, err = acct.IncreaseAsset(common.StrToName(assetaccount), big.NewInt(0), systemassetid, tGas, &accountmanager.IncAsset{
			Amount:  new(big.Int).Mul(big.NewInt(10000000), big.NewInt(1e18)),
			To:      accountName,
			AssetId: ast.AssetId,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
		// acct.SetAssetOwner()
		// acct.DestroyAsset()
	})
}

func TestDPOS(t *testing.T) {
	SkipConvey("DPOS", t, func() {
		api := NewAPI(rpchost)
		var systempriv, _ = crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)
		priv, pub := GenerateKey()
		accountName := common.StrToName(GenerateAccountName("prod", 8))
		hash, err := sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.CreateAccountAction{
			AccountName: accountName,
			PublicKey:   pub,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		priv2, pub2 := GenerateKey()
		accountName2 := common.StrToName(GenerateAccountName("voter", 8))
		hash, err = sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.CreateAccountAction{
			AccountName: accountName2,
			PublicKey:   pub2,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// RegCandidate
		acct := NewAccount(api, accountName, priv, systemassetid, math.MaxUint64, true, chainid)
		acct2 := NewAccount(api, accountName2, priv2, systemassetid, math.MaxUint64, true, chainid)
		hash, err = acct.RegCandidate(common.StrToName(dposaccount), new(big.Int).Div(tValue, big.NewInt(3)), systemassetid, tGas, &dpos.RegisterCandidate{
			URL: fmt.Sprintf("www.%s.com", accountName.String()),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// VoteCandidate
		hash, err = acct2.VoteCandidate(common.StrToName(dposaccount), new(big.Int).Div(tValue, big.NewInt(3)), systemassetid, tGas, &dpos.VoteCandidate{
			Candidate: accountName.String(),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// VoteCandidate
		hash, err = acct2.VoteCandidate(common.StrToName(dposaccount), new(big.Int).Div(tValue, big.NewInt(3)), systemassetid, tGas, &dpos.VoteCandidate{
			Candidate: systemaccount,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		hash, err = sysAcct.KickedCandidate(common.StrToName(dposaccount), new(big.Int).Mul(tValue, big.NewInt(0)), systemassetid, tGas, &dpos.KickedCandidate{
			Candidates: []string{accountName.String()},
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// UnRegCandidate
		hash, err = acct.UnRegCandidate(common.StrToName(dposaccount), new(big.Int).Mul(tValue, big.NewInt(0)), systemassetid, tGas)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func TestManual(t *testing.T) {
	SkipConvey("Manual", t, func() {
		api := NewAPI(rpchost)
		var systempriv, _ = crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)

		hash, err := sysAcct.KickedCandidate(common.StrToName(dposaccount), new(big.Int).Mul(tValue, big.NewInt(0)), systemassetid, tGas, &dpos.KickedCandidate{
			Candidates: []string{"ftcandidate1", "ftcandidate2", "ftcandidate3"},
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func createAccount(sysAcct *Account, api *API) (*Account, error) {
	priv, pub := GenerateKey()
	accountName := common.StrToName(GenerateAccountName("test", 8))
	if _, err := sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.CreateAccountAction{
		AccountName: accountName,
		PublicKey:   pub,
	}); err != nil {
		return nil, err
	}
	return NewAccount(api, accountName, priv, systemassetid, math.MaxUint64, true, chainid), nil
}

func TestContract(t *testing.T) {
	Convey("Contract", t, func() {
		api := NewAPI(rpchost)
		var systempriv, _ = crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)

		// CreateAccount
		acct, err := createAccount(sysAcct, api)
		So(err, ShouldBeNil)

		// deploy contract ./test/asset.sol
		input, err := formCreateContractInput(AssetAbi, AssetBin)
		So(err, ShouldBeNil)
		hash, err := acct.CreateContract(systemassetid, tGas, input)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// issue asset in contract
		assetName := GenerateAccountName("test", 8)
		input, err = formIssueAssetInput(AssetAbi, assetName+","+assetName+",10000000000,10,"+acct.name.String()+",20000000000,"+acct.name.String()+",,this is contracgt asset")
		So(err, ShouldBeNil)
		hash, err = acct.CallContract(systemassetid, tGas, input)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
		ast, err := api.AssetInfoByName(assetName)
		So(err, ShouldBeNil)
		So(ast.Owner, ShouldEqual, acct.name) // compare name

		// increase asset in contract
		accountInfo, err := api.AccountInfo(acct.name.String())
		So(err, ShouldBeNil)
		balance, err := accountInfo.GetBalanceByID(ast.AssetId)
		So(err, ShouldBeNil)
		increment := big.NewInt(100000)
		input, err = formIncreaseAssetInput(AssetAbi, big.NewInt(int64(ast.GetAssetId())),
			common.BigToAddress(big.NewInt(int64(accountInfo.AccountID))), increment)
		So(err, ShouldBeNil)
		hash, err = acct.CallContract(systemassetid, tGas, input)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		newAsset, err := api.AssetInfoByName(assetName)
		So(err, ShouldBeNil)
		So(big.NewInt(0).Add(ast.Amount, increment), ShouldResemble, newAsset.Amount) // compare asset amount

		newAccountInfo, err := api.AccountInfo(acct.name.String())
		So(err, ShouldBeNil)
		newBalance, err := newAccountInfo.GetBalanceByID(ast.AssetId)
		So(err, ShouldBeNil)
		So(big.NewInt(0).Add(balance, increment), ShouldResemble, newBalance) // compare account blanace

		// transfer asset in contract
		toAcct, err := createAccount(sysAcct, api)
		So(err, ShouldBeNil)
		toAcctInfo, err := api.AccountInfo(toAcct.name.String())
		So(err, ShouldBeNil)
		input, err = formTransferAssetInput(AssetAbi, big.NewInt(int64(ast.AssetId)), common.BigToAddress(big.NewInt(int64(toAcctInfo.AccountID))), big.NewInt(1))
		So(err, ShouldBeNil)
		hash, err = acct.CallContract(systemassetid, tGas, input)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		sendAccountInfo, err := api.AccountInfo(acct.name.String())
		So(err, ShouldBeNil)
		senderBalance, err := sendAccountInfo.GetBalanceByID(ast.AssetId)
		So(err, ShouldBeNil)
		So(newBalance.Sub(newBalance, big.NewInt(1)), ShouldResemble, senderBalance) // compare sender blanace

		recipientAccountInfo, err := api.AccountInfo(toAcct.name.String())
		So(err, ShouldBeNil)
		recipientBalance, err := recipientAccountInfo.GetBalanceByID(ast.AssetId)
		So(err, ShouldBeNil)
		So(big.NewInt(1), ShouldResemble, recipientBalance) // compare recipient blanace

		// change asset owner in contract
		input, err = formChangeAssetOwner(AssetAbi, common.BigToAddress(big.NewInt(int64(toAcctInfo.AccountID))), big.NewInt(int64(ast.AssetId))) //22168
		So(err, ShouldBeNil)
		hash, err = acct.CallContract(systemassetid, tGas, input)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
		newOwnerAsset, err := api.AssetInfoByName(assetName)
		So(err, ShouldBeNil)
		So(newOwnerAsset.Owner, ShouldEqual, toAcct.name) // compare asset owner

		// destory asset in contract
		input, err = formDestroyAsset(AssetAbi, big.NewInt(int64(ast.AssetId)), senderBalance)
		So(err, ShouldBeNil)
		hash, err = acct.CallContract(systemassetid, tGas, input)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		destroyAccountInfo, err := api.AccountInfo(acct.name.String())
		So(err, ShouldBeNil)
		destroyBalance, err := destroyAccountInfo.GetBalanceByID(ast.AssetId)
		So(err, ShouldBeNil)
		So(big.NewInt(0), ShouldResemble, destroyBalance) // compare destory balance
	})
}
