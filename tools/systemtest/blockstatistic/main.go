package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	tcommon "github.com/fractalplatform/systemtest/common"
)

type blockinfo struct {
	Heignt    uint64   `json:"number"`
	Size      int64    `json:"size"`
	Miner     string   `json:"miner"`
	Txs       []string `json:"transactions"`
	TimeStamp int64    `json:"timestamp"`
}

func parseBlockInfo(api *tcommon.API, height uint64) (info *blockinfo, err error) {
	result := map[string]interface{}{}
	if height == math.MaxUint64 {
		result, err = api.CurrentBlock(false)
	} else {
		result, err = api.BlockByNumber(int64(height), false)
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
	rpcHost := flag.String("u", "http://localhost:8545", "RPC host地址")
	_start := flag.Uint64("s", 1, "统计的起始高度")
	_end := flag.Uint64("e", 0, "统计的结束高度")
	bDetails := flag.Bool("d", false, "是否显示区块细节情况")
	intervalTime := flag.Int64("i", 3000, "出块时间间隔(单位毫秒)")
	blockRepeat := flag.Int("r", 6, "单个节点一次出块个数")
	validators := flag.Int("v", 3, "出块人列表个数")
	flag.Parse()

	startH := *_start
	endH := *_end
	itime := (*intervalTime) * int64(time.Millisecond)
	vtime := itime * int64(*blockRepeat)
	etime := vtime * int64(*validators)
	api := tcommon.NewAPI(*rpcHost)
	info, err := parseBlockInfo(api, math.MaxUint64)
	if err != nil {
		println("Can`t get block end height.")
		return
	}
	if startH < 0 {
		startH = 0
	}
	if endH == 0 || endH > info.Heignt {
		endH = info.Heignt
	}

	var timeStart, timeEnd, privts int64
	var totalTxsCount, totalBlockSize int64

	if *bDetails {
		fmt.Println("===================Detail=======================")
	}

	for i := startH; i <= endH; i++ {
		info, err := parseBlockInfo(api, i)
		if err != nil {
			panic(err)
		}
		timestamp := info.TimeStamp
		cntTxs := int64(len(info.Txs))

		if i == startH {
			timeStart = timestamp
			privts = timestamp
		} else if i == endH {
			timeEnd = timestamp
		}
		totalTxsCount += cntTxs
		totalBlockSize += info.Size

		printDetails := func(miner string, ch int64) {
			if timestamp == privts || timestamp%etime/vtime != (timestamp-itime)%etime/vtime {
				fmt.Printf("\n%s:", miner)
			}
			fmt.Printf("%5d(%05d)", ch, info.Heignt)
		}
		if *bDetails {
			ttimestamp := timestamp
			for ttimestamp-privts >= 2*itime {
				timestamp = privts + itime
				if timestamp%etime%vtime == 0 && ttimestamp-timestamp >= vtime {
					printDetails("==========", -1)
				} else {
					printDetails(info.Miner, -1)
				}
				privts = timestamp
			}
			timestamp = ttimestamp
			printDetails(info.Miner, int64(cntTxs))
			privts = timestamp
		}
	}

	if *bDetails {
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
	fmt.Println("Block Num:", startH, "=>", endH, "count", endH-startH)
	fmt.Println("Time From:", tStart.Format("2006-01-02 15:04:05.99"))
	fmt.Println("Time To  :", tEnd.Format("2006-01-02 15:04:05.99"))
	fmt.Println("Elapse   :", timeElapse/(60*60), "H:", timeElapse%(60*60)/60, "M:", timeElapse%(60*60)%60, "S")
	fmt.Println("Total Txs:", NumberFormatInt(totalTxsCount), "\t\tSpeed:", float64(totalTxsCount)/float64(timeElapse), "Tx/s")
	fmt.Println("Total Len:", NumberFormatInt(totalBlockSize), "B", "\tSpeed:", float64(totalBlockSize)/float64(timeElapse), "B/s")
}
