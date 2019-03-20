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

package txpool

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

var (
	evictionInterval    = time.Minute     // Time interval to check for evictable transactions
	statsReportInterval = 8 * time.Second // Time interval to report transaction pool stats
)

const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
)

// TxStatus is the current status of a transaction as seen by the tp.
type TxStatus uint

const (
	TxStatusUnknown TxStatus = iota
	TxStatusQueued
	TxStatusPending
	TxStatusIncluded
)

// blockChain provides the state of blockchain and current gas limit to do
// some pre checks in tx pool and feed subscribers.
type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
}

// TxPool contains all currently known transactions.
type TxPool struct {
	config                Config
	gasPrice              *big.Int
	chain                 blockChain
	signer                types.Signer
	chainHeadCh           chan *event.Event
	chainHeadSub          event.Subscription
	curAccountManager     *am.AccountManager
	pendingAccountManager *am.AccountManager
	currentMaxGas         uint64      // Current gas limit for transaction caps
	locals                *accountSet // Set of local transaction to exempt from eviction rules
	journal               *txJournal  // Journal of local transaction to back up to disk
	pending               map[common.Name]*txList
	queue                 map[common.Name]*txList
	beats                 map[common.Name]time.Time // Last heartbeat from each known account
	all                   *txLookup                 // All transactions to allow lookups
	priced                *txPricedList

	mu sync.RWMutex
	wg sync.WaitGroup // for shutdown sync
}

// New creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func New(config Config, chainconfig *params.ChainConfig, bc blockChain) *TxPool {
	signer := types.NewSigner(chainconfig.ChainID)
	all := newTxLookup()
	tp := &TxPool{
		config:      config.check(),
		chain:       bc,
		signer:      signer,
		locals:      newAccountSet(signer),
		chainHeadCh: make(chan *event.Event, chainHeadChanSize),
		pending:     make(map[common.Name]*txList),
		queue:       make(map[common.Name]*txList),
		beats:       make(map[common.Name]time.Time),
		all:         all,
		priced:      newTxPricedList(all),
		gasPrice:    new(big.Int).SetUint64(config.PriceLimit),
	}
	tp.reset(nil, bc.CurrentBlock().Header())

	// If local transactions and journaling is enabled, load from disk
	if !config.NoLocals && config.Journal != "" {
		tp.journal = newTxJournal(config.Journal)
		if err := tp.journal.load(tp.AddLocals); err != nil {
			log.Warn("Failed to load transaction journal", "err", err)
		}
		if err := tp.journal.rotate(tp.local()); err != nil {
			log.Warn("Failed to rotate transaction journal", "err", err)
		}
	}

	// Subscribe feeds from blockchain
	tp.chainHeadSub = event.Subscribe(nil, tp.chainHeadCh, event.ChainHeadEv, &types.Block{})

	NewTxpoolStation(tp)
	// Start the feed loop and return
	tp.wg.Add(1)
	go tp.loop()
	return tp
}

// loop is the transaction pool's main feed loop, waiting for and reacting to
// outside blockchain feeds as well as for various reporting and transaction
// eviction feeds.
func (tp *TxPool) loop() {
	defer tp.wg.Done()

	// Start the stats reporting and transaction eviction tickers
	var prevPending, prevQueued, prevStales int

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()

	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()

	journal := time.NewTicker(tp.config.Rejournal)
	defer journal.Stop()

	// Track the previous head headers for transaction reorgs
	head := tp.chain.CurrentBlock()

	// Keep waiting for and reacting to the various feeds
	for {
		select {
		// Handle ChainHeadfeed
		case ev := <-tp.chainHeadCh:
			block := ev.Data.(*types.Block)
			if block != nil {
				tp.mu.Lock()
				tp.reset(head.Header(), block.Header())
				head = block
				tp.mu.Unlock()
			}
			// Be unsubscribed due to system stopped
		case <-tp.chainHeadSub.Err():
			return
			// Handle stats reporting ticks
		case <-report.C:
			tp.mu.RLock()
			pending, queued := tp.stats()
			stales := tp.priced.stales
			tp.mu.RUnlock()

			if pending != prevPending || queued != prevQueued || stales != prevStales {
				log.Debug("Transaction pool status report", "executable", pending, "queued", queued, "stales", stales)
				prevPending, prevQueued, prevStales = pending, queued, stales
			}

			// Handle inactive account transaction eviction
		case <-evict.C:
			tp.mu.Lock()
			for addr := range tp.queue {
				// Skip local transactions from the eviction mechanism
				if tp.locals.contains(addr) {
					continue
				}
				// Any non-locals old enough should be removed
				if time.Since(tp.beats[addr]) > tp.config.Lifetime {
					for _, tx := range tp.queue[addr].Flatten() {
						tp.removeTx(tx.Hash(), true)
					}
				}
			}
			tp.mu.Unlock()

			// Handle local transaction journal rotation
		case <-journal.C:
			if tp.journal != nil {
				tp.mu.Lock()
				if err := tp.journal.rotate(tp.local()); err != nil {
					log.Warn("Failed to rotate local tx journal", "err", err)
				}
				tp.mu.Unlock()
			}
		}
	}
}

