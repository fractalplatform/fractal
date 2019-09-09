package rpc

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
	"math/big"
	"strconv"
)

func regProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey, url string, stake *big.Int) (common.Hash, error) {
	rp := dpos.RegisterProducer{
		Url:   url,
		Stake: stake,
	}

	rawdata, err := rlp.EncodeToBytes(rp)
	if err != nil {
		return common.Hash{}, err
	}

	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.RegProducer, accountName, accountName, nonce, 1, Gaslimit, big.NewInt(1e5), rawdata, fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, err
}

func updateProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey, url string, stake *big.Int) (common.Hash, error) {
	rp := dpos.RegisterProducer{
		Url:   url,
		Stake: stake,
	}

	rawdata, err := rlp.EncodeToBytes(rp)
	if err != nil {
		return common.Hash{}, err
	}

	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.UpdateProducer, accountName, "", nonce, 1, Gaslimit, nil, rawdata, fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, err
}

func unRegProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey) (common.Hash, error) {
	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.UnregProducer, accountName, "", nonce, 1, Gaslimit, nil, nil, fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, err
}

func voteProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey, producer string, stake *big.Int) (common.Hash, error) {
	arg := &dpos.VoteProducer{
		Producer: producer,
		Stake:    stake,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.VoteProducer, accountName, "", nonce, 1, Gaslimit, nil, payload, fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, err
}

func changeProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey, producer string) (common.Hash, error) {
	arg := &dpos.ChangeProducer{
		Producer: producer,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.ChangeProducer, accountName, "", nonce, 1, Gaslimit, nil, payload, fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, err
}

func unvoteProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey) (common.Hash, error) {
	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.UnvoteProducer, accountName, "", nonce, 1, Gaslimit, nil, nil, fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, err
}

