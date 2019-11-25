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
	"sort"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
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
	Skip           bool
}

func (info *CandidateInfo) copy() *CandidateInfo {
	ret := *info
	ret.Balance = new(big.Int).Set(info.Balance)
	return &ret
}

func (info *CandidateInfo) update(newinfo *CandidateInfo) {
	info.SignAccount = newinfo.SignAccount
	info.RegisterNumber = newinfo.RegisterNumber
	if newinfo.Balance.Sign() > 0 {
		info.Balance.Add(info.Balance, newinfo.Balance)
		oldSum := info.WeightedSum()
		totalSum := oldSum.Add(oldSum, newinfo.WeightedSum())
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
			return strings.Compare(info_j.OwnerAccount, info_j.OwnerAccount) < 0
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
	parent        *types.Header
	stateDB       *state.StateDB
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
		info := &CandidateInfo{}
		info.Load(c.stateDB, n)
		c.candidates.info[n] = info
		if c.parent != nil && n == c.parent.Coinbase {
			c.minerIndex = uint64(i)
		}
	}
}

// return timestamp of parent+n
func (c *Consensus) timeSlot(n uint64) uint64 {
	ontime := genesisTime + (c.parent.Number+c.LackBlock+n)*blockDuration
	return ontime
}

// return miner of parent+n
func (c *Consensus) minerSlot(n uint64) string {
	numMiner := maxMiner
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

// return next miner
func (c *Consensus) nextMiner() int {
	now := uint64(time.Now().Unix())
	for i := 1; i <= c.candidates.Len()+maxPauseBlock; i++ {
		nextTimeout := c.timeSlot(uint64(i))
		if now < nextTimeout {
			for j := 0; j < c.candidates.Len(); j++ {
				miner := c.minerSlot(uint64(i + j))
				if c.candidates.info[miner].Skip {
					continue
				}
				fmt.Println("nextMiner", "i", i, "j", j, "now", now, "next", nextTimeout)
				return i + j
			}
			return -1
		}
	}
	return -1
}

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

func (c *Consensus) Show(miner string) {
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
		fmt.Println("\t", i, n, info.WeightedSum(), info.Skip)
	}
}

func (c *Consensus) MineDelay(miner string) time.Duration {
	// just beta
	c.initRequrie()

	c.Show(miner)

	now := time.Now().Unix()
	i := c.nextMiner()
	if i < 1 {
		fmt.Println("i<1:", i)
		return time.Duration(int64(c.timeSlot(1))-now) * time.Second
	}
	nextMiner := c.minerSlot(uint64(i))
	if nextMiner == miner {
		ontime := int64(c.timeSlot(uint64(i) - 1))
		if ontime > now {
			fmt.Println("i-1:", i, ontime, now)
			return time.Duration(ontime-now) * time.Second
		}
		fmt.Println("i-x:", i, ontime, now)
		c.minerOffset = uint64(i)
		return 0
	}
	fmt.Println("i-2:", i)
	return time.Duration(int64(c.timeSlot(uint64(i)))-now) * time.Second
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
	for i := uint64(1); i < minerIndex; i++ {
		skipMiner := c.minerSlot(i)
		if skipMiner == miner {
			continue
		}
		info := c.candidates.info[skipMiner]
		if !info.Skip {
			info.Skip = true // skip and dec weight
			info.DecWeight()
		} else {
			info.Skip = false // skip and reset skip
		}
		info.Store(c.stateDB)
	}
	if minerIndex > 1 {
		c.LackBlock += uint64(minerIndex) - 1
		c.minerIndex += uint64(minerIndex) - 1
		c.storeLackBlock()
	}
	info := c.candidates.info[miner]
	info.IncWeight()
	info.Store(c.stateDB)
	c.candidates.sort()
	c.storeCandidates()
	return nil
}

func (c *Consensus) CallTx(action *types.Action, pm IPM) ([]byte, error) {
	// just beta
	c.initRequrie()

	if action.Value().Sign() > 0 {
		if action.AssetID() != MinerAssetID {
			return nil, fmt.Errorf("assetID must be %d", MinerAssetID)
		}
		if err := pm.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value()); err != nil {
			return nil, err
		}
	}
	//TODO: return or lock balance
	var success bool
	var info, newinfo *CandidateInfo
	if action.Value().Sign() > 0 {
		fmt.Println("signer", action.Sender())
		success, newinfo, info = c.pushCandidate(action.Sender(), action.Sender(), action.Value())
	} else {
		success, info = c.removeCandidate(action.Sender())
	}
	if success {
		c.storeCandidates()
		if newinfo != nil {
			newinfo.Store(c.stateDB)
		}
		if info != nil {
			info.Store(c.stateDB)
			err := pm.TransferAsset(MinerAccount, info.OwnerAccount, MinerAssetID, info.Balance)
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("wrong candidate")
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
	signerAccount, err := pm.getAccount(signerInfo.SignAccount)
	if err != nil {
		fmt.Println("miner:", miner, "signer:", signerInfo.SignAccount, "err:", err)
		return block, err
	}
	keyAddress := crypto.PubkeyToAddress(priKey.PublicKey)
	if signerAccount.Address.Compare(keyAddress) != 0 {
		return block, errors.New("illegal private key")
	}
	block.Head.Sign, err = pm.Sign(block.Header().SignHash, priKey)
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
	minerIndex := c.toOffset(header.Difficulty)
	if c.minerSlot(minerIndex) != miner {
		return errors.New("wrong miner")
	}
	// 4. verify block time
	timeSlot := c.timeSlot(uint64(minerIndex))
	if header.Time != timeSlot {
		return fmt.Errorf("wrong block.Time, get %d want %d slot:%d", header.Time, timeSlot, minerIndex)
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
	signerAccount, err := pm.getAccount(signerInfo.SignAccount)
	if err != nil {
		return err
	}
	b, err := pm.Recover(header.Sign, header.SignHash)
	if err != nil {
		return err
	}
	recPub, _ := crypto.UnmarshalPubkey(b)
	recAddress := crypto.PubkeyToAddress(*recPub)
	if signerAccount.Address.Compare(recAddress) != 0 {
		return errors.New("illegal signature")
	}

	return nil
}