// lockedReset is a wrapper around reset to allow calling it in a thread safe
// manner. This method is only ever used in the tester!
func (tp *TxPool) lockedReset(oldHead, newHead *types.Header) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.reset(oldHead, newHead)
}

// reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
func (tp *TxPool) reset(oldHead, newHead *types.Header) {
	// If we're reorging an old state, reinject all dropped transactions
	var reinject []*types.Transaction

	if oldHead != nil && oldHead.Hash() != newHead.ParentHash {
		// If the reorg is too deep, avoid doing it (will happen during fast sync)
		oldNum := oldHead.Number.Uint64()
		newNum := newHead.Number.Uint64()

		if depth := uint64(math.Abs(float64(oldNum) - float64(newNum))); depth > 64 {
			log.Debug("Skipping deep transaction reorg", "depth", depth)
		} else {
			// Reorg seems shallow enough to pull in all transactions into memory
			var discarded, included []*types.Transaction

			var (
				rem = tp.chain.GetBlock(oldHead.Hash(), oldHead.Number.Uint64())
				add = tp.chain.GetBlock(newHead.Hash(), newHead.Number.Uint64())
			)
			for rem.NumberU64() > add.NumberU64() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = tp.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
			}
			for add.NumberU64() > rem.NumberU64() {
				included = append(included, add.Transactions()...)
				if add = tp.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			for rem.Hash() != add.Hash() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = tp.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
				included = append(included, add.Transactions()...)
				if add = tp.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			reinject = types.TxDifference(discarded, included)
		}
	}
	// Initialize the internal state to the current head
	if newHead == nil {
		newHead = tp.chain.CurrentBlock().Header() // Special case during testing
	}
	statedb, err := tp.chain.StateAt(newHead.Root)
	if err != nil {
		log.Error("Failed to reset txpool state", "err", err)
		return
	}
	tp.curAccountManager, err = am.NewAccountManager(statedb)
	if err != nil {
		log.Error("Failed to create current NewAccountManager", "err", err)
		return
	}
	tp.pendingAccountManager, err = am.NewAccountManager(statedb.Copy())
	if err != nil {
		log.Error("Failed to create pending  NewAccountManager state", "err", err)
		return
	}
	tp.currentMaxGas = newHead.GasLimit
	// Inject any transactions discarded due to reorgs
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	SenderCacher.recover(tp.signer, reinject)
	tp.addTxsLocked(reinject, false)

	// validate the pool of pending transactions, this will remove
	// any transactions that have been included in the block or
	// have been invalidated because of another transaction (e.g.
	// higher gas price)
	tp.demoteUnexecutables()

	// Update all accounts to the latest known pending nonce
	for name, list := range tp.pending {
		txs := list.Flatten() // Heavy but will be cached and is needed by the miner anyway
		// todo change transaction action nonce
		if err := tp.pendingAccountManager.SetNonce(name, txs[len(txs)-1].GetActions()[0].Nonce()+1); err != nil {

			if err != am.ErrAccountIsDestroy {
				log.Error("Failed to pendingAccountManager SetNonce", "err", err)
				return
			} else {
				delete(tp.pending, name)
				delete(tp.beats, name)
				delete(tp.queue, name)
				log.Debug("Remove all destory account ", "name", name)
			}

		}
	}
	// Check the queue and move transactions over to the pending if possible
	// or remove those that have become invalid
	tp.promoteExecutables(nil)
}

