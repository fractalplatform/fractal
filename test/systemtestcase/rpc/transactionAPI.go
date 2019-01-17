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

package rpc

import (
	"fmt"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

func GetReceiptByTxHash(hash common.Hash) (*types.RPCReceipt, error) {
	receipt := &types.RPCReceipt{}
	err := ClientCall("ft_getTransactionReceipt", receipt, hash.Hex())

	return receipt, err
}

func DelayGetReceiptByTxHash(txHash common.Hash, maxTime uint) (*types.RPCReceipt, bool, error) {
	for maxTime > 0 {
		time.Sleep(time.Duration(1) * time.Second)
		receipt, err := GetReceiptByTxHash(txHash)
		if err != nil {
			fmt.Println("DelayGetReceiptByTxHash:" + err.Error())
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

func GetTxpoolStatus() (int, int) {
	result := map[string]int{}
	ClientCall("txpool_status", result)
	return result["pending"], result["queue"]
}

func IsTxpoolFull() bool {
	pendingTxNum, _ := GetTxpoolStatus()
	return pendingTxNum >= MaxTxNumInTxpool
}
