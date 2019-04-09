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
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
	jww "github.com/spf13/jwalterweatherman"

	tc "github.com/fractalplatform/fractal/test/common"
)

var (
	minerprikey, _   = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	minerpubkey      = common.HexToPubKey("0x047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd")
	newPrivateKey, _ = crypto.HexToECDSA("8ee847ae5974a13ce9df66083e453ea1e0f7995379ed027a98e827aa8b6bc211")
	gaslimit         = uint64(2000000)
	minername        = common.Name("ftsystemio")
	toname           = common.Name("testtest11")
	issueAmount      = new(big.Int).Mul(big.NewInt(10), big.NewInt(1e10))
	inCreateAmount   = big.NewInt(100000000)
	indexstr         = "abcdefghijklmnopqrstuvwxyz0123456789"
	basefrom         = "newnamefrom%s"
	baseto           = "newnameto%s"
	testbase         = "testtest"
	testname1        = ""
)

type GenAction struct {
	*types.Action
	PrivateKey *ecdsa.PrivateKey
}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)
}

func GeneragePubKey() (common.PubKey, *ecdsa.PrivateKey) {
	prikey, _ := crypto.GenerateKey()
	return common.BytesToPubKey(crypto.FromECDSAPub(&prikey.PublicKey)), prikey
}