// Stop terminates the transaction tp.
func (tp *TxPool) Stop() {
	// Unsubscribe subscriptions registered from blockchain
	tp.chainHeadSub.Unsubscribe()
	tp.wg.Wait()

	if tp.journal != nil {
		tp.journal.close()
	}
	log.Info("Transaction pool stopped")
}

// GasPrice returns the current gas price enforced by the transaction tp.
func (tp *TxPool) GasPrice() *big.Int {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	return new(big.Int).Set(tp.gasPrice)
}

// SetGasPrice updates the minimum price required by the transaction pool for a
// new transaction, and drops all transactions below this threshold.
func (tp *TxPool) SetGasPrice(price *big.Int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.gasPrice = price
	for _, tx := range tp.priced.Cap(price, tp.locals) {
		tp.removeTx(tx.Hash(), false)
	}
	log.Info("Transaction pool price threshold updated", "price", price)
}

// State returns the virtual managed state of the transaction tp.
func (tp *TxPool) State() *am.AccountManager {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.pendingAccountManager
}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (tp *TxPool) Stats() (int, int) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.stats()
}

// stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (tp *TxPool) stats() (int, int) {
	pending := 0
	for _, list := range tp.pending {
		pending += list.Len()
	}
	queued := 0
	for _, list := range tp.queue {
		queued += list.Len()
	}
	return pending, queued
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (tp *TxPool) Content() (map[common.Name][]*types.Transaction, map[common.Name][]*types.Transaction) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	pending := make(map[common.Name][]*types.Transaction)
	for addr, list := range tp.pending {
		pending[addr] = list.Flatten()
	}
	queued := make(map[common.Name][]*types.Transaction)
	for addr, list := range tp.queue {
		queued[addr] = list.Flatten()
	}
	return pending, queued
}

// Pending retrieves all currently processable transactions, groupped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (tp *TxPool) Pending() (map[common.Name][]*types.Transaction, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	pending := make(map[common.Name][]*types.Transaction)
	for addr, list := range tp.pending {
		pending[addr] = list.Flatten()
	}
	return pending, nil
}

// local retrieves all currently known local transactions, groupped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (tp *TxPool) local() map[common.Name][]*types.Transaction {
	txs := make(map[common.Name][]*types.Transaction)
	for addr := range tp.locals.accounts {
		if list := tp.pending[addr]; list != nil {
			txs[addr] = append(txs[addr], list.Flatten()...)
		}
		if list := tp.queue[addr]; list != nil {
			txs[addr] = append(txs[addr], list.Flatten()...)
		}
	}
	return txs
}

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (price and size).
func (tp *TxPool) validateTx(tx *types.Transaction, local bool) error {
	validateAction := func(tx *types.Transaction, action *types.Action) error {
		from := action.Sender()

		// Drop non-local transactions under our own minimal accepted gas price
		local = local || tp.locals.contains(from) // account may be local even if the transaction arrived from the network
		if !local && tp.gasPrice.Cmp(tx.GasPrice()) > 0 {
			return ErrUnderpriced
		}
		// Ensure the transaction adheres to nonce ordering
		nonce, err := tp.curAccountManager.GetNonce(from)
		if err != nil {
			return err
		}
		// todo change action nonce
		if nonce > action.Nonce() {
			return ErrNonceTooLow
		}

		// Transactor should have enough funds to cover the gas costs
		balance, err := tp.curAccountManager.GetAccountBalanceByID(from, tx.GasAssetID())
		if err != nil {
			return err
		}

		gascost := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(action.Gas()))
		if balance.Cmp(gascost) < 0 {
			return ErrInsufficientFundsForGas
		}

		// Transactor should have enough funds to cover the value costs
		balance, err = tp.curAccountManager.GetAccountBalanceByID(from, action.AssetID())
		if err != nil {
			return err
		}

		value := action.Value()
		if tx.GasAssetID() == action.AssetID() {
			value.Add(value, gascost)
		}

		if balance.Cmp(value) < 0 {
			return ErrInsufficientFundsForValue
		}

		//
		if action.CheckValue() != true {
			return ErrInvalidValue
		}

		intrGas, err := IntrinsicGas(action)
		if err != nil {
			return err
		}

		if action.Gas() < intrGas {
			return ErrIntrinsicGas
		}
		return nil

	}

	// Heuristic limit, reject transactions over 32KB to prfeed DOS attacks
	if tx.Size() > 32*1024 {
		return ErrOversizedData
	}

	// Make sure the transaction is signed properly
	if err := tp.curAccountManager.RecoverTx(tp.signer, tx); err != nil {
		log.Error("account Manager reocver faild ", "err", err)
		return ErrInvalidSender
	}

	// Transaction action  value can't be negative.
	var allgas uint64
	for _, a := range tx.GetActions() {
		if a.Value().Sign() < 0 {
			return ErrNegativeValue
		}
		if err := validateAction(tx, a); err != nil {
			return err
		}
		allgas += a.Gas()
	}

	// Ensure the transaction doesn't exceed the current block limit gas.
	if tp.currentMaxGas < allgas {
		return ErrGasLimit
	}

	return nil
}

