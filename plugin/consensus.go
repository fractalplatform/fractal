// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// 1. 基准时间轴 o
// 2. 严格时间轴 o
// 3. 不出块则跳过时间窗 o
// 4. 后续出块人依次出块并包含投票(交易?) o

// 1. 缺块空时间窗(VDF?)
// 2. 总值 = t * p ; max(p) = 100
// 3. 支持合约？

// TODO:
// 1. 签名
// 2. 出块公钥
// 3. 难度
// 4. 回退

const (
	ConsensusKey     = "consensus"
	CandidateKey     = "candidates"
	LackBlock        = "lackblock"
	Epoch            = "Epoch"
	CandidateInfoKey = "info_"
	// account
	MinerAssetID = uint64(0)
)

var (
	minLockAmount = big.NewInt(1)
	maxCandidates = 32
	maxMiner      = uint64(21)
	maxWeight     = uint64(100)
	minWeight     = 0
	blockDuration = uint64(3)
	MinerAccount  = "fractaldpos"
	genesisTime   uint64
	maxPauseBlock = 864000
)

type CandidateInfo struct {
	OwnerAccount   string
	SignAccount    string
	RegisterNumber uint64
	Weight         uint64
	Balance        *big.Int
	Epoch          uint64
	//	Skip           bool
}

func (info *CandidateInfo) copy() *CandidateInfo {
	ret := *info
	ret.Balance = new(big.Int).Set(info.Balance)
	return &ret
}

func (info *CandidateInfo) update(newinfo *CandidateInfo) {
	if len(newinfo.SignAccount) != 0 {
		info.SignAccount = newinfo.SignAccount
	}
	info.RegisterNumber = newinfo.RegisterNumber
	if newinfo.Balance.Sign() > 0 {
		totalSum := info.WeightedSum()
		totalSum.Add(totalSum, newinfo.WeightedSum())
		info.Balance.Add(info.Balance, newinfo.Balance)
		info.Weight = totalSum.Div(totalSum, info.Balance).Uint64()
	}
	info.SignAccount = newinfo.SignAccount
}

func (info *CandidateInfo) IncWeight() uint64 {
	if info.Weight >= maxWeight {
		return maxWeight
	}
	info.Weight++
	return info.Weight
}

func (info *CandidateInfo) DecWeight() uint64 {
	info.Weight = info.Weight * 90 / 100
	return info.Weight
}

func (info *CandidateInfo) WeightedSum() *big.Int {
	z := big.NewInt(int64(info.Weight))
	return z.Mul(info.Balance, z)
}

func (info *CandidateInfo) Store(stateDB *state.StateDB) {
	b, _ := rlp.EncodeToBytes(info)
	stateDB.Put(ConsensusKey, CandidateInfoKey+info.OwnerAccount, b)
}

func (info *CandidateInfo) Load(stateDB *state.StateDB, owner string) {
	b, _ := stateDB.Get(ConsensusKey, CandidateInfoKey+owner)
	rlp.DecodeBytes(b, info)
}

func (info *CandidateInfo) signer() string {
	if len(info.SignAccount) == 0 {
		return info.OwnerAccount
	}
	return info.SignAccount
}

type Candidates struct {
	listSort []string
	info     map[string]*CandidateInfo
}

func (candidates *Candidates) Len() int {
	return len(candidates.listSort)
}

func (candidates *Candidates) Less(i, j int) bool {
	info_i := candidates.info[candidates.listSort[i]]
	info_j := candidates.info[candidates.listSort[j]]
	isless := info_i.WeightedSum().Cmp(info_j.WeightedSum())
	if isless == 0 {
		if info_i.RegisterNumber == info_j.RegisterNumber {
			return strings.Compare(info_i.OwnerAccount, info_j.OwnerAccount) < 0
		}
		return info_i.RegisterNumber > info_j.RegisterNumber
	}
	return isless < 0
}

func (candidates *Candidates) Swap(i, j int) {
	candidates.listSort[i], candidates.listSort[j] = candidates.listSort[j], candidates.listSort[i]
}

func (candidates *Candidates) sort() {
	sort.Sort(sort.Reverse(candidates))
}

func (candidates *Candidates) getInfoCopy(account string) *CandidateInfo {
	if info := candidates.info[account]; info != nil {
		return info.copy()
	}
	return nil
}

