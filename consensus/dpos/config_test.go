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
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestRLP(t *testing.T) {
	val, err := DefaultConfig.EncodeRLP()
	if err != nil {
		panic(fmt.Sprintf("EncodeRLP - %v", err))
	}

	_ = val
	if err := DefaultConfig.DecodeRLP(val); err != nil {
		panic(fmt.Sprintf("DecodeRLP - %v", err))
	}

	val, _ = json.Marshal(DefaultConfig)
	t.Log("config  ", string(val))
	t.Log("safesize", DefaultConfig.consensusSize())

	now := uint64(time.Now().UnixNano())
	slot := DefaultConfig.slot(now)
	nslot := DefaultConfig.nextslot(now)
	second := uint64(time.Second)
	layout := "2006-01-02 15:04:05.999999999"
	t.Log("epoch    ", DefaultConfig.epochInterval()/uint64(time.Millisecond), "ms")
	t.Log("interval ", DefaultConfig.blockInterval()/uint64(time.Millisecond), "ms")
	t.Log("now      ", time.Unix(int64(now/second), int64(now%second)).Format(layout))
	t.Log("cur slot ", time.Unix(int64(slot/second), int64(slot%second)).Format(layout), "offset", DefaultConfig.getoffset(slot))
	t.Log("next slot", time.Unix(int64(nslot/second), int64(nslot%second)).Format(layout), "offset", DefaultConfig.getoffset(nslot))

	DefaultConfig.BlockInterval = 500
	DefaultConfig.blockInter.Store(DefaultConfig.BlockInterval * uint64(time.Millisecond))
	DefaultConfig.epochInter.Store(DefaultConfig.blockInterval() * DefaultConfig.BlockFrequency * DefaultConfig.CandidateScheduleSize)
	now = uint64(time.Now().UnixNano())
	slot = DefaultConfig.slot(now)
	nslot = DefaultConfig.nextslot(now)
	second = uint64(time.Second)
	layout = "2006-01-02 15:04:05.999999999"
	t.Log("epoch    ", DefaultConfig.epochInterval()/uint64(time.Millisecond), "ms")
	t.Log("interval ", DefaultConfig.blockInterval()/uint64(time.Millisecond), "ms")
	t.Log("now      ", time.Unix(int64(now/second), int64(now%second)).Format(layout))
	t.Log("cur slot ", time.Unix(int64(slot/second), int64(slot%second)).Format(layout), "offset", DefaultConfig.getoffset(slot))
	t.Log("next slot", time.Unix(int64(nslot/second), int64(nslot%second)).Format(layout), "offset", DefaultConfig.getoffset(nslot))
}
