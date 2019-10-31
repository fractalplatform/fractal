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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fractalplatform/fractal/blockchain/downloader"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/params"
	pm "github.com/fractalplatform/fractal/plugin"
	p "github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

const (

	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
	// BlockInterval todo move config
	BlockInterval = 3000000000
)

type context interface {
	chainContext
	txPoolContext
	p.Processor
}

type chainContext interface {
	Config() *params.ChainConfig
	CurrentHeader() *types.Header
	StateAt(common.Hash) (*state.StateDB, error)
	WriteBlockWithState(*types.Block, []*types.Receipt, *state.StateDB) (bool, error)
}

type txPoolContext interface {
	Pending() (map[string][]*types.Transaction, error)
}

func slot(timestamp uint64) uint64 {
	return ((timestamp + BlockInterval/10) / BlockInterval * BlockInterval)
}

// Worker is the main object which takes care of applying messages to the new state
type Worker struct {
	context
	manger        pm.IPM
	mu            sync.Mutex
	delayDuration uint64
	coinbase      string
	privKeys      []*ecdsa.PrivateKey
	pubKeys       [][]byte
	extra         []byte

	wg        sync.WaitGroup
	mining    int32
	quitWork1 chan struct{}
	quitWork  chan struct{}
	wgWork    sync.WaitGroup
	quit      chan struct{}
	force     bool
}

func newWorker(manger pm.IPM, c context) *Worker {
	worker := &Worker{
		context:   c,
		manger:    manger,
		quitWork1: make(chan struct{}),
		quit:      make(chan struct{}),
	}
	return worker
}

