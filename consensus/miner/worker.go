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

package miner

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/blockchain"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

const (
	// txChanSize is the size of channel listening to NewTxsEvent.
	txChanSize = 4096
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
)

// Worker is the main object which takes care of applying messages to the new state
type Worker struct {
	consensus.IConsensus

	mu       sync.Mutex
	coinbase string
	privKeys []*ecdsa.PrivateKey
	pubKeys  [][]byte
	extra    []byte

	currentWork *Work

	wg         sync.WaitGroup
	mining     int32
	quitWork   chan struct{}
	quitWorkRW sync.RWMutex
	quit       chan struct{}
	force      bool
}

func newWorker(consensus consensus.IConsensus) *Worker {
	worker := &Worker{
		IConsensus: consensus,
		quit:       make(chan struct{}),
	}
	go worker.update()
	worker.commitNewWork(time.Now().UnixNano(), nil)
	return worker
}

// update keeps track of events.
func (worker *Worker) update() {
	txsCh := make(chan *event.Event, txChanSize)
	txsSub := event.Subscribe(nil, txsCh, event.TxEv, []*types.Transaction{})
	defer txsSub.Unsubscribe()

	chainHeadCh := make(chan *event.Event, chainHeadChanSize)
	chainHeadSub := event.Subscribe(nil, chainHeadCh, event.ChainHeadEv, &types.Block{})
	defer chainHeadSub.Unsubscribe()
out:
	for {
		select {
		case ev := <-chainHeadCh:
			// Handle ChainHeadEvent
			if atomic.LoadInt32(&worker.mining) != 0 {
				if blk := ev.Data.(*types.Block); strings.Compare(blk.Coinbase().String(), worker.coinbase) != 0 {
					worker.quitWorkRW.Lock()
					if worker.quitWork != nil {
						log.Debug("old parent hash coming, will be closing current work", "timestamp", worker.currentWork.currentHeader.Time)
						close(worker.quitWork)
						worker.quitWork = nil
					}
					worker.quitWorkRW.Unlock()
				}
			}
		case ev := <-txsCh:
			// Apply transactions to the pending state if we're not mining.
			if atomic.LoadInt32(&worker.mining) == 0 {
				worker.wg.Wait()
				txs := make(map[common.Name][]*types.Transaction)
				for _, tx := range ev.Data.([]*types.Transaction) {
					action := tx.GetActions()[0]
					from := action.Sender()
					txs[from] = append(txs[from], tx)
				}
				worker.commitTransactions(worker.currentWork, types.NewTransactionsByPriceAndNonce(txs), math.MaxUint64)
			}
			// System stopped
		case <-txsSub.Err():
			break out
		case <-chainHeadSub.Err():
			break out
		}
	}
}

func (worker *Worker) start(force bool) {
	if !atomic.CompareAndSwapInt32(&worker.mining, 0, 1) {
		log.Warn("worker already started")
		return
	}
	worker.force = force
	go worker.mintLoop()
}

func (worker *Worker) mintLoop() {
	worker.wg.Add(1)
	defer worker.wg.Done()
	dpos, ok := worker.Engine().(*dpos.Dpos)
	if !ok {
		panic("only support dpos engine")
	}
	dpos.SetSignFn(func(content []byte, state *state.StateDB) ([]byte, error) {
		accountDB, err := accountmanager.NewAccountManager(state)
		if err != nil {
			return nil, err
		}
		for index, privKey := range worker.privKeys {
			if err := accountDB.IsValidSign(common.StrToName(worker.coinbase), common.BytesToPubKey(worker.pubKeys[index])); err == nil {
				return crypto.Sign(content, privKey)
			}
		}
		return nil, fmt.Errorf("not found match private key for sign")
	})
	interval := int64(dpos.BlockInterval())
	timer := time.NewTimer(time.Duration(interval - (time.Now().UnixNano() % interval)))
	defer timer.Stop()
	for {
		select {
		case now := <-timer.C:
			worker.quitWorkRW.Lock()
			if worker.quitWork != nil {
				close(worker.quitWork)
				worker.quitWork = nil
				log.Debug("next time coming, will be closing current work", "timestamp", worker.currentWork.currentHeader.Time)
			}
			worker.quitWorkRW.Unlock()
			quit := make(chan struct{})
			worker.mintBlock(int64(dpos.Slot(uint64(now.UnixNano()))), quit)
			timer.Reset(time.Duration(interval - (time.Now().UnixNano() % interval)))
		case <-worker.quit:
			worker.quit = make(chan struct{})
			return
		}
	}
}

