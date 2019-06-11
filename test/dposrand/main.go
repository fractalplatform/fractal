package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"

	"github.com/fractalplatform/fractal/accountmanager"

	"github.com/fractalplatform/fractal/consensus/dpos"

	"github.com/fractalplatform/fractal/sdk"
)

func main() {
	_rpchost := flag.String("u", "http://127.0.0.1:8545", "RPC host地址")
	_detail := flag.Bool("d", true, "是否显示投票细节情况")
	flag.Parse()

	// init
	api := sdk.NewAPI(*_rpchost)
	chainCfg, err := api.GetChainConfig()
	if err != nil {
		panic(err)
	}
	epochInterval := chainCfg.DposCfg.EpochInterval * uint64(time.Millisecond)
	epochFunc := func(timestamp uint64) uint64 {
		return (timestamp-chainCfg.ReferenceTime)/epochInterval + 1
	}
	unitStakeFunc := func() *big.Int {
		return new(big.Int).Mul(chainCfg.DposCfg.UnitStake, big.NewInt(1e18))
	}

	accts := []*accountmanager.Account{}
	for id := uint64(4097); ; id++ {
		acct, err := api.AccountInfoByID(id)
		if err != nil || acct.AccountID < 4097 {
			break
		}
		if acct.AcctName.String() == chainCfg.DposName ||
			acct.AcctName.String() == chainCfg.AccountName ||
			acct.AcctName.String() == chainCfg.AssetName ||
			acct.AcctName.String() == chainCfg.FeeName {
			continue
		}
		accts = append(accts, acct)
	}
	acctscnt := len(accts)

	priv, _ := crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	rand.Seed(time.Now().UnixNano())
	prevEpoch := epochFunc(getFirstBlockTimestamp(api))
	fmt.Println("start epoch", prevEpoch, "interval", chainCfg.DposCfg.EpochInterval)
	for {
		epoch := epochFunc(getCurrentBlockTimestamp(api))
		if epoch == prevEpoch {
			time.Sleep(time.Second)
			continue
		}
		candidates := getCandidates(api, epoch)
		candidatescnt := len(candidates)
		var rw sync.RWMutex
		voteNum := 0
		voteQ := big.NewInt(0)
		fmt.Printf("==========================%d==============================\n", epoch)
		nthread := runtime.NumCPU()
		if acctscnt < nthread {
			nthread = acctscnt
		}
		wg := sync.WaitGroup{}
		wg.Add(nthread)
		for i := 0; i < nthread; i++ {
			go func(index int) {
				wg.Done()
				for {
					n := (acctscnt/nthread)*index + rand.Intn(acctscnt/nthread)
					name := accts[n].AcctName
					acct := sdk.NewAccount(api, name, priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)
					cnt := rand.Intn(5) + 1
					for stake := availableStake(api, epoch, name.String()); stake.Cmp(new(big.Int).Mul(unitStakeFunc(), chainCfg.DposCfg.VoterMinQuantity)) == 1; {
						candidateInfo := candidates[rand.Intn(candidatescnt)]
						if candidateInfo.Type != dpos.Normal {
							continue
						}
						q := big.NewInt(int64(rand.Intn(100)) + chainCfg.DposCfg.VoterMinQuantity.Int64())
						stake := new(big.Int).Mul(unitStakeFunc(), q)
						hash, err := acct.VoteCandidate(common.StrToName(chainCfg.DposName), big.NewInt(0), chainCfg.SysTokenID, 500000, &dpos.VoteCandidate{
							Candidate: candidateInfo.Name,
							Stake:     stake,
						})
						if *_detail {
							fmt.Printf("%04d: %v ==> %v %v(%v), hash %v, err %v\n", index, name, candidateInfo.Name, stake, q, hash.String(), err)
						}
						if err != nil || getTxEpoch(api, hash) != epoch {
							break
						}
						rw.Lock()
						voteNum++
						voteQ = new(big.Int).Add(voteQ, q)
						candidateInfo.TotalQuantity = new(big.Int).Add(candidateInfo.TotalQuantity, q)
						rw.Unlock()
						cnt--
						if cnt == 0 {
							break
						}
					}
					if tepoch := epochFunc(getCurrentBlockTimestamp(api)); epoch != tepoch {
						break
					}
				}
			}(i)
		}
		wg.Wait()
		bts, _ := json.Marshal(candidates)
		fmt.Printf("%v(%v:%v) result:\n %s\n", voteNum, voteQ, epoch, string(bts))
		fmt.Printf("==========================END==============================\n")
		prevEpoch = epoch
	}
}

func getFirstBlockTimestamp(api *sdk.API) uint64 {
	ret, err := api.GetBlockByNumber(1, false)
	if err != nil {
		panic(err)
	}
	return uint64(ret["timestamp"].(float64))
}

func getCurrentBlockTimestamp(api *sdk.API) uint64 {
	ret, err := api.GetCurrentBlock(false)
	if err != nil {
		panic(err)
	}
	return uint64(ret["timestamp"].(float64))
}

func getTxEpoch(api *sdk.API, hash common.Hash) uint64 {
	ret, err := api.GetTransactionByHash(hash)
	if err != nil {
		panic(err)
	}
	epoch, err := api.DposEpoch(ret.BlockNumber)
	if err != nil {
		panic(err)
	}
	return epoch
}
func availableStake(api *sdk.API, epoch uint64, name string) *big.Int {
	stake, err := api.DposAvailableStake(epoch, name)
	if err != nil {
		panic(err)
	}
	return stake
}

func getCandidates(api *sdk.API, epoch uint64) []*dpos.CandidateInfo {
	ret, err := api.DposCandidates(epoch, true)
	if err != nil {
		panic(err)
	}
	candidates := []*dpos.CandidateInfo{}
	bts, err := json.Marshal(ret)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bts, &candidates)
	if err != nil {
		panic(err)
	}
	return candidates
}
