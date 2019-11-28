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
	"sync"
	"sync/atomic"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/plugin"
	p "github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// TODO
// 1. stop
// 2. history header (parent)
// 3. sign

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
	GetBlockByNumber(number uint64) *types.Block
	StateAt(common.Hash) (*state.StateDB, error)
	WriteBlockWithState(*types.Block, []*types.Receipt, *state.StateDB) (bool, error)
}

type txPoolContext interface {
	Pending() (map[string][]*types.Transaction, error)
}

// Worker is the main object which takes care of applying messages to the new state
type Worker struct {
	context
	manger   plugin.IPM
	mu       sync.Mutex
	delayMs  time.Duration
	coinbase string
	priKey   *ecdsa.PrivateKey
	extra    []byte

	wg       sync.WaitGroup
	mining   int32
	quitWork chan struct{}
	wgWork   sync.WaitGroup
	quit     chan struct{}
	force    bool
}

func newWorker(c context) *Worker {
	worker := &Worker{
		context:  c,
		quitWork: make(chan struct{}, 1), // atleast 1
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
		case <-chainHeadCh:
			select {
			case worker.quitWork <- struct{}{}:
			default:
			}
		case <-worker.quit:
			break out
		case <-chainHeadSub.Err():
			break out
		}
	}
}

func (worker *Worker) start(force bool) {
	worker.quit = make(chan struct{})
	worker.force = force
	worker.wg.Add(2)
	go func() {
		worker.update()
		worker.wg.Done()
	}()
	go func() {
		worker.mintLoop()
		worker.wg.Done()
	}()
}

func (worker *Worker) mintLoop() {
	for {
		header := worker.CurrentHeader()
		state, err := worker.StateAt(header.Root)
		if err != nil {
			log.Error("Can't find state", "err", err, "root", header.Root, "hash", header.Hash(), "number", header.Number)
			return
		}
		fmt.Println("NewPM:", header.Number, header.Root)
		pm := plugin.NewPM(state)
		pm.Init(0, header)
		if delay := pm.MineDelay(worker.coinbase); delay > 0 {
			delayCh := time.NewTimer(delay)
			select {
			case <-delayCh.C:
			case <-worker.quit:
				return
			case <-worker.quitWork:
				continue
			}
			continue
		}
		worker.quitWork = make(chan struct{})
		worker.mintBlock(state, pm, header)
	}
}

func (worker *Worker) mintBlock(state *state.StateDB, pm plugin.IPM, header *types.Header) {
	bstart := time.Now()
	block, err := worker.commitNewWork(pm, state, header)
	if err == nil {

		{

			verifyState, err := worker.StateAt(header.Root)
			if err != nil {
				panic(err)
			}
			verifyPM := plugin.NewPM(verifyState)
			verifyPM.Init(0, header)
			err1 := verifyPM.VerifySeal(block.Header(), verifyPM)
			fmt.Println("VerifySeal:", err1)
			err2 := verifyPM.Verify(block.Header())
			fmt.Println("Verify:", err2)
		}

		log.Info("Mined new block", "candidate", block.Coinbase(), "number", block.Number(), "hash", block.Hash().String(), "time", block.Time().Int64(), "txs", len(block.Txs), "gas", block.GasUsed(), "elapsed", common.PrettyDuration(time.Since(bstart)))
		return
	}
	log.Warn("failed to mint block", "timestamp", header.Time, "err", err)
}

func (worker *Worker) stop() {
	close(worker.quit)
	worker.wg.Wait()
}
func (worker *Worker) setDelayDuration(delayMS uint64) error {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.delayMs = time.Duration(delayMS) * time.Millisecond
	return nil
}

func (worker *Worker) setCoinbase(name string, privateKey *ecdsa.PrivateKey) {
	// _, _ := worker.StateAt(worker.CurrentHeader().Root)
	// 	manger := plugin.NewPM(state)
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.coinbase = name
	worker.priKey = privateKey
}

func (worker *Worker) setExtra(extra []byte) {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.extra = extra
}

func (worker *Worker) commitNewWork(pm plugin.IPM, state *state.StateDB, parent *types.Header) (*types.Block, error) {
	start := time.Now()
	// must not use state.Copy() !
	sealState, err := worker.StateAt(parent.Root)
	if err != nil {
		return nil, err
	}
	header := &types.Header{Coinbase: worker.coinbase}
	if err := pm.Prepare(header); err != nil {
		return nil, err
	}
	work := &Work{
		currentHeader:   header,
		currentState:    state,
		currentTxs:      []*types.Transaction{},
		currentReceipts: []*types.Receipt{},
		currentGasPool:  new(common.GasPool).AddGas(header.GasLimit),
		currentCnt:      0,
		quit:            worker.quitWork,
	}

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
	if err := worker.commitTransactions(work, txs, BlockInterval); err != nil {
		return nil, err
	}

	blk, err := pm.Finalize(work.currentHeader, work.currentTxs, work.currentReceipts)
	if err != nil {
		return nil, fmt.Errorf("finalize block, err: %v", err)
	}

	work.currentBlock = blk

	sealPm := plugin.NewPM(sealState)
	sealPm.Init(0, parent)
	block, err := sealPm.Seal(work.currentBlock, worker.priKey, sealPm)
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

	// wait send
	if worker.delayMs > 0 {
		<-time.NewTimer(worker.delayMs).C
	}
	event.SendEvent(&event.Event{Typecode: event.NewMinedEv, Data: block})
	return block, nil
}

func (worker *Worker) commitTransactions(work *Work, txs *types.TransactionsByPriceAndNonce, interval uint64) error {
	var coalescedLogs []*types.Log
	endTimeStamp := work.currentHeader.Time*uint64(time.Second) + interval - 2*interval/5
	endTime := time.Unix((int64)(endTimeStamp)/(int64)(time.Second), (int64)(endTimeStamp)%(int64)(time.Second))

	for {
		select {
		case <-worker.quit:
			return errors.New("worker quit")
		case <-work.quit:
			return errors.New("work canceld")
		default:
		}
		if work.currentGasPool.Gas() < params.GasTableInstance.ActionGas {
			log.Debug("Not enough gas for further transactions", "have", work.currentGasPool, "want", params.GasTableInstance.ActionGas)
			break
		}

		if interval != math.MaxUint64 && uint64(time.Now().UnixNano()) >= endTimeStamp {
			log.Debug("Not enough time for further transactions", "timestamp", work.currentHeader.Time)
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
	quit            chan struct{}
}