func (worker *Worker) mintBlock(timestamp int64, quit chan struct{}) {
	worker.quitWorkRW.Lock()
	worker.quitWork = quit
	worker.quitWorkRW.Unlock()
	defer func() {
		worker.quitWorkRW.Lock()
		worker.quitWork = nil
		worker.quitWorkRW.Unlock()
	}()
	cdpos := worker.Engine().(*dpos.Dpos)
	header := worker.CurrentHeader()
	state, err := worker.StateAt(header.Root)
	if err != nil {
		log.Error("failed to mint block", "timestamp", timestamp, "err", err)
		return
	}
	if err := cdpos.IsValidateCandidate(worker, header, uint64(timestamp), worker.coinbase, worker.pubKeys, state, worker.force); err != nil {
		switch err {
		case dpos.ErrSystemTakeOver:
			fallthrough
		case dpos.ErrTooMuchRreversible:
			fallthrough
		case dpos.ErrIllegalCandidateName:
			fallthrough
		case dpos.ErrIllegalCandidatePubKey:
			log.Error("failed to mint the block", "timestamp", timestamp, "err", err)
		default:
			log.Debug("failed to mint the block", "timestamp", timestamp, "err", err)
		}
		return
	}

	bstart := time.Now()
outer:

	for {
		select {
		case <-quit:
			return
		default:
		}
		block, err := worker.commitNewWork(timestamp, quit)
		if err == nil {
			log.Info("Mined new block", "candidate", block.Coinbase(), "number", block.Number(), "hash", block.Hash().String(), "time", block.Time().Int64(), "txs", len(block.Txs), "gas", block.GasUsed(), "diff", block.Difficulty(), "elapsed", common.PrettyDuration(time.Since(bstart)))
			break outer
		}
		if strings.Contains(err.Error(), "mint") {
			log.Error("failed to mint block", "timestamp", timestamp, "err", err)
			break outer
		} else if strings.Contains(err.Error(), "wait") {
			time.Sleep(time.Duration(cdpos.BlockInterval() / 10))
		}

		log.Warn("failed to mint block", "timestamp", timestamp, "err", err)
	}
}

func (worker *Worker) stop() {
	if !atomic.CompareAndSwapInt32(&worker.mining, 1, 0) {
		log.Warn("woker already stopped")
		return
	}
	close(worker.quit)
}

func (worker *Worker) setCoinbase(name string, privKeys []*ecdsa.PrivateKey) {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.coinbase = name
	worker.privKeys = privKeys
	worker.pubKeys = nil
	for _, privkey := range privKeys {
		worker.pubKeys = append(worker.pubKeys, crypto.FromECDSAPub(&privkey.PublicKey))
	}
}

func (worker *Worker) setExtra(extra []byte) {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.extra = extra
}

func (worker *Worker) pending() (*types.Block, *state.StateDB) {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	return worker.currentWork.currentBlock, worker.currentWork.currentState
}

