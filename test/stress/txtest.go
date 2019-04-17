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
	"flag"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	tc "github.com/fractalplatform/fractal/test/common"
	testcommon "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	minerprikey, _   = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	minerpubkey      = common.HexToPubKey("0x047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd")
	newPrivateKey, _ = crypto.HexToECDSA("8ee847ae5974a13ce9df66083e453ea1e0f7995379ed027a98e827aa8b6bc211")
	gaslimit         = uint64(2000000)
	minername        = common.Name("ftsystemio")
	toname           = common.Name("testtest1")
	issueAmount      = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e1))
	inCreateAmount   = big.NewInt(1)
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

func createAccount(accountName common.Name, founder common.Name, from, newname common.Name, nonce uint64, prikey *ecdsa.PrivateKey, pubkey common.PubKey) {
	account := &accountmanager.AccountAction{
		AccountName: accountName,
		Founder:     founder,
		ChargeRatio: 80,
		PublicKey:   pubkey,
	}
	payload, err := rlp.EncodeToBytes(account)
	if err != nil {
		panic("rlp payload err")
	}
	key := types.MakeKeyPair(prikey, []uint64{0})
	sendTransferTx(types.CreateAccount, from, newname, nonce, 1, nil, payload, []*types.KeyPair{key})
}

func updateAccount(accountName common.Name, founder common.Name, from, newname common.Name, nonce uint64, prikey *ecdsa.PrivateKey, pubkey common.PubKey) {
	account := &accountmanager.AccountAction{
		AccountName: accountName,
		Founder:     founder,
		ChargeRatio: 80,
		PublicKey:   pubkey,
	}
	payload, err := rlp.EncodeToBytes(account)
	if err != nil {
		panic("rlp payload err")
	}
	key := types.MakeKeyPair(prikey, []uint64{0})
	sendTransferTx(types.UpdateAccount, from, newname, nonce, 1, nil, payload, []*types.KeyPair{key})
}

func transfer(from, to common.Name, amount *big.Int, nonce uint64, prikey *ecdsa.PrivateKey) {
	key := types.MakeKeyPair(prikey, []uint64{0})
	sendTransferTx(types.Transfer, from, to, nonce, 1, amount, nil, []*types.KeyPair{key})
}

func sendTransferTx(txType types.ActionType, from, to common.Name, nonce, assetID uint64, value *big.Int, input []byte, keys []*types.KeyPair) {
	action := types.NewAction(txType, from, to, nonce, assetID, gaslimit, value, input)
	//gasprice := big.NewInt(2)
	gasprice, _ := testcommon.GasPrice()
	tx := types.NewTransaction(1, gasprice, action)

	signer := types.MakeSigner(big.NewInt(1))
	err := types.SignActionWithMultiKey(action, tx, signer, keys)
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

func main() {

	var timecount int
	flag.IntVar(&timecount, "timecount", 0, "this is time")
	toname1 := flag.String("to", "testtest1", "toname")
	minername1 := flag.String("from", "ftsystemio", "fromname")
	flag.Parse()
	nonce, _ := tc.GetNonce(common.Name(*minername1))
	createAccount(common.Name(*toname1), "", common.Name(*minername1), common.Name(*toname1), nonce, minerprikey, minerpubkey)
	nonce++
	for {

		jww.INFO.Println(nonce)

		transfer(common.Name(*minername1), common.Name(*toname1), issueAmount, nonce, minerprikey)
		time.Sleep(time.Duration(timecount) * time.Millisecond)
		nonce++
		//updateAccount(toname,toname,minername, toname, nonce, minerprikey, minerpubkey)

	}

}
