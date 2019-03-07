// Copyright 2018 The Fractal Team Authors
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

package dpos

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	lru "github.com/hashicorp/golang-lru"
)

var (
	errMissingSignature           = errors.New("extra-data 65 byte suffix signature missing")
	errMismatchSignerAndValidator = errors.New("mismatch block signer and producer")
	errInvalidMintBlockTime       = errors.New("invalid time to mint the block")
	errInvalidBlockProducer       = errors.New("invalid block producer")
	errInvalidTimestamp           = errors.New("invalid timestamp")
	ErrIllegalProducerName        = errors.New("illegal producer name")
	ErrIllegalProducerPubKey      = errors.New("illegal producer pubkey")
	ErrTooMuchRreversible         = errors.New("too much rreversible blocks")
	errUnknownBlock               = errors.New("unknown block")
	extraSeal                     = 65
	timeOfGenesisBlock            int64
)

type stateDB struct {
	name    string
	assetid uint64
	state   *state.StateDB
}

func (s *stateDB) GetSnapshot(key string, timestamp uint64) ([]byte, error) {
	return s.state.GetSnapshot(s.name, key, timestamp)
}

func (s *stateDB) Get(key string) ([]byte, error) {
	return s.state.Get(s.name, key)
}
func (s *stateDB) Put(key string, value []byte) error {
	s.state.Put(s.name, key, value)
	return nil
}
func (s *stateDB) Delete(key string) error {
	s.state.Delete(s.name, key)
	return nil
}
func (s *stateDB) Delegate(from string, amount *big.Int) error {
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return err
	}
	return accountDB.TransferAsset(common.StrToName(from), common.StrToName(s.name), s.assetid, amount)
}
func (s *stateDB) Undelegate(to string, amount *big.Int) error {
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return err
	}
	return accountDB.TransferAsset(common.StrToName(s.name), common.StrToName(to), s.assetid, amount)
}
func (s *stateDB) IncAsset2Acct(from string, to string, amount *big.Int) error {
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return err
	}
	return accountDB.IncAsset2Acct(common.StrToName(from), common.StrToName(to), s.assetid, amount)
}
func (s *stateDB) IsValidSign(name string, pubkey []byte) bool {
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return false
	}
	return accountDB.IsValidSign(common.StrToName(name), types.ActionType(0), common.BytesToPubKey(pubkey)) == nil
}

// Genesis dpos genesis store
func Genesis(cfg *Config, state *state.StateDB, height uint64) error {
	db := &LDB{
		IDatabase: &stateDB{
			name:  cfg.AccountName,
			state: state,
		},
	}
	if err := db.SetProducer(&producerInfo{
		Name:          cfg.SystemName,
		URL:           cfg.SystemURL,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
		Height:        0,
	}); err != nil {
		return err
	}

	activatedProducerSchedule := []string{}
	for i := uint64(0); i < cfg.ProducerScheduleSize; i++ {
		activatedProducerSchedule = append(activatedProducerSchedule, cfg.SystemName)
	}
	if err := db.SetState(&globalState{
		Height:                    height,
		ActivatedTotalQuantity:    big.NewInt(0),
		ActivatedProducerSchedule: activatedProducerSchedule,
	}); err != nil {
		return err
	}
	if err := db.SetState(&globalState{
		Height:                    height + 1,
		ActivatedTotalQuantity:    big.NewInt(0),
		ActivatedProducerSchedule: activatedProducerSchedule,
	}); err != nil {
		return err
	}
	return nil
}

// SignFn signature function
type SignFn func([]byte, *state.StateDB) ([]byte, error)

// Dpos dpos engine
type Dpos struct {
	rw sync.RWMutex

	signFn SignFn

	config *Config

	// cache
	bftIrreversibles *lru.Cache
}

// New creates a DPOS consensus engine
func New(config *Config, chain consensus.IChainReader) *Dpos {
	dpos := &Dpos{
		config: config,
	}
	dpos.bftIrreversibles, _ = lru.New(int(config.ProducerScheduleSize))
	return dpos
}

// SetConfig set dpos config
func (dpos *Dpos) SetConfig(config *Config) {
	dpos.rw.Lock()
	defer dpos.rw.Unlock()
	dpos.config = config
}

// Config return dpos config
func (dpos *Dpos) Config() *Config {
	dpos.rw.RLock()
	defer dpos.rw.RUnlock()
	return dpos.config
}

// SetSignFn set signature function
func (dpos *Dpos) SetSignFn(signFn SignFn) {
	dpos.rw.Lock()
	defer dpos.rw.Unlock()
	dpos.signFn = signFn
}

// Author implements consensus.Engine, returning the header's coinbase
func (dpos *Dpos) Author(header *types.Header) (common.Name, error) {
	return header.Coinbase, nil
}

