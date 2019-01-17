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
	//"bytes"
	//"io/ioutil"
	"math/big"
	//"strings"
	"fmt"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	testcommon "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	minproducerstake = big.NewInt(0).Mul(dpos.DefaultConfig.UnitStake, dpos.DefaultConfig.ProducerMinQuantity)
	minvoterstake    = big.NewInt(0).Mul(dpos.DefaultConfig.UnitStake, dpos.DefaultConfig.VoterMinQuantity)
	big1             = big.NewInt(1)

	privateKey, _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	from          = common.Name("ftsystemio")
	p1            = common.Name("producer1")
	p2            = common.Name("producer2")
	newFrom       = common.Name("newfromname")
	noNeedUser    = common.Name("")
	noNeedNum     = big.NewInt(0)

	assetID = uint64(1)

	nonce = uint64(0)

	gasLimit = uint64(2000000)
)

func generateAccount() {
	nonce, _ = testcommon.GetNonce(from)

	newPrivateKey, _ := crypto.GenerateKey()
	pubKey := common.BytesToPubKey(crypto.FromECDSAPub(&newPrivateKey.PublicKey))

	balance, _ := testcommon.GetAccountBalanceByID(from, assetID)
	balance.Div(balance, big.NewInt(10))

	newFrom = common.Name(fmt.Sprintf("newfromname%d", nonce))
	//contractAddr = common.Name(fmt.Sprintf("multiasset%d", nonce))

	sendTransferTx(types.CreateAccount, from, newFrom, nonce, assetID, balance, pubKey.Bytes())
	nonce++
	sendTransferTx(types.CreateAccount, from, p1, nonce, assetID, balance, pubKey.Bytes())
	nonce++
	sendTransferTx(types.CreateAccount, from, p2, nonce, assetID, balance, pubKey.Bytes())

	for {
		time.Sleep(3 * time.Second)
		bnf, _ := testcommon.CheckAccountIsExist(newFrom)
		bp1, _ := testcommon.CheckAccountIsExist(p1)
		bp2, _ := testcommon.CheckAccountIsExist(p2)
		if bnf && bp1 && bp2 {
			jww.INFO.Println("account created... ")
			break
		}
	}

	from = newFrom
	privateKey = newPrivateKey
}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)

	generateAccount()
}

func main() {
	jww.INFO.Println("Test dpos votes transactions...")

	commonUnregProducer()

	sendRegProducer()
	sendUpdateProducer()
	sendVoteProducer()

	for {
		// check result
		time.Sleep(3 * time.Second)

		p1_fields, err := testcommon.GetDposAccount(p1)
		p2_fields, _ := testcommon.GetDposAccount(p2)
		if true {
			jww.INFO.Println("p1:", p1_fields)
			jww.INFO.Println("p2:", p2_fields)
			jww.INFO.Println("err", err)
			break
		}
	}
	sendChangeProducer()
	//
	sendUnvoteProducer()
	//
	for {
		// check result
		time.Sleep(3 * time.Second)

		p1_fields, _ := testcommon.GetDposAccount(p1)
		p2_fields, _ := testcommon.GetDposAccount(p2)
		if true {
			jww.INFO.Println("p1:", p1_fields)
			jww.INFO.Println("p2:", p2_fields)
			break
		}
	}
	sendVoteProducer()
	sendRemoveVoter()
	//
	for {
		// check result
		time.Sleep(3 * time.Second)

		p1_fields, _ := testcommon.GetDposAccount(p1)
		p2_fields, _ := testcommon.GetDposAccount(p2)
		if true {
			jww.INFO.Println("p1:", p1_fields)
			jww.INFO.Println("p2:", p2_fields)
			break
		}
	}
	sendUnregProducer()
}

func sendRegProducer() {
	jww.INFO.Println("test sendRegProducer... ")

	rp := dpos.RegisterProducer{
		Url:   "https://www.test_url.xxx/",
		Stake: minproducerstake,
	}

	rawdata, err := rlp.EncodeToBytes(rp)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	nonce++
	sendTransferTx(types.RegProducer, p1, noNeedUser, nonce, assetID, noNeedNum, rawdata)
	nonce++
	sendTransferTx(types.RegProducer, p2, noNeedUser, nonce, assetID, noNeedNum, rawdata)
	jww.INFO.Println("test sendRegProducer... done ")
}

