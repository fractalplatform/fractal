package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/params"

	"github.com/fractalplatform/fractal/sdk"
)

type blockinfo struct {
	Height    uint64   `json:"number"`
	Size      int64    `json:"size"`
	Miner     string   `json:"miner"`
	Txs       []string `json:"transactions"`
	TimeStamp int64    `json:"timestamp"`
}

func getBlockInfo(api *sdk.API, height uint64) (info *blockinfo, err error) {
	result := map[string]interface{}{}
	if height == math.MaxUint64 {
		result, err = api.GetCurrentBlock(false)
	} else {
		result, err = api.GetBlockByNumber(int64(height), false)
	}
	if err != nil {
		return
	}
	cj, _ := json.Marshal(result)
	info = &blockinfo{}
	err = json.Unmarshal(cj, info)
	return info, nil
}

//格式化数值    1,234,567,898.55
func NumberFormat(str string) string {
	length := len(str)
	if length < 4 {
		return str
	}
	arr := strings.Split(str, ".") //用小数点符号分割字符串,为数组接收
	length1 := len(arr[0])
	if length1 < 4 {
		return str
	}
	count := (length1 - 1) / 3
	for i := 0; i < count; i++ {
		arr[0] = arr[0][:length1-(i+1)*3] + "," + arr[0][length1-(i+1)*3:]
	}
	return strings.Join(arr, ".") //将一系列字符串连接为一个字符串，之间用sep来分隔。
}
func NumberFormatInt(str int64) string {
	i := strconv.FormatInt(str, 10)
	return NumberFormat(i)
}

func main() {
	_rpchost := flag.String("u", "http://192.168.2.11:3090", "RPC host地址")
	_start := flag.Uint64("s", 1, "统计的起始高度")
	_end := flag.Uint64("e", 0, "统计的结束高度")
	_detail := flag.Bool("d", true, "是否显示区块细节情况")
	flag.Parse()

	// init
	api := sdk.NewAPI(*_rpchost)
	chainCfg, err := api.GetChainConfig()
	if err != nil {
		panic(fmt.Sprintf("init err %v", err))
	}
	curblk, err := getBlockInfo(api, math.MaxUint64)
	if err != nil {
		panic(err)
	}
	if *_end == 0 || *_end > curblk.Height {
		*_end = curblk.Height
	}
	blockInterval := chainCfg.DposCfg.BlockInterval * uint64(time.Millisecond)
	mepochInterval := blockInterval * chainCfg.DposCfg.BlockFrequency * chainCfg.DposCfg.CandidateScheduleSize
	epochInterval := chainCfg.DposCfg.EpochInterval * uint64(time.Millisecond)
	epochFunc := func(timestamp uint64) uint64 {
		return (timestamp-chainCfg.ReferenceTime)/epochInterval + 1
	}
	epochTimeStampFunc := func(epoch uint64) uint64 {
		return (epoch-1)*epochInterval + chainCfg.ReferenceTime
	}
	offsetFunc := func(timestamp uint64, fid uint64) uint64 {
		interval := blockInterval
		if fid >= params.ForkID2 {
			interval = 0
		}
		offset := uint64(timestamp-interval) % epochInterval % mepochInterval
		offset /= blockInterval * chainCfg.DposCfg.BlockFrequency
		return offset
	}

	// done
	var timeStart, timeEnd int64
	var totalTxsCount, totalBlockSize int64
	var prevTime int64
	lastminer := ""
	for i := *_start; i <= *_end; i++ {
		blk, err := getBlockInfo(api, i)
		if err != nil {
			panic(err)
		}
		if i == *_start {
			timeStart = blk.TimeStamp
			prevTime = blk.TimeStamp
		} else if i == *_end {
			timeEnd = blk.TimeStamp
		}
		totalTxsCount += int64(len(blk.Txs))
		totalBlockSize += blk.Size

		printDetails := func(timestamp int64, miner string, txs uint64, height uint64) {
			if prevTime == blk.TimeStamp || epochFunc(uint64(prevTime)) != epochFunc(uint64(timestamp)) {
				epoch := epochFunc(uint64(timestamp))
				vcandidates, err := api.DposValidCandidates(epoch)
				if err != nil {
					panic(err)
				}
				fmt.Println("\n==========================周期==========================")
				fmt.Printf("epoch %d height %d\n", epoch, blk.Height)
				fmt.Printf("activatedCandidateSchedule %v\n", vcandidates["activatedCandidateSchedule"])
				fmt.Printf("badCandidateIndexSchedule %v\n", vcandidates["badCandidateIndexSchedule"])
				fmt.Printf("usingCandidateIndexSchedule %v\n", vcandidates["usingCandidateIndexSchedule"])
				fmt.Println("==========================周期==========================")
			}
			offset := offsetFunc(uint64(timestamp), params.ForkID2)
			if prevTime == blk.TimeStamp || offset != offsetFunc(uint64(prevTime), params.ForkID2) {
				mepoch := (uint64(blk.TimeStamp) - epochTimeStampFunc(epochFunc(uint64(blk.TimeStamp)))) / mepochInterval / 10
				fmt.Printf("\n%03d-%03d-%05d(%s):", mepoch, offset, height, miner)
				lastminer = miner
			}
			if lastminer != miner {
				fmt.Printf("%5d-%05d(%s)", txs, height, miner)
			} else {
				fmt.Printf("%5d", txs)
			}
		}
		if *_detail {
			for blk.TimeStamp-prevTime >= 2*int64(blockInterval) {
				timestamp := prevTime + int64(blockInterval)
				printDetails(timestamp, "******", 0, 0)
				prevTime = timestamp
			}
			printDetails(blk.TimeStamp, blk.Miner, uint64(len(blk.Txs)), blk.Height)
			prevTime = blk.TimeStamp
		}
	}

	if *_detail {
		fmt.Printf("\n")
	}
	if timeEnd < timeStart {
		panic("启动时间大于结束时间")
	}
	tStart := time.Unix(timeStart/int64(time.Second), timeStart%int64(time.Second))
	tEnd := time.Unix(timeEnd/int64(time.Second), timeEnd%int64(time.Second))
	timeElapse := (timeEnd - timeStart) / int64(time.Second)
	if timeElapse == 0 {
		timeElapse = 1
	}
	fmt.Println("===================Result=======================")
	fmt.Println("Block Num:", *_start, "=>", *_end, "count", *_end-*_start)
	fmt.Println("Time From:", tStart.Format("2006-01-02 15:04:05.99"))
	fmt.Println("Time To  :", tEnd.Format("2006-01-02 15:04:05.99"))
	fmt.Println("Elapse   :", timeElapse/(60*60), "H:", timeElapse%(60*60)/60, "M:", timeElapse%(60*60)%60, "S")
	fmt.Println("Total Txs:", NumberFormatInt(totalTxsCount), "\t\tSpeed:", float64(totalTxsCount)/float64(timeElapse), "Tx/s")
	fmt.Println("Total Len:", NumberFormatInt(totalBlockSize), "B", "\tSpeed:", float64(totalBlockSize)/float64(timeElapse), "B/s")
}
