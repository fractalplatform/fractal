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
	"encoding/binary"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/types"
)

const (
	//	maxKonwnTxs      = 32
	//	txsSendDelay     = 50 * time.Millisecond
	//	txsSendThreshold = 32
	cacheBits = 12
	cacheSize = 1 << cacheBits
	cacheMask = cacheSize - 1
)

type peerInfo struct {
	peer router.Station
	idle int32
}

// set peer idle
func (p *peerInfo) setIdle() {
	atomic.StoreInt32(&p.idle, 0)
}

// if peer is idle, then set busy and return true. otherwise do nothing and return false.
func (p *peerInfo) trySetBusy() bool {
	return atomic.CompareAndSwapInt32(&p.idle, 0, 1)
}

type bloomPath struct {
	hash  common.Hash
	bloom *types.Bloom
}

// return count of bits that set
func (p *bloomPath) getBloomSetBits() int {
	getBits := func(b []byte) int {
		num := binary.BigEndian.Uint64(b)
		ret := 0
		for num > 0 {
			ret++
			num &= num - 1
		}
		return ret
	}
	ret := 0
	for i := 0; i < 256; i += 8 {
		ret += getBits((*p.bloom)[i : i+8])
	}
	return ret
}

// add path to bloom
func (p *bloomPath) addPath(path string) {
	p.bloom.Add(new(big.Int).SetBytes([]byte(path)))
}

// merge two bloom
func (p *bloomPath) addBloom(b *types.Bloom) {
	if b != nil && *b != *p.bloom {
		lb := p.bloom.Big()
		rb := b.Big()
		p.bloom.SetBytes(lb.Or(lb, rb).Bytes())
	}
}

// return true if the path was in bloom, otherwise false
func (p *bloomPath) testPath(path string) bool {
	return p.bloom.TestBytes([]byte(path))
}

// reset bloom with new hash and bloom
func (p *bloomPath) reset(hash common.Hash, bloom *types.Bloom) {
	p.hash = hash
	p.bloom = bloom
	if p.bloom == nil {
		p.bloom = &types.Bloom{}
	}
}

type txsCache struct {
	cache [cacheSize]bloomPath
}

func (c *txsCache) getTarget(hash common.Hash) *bloomPath {
	index := (int(hash[0]) | (int(hash[1]) << 8)) & cacheMask
	return &c.cache[index]
}

func (c *txsCache) ttlCheck(tx *types.Transaction) {
	hash := tx.Hash()
	target := c.getTarget(hash)
	if target.hash != hash {
		return
	}
	if target.getBloomSetBits() > len(*target.bloom)*3 {
		target.reset(hash, &types.Bloom{})
	}
}

// add a transaction in to the cache
func (c *txsCache) addTx(tx *types.Transaction, bloom *types.Bloom, path string) {
	hash := tx.Hash()
	target := c.getTarget(hash)
	if target.hash != hash {
		target.reset(hash, bloom)
	} else {
		target.addBloom(bloom)
	}
	target.addPath(path)
}

// return a *bloom of the transaction
func (c *txsCache) getTxBloom(tx *types.Transaction) *types.Bloom {
	hash := tx.Hash()
	target := c.getTarget(hash)
	if target.hash != hash {
		return &types.Bloom{}
	}
	return target.bloom
}

func (c *txsCache) copyTxBloom(tx *types.Transaction, dest *types.Bloom) *types.Bloom {
	src := c.getTxBloom(tx)
	copy((*dest)[:], (*src)[:])
	return dest
}

// return true if cache have the transaction
func (c *txsCache) hadTx(tx *types.Transaction) bool {
	hash := tx.Hash()
	target := c.getTarget(hash)
	return target.hash == hash
}

// return true if transaction was already though the path
func (c *txsCache) txHadPath(tx *types.Transaction, path string) bool {
	hash := tx.Hash()
	target := c.getTarget(hash)
	if target.hash != hash {
		return false
	}
	return target.testPath(path)
}

// TxpoolStation is responsible for transaction broadcasting and receiving
type TxpoolStation struct {
	txChan       chan *router.Event
	txpool       *TxPool
	peers        map[string]*peerInfo
	cache        *txsCache
	delayedTxs   []*types.Transaction
	quit         chan struct{}
	loopWG       sync.WaitGroup
	maxGorouting int64
	numGorouting int64
	subs         []router.Subscription
}

// NewTxpoolStation create a new TxpoolStation
func NewTxpoolStation(txpool *TxPool) *TxpoolStation {
	station := &TxpoolStation{
		txChan:       make(chan *router.Event, 1024),
		txpool:       txpool,
		peers:        make(map[string]*peerInfo),
		cache:        &txsCache{},
		maxGorouting: 1024,
		numGorouting: 0,
		quit:         make(chan struct{}),
		subs:         make([]router.Subscription, 4),
	}
	station.subs[0] = router.Subscribe(nil, station.txChan, router.P2PTxMsg, []*TransactionWithPath{}) // recive txs form remote
	station.subs[1] = router.Subscribe(nil, station.txChan, router.NewPeerPassedNotify, nil)           // new peer is handshake completed
	station.subs[2] = router.Subscribe(nil, station.txChan, router.DelPeerNotify, new(string))         // new peer is handshake completed
	station.subs[3] = router.Subscribe(nil, station.txChan, router.NewTxs, []*types.Transaction{})     // NewTxs recived , prepare to broadcast
	station.loopWG.Add(1)
	go station.handleMsg()
	return station
}

