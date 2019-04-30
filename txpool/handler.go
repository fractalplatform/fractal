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
	"sync/atomic"
	"time"

	"github.com/fractalplatform/fractal/common"
	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/types"
)

const (
	maxKonwnTxs      = 32
	txsSendDelay     = 50 * time.Millisecond
	txsSendThreshold = 32
	cacheBits        = 10
	cacheSize        = 1 << cacheBits
	cacheMask        = cacheSize - 1
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
	getBits := func(num uint64) int {
		ret := 0
		for num > 0 {
			ret++
			num &= num - 1
		}
		return ret
	}
	ret := getBits(binary.BigEndian.Uint64((*p.bloom)[0:8]))
	ret += getBits(binary.BigEndian.Uint64((*p.bloom)[8:16]))
	ret += getBits(binary.BigEndian.Uint64((*p.bloom)[16:24]))
	ret += getBits(binary.BigEndian.Uint64((*p.bloom)[24:32]))
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
	txChan     chan *router.Event
	txpool     *TxPool
	peers      map[string]*peerInfo
	cache      *txsCache
	delayedTxs []*types.Transaction
}

// NewTxpoolStation create a new TxpoolStation
func NewTxpoolStation(txpool *TxPool) *TxpoolStation {
	station := &TxpoolStation{
		txChan: make(chan *router.Event, 1024),
		txpool: txpool,
		peers:  make(map[string]*peerInfo),
		cache:  &txsCache{},
	}
	router.Subscribe(nil, station.txChan, router.P2PTxMsg, []*TransactionWithPath{}) // recive txs form remote
	router.Subscribe(nil, station.txChan, router.NewPeerPassedNotify, nil)           // new peer is handshake completed
	router.Subscribe(nil, station.txChan, router.DelPeerNotify, new(string))         // new peer is handshake completed
	router.Subscribe(nil, station.txChan, router.NewTxs, []*types.Transaction{})     // NewTxs recived , prepare to broadcast
	go station.handleMsg()
	return station
}

// add []transactions in to the cache, and return the []transaction that the cache don't have
func (s *TxpoolStation) addTxs(txs []*TransactionWithPath, from string) []*types.Transaction {
	rtxs := make([]*types.Transaction, 0, len(txs))
	for _, tx := range txs {
		if !s.cache.hadTx(tx.Tx) {
			rtxs = append(rtxs, tx.Tx)
		}
		s.cache.addTx(tx.Tx, tx.Bloom, from)
	}
	return rtxs
}

func (s *TxpoolStation) broadcast(txs []*types.Transaction) {
	sendTask := make(map[*peerInfo][]*TransactionWithPath)
	addToTask := func(txObj *TransactionWithPath) bool {
		txSend := 0
		retransmit := true // retransmit = true, if the tx don't send because of all peers were busy
		tx := txObj.Tx
		s.cache.ttlCheck(tx)
		for name, peerInfo := range s.peers {
			if txSend > 3 {
				break
			}
			if s.cache.txHadPath(tx, name) {
				retransmit = false
				continue
			}
			if _, ok := sendTask[peerInfo]; ok {
				s.cache.addTx(tx, nil, name)
				sendTask[peerInfo] = append(sendTask[peerInfo], txObj)
				txSend++
				continue
			}
			if !peerInfo.trySetBusy() {
				continue
			}
			s.cache.addTx(tx, nil, name)
			sendTask[peerInfo] = []*TransactionWithPath{txObj}
			txSend++
		}
		if txSend > 0 {
			s.cache.copyTxBloom(tx, txObj.Bloom)
			return false
		}
		return retransmit
	}

	oldTxs := s.delayedTxs[:]
	s.delayedTxs = s.delayedTxs[:0]
	oldTxs = append(oldTxs, txs...)

	for _, tx := range oldTxs {
		txObj := &TransactionWithPath{Tx: tx, Bloom: &types.Bloom{}}
		retransmit := addToTask(txObj)
		if retransmit {
			s.delayedTxs = append(s.delayedTxs, tx)
		}
	}

	if len(sendTask) == 0 {
		return
	}

	go func() {
		for peerInfo, txs := range sendTask {
			router.SendTo(nil, peerInfo.peer, router.P2PTxMsg, txs)
			peerInfo.setIdle()
		}
	}()
}

func (s *TxpoolStation) handleMsg() {
	for {
		e := <-s.txChan
		switch e.Typecode {
		case router.NewTxs:
			txs := e.Data.([]*types.Transaction)
			s.broadcast(txs)
		case router.P2PTxMsg:
			txs := e.Data.([]*TransactionWithPath)
			//fmt.Printf("bloom:%x\n", *txs[0].Bloom)
			rawTxs := s.addTxs(txs, e.From.Name())
			if len(rawTxs) > 0 {
				go s.txpool.AddRemotes(rawTxs)
			}
		case router.NewPeerPassedNotify:
			newpeer := &peerInfo{peer: e.From, idle: 1}
			s.peers[e.From.Name()] = newpeer
			s.syncTransactions(newpeer)
		case router.DelPeerNotify:
			delete(s.peers, e.From.Name())
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
	go func() {
		router.SendTo(nil, peer.peer, router.P2PTxMsg, txs)
		peer.setIdle()
	}()
}
