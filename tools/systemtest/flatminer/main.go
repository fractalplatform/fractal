package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
	tcommon "github.com/fractalplatform/systemtest/common"
	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelDebug)
}

type rpcTransaction struct {
	*types.RPCTransaction
	receipts []*types.RPCActionResult
}

type rpcBlock struct {
	types.Header
	Hash common.Hash       `json:"hash"`
	Size uint64            `json:"size"`
	Txs  []*rpcTransaction `json:"-"`
}

func getCurrentBlockInfo(api *tcommon.API) *rpcBlock {
	ret, _ := api.CurrentBlock(false)
	cj, _ := json.Marshal(ret)
	blk := &rpcBlock{}
	json.Unmarshal(cj, blk)
	return blk
}

func getBlockInfo(api *tcommon.API, height int64) (*rpcBlock, error) {
	ret, err := api.BlockByNumber(height, false)
	if err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, nil
	}

	ret["extraData"] = ""
	cj, _ := json.Marshal(ret)
	blk := &rpcBlock{}
	if err := json.Unmarshal(cj, blk); err != nil {
		return nil, err
	}
	infs := ret["transactions"].([]interface{})
	for _, inf := range infs {
		hash := common.HexToHash(inf.(string))
		rpctx, err := api.TransactionByHash(hash)
		if err != nil {
			return nil, err
		}
		tx := &rpcTransaction{
			RPCTransaction: rpctx,
		}

		receipt, err := api.TransactionReceiptByHash(tx.Hash)
		if err != nil {
			return nil, err
		}

		tx.receipts = receipt.ActionResults

		if len(tx.RPCActions) != 1 {
			panic("unsupport actions more than 1 in tx")
		}
		if len(tx.receipts) != 1 {
			panic("unsupport receipts more than 1 in tx")
		}
		blk.Txs = append(blk.Txs, tx)
	}
	return blk, nil
}