func (tp *TxPool) add(tx *types.Transaction, local bool) (bool, error) {
	// If the transaction is already known, discard it
	hash := tx.Hash()
	if tp.all.Get(hash) != nil {
		log.Trace("Discarding already known transaction", "hash", hash)
		return false, fmt.Errorf("known transaction: %x", hash)
	}
	// If the transaction fails basic validation, discard it
	if err := tp.validateTx(tx, local); err != nil {
		log.Trace("Discarding invalid transaction", "hash", hash, "err", err)
		return false, err
	}
	// If the transaction pool is full, discard underpriced transactions
	if uint64(tp.all.Count()) >= tp.config.GlobalSlots+tp.config.GlobalQueue {
		// If the new transaction is underpriced, don't accept it
		if !local && tp.priced.Underpriced(tx, tp.locals) {
			log.Trace("Discarding underpriced transaction", "hash", hash, "price", tx.GasPrice())
			return false, ErrUnderpriced
		}
		// New transaction is better than our worse ones, make room for it
		drop := tp.priced.Discard(tp.all.Count()-int(tp.config.GlobalSlots+tp.config.GlobalQueue-1), tp.locals)
		for _, tx := range drop {
			log.Trace("Discarding freshly underpriced transaction", "hash", tx.Hash(), "price", tx.GasPrice())
			tp.removeTx(tx.Hash(), false)
		}
	}

	// If the transaction is replacing an already pending one, do directly
	// todo Change action
	from := tx.GetActions()[0].Sender()
	if list := tp.pending[from]; list != nil && list.Overlaps(tx) {
		// Nonce already pending, check if required price bump is met
		inserted, old := list.Add(tx, tp.config.PriceBump)
		if !inserted {
			return false, ErrReplaceUnderpriced
		}
		// New transaction is better, replace old one
		if old != nil {
			tp.all.Remove(old.Hash())
			tp.priced.Removed()
		}
		tp.all.Add(tx)
		tp.priced.Put(tx)
		tp.journalTx(from, tx)

		log.Trace("Pooled new executable transaction", "hash", hash, "from", from)

		// We've directly injected a replacement transaction, notify subsystems
		events := []*event.Event{
			{Typecode: event.TxEv, Data: []*types.Transaction{tx}},
			{To: event.GetStationByName("broadcast"), Typecode: event.P2PTxMsg, Data: []*types.Transaction{tx}},
		}
		go event.SendEvents(events)

		return old != nil, nil
	}
	// New transaction isn't replacing a pending one, push into queue
	replace, err := tp.enqueueTx(hash, tx)
	if err != nil {
		return false, err
	}
	// Mark local addresses and journal local transactions
	if local {
		tp.locals.add(from)
	}
	tp.journalTx(from, tx)

	log.Trace("Pooled new future transaction", "hash", hash, "from", from)
	return replace, nil
}

// enqueueTx inserts a new transaction into the non-executable transaction queue.
//
// Note, this method assumes the pool lock is held!
func (tp *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction) (bool, error) {
	// Try to insert the transaction into the future queue
	from := tx.GetActions()[0].Sender()

	if tp.queue[from] == nil {
		tp.queue[from] = newTxList(false)
	}
	inserted, old := tp.queue[from].Add(tx, tp.config.PriceBump)
	if !inserted {
		// An older transaction was better, discard this
		return false, ErrReplaceUnderpriced
	}
	// Discard any previous transaction and mark this
	if old != nil {
		tp.all.Remove(old.Hash())
		tp.priced.Removed()
	}
	if tp.all.Get(hash) == nil {
		tp.all.Add(tx)
		tp.priced.Put(tx)
	}
	return old != nil, nil
}

