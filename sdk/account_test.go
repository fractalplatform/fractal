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

	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/params"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	systemaccount  = params.DefaultChainconfig.SysName
	accountaccount = params.DefaultChainconfig.AccountName
	systemassetid  = uint64(0)
	chainid        = big.NewInt(1)
	tValue         = new(big.Int).Mul(big.NewInt(300000), big.NewInt(1e18))
	tGas           = uint64(20000000)

	AssetAbi   = "./test/contract/Asset.abi"
	AssetBin   = "./test/contract/Asset.bin"
	api        = NewAPI("http://127.0.0.1:8545")
	syspriv, _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	decimals   = big.NewInt(1)
	chainCfg   *params.ChainConfig
	sysAct     *Account
)

var (
	val  = big.NewInt(100)
	gas  = uint64(30000000)
	name = GenerateAccountName("sdktest", 8)
	priv = syspriv

	astname   = GenerateAccountName("sdkasset", 8)
	astsymbol = "sas"
	astamount = big.NewInt(100000000000)
)

func init() {
	cfg, err := api.GetChainConfig()
	if err != nil {
		panic(fmt.Sprintf("init err %v", err))
	}
	chainCfg = cfg
	for i := uint64(0); i < chainCfg.SysTokenDecimals; i++ {
		decimals = new(big.Int).Mul(decimals, big.NewInt(10))
	}
	sysAct = NewAccount(api, common.StrToName(chainCfg.SysName), syspriv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
}

func TestCreateAccount(t *testing.T) {
	Convey("CreateAccount", t, func() {
		pub := common.BytesToPubKey(crypto.FromECDSAPub(&priv.PublicKey))
		hash, err := sysAct.CreateAccount(common.StrToName(chainCfg.AccountName), new(big.Int).Mul(val, decimals), chainCfg.SysTokenID, gas, &accountmanager.CreateAccountAction{
			AccountName: common.StrToName(name),
			PublicKey:   pub,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func TestUpdateAccount(t *testing.T) {
	Convey("UpdateAccount", t, func() {
		act := NewAccount(api, common.StrToName(name), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
		hash, err := act.UpdateAccount(common.StrToName(chainCfg.AccountName), big.NewInt(0), chainCfg.SysTokenID, gas, &accountmanager.UpdataAccountAction{
			Founder: common.StrToName(name),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func TestUpdateAccountAuthor(t *testing.T) {
	Convey("UpdateAccountAuthor", t, func() {
		act := NewAccount(api, common.StrToName(name), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
		hash, err := act.UpdateAccountAuthor(common.StrToName(chainCfg.AccountName), big.NewInt(0), chainCfg.SysTokenID, gas, &accountmanager.AccountAuthorAction{
			Threshold:             1,
			UpdateAuthorThreshold: 1,
			AuthorActions: []*accountmanager.AuthorAction{
				&accountmanager.AuthorAction{
					ActionType: accountmanager.UpdateAuthor,
					Author: &common.Author{
						Owner:  act.Pubkey(),
						Weight: 1,
					},
				},
			},
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func TestIssueAsset(t *testing.T) {
	Convey("IssueAsset", t, func() {
		hash, err := sysAct.IssueAsset(common.StrToName(chainCfg.AssetName), big.NewInt(0), chainCfg.SysTokenID, gas, &accountmanager.IssueAsset{
			AssetName: astname,
			Symbol:    astsymbol,
			Amount:    new(big.Int).Mul(astamount, decimals),
			Decimals:  chainCfg.SysTokenDecimals,
			Founder:   common.StrToName(chainCfg.SysName),
			Owner:     common.StrToName(chainCfg.SysName),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestUpdateAsset(t *testing.T) {
	Convey("UpdateAsset", t, func() {
		ast, err := api.AssetInfoByName(astname)
		So(err, ShouldBeNil)

		hash, err := sysAct.UpdateAsset(common.StrToName(chainCfg.AssetName), big.NewInt(0), chainCfg.SysTokenID, gas, &accountmanager.UpdateAsset{
			AssetID: ast.AssetId,
			Founder: common.StrToName(chainCfg.SysName),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestSetAssetOwner(t *testing.T) {
	Convey("SetAssetOwner", t, func() {
		ast, err := api.AssetInfoByName(astname)
		So(err, ShouldBeNil)

		hash, err := sysAct.SetAssetOwner(common.StrToName(chainCfg.AssetName), big.NewInt(0), chainCfg.SysTokenID, gas, &accountmanager.UpdateAssetOwner{
			AssetID: ast.AssetId,
			Owner:   common.StrToName(chainCfg.SysName),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestDestroyAsset(t *testing.T) {
	Convey("DestroyAsset", t, func() {
		ast, err := api.AssetInfoByName(astname)
		So(err, ShouldBeNil)

		hash, err := sysAct.DestroyAsset(common.StrToName(chainCfg.AssetName), new(big.Int).Mul(astamount, decimals), ast.AssetId, gas)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestIncreaseAsset(t *testing.T) {
	Convey("IncreaseAsset", t, func() {
		ast, err := api.AssetInfoByName(astname)
		So(err, ShouldBeNil)

		hash, err := sysAct.IncreaseAsset(common.StrToName(chainCfg.AssetName), big.NewInt(0), chainCfg.SysTokenID, gas, &accountmanager.IncAsset{
			AssetId: ast.AssetId,
			Amount:  new(big.Int).Mul(astamount, decimals),
			To:      common.StrToName(chainCfg.SysName),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestTransfer(t *testing.T) {
	Convey("Transfer", t, func() {
		hash, err := sysAct.Transfer(common.StrToName(name), val, chainCfg.SysTokenID, gas)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func TestRegCandidate(t *testing.T) {
	SkipConvey("RegCandidate", t, func() {
		// RegCandidate
		act := NewAccount(api, common.StrToName(name), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
		hash, err := act.RegCandidate(common.StrToName(chainCfg.DposName), new(big.Int).Mul(new(big.Int).Div(val, big.NewInt(4)), decimals), chainCfg.SysTokenID, gas, &dpos.RegisterCandidate{
			URL: fmt.Sprintf("www.%s.com", name),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestUpdateCandidate(t *testing.T) {
	SkipConvey("UpdateCandidate", t, func() {
		// UpdateCandidate
		act := NewAccount(api, common.StrToName(name), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
		hash, err := act.UpdateCandidate(common.StrToName(chainCfg.DposName), new(big.Int).Mul(new(big.Int).Div(val, big.NewInt(4)), decimals), chainCfg.SysTokenID, gas, &dpos.UpdateCandidate{
			URL: fmt.Sprintf("www.%s.com", name),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestUnRegCandidate(t *testing.T) {
	SkipConvey("UnRegCandidate", t, func() {
		// UnRegCandidate
		act := NewAccount(api, common.StrToName(name), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
		hash, err := act.UnRegCandidate(common.StrToName(chainCfg.DposName), big.NewInt(0), chainCfg.SysTokenID, gas)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func TestRefundCandidate(t *testing.T) {
	SkipConvey("RefundCandidate", t, func() {
		// RefundCandidate
		act := NewAccount(api, common.StrToName(name), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
		hash, err := act.RefundCandidate(common.StrToName(chainCfg.DposName), big.NewInt(0), chainCfg.SysTokenID, gas)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestVoteCandidate(t *testing.T) {
	SkipConvey("VoteCandidate", t, func() {
		// VoteCandidate
		act := NewAccount(api, common.StrToName(name), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
		hash, err := act.VoteCandidate(common.StrToName(chainCfg.DposName), big.NewInt(0), chainCfg.SysTokenID, gas, &dpos.VoteCandidate{
			Candidate: chainCfg.SysName,
			Stake:     new(big.Int).Mul(new(big.Int).Div(val, big.NewInt(4)), decimals),
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}
func TestKickedCandidate(t *testing.T) {
	SkipConvey("KickedCandidate", t, func() {
		// KickedCandidate
		hash, err := sysAct.KickedCandidate(common.StrToName(chainCfg.DposName), big.NewInt(0), chainCfg.SysTokenID, gas, &dpos.KickedCandidate{
			Candidates: []string{name},
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func TestExitTakeOver(t *testing.T) {
	SkipConvey("ExitTakeOver", t, func() {
		// ExitTakeOver
		hash, err := sysAct.ExitTakeOver(common.StrToName(chainCfg.DposName), big.NewInt(0), chainCfg.SysTokenID, gas)
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)
	})
}

func createAccount(sysAcct *Account, api *API) (*Account, error) {
	priv := GenerateKey()
	pub := common.BytesToPubKey(crypto.FromECDSAPub(&priv.PublicKey))
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
		sysAcct := NewAccount(api, common.StrToName(systemaccount), syspriv, systemassetid, math.MaxUint64, true, chainid)

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
