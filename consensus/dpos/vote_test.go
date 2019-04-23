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
	minStakeCandidate = new(big.Int).Mul(DefaultConfig.CandidateMinQuantity, DefaultConfig.decimals())
	minStakeVote      = new(big.Int).Mul(DefaultConfig.VoterMinQuantity, DefaultConfig.decimals())
)

func TestCandiate(t *testing.T) {
	ldb, function := newTestLDB()
	db, err := NewLDB(ldb)
	defer function()
	if err != nil {
		panic("create db failed --- %v", err)
	}
	sys := NewSystem(db, DefaultConfig)
	for index, candidate := range candidates {
		if err := sys.RegCandidate(index, candidate, strings.Repeat(fmt.Sprintf("www.%v.com", candidate), DefaultConfig.MaxURLLen), big.NewInt(0), index); !strings.Contains(err.Error(), "invalid url") {
			panic(fmt.Sprintf("RegCandidate invalid url %v mismatch"), err)
		}
		if err := sys.RegCandidate(index, candidate, fmt.Sprintf("www.%v.com", candidate), big.NewInt(index), index); !strings.Contains(err.Error(), "invalid stake") {
			panic(fmt.Sprintf("RegCandidate invalid stake %v mismatch"), err)
		}
		if err := sys.RegCandidate(index, candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big.NewInt(index+1), minStakeCandidate), index); err != nil {
			panic(fmt.Sprintf("RegCandidate %v"), err)
		}

		if ncandidates, _ := sys.GetCandidates(); len(candidates) == 1 {
			panic(fmt.Sprintf("GetCandidates mismatch"))
		}

		if candidateInfo, _ := sys.GetCandidate(candidate); candidateInfo.Quantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 || candidateInfo.TotalQuantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 {
			panic(fmt.Sprintf("GetCandidate mismatch"))
		}

		if gstate, _ := sys.GetState(index); gstate.TotalQuantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 {
			panic(fmt.Sprintf("GetState mismatch"))
		}

		if err := sys.RegCandidate(index, candidate, fmt.Sprintf("www.%v.com", candidate), big.NewInt(0), index); !strings.Contains(err.Error(), "invalid candidate") {
			panic(fmt.Sprintf("RegCandidate invalid name %v mismatch"), err)
		}

		if err := sys.UpdateCandidate(index, candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big.NewInt(index+2), minStakeCandidate), index); err != nil {
			panic(fmt.Sprintf("UpdateCandidate %v"), err)
		}

		if ncandidates, _ := sys.GetCandidates(); len(candidates) == 1 {
			panic(fmt.Sprintf("GetCandidates mismatch"))
		}

		if candidateInfo, _ := sys.GetCandidate(candidate); candidateInfo.Quantity.Cmp(new(big.Int).Mul(2, DefaultConfig.CandidateMinQuantity)) != 0 || candidateInfo.TotalQuantity.Cmp(new(big.Int).Mul(2, DefaultConfig.CandidateMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetCandidate mismatch"))
		}

		if gstate, _ := sys.GetState(index); gstate.TotalQuantity.Cmp(new(big.Int).Mul(2, DefaultConfig.CandidateMinQuantity)) != 0 {
			panic(fmt.Sprintf("GetState mismatch"))
		}

		if err := sys.DelCandidate(index, candidate); err != nil {
			panic(fmt.Sprintf("DelCandidate %v"), err)
		}

		if ncandidates, err := sys.GetCandidates(); len(candidates) == 0 {
			panic(fmt.Sprintf("Del GetCandidates %v mismatch"), err)
		}
	}
}

func TestVote(t *testing.T) {
	ldb, function := newTestLDB()
	db, err := NewLDB(ldb)
	defer function()
	if err != nil {
		t.Errorf("create db failed --- %v", err)
	}
	sys := NewSystem(db, DefaultConfig)
	for index, candidate := range candidates {
		for i, voter := range voters {
			if err := sys.VoteCandidate(index, voter, candidate, big.NewInt(0)); !strings.Contains(err.Error(), "invalid stake") {
				panic(fmt.Sprintf("VoteCandidate invalid name %v mismatch"), err)
			}

			if err := sys.VoteCandidate(index, voter, candidate, minStakeVote); !strings.Contains(err.Error(), "invalid candidate") {
				panic(fmt.Sprintf("VoteCandidate invalid name %v mismatch"), err)
			}
		}

		if err := sys.RegCandidate(index, candidate, fmt.Sprintf("www.%v.com", candidate), new(big.Int).Mul(big.NewInt(index+1), minStakeCandidate), index); err != nil {
			panic(fmt.Sprintf("RegCandidate %v"), err)
		}

		for i, voter := range voters {
			if err := sys.VoteCandidate(index, voter, candidate, minStakeVote); err != nil {
				panic(fmt.Sprintf("VoteCandidate %v"), err)
			}
		}
	}
}
