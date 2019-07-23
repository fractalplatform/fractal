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
	"strings"
	"testing"
)

var (
	big0              = big.NewInt(0)
	big1              = big.NewInt(1)
	big2              = big.NewInt(2)
	big3              = big.NewInt(3)
	big10             = big.NewInt(10)
	minStakeCandidate = new(big.Int).Mul(DefaultConfig.CandidateMinQuantity, DefaultConfig.unitStake())
	minStakeVote      = new(big.Int).Mul(DefaultConfig.VoterMinQuantity, DefaultConfig.unitStake())
)

func TestCandiate(t *testing.T) {
	ldb, function := newTestLDB()
	db, err := NewLDB(ldb)
	defer function()
	if err != nil {
		panic(fmt.Errorf("create db failed --- %v", err))
	}
	sys := &System{
		config: DefaultConfig,
		IDB:    db,
	}
	_ = sys
	for index, candidate := range candidates {
		if err := db.SetState(&GlobalState{
			Epoch:         uint64(index),
			PreEpoch:      uint64(index),
			TotalQuantity: big.NewInt(0),
		}); err != nil {
			panic(fmt.Errorf("SetState --- %v", err))
		}
		if err := sys.IDB.SetAvailableQuantity(uint64(index), candidate, new(big.Int).Mul(big10, minStakeCandidate)); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := sys.RegCandidate(uint64(index), candidate, strings.Repeat(fmt.Sprintf("www.%v.com", candidate), int(DefaultConfig.MaxURLLen)), new(big.Int).Mul(big1, minStakeCandidate), uint64(index), 0); !strings.Contains(err.Error(), "invalid url") {
			panic(fmt.Sprintf("RegCandidate invalid url %v mismatch", err))
		}
		if err := sys.RegCandidate(uint64(index), candidate, fmt.Sprintf("www.%v.com", candidate), big1, uint64(index), 0); !strings.Contains(err.Error(), "non divisibility") {
			panic(fmt.Sprintf("RegCandidate invalid stake %v mismatch", err))
		}
		if err := sys.RegCandidate(uint64(index), candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big0, minStakeCandidate), uint64(index), 0); !strings.Contains(err.Error(), "insufficient") {
			panic(fmt.Sprintf("RegCandidate invalid stake %v mismatch", err))
		}
		if err := sys.RegCandidate(uint64(index), candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big1, minStakeCandidate), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("RegCandidate %v", err))
		}

		if ncandidates, _ := sys.GetCandidates(uint64(index)); len(ncandidates) != 1 {
			panic(fmt.Sprintf("GetCandidates mismatch"))
		}

		if candidateInfo, _ := sys.GetCandidate(uint64(index), candidate); candidateInfo.Quantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 || candidateInfo.TotalQuantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 {
			panic(fmt.Sprintf("GetCandidate mismatch"))
		}

		if gstate, _ := sys.GetState(uint64(index)); gstate.TotalQuantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 {
			panic(fmt.Sprintf("GetState mismatch"))
		}

		if err := sys.RegCandidate(uint64(index), candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big1, minStakeCandidate), uint64(index), 0); !strings.Contains(err.Error(), "invalid candidate") {
			panic(fmt.Sprintf("RegCandidate invalid name %v mismatch", err))
		}

		if err := sys.UpdateCandidate(uint64(index), candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big2, minStakeCandidate), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("UpdateCandidate %v", err))
		}

		if ncandidates, _ := sys.GetCandidates(uint64(index)); len(ncandidates) != 1 {
			panic(fmt.Sprintf("GetCandidates num mismatch"))
		}

		if candidateInfo, _ := sys.GetCandidate(uint64(index), candidate); candidateInfo.Quantity.Cmp(new(big.Int).Mul(big3, DefaultConfig.CandidateMinQuantity)) != 0 || candidateInfo.TotalQuantity.Cmp(new(big.Int).Mul(big3, DefaultConfig.CandidateMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetCandidate quantity mismatch"))
		}

		if gstate, _ := sys.GetState(uint64(index)); gstate.TotalQuantity.Cmp(new(big.Int).Mul(big3, DefaultConfig.CandidateMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetState mismatch"))
		}

		if err := sys.UnregCandidate(uint64(index), candidate, uint64(index), 0); err != nil {
			panic(fmt.Sprintf("UnregCandidate %v", err))
		}

		// if err := sys.RefundCandidate(uint64(index), candidate, uint64(index)); err == nil {
		// 	panic(fmt.Sprintf("RefundCandidate %v", err))
		// }

		if ncandidates, _ := sys.GetCandidates(uint64(index)); len(ncandidates) != 1 {
			panic(fmt.Sprintf("UnregCandidate %v mismatch", err))
		}
	}

}

