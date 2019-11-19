package main

import (
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	pm "github.com/fractalplatform/fractal/plugin"
	testcommon "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

func sendTx() error {

	privateKey, _ := crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")

	act := &pm.CreateAccountAction{
		Name:   "testaccount",
		Desc:   "system account",
		Pubkey: common.HexToPubKey("047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"),
	}

	payload, err := rlp.EncodeToBytes(act)
	if err != nil {
		return err
	}

	from, to := "fractal.founder", "fractal.account"
	value := big.NewInt(1)
	gasLimit := uint64(20000000)

	action := types.NewAction(pm.CreateAccount, from, to, 0, 0, gasLimit, value, payload, nil)
	gasprice, err := testcommon.GasPrice()
	if err != nil {
		return err
	}

	tx := types.NewTransaction(0, gasprice, action)

	signer, _ := pm.NewSigner(big.NewInt(1))

	d, err := signer.Sign(tx.SignHash, privateKey)
	if err != nil {
		return err
	}

	action.WithSignature(d)
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}

	hash, err := testcommon.SendRawTx(rawtx)
	if err != nil {
		return err
	}

	fmt.Println("result hash: ", hash.Hex())
	return nil
}

func main() {
	if err := sendTx(); err != nil {
		fmt.Println("send transaction error", err.Error())
	}
}