// journalTx adds the specified transaction to the local disk journal if it is
// deemed to have been sent from a local account.
func (tp *TxPool) journalTx(from common.Name, tx *types.Transaction) {
	// Only journal if it's enabled and the transaction is local
	if tp.journal == nil || !tp.locals.contains(from) {
		return
	}
	if err := tp.journal.insert(tx); err != nil {
		log.Warn("Failed to journal local transaction", "err", err)
	}
}

// promoteTx adds a transaction to the pending (processable) list of transactions
// and returns whether it was inserted or an older was better.
//
// Note, this method assumes the pool lock is held!
func (tp *TxPool) promoteTx(name common.Name, hash common.Hash, tx *types.Transaction) bool {
	// Try to insert the transaction into the pending queue
	if tp.pending[name] == nil {
		tp.pending[name] = newTxList(true)
	}
	list := tp.pending[name]

	inserted, old := list.Add(tx, tp.config.PriceBump)
	if !inserted {
		// An older transaction was better, discard this
		tp.all.Remove(hash)
		tp.priced.Removed()

		return false
	}
	// Otherwise discard any previous transaction and mark this
	if old != nil {
		tp.all.Remove(old.Hash())
		tp.priced.Removed()

	}
	// Failsafe to work around direct pending inserts (tests)
	if tp.all.Get(hash) == nil {
		tp.all.Add(tx)
		tp.priced.Put(tx)
	}
	// Set the potentially new pending nonce and notify any subsystems of the new tx
	tp.beats[name] = time.Now()

	// todo action
	tp.pendingAccountManager.SetNonce(name, tx.GetActions()[0].Nonce()+1)
	return true
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one in the mean time, ensuring it goes around the local
// pricing constraints.
func (tp *TxPool) AddLocal(tx *types.Transaction) error {
	return tp.addTx(tx, !tp.config.NoLocals)
}

// AddRemote enqueues a single transaction into the pool if it is valid. If the
// sender is not among the locally tracked ones, full pricing constraints will
// apply.
func (tp *TxPool) AddRemote(tx *types.Transaction) error {
	return tp.addTx(tx, false)
}

// AddLocals enqueues a batch of transactions into the pool if they are valid,
// marking the senders as a local ones in the mean time, ensuring they go around
// the local pricing constraints.
func (tp *TxPool) AddLocals(txs []*types.Transaction) []error {
	return tp.addTxs(txs, !tp.config.NoLocals)
}

// AddRemotes enqueues a batch of transactions into the pool if they are valid.
// If the senders are not among the locally tracked ones, full pricing constraints
// will apply.
func (tp *TxPool) AddRemotes(txs []*types.Transaction) []error {
	return tp.addTxs(txs, false)
}

// addTx enqueues a single transaction into the pool if it is valid.
func (tp *TxPool) addTx(tx *types.Transaction, local bool) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Try to inject the transaction and update any state
	replace, err := tp.add(tx, local)
	if err != nil {
		return err
	}
	// If we added a new transaction, run promotion checks and return
	if !replace {
		// todo
		from := tx.GetActions()[0].Sender()
		tp.promoteExecutables([]common.Name{from})
	}
	return nil
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (tp *TxPool) addTxs(txs []*types.Transaction, local bool) []error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	return tp.addTxsLocked(txs, local)
}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (tp *TxPool) addTxsLocked(txs []*types.Transaction, local bool) []error {
	// Add the batch of transaction, tracking the accepted ones
	dirty := make(map[common.Name]struct{})
	errs := make([]error, len(txs))

	for i, tx := range txs {
		var replace bool
		if replace, errs[i] = tp.add(tx, local); errs[i] == nil && !replace {
			from := tx.GetActions()[0].Sender()
			dirty[from] = struct{}{}
		}
	}
	// Only reprocess the internal state if something was actually added
	if len(dirty) > 0 {
		addrs := make([]common.Name, 0, len(dirty))
		for addr := range dirty {
			addrs = append(addrs, addr)
		}
		tp.promoteExecutables(addrs)
	}
	return errs
}