func (candidates *Candidates) insert(account string, newinfo *CandidateInfo) (bool, *CandidateInfo) {
	if _, exist := candidates.info[account]; !exist {
		candidates.listSort = append(candidates.listSort, account)
	}
	candidates.info[account] = newinfo
	candidates.sort()
	if candidates.Len() > maxCandidates {
		replaced := candidates.listSort[candidates.Len()-1]
		info := candidates.remove(replaced)
		return replaced != account, info
	}
	return true, nil // no one out
}

func (candidates *Candidates) remove(account string) *CandidateInfo {
	info, exist := candidates.info[account]
	if !exist {
		return nil
	}
	for i, name := range candidates.listSort {
		if name == account {
			copy(candidates.listSort[i:], candidates.listSort[i+1:])
			candidates.listSort = candidates.listSort[:candidates.Len()-1]
			delete(candidates.info, account)
			return info
		}
	}
	return nil // never goto here
}

type Consensus struct {
	isInit        bool
	BlockGasLimit uint64
	LackBlock     uint64
	candidates    *Candidates
	minerIndex    uint64
	minerOffset   uint64
	blockEpoch    uint64 // 轮数
	epochNum      uint64 // 该轮已出块数

	parent  *types.Header
	stateDB *state.StateDB

	rnd      *rand.Rand // just optimize
	rndCount int        // just optimize
	rndNum   int        // just optimize
}

func NewConsensus(stateDB *state.StateDB) *Consensus {
	c := &Consensus{
		parent:  nil,
		stateDB: stateDB,
		candidates: &Candidates{
			info: make(map[string]*CandidateInfo),
		},
	}
	c.loadCandidates()
	c.loadLackBlock()
	c.loadEpoch()
	for i, n := range c.candidates.listSort {
		info := &CandidateInfo{}
		info.Load(c.stateDB, n)
		c.candidates.info[n] = info
		if c.parent != nil && n == c.parent.Coinbase {
			c.minerIndex = uint64(i)
		}
	}
	return c
}

func (c *Consensus) AccountName() string {
	return MinerAccount
}

func (c *Consensus) initRequrie() {
	if !c.isInit {
		panic("Consensus need Init() before call")
	}
}

func (c *Consensus) Init(_genesisTime uint64, parent *types.Header) {
	if genesisTime == 0 {
		genesisTime = _genesisTime
		fmt.Println("genesisTime", genesisTime, MinerAccount)
	}
	c.parent = parent
	c.isInit = true

	for i, n := range c.candidates.listSort {
		if c.parent != nil && n == c.parent.Coinbase {
			c.minerIndex = uint64(i)
		}
	}
}

// return timestamp of parent+n
func (c *Consensus) timeSlot(epoch uint64) uint64 {
	ontime := genesisTime + (c.parent.Number+c.LackBlock+epoch)*blockDuration
	return ontime
}

// return miner of parent+n
// n = rndIndex
func (c *Consensus) minerSlot(n, epoch uint64) string {
	numMiner := maxMiner + epoch - 1
	if numMiner > uint64(c.candidates.Len()) {
		numMiner = uint64(c.candidates.Len())
	}
	index := (c.minerIndex + n) % numMiner
	return c.candidates.listSort[index]
}

func (c *Consensus) storeLackBlock() {
	b, _ := rlp.EncodeToBytes(c.LackBlock)
	c.stateDB.Put(ConsensusKey, LackBlock, b)
}
func (c *Consensus) loadLackBlock() {
	b, _ := c.stateDB.Get(ConsensusKey, LackBlock)
	rlp.DecodeBytes(b, &c.LackBlock)
}

func (c *Consensus) storeEpoch() {
	epoch := (c.blockEpoch << 8) + (c.epochNum & 0xff)
	b, _ := rlp.EncodeToBytes(epoch)
	c.stateDB.Put(ConsensusKey, Epoch, b)
}
func (c *Consensus) loadEpoch() {
	var epoch uint64
	b, _ := c.stateDB.Get(ConsensusKey, Epoch)
	rlp.DecodeBytes(b, &epoch)
	c.blockEpoch = epoch >> 8
	c.epochNum = epoch & 0xff
}

func (c *Consensus) storeCandidates() {
	b, _ := rlp.EncodeToBytes(c.candidates.listSort)
	c.stateDB.Put(ConsensusKey, CandidateKey, b)
}

func (c *Consensus) loadCandidates() {
	b, _ := c.stateDB.Get(ConsensusKey, CandidateKey)
	rlp.DecodeBytes(b, &c.candidates.listSort)
}

