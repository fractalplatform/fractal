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

	"os"
	"strconv"
	"sync"

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	tc "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	minerprikey []*ecdsa.PrivateKey
	minername   []common.Name
	ipc         []string
	index       int

	basefrom []string
	baseto   []string

	ftproducer1key, _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	ftproducer1name   = common.Name("ftproducer1")

	ftproducer2key, _ = crypto.HexToECDSA("9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658")
	ftproducer2name   = common.Name("ftproducer2")

	ftproducer3key, _ = crypto.HexToECDSA("8605cf6e76c9fc8ac079d0f841bd5e99bd3ad40fdd56af067993ed14fc5bfca8")
	ftproducer3name   = common.Name("ftproducer3")

	gaslimit       = uint64(200000)
	issueAmount    = big.NewInt(100000000000000)
	inCreateAmount = big.NewInt(100000000)
	indexstr       = "abcdefghijklmnopqrstuvwxyz0123456789"
)

type GenAction struct {
	*types.Action
	PrivateKey *ecdsa.PrivateKey
}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)

	sysprikey, _ := crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	//syspubkey := common.HexToPubKey("0x047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd")
	minerprikey = append(minerprikey, sysprikey)
	minername = append(minername, params.DefaultChainconfig.SysName)
	minerprikey = append(minerprikey, ftproducer1key, ftproducer2key, ftproducer3key)
	minername = append(minername, ftproducer1name, ftproducer2name, ftproducer3name)
	ipc = append(ipc, "/home/Fractal/piTest/node_01/data/ft.ipc", "/home/Fractal/piTest/node_02/data/ft.ipc", "/home/Fractal/piTest/node_03/data/ft.ipc")

	basefrom = append(basefrom, "newnamefrom1%s", "newnamefrom2%s", "newnamefrom3%s")
	baseto = append(baseto, "newnameto1%s", "newnameto2%s", "newnameto3%s")
}

func GeneragePubKey() (common.PubKey, *ecdsa.PrivateKey) {
	prikey, _ := crypto.GenerateKey()
	return common.BytesToPubKey(crypto.FromECDSAPub(&prikey.PublicKey)), prikey
}

