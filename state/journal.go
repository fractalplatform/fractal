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

package state

import (
	"github.com/fractalplatform/fractal/common"
)

type journalEntry interface {
	revert(*StateDB)
	dirtied() *string
}

type journal struct {
	entries []journalEntry // Current changes tracked by the journal
	dirties map[string]int // Dirty key and the number of changes
}

func newJournal() *journal {
	return &journal{
		dirties: make(map[string]int),
	}
}

func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if key := entry.dirtied(); key != nil {
		j.dirties[*key]++
	}
}

func (j *journal) revert(statedb *StateDB, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		j.entries[i].revert(statedb)

		if key := j.entries[i].dirtied(); key != nil {
			if j.dirties[*key]--; j.dirties[*key] == 0 {
				delete(j.dirties, *key)
			}
		}
	}
	j.entries = j.entries[:snapshot]
}

func (j *journal) dirty(key string) {
	j.dirties[key]++
}

func (j *journal) length() int {
	return len(j.entries)
}

type (
	stateChange struct {
		key      *string
		prevalue []byte
	}

	refundChange struct {
		prev uint64
	}
	addLogChange struct {
		txhash common.Hash
	}
	addPreimageChange struct {
		hash common.Hash
	}
)

func (ch stateChange) revert(s *StateDB) {
	s.set(*ch.key, ch.prevalue)
}

func (ch stateChange) dirtied() *string {
	return ch.key
}

func (ch refundChange) revert(s *StateDB) {
	s.refund = ch.prev
}

func (ch refundChange) dirtied() *string {
	return nil
}

func (ch addLogChange) revert(s *StateDB) {
	logs := s.logs[ch.txhash]
	if len(logs) == 1 {
		delete(s.logs, ch.txhash)
	} else {
		s.logs[ch.txhash] = logs[:len(logs)-1]
	}
	s.logSize--
}

func (ch addLogChange) dirtied() *string {
	return nil
}

func (ch addPreimageChange) revert(s *StateDB) {
	delete(s.preimages, ch.hash)
}

func (ch addPreimageChange) dirtied() *string {
	return nil
}