func (c *Consensus) removeCandidate(delCandidate string) (bool, *CandidateInfo) {
	if c.candidates.Len() == 0 {
		return false, nil // impossible?
	}
	if info := c.candidates.info[delCandidate]; info != nil {
		if c.parent.Number-info.RegisterNumber < (15*24*3600)/blockDuration {
			return false, nil
		}
	}
	info := c.candidates.remove(delCandidate)
	return info != nil, info
}

func (c *Consensus) pushCandidate(newCandidate string, signAccount string, lockAmount *big.Int) (bool, *CandidateInfo, *CandidateInfo) {
	newinfo := &CandidateInfo{
		OwnerAccount: newCandidate,
		SignAccount:  signAccount,
		Weight:       90,
		Balance:      lockAmount,
		Epoch:        0,
	}
	if c.parent != nil {
		newinfo.RegisterNumber = c.parent.Number + 1
	}
	if oldinfo := c.candidates.getInfoCopy(newCandidate); oldinfo != nil {
		oldinfo.update(newinfo)
		newinfo = oldinfo
	}
	if newinfo.WeightedSum().Cmp(minLockAmount) < 0 {
		return false, nil, nil
	}
	success, replaced := c.candidates.insert(newCandidate, newinfo)
	return success, newinfo, replaced
}

func (c *Consensus) nIndex(n int) int {
	if c.rnd == nil || c.rndCount > n {
		c.rnd = c.pseudoRand()
		c.rndCount = 0
		c.rndNum = c.rnd.Int()
	}
	for c.rndCount < n {
		c.rndNum = c.rnd.Int()
		c.rndCount++
	}
	return c.rndNum
}

func (c *Consensus) epochToIndex(epoch int) (int, int) {
	rndIndex := c.nIndex(epoch)
	change := make(map[string]uint64)
	for j := 0; j < c.candidates.Len(); j++ {
		minerIndex := rndIndex + j
		miner := c.minerSlot(uint64(minerIndex), uint64(epoch))
		minerEpoch, exist := change[miner]
		if !exist {
			minerEpoch = c.candidates.info[miner].Epoch
		}
		fmt.Println("len", c.candidates.Len(), "rnd_i", rndIndex, "epoch", epoch, "plus", j, "minerEpoch", minerEpoch, "blockEpoch", c.blockEpoch, "epochNum", c.epochNum)
		if minerEpoch <= c.blockEpoch {
			return epoch, minerIndex
		}
		change[miner] = c.blockEpoch + 1
	}
	return epoch, rndIndex + c.candidates.Len()
}

// return next miner
func (c *Consensus) nextMiner() (int, int) {
	now := uint64(time.Now().Unix())
	for i := 1; i <= c.candidates.Len()+maxPauseBlock; i++ {
		nextTimeout := c.timeSlot(uint64(i))
		if now < nextTimeout {
			return c.epochToIndex(i)
		}
	}
	return -1, -1
}

/*
// return distance between miner and parent.Coinbase
func (c *Consensus) searchMiner(miner string) int {
	for i := 1; i <= c.candidates.Len()+maxPauseBlock; i++ {
		if miner == c.minerSlot(uint64(i)) {
			if c.candidates.info[miner].Skip {
				return -1
			}
			return i
		}
	}
	return -1
}
*/

var pn = 0

func xpanic(n int) {
	pn++
	if pn > n {
		panic(pn)
	}
}

func (c *Consensus) Show(miner string, nextMiner string) {

	//xpanic(10)

	fmt.Println("-----------------")
	fmt.Println("parent:", c.parent.Number)
	fmt.Println("parent:", c.parent.Time)
	fmt.Println("now:", time.Now().Unix())
	fmt.Println("LackBlock:", c.LackBlock)
	fmt.Println("candidates:", c.candidates.Len())
	fmt.Println("minerIndex:", c.minerIndex)
	fmt.Println("miner:", miner)
	fmt.Println("genesisTime", genesisTime, MinerAccount)
	for i, n := range c.candidates.listSort {
		info := c.candidates.info[n]
		fmt.Print("\t")
		align := "   "
		if n == nextMiner {
			fmt.Print(">")
			align = align[1:]
		}
		if n == miner {
			fmt.Print("*")
			align = align[1:]
		}
		fmt.Print(align)
		fmt.Printf("%02d %s %v %v\n", i, n, info.WeightedSum(), info)
	}
}