func removeVoter(fromAccount string, fromPriKey *ecdsa.PrivateKey, voter string) (common.Hash, error) {
	arg := &dpos.RemoveVoter{
		Voter: voter,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	accountName := common.Name(fromAccount)
	nonce, err := GetNonce(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	gc := NewGeAction(types.RemoveVoter, accountName, "", nonce, 1, Gaslimit, nil, payload, fromPriKey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	txHash, err := SendTxTest(gcs)
	return txHash, err
}

func getDPosAccount(name common.Name) (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	err := ClientCall("dpos_account", &fields, name.String())
	return fields, err
}

type producerInfo struct {
	Name          string   // producer name
	URL           string   // producer url
	Quantity      *big.Int // producer stake quantity
	TotalQuantity *big.Int // producer total stake quantity
	Height        uint64   // timestamp
}

func ValidateAllProducers() error {
	producers, err := getAllProducers()
	if err != nil {
		return err
	}
	for _, producer := range producers {
		result, err := json.Marshal(producer)
		fmt.Println(string(result))
		producerInfo, err := GetProducerInfo(producer.Name)
		if err != nil {
			return errors.New("无法正确获取所有的生产者：" + err.Error())
		}
		if !isProducer(producerInfo) {
			return errors.New("无法正确获取所有的生产者")
		}
	}
	return nil
}

func setMinerCoinbase(name string, privKey string) error {
	err := ClientCall("miner_setCoinbase", name, privKey)
	return err
}

func SetSpecifiedMinerCoinbase(nodeIp string, nodePort int64, name string, privKey string) error {
	err := ClientCallWithAddr(nodeIp, nodePort, "miner_setCoinbase", name, privKey)
	return err
}

func getAllProducers() ([]producerInfo, error) {
	fields := []producerInfo{}
	err := ClientCall("dpos_producers", &fields)
	return fields, err
}

func getDPosConfigInfo() (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	err := ClientCall("dpos_info", &fields)
	return fields, err
}

func isGlobalConfigInfo(globalInfo map[string]interface{}) bool {
	_, ok1 := globalInfo["ActivatedMinQuantity"]
	_, ok2 := globalInfo["BlockFrequency"]
	_, ok3 := globalInfo["ProducerScheduleSize"]
	_, ok4 := globalInfo["VoterMinQuantity"]
	_, ok5 := globalInfo["UnitStake"]

	return ok1 && ok2 && ok3 && ok4 && ok5
}

func getDPosGlobalInfo() (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	err := ClientCall("dpos_validateEpcho", &fields)
	return fields, err
}

func isGlobalInfo(globalInfo map[string]interface{}) bool {
	_, scheduleOK := globalInfo["ActivatedProducerSchedule"]
	_, heightOK := globalInfo["Height"]
	_, updateOK := globalInfo["ActivatedProducerScheduleUpdate"]
	_, activateQuantityOK := globalInfo["ActivatedTotalQuantity"]
	_, totalQuantityOK := globalInfo["TotalQuantity"]

	return scheduleOK && heightOK && updateOK && activateQuantityOK && totalQuantityOK
}

func GetProducerInfo(producer string) (map[string]interface{}, error) {
	producerInfo, err := getDPosAccount(common.Name(producer))
	if err != nil {
		return nil, errors.New("无法获取生产者信息：" + err.Error())
	}
	if !isProducer(producerInfo) {
		return nil, errors.New("无法获取生产者信息")
	}
	return producerInfo, nil
}

func GetVoterInfo(voter string) (map[string]interface{}, error) {
	voterInfo, err := getDPosAccount(common.Name(voter))
	if err != nil {
		return nil, errors.New("无法获取投票者信息：" + err.Error())
	}
	if !isVoter(voterInfo) {
		return nil, errors.New("无法获取投票者信息：")
	}
	return voterInfo, nil
}

func compareProducerAfterVote(oldProducerInfo, newProducerInfo map[string]interface{}, stake *big.Int) bool {
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, dpos.DefaultConfig.UnitStake, m)
	oldQuantity, _ := oldProducerInfo["Quantity"]
	oldTotalQuantity, _ := oldProducerInfo["TotalQuantity"]
	newQuantity, _ := newProducerInfo["Quantity"]
	newTotalQuantity, _ := newProducerInfo["TotalQuantity"]
	if newQuantity.(*big.Int).Sub(newQuantity.(*big.Int), oldQuantity.(*big.Int)).Cmp(q) != 0 {
		return false
	}
	if newTotalQuantity.(*big.Int).Sub(newTotalQuantity.(*big.Int), oldTotalQuantity.(*big.Int)).Cmp(q) != 0 {
		return false
	}
	return true
}

func compareVoterAfterVote(oldVoterInfo, newVoterInfo map[string]interface{}, stake *big.Int) bool {
	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, dpos.DefaultConfig.UnitStake, m)
	oldQuantity, _ := oldVoterInfo["Quantity"]
	newQuantity, _ := newVoterInfo["Quantity"]
	if oldQuantity.(*big.Int).Sub(oldQuantity.(*big.Int), newQuantity.(*big.Int)).Cmp(q) != 0 {
		return false
	}
	return true
}

func isProducer(producerInfo map[string]interface{}) bool {
	_, urlOK := producerInfo["URL"]
	_, heightOK := producerInfo["Height"]
	_, nameOK := producerInfo["Name"]
	_, quantityOK := producerInfo["Quantity"]
	_, totalQuantityOK := producerInfo["TotalQuantity"]

	return urlOK && heightOK && nameOK && quantityOK && totalQuantityOK
}

func isVoter(voterInfo map[string]interface{}) bool {
	_, heightOK := voterInfo["Height"]
	_, nameOK := voterInfo["Name"]
	_, quantityOK := voterInfo["Quantity"]
	_, producerOK := voterInfo["Producer"]

	return heightOK && nameOK && quantityOK && producerOK
}

func ValidateActiveProducer() error {
	dposGlobalInfo, err := getDPosGlobalInfo()
	if err != nil {
		return errors.New("无法获取全局投票信息：" + err.Error())
	}
	if !isGlobalInfo(dposGlobalInfo) {
		return errors.New("无法获取全局投票信息")
	}
	newActivatedProducerSchedule, _ := dposGlobalInfo["ActivatedProducerSchedule"].([]string)
	for _, activeProducer := range newActivatedProducerSchedule {
		accountExist, err := AccountIsExist(activeProducer)

		if !accountExist || err != nil {
			return errors.New("有Active生产者的账号已不存在：" + activeProducer)
		}
	}
	return nil
}

func GetActiveProducerStat() (map[string]uint, error) {
	dposGlobalInfo, err := getDPosGlobalInfo()
	if err != nil {
		return nil, errors.New("无法获取全局投票信息：" + err.Error())
	}
	if !isGlobalInfo(dposGlobalInfo) {
		return nil, errors.New("无法获取全局投票信息")
	}
	activeProducerStat := map[string]uint{}
	newActivatedProducerSchedule, _ := dposGlobalInfo["ActivatedProducerSchedule"].([]string)
	for _, activeProducer := range newActivatedProducerSchedule {
		accountExist, err := AccountIsExist(activeProducer)

		if !accountExist || err != nil {
			return nil, errors.New("有Active生产者的账号已不存在：" + activeProducer)
		}
		activeProducerStat[activeProducer]++
	}
	return activeProducerStat, nil
}

func isInActiveProducerList(producer string) (bool, error) {
	dposGlobalInfo, err := getDPosGlobalInfo()
	if err != nil || !isGlobalInfo(dposGlobalInfo) {
		return false, errors.New("无法获取全局投票信息：" + err.Error())
	}
	newActivatedProducerSchedule, _ := dposGlobalInfo["ActivatedProducerSchedule"].([]string)

	return isActiveProducer(newActivatedProducerSchedule, producer), nil
}

func isActiveProducer(activeProducers []string, producer string) bool {
	for _, activeProducer := range activeProducers {
		if activeProducer == producer {
			return true
		}
	}
	return false
}
func IsProducerExist(producerName string) bool {
	producerInfo, err := getDPosAccount(common.Name(producerName))
	if err != nil {
		return false
	}
	return isProducer(producerInfo)
}

func IsVoterExist(voterName string) bool {
	voterInfo, err := getDPosAccount(common.Name(voterName))
	if err != nil {
		return false
	}
	return isVoter(voterInfo)
}

// 判断某生产者在当前时刻是否有可能进入active生产者列表，注意这里得到的并不是确定性结果，因为只要随后投票发生改变，都有可能影响生产者列表的产生
func couldBeInActiveProducerList(producer string) (bool, error) {
	producerInfo, err := GetProducerInfo(producer)
	if err != nil || !isProducer(producerInfo) {
		return false, errors.New("取消投票后无法获取新生产者信息：" + err.Error())
	}

	configInfo, err := getDPosConfigInfo()
	if err != nil || !isGlobalConfigInfo(configInfo) {
		return false, errors.New("投票者取消投票后，无法获取dpos配置信息：" + err.Error())
	}
	producerScheduleSize := configInfo["ProducerScheduleSize"].(uint64)

	producers, err := getAllProducers()
	if err != nil {
		return false, errors.New("投票者取消投票后，无法获取所有生产者信息：" + err.Error())
	}
	count := 0
	for _, producer := range producers {
		if producer.TotalQuantity.Cmp(producerInfo["TotalQuantity"].(*big.Int)) > 0 {
			count++
		}
	}
	return uint64(count) < producerScheduleSize, nil
}

func curRankingInProducers(producerInfo map[string]interface{}, producers []producerInfo) (int, error) {
	count := 0
	for _, producer := range producers {
		if producer.TotalQuantity.Cmp(producerInfo["TotalQuantity"].(*big.Int)) > 0 {
			count++
		}
	}
	return count + 1, nil
}

func curRanking(producerInfo map[string]interface{}) (int, error) {
	producers, err := getAllProducers()
	if err != nil {
		return -1, errors.New("无法获取所有生产者信息：" + err.Error())
	}
	return curRankingInProducers(producerInfo, producers)
}

func curRankingOf(producer string) (int, error) {
	producerInfo, err := GetProducerInfo(producer)
	if err != nil || !isProducer(producerInfo) {
		return -1, errors.New("无法获取新生产者信息：" + err.Error())
	}
	return curRanking(producerInfo)
}

func checkDPosGlobalInfo(oldDPosGlobalInfo map[string]interface{}, newDPosGlobalInfo map[string]interface{}, stake *big.Int, bAdded bool) error {
	if value, ok := oldDPosGlobalInfo["TotalQuantity"]; ok {
		fmt.Printf("oldDPosGlobalInfo-TotalQuantity %v\n", strconv.FormatFloat(value.(float64), 'f', -1, 64))
	}

	oldTotalVoteQuantity, oldOK := oldDPosGlobalInfo["TotalQuantity"]
	newTotalVoteQuantity, newOK := newDPosGlobalInfo["TotalQuantity"]

	if !oldOK || !newOK {
		return errors.New("无法获取新旧全局信息中的投票总数")
	}
	oldQuantity, parseOldErr := strconv.ParseInt(strconv.FormatFloat(oldTotalVoteQuantity.(float64), 'f', -1, 64), 10, 64)
	newQuantity, parseNewErr := strconv.ParseInt(strconv.FormatFloat(newTotalVoteQuantity.(float64), 'f', -1, 64), 10, 64)
	if parseOldErr != nil || parseNewErr != nil {
		return errors.New("投票总数类型转换失败")
	}

	m := big.NewInt(0)
	q, _ := new(big.Int).DivMod(stake, dpos.DefaultConfig.UnitStake, m)
	if !bAdded {
		q = q.Mul(q, big.NewInt(-1))
	}
	if big.NewInt(0).Sub(big.NewInt(oldQuantity), big.NewInt(newQuantity)).Cmp(q) != 0 {
		return errors.New("全局投票总数不对")
	}
	return nil
}

func runTxFuncThenCheckGlobalInfo(txFunc func() error, stake *big.Int, bAdded bool) error {
	oldDPosGlobalInfo, err := getDPosGlobalInfo()
	if err != nil || !isGlobalInfo(oldDPosGlobalInfo) {
		if err != nil {
			return errors.New("无法获取全局投票信息：" + err.Error())
		} else {
			return errors.New("无法获取全局投票信息")
		}
	}

	err = txFunc()
	if err != nil {
		return err
	}

	newDPosGlobalInfo, err := getDPosGlobalInfo()
	if err != nil || !isGlobalInfo(newDPosGlobalInfo) {
		if err != nil {
			return errors.New("无法获取全局投票信息：" + err.Error())
		} else {
			return errors.New("无法获取全局投票信息")
		}
	}
	if err := checkDPosGlobalInfo(oldDPosGlobalInfo, newDPosGlobalInfo, stake, bAdded); err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func RegProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey, url string, stake *big.Int) error {

	err := runTxFuncThenCheckGlobalInfo(func() error {
		err := runTxAndCheckReceipt(func() (common.Hash, error) {
			txHash, err := regProducer(fromAccount, fromPriKey, url, stake)
			if err != nil {
				return common.Hash{}, errors.New("注册生产者交易失败：" + err.Error())
			}
			return txHash, nil
		}, fromAccount)

		if err != nil {
			return err
		}

		if !IsProducerExist(fromAccount) {
			return errors.New("无法获取生产者信息")
		}
		return nil
	}, stake, true)
	if err != nil {
		return errors.New("注册生产者发生错误：" + err.Error())
	}

	return nil
}

func UnRegProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey) error {
	producerInfo, err := GetProducerInfo(fromAccount)
	if err != nil || !isProducer(producerInfo) {
		return errors.New("注销前无法获取生产者信息：" + err.Error())
	}
	stake, _ := producerInfo["TotalQuantity"]

	err = runTxFuncThenCheckGlobalInfo(func() error {
		err := runTxAndCheckReceipt(func() (common.Hash, error) {
			txHash, err := unRegProducer(fromAccount, fromPriKey)
			if err != nil {
				return common.Hash{}, errors.New("注销生产者交易失败：" + err.Error())
			}
			return txHash, nil
		}, fromAccount)

		if err != nil {
			return err
		}

		if IsProducerExist(fromAccount) {
			return errors.New("生产者未注销成功，依然能查询到")
		}
		return nil
	}, stake.(*big.Int), false)

	if err != nil {
		return errors.New("注销生产者发生错误：" + err.Error())
	}

	return nil
}

func UpdateProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey, url string, stake *big.Int) error {
	err := runTxFuncThenCheckGlobalInfo(func() error {
		oldProducerInfo, err := GetProducerInfo(fromAccount)
		if err != nil {
			return errors.New("更新前无法获取生产者信息：" + err.Error())
		}

		err = runTxAndCheckReceipt(func() (common.Hash, error) {
			txHash, err := updateProducer(fromAccount, fromPriKey, url, stake)
			if err != nil {
				return common.Hash{}, errors.New("更新生产者交易失败：" + err.Error())
			}
			return txHash, nil
		}, fromAccount)

		if err != nil {
			return err
		}

		if !IsProducerExist(fromAccount) {
			return errors.New("无法获取生产者信息")
		}

		newProducerInfo, err := GetProducerInfo(fromAccount)
		if err != nil {
			return errors.New("无法获取更新后的生产者信息：" + err.Error())
		}
		newUrl, _ := newProducerInfo["URL"]
		if newUrl != url {
			return errors.New("生产者的URL更新失败")
		}

		if !compareProducerAfterVote(oldProducerInfo, newProducerInfo, stake) {
			return errors.New("更新生产者投票信息后，投票结果不一致")
		}
		return nil
	}, stake, true)

	if err != nil {
		return errors.New("更新生产者发生错误：" + err.Error())
	}

	return nil
}

