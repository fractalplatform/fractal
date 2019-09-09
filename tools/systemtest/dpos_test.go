package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/fractalplatform/fractal/crypto"
	. "github.com/fractalplatform/systemtest/rpc"
	jww "github.com/spf13/jwalterweatherman"
)

var nodePortMap = map[string][]int64{"192.168.2.13": []int64{13100, 13101, 13102, 13103, 13104},
	"192.168.2.14": []int64{14100, 14101, 14102, 14103, 14104},
	"192.168.2.15": []int64{15100, 15101, 15102, 15103, 15104},
}

func startDPos(nodeNum int) bool {
	// 查看当前网络出块节点是否已满足要求，如已满足，则无需再次初始化网络
	// check whether current dpos network satisfies the condition of packing block, if satisfy, return
	activeProducerStat, err := GetActiveProducerStat()
	if err != nil {
		jww.INFO.Println(err.Error())
		return false
	}
	if len(activeProducerStat) >= nodeNum {
		return true
	}

	// 检查p2p连接是否都已经完成
	// check whether the p2p connection has been finished
	for ip, ports := range nodePortMap {
		for _, port := range ports {
			peerCount, err := GetPeerCount(ip, port)
			if err != nil {
				jww.INFO.Println(err.Error())
				return false
			}
			jww.INFO.Printf("Peer（%s : %d） node count  = %d", ip, port, peerCount)
		}
	}

	var accounts []string = make([]string, nodeNum)
	var priKeys []*ecdsa.PrivateKey = make([]*ecdsa.PrivateKey, nodeNum)
	var pkStrs []string = make([]string, nodeNum)

	// 创建十五个新账号，并给它们转账
	// create 15 new accounts, and then transfer a mount of asset to them
	for i := 0; i < nodeNum; i++ {
		accounts[i], priKeys[i] = registerAccount("dpos", 5)
		TransferAsset(SystemAccount, accounts[i], 1, getAssetAmountByStake(big.NewInt(100)), SystemAccountPriKey)
		pkStrs[i] = hex.EncodeToString(crypto.FromECDSA(priKeys[i]))
	}

	// 修改节点的coinbase信息
	// modify node's coinbase info
	minerNum := 0
	for ip, ports := range nodePortMap {
		for _, port := range ports {
			SetSpecifiedMinerCoinbase(ip, port, accounts[minerNum], pkStrs[minerNum])
			minerNum++
		}
	}

	// 账号注册为生产者
	// register producer
	for i := 0; i < nodeNum; i++ {
		RegProducer(accounts[i], priKeys[i], "www.xxx.com", getAssetAmountByStake(big.NewInt(100)))
	}

	maxWaitTime := 10
	for maxWaitTime > 0 {
		jww.INFO.Printf("check new active producer")
		activeProducerStat, _ := GetActiveProducerStat()
		if len(activeProducerStat) >= nodeNum {
			jww.INFO.Printf("%d active producers has been started.", len(activeProducerStat))
			return true
		}
		time.Sleep(time.Duration(3 * int64(time.Second)))
		maxWaitTime--
	}
	return false
}

func TestStartDPos(t *testing.T) {
	Convey("启动十五个生产者(start up 15 producers)", t, func() {
		So(startDPos(15), ShouldBeTrue)
	})
}