func main() {
	rpcHost := flag.String("rpc", "http://192.168.2.13:8445", "rpc host地址")
	sysacctn := flag.String("sysacct", "ftsystemio", "系统账户")
	dposacctn := flag.String("dposacct", "ftsystemdpos", "dpos系统账户")
	systoken := flag.String("systoken", "ftfoundation", "系统代币名字")
	systokenaccount := flag.String("systokenaccount", "100000000000000000000000000000", "系统代币初始值")
	blkReward := flag.Int64("reward", 5, "区块奖励")
	blkExtraReward := flag.Int64("extrareward", 1, "额外奖励")
	blkInterval := flag.Int64("blockinterval", 3000, "出块时间间隔(单位毫秒)")
	blkRepeat := flag.Int64("blockrepeat", 6, "出块个数")
	blkValidators := flag.Int64("validators", 3, "出块人个数")
	flag.Parse()

	blockReward := new(big.Int).Mul(big.NewInt(*blkReward), big.NewInt(1e18))
	blockExtraReward := new(big.Int).Mul(big.NewInt(*blkExtraReward), big.NewInt(1e18))
	vtime := (*blkInterval) * (*blkRepeat) * int64(time.Millisecond)
	etime := vtime * (*blkValidators)

	api := tcommon.NewAPI(*rpcHost)
	sysasset, err := api.AssetInfoByName(*systoken)
	if err != nil {
		panic(fmt.Sprintf("get asset info %v err %v", *systoken, err))
	}
	systokenid := sysasset.AssetId
	systokendecimals := big.NewInt(1)
	for i := uint64(0); i < sysasset.Decimals; i++ {
		systokendecimals = new(big.Int).Mul(systokendecimals, big.NewInt(10))
	}
	info, err := api.DposInfo()
	if err != nil {
		panic(fmt.Sprintf("get dpos info err %v", err))
	}
	unitstake := new(big.Int).Mul(big.NewInt(int64(info["UnitStake"].(float64))), systokendecimals)

	accts := map[string]map[uint64]*big.Int{}
	accts[*sysacctn] = map[uint64]*big.Int{}
	amount := new(big.Int)
	amount.SetString(*systokenaccount, 10)
	accts[*sysacctn][systokenid] = amount

	accts[*dposacctn] = map[uint64]*big.Int{}
	accts[*dposacctn][systokenid] = big.NewInt(0)

	changeaccts := map[string]string{}
	changeacts := map[uint64]uint64{}

	prouducers := map[string]*big.Int{} // producer --> total quantity
	prouducers[*sysacctn] = big.NewInt(0)

	extraCnt := int64(0)
	height := int64(1)
	for {
		blk, err := getBlockInfo(api, height)
		if err != nil {
			panic(err)
		}
		if blk == nil {
			for name := range changeaccts {
				for id, balance := range accts[name] {
					tbalance, _ := api.BalanceByAssetID(name, id)
					if balance.Cmp(tbalance) != 0 {
						if getCurrentBlockInfo(api).Number.Int64() >= height {
							jww.WARN.Println(height-1, name, "wrong", id, "balance", "have", tbalance, "except", balance, "actions", changeacts)
						} else {
							jww.ERROR.Println(height-1, name, "wrong", id, "balance", "have", tbalance, "except", balance, "actions", changeacts)
						}
					} else {
						jww.INFO.Println(height-1, name, "right", id, "balance", balance)
					}
				}

				for prouducer, quantity := range prouducers {
					ret, _ := api.DposAccount(prouducer)
					if len(ret) == 0 && quantity.Cmp(big.NewInt(0)) == 0 {
						continue
					}
					tquantity := big.NewInt(int64(ret["Quantity"].(float64)))
					if quantity.Cmp(tquantity) != 0 {
						if getCurrentBlockInfo(api).Number.Int64() >= height {
							jww.WARN.Println(height-1, prouducer, "wrong quantity", "have", tquantity, "except", quantity, "actions", changeacts)
						} else {
							jww.ERROR.Println(height-1, prouducer, "wrong quantity", "have", tquantity, "except", quantity, "actions", changeacts)
						}
					} else {
						jww.INFO.Println(height-1, prouducer, "right quantity", quantity)
					}
				}
			}
			changeaccts = map[string]string{}
			changeacts = map[uint64]uint64{}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		miner := blk.Coinbase.String()
		if _, ok := accts[miner]; !ok {
			accts[miner] = map[uint64]*big.Int{}
		}

		for _, tx := range blk.Txs {
			gasid := tx.GasAssetID
			gasprice := tx.GasPrice

			gasused := int64(0)
			receipt := tx.receipts[0]
			gasused += int64(receipt.GasUsed)

			fee := new(big.Int).Mul(gasprice, big.NewInt(gasused))

			//该笔交易手续费收入
			if _, ok := accts[miner][gasid]; !ok {
				accts[miner][gasid] = big.NewInt(0)
			}
			accts[miner][gasid] = new(big.Int).Add(accts[miner][gasid], fee)
			jww.DEBUG.Println(height, miner, "miner+fee", fee)

			action := tx.RPCActions[0]
			changeacts[action.Type] = action.Type
			if tx.receipts[0].Status == 1 {
				//该笔交易支出
				from := action.From.String()
				accts[from][gasid] = new(big.Int).Sub(accts[from][gasid], fee)
				jww.DEBUG.Println(height, from, "tranfer-fee", fee)
				accts[from][action.AssetID] = new(big.Int).Sub(accts[from][action.AssetID], action.Amount)
				jww.DEBUG.Println(height, from, "tranfer-account", action.Amount)
				changeaccts[from] = from

				//该笔交易收入
				to := action.To.String()
				if _, ok := accts[to]; !ok {
					accts[to] = map[uint64]*big.Int{}
				}
				if _, ok := accts[to][action.AssetID]; !ok {
					accts[to][action.AssetID] = big.NewInt(0)
				}
				accts[to][action.AssetID] = new(big.Int).Add(accts[to][action.AssetID], action.Amount)
				jww.DEBUG.Println(height, to, "tranfer+account", action.Amount)
				changeaccts[action.To.String()] = action.To.String()

				//附加
				switch types.ActionType(action.Type) {
				case types.DeleteAccount:
				case types.IssueAsset:
					arg := &asset.AssetObject{}
					rlp.DecodeBytes(action.Payload, arg)
					asset, _ := api.AssetInfoByName(arg.AssetName)
					owner := arg.Owner.String()
					if _, ok := accts[owner]; !ok {
						accts[owner] = map[uint64]*big.Int{}
					}
					if _, ok := accts[owner][asset.AssetId]; !ok {
						accts[owner][asset.AssetId] = big.NewInt(0)
					}
					accts[owner][asset.AssetId] = new(big.Int).Add(accts[owner][asset.AssetId], arg.Amount)
					jww.DEBUG.Println(height, owner, "issue+account", arg.Amount)
					changeaccts[owner] = owner
				case types.IncreaseAsset:
					arg := &asset.AssetObject{}
					rlp.DecodeBytes(action.Payload, arg)
					asset, _ := api.AssetInfoByName(arg.AssetName)
					owner := asset.Owner.String()
					if _, ok := accts[owner]; !ok {
						accts[owner] = map[uint64]*big.Int{}
					}
					if _, ok := accts[owner][asset.AssetId]; !ok {
						accts[owner][asset.AssetId] = big.NewInt(0)
					}
					accts[owner][arg.AssetId] = new(big.Int).Add(accts[owner][arg.AssetId], arg.Amount)
					jww.DEBUG.Println(height, owner, "increase+account", arg.Amount)
					changeaccts[owner] = owner
				case types.SetAssetOwner:
					// nothing
				case types.RegProducer:
					arg := &dpos.RegisterProducer{}
					rlp.DecodeBytes(action.Payload, arg)
					accts[from][systokenid] = new(big.Int).Sub(accts[from][systokenid], arg.Stake)
					jww.DEBUG.Println(height, from, "reg-delegate", arg.Stake)
					accts[*dposacctn][systokenid] = new(big.Int).Add(accts[*dposacctn][systokenid], arg.Stake)
					jww.DEBUG.Println(height, *dposacctn, "reg+delegate", arg.Stake)
					changeaccts[*dposacctn] = *dposacctn
					if _, ok := prouducers[from]; !ok {
						prouducers[from] = big.NewInt(0)
					}
					prouducers[from] = new(big.Int).Add(prouducers[from], new(big.Int).Div(arg.Stake, unitstake))
				case types.UpdateProducer:
					arg := &dpos.UpdateProducer{}
					rlp.DecodeBytes(action.Payload, arg)
					ret, _ := api.DposAccount(from)
					stake := new(big.Int).Mul(big.NewInt(int64(ret["Quantity"].(float64))), unitstake)
					actualstake := new(big.Int).Sub(arg.Stake, stake)
					accts[from][systokenid] = new(big.Int).Sub(accts[from][systokenid], actualstake)
					jww.DEBUG.Println(height, from, "update-delegate", actualstake)
					accts[*dposacctn][systokenid] = new(big.Int).Add(accts[*dposacctn][systokenid], actualstake)
					jww.DEBUG.Println(height, *dposacctn, "update+delegate", actualstake)
					changeaccts[*dposacctn] = action.To.String()
					prouducers[from] = new(big.Int).Add(prouducers[from], new(big.Int).Div(actualstake, unitstake))
				case types.UnregProducer:
					ret, _ := api.DposAccount(from)
					if len(ret) == 0 {
						break
					}
					stake := new(big.Int).Mul(big.NewInt(int64(ret["Quantity"].(float64))), unitstake)
					accts[from][systokenid] = new(big.Int).Add(accts[from][systokenid], stake)
					jww.DEBUG.Println(height, from, "reg+undelegate", stake)
					accts[*dposacctn][systokenid] = new(big.Int).Sub(accts[*dposacctn][systokenid], stake)
					jww.DEBUG.Println(height, *dposacctn, "reg-undelegate", stake)
					changeaccts[*dposacctn] = *dposacctn
					prouducers[from] = new(big.Int).Sub(prouducers[from], new(big.Int).Div(stake, unitstake))
				case types.VoteProducer:
					arg := &dpos.VoteProducer{}
					rlp.DecodeBytes(action.Payload, arg)
					accts[from][systokenid] = new(big.Int).Sub(accts[from][systokenid], arg.Stake)
					jww.DEBUG.Println(height, from, "vote-delegate", arg.Stake)
					accts[*dposacctn][systokenid] = new(big.Int).Add(accts[*dposacctn][systokenid], arg.Stake)
					jww.DEBUG.Println(height, *dposacctn, "vote+delegate", arg.Stake)
					changeaccts[*dposacctn] = *dposacctn
					prouducers[arg.Producer] = new(big.Int).Add(prouducers[arg.Producer], new(big.Int).Div(arg.Stake, unitstake))
				case types.ChangeProducer:
					// nothing
					ret, _ := api.DposAccount(from)
					quantity := big.NewInt(int64(ret["Quantity"].(float64)))
					producer := ret["Producer"].(string)
					prouducers[producer] = new(big.Int).Sub(prouducers[producer], quantity)
					arg := &dpos.ChangeProducer{}
					rlp.DecodeBytes(action.Payload, arg)
					prouducers[arg.Producer] = new(big.Int).Add(prouducers[arg.Producer], quantity)
				case types.UnvoteProducer:
					ret, _ := api.DposAccount(from)
					stake := new(big.Int).Mul(big.NewInt(int64(ret["Quantity"].(float64))), unitstake)
					accts[from][systokenid] = new(big.Int).Add(accts[from][systokenid], stake)
					jww.DEBUG.Println(height, from, "vote+undelegate", stake)
					accts[*dposacctn][systokenid] = new(big.Int).Sub(accts[*dposacctn][systokenid], stake)
					jww.DEBUG.Println(height, *dposacctn, "vote-undelegate", stake)
					changeaccts[*dposacctn] = *dposacctn
					producer := ret["Producer"].(string)
					prouducers[producer] = new(big.Int).Sub(prouducers[producer], new(big.Int).Div(stake, unitstake))
				case types.RemoveVoter:
					arg := &dpos.RemoveVoter{}
					rlp.DecodeBytes(action.Payload, arg)
					ret, _ := api.DposAccount(arg.Voter)
					stake := new(big.Int).Mul(big.NewInt(int64(ret["Quantity"].(float64))), unitstake)
					accts[arg.Voter][systokenid] = new(big.Int).Add(accts[arg.Voter][systokenid], stake)
					jww.DEBUG.Println(height, arg.Voter, "vote+undelegate", stake)
					changeaccts[arg.Voter] = arg.Voter
					accts[*dposacctn][systokenid] = new(big.Int).Sub(accts[*dposacctn][systokenid], stake)
					jww.DEBUG.Println(height, *dposacctn, "vote-undelegate", stake)
					changeaccts[*dposacctn] = *dposacctn
					producer := ret["Producer"].(string)
					prouducers[producer] = new(big.Int).Sub(prouducers[producer], new(big.Int).Div(stake, unitstake))
				}
			}
		}

		if _, ok := accts[miner][systokenid]; !ok {
			accts[miner][systokenid] = big.NewInt(0)
		}
		//区块奖励
		accts[miner][systokenid] = new(big.Int).Add(accts[miner][systokenid], blockReward)
		jww.DEBUG.Println(height, miner, "miner+reward", blockReward)

		//额外奖励
		if blk.Time.Int64()%etime%vtime == 0 {
			jww.DEBUG.Println(height, miner, "miner+extrareward", extraCnt, new(big.Int).Mul(big.NewInt(extraCnt), blockExtraReward))
			accts[miner][systokenid] = new(big.Int).Add(accts[miner][systokenid], new(big.Int).Mul(big.NewInt(extraCnt), blockExtraReward))
			extraCnt = 0
		}
		extraCnt++
		changeaccts[miner] = miner
		height++
	}
}