func VoteProducer(fromAccount string, fromPriKey *ecdsa.PrivateKey, producer string, stake *big.Int) error {
	err := runTxFuncThenCheckGlobalInfo(func() error {
		oldProducerInfo, err := GetProducerInfo(producer)
		if err != nil {
			return errors.New("投票前无法获取生产者信息：" + err.Error())
		}
		oldVoterInfo, err := GetVoterInfo(fromAccount)
		if err != nil {
			return errors.New("投票前无法获取投票者信息：" + err.Error())
		}

		err = runTxAndCheckReceipt(func() (common.Hash, error) {
			txHash, err := voteProducer(fromAccount, fromPriKey, producer, stake)
			if err != nil {
				return common.Hash{}, errors.New("投票交易失败：" + err.Error())
			}
			return txHash, nil
		}, fromAccount)

		if err != nil {
			return err
		}

		newProducerInfo, err := GetProducerInfo(producer)
		if err != nil {
			return errors.New("投票后无法获取生产者信息：" + err.Error())
		}
		newVoterInfo, err := GetVoterInfo(fromAccount)
		if err != nil {
			return errors.New("投票后无法获取投票者信息：" + err.Error())
		}

		if !compareProducerAfterVote(oldProducerInfo, newProducerInfo, stake) {
			return errors.New("投票者投票后，生产者的投票结果信息不对")
		}

		if !compareVoterAfterVote(oldVoterInfo, newVoterInfo, stake) {
			return errors.New("投票者投票后，投票者的投票结果信息不对")
		}
		return nil
	}, stake, true)

	if err != nil {
		return errors.New("更新生产者发生错误：" + err.Error())
	}

	return nil
}

