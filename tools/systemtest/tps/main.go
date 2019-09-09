package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"math/big"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	tcommon "github.com/fractalplatform/systemtest/common"
)

var (
	gasLimit = uint64(900000)
)

func pause() {
	fmt.Println("请按任意键继续...")
	var c rune
	fmt.Scanf("%c", &c)
}

func main() {
	rpcHost := flag.String("rpc", "http://localhost:8545", "rpc host地址")
	sysacctn := flag.String("sysacct", "ftsystemio:289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032", "发行账户私钥")
	systoken := flag.String("systoken", "ftperfoundation", "系统代币名字")
	delay := flag.Int64("delay", 100, "发送交易延时(ms)")
	autodelay := flag.Int64("autodelay", 1, "调整发送交易延时(s)")
	intThreadCnt := flag.Int("t", 0, "测试用线程数,默认为CPU个数")
	chainID := flag.Int64("chainid", 1, "chain id")
	flag.Parse()
	if *intThreadCnt == 0 {
		*intThreadCnt = runtime.NumCPU()
	}

	api := tcommon.NewAPI(*rpcHost)
	// issuer
	splits := strings.Split(*sysacctn, ":")
	if len(splits) != 2 {
		println("系统账户选项参数出错啦~~~(账户名:私钥)不合法", *sysacctn)
		return
	}
	if !common.IsValidName(splits[0]) {
		println("系统账户选项参数出错啦~~~账户名不合法", splits[0])
		return
	}
	priv, err := crypto.HexToECDSA(splits[1])
	if err != nil {
		println("系统账户选项参数出错啦~~~私钥不合法~~~", splits[1])
		return
	}
	acct, err := api.AccountInfo(splits[0])
	if err != nil {
		panic(fmt.Sprintf("get account info %v err %v", splits[0], err))
	} else if bytes.Compare(acct.PublicKey.Bytes(), common.BytesToPubKey(crypto.FromECDSAPub(&priv.PublicKey)).Bytes()) != 0 {
		println("系统账户选项参数出错啦~~~私钥不匹配~~~", splits[1])
		return
	}
	sysname := common.StrToName(splits[0])
	syspriv := priv

	asset, err := api.AssetInfoByName(*systoken)
	if err != nil {
		panic(fmt.Sprintf("get asset info %v err %v", *systoken, err))
	}
	systokenid := asset.AssetId
	decimals := big.NewInt(1)
	for i := uint64(0); i < asset.Decimals; i++ {
		decimals = new(big.Int).Mul(decimals, big.NewInt(10))
	}
	sysacct := tcommon.NewAccount(api, sysname, syspriv, systokenid, math.MaxUint64, true, big.NewInt(*chainID))

	initAccounts := map[common.Name]*tcommon.Account{}
	for i := 0; i < *intThreadCnt; i++ {
		aname := common.StrToName(tcommon.GenerateAccountName("tps", 10))
		apriv, apub := tcommon.GenerateKey()
		acct := tcommon.NewAccount(api, aname, apriv, systokenid, math.MaxUint64, true, big.NewInt(*chainID))
		initAccounts[aname] = acct
		hash, err := sysacct.CreateAccount(aname, new(big.Int).Mul(big.NewInt(500000), decimals), systokenid, gasLimit, apub)
		if err != nil {
			panic(err)
		}
		fmt.Println("create account", aname, "hash:", hash.String())
	}
	fmt.Println("create account...", *intThreadCnt)
	pause()

	// for to := range initAccounts {
	// 	hash, err := sysacct.Transfer(to, new(big.Int).Mul(big.NewInt(100000), decimals), systokenid, gasLimit)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("transfer account", to, "hash:", hash.String())
	// }
	// pause()

	// for to, acct := range initAccounts {
	// 	hash, err := acct.RegProducer(to, new(big.Int).Mul(big.NewInt(0), decimals), systokenid, gasLimit, fmt.Sprintf("www.%v.com", to), new(big.Int).Mul(big.NewInt(10000), decimals))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("regproducer account", to, "hash:", hash.String())
	// }
	// pause()

	atomic.StoreInt64(delay, (*delay)*int64(time.Millisecond))
	if *autodelay > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration((*autodelay) * int64(time.Second)))
			for {
				select {
				case <-ticker.C:
					ndelay := atomic.AddInt64(delay, int64(-1)*int64(time.Millisecond))
					if ndelay == 0 {
						return
					}
				}
			}
		}()
	}

	wg := &sync.WaitGroup{}
	wg.Add(*intThreadCnt)
	for name, acct := range initAccounts {
		go func(name common.Name, acct *tcommon.Account) {
			defer wg.Done()
			// create rand account
			tpriv, tpub := tcommon.GenerateKey()
			tname := tcommon.GenerateAccountName("tpstmp", 10)
			value := int64(10000)
			if _, err := acct.CreateAccount(common.StrToName(tname), new(big.Int).Mul(big.NewInt(value), decimals), systokenid, gasLimit, tpub); err != nil {
				fmt.Println("create account err", err)
				return
			}
			tacct := tcommon.NewAccount(api, common.StrToName(tname), tpriv, systokenid, 0, false, big.NewInt(*chainID))
			for {
				hash, err := tacct.Transfer(name, big.NewInt(1), systokenid, gasLimit)
				if err != nil {
					fmt.Println("tranfer account err", err)
					return
				} else {
					fmt.Println("tranfer", hash.String(), "delay", time.Duration(*delay))
				}
				time.Sleep(time.Duration(*delay))
			}
		}(name, acct)
	}
	wg.Wait()
}