// add []transactions in to the cache, and return the []transaction that the cache don't have
func (s *TxpoolStation) addTxs(txs []*TransactionWithPath, from string) []*types.Transaction {
	rtxs := make([]*types.Transaction, 0, len(txs))
	for _, tx := range txs {
		//		if !s.cache.hadTx(tx.Tx) {
		rtxs = append(rtxs, tx.Tx)
		//		}
		s.cache.addTx(tx.Tx, tx.Bloom, from)
	}
	return rtxs
}

func (s *TxpoolStation) broadcast(txs []*types.Transaction) {
	if len(s.peers) == 0 {
		return
	}
	minSend := int(s.txpool.config.MinBroadcast)
	maxSend := len(s.peers) / int(s.txpool.config.RatioBroadcast)
	if maxSend < minSend {
		maxSend = minSend
	}
	sendTask := make(map[*peerInfo][]*TransactionWithPath)
	addToTask := func(name string, peerInfo *peerInfo, txObj *TransactionWithPath) bool {
		tx := txObj.Tx
		if _, ok := sendTask[peerInfo]; ok {
			s.cache.addTx(tx, txObj.Bloom, name)
			sendTask[peerInfo] = append(sendTask[peerInfo], txObj)
			return true
		}
		if !peerInfo.trySetBusy() {
			return false
		}
		s.cache.addTx(tx, txObj.Bloom, name)
		sendTask[peerInfo] = []*TransactionWithPath{txObj}
		return true
	}
	addToTaskAtLeast3 := func(txObj *TransactionWithPath) bool {
		txSend := 0
		tx := txObj.Tx
		s.cache.ttlCheck(tx)
		skipedPeers := make(map[string]*peerInfo, len(s.peers))
		for name, peerInfo := range s.peers {
			if txSend > maxSend {
				break
			}
			if s.cache.txHadPath(tx, name) {
				skipedPeers[name] = peerInfo
				continue
			}
			if addToTask(name, peerInfo, txObj) {
				txSend++
			}
		}
		for name, peerInfo := range skipedPeers {
			if txSend >= minSend {
				break
			}
			if addToTask(name, peerInfo, txObj) {
				txSend++
			}
		}
		s.cache.copyTxBloom(tx, txObj.Bloom)
		return txSend == 0
	}

	oldTxs := s.delayedTxs[:]
	s.delayedTxs = s.delayedTxs[:0]
	oldTxs = append(oldTxs, txs...)

	for _, tx := range oldTxs {
		txObj := &TransactionWithPath{Tx: tx, Bloom: &types.Bloom{}}
		retransmit := addToTaskAtLeast3(txObj)
		if retransmit {
			s.delayedTxs = append(s.delayedTxs, tx)
		}
	}

	if len(sendTask) == 0 {
		return
	}

	s.loopWG.Add(1)
	go func() {
		for peerInfo, txs := range sendTask {
			router.SendTo(nil, peerInfo.peer, router.P2PTxMsg, txs)
			peerInfo.setIdle()
		}
		s.loopWG.Done()
	}()
}

func (s *TxpoolStation) handleMsg() {
	defer s.loopWG.Done()
	for {
		select {
		case <-s.quit:
			return
		case e := <-s.txChan:
			switch e.Typecode {
			case router.NewTxs:
				txs := e.Data.([]*types.Transaction)
				s.broadcast(txs)
			case router.P2PTxMsg:
				if atomic.LoadInt64(&s.numGorouting) >= s.maxGorouting {
					continue
				}
				atomic.AddInt64(&s.numGorouting, 1)
				txs := e.Data.([]*TransactionWithPath)
				//fmt.Printf("bloom:%x\n", *txs[0].Bloom)
				rawTxs := s.addTxs(txs, e.From.Name())
				if len(rawTxs) > 0 {
					s.loopWG.Add(1)
					go func() {
						s.txpool.AddRemotes(rawTxs)
						atomic.AddInt64(&s.numGorouting, -1)
						s.loopWG.Done()
					}()
				}
			case router.NewPeerPassedNotify:
				newpeer := &peerInfo{peer: e.From, idle: 1}
				s.peers[e.From.Name()] = newpeer
				s.delayedTxs = s.delayedTxs[:0]
				s.syncTransactions(newpeer)
			case router.DelPeerNotify:
				delete(s.peers, e.From.Name())
				if len(s.peers) == 0 {
					s.delayedTxs = s.delayedTxs[:0]
				}
			}
		}
	}
}

func (s *TxpoolStation) syncTransactions(peer *peerInfo) {
	var txs []*TransactionWithPath
	pending, _ := s.txpool.Pending()
	for _, batch := range pending {
		for _, tx := range batch {
			bloom := s.cache.copyTxBloom(tx, &types.Bloom{})
			txs = append(txs, &TransactionWithPath{Tx: tx, Bloom: bloom})
		}
	}
	if len(txs) == 0 {
		peer.setIdle()
		return
	}
	s.loopWG.Add(1)
	go func() {
		router.SendTo(nil, peer.peer, router.P2PTxMsg, txs)
		peer.setIdle()
		s.loopWG.Done()
	}()
}

func (s *TxpoolStation) Stop() {
	log.Info("TxpoolHandler stopping...")
	close(s.quit)
	for _, sub := range s.subs {
		sub.Unsubscribe()
	}
	s.loopWG.Wait()
	log.Info("TxpoolHandler stopped.")
}
