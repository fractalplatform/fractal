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
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/params"
)

type minerStatus int32

const (
	Stopped int32 = iota
	Starting
	Started
	Stopping
)

// Miner creates blocks and searches for proof values.
type Miner struct {
	worker *Worker
	mining int32
}

var (
	minerOpErr = []string{
		Stopped:  "miner already stoped",
		Starting: "miner is starting...",
		Started:  "miner already started",
		Stopping: "miner is stopping...",
	}
)

// NewMiner creates a miner.
func NewMiner(c context) *Miner {
	return &Miner{
		worker: newWorker(c),
		mining: Stopped,
	}
}

// Start start worker
func (miner *Miner) Start(force bool) bool {
	if !atomic.CompareAndSwapInt32(&miner.mining, Stopped, Starting) {
		return false
	}
	log.Info("Starting mining operation")
	miner.worker.start(force)
	atomic.StoreInt32(&miner.mining, Started)
	return true
}

// Stop stop worker
func (miner *Miner) Stop() bool {
	if !atomic.CompareAndSwapInt32(&miner.mining, Started, Stopping) {
		return false
	}
	log.Info("Stopping mining operation")
	miner.worker.stop()
	atomic.StoreInt32(&miner.mining, Stopped)
	return true
}

// Mining worker is working
func (miner *Miner) Mining() bool {
	return atomic.LoadInt32(&miner.mining) > 0
}

// SetCoinbase coinbase name & private key
func (miner *Miner) SetCoinbase(name string, privKey string) error {
	bts, err := hex.DecodeString(privKey)
	if err != nil {
		return err
	}
	priv, err := crypto.ToECDSA(bts)
	if err != nil {
		return err
	}
	miner.worker.setCoinbase(name, priv)
	return nil
}

// SetDelayDuration delay broadcast block when mint block (unit:ms)
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