func (worker *Worker) commitNewWork(timestamp int64, quit chan struct{}) (*types.Block, error) {
	parent := worker.CurrentHeader()
	dpos := worker.Engine().(*dpos.Dpos)
	if time.Now().UnixNano() >= timestamp+int64(dpos.BlockInterval()) {
		return nil, errors.New("mint the ingore block")
	}
	if parent.Time.Int64() >= timestamp {
		return nil, errors.New("mint the future block")
	}
	// if dpos.IsFirst(uint64(timestamp)) && parent.Time.Int64() != timestamp-int64(dpos.BlockInterval()) && timestamp-time.Now().UnixNano() >= int64(dpos.BlockInterval())/10 {
	if parent.Number.Uint64() > 0 && dpos.IsFirst(uint64(timestamp)) && parent.Time.Int64() != timestamp-int64(dpos.BlockInterval()) && time.Now().UnixNano()-timestamp <= 2*int64(dpos.BlockInterval())/5 {
		return nil, errors.New("wait for last block arrived")
	}

	number := parent.Number
	pblk := worker.GetBlock(parent.Hash(), parent.Number.Uint64())
	if pblk == nil {
		log.Error("parent is nil", "number", parent.Number.Uint64())
		return nil, errors.New("parent is nil")
	}
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     new(big.Int).Add(number, big.NewInt(1)),
		GasLimit:   worker.calcGasLimit(pblk),
		Extra:      worker.extra,
		Time:       big.NewInt(timestamp),
		Difficulty: worker.CalcDifficulty(worker.IConsensus, uint64(timestamp), parent),
	}
	if common.IsValidAccountName(worker.coinbase) {
		header.Coinbase = common.StrToName(worker.coinbase)
		header.ProposedIrreversible = dpos.CalcProposedIrreversible(worker, parent, false)
	}
	state, err := worker.StateAt(parent.Root)
	if err != nil {
		return nil, fmt.Errorf("get parent state %v, err: %v ", header.Root, err)
	}

	// fill ForkID
	if err := worker.FillForkID(header, state); err != nil {
		return nil, err
	}

	work := &Work{
		currentHeader:   header,
		currentState:    state,
		currentTxs:      []*types.Transaction{},
		currentReceipts: []*types.Receipt{},
		currentGasPool:  new(common.GasPool).AddGas(header.GasLimit),
		currentCnt:      0,
		quit:            quit,
	}
	worker.mu.Lock()
	worker.currentWork = work
	worker.mu.Unlock()

	if err := worker.Prepare(worker.IConsensus, work.currentHeader, work.currentTxs, work.currentReceipts, work.currentState); err != nil {
		return nil, fmt.Errorf("prepare header for mining, err: %v", err)
	}

	pending, err := worker.Pending()
	if err != nil {
		return nil, fmt.Errorf("got error when fetch pending transactions, err: %v", err)
	}

	txs := types.NewTransactionsByPriceAndNonce(pending)
	if err := worker.commitTransactions(work, txs, dpos.BlockInterval()); err != nil {
		return nil, err
	}

	if atomic.LoadInt32(&worker.mining) == 1 {
		blk, err := worker.Finalize(worker.IConsensus, work.currentHeader, work.currentTxs, work.currentReceipts, work.currentState)
		if err != nil {
			return nil, fmt.Errorf("finalize block, err: %v", err)
		}

		work.currentBlock = blk

		block, err := worker.Seal(worker.IConsensus, work.currentBlock, nil)
		if err != nil {
			return nil, fmt.Errorf("seal block, err: %v", err)
		}
		var logs []*types.Log
		for _, r := range work.currentReceipts {
			for _, l := range r.Logs {
				l.BlockHash = block.Hash()
			}
			logs = append(logs, r.Logs...)
		}
		for _, log := range work.currentState.Logs() {
			log.BlockHash = block.Hash()
		}

		if bytes.Compare(block.ParentHash().Bytes(), worker.CurrentHeader().Hash().Bytes()) != 0 {
			return nil, fmt.Errorf("old parent hash")
		}
		if _, err := worker.WriteBlockWithState(block, work.currentReceipts, work.currentState); err != nil {
			return nil, fmt.Errorf("writing block to chain, err: %v", err)
		}

		event.SendEvent(&event.Event{Typecode: event.ChainHeadEv, Data: block})
		event.SendEvent(&event.Event{Typecode: event.NewMinedEv, Data: blockchain.NewMinedBlockEvent{
			Block: block,
		}})
		return block, nil
	}
	block := types.NewBlock(work.currentHeader, work.currentTxs, work.currentReceipts)
	work.currentBlock = block
	return block, nil
}