func createAccount(from, newname common.Name, nonce uint64, prikey *ecdsa.PrivateKey, pubkey common.PubKey) (common.Hash, error) {
	to := newname
	gc := newGeAction(types.CreateAccount, from, to, nonce, 1, gaslimit, nil, pubkey[:], prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return sendTxTest(gcs)
}

func issueAsset(from, Owner common.Name, amount *big.Int, assetname string, nonce uint64, prikey *ecdsa.PrivateKey) (common.Hash, error) {
	ast := &asset.AssetObject{
		AssetName: assetname,
		Symbol:    fmt.Sprintf("symbol%d", nonce),
		Amount:    amount,
		Decimals:  2,
		Owner:     Owner,
		Founder:    from,
		UpperLimit: big.NewInt(500000000000000000),
	}
	payload, err := rlp.EncodeToBytes(ast)
	if err != nil {
		panic("rlp payload err")
	}
	gc := newGeAction(types.IssueAsset, from, "", nonce, 1, gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return sendTxTest(gcs)
}

func transfer(from, to common.Name, amount *big.Int, nonce uint64, prikey *ecdsa.PrivateKey) (common.Hash, error) {
	gc := newGeAction(types.Transfer, from, to, nonce, 1, gaslimit, amount, nil, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return sendTxTest(gcs)
}

func newGeAction(at types.ActionType, from, to common.Name, nonce uint64, assetid uint64, gaslimit uint64, amount *big.Int, payload []byte, prikey *ecdsa.PrivateKey) *GenAction {
	action := types.NewAction(at, from, to, nonce, assetid, gaslimit, amount, payload)
	return &GenAction{
		Action:     action,
		PrivateKey: prikey,
	}
}

func sendTxTest(gcs []*GenAction) (common.Hash, error) {
	//nonce := GetNonce(sendaddr, "latest")
	signer := types.NewSigner(params.DefaultChainconfig.ChainID)
	var actions []*types.Action
	for _, v := range gcs {
		actions = append(actions, v.Action)
	}
	tx := types.NewTransaction(uint64(1), big.NewInt(2), actions...)
	for _, v := range gcs {
		err := types.SignAction(v.Action, tx, signer, v.PrivateKey)
		if err != nil {
			panic(fmt.Sprintf("SignAction err %v", err))
		}
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		jww.ERROR.Fatalln(err)
	}
	return tc.SendRawTx(rawtx)
}

func stressTest() {
	if len(os.Args) < 3 {
		panic("argument not enough")
	}
	threadsize, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic("argument err")
	}

	index, err = strconv.Atoi(os.Args[2])
	if err != nil {
		panic("argument err")
	}

	if threadsize >= len(indexstr) {
		panic("first argument over")
	}
	type UserInfo struct {
		name   common.Name
		pubkey common.PubKey
		prikey *ecdsa.PrivateKey
	}

	//tc.SetDefultURL(ipc[index])
	newusersfrom := make([]UserInfo, threadsize)
	newusersto := make([]UserInfo, threadsize)
	for i := 0; i < threadsize; i++ {
		newusersfrom[i].name = common.Name(fmt.Sprintf(basefrom[index], string(indexstr[i])))
		newusersfrom[i].pubkey, newusersfrom[i].prikey = GeneragePubKey()
		newusersto[i].name = common.Name(fmt.Sprintf(baseto[index], string(indexstr[i])))
		newusersto[i].pubkey, newusersto[i].prikey = GeneragePubKey()
	}
	nonce, _ := tc.GetNonce(minername[index])
	for i := 0; i < threadsize; i++ {
		h, err := createAccount(minername[index], newusersfrom[i].name, nonce, minerprikey[index], newusersfrom[i].pubkey)
		fmt.Println("create account from :", h.String(), "err:", err)
		nonce += 1
		h, err = createAccount(minername[index], newusersto[i].name, nonce, minerprikey[index], newusersto[i].pubkey)
		fmt.Println("create account to :", h.String(), "err:", err)
		nonce += 1
	}
	flag := true
	for {
		flag = true
		jww.INFO.Printf("check createAccount \n")
		time.Sleep(time.Duration(7) * time.Second)
		for i := 0; i < threadsize; i++ {
			act, _ := tc.GetAccountByName(newusersfrom[i].name)
			jww.INFO.Println("name", act.AcctName, "check createAccount 1", newusersfrom[i].name)
			if act.AcctName != newusersfrom[i].name {
				flag = false
			}
			act, _ = tc.GetAccountByName(newusersto[i].name)
			jww.INFO.Println("name", act.AcctName, "check createAccount 2", newusersto[i].name)
			if act.AcctName != newusersto[i].name {
				flag = false
			}
		}
		if flag {
			break
		}
	}
	jww.INFO.Printf("all account create success \n")
	nonce, _ = tc.GetNonce(minername[index])
	for i := 0; i < threadsize; i++ {
		h, err := transfer(minername[index], newusersfrom[i].name, big.NewInt(10000000000000), nonce, minerprikey[index])
		fmt.Println("transfer to prodcer :", h.String(), "err:", err)
		nonce += 1
	}
	for {
		flag = true
		jww.INFO.Printf("check transfer \n")
		time.Sleep(time.Duration(7) * time.Second)
		for i := 0; i < len(newusersfrom); i++ {
			balance, _ := tc.GetAccountBalanceByID(newusersfrom[i].name, 1)
			fmt.Println("name", newusersfrom[i].name, "balance:", balance.Int64())
			if balance.Int64() < 10000000000000 {
				flag = false
			}
		}
		if flag {
			break
		}
	}
	jww.INFO.Printf("all account transfer success \n")
	if len(os.Args) < 2 {
		panic("argument not enough")
	}
	type tmpasset struct {
		from    common.Name
		to      common.Name
		assetid uint64
	}
	ch := make(chan tmpasset, 100)
	txsnumber := make([]*uint64, threadsize)
	for i := 0; i < threadsize; i++ {
		txsnumber[i] = new(uint64)
	}
	var wg sync.WaitGroup
	for i := 0; i < threadsize; i++ {
		wg.Add(1)
		go func(from, to common.Name, prikey *ecdsa.PrivateKey, c chan tmpasset, txnumber *uint64) {
			defer wg.Add(-1)
			assetname := "na" + from.String()
			innonce, _ := tc.GetNonce(from)
			h, err := issueAsset(from, from, big.NewInt(100000000000000), assetname, innonce, prikey)
			fmt.Println("issueAsset :", h.String(), "err:", err)
			innonce += 1
			var astInfo *asset.AssetObject
			for {
				jww.INFO.Printf("check issueAsset %v\n", from)
				time.Sleep(time.Duration(7) * time.Second)
				astInfo, err = tc.GetAssetInfoByName(assetname)
				jww.INFO.Println(astInfo, err)
				fmt.Printf("astInfo.AssetName:[%s] assetname:[%s]\n", astInfo.AssetName, assetname)
				if astInfo != nil && astInfo.AssetName == assetname {
					//jww.INFO.Println("issueasset success ", assetname, astInfo.AssetId)
					jww.INFO.Println("assetid ", astInfo.AssetId)
					break
				}
			}
			jww.INFO.Println("start sent transfer forevery \n")
			flag := true
			for {
				gc := newGeAction(types.Transfer, from, to, innonce, astInfo.AssetId, gaslimit, big.NewInt(1), nil, prikey)
				var gcs []*GenAction
				gcs = append(gcs, gc)
				sendTxTest(gcs)
				innonce += 1
				if flag {
					for {
						jww.INFO.Printf("check balance %v\n", to)
						time.Sleep(time.Duration(7) * time.Second)
						balance, _ := tc.GetAccountBalanceByID(to, astInfo.AssetId)
						if balance.Int64() >= 1 {
							jww.INFO.Printf("Transfer to %v success\n", to)
							flag = false
							c <- tmpasset{from, to, astInfo.AssetId}
							break
						}
					}
				}
				*txnumber += 1
			}
		}(newusersfrom[i].name, newusersto[i].name, newusersfrom[i].prikey, ch, txsnumber[i])
	}
	total := 0
	var toassets []tmpasset
	for {
		tmp := <-ch
		toassets = append(toassets, tmp)
		total += 1
		if total >= threadsize {
			break
		}
	}
	type calculatetps struct {
		tm      int64
		balance int64
	}
	caltps := make(map[common.Name]*calculatetps, len(toassets))
	tm := time.Now().Unix()
	for i := 0; i < len(toassets); i++ {
		caltps[toassets[i].to] = &calculatetps{tm, 0}
	}

	go func() {
		defer wg.Add(-1)
		nonce, _ := tc.GetNonce(minername[index])
		for {
			time.Sleep(time.Duration(20) * time.Second)
			for i := 0; i < threadsize; i++ {
				h, err := transfer(minername[index], toassets[i].from, big.NewInt(100000), nonce, minerprikey[index])
				jww.INFO.Println("transfer to custom :", h.String(), "err:", err)
				nonce += 1
			}
		}
	}()
	lastnumber := uint64(0)
	lasttime := time.Now().Unix()
	for {
		time.Sleep(time.Duration(10) * time.Second)
		tmpnumber := uint64(0)
		actualnumber := uint64(0)
		tmptime := time.Now().Unix()
		for i := 0; i < threadsize; i++ {
			tmpnumber += *txsnumber[i]
			balance, _ := tc.GetAccountBalanceByID(toassets[i].to, toassets[i].assetid)
			actualnumber += balance.Uint64()
		}
		jww.INFO.Printf(" send tx number: %d\t succes num: %d\t tps: %d/s",
			tmpnumber, actualnumber, (actualnumber-lastnumber)/uint64(tmptime-lasttime))
		lasttime = tmptime
		lastnumber = actualnumber
	}

	wg.Wait()
}

func main() {
	stressTest()
}