func (c *Consensus) pseudoRand() *rand.Rand {
	c.initRequrie()
	return rand.New(rand.NewSource(new(big.Int).SetBytes(c.parent.Proof).Int64()))
}

// 1. epoch: 表示出块slot
// 2. rnd: 表示通过该epoch得出的miner序号
func (c *Consensus) MineDelay(miner string) time.Duration {
	// just beta
	c.initRequrie()

	now := time.Now().Unix()
	epoch, rndIndex := c.nextMiner()
	if epoch < 1 {
		fmt.Println("epoch-wrong:", epoch)
		return time.Duration(blockDuration) * time.Second
	}
	nextMiner := c.minerSlot(uint64(rndIndex), uint64(epoch))

	c.Show(miner, nextMiner)

	if nextMiner == miner {
		ontime := int64(c.timeSlot(uint64(epoch) - 1))
		if ontime > now {
			fmt.Println("epoch-ready:", epoch, ontime, now)
			return time.Duration(ontime-now) * time.Second
		}
		fmt.Println("epoch-go:", epoch, ontime, now)
		c.minerOffset = uint64(epoch)
		return 0
	}
	fmt.Println("epoch-wait:", epoch, c.timeSlot(uint64(epoch)), now)
	return time.Duration(int64(c.timeSlot(uint64(epoch)))-now) * time.Second
}

func (c *Consensus) Prepare(header *types.Header) error {
	// just beta
	c.initRequrie()
	minerIndex := c.minerOffset
	if minerIndex == 0 {
		minerIndex = c.toOffset(header.Difficulty)
	}
	if minerIndex == 0 {
		return errors.New("minerIndex must greater than zero")
	}
	/*
		minerIndex := c.searchMiner(miner)
		if minerIndex < 0 {
			return nil
		}
	*/
	header.Difficulty = c.toDifficult(minerIndex)
	header.Time = c.timeSlot(minerIndex) // this code must be here
	header.ParentHash = c.parent.Hash()
	header.Number = c.parent.Number + 1
	header.GasLimit = params.BlockGasLimit

	miner := header.Coinbase
	start := time.Now().Unix()
	for i := uint64(1); i < minerIndex; i++ {
		_, rndIndex := c.epochToIndex(int(i))
		if time.Now().Unix()-start > 2 {
			start = time.Now().Unix()
			return errors.New("too long to Prepare")
		}
		skipMiner := c.minerSlot(uint64(rndIndex), i)
		info := c.candidates.info[skipMiner]
		info.Epoch = c.blockEpoch + 1
		info.DecWeight()
		info.Store(c.stateDB)
	}

	if minerIndex > 1 {
		c.LackBlock += uint64(minerIndex) - 1
		c.storeLackBlock()
	}
	info := c.candidates.info[miner]
	info.IncWeight()
	info.Epoch = c.blockEpoch + 1
	info.Store(c.stateDB)
	c.epochNum++
	halfMiner := maxMiner / 2
	if halfMiner > uint64(c.candidates.Len()) {
		halfMiner = uint64(c.candidates.Len())
	}
	if c.epochNum >= halfMiner {
		c.blockEpoch++
		c.epochNum = 0
	}
	c.storeEpoch()
	c.candidates.sort()
	c.storeCandidates()
	/*
		if priKey != nil {
			header.Proof = crypto.VRF_Proof(priKey, c.parent.SignHash())
		}
	*/
	return nil
}

func (c *Consensus) CallTx(tx *envelope.PluginTx, pm IPM) ([]byte, error) {
	// just beta
	c.initRequrie()
	var success bool
	var info, newinfo *CandidateInfo
	switch tx.PayloadType() {
	case RegisterMiner:
		if tx.Value().Sign() > 0 {
			if tx.GetAssetID() != MinerAssetID {
				return nil, fmt.Errorf("assetID must be %d", MinerAssetID)
			}
			if err := pm.TransferAsset(tx.Sender(), tx.Recipient(), tx.GetAssetID(), tx.Value()); err != nil {
				return nil, err
			}
		}
		var signAccount string
		if len(tx.GetPayload()) > 0 {
			if err := rlp.DecodeBytes(tx.GetPayload(), &signAccount); err != nil {
				return nil, err
			}
		}
		if len(signAccount) > 0 {
			err := pm.AccountIsExist(signAccount)
			if err != nil {
				return nil, err
			}
		}
		success, newinfo, info = c.pushCandidate(tx.Sender(), signAccount, tx.Value())
	case UnregisterMiner:
		if tx.Value().Sign() > 0 {
			return nil, errors.New("msg.value must be zero")
		}
		success, info = c.removeCandidate(tx.Sender())
	default:
		return nil, ErrWrongTransaction
	}
	if !success {
		return nil, errors.New("wrong candidate")
	}
	c.storeCandidates()
	if newinfo != nil {
		newinfo.Store(c.stateDB)
	}
	if info != nil {
		if err := pm.TransferAsset(MinerAccount, info.OwnerAccount, MinerAssetID, info.Balance); err != nil {
			return nil, err
		}
		info.Balance = big.NewInt(0)
		info.Store(c.stateDB)
	}
	return nil, nil
}