// Prepare initializes the consensus fields of a block header according to the rules of a particular engine. The changes are executed inline.
func (dpos *Dpos) Prepare(chain consensus.IChainReader, header *types.Header, txs []*types.Transaction, receipts []*types.Receipt, state *state.StateDB) error {
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)
	return nil
}

// Finalize assembles the final block.
func (dpos *Dpos) Finalize(chain consensus.IChainReader, header *types.Header, txs []*types.Transaction, receipts []*types.Receipt, state *state.StateDB) (*types.Block, error) {
	if chain == nil {
		header.Root = state.IntermediateRoot()
		return types.NewBlock(header, txs, receipts), nil
	}
	parent := chain.GetHeaderByHash(header.ParentHash)

	sys := &System{
		config: dpos.config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:    dpos.config.AccountName,
				assetid: chain.Config().SysTokenID,
				state:   state,
			},
		},
	}

	counter := int64(0)
	parentEpoch := dpos.config.epoch(parent.Time.Uint64())
	currentEpoch := dpos.config.epoch(header.Time.Uint64())
	if parentEpoch != currentEpoch {
		tparent := parent
		for dpos.config.epoch(tparent.Number.Uint64()) == dpos.config.epoch(parent.Time.Uint64()) {
			counter++
			tparent = chain.GetHeaderByHash(tparent.ParentHash)
		}
		// next epoch
		sys.updateElectedProducers(header.Time.Uint64())
	}

	extraReward := new(big.Int).Mul(dpos.config.extraBlockReward(), big.NewInt(counter))
	reward := new(big.Int).Add(dpos.config.blockReward(), extraReward)
	sys.IncAsset2Acct(dpos.config.SystemName, header.Coinbase.String(), reward)
	sys.onblock(header.Number.Uint64())
	header.Root = state.IntermediateRoot()
	dpos.bftIrreversibles.Add(header.Coinbase, header.ProposedIrreversible)
	return types.NewBlock(header, txs, receipts), nil
}

// Seal generates a new block for the given input block with the local miner's seal place on top.
func (dpos *Dpos) Seal(chain consensus.IChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error) {
	header := block.Header()
	number := header.Number.Uint64()
	if number == 0 {
		return nil, errUnknownBlock
	}

	parent := chain.GetHeader(header.ParentHash, number-1)
	state, err := chain.StateAt(parent.Root)
	if err != nil {
		return nil, err
	}
	sighash, err := dpos.signFn(signHash(header, chain.Config().ChainID.Bytes()).Bytes(), state)
	if err != nil {
		return nil, err
	}
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	return block.WithSeal(header), nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to the consensus rules of the given engine.
func (dpos *Dpos) VerifySeal(chain consensus.IChainReader, header *types.Header) error {
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	parent := chain.GetHeader(header.ParentHash, number-1)
	if dpos.config.nextslot(parent.Time.Uint64()) > header.Time.Uint64() {
		return errInvalidTimestamp
	}
	proudcer := header.Coinbase.String()
	state, err := chain.StateAt(parent.Root)
	if err != nil {
		return err
	}

	pubkey, err := ecrecover(header, chain.Config().ChainID.Bytes())
	if err != nil {
		return err
	}

	if err := dpos.IsValidateProducer(chain, header.Number.Uint64()-1, header.Time.Uint64(), proudcer, [][]byte{pubkey}, state); err != nil {
		return err
	}

	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm.
// It returns the difficulty that a new block should have when created at time given the parent block's time and difficulty.
func (dpos *Dpos) CalcDifficulty(chain consensus.IChainReader, time uint64, parent *types.Header) *big.Int {
	// return the current height as difficulty
	if timeOfGenesisBlock == 0 {
		if genesisBlock := chain.GetHeaderByNumber(0); genesisBlock != nil {
			timeOfGenesisBlock = genesisBlock.Time.Int64()
		}
	}
	return big.NewInt((int64(time)-timeOfGenesisBlock)/int64(dpos.config.blockInterval()) + 1)
}

//IsValidateProducer current producer
func (dpos *Dpos) IsValidateProducer(chain consensus.IChainReader, height uint64, timestamp uint64, producer string, pubkeys [][]byte, state *state.StateDB) error {
	if timestamp%dpos.BlockInterval() != 0 {
		return errInvalidMintBlockTime
	}

	if height-dpos.CalcProposedIrreversible(chain) >= dpos.config.ProducerScheduleSize*dpos.config.BlockFrequency {
		return ErrTooMuchRreversible
	}

	db := &stateDB{
		name:  dpos.config.AccountName,
		state: state,
	}

	if !common.IsValidName(producer) {
		return ErrIllegalProducerName
	}

	has := false
	for _, pubkey := range pubkeys {
		if db.IsValidSign(producer, pubkey) {
			has = true
		}
	}
	if !has {
		return ErrIllegalProducerPubKey
	}

	targetTime := big.NewInt(int64(timestamp - dpos.config.DelayEcho*dpos.config.epochInterval()))
	// find target block
	var pheader *types.Header
	for height > 0 {
		pheader = chain.GetHeaderByNumber(height)
		if pheader.Time.Cmp(targetTime) != 1 {
			break
		} else {
			height--
		}
	}

	sys := &System{
		config: dpos.config,
		IDB: &LDB{
			IDatabase: db,
		},
	}

	gstate, err := sys.GetState(height)
	if err != nil {
		return err
	}
	offset := dpos.config.getoffset(timestamp)
	if gstate == nil || offset >= uint64(len(gstate.ActivatedProducerSchedule)) || strings.Compare(gstate.ActivatedProducerSchedule[offset], producer) != 0 {
		return fmt.Errorf("%v %v, except %v index %v (%v) ", errInvalidBlockProducer, producer, gstate.ActivatedProducerSchedule, offset, timestamp/dpos.config.epochInterval())
	}
	return nil
}

// BlockInterval block interval
func (dpos *Dpos) BlockInterval() uint64 {
	return dpos.config.blockInterval()
}

// Slot
func (dpos *Dpos) Slot(timestamp uint64) uint64 {
	return dpos.config.slot(timestamp)
}

// IsFirst the first of producer
func (dpos *Dpos) IsFirst(timestamp uint64) bool {
	return timestamp%dpos.config.epochInterval()%(dpos.config.blockInterval()*dpos.config.BlockFrequency) == 0
}

func (dpos *Dpos) GetDelegatedByTime(name string, timestamp uint64, state *state.StateDB) (*big.Int, error) {
	sys := &System{
		config: dpos.config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:  dpos.config.AccountName,
				state: state,
			},
		},
	}
	return sys.GetDelegatedByTime(name, timestamp)
}

