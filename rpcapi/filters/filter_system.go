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

package filters

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/event"
	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
)

// Type determines the kind of filter and is used to put the filter in to
// the correct bucket when added.
type Type byte

const (
	// UnknownSubscription indicates an unknown subscription type
	UnknownSubscription Type = iota
	// LogsSubscription queries for new or removed (chain reorg) logs
	LogsSubscription
	// transactions entering the pending state
	PendingTransactionsSubscription
	// BlocksSubscription queries hashes for blocks that are imported
	BlocksSubscription
	// LastSubscription keeps track of the last index
	LastIndexSubscription
)

const (

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
	// rmLogsChanSize is the size of channel listening to RemovedLogsEvent.
	rmLogsChanSize = 10
	// logsChanSize is the size of channel listening to LogsEvent.
	logsChanSize = 10
	// chainEvChanSize is the size of channel listening to ChainEvent.
	chainEvChanSize = 10
)

var (
	ErrInvalidSubscriptionID = errors.New("invalid id")
)

type subscription struct {
	id        rpc.ID
	typ       Type
	created   time.Time
	logsCrit  FilterQuery
	logs      chan []*types.Log
	hashes    chan []common.Hash
	headers   chan *types.Header
	installed chan struct{} // closed when the filter is installed
	err       chan error    // closed when the filter is uninstalled
}

// EventSystem creates subscriptions, processes events and broadcasts them to the
// subscription which match the subscription criteria.
type EventSystem struct {
	backend  Backend
	lastHead *types.Header

	// Subscriptions
	txsSub   event.Subscription // Subscription for new transaction event
	chainSub event.Subscription // Subscription for new chain event

	// Channels
	install   chan *subscription // install filter for event notification
	uninstall chan *subscription // remove filter for event notification
	txsCh     chan *router.Event // Channel to receive new transactions event
	chainCh   chan *router.Event // Channel to receive new chain event
}

// NewEventSystem creates a new manager that listens for event on the given mux,
// parses and filters them. It uses the all map to retrieve filter changes. The
// work loop holds its own index that is used to forward events to filters.
//
// The returned manager has a loop that needs to be stopped with the Stop function
// or by stopping the given mux.
func NewEventSystem(backend Backend) *EventSystem {
	m := &EventSystem{
		backend:   backend,
		install:   make(chan *subscription),
		uninstall: make(chan *subscription),
		txsCh:     make(chan *router.Event, txChanSize),
		chainCh:   make(chan *router.Event, chainEvChanSize),
	}

	// Subscribe events
	m.txsSub = router.Subscribe(nil, m.txsCh, router.NewTxs, []*types.Transaction{})
	m.chainSub = router.Subscribe(nil, m.chainCh, router.ChainHeadEv, &types.Block{})

	go m.eventLoop()
	return m
}

// Subscription is created when the client registers itself for a particular event.
type Subscription struct {
	ID        rpc.ID
	f         *subscription
	es        *EventSystem
	unsubOnce sync.Once
}

// Err returns a channel that is closed when unsubscribed.
func (sub *Subscription) Err() <-chan error {
	return sub.f.err
}

// Unsubscribe uninstalls the subscription from the event broadcast loop.
func (sub *Subscription) Unsubscribe() {
	sub.unsubOnce.Do(func() {
	uninstallLoop:
		for {
			// write uninstall request and consume logs/hashes. This prevents
			// the eventLoop broadcast method to deadlock when writing to the
			// filter event channel while the subscription loop is waiting for
			// this method to return (and thus not reading these events).
			select {
			case sub.es.uninstall <- sub.f:
				break uninstallLoop
			case <-sub.f.logs:
			case <-sub.f.hashes:
			case <-sub.f.headers:
			}
		}

		// wait for filter to be uninstalled in work loop before returning
		// this ensures that the manager won't use the event channel which
		// will probably be closed by the client asap after this method returns.
		<-sub.Err()
	})
}

// subscribe installs the subscription in the event broadcast loop.
func (es *EventSystem) subscribe(sub *subscription) *Subscription {
	es.install <- sub
	<-sub.installed
	return &Subscription{ID: sub.id, f: sub, es: es}
}

// SubscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel. Default value for the from and to
// block is "latest". If the fromBlock > toBlock an error is returned.
func (es *EventSystem) SubscribeLogs(crit FilterQuery, logs chan []*types.Log) (*Subscription, error) {
	return es.subscribeLogs(crit, logs), nil
}

// subscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel.
func (es *EventSystem) subscribeLogs(crit FilterQuery, logs chan []*types.Log) *Subscription {
	sub := &subscription{
		id:        rpc.NewID(),
		typ:       LogsSubscription,
		logsCrit:  crit,
		created:   time.Now(),
		logs:      logs,
		hashes:    make(chan []common.Hash),
		headers:   make(chan *types.Header),
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}

// SubscribeNewHeads creates a subscription that writes the header of a block that is
// imported in the chain.
func (es *EventSystem) SubscribeNewHeads(headers chan *types.Header) *Subscription {
	sub := &subscription{
		id:        rpc.NewID(),
		typ:       BlocksSubscription,
		created:   time.Now(),
		logs:      make(chan []*types.Log),
		hashes:    make(chan []common.Hash),
		headers:   headers,
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}

// SubscribePendingTxs creates a subscription that writes transaction hashes for
// transactions that enter the transaction pool.
func (es *EventSystem) SubscribePendingTxs(hashes chan []common.Hash) *Subscription {
	sub := &subscription{
		id:        rpc.NewID(),
		typ:       PendingTransactionsSubscription,
		created:   time.Now(),
		logs:      make(chan []*types.Log),
		hashes:    hashes,
		headers:   make(chan *types.Header),
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}

type filterIndex map[Type]map[rpc.ID]*subscription

// broadcast event to filters that match criteria.
func (es *EventSystem) broadcast(filters filterIndex, ev *router.Event) {
	if ev == nil {
		return
	}
	switch ev.Typecode {
	case router.NewTxs:
		txs := ev.Data.([]*types.Transaction)
		hashes := make([]common.Hash, 0, len(txs))
		for _, tx := range txs {
			hashes = append(hashes, tx.Hash())
		}
		for _, f := range filters[PendingTransactionsSubscription] {
			f.hashes <- hashes
		}
	case router.ChainHeadEv:
		block := ev.Data.(*types.Block)
		for _, f := range filters[BlocksSubscription] {
			f.headers <- block.Header()
		}
		if len(filters[LogsSubscription]) > 0 {
			es.lightFilterNewHead(block.Header(), func(header *types.Header, remove bool) {
				for _, f := range filters[LogsSubscription] {
					if matchedLogs := es.lightFilterLogs(header, f.logsCrit.Accounts, f.logsCrit.Topics, remove); len(matchedLogs) > 0 {
						f.logs <- matchedLogs
					}
				}
			})
		}
	}
}

func (es *EventSystem) lightFilterNewHead(newHeader *types.Header, callBack func(*types.Header, bool)) {
	oldh := es.lastHead
	es.lastHead = newHeader
	if oldh == nil {
		return
	}
	newh := newHeader
	// find common ancestor, create list of rolled back and new block hashes
	var oldHeaders, newHeaders []*types.Header
	for oldh.Hash() != newh.Hash() {
		if oldh.Number.Uint64() >= newh.Number.Uint64() {
			oldHeaders = append(oldHeaders, oldh)
			oldh = rawdb.ReadHeader(es.backend.ChainDb(), oldh.ParentHash, oldh.Number.Uint64()-1)
		}
		if oldh.Number.Uint64() < newh.Number.Uint64() {
			newHeaders = append(newHeaders, newh)
			newh = rawdb.ReadHeader(es.backend.ChainDb(), newh.ParentHash, newh.Number.Uint64()-1)
			if newh == nil {
				// happens when CHT syncing, nothing to do
				newh = oldh
			}
		}
	}
	// roll back old blocks
	for _, h := range oldHeaders {
		callBack(h, true)
	}
	// check new blocks (array is in reverse order)
	for i := len(newHeaders) - 1; i >= 0; i-- {
		callBack(newHeaders[i], false)
	}
}

// filter logs of a single header in light client mode
func (es *EventSystem) lightFilterLogs(header *types.Header, accounts []common.Name, topics [][]common.Hash, remove bool) []*types.Log {
	if bloomFilter(header.Bloom, accounts, topics) {
		// Get the logs of the block
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		logsList, err := es.backend.GetLogs(ctx, header.Hash())
		if err != nil {
			return nil
		}
		var unfiltered []*types.Log
		for _, logs := range logsList {
			for _, log := range logs {
				logcopy := *log
				//				logcopy.Removed = remove
				unfiltered = append(unfiltered, &logcopy)
			}
		}
		logs := filterLogs(unfiltered, accounts, topics)
		if len(logs) > 0 && logs[0].TxHash == (common.Hash{}) {
			// We have matching but non-derived logs
			receipts, err := es.backend.GetReceipts(ctx, header.Hash())
			if err != nil {
				return nil
			}
			unfiltered = unfiltered[:0]
			for _, receipt := range receipts {
				for _, log := range receipt.Logs {
					logcopy := *log
					//					logcopy.Removed = remove
					unfiltered = append(unfiltered, &logcopy)
				}
			}
			logs = filterLogs(unfiltered, accounts, topics)
		}
		return logs
	}
	return nil
}

// eventLoop (un)installs filters and processes mux events.
func (es *EventSystem) eventLoop() {
	// Ensure all subscriptions get cleaned up
	defer func() {
		//		es.pendingLogSub.Unsubscribe()
		es.txsSub.Unsubscribe()
		//es.logsSub.Unsubscribe()
		//es.rmLogsSub.Unsubscribe()
		es.chainSub.Unsubscribe()
	}()

	index := make(filterIndex)
	for i := UnknownSubscription; i < LastIndexSubscription; i++ {
		index[i] = make(map[rpc.ID]*subscription)
	}

	for {
		select {
		// Handle subscribed events
		case ev := <-es.txsCh:
			es.broadcast(index, ev)
		case ev := <-es.chainCh:
			es.broadcast(index, ev)

		case f := <-es.install:
			index[f.typ][f.id] = f
			close(f.installed)
		case f := <-es.uninstall:
			delete(index[f.typ], f.id)
			close(f.err)
		// System stopped
		case <-es.txsSub.Err():
			return
		case <-es.chainSub.Err():
			return
		}
	}
}
