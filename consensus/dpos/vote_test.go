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
	"math/big"
	"testing"
)

var (
	mincandidatestake = big.NewInt(0).Mul(DefaultConfig.unitStake(), DefaultConfig.CandidateMinQuantity)
	minvoterstake     = big.NewInt(0).Mul(DefaultConfig.unitStake(), DefaultConfig.VoterMinQuantity)
	big1              = big.NewInt(1)
)

func TestVote(t *testing.T) {
	ldb, function := newTestLDB()
	db, err := NewLDB(ldb)
	defer function()
	if err != nil {
		t.Errorf("create db failed --- %v", err)
	}
	dpos := &System{
		config: DefaultConfig,
		IDB:    db,
	}
	dpos.SetState(&globalState{
		Height:                 0,
		ActivatedTotalQuantity: big.NewInt(0),
	})

	// RegCandidate
	candidate := "testcandidate"
	url := "testurl"
	stake := big.NewInt(0).Sub(mincandidatestake, big1)

	if _, err := dpos.UnregCandidate(candidate); err == nil {
		t.Errorf("UnregCandidate should failed --- %v", err)
	}

	err = dpos.RegCandidate(candidate, url, stake)
	if err == nil {
		t.Errorf("RegCandidate should failed --- %v", err)
	}

	if _, err := dpos.UnregCandidate(candidate); err == nil {
		t.Errorf("UnregCandidate should failed --- %v", err)
	}

	err = dpos.RegCandidate(candidate, url, mincandidatestake)
	if nil != err {
		t.Errorf("RegCandidate failed --- %v", err)
	}

	if gstate, err := dpos.GetState(LastBlockHeight); err != nil || gstate.TotalQuantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 {
		t.Errorf("gstate totalQuantity mismatch --- %v(%v, %v)", err, gstate.TotalQuantity, DefaultConfig.CandidateMinQuantity)
	}

	// GetCandidate
	if prod, err := dpos.GetCandidate(candidate); err != nil {
		t.Errorf("GetCandidate failed --- %v", err)
	} else if prod.Name != candidate || prod.URL != url || prod.Quantity.Cmp(DefaultConfig.CandidateMinQuantity) != 0 || prod.Quantity.Cmp(prod.TotalQuantity) != 0 {
		t.Errorf("candidate info not match")
	}

	// Candidates
	if prods, err := dpos.Candidates(); err != nil || len(prods) != 1 {
		t.Errorf("candidates mismatch")
	}

	// CandidatesSize
	if size, err := dpos.CandidatesSize(); err != nil || size != 1 {
		t.Errorf("candidates mismatch")
	}

	// VoteCandidate
	voter := "testvoter"
	vstake := big.NewInt(0).Sub(minvoterstake, big1)
	err = dpos.VoteCandidate(voter, candidate, vstake)
	if err == nil {
		t.Errorf("VoteCandidate should failed --- %v", err)
	}

	err = dpos.VoteCandidate(voter, candidate, minvoterstake)
	if nil != err {
		t.Errorf("VoterCandidate failed --- %v", err)
	}

	prod, _ := dpos.GetCandidate(candidate)
	gstate, _ := dpos.GetState(LastBlockHeight)
	if prod.TotalQuantity.Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("gstate totalQuantity mismatch --- %v(%v, %v)", err, gstate.TotalQuantity, prod.TotalQuantity)
	} else if new(big.Int).Sub(prod.TotalQuantity, prod.Quantity).Cmp(DefaultConfig.VoterMinQuantity) != 0 {
		t.Errorf("candidate totalQuantity mismatch --- %v(%v, %v)", err, prod.TotalQuantity, prod.Quantity)
	}

	// GetVoter
	if vote, err := dpos.GetVoter(voter); err != nil {
		t.Errorf("GetVoter failed --- %v", err)
	} else if vote.Name != voter || vote.Candidate != candidate || vote.Quantity.Cmp(DefaultConfig.VoterMinQuantity) != 0 {
		t.Errorf("voter info not match --- %v", err)
	}

	// voter cant reg candidate
	err = dpos.RegCandidate(voter, url, mincandidatestake)
	if err.Error() != "invalid candidate testvoter(alreay vote to testcandidate)" {
		t.Errorf("wrong err type --- %v", err)
	}

	// test change
	candidate2 := "testcandidate2"
	url2 := "testurl2"
	dpos.RegCandidate(candidate2, url2, mincandidatestake)
	// Candidates
	if prods, err := dpos.Candidates(); err != nil || len(prods) != 2 {
		t.Errorf("candidates mismatch")
	}

	if err := dpos.ChangeCandidate(voter, candidate2); err != nil {
		t.Errorf("ChangeCandidate failed --- %v", err)
	}

	vote, _ := dpos.GetVoter(voter)
	prod, _ = dpos.GetCandidate(candidate)
	prod2, _ := dpos.GetCandidate(candidate2)
	gstate, _ = dpos.GetState(LastBlockHeight)

	if vote.Candidate != prod2.Name || vote.Quantity.Cmp(DefaultConfig.VoterMinQuantity) != 0 ||
		prod.Quantity.Cmp(prod.TotalQuantity) != 0 || new(big.Int).Add(prod.TotalQuantity, prod2.TotalQuantity).Cmp(gstate.TotalQuantity) != 0 {
		t.Log(prod2.TotalQuantity, gstate.TotalQuantity)
		t.Error("Change stake not work")
	}

	if _, err := dpos.UnvoteCandidate(voter); err != nil {
		t.Errorf("UnvoteCandidate failed --- %v", err)
	} else if vote, err := dpos.GetVoter(voter); err != nil || vote != nil {
		t.Errorf("UnvoteCandidate failed --- %v", err)
	}
	prod2, _ = dpos.GetCandidate(candidate2)
	gstate, _ = dpos.GetState(LastBlockHeight)
	if prod.Quantity.Cmp(prod.TotalQuantity) != 0 ||
		prod2.Quantity.Cmp(prod2.TotalQuantity) != 0 || new(big.Int).Add(prod.TotalQuantity, prod2.TotalQuantity).Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("UnvoteCandidate failed")
	}

	if _, err := dpos.UnregCandidate(candidate); err != nil {
		t.Errorf("UnregCandidate failed --- %v", err)
	} else if prod, err := dpos.GetCandidate(candidate); err != nil || prod != nil {
		t.Errorf("UnregCandidate failed --- %v", err)
	} else if gstate, _ = dpos.GetState(LastBlockHeight); prod2.TotalQuantity.Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("UnvoteCandidate failed mismatch %v %v", prod2.TotalQuantity, gstate.TotalQuantity)
	}

	// activate dpos state
	DefaultConfig.safeSize.Store(uint64(2))
	pmq2 := big.NewInt(0).Mul(DefaultConfig.CandidateMinQuantity, big.NewInt(2))
	DefaultConfig.ActivatedMinQuantity = big.NewInt(0).Add(pmq2, big1)

	err = dpos.RegCandidate(candidate, url, mincandidatestake)
	if err != nil {
		t.Errorf("RegCandidate err %v", err)
	}

	// register again
	err = dpos.RegCandidate(candidate, url, mincandidatestake)
	if err.Error() != "invalid candidate testcandidate(already exist)" {
		t.Errorf("wrong err: %v", err)
	}

	err = dpos.VoteCandidate(voter, candidate, minvoterstake)
	if nil != err {
		t.Errorf("VoterCandidate failed --- %v", err)
	}

	//t.Log(dpos.isdpos())
	_, err = dpos.UnregCandidate(candidate)
	if err.Error() != "already has voter" {
		t.Errorf("wrong err: %v", err)
	}

	_, err = dpos.UnvoteCandidate(voter)
	if err != nil {
		t.Errorf("wrong err: %v", err)
	}
}