func (worker *Worker) commitTransactions(work *Work, txs *types.TransactionsByPriceAndNonce, interval uint64) error {
	var coalescedLogs []*types.Log
	for {
		select {
		case <-work.quit:
			return fmt.Errorf("mined block timestamp %v missing --- signal", work.currentHeader.Time.Int64())
		default:
		}
		if work.currentGasPool.Gas() < params.ActionGas {
			log.Debug("Not enough gas for further transactions", "have", work.currentGasPool, "want", params.ActionGas)
			break
		}

		if interval != math.MaxUint64 && uint64(time.Now().UnixNano())+2*interval/5 >= work.currentHeader.Time.Uint64()+interval {
			log.Debug("Not enough time for further transactions", "timestamp", work.currentHeader.Time.Int64())
			break
		}

		// Retrieve the next transaction and abort if all done
		tx := txs.Peek()

		if tx == nil {
			break
		}

		action := tx.GetActions()[0]

		if strings.Compare(work.currentHeader.Coinbase.String(), worker.Config().SysName) != 0 {
			switch action.Type() {
			case types.KickedCandidate:
				fallthrough
			case types.ExitTakeOver:
				continue
			default:
			}
		}

		from := action.Sender()
		// Start executing the transaction
		work.currentState.Prepare(tx.Hash(), common.Hash{}, work.currentCnt)

		logs, err := worker.commitTransaction(work, tx)
		switch err {
		case common.ErrGasLimitReached:
			// Pop the current out-of-gas transaction without shifting in the next from the account
			log.Trace("Gas limit exceeded for current block", "sender", from)
			txs.Pop()

		case processor.ErrNonceTooLow:
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", action.Nonce())
			txs.Shift()

		case processor.ErrNonceTooHigh:
			// Reorg notification data race between the transaction pool and miner, skip account =
			log.Trace("Skipping account with hight nonce", "sender", from, "nonce", action.Nonce())
			txs.Pop()

		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
			coalescedLogs = append(coalescedLogs, logs...)
			work.currentCnt++
			txs.Shift()

		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			log.Debug("Transaction failed, account skipped", "hash", tx.Hash(), "err", err)
			txs.Shift()
		}
	}

	_ = coalescedLogs
	return nil
}

func (worker *Worker) commitTransaction(work *Work, tx *types.Transaction) ([]*types.Log, error) {
	snap := work.currentState.Snapshot()
	var name *common.Name
	if common.IsValidAccountName(work.currentHeader.Coinbase.String()) {
		name = new(common.Name)
		*name = common.StrToName(work.currentHeader.Coinbase.String())
	}

	receipt, _, err := worker.ApplyTransaction(name, work.currentGasPool, work.currentState, work.currentHeader, tx, &work.currentHeader.GasUsed, vm.Config{})
	if err != nil {
		work.currentState.RevertToSnapshot(snap)
		return nil, err
	}
	work.currentTxs = append(work.currentTxs, tx)
	work.currentReceipts = append(work.currentReceipts, receipt)
	return receipt.Logs, nil
}

func (worker *Worker) calcGasLimit(parent *types.Block) uint64 {
	if atomic.LoadInt32(&worker.mining) == 0 {
		return math.MaxUint64
	}
	return worker.IConsensus.CalcGasLimit(parent)
}

type Work struct {
	currentCnt      int
	currentGasPool  *common.GasPool
	currentHeader   *types.Header
	currentTxs      []*types.Transaction
	currentReceipts []*types.Receipt
	currentBlock    *types.Block
	currentState    *state.StateDB
	quit            chan struct{}
}