// Status returns the status (unknown/pending/queued) of a batch of transactions
// identified by their hashes.
func (tp *TxPool) Status(hashes []common.Hash) []TxStatus {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	status := make([]TxStatus, len(hashes))
	for i, hash := range hashes {
		if tx := tp.all.Get(hash); tx != nil {
			from := tx.GetActions()[0].Sender()
			nonce := tx.GetActions()[0].Nonce()
			if tp.pending[from] != nil && tp.pending[from].txs.items[nonce] != nil {
				status[i] = TxStatusPending
			} else {
				status[i] = TxStatusQueued
			}
		}
	}
	return status
}

// Get returns a transaction if it is contained in the pool
// and nil otherwise.
func (tp *TxPool) Get(hash common.Hash) *types.Transaction {
	return tp.all.Get(hash)
}

// removeTx removes a single transaction from the queue, moving all subsequent
// transactions back to the future queue.
func (tp *TxPool) removeTx(hash common.Hash, outofbound bool) {
	// Fetch the transaction we wish to delete
	tx := tp.all.Get(hash)
	if tx == nil {
		return
	}
	from := tx.GetActions()[0].Sender()

	// Remove it from the list of known transactions
	tp.all.Remove(hash)
	if outofbound {
		tp.priced.Removed()
	}
	// Remove the transaction from the pending lists and reset the account nonce
	if pending := tp.pending[from]; pending != nil {
		if removed, invalids := pending.Remove(tx); removed {
			// If no more pending transactions are left, remove the list
			if pending.Empty() {
				delete(tp.pending, from)
				delete(tp.beats, from)
			}
			// Postpone any invalidated transactions
			for _, tx := range invalids {
				tp.enqueueTx(tx.Hash(), tx)
			}

			nonce := tx.GetActions()[0].Nonce()

			// Update the account nonce if needed
			pnonce, err := tp.pendingAccountManager.GetNonce(from)
			if err != nil && err != am.ErrAccountNotExist {
				log.Error("removeTx pending account manager get nonce err ", "name", from, "err", err)
			}
			if pnonce > nonce {
				if err := tp.pendingAccountManager.SetNonce(from, nonce); err != nil {
					log.Error("removeTx pending account manager set nonce err ", "name", from, "err", err)
				}
			}
			return
		}
	}
	// Transaction is in the future queue
	if future := tp.queue[from]; future != nil {
		future.Remove(tx)
		if future.Empty() {
			delete(tp.queue, from)
		}
	}
}