// Engine an engine
func (dpos *Dpos) Engine() consensus.IEngine {
	return dpos
}

func (dpos *Dpos) CalcBFTIrreversible() uint64 {
	irreversibles := UInt64Slice{}
	keys := dpos.bftIrreversibles.Keys()
	for _, key := range keys {
		if irreversible, ok := dpos.bftIrreversibles.Get(key); ok {
			irreversibles = append(irreversibles, irreversible.(uint64))
		}
	}

	if len(irreversibles) == 0 {
		return 0
	}

	sort.Sort(irreversibles)

	/// 2/3 must be greater, so if I go 1/3 into the list sorted from low to high, then 2/3 are greater
	return irreversibles[(len(irreversibles)-1)/3]
}

func (dpos *Dpos) CalcProposedIrreversible(chain consensus.IChainReader) uint64 {
	curHeader := chain.CurrentHeader()
	state, _ := chain.StateAt(curHeader.Root)
	sys := &System{
		config: dpos.config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:    dpos.config.AccountName,
				assetid: chain.Config().SysTokenID,
				state:   state,
			},
		},
	}
	producerMap := make(map[string]uint64)
	for curHeader.Number.Uint64() > 0 {
		if curHeader.Number.Uint64() <= dpos.config.DelayEcho*dpos.config.ProducerScheduleSize*dpos.config.BlockFrequency {
			return curHeader.Number.Uint64()
		} else if gstate, _ := sys.GetState(curHeader.Number.Uint64() - dpos.config.DelayEcho*dpos.config.ProducerScheduleSize*dpos.config.BlockFrequency); !sys.isdpos(gstate) {
			return curHeader.Number.Uint64()
		}
		producerMap[curHeader.Coinbase.String()] = dpos.config.epoch(curHeader.Time.Uint64())
		if uint64(len(producerMap)) >= dpos.config.consensusSize() {
			return curHeader.Number.Uint64()
		}
		curHeader = chain.GetHeaderByHash(curHeader.ParentHash)
	}
	return 0
}

func ecrecover(header *types.Header, extra []byte) ([]byte, error) {
	// If the signature's already cached, return that
	if len(header.Extra) < extraSeal {
		return nil, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]
	pubkey, err := crypto.Ecrecover(signHash(header, extra).Bytes(), signature)
	if err != nil {
		return nil, err
	}
	return pubkey, nil
}

func signHash(header *types.Header, extra []byte) (hash common.Hash) {
	theader := types.CopyHeader(header)
	theader.Extra = theader.Extra[:len(theader.Extra)-extraSeal]
	return theader.Hash()
}

// UInt64Slice attaches the methods of sort.Interface to []uint64, sorting in increasing order.
type UInt64Slice []uint64

func (s UInt64Slice) Len() int           { return len(s) }
func (s UInt64Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s UInt64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
