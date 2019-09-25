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
	"fmt"
	"math/big"
	"testing"
)

func TestConfig(t *testing.T) {
	if 0 != DefaultConfig.decimals().Cmp(big.NewInt(1000000000000000000)) {
		panic(fmt.Errorf("Config decimals mismatch"))
	}

	if 0 != DefaultConfig.extraBlockReward().Cmp(big.NewInt(1000000000000000000)) {
		panic(fmt.Errorf("Config extraBlockReward mismatch"))
	}

	if 0 != DefaultConfig.blockReward().Cmp(big.NewInt(5000000000000000000)) {
		panic(fmt.Errorf("Config blockReward mismatch"))
	}

	if 3000000000 != DefaultConfig.blockInterval() {
		panic(fmt.Errorf("Config blockInterval mismatch"))
	}

	if 54000000000 != DefaultConfig.mepochInterval() {
		panic(fmt.Errorf("Config mepochInterval mismatch"))
	}

	if 1080000000000 != DefaultConfig.epochInterval() {
		panic(fmt.Errorf("Config epochInterval mismatch"))
	}

	if 3 != DefaultConfig.consensusSize() {
		panic(fmt.Errorf("Config consensusSize mismatch"))
	}

	if 3 != DefaultConfig.consensusSize() {
		panic(fmt.Errorf("Config Cache consensusSize mismatch"))
	}

	if 0 != DefaultConfig.slot(1567591745) {
		panic(fmt.Errorf("Config slot mismatch"))
	}

	if 3000000000 != DefaultConfig.nextslot(1567591745) {
		panic(fmt.Errorf("Config nextslot mismatch"))
	}

	if 0 != DefaultConfig.getoffset(1567591745, 1) {
		panic(fmt.Errorf("Config getoffset mismatch"))
	}

	if 0 != DefaultConfig.getoffset(1567591745, 2) {
		panic(fmt.Errorf("Config getoffset mismatch"))
	}

	if 15639786 != DefaultConfig.epoch(1567591745) {
		panic(fmt.Errorf("Config epoch mismatch"))
	}

	if 1555777080000000000 != DefaultConfig.epochTimeStamp(2) {
		panic(fmt.Errorf("Config epochTimeStamp mismatch"))
	}

	if 0 != DefaultConfig.shouldCounter(1567591745000, 1567591745000) {
		panic(fmt.Errorf("Config epochTimeStamp mismatch"))
	}

	if 3 != DefaultConfig.shouldCounter(1521370523000, 1567591745000) {
		panic(fmt.Errorf("Config epochTimeStamp mismatch"))
	}

	if 10 != DefaultConfig.minMEpoch() {
		panic(fmt.Errorf("Config minMEpoch mismatch"))
	}

	if 180 != DefaultConfig.minBlockCnt() {
		panic(fmt.Errorf("Config minBlockCnt mismatch"))
	}

	if err := DefaultConfig.IsValid(); err != nil {
		panic(fmt.Errorf("Config IsValid err %v", err))
	}

	DefaultConfig.epochInter.Store(uint64(1070000))
	if err := DefaultConfig.IsValid(); err == nil {
		panic(fmt.Errorf("Config IsValid err %v", err))
	}
}
