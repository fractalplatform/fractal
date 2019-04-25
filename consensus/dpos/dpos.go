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

	"github.com/ethereum/go-ethereum/log"
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
	errMismatchSignerAndValidator = errors.New("mismatch block signer and candidate")
	errInvalidMintBlockTime       = errors.New("invalid time to mint the block")
	errInvalidBlockCandidate      = errors.New("invalid block candidate")
	errInvalidTimestamp           = errors.New("invalid timestamp")
	ErrIllegalCandidateName       = errors.New("illegal candidate name")
	ErrIllegalCandidatePubKey     = errors.New("illegal candidate pubkey")
	ErrTooMuchRreversible         = errors.New("too much rreversible blocks")
	ErrSystemTakeOver             = errors.New("system account take over")
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
func (s *stateDB) Undelegate(to string, amount *big.Int) (*types.Action, error) {
	action := types.NewAction(types.Transfer, common.StrToName(s.name), common.StrToName(to), 0, s.assetid, 0, amount, nil)
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return action, err
	}
	return action, accountDB.TransferAsset(common.StrToName(s.name), common.StrToName(to), s.assetid, amount)
}
func (s *stateDB) IncAsset2Acct(from string, to string, amount *big.Int) (*types.Action, error) {
	action := types.NewAction(types.IncreaseAsset, common.StrToName(s.name), common.StrToName(to), 0, s.assetid, 0, amount, nil)
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return action, err
	}
	return action, accountDB.IncAsset2Acct(common.StrToName(from), common.StrToName(to), s.assetid, amount)
}
func (s *stateDB) IsValidSign(name string, pubkey []byte) bool {
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return false
	}
	return accountDB.IsValidSign(common.StrToName(name), common.BytesToPubKey(pubkey)) == nil
}
func (s *stateDB) GetBalanceByTime(name string, timestamp uint64) (*big.Int, error) {
	accountDB, err := accountmanager.NewAccountManager(s.state)
	if err != nil {
		return big.NewInt(0), err
	}
	return accountDB.GetBalanceByTime(common.StrToName(name), s.assetid, 1, timestamp)
}