func TestVote(t *testing.T) {
	ldb, function := newTestLDB()
	db, err := NewLDB(ldb)
	defer function()
	if err != nil {
		panic(fmt.Errorf("create db failed --- %v", err))
	}
	sys := &System{
		config: DefaultConfig,
		IDB:    db,
	}
	_ = sys

	for index, candidate := range candidates {
		if err := db.SetState(&GlobalState{
			Epoch:         uint64(index),
			PreEpoch:      uint64(index),
			TotalQuantity: big.NewInt(0),
		}); err != nil {
			panic(fmt.Errorf("SetState --- %v", err))
		}
		if err := sys.IDB.SetAvailableQuantity(uint64(index), candidate, new(big.Int).Mul(big10, minStakeCandidate)); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := sys.RegCandidate(uint64(index), candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big1, minStakeCandidate), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("RegCandidate %v", err))
		}
		if err := sys.RefundCandidate(uint64(index), candidate, uint64(index), 0); err == nil {
			panic(fmt.Sprintf("RefundCandidate %v", err))
		}
	}

	for index, voter := range voters {
		if err := sys.IDB.SetAvailableQuantity(uint64(index), voter, new(big.Int).Mul(big10, minStakeVote)); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := sys.VoteCandidate(uint64(index), voter, "test", new(big.Int).Mul(big1, minStakeVote), uint64(index), 0); !strings.Contains(err.Error(), "invalid candidate") {
			panic(fmt.Sprintf("VoteCandidate invalid candidate %v mismatch", err))
		}

		if err := sys.VoteCandidate(uint64(index), voter, candidates[index], big1, uint64(index), 0); !strings.Contains(err.Error(), "non divisibility") {
			panic(fmt.Sprintf("VoteCandidate invalid stake %v mismatch", err))
		}

		if err := sys.VoteCandidate(uint64(index), voter, candidates[index], new(big.Int).Mul(big0, minStakeVote), uint64(index), 0); !strings.Contains(err.Error(), "insufficient") {
			panic(fmt.Sprintf("VoteCandidate invalid stake %v mismatch", err))
		}

		if err := sys.VoteCandidate(uint64(index), voter, candidates[index], new(big.Int).Mul(big1, minStakeVote), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("VoteCandidate --- %v", err))
		}

		if candidateInfo, _ := sys.GetCandidate(uint64(index), candidates[index]); new(big.Int).Sub(candidateInfo.TotalQuantity, candidateInfo.Quantity).Cmp(new(big.Int).Mul(big1, DefaultConfig.VoterMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetCandidate mismatch"))
		}

		if err := sys.VoteCandidate(uint64(index), voter, candidates[index], new(big.Int).Mul(big1, minStakeVote), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("VoteCandidate --- %v", err))
		}

		if candidateInfo, _ := sys.GetCandidate(uint64(index), candidates[index]); new(big.Int).Sub(candidateInfo.TotalQuantity, candidateInfo.Quantity).Cmp(new(big.Int).Mul(big2, DefaultConfig.VoterMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetCandidate mismatch"))
		} else if candidateInfo.Type != Normal {
			panic(fmt.Sprintf("GetCandidate  type mismatch"))
		}
	}
}

func TestCandidateVote(t *testing.T) {
	ldb, function := newTestLDB()
	db, err := NewLDB(ldb)
	defer function()
	if err != nil {
		panic(fmt.Errorf("create db failed --- %v", err))
	}
	sys := &System{
		config: DefaultConfig,
		IDB:    db,
	}
	_ = sys

	for index, candidate := range candidates {
		if err := db.SetState(&GlobalState{
			Epoch:         uint64(index),
			PreEpoch:      uint64(index),
			TotalQuantity: big.NewInt(0),
		}); err != nil {
			panic(fmt.Errorf("SetState --- %v", err))
		}
		if err := sys.IDB.SetAvailableQuantity(uint64(index), candidate, new(big.Int).Mul(big10, minStakeCandidate)); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := sys.RegCandidate(uint64(index), candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big1, minStakeCandidate), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("RegCandidate %v", err))
		}
		// if err := sys.UnregCandidate(uint64(index), candidate, uint64(index)); err != nil {
		// 	panic(fmt.Sprintf("UnregCandidate %v", err))
		// }
	}

	for index, voter := range voters {
		if err := sys.IDB.SetAvailableQuantity(uint64(index), voter, new(big.Int).Mul(big10, minStakeVote)); err != nil {
			panic(fmt.Errorf("SetAvailableQuantity --- %v", err))
		}
		if err := sys.VoteCandidate(uint64(index), voter, "test", new(big.Int).Mul(big1, minStakeVote), uint64(index), 0); !strings.Contains(err.Error(), "invalid candidate") {
			panic(fmt.Sprintf("VoteCandidate invalid candidate %v mismatch", err))
		}

		// if err := sys.VoteCandidate(uint64(index), voter, candidates[index], big1, uint64(index)); !strings.Contains(err.Error(), "not in normal") {
		// 	panic(fmt.Sprintf("VoteCandidate invalid stake %v mismatch", err))
		// }

		// if err := sys.VoteCandidate(uint64(index), voter, candidates[index], new(big.Int).Mul(big1, minStakeVote), uint64(index)); !strings.Contains(err.Error(), "insufficient") {
		// 	panic(fmt.Sprintf("VoteCandidate invalid stake %v mismatch", err))
		// }

		if err := sys.VoteCandidate(uint64(index), voter, candidates[index], new(big.Int).Mul(big1, minStakeVote), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("VoteCandidate --- %v", err))
		}

		if candidateInfo, _ := sys.GetCandidate(uint64(index), candidates[index]); new(big.Int).Sub(candidateInfo.TotalQuantity, candidateInfo.Quantity).Cmp(new(big.Int).Mul(big1, DefaultConfig.VoterMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetCandidate mismatch"))
		}

		if err := sys.VoteCandidate(uint64(index), voter, candidates[index], new(big.Int).Mul(big1, minStakeVote), uint64(index), 0); err != nil {
			panic(fmt.Sprintf("VoteCandidate --- %v", err))
		}

		if candidateInfo, _ := sys.GetCandidate(uint64(index), candidates[index]); new(big.Int).Sub(candidateInfo.TotalQuantity, candidateInfo.Quantity).Cmp(new(big.Int).Mul(big2, DefaultConfig.VoterMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetCandidate mismatch"))
		} else if candidateInfo.Type != Normal {
			panic(fmt.Sprintf("GetCandidate  type mismatch"))
		}
	}
}
