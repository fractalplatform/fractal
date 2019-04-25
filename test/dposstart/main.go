package main

import (
	"flag"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	args "github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/sdk"
)

var (
	gasLimit = uint64(210000)
	gasPrice = big.NewInt(10000000000)
)

func main() {
	rpcHost := flag.String("u", "http://localhost:8545", "RPC host地址")
	issueHex := flag.String("s", "ftsystemio:289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032", "系统账户私钥")
	systoken := flag.String("systoken", "ftoken", "系统代币名字")
	value := flag.Int64("n", 10000000, "生成者代币质押数量(最小单位),1000万")
	chainID := flag.Int64("chainid", 1, "chain id")
	flag.Parse()

	candidates := flag.Args()
	if len(candidates) == 0 {
		candidates = append(candidates, "ftcandidate1:289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
		candidates = append(candidates, "ftcandidate2:9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658")
		candidates = append(candidates, "ftcandidate3:8605cf6e76c9fc8ac079d0f841bd5e99bd3ad40fdd56af067993ed14fc5bfca8")
	}

	fmt.Println("RPC:", rpcHost)
	fmt.Println("系统账户:", *issueHex)
	fmt.Println("系统代币:", *systoken)
	fmt.Println("质押数量:", *value)
	fmt.Println("生成者数量:", len(candidates))
	fmt.Println("生成者列表:", candidates)
	api := sdk.NewAPI(*rpcHost)

	sysasset, err := api.AssetInfoByName(*systoken)
	if err != nil {
		panic(fmt.Sprintf("get asset info %v err %v", *systoken, err))
	}
	systokenid := sysasset.AssetId
	systokendecimals := big.NewInt(1)
	for i := uint64(0); i < sysasset.Decimals; i++ {
		systokendecimals = new(big.Int).Mul(systokendecimals, big.NewInt(10))
	}

	// issuer
	splits := strings.Split(*issueHex, ":")
	if len(splits) != 2 {
		println("系统账户出错啦~~~", *issueHex)
		return
	}
	if !common.IsValidAccountName(splits[0]) {
		println("系统账户非法啦~~~", splits[0])
		return
	}
	issuerName := common.StrToName(splits[0])
	issuerPriv, err := crypto.HexToECDSA(splits[1])
	if err != nil {
		println("系统账户私钥出错啦~~~", issueHex)
		return
	}

	if len(candidates) < 3 {
		println("生产者个数不能小于3~~~", len(candidates))
		return
	}
	issuerAcct := sdk.NewAccount(api, issuerName, issuerPriv, systokenid, math.MaxUint64, true, big.NewInt(*chainID))

	prods := map[common.Name]*sdk.Account{}
	for _, privHex := range candidates {
		splits := strings.Split(privHex, ":")
		if len(splits) != 2 {
			println("生产者账户出错啦~~~", privHex)
			return
		}
		if !common.IsValidAccountName(splits[0]) {
			println("生产者账户非法啦~~~", splits[0])
			return
		}
		name := common.StrToName(splits[0])
		priv, err := crypto.HexToECDSA(splits[1])
		if err != nil {
			println("生产者账户私钥出错啦~~~", privHex)
			return
		}
		prods[name] = sdk.NewAccount(api, name, priv, systokenid, math.MaxUint64, true, big.NewInt(*chainID))
	}

	delegateValue := new(big.Int).Mul(big.NewInt(*value), systokendecimals)
	issueValue := new(big.Int).Mul(big.NewInt(*value+10), systokendecimals)

	fmt.Println("转账金额/人:", issueValue)
	fmt.Println("质押金额/人:", delegateValue)

	for to, acct := range prods {
		existed, _ := api.AccountIsExist(to.String())
		if !existed {
			hash, err := issuerAcct.CreateAccount(issuerName, issueValue, systokenid, gasLimit, &accountmanager.AccountAction{
				AccountName: to,
				PublicKey:   acct.Pubkey(),
			})
			fmt.Println(hash.String(), ":", issuerName, "create & transfer", issueValue, fmt.Sprintf("(%v)", systokenid), "to", to, "error", err)
		} else {
			hash, err := issuerAcct.Transfer(to, issueValue, systokenid, gasLimit)
			fmt.Println(hash.String(), ":", issuerName, "transfer", issueValue, fmt.Sprintf("(%v)", systokenid), "to", to, "error", err)
		}
	}

	// reg candidates
	for name, acct := range prods {
		value := big.NewInt(1e5)
		hash, err := acct.RegCandidate(name, big.NewInt(0), systokenid, gasLimit, &args.RegisterCandidate{
			Url:   "www." + name.String() + ".io",
			Stake: delegateValue,
		})
		fmt.Println(hash.String(), ":", issuerName, "reg", delegateValue, "& transfer", value, "(", systokenid, ")", "to ", name, "error", err)
	}

	fmt.Println("Dpos启动完成.")
}