// promoteExecutables moves transactions that have become processable from the
// future queue to the set of pending transactions. During this process, all
// invalidated transactions (low nonce, low balance) are deleted.
func (tp *TxPool) promoteExecutables(accounts []common.Name) {
	// Track the promoted transactions to broadcast them at once
	var promoted []*types.Transaction

	// Gather all the accounts potentially needing updates
	if accounts == nil {
		accounts = make([]common.Name, 0, len(tp.queue))
		for addr := range tp.queue {
			accounts = append(accounts, addr)
		}
	}
	// Iterate over all accounts and promote any executable transactions
	for _, addr := range accounts {
		list := tp.queue[addr]
		if list == nil {
			continue // Just in case someone calls with a non existing account
		}
		// Drop all transactions that are deemed too old (low nonce)
		nonce, err := tp.curAccountManager.GetNonce(addr)
		if err != nil {
			log.Error("promoteExecutables current account manager get nonce err", "name", addr, "err", err)
		}
		for _, tx := range list.Forward(nonce) {
			hash := tx.Hash()
			log.Trace("Removed old queued transaction", "hash", hash)
			tp.all.Remove(hash)
			tp.priced.Removed()
		}
		// Drop all transactions that are too costly (low balance or out of gas)
		// todo assetID
		balance, err := tp.curAccountManager.GetAccountBalanceByID(addr, tp.config.GasAssetID)
		if err != nil {
			log.Error("promoteExecutables current account manager get balance err ", "name", addr, "assetID", tp.config.GasAssetID, "err", err)
		}
		drops, _ := list.Filter(balance, tp.currentMaxGas, tp.curAccountManager.GetAccountBalanceByID)
		for _, tx := range drops {
			hash := tx.Hash()
			log.Trace("Removed unpayable queued transaction", "hash", hash)
			tp.all.Remove(hash)
			tp.priced.Removed()
		}

		// Gather all executable transactions and promote them
		nonce, err = tp.pendingAccountManager.GetNonce(addr)
		if err != nil && err != am.ErrAccountNotExist {
			log.Error("promoteExecutables pending account manager get nonce err ", "name", addr, "err", err)
		}
		for _, tx := range list.Ready(nonce) {
			hash := tx.Hash()
			if tp.promoteTx(addr, hash, tx) {
				log.Trace("Promoting queued transaction", "hash", hash)
				promoted = append(promoted, tx)
			}
		}
		// Drop all transactions over the allowed limit
		if !tp.locals.contains(addr) {
			for _, tx := range list.Cap(int(tp.config.AccountQueue)) {
				hash := tx.Hash()
				tp.all.Remove(hash)
				tp.priced.Removed()
				log.Trace("Removed cap-exceeding queued transaction", "hash", hash)
			}
		}
		// Delete the entire queue entry if it became empty.
		if list.Empty() {
			delete(tp.queue, addr)
		}
	}
	// Notify subsystem for new promoted transactions.
	if len(promoted) > 0 {
		// go event.SendEvent(&event.Event{Typecode: event.TxEv, Data: promoted})
		events := []*event.Event{
			{Typecode: event.TxEv, Data: promoted},
			{To: event.GetStationByName("broadcast"), Typecode: event.P2PTxMsg, Data: promoted},
		}
		go event.SendEvents(events)
	}
	// If the pending limit is overflown, start equalizing allowances
	pending := uint64(0)
	for _, list := range tp.pending {
		pending += uint64(list.Len())
	}
	if pending > tp.config.GlobalSlots {
		// Assemble a spam order to penalize large transactors first
		spammers := prque.New()
		for addr, list := range tp.pending {
			// Only evict transactions from high rollers
			if !tp.locals.contains(addr) && uint64(list.Len()) > tp.config.AccountSlots {
				spammers.Push(addr, float32(list.Len()))
			}
		}
		// Gradually drop transactions from offenders
		offenders := []common.Name{}
		for pending > tp.config.GlobalSlots && !spammers.Empty() {
			// Retrieve the next offender if not local address
			offender, _ := spammers.Pop()
			offenders = append(offenders, offender.(common.Name))

			// Equalize balances until all the same or below threshold
			if len(offenders) > 1 {
				// Calculate the equalization threshold for all current offenders
				threshold := tp.pending[offender.(common.Name)].Len()

				// Iteratively reduce all offenders until below limit or threshold reached
				for pending > tp.config.GlobalSlots && tp.pending[offenders[len(offenders)-2]].Len() > threshold {
					for i := 0; i < len(offenders)-1; i++ {
						list := tp.pending[offenders[i]]
						for _, tx := range list.Cap(list.Len() - 1) {
							// Drop the transaction from the global pools too
							hash := tx.Hash()
							tp.all.Remove(hash)
							tp.priced.Removed()

							// Update the account nonce to the dropped transaction
							// todo change action
							pnonce, err := tp.pendingAccountManager.GetNonce(offenders[i])
							if err != nil && err != am.ErrAccountNotExist {
								log.Error("promoteExecutables pending account manager get nonce err ", "name", offenders[i], "err", err)
							}
							if nonce := tx.GetActions()[0].Nonce(); pnonce > nonce {
								if err := tp.pendingAccountManager.SetNonce(offenders[i], nonce); err != nil {
									log.Error("promoteExecutables pending account manager set nonce err ", "name", offenders[i], "nonce", nonce, "err", err)
								}
							}
							log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
						}
						pending--
					}
				}
			}
		}
		// If still above threshold, reduce to limit or min allowance
		if pending > tp.config.GlobalSlots && len(offenders) > 0 {
			for pending > tp.config.GlobalSlots && uint64(tp.pending[offenders[len(offenders)-1]].Len()) > tp.config.AccountSlots {
				for _, addr := range offenders {
					list := tp.pending[addr]
					for _, tx := range list.Cap(list.Len() - 1) {
						// Drop the transaction from the global pools too
						hash := tx.Hash()
						tp.all.Remove(hash)
						tp.priced.Removed()

						// Update the account nonce to the dropped transaction
						pnonce, err := tp.pendingAccountManager.GetNonce(addr)
						if err != nil && err != am.ErrAccountNotExist {
							log.Error("promoteExecutables pending account manager get nonce err ", "name", addr, "err", err)
						}
						if nonce := tx.GetActions()[0].Nonce(); pnonce > nonce {
							tp.pendingAccountManager.SetNonce(addr, nonce)
						}
						log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
					}
					pending--
				}
			}
		}
	}
	// If we've queued more transactions than the hard limit, drop oldest ones
	queued := uint64(0)
	for _, list := range tp.queue {
		queued += uint64(list.Len())
	}
	if queued > tp.config.GlobalQueue {
		// Sort all accounts with queued transactions by heartbeat
		addresses := make(namesByHeartbeat, 0, len(tp.queue))
		for addr := range tp.queue {
			if !tp.locals.contains(addr) { // don't drop locals
				addresses = append(addresses, nameByHeartbeat{addr, tp.beats[addr]})
			}
		}
		sort.Sort(addresses)

		// Drop transactions until the total is below the limit or only locals remain
		for drop := queued - tp.config.GlobalQueue; drop > 0 && len(addresses) > 0; {
			addr := addresses[len(addresses)-1]
			list := tp.queue[addr.name]

			addresses = addresses[:len(addresses)-1]

			// Drop all transactions if they are less than the overflow
			if size := uint64(list.Len()); size <= drop {
				for _, tx := range list.Flatten() {
					tp.removeTx(tx.Hash(), true)
				}
				drop -= size
				continue
			}
			// Otherwise drop only last few transactions
			txs := list.Flatten()
			for i := len(txs) - 1; i >= 0 && drop > 0; i-- {
				tp.removeTx(txs[i].Hash(), true)
				drop--
			}
		}
	}
}