func sendUnregProducer() {
	jww.INFO.Println("test sendUnregProducer... ")

	nonce++
	sendTransferTx(types.UnregProducer, p1, noNeedUser, nonce, assetID, noNeedNum, nil)
	nonce++
	sendTransferTx(types.UnregProducer, p2, noNeedUser, nonce, assetID, noNeedNum, nil)
	jww.INFO.Println("test sendUnregProducer... done ")
}

func sendUpdateProducer() {
	jww.INFO.Println("test sendUpdateProducer... ")

	rp := dpos.UpdateProducer{
		Url:   "https://www.test_url_p1.xxx/",
		Stake: minproducerstake,
	}

	rawdata, err := rlp.EncodeToBytes(rp)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}
	nonce++
	sendTransferTx(types.UpdateProducer, p1, noNeedUser, nonce, assetID, noNeedNum, rawdata)
	jww.INFO.Println("test sendUpdateProducer... done ")
}

func sendRemoveVoter() {
	jww.INFO.Println("test sendRemoveVoter... ")

	rp := dpos.RemoveVoter{
		Voter: from.String(),
	}

	rawdata, err := rlp.EncodeToBytes(rp)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	nonce++
	sendTransferTx(types.RemoveVoter, p1, noNeedUser, nonce, assetID, noNeedNum, rawdata)
	jww.INFO.Println("test sendRemoveVoter... done ")
}

func sendVoteProducer() {
	jww.INFO.Println("test sendVoteProducer... ")

	rp := dpos.VoteProducer{
		Producer: p1.String(),
		Stake:    minvoterstake,
	}

	rawdata, err := rlp.EncodeToBytes(rp)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}
	nonce++
	sendTransferTx(types.VoteProducer, from, noNeedUser, nonce, assetID, noNeedNum, rawdata)
	jww.INFO.Println("test sendVoteProducer... done ")
}

func sendChangeProducer() {
	jww.INFO.Println("test sendChangeProducer... ")

	rp := dpos.ChangeProducer{
		Producer: p2.String(),
	}

	rawdata, err := rlp.EncodeToBytes(rp)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}
	nonce++
	sendTransferTx(types.ChangeProducer, from, noNeedUser, nonce, assetID, noNeedNum, rawdata)
	jww.INFO.Println("test sendChangeProducer... done")
}

func sendUnvoteProducer() {
	jww.INFO.Println("test sendUnvoteProducer... ")

	nonce++
	sendTransferTx(types.UnvoteProducer, from, noNeedUser, nonce, assetID, noNeedNum, nil)
	jww.INFO.Println("test sendUnvoteProducer... done ")
}

// sendTransferTx
func sendTransferTx(txType types.ActionType, from, to common.Name, nonce, assetID uint64, value *big.Int, input []byte) {
	action := types.NewAction(txType, from, to, nonce, assetID, gasLimit, value, input)
	gp, _ := testcommon.GasPrice()
	tx := types.NewTransaction(assetID, gp, action)

	signer := types.MakeSigner(big.NewInt(1))
	err := types.SignAction(action, tx, signer, privateKey)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}

	if hash, err := testcommon.SendRawTx(rawtx); err != nil {
		jww.ERROR.Fatalln("send failed: ", err)
	} else {
		jww.INFO.Println("result hash: ", hash.Hex())
	}
}

func commonUnregProducer() {
	jww.INFO.Println("test commonUnregProducer... ")

	acct := testcommon.NewAccount(p1, privateKey, 1, 0, nil)

	rawtx := acct.UnRegProducer(p1, big1, 1, gasLimit)

	if hash, err := testcommon.SendRawTx(rawtx); err != nil {
		jww.ERROR.Fatalln("send failed: h", err, hash.Hex())
	} else {
		jww.INFO.Println("result hash: ", hash.Hex())
	}

	jww.INFO.Println("test commonUnregProducer... done ")
}