// Genesis dpos genesis store
func Genesis(cfg *Config, state *state.StateDB, timestamp uint64, height uint64) error {
	sys := NewSystem(state, cfg)
	if err := sys.SetCandidate(&CandidateInfo{
		Name:          cfg.SystemName,
		URL:           cfg.SystemURL,
		Quantity:      big.NewInt(0),
		TotalQuantity: big.NewInt(0),
		Height:        height,
	}); err != nil {
		return err
	}

	epcho := cfg.epoch(timestamp)
	if err := sys.SetState(&GlobalState{
		Epcho:                  epcho,
		PreEpcho:               epcho,
		ActivatedTotalQuantity: big.NewInt(0),
		TotalQuantity:          big.NewInt(0),
		Height:                 height,
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
	dpos.bftIrreversibles, _ = lru.New(int(config.CandidateScheduleSize))
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
	sys := NewSystem(state, dpos.config)
	parent := chain.GetHeaderByHash(header.ParentHash)
	pepcho := dpos.config.epoch(parent.Time.Uint64())
	epcho := dpos.config.epoch(header.Time.Uint64())
	if pepcho != epcho {
		log.Debug("UpdateElectedCandidates", "prev", pepcho, "curr", epcho, "height", parent.Number.Uint64(), "time", parent.Time.Uint64())
		sys.UpdateElectedCandidates(pepcho, epcho, parent.Number.Uint64())
	}
	return nil
}

// Finalize assembles the final block.
func (dpos *Dpos) Finalize(chain consensus.IChainReader, header *types.Header, txs []*types.Transaction, receipts []*types.Receipt, state *state.StateDB) (*types.Block, error) {
	if chain == nil {
		header.Root = state.IntermediateRoot()
		return types.NewBlock(header, txs, receipts), nil
	}
	sys := NewSystem(state, dpos.config)
	parent := chain.GetHeaderByHash(header.ParentHash)
	counter := int64(0)
	if parent.Number.Uint64() > 0 && (dpos.CalcProposedIrreversible(chain, parent, true) == 0 || header.Time.Uint64()-parent.Time.Uint64() > 2*dpos.config.mepochInterval()) {
		if systemio := strings.Compare(header.Coinbase.String(), dpos.config.SystemName) == 0; systemio {
			latest, err := sys.GetState(LastEpcho)
			if err != nil {
				return nil, err
			}
			latest.TakeOver = true
			if err := sys.SetState(latest); err != nil {
				return nil, err
			}
		}
	}

	candidate, err := sys.GetCandidate(header.Coinbase.String())
	if err != nil {
		return nil, err
	} else if candidate != nil {
		candidate.Counter++
		if err := sys.SetCandidate(candidate); err != nil {
			return nil, err
		}
	}

	extraReward := new(big.Int).Mul(dpos.config.extraBlockReward(), big.NewInt(counter))
	reward := new(big.Int).Add(dpos.config.blockReward(), extraReward)
	sys.IncAsset2Acct(dpos.config.SystemName, header.Coinbase.String(), reward)

	blk := types.NewBlock(header, txs, receipts)

	// first hard fork at a specific height
	// If the block height is greater than or equal to the hard forking height,
	// the fork function will take effect. This function is valid only in the test network.
	if err := chain.ForkUpdate(blk, state); err != nil {
		return nil, err
	}

	// update state root at the end
	blk.Head.Root = state.IntermediateRoot()
	if strings.Compare(header.Coinbase.String(), dpos.config.SystemName) == 0 {
		dpos.bftIrreversibles.Purge()
	}
	dpos.bftIrreversibles.Add(header.Coinbase, header.ProposedIrreversible)
	return blk, nil
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

	if err := dpos.IsValidateCandidate(chain, parent, header.Time.Uint64(), proudcer, [][]byte{pubkey}, state, true); err != nil {
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

//IsValidateCandidate current candidate
func (dpos *Dpos) IsValidateCandidate(chain consensus.IChainReader, parent *types.Header, timestamp uint64, candidate string, pubkeys [][]byte, state *state.StateDB, force bool) error {
	if timestamp%dpos.BlockInterval() != 0 {
		return errInvalidMintBlockTime
	}

	if !common.IsValidAccountName(candidate) {
		return ErrIllegalCandidateName
	}

	db := &stateDB{
		name:  dpos.config.AccountName,
		state: state,
	}
	has := false
	for _, pubkey := range pubkeys {
		if db.IsValidSign(candidate, pubkey) {
			has = true
		}
	}
	if !has {
		return ErrIllegalCandidatePubKey
	}

	sys := NewSystem(state, dpos.config)
	gstate, err := sys.GetState(LastEpcho)
	if err != nil {
		return err
	}

	systemio := strings.Compare(candidate, dpos.config.SystemName) == 0
	if gstate.TakeOver {
		if force && systemio {
			return nil
		}
		return ErrSystemTakeOver
	} else if parent.Number.Uint64() > 0 && (dpos.CalcProposedIrreversible(chain, parent, true) == 0 || timestamp-parent.Time.Uint64() > 2*dpos.config.mepochInterval()) {
		if force && systemio {
			// first take over
			return nil
		}
		return ErrTooMuchRreversible
	}

	pstate, err := sys.GetState(gstate.PreEpcho)
	if err != nil {
		return err
	}
	offset := dpos.config.getoffset(timestamp)
	if pstate == nil || offset >= uint64(len(pstate.ActivatedCandidateSchedule)) || strings.Compare(pstate.ActivatedCandidateSchedule[offset], candidate) != 0 {
		return fmt.Errorf("%v %v, except %v index %v (%v) ", errInvalidBlockCandidate, candidate, pstate.ActivatedCandidateSchedule, offset, pstate.Epcho)
	}
	return nil
}

// BlockInterval block interval
func (dpos *Dpos) BlockInterval() uint64 {
	return dpos.config.blockInterval()
}

// Slot slot
func (dpos *Dpos) Slot(timestamp uint64) uint64 {
	return dpos.config.slot(timestamp)
}

// IsFirst the first of candidate
func (dpos *Dpos) IsFirst(timestamp uint64) bool {
	return timestamp%(dpos.config.blockInterval()*dpos.config.BlockFrequency) == 0
}

// GetDelegatedByTime get delegate of candidate
func (dpos *Dpos) GetDelegatedByTime(candidate string, timestamp uint64, state *state.StateDB) (*big.Int, *big.Int, uint64, error) {
	sys := NewSystem(state, dpos.config)
	return sys.GetDelegatedByTime(candidate, timestamp)
}

// Engine an engine
func (dpos *Dpos) Engine() consensus.IEngine {
	return dpos
}

// CalcBFTIrreversible calc irreversible
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

// CalcProposedIrreversible calc irreversible
func (dpos *Dpos) CalcProposedIrreversible(chain consensus.IChainReader, parent *types.Header, strict bool) uint64 {
	curHeader := chain.CurrentHeader()
	if parent != nil {
		curHeader = chain.GetHeaderByHash(parent.Hash())
	}
	candidateMap := make(map[string]uint64)
	timestamp := curHeader.Time.Uint64()
	for curHeader.Number.Uint64() > 0 {
		if strings.Compare(curHeader.Coinbase.String(), dpos.config.SystemName) == 0 {
			return curHeader.Number.Uint64()
		}
		if strict && timestamp-curHeader.Time.Uint64() >= 2*dpos.config.mepochInterval() {
			break
		}
		candidateMap[curHeader.Coinbase.String()]++
		if uint64(len(candidateMap)) >= dpos.config.consensusSize() {
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