// 1：在投票者改变生产者前，先获取全局投票信息、投票者和生产者信息
// 2：投票给其它生产者
// 3：再次获取全局投票信息、投票者和生产者信息，跟之前获取的信息进行比较
func ChangeProducer(voter string, voterPriKey *ecdsa.PrivateKey, producer string) error {
	err := runTxFuncThenCheckGlobalInfo(func() error {
		oldVoterInfo, err := GetVoterInfo(voter)
		if err != nil {
			return errors.New("改变生产者前无法获取投票者信息：" + err.Error())
		}
		stake := oldVoterInfo["Quantity"].(*big.Int)
		oldProducer := oldVoterInfo["Producer"].(string)
		oldFirstProducerInfo, err := GetProducerInfo(oldProducer)
		if err != nil {
			return errors.New("改变生产者前无法获取原生产者信息：" + err.Error())
		}
		oldSecondProducerInfo, err := GetProducerInfo(producer)
		if err != nil {
			return errors.New("改变生产者前无法获取新生产者信息：" + err.Error())
		}

		err = runTxAndCheckReceipt(func() (common.Hash, error) {
			txHash, err := changeProducer(voter, voterPriKey, producer)
			if err != nil {
				return common.Hash{}, errors.New("改变生产者交易失败：" + err.Error())
			}
			return txHash, nil
		}, voter)

		if err != nil {
			return err
		}

		newVoterInfo, err := GetVoterInfo(voter)
		if err != nil {
			return errors.New("改变生产者后无法获取投票者信息：" + err.Error())
		}
		newFirstProducerInfo, err := GetProducerInfo(oldProducer)
		if err != nil {
			return errors.New("改变生产者后无法获取新生产者信息：" + err.Error())
		}
		newSecondProducerInfo, err := GetProducerInfo(producer)
		if err != nil {
			return errors.New("改变生产者后无法获取新生产者信息：" + err.Error())
		}

		if !compareProducerAfterVote(oldFirstProducerInfo, newFirstProducerInfo, stake.Mul(stake, big.NewInt(-1))) {
			return errors.New("投票者改变生产者后，生产者的投票结果信息不对")
		}

		if !compareProducerAfterVote(oldSecondProducerInfo, newSecondProducerInfo, stake) {
			return errors.New("投票者改变生产者后，生产者的投票结果信息不对")
		}

		if !compareVoterAfterVote(oldVoterInfo, newVoterInfo, new(big.Int)) {
			return errors.New("投票者改变生产者后，投票者的投票结果信息不对")
		}
		return nil
	}, big.NewInt(0), true)

	if err != nil {
		return errors.New("更新生产者发生错误：" + err.Error())
	}

	return nil
}