// update keeps track of events.
func (worker *Worker) update() {
	chainHeadCh := make(chan *event.Event, chainHeadChanSize)
	chainHeadSub := event.Subscribe(nil, chainHeadCh, event.ChainHeadEv, &types.Block{})
	defer chainHeadSub.Unsubscribe()
out:
	for {
		select {
		case ev := <-chainHeadCh:
			// Handle ChainHeadEvent
			if atomic.LoadInt32(&worker.mining) != 0 {
				if blk := ev.Data.(*types.Block); strings.Compare(blk.Coinbase(), worker.coinbase) != 0 {
					worker.quitWork1 <- struct{}{}
				}
			}
		case <-worker.quit:
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
	worker.quit = make(chan struct{})
	worker.force = force
	worker.wg.Add(1)
	go func() {
		worker.mintLoop()
		worker.wg.Done()
	}()
	worker.wg.Add(1)
	go func() {
		worker.update()
		worker.wg.Done()
	}()
}

func (worker *Worker) mintLoop() {
	interval := int64(BlockInterval)
	c := make(chan time.Time)
	to := time.Now()
	worker.utimerTo(to.Add(time.Duration(interval-(to.UnixNano()%interval))), c)
	for {
		select {
		case <-worker.quitWork1:
			if worker.quitWork != nil {
				log.Debug("old parent hash coming, will be closing current work")
				close(worker.quitWork)
				worker.quitWork = nil
			}
		case now := <-c:
			if worker.quitWork != nil {
				close(worker.quitWork)
				worker.quitWork = nil
			}
			worker.wgWork.Wait()
			worker.quitWork = make(chan struct{})
			timestamp := int64(slot(uint64(now.UnixNano())))
			worker.wg.Add(1)
			worker.wgWork.Add(1)
			go func(quit chan struct{}) {
				worker.mintBlock(timestamp, quit)
				worker.wgWork.Done()
				worker.wg.Done()
			}(worker.quitWork)
			to := time.Now()
			worker.utimerTo(to.Add(time.Duration(interval-(to.UnixNano()%interval))), c)
		case <-worker.quit:
			return
		}
	}
}

func (worker *Worker) mintBlock(timestamp int64, quit chan struct{}) {
	bstart := time.Now()
	log.Debug("mint block", "timestamp", timestamp)
	for {
		select {
		case <-worker.quit:
			return
		case <-quit:
			return
		default:
		}

		header := worker.CurrentHeader()
		// state, err := worker.StateAt(header.Root)
		// if err != nil {
		// 	log.Error("failed to mint block", "timestamp", timestamp, "err", err)
		// 	return
		// }
		// theader := &types.Header{}
		// if err := cdpos.IsValidateCandidate(worker, header, uint64(timestamp), worker.coinbase, worker.pubKeys, state, worker.force, theader.CurForkID()); err != nil {
		// 	switch err {
		// 	case dpos.ErrSystemTakeOver:
		// 		fallthrough
		// 	case dpos.ErrTooMuchRreversible:
		// 		fallthrough
		// 	case dpos.ErrIllegalCandidateName:
		// 		fallthrough
		// 	case dpos.ErrIllegalCandidatePubKey:

		// 		log.Warn("failed to mint the block", "timestamp", timestamp, "err", err, "candidate", worker.coinbase)
		// 	default:
		// 		log.Debug("failed to mint the block", "timestamp", timestamp, "err", err)
		// 	}
		// 	return
		// }
		block, err := worker.commitNewWork(timestamp, header, quit)
		if err == nil {
			log.Info("Mined new block", "candidate", block.Coinbase(), "number", block.Number(), "hash", block.Hash().String(), "time", block.Time().Int64(), "txs", len(block.Txs), "gas", block.GasUsed(), "diff", block.Difficulty(), "elapsed", common.PrettyDuration(time.Since(bstart)))
			break
		}
		if strings.Contains(err.Error(), "mint") {
			log.Error("failed to mint block", "timestamp", timestamp, "err", err)
			break
		} else if strings.Contains(err.Error(), "wait") {
			worker.usleepTo(time.Now().Add(time.Duration(BlockInterval / 10)))
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
	worker.wg.Wait()
}
func (worker *Worker) setDelayDuration(delay uint64) error {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.delayDuration = delay
	return nil
}

func (worker *Worker) setCoinbase(name string, privateKeys []*ecdsa.PrivateKey) {
	// _, _ := worker.StateAt(worker.CurrentHeader().Root)
	// 	manger := pm.NewPM(state)
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.coinbase = name
	worker.privKeys = privateKeys
	worker.pubKeys = nil
	for _, privateKey := range privateKeys {
		pubkey := crypto.FromECDSAPub(&privateKey.PublicKey)
		// 		// todo manger.Recover()
		// 		// if err := mgr.IsValidSign(common.StrToName(name), common.BytesToPubKey(pubkey)); err == nil {
		// 		// 	log.Info("setCoinbase[valid]", "coinbase", name, fmt.Sprintf("pubKey_%03d", index), common.BytesToPubKey(pubkey).String())
		// 		// } else {
		// 		// 	log.Warn("setCoinbase[invalid]", "coinbase", name, fmt.Sprintf("pubKey_%03d", index), common.BytesToPubKey(pubkey).String(), "detail", err)
		// 		// }
		worker.pubKeys = append(worker.pubKeys, pubkey)
	}
}

func (worker *Worker) setExtra(extra []byte) {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.extra = extra
}

func (worker *Worker) commitNewWork(timestamp int64, parent *types.Header, quit chan struct{}) (*types.Block, error) {
	// 	dpos := worker.Engine().(*dpos.Dpos)
	if t := time.Now(); t.UnixNano() >= timestamp+int64(BlockInterval) {
		return nil, fmt.Errorf("mint the ingore block, need %v, now %v, sub %v", timestamp, t.UnixNano(), t.Sub(time.Unix(timestamp/int64(time.Second), timestamp%int64(time.Second))))
	}
	if parent.Time.Int64() >= timestamp {
		return nil, errors.New("mint the old block")
	}
	if parent.Number.Uint64() > 0 &&
		parent.Time.Int64()+int64(BlockInterval) < timestamp &&
		time.Now().UnixNano()-timestamp <= 2*int64(BlockInterval)/5 {
		return nil, errors.New("wait for last block arrived")
	}

	number := parent.Number
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     new(big.Int).Add(number, big.NewInt(1)),
		GasLimit:   params.BlockGasLimit,
		Extra:      worker.extra,
		Time:       big.NewInt(timestamp),
		//	Difficulty: worker.CalcDifficulty(worker.IConsensus, uint64(timestamp), parent),
	}
	header.Coinbase = worker.coinbase
	// header.ProposedIrreversible = dpos.CalcProposedIrreversible(worker, parent, false)

	state, err := worker.StateAt(parent.Root)
	if err != nil {
		return nil, fmt.Errorf("get parent state %v, err: %v ", header.Root, err)
	}

	work := &Work{
		currentHeader:   header,
		currentState:    state,
		currentTxs:      []*types.Transaction{},
		currentReceipts: []*types.Receipt{},
		currentGasPool:  new(common.GasPool).AddGas(header.GasLimit),
		currentCnt:      0,
	}

	if err := worker.manger.Prepare(work.currentHeader, work.currentTxs, work.currentReceipts, work.currentState); err != nil {
		return nil, fmt.Errorf("prepare header for mining, err: %v", err)
	}

	start := time.Now()
	pending, err := worker.Pending()
	if err != nil {
		return nil, fmt.Errorf("got error when fetch pending transactions, err: %v", err)
	}
	var txsLen int
	for _, txs := range pending {
		txsLen = txsLen + len(txs)
	}
	log.Debug("worker get pending txs from txpool", "len", txsLen, "since", time.Since(start))

	txs := types.NewTransactionsByPriceAndNonce(pending)
	if err := worker.commitTransactions(work, txs, BlockInterval, quit); err != nil {
		return nil, err
	}

	if atomic.LoadInt32(&worker.mining) == 1 {
		blk, err := worker.manger.Finalize(parent, work.currentHeader, work.currentTxs, work.currentReceipts, work.currentState)
		if err != nil {
			return nil, fmt.Errorf("finalize block, err: %v", err)
		}

		work.currentBlock = blk

		block, err := worker.manger.Seal(work.currentBlock)
		if err != nil {
			return nil, fmt.Errorf("seal block, err: %v", err)
		}
		for _, r := range work.currentReceipts {
			for _, l := range r.Logs {
				l.BlockHash = block.Hash()
			}
		}
		for _, log := range work.currentState.Logs() {
			log.BlockHash = block.Hash()
		}

		if _, err := worker.WriteBlockWithState(block, work.currentReceipts, work.currentState); err != nil {
			return nil, fmt.Errorf("writing block to chain, err: %v", err)
		}
		time.Sleep(time.Duration(worker.delayDuration * uint64(time.Millisecond)))

		event.SendEvent(&event.Event{Typecode: event.ChainHeadEv, Data: block})
		event.SendEvent(&event.Event{Typecode: event.NewMinedEv, Data: downloader.NewMinedBlockEvent{
			Block: block,
		}})
		return block, nil
	}
	work.currentBlock = types.NewBlock(work.currentHeader, work.currentTxs, work.currentReceipts)
	return work.currentBlock, nil
}

func (worker *Worker) commitTransactions(work *Work, txs *types.TransactionsByPriceAndNonce, interval uint64, quit chan struct{}) error {
	var coalescedLogs []*types.Log
	endTimeStamp := work.currentHeader.Time.Uint64() + interval - 2*interval/5
	endTime := time.Unix((int64)(endTimeStamp)/(int64)(time.Second), (int64)(endTimeStamp)%(int64)(time.Second))
	for {
		select {
		case <-worker.quit:
			return fmt.Errorf("mint the quit block")
		case <-quit:
			return fmt.Errorf("mint the quit block")
		default:
		}
		if work.currentGasPool.Gas() < params.GasTableInstance.ActionGas {
			log.Debug("Not enough gas for further transactions", "have", work.currentGasPool, "want", params.GasTableInstance.ActionGas)
			break
		}

		if interval != math.MaxUint64 && uint64(time.Now().UnixNano()) >= endTimeStamp {
			log.Debug("Not enough time for further transactions", "timestamp", work.currentHeader.Time.Int64())
			break
		}

		// Retrieve the next transaction and abort if all done
		tx := txs.Peek()

		if tx == nil {
			break
		}

		action := tx.GetActions()[0]

		from := action.Sender()
		// Start executing the transaction
		work.currentState.Prepare(tx.Hash(), common.Hash{}, work.currentCnt)

		logs, err := worker.commitTransaction(work, tx, endTime)
		switch err {
		case vm.ErrExecOverTime:
			log.Trace("Skipping transaction exec over time", "hash", tx.Hash())
			txs.Pop()
		case common.ErrGasLimitReached:
			// Pop the current out-of-gas transaction without shifting in the next from the account
			log.Trace("Gas limit exceeded for current block", "sender", from)
			txs.Pop()

		case p.ErrNonceTooLow:
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", action.Nonce())
			txs.Shift()

		case p.ErrNonceTooHigh:
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

func (worker *Worker) commitTransaction(work *Work, tx *types.Transaction, endTime time.Time) ([]*types.Log, error) {
	snap := work.currentState.Snapshot()

	receipt, _, err := worker.ApplyTransaction(&work.currentHeader.Coinbase, work.currentGasPool, work.currentState, work.currentHeader, tx, &work.currentHeader.GasUsed, vm.Config{
		EndTime: endTime,
	})
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
	return params.BlockGasLimit
}

type Work struct {
	currentCnt      int
	currentGasPool  *common.GasPool
	currentHeader   *types.Header
	currentTxs      []*types.Transaction
	currentReceipts []*types.Receipt
	currentBlock    *types.Block
	currentState    *state.StateDB
}

func (worker *Worker) usleepTo(to time.Time) {
	for {
		select {
		case <-worker.quit:
			return
		default:
		}
		if time.Now().UnixNano() >= to.UnixNano() {
			break
		}
		time.Sleep(time.Millisecond)
	}
}
func (worker *Worker) utimerTo(to time.Time, c chan time.Time) {
	worker.wg.Add(1)
	go func(c chan time.Time) {
		worker.usleepTo(to)
		select {
		case c <- to:
		case <-worker.quit:
			//default: // worker.quit is closed
		}
		worker.wg.Done()
	}(c)
}
