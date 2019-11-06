package main

import (
	"fmt"
	"math/big"

	testcommon "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

func sendTx() error {

	from, to := "", ""
	value := big.NewInt(1)
	gasLimit := uint64(20000000)

	action := types.NewAction(types.Transfer, from, to, 0, 0, gasLimit, value, nil, nil)
	gasprice, err := testcommon.GasPrice()
	if err != nil {
		return err
	}

	tx := types.NewTransaction(0, gasprice, action)

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
