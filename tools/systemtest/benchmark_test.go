package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
	. "github.com/fractalplatform/systemtest/rpc"
	. "github.com/smartystreets/goconvey/convey"
	jww "github.com/spf13/jwalterweatherman"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	wg            sync.WaitGroup
	success       int32
	fail          int32
	dposLoop      int64 = 3
	blocksPerNode       = 6
	curLoopTxNum  uint  = 0
	txAddPerStep        = uint(1000)
	normalTPS     uint  = 500
)

func transfer(from, to string, assetId uint64, amount *big.Int, prikey *ecdsa.PrivateKey) (common.Hash, error) {
	nonce, err := GetNonce(common.Name(from))
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.Transfer, common.Name(from), common.Name(to), nonce, assetId, Gaslimit, amount, nil, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return SendTxTest(gcs)
}
func createNewAccountAutoGenName() {
	defer wg.Done()
	defer func() {
		if (fail+success)%100 == 0 {
			//jww.INFO.Printf("Executed full TX number, fail=%d, success=%d", fail, success)
		}
	}()
	accountName := GenerateAccountName("bm", 5)
	err := createNewAccount(SystemAccount, SystemAccountPriKey, accountName, nil)
	if err != nil {
		fmt.Println(err.Error())
		atomic.AddInt32(&fail, 1)
		return
	}
	atomic.AddInt32(&success, 1)
}

func createAccountAndTransfer() {
	defer wg.Done()
	defer func() {
		if (fail+success)%100 == 0 {
			//jww.INFO.Printf("Executed full TX number, fail=%d, success=%d", fail, success)
		}
	}()
	accountName := GenerateAccountName("bm", 8)
	err := createNewAccount(SystemAccount, SystemAccountPriKey, accountName, nil)
	if err != nil {
		jww.INFO.Println(err.Error())
		atomic.AddInt32(&fail, 2)
		return
	}
	_, err = transfer(SystemAccount, accountName, 1, big.NewInt(1), SystemAccountPriKey)
	//err = TransferAsset(SystemAccount, accountName, 1, big.NewInt(1), SystemAccountPriKey)
	if err != nil {
		jww.INFO.Println(err.Error())
		atomic.AddInt32(&fail, 1)
		return
	}
	atomic.AddInt32(&success, 2)
}

func batchRunTx(txFunc func(), txNumInFunc uint) {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)
	//curBlockHeight, err := GetCurrentBlockHeight()
	//jww.INFO.Printf("Current block height:" + strconv.Itoa(int(curBlockHeight)))
	//So(err, ShouldBeNil)
	totalBlockTxNum := uint(0)

	for curLoopTxNum < 100 {
		time.Sleep(time.Duration(3 * int64(time.Second)))
		jww.INFO.Printf("--------------Enter next loop----------------")
		success = 0
		fail = 0

		//startTime := time.Now()
		result := make(map[int64]int)
		totalBlockTxNum = 0
		SentTxNum = 0

		curBlockHeight, _ := GetCurrentBlockHeight()
		curBlockHeight++
		jww.INFO.Printf("Current block height:" + strconv.Itoa(int(curBlockHeight)))
		lastSentNum := int32(0)
		go func() {
			for {
				time.Sleep(time.Duration(200 * int64(time.Millisecond)))

				pendingNum, queueNum := GetTxpoolStatus()
				if lastSentNum != SentTxNum {
					jww.INFO.Printf("TX Stat: sentTxNum=%d, pendingNum=%d, queueNum=%d", SentTxNum, pendingNum, queueNum)
				}
				lastSentNum = SentTxNum

				txNum, err := GetTxNumByBlockHeight(curBlockHeight)
				if err != nil || txNum < 0 {
					continue
				}
				result[curBlockHeight] = txNum
				jww.INFO.Printf("TX Stat: sentTxNum=%d, success=%d, fail=%d, height=%d, txNum=%d",
					SentTxNum, success, fail, curBlockHeight, txNum)
				curBlockHeight++
				totalBlockTxNum += uint(txNum)
				if totalBlockTxNum == txAddPerStep*txNumInFunc {
					//timeVal := time.Since(startTime).Seconds()
					//jww.INFO.Printf("tps = %f", float64(totalBlockTxNum) / timeVal)
					break
				}
				if uint(success+fail) == txAddPerStep*txNumInFunc {
					jww.INFO.Printf("result: %v", result)
					break
				}
			}
		}()

		wg.Add(int(txAddPerStep))
		for i := 0; i < int(txAddPerStep); i++ {
			go txFunc()
		}

		wg.Wait()
		curLoopTxNum++
	}
}

func TestBatchCreateNormalAccount(t *testing.T) {
	Convey("用系统账户批量创建新账户，观察交易打包和出块情况(System account create a lot of new account, and check the block info)", t, func() {
		batchRunTx(createNewAccountAutoGenName, 1)
	})
}

func TestBatchCreateAccountAndTransfer(t *testing.T) {
	Convey("用系统账户批量创建新账户，并给其转账(System account creates a lot of new account, then transfer a mount of asset to them)", t, func() {
		batchRunTx(createAccountAndTransfer, 2)
	})
}