func createAccount(accountName common.Name, founder common.Name, from, newname common.Name, nonce uint64, publickey common.PubKey, prikey *ecdsa.PrivateKey) {
	account := &accountmanager.AccountAction{
		AccountName: accountName,
		Founder:     founder,
		ChargeRatio: 80,
		PublicKey:   publickey,
	}
	payload, err := rlp.EncodeToBytes(account)
	if err != nil {
		panic("rlp payload err")
	}
	to := newname
	gc := newGeAction(types.CreateAccount, from, to, nonce, 1, gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	sendTxTest(gcs)
}

func updateAccount(from, founder common.Name, nonce uint64, privatekey *ecdsa.PrivateKey) {
	account := &accountmanager.AccountAction{
		AccountName: from,
		Founder:     founder,
		ChargeRatio: 80,
	}
	payload, err := rlp.EncodeToBytes(account)
	if err != nil {
		panic("rlp payload err")
	}
	gc := newGeAction(types.UpdateAccount, from, "", nonce, 1, gaslimit, nil, payload, privatekey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	sendTxTest(gcs)
}

func updateAccountAuthor(from common.Name, oldpubkey, newpubkey common.PubKey, nonce uint64, privatekey *ecdsa.PrivateKey) {
	oldauthor := &common.Author{
		Owner: oldpubkey,
	}
	oldauthoraction := &accountmanager.AuthorAction{accountmanager.DeleteAuthor, oldauthor}
	newauthor := &common.Author{
		Owner:  newpubkey,
		Weight: 1,
	}
	newauthoraction := &accountmanager.AuthorAction{accountmanager.AddAuthor, newauthor}
	accountauthor := &accountmanager.AccountAuthorAction{
		AuthorActions: []*accountmanager.AuthorAction{oldauthoraction, newauthoraction},
	}
	payload, err := rlp.EncodeToBytes(accountauthor)
	if err != nil {
		panic("rlp payload err")
	}
	gc := newGeAction(types.UpdateAccountAuthor, from, "", nonce, 1, gaslimit, nil, payload, privatekey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	sendTxTest(gcs)
}

func issueAsset(atype types.ActionType, assetid uint64, from, Owner, founder common.Name, amount *big.Int, assetname string, nonce uint64, prikey *ecdsa.PrivateKey) {
	ast := &asset.AssetObject{
		AssetId:    assetid,
		AssetName:  assetname,
		Symbol:     fmt.Sprintf("symbol%d", nonce),
		Amount:     amount,
		Decimals:   2,
		Founder:    founder,
		AddIssue:   nil,
		Owner:      Owner,
		UpperLimit: big.NewInt(100000000000000000),
	}
	payload, err := rlp.EncodeToBytes(ast)
	if err != nil {
		panic("rlp payload err")
	}
	gc := newGeAction(atype, from, "", nonce, 1, gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	sendTxTest(gcs)
}

func increaseAsset(from, to common.Name, assetid uint64, nonce uint64, prikey *ecdsa.PrivateKey) {
	ast := &accountmanager.IncAsset{
		AssetId: assetid,
		To:      to,
		Amount:  inCreateAmount,
	}
	payload, err := rlp.EncodeToBytes(ast)
	if err != nil {
		panic("rlp payload err")
	}
	gc := newGeAction(types.IncreaseAsset, from, "", nonce, 1, gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	sendTxTest(gcs)
}

func transfer(from, to common.Name, amount *big.Int, nonce uint64, prikey *ecdsa.PrivateKey) {
	gc := newGeAction(types.Transfer, from, to, nonce, 1, gaslimit, amount, nil, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	sendTxTest(gcs)
}

func newGeAction(at types.ActionType, from, to common.Name, nonce uint64, assetid uint64, gaslimit uint64, amount *big.Int, payload []byte, prikey *ecdsa.PrivateKey) *GenAction {
	action := types.NewAction(at, from, to, nonce, assetid, gaslimit, amount, payload)

	return &GenAction{
		Action:     action,
		PrivateKey: prikey,
	}
}

func sendTxTest(gcs []*GenAction) {
	signer := types.NewSigner(big.NewInt(1))
	var actions []*types.Action
	for _, v := range gcs {
		actions = append(actions, v.Action)
	}
	tx := types.NewTransaction(uint64(1), big.NewInt(1), actions...)
	for _, v := range gcs {
		keypair := types.MakeKeyPair(v.PrivateKey, []uint64{0})
		err := types.SignActionWithMultiKey(v.Action, tx, signer, []*types.KeyPair{keypair})
		if err != nil {
			panic(fmt.Sprintf("SignAction err %v", err))
		}

	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}
	hash, err := tc.SendRawTx(rawtx)
	if err != nil {
		panic(err)
	}
	jww.INFO.Printf("hash: %x", hash)
}

var (
	pub1 = "0x0468cba7890aae10f3dde57d269cf7c4ba14cc0efc2afee86791b0a22b794820febdb2e5c6c56878a308e7f62ad2d75739de40313a72975c993dd76a5301a03d12"
	pri1 = "357a2cbdd91686dcbe2c612e9bed85d4415f62446440839466bf7b2f1ab135b7"

	pub2 = "0x04fa0b2a9b2d0542bf2912c4c6500ba64a26652e302370ed5645b1c32df50fbe7a5f12da0b278638e1df6753a7c6ac09e68cb748cfe6d45102114f52e95e9ed652"
	pri2 = "340cde826336f1adb8673ec945819d073af00cffb5c174542e35ff346445e213"

	pubkey1    = common.HexToPubKey(pub1)
	prikey1, _ = crypto.HexToECDSA(pri1)

	pubkey2    = common.HexToPubKey(pub2)
	prikey2, _ = crypto.HexToECDSA(pri2)
)

func main() {
	nonce, _ := tc.GetNonce(minername)
	//pub, pri := GeneragePubKey()
	createAccount(toname, "", minername, toname, nonce, pubkey1, minerprikey)
	nonce++
	transfer(minername, toname, issueAmount, nonce, minerprikey)
	nonce++

	newname2 := common.Name("testtest12")
	//pub1, _ := GeneragePubKey()
	createAccount(newname2, "", minername, newname2, nonce, pubkey2, minerprikey)
	nonce++

	transfer(minername, newname2, issueAmount, nonce, minerprikey)
	nonce++

	pubkey3, prikey3 := GeneragePubKey()
	newname3 := common.Name("testtest13")
	createAccount(newname3, "", minername, newname3, nonce, pubkey3, minerprikey)
	nonce++

	transfer(minername, newname3, issueAmount, nonce, minerprikey)
	nonce++

	time.Sleep(time.Duration(3) * time.Second)

	t3nonce, _ := tc.GetNonce(newname3)
	updateAccount(newname3, newname2, t3nonce, prikey3)
	time.Sleep(time.Duration(3) * time.Second)

	t3nonce, _ = tc.GetNonce(newname3)
	updateAccountAuthor(newname3, pubkey3, pubkey2, t3nonce, prikey3)
	time.Sleep(time.Duration(3) * time.Second)

	t3nonce, _ = tc.GetNonce(newname3)
	issueAsset(types.IssueAsset, 0, newname3, newname3, newname3, big.NewInt(10000000000000), "testnewasset", t3nonce, prikey2)
	time.Sleep(time.Duration(3) * time.Second)

	t3nonce++
	increaseAsset(newname3, newname3, 2, t3nonce, prikey2)
	time.Sleep(time.Duration(3) * time.Second)

	t3nonce++
	issueAsset(types.SetAssetOwner, 2, newname3, toname, "", big.NewInt(100000), "testnewasset", t3nonce, prikey2)
	time.Sleep(time.Duration(3) * time.Second)

	t3nonce++
	issueAsset(types.DestroyAsset, 2, toname, "", "", big.NewInt(100000), "testnewasset", t3nonce, prikey1)

}