// demoteUnexecutables removes invalid and processed transactions from the pools
// executable/pending queue and any subsequent transactions that become unexecutable
// are moved back into the future queue.
func (tp *TxPool) demoteUnexecutables() {
	// Iterate over all accounts and demote any non-executable transactions
	for addr, list := range tp.pending {
		nonce, err := tp.curAccountManager.GetNonce(addr)
		if err != nil && err != am.ErrAccountNotExist {
			log.Error("promoteExecutables current account manager get nonce err ", "name", addr, "err", err)
		}

		// Drop all transactions that are deemed too old (low nonce)
		for _, tx := range list.Forward(nonce) {
			hash := tx.Hash()
			log.Trace("Removed old pending transaction", "hash", hash)
			tp.all.Remove(hash)
			tp.priced.Removed()
		}

		// Drop all transactions that are too costly (low balance or out of gas), and queue any invalids back for later
		gasBalance, err := tp.curAccountManager.GetAccountBalanceByID(addr, tp.config.GasAssetID)
		if err != nil && err != am.ErrAccountNotExist {
			log.Error("promoteExecutables current account manager get balance err ", "name", addr, "assetID", tp.config.GasAssetID, "err", err)
		}

		drops, invalids := list.Filter(gasBalance, tp.currentMaxGas, tp.curAccountManager.GetAccountBalanceByID)
		for _, tx := range drops {
			hash := tx.Hash()
			log.Trace("Removed unpayable pending transaction", "hash", hash)
			tp.all.Remove(hash)
			tp.priced.Removed()
		}

		for _, tx := range invalids {
			hash := tx.Hash()
			log.Trace("Demoting pending transaction", "hash", hash)
			tp.enqueueTx(hash, tx)
		}
		// If there's a gap in front, alert (should never happen) and postpone all transactions
		if list.Len() > 0 && list.txs.Get(nonce) == nil {
			for _, tx := range list.Cap(0) {
				hash := tx.Hash()
				log.Error("Demoting invalidated transaction", "hash", hash)
				tp.enqueueTx(hash, tx)
			}
		}
		// Delete the entire queue entry if it became empty.
		if list.Empty() {
			delete(tp.pending, addr)
			delete(tp.beats, addr)
		}
	}
}