func UnvoteProducer(voter string, voterPriKey *ecdsa.PrivateKey) error {
	oldVoterInfo, err := GetVoterInfo(voter)
	if err != nil || !isVoter(oldVoterInfo) {
		return errors.New("投票者取消投票前无法获取投票者信息：" + err.Error())
	}
	stake := oldVoterInfo["Quantity"].(*big.Int)

	err = runTxFuncThenCheckGlobalInfo(func() error {
		oldProducer := oldVoterInfo["Producer"].(string)
		oldProducerInfo, err := GetProducerInfo(oldProducer)
		if err != nil || !isProducer(oldProducerInfo) {
			return errors.New("投票者取消投票前无法获取原生产者信息：" + err.Error())
		}

		err = runTxAndCheckReceipt(func() (common.Hash, error) {
			txHash, err := unvoteProducer(voter, voterPriKey)
			if err != nil {
				return common.Hash{}, errors.New("投票者取消投票的交易失败：" + err.Error())
			}
			return txHash, nil
		}, voter)

		if err != nil {
			return err
		}

		newVoterInfo, err := GetVoterInfo(voter)
		if isVoter(newVoterInfo) {
			return errors.New("取消投票后竟然还能获取投票者信息")
		}
		newProducerInfo, err := GetProducerInfo(oldProducer)
		if err != nil || !isProducer(oldProducerInfo) {
			return errors.New("取消投票后无法获取新生产者信息：" + err.Error())
		}

		if !compareProducerAfterVote(oldProducerInfo, newProducerInfo, stake.Mul(stake, big.NewInt(-1))) {
			return errors.New("投票者取消投票后，生产者的投票结果信息不对")
		}
		return nil
	}, stake, false)

	if err != nil {
		return errors.New("投票者取消投票发生错误：" + err.Error())
	}

	return nil
}

func ProducerRemoveVoter(producer string, producerPriKey *ecdsa.PrivateKey, voter string) error {
	oldVoterInfo, err := GetVoterInfo(voter)
	if err != nil || !isVoter(oldVoterInfo) {
		return errors.New("生产者在取消投票者的投票之前无法获取投票者信息：" + err.Error())
	}
	stake := oldVoterInfo["Quantity"].(*big.Int)

	err = runTxFuncThenCheckGlobalInfo(func() error {
		oldProducerInfo, err := GetProducerInfo(producer)
		if err != nil || !isProducer(oldProducerInfo) {
			return errors.New("无法获取原生产者信息：" + err.Error())
		}

		err = runTxAndCheckReceipt(func() (common.Hash, error) {
			txHash, err := removeVoter(producer, producerPriKey, voter)
			if err != nil {
				return common.Hash{}, errors.New("生产者取消投票的交易失败：" + err.Error())
			}
			return txHash, nil
		}, voter)

		if err != nil {
			return err
		}

		newVoterInfo, err := GetVoterInfo(voter)
		if isVoter(newVoterInfo) {
			return errors.New("生产者取消投票后竟然还能获取投票者信息")
		}
		newProducerInfo, err := GetProducerInfo(producer)
		if err != nil || !isProducer(oldProducerInfo) {
			return errors.New("生产者取消投票后无法获取新生产者信息：" + err.Error())
		}

		if !compareProducerAfterVote(oldProducerInfo, newProducerInfo, stake.Mul(stake, big.NewInt(-1))) {
			return errors.New("投票者取消投票后，生产者的投票结果信息不对")
		}
		return nil
	}, stake, false)

	if err != nil {
		return errors.New("投票者取消投票发生错误：" + err.Error())
	}

	return nil
}
