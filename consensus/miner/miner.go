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
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
)

// Miner creates blocks and searches for proof values.
type Miner struct {
	worker *Worker

	mining      int32
	canStart    int32 // can start indicates whether we can start the mining operation
	shouldStart int32 // should start indicates whether we should start after sync
}

// NewMiner creates a miner.
func NewMiner(consensus consensus.IConsensus) *Miner {
	miner := &Miner{
		worker:   newWorker(consensus),
		canStart: 1,
	}
	go miner.update()
	return miner
}

// update keeps track of events.
func (miner *Miner) update() {
	// 	downloaderEventChan := make(chan)
	// 	downloaderEventSub := event.Subscription{}
	//  defer downloaderEventSub.Unsubscribe()
	// out:
	// 	for {
	// 		select {
	// 		case ev := range downloaderEventChan:
	// 			switch ev.Data.(type) {
	// 			case downloader.StartEvent:
	// 				atomic.StoreInt32(&miner.canStart, 0)
	// 				if miner.Mining() {
	// 					miner.Stop()
	// 					atomic.StoreInt32(&miner.shouldStart, 1)
	// 					log.Info("Mining aborted due to sync")
	// 				}
	// 			case downloader.DoneEvent, downloader.FailedEvent:
	// 				shouldStart := atomic.LoadInt32(&miner.shouldStart) == 1
	// 				atomic.StoreInt32(&miner.canStart, 1)
	// 				atomic.StoreInt32(&miner.shouldStart, 0)
	// 				if shouldStart {
	// 					miner.Start()
	// 				}
	// 				// unsubscribe. we're only interested in this event once
	// 				// stop immediately and ignore all further pending events
	// 				break out
	// 			}
	// 		case ev := range downloaderEventSub.Err()
	// 			break out
	// 		}
	// 	}
}

// Start start worker
func (miner *Miner) Start(force bool) bool {
	atomic.StoreInt32(&miner.shouldStart, 1)
	if atomic.LoadInt32(&miner.canStart) == 0 {
		log.Error("Network syncing, will start miner afterwards")
		return false
	}
	if !atomic.CompareAndSwapInt32(&miner.mining, 0, 1) {
		log.Error("miner already started")
		return false
	}
	log.Info("Starting mining operation")
	miner.worker.start(force)
	return true
}

// Stop stop worker
func (miner *Miner) Stop() bool {
	if !atomic.CompareAndSwapInt32(&miner.mining, 1, 0) {
		log.Error("miner already stopped")
		return false
	}
	log.Info("Stopping mining operation")
	atomic.StoreInt32(&miner.shouldStart, 0)
	miner.worker.stop()
	return true
}

// Mining wroker is wroking
func (miner *Miner) Mining() bool {
	return atomic.LoadInt32(&miner.mining) > 0
}

// SetCoinbase coinbase name & private key
func (miner *Miner) SetCoinbase(name string, privKeys []string) error {
	privs := make([]*ecdsa.PrivateKey, 0, len(privKeys))
	for _, privKey := range privKeys {
		bts, err := hex.DecodeString(privKey)
		if err != nil {
			return err
		}
		priv, err := crypto.ToECDSA(bts)
		if err != nil {
			return err
		}
		privs = append(privs, priv)
	}

	miner.worker.setCoinbase(name, privs)
	return nil
}

// SetDelayDuration delay broacast block when mint block (unit:ms)
func (miner *Miner) SetDelayDuration(delayDuration uint64) error {
	return miner.worker.setDelayDuration(delayDuration)
}

// SetExtra extra data
func (miner *Miner) SetExtra(extra []byte) error {
	if uint64(len(extra)) > params.MaximumExtraDataSize-65 {
		err := fmt.Errorf("extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize-65)
		log.Warn("SetExtra", "error", err)
		return err
	}
	miner.worker.setExtra(extra)
	return nil
}
