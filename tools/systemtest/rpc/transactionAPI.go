package rpc

import (
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

func GetReceiptByTxHash(hash common.Hash) (*types.RPCReceipt, error) {
	receipt := &types.RPCReceipt{}
	err := ClientCall("ft_getTransactionReceipt", receipt, hash.Hex())
	if err != nil {
		fmt.Println("ft_getTransactionReceipt:" + hash.Hex() + ", err=" + err.Error())
	}
	return receipt, err
}

func DelayGetReceiptByTxHash(txHash common.Hash, maxTime uint) (*types.RPCReceipt, bool, error) {
	for maxTime > 0 {
		time.Sleep(time.Duration(1) * time.Second)
		receipt, err := GetReceiptByTxHash(txHash)
		if err != nil {
			return nil, false, err
		}

		//json, _ := json.Marshal(receipt)
		//fmt.Println("DelayGetReceiptByTxHash：" + string(json))
		if receipt.BlockNumber > 0 {
			return receipt, false, nil
		}
		maxTime--
	}
	return &types.RPCReceipt{}, maxTime == 0, nil
}

func GetAllTxInPendingAndQueue() (map[string]map[string]map[string]string, error) {
	allNotExecutedTxs := map[string]map[string]map[string]string{}
	err := ClientCall("txpool_inspect", &allNotExecutedTxs)

	return allNotExecutedTxs, err
}

func GetTxpoolStatus() (int, int) {
	result := map[string]int{}
	ClientCall("txpool_status", &result)
	return result["pending"], result["queue"]
}

func IsTxpoolFull() bool {
	pendingTxNum, _ := GetTxpoolStatus()
	return pendingTxNum >= MaxTxNumInTxpool
}

func CheckTxInPendingOrQueue(accountName string, txHash common.Hash) (bool, bool, error) {
	allNotExecutedTxs, err := GetAllTxInPendingAndQueue()
	if err != nil {
		return false, false, errors.New("无法获取pending和queue中的txs")
	}
	pendingTxs, _ := allNotExecutedTxs["pending"]
	queuedTxs, _ := allNotExecutedTxs["queued"]

	bInPending := false
	bInQueued := false
	accountTxs, ok := pendingTxs[accountName]
	if ok {
		if _, ok = accountTxs[txHash.Hex()]; ok {
			bInPending = true
		}
	}

	accountTxs, ok = queuedTxs[accountName]
	if ok {
		if _, ok = accountTxs[txHash.Hex()]; ok {
			bInQueued = true
		}
	}
	return bInPending, bInQueued, nil
}
