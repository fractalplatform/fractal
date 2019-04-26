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
	systemprivkey   = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
	systemassetname = params.DefaultChainconfig.SysToken
	systemassetid   = uint64(1)
	chainid         = big.NewInt(1)
	tValue          = new(big.Int).Mul(big.NewInt(300000), big.NewInt(1e18))
	tGas            = uint64(90000)
)

func TestAccount(t *testing.T) {
	Convey("Account", t, func() {
		api := NewAPI(rpchost)
		var systempriv, _ = crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)
		// CreateAccount
		priv, pub := GenerateKey()
		accountName := common.StrToName(GenerateAccountName("test", 8))
		hash, err := sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.AccountAction{
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
		_, npub := GenerateKey()
		hash, err = acct.UpdateAccount(common.StrToName(accountaccount), new(big.Int).Mul(tValue, big.NewInt(0)), systemassetid, tGas, &accountmanager.AccountAction{
			AccountName: accountName,
			PublicKey:   npub,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		// DestroyAccount
	})
}

func TestAsset(t *testing.T) {
	// Convey("types.IssueAsset", t, func() {
	// 	api := NewAPI(rpchost)
	// 	var systempriv, _ = crypto.HexToECDSA(systemprivkey)
	// 	sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)
	// 	priv, pub := GenerateKey()
	// 	accountName := common.StrToName(GenerateAccountName("test", 8))
	// 	hash, err := sysAcct.CreateAccount(common.StrToName(systemaccount), tValue, systemassetid, tGas, &accountmanager.AccountAction{
	// 		AccountName: accountName,
	// 		PublicKey:   pub,
	// 	})
	// 	So(err, ShouldBeNil)
	// 	So(hash, ShouldNotBeNil)

	// 	acct := NewAccount(api, accountName, priv, systemassetid, math.MaxUint64, true, chainid)
	// 	assetname := common.StrToName(GenerateAccountName("asset", 8)).String()
	// 	// IssueAsset
	// 	hash, err = acct.IssueAsset(accountName, new(big.Int).Div(tValue, big.NewInt(10)), systemassetid, tGas, &asset.AssetObject{
	// 		AssetName:  assetname,
	// 		Symbol:     assetname[len(assetname)-4:],
	// 		Amount:     new(big.Int).Mul(big.NewInt(10000000), big.NewInt(1e18)),
	// 		Decimals:   18,
	// 		Owner:      accountName,
	// 		Founder:    accountName,
	// 		AddIssue:   big.NewInt(0),
	// 		UpperLimit: big.NewInt(0),
	// 	})
	// 	So(err, ShouldBeNil)
	// 	So(hash, ShouldNotBeNil)

	// 	// acct.UpdateAsset()
	// 	// acct.IncreaseAsset()
	// 	// acct.SetAssetOwner()
	// 	// acct.DestroyAsset()
	// })
}

func TestDPOS(t *testing.T) {
	SkipConvey("DPOS", t, func() {
		api := NewAPI(rpchost)
		var systempriv, _ = crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(api, common.StrToName(systemaccount), systempriv, systemassetid, math.MaxUint64, true, chainid)
		priv, pub := GenerateKey()
		accountName := common.StrToName(GenerateAccountName("prod", 8))
		hash, err := sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.AccountAction{
			AccountName: accountName,
			PublicKey:   pub,
		})
		So(err, ShouldBeNil)
		So(hash, ShouldNotBeNil)

		priv2, pub2 := GenerateKey()
		accountName2 := common.StrToName(GenerateAccountName("voter", 8))
		hash, err = sysAcct.CreateAccount(common.StrToName(accountaccount), tValue, systemassetid, tGas, &accountmanager.AccountAction{
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