// Finalize assembles the final block.
func (c *Consensus) Finalize(header *types.Header, txs []*types.Transaction, receipts []*types.Receipt) (*types.Block, error) {
	// just beta
	c.initRequrie()
	// info.Dec or Inc
	header.Root = c.stateDB.IntermediateRoot()
	return types.NewBlock(header, txs, receipts), nil
}

func (c *Consensus) toDifficult(offset uint64) uint64 {
	return uint64(c.candidates.Len()+maxPauseBlock) - offset + 1
}

func (c *Consensus) toOffset(difficult uint64) uint64 {
	max := uint64(c.candidates.Len() + maxPauseBlock)
	if max >= difficult {
		return max - difficult + 1
	}
	return 0
}

func (c *Consensus) Seal(block *types.Block, priKey *ecdsa.PrivateKey, pm IPM) (*types.Block, error) {
	// just beta
	c.initRequrie()

	miner := block.Coinbase()
	signerInfo, exist := c.candidates.info[miner]
	if !exist {
		return block, errors.New("illegal miner")
	}
	var err error
	block.Head.Proof = crypto.VRF_Proof(priKey, c.parent.Hash().Bytes())
	block.Head.Sign, err = pm.AccountSign(signerInfo.signer(), priKey, pm, block.Header().SignHash)
	return block, err
}

func (c *Consensus) Verify(header *types.Header) error {
	// just beta
	c.initRequrie()

	// TODO: verify header
	// 1. verify number
	if header.Number != c.parent.Number+1 {
		return fmt.Errorf("wrong block.Number, get %d want %d", header.Number, c.parent.Number+1)
	}
	// 2. verify Parent hash
	if header.ParentHash != c.parent.Hash() {
		return fmt.Errorf("wrong block.ParentHash, get %s want %s", header.ParentHash, c.parent.Hash())
	}
	// 3. verify miner
	/*
		minerIndex := c.searchMiner(miner)
		if minerIndex < 0 {
			return errors.New("can not find miner or miner skiped")
		}
	*/
	miner := header.Coinbase
	minerEpoch := c.toOffset(header.Difficulty)
	epoch, rndIndex := c.epochToIndex(int(minerEpoch))
	if c.minerSlot(uint64(rndIndex), uint64(epoch)) != miner {
		return errors.New("wrong miner")
	}
	// 4. verify block time
	timeSlot := c.timeSlot(uint64(minerEpoch))
	if header.Time != timeSlot {
		return fmt.Errorf("wrong block.Time, get %d want %d slot:%d", header.Time, timeSlot, minerEpoch)
	}
	now := time.Now().Unix()
	maxTime := uint64(now) + blockDuration*5
	if timeSlot > maxTime {
		return fmt.Errorf("wrong time slot, get %d want <=%d", timeSlot, maxTime)
	}
	// 5. verify weighSum?
	// 6. verify Sign?
	// 7. verify Version
	// 8. verify ExtData?
	return nil
}

func (c *Consensus) VerifySeal(header *types.Header, pm IPM) error {
	// just beta
	c.initRequrie()
	miner := header.Coinbase
	signerInfo, exist := c.candidates.info[miner]
	if !exist {
		return errors.New("illegal miner")
	}
	ecpub, err := pm.AccountVerify(signerInfo.signer(), pm, header.Sign, header.SignHash)
	if err != nil {
		return err
	}

	if !crypto.VRF_Verify(ecpub, c.parent.Hash().Bytes(), header.Proof) {
		return errors.New("VRF Verify error")
	}
	return nil
}

func (c *Consensus) Sol_Sprintf(_ interface{}, fmtstr string, name string, age *big.Int) (string, error) {
	return fmt.Sprintf(fmtstr, name, age.Int64()), nil
}
