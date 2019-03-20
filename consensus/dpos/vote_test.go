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
	mincadidatestake = big.NewInt(0).Mul(DefaultConfig.unitStake(), DefaultConfig.CadidateMinQuantity)
	minvoterstake    = big.NewInt(0).Mul(DefaultConfig.unitStake(), DefaultConfig.VoterMinQuantity)
	big1             = big.NewInt(1)
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

	// RegCadidate
	cadidate := "testcadidate"
	url := "testurl"
	stake := big.NewInt(0).Sub(mincadidatestake, big1)

	if err := dpos.UnregCadidate(cadidate); err == nil {
		t.Errorf("UnregCadidate should failed --- %v", err)
	}

	err = dpos.RegCadidate(cadidate, url, stake)
	if err == nil {
		t.Errorf("RegCadidate should failed --- %v", err)
	}

	if err := dpos.UnregCadidate(cadidate); err == nil {
		t.Errorf("UnregCadidate should failed --- %v", err)
	}

	err = dpos.RegCadidate(cadidate, url, mincadidatestake)
	if nil != err {
		t.Errorf("RegCadidate failed --- %v", err)
	}

	if gstate, err := dpos.GetState(LastBlockHeight); err != nil || gstate.TotalQuantity.Cmp(DefaultConfig.CadidateMinQuantity) != 0 {
		t.Errorf("gstate totalQuantity mismatch --- %v(%v, %v)", err, gstate.TotalQuantity, DefaultConfig.CadidateMinQuantity)
	}

	// GetCadidate
	if prod, err := dpos.GetCadidate(cadidate); err != nil {
		t.Errorf("GetCadidate failed --- %v", err)
	} else if prod.Name != cadidate || prod.URL != url || prod.Quantity.Cmp(DefaultConfig.CadidateMinQuantity) != 0 || prod.Quantity.Cmp(prod.TotalQuantity) != 0 {
		t.Errorf("cadidate info not match")
	}

	// Cadidates
	if prods, err := dpos.Cadidates(); err != nil || len(prods) != 1 {
		t.Errorf("cadidates mismatch")
	}

	// CadidatesSize
	if size, err := dpos.CadidatesSize(); err != nil || size != 1 {
		t.Errorf("cadidates mismatch")
	}

	// VoteCadidate
	voter := "testvoter"
	vstake := big.NewInt(0).Sub(minvoterstake, big1)
	err = dpos.VoteCadidate(voter, cadidate, vstake)
	if err == nil {
		t.Errorf("VoteCadidate should failed --- %v", err)
	}

	err = dpos.VoteCadidate(voter, cadidate, minvoterstake)
	if nil != err {
		t.Errorf("VoterCadidate failed --- %v", err)
	}

	prod, _ := dpos.GetCadidate(cadidate)
	gstate, _ := dpos.GetState(LastBlockHeight)
	if prod.TotalQuantity.Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("gstate totalQuantity mismatch --- %v(%v, %v)", err, gstate.TotalQuantity, prod.TotalQuantity)
	} else if new(big.Int).Sub(prod.TotalQuantity, prod.Quantity).Cmp(DefaultConfig.VoterMinQuantity) != 0 {
		t.Errorf("cadidate totalQuantity mismatch --- %v(%v, %v)", err, prod.TotalQuantity, prod.Quantity)
	}

	// GetVoter
	if vote, err := dpos.GetVoter(voter); err != nil {
		t.Errorf("GetVoter failed --- %v", err)
	} else if vote.Name != voter || vote.Cadidate != cadidate || vote.Quantity.Cmp(DefaultConfig.VoterMinQuantity) != 0 {
		t.Errorf("voter info not match --- %v", err)
	}

	// voter cant reg cadidate
	err = dpos.RegCadidate(voter, url, mincadidatestake)
	if err.Error() != "invalid cadidate testvoter(alreay vote to testcadidate)" {
		t.Errorf("wrong err type --- %v", err)
	}

	// test change
	cadidate2 := "testcadidate2"
	url2 := "testurl2"
	dpos.RegCadidate(cadidate2, url2, mincadidatestake)
	// Cadidates
	if prods, err := dpos.Cadidates(); err != nil || len(prods) != 2 {
		t.Errorf("cadidates mismatch")
	}

	if err := dpos.ChangeCadidate(voter, cadidate2); err != nil {
		t.Errorf("ChangeCadidate failed --- %v", err)
	}

	vote, _ := dpos.GetVoter(voter)
	prod, _ = dpos.GetCadidate(cadidate)
	prod2, _ := dpos.GetCadidate(cadidate2)
	gstate, _ = dpos.GetState(LastBlockHeight)

	if vote.Cadidate != prod2.Name || vote.Quantity.Cmp(DefaultConfig.VoterMinQuantity) != 0 ||
		prod.Quantity.Cmp(prod.TotalQuantity) != 0 || new(big.Int).Add(prod.TotalQuantity, prod2.TotalQuantity).Cmp(gstate.TotalQuantity) != 0 {
		t.Log(prod2.TotalQuantity, gstate.TotalQuantity)
		t.Error("Change stake not work")
	}

	if err := dpos.UnvoteCadidate(voter); err != nil {
		t.Errorf("UnvoteCadidate failed --- %v", err)
	} else if vote, err := dpos.GetVoter(voter); err != nil || vote != nil {
		t.Errorf("UnvoteCadidate failed --- %v", err)
	}
	prod2, _ = dpos.GetCadidate(cadidate2)
	gstate, _ = dpos.GetState(LastBlockHeight)
	if prod.Quantity.Cmp(prod.TotalQuantity) != 0 ||
		prod2.Quantity.Cmp(prod2.TotalQuantity) != 0 || new(big.Int).Add(prod.TotalQuantity, prod2.TotalQuantity).Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("UnvoteCadidate failed")
	}

	if err := dpos.UnregCadidate(cadidate); err != nil {
		t.Errorf("UnregCadidate failed --- %v", err)
	} else if prod, err := dpos.GetCadidate(cadidate); err != nil || prod != nil {
		t.Errorf("UnregCadidate failed --- %v", err)
	} else if gstate, _ = dpos.GetState(LastBlockHeight); prod2.TotalQuantity.Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("UnvoteCadidate failed mismatch %v %v", prod2.TotalQuantity, gstate.TotalQuantity)
	}

	// activate dpos state
	DefaultConfig.safeSize.Store(uint64(2))
	pmq2 := big.NewInt(0).Mul(DefaultConfig.CadidateMinQuantity, big.NewInt(2))
	DefaultConfig.ActivatedMinQuantity = big.NewInt(0).Add(pmq2, big1)

	err = dpos.RegCadidate(cadidate, url, mincadidatestake)
	if err != nil {
		t.Errorf("RegCadidate err %v", err)
	}

	// register again
	err = dpos.RegCadidate(cadidate, url, mincadidatestake)
	if err.Error() != "invalid cadidate testcadidate(already exist)" {
		t.Errorf("wrong err: %v", err)
	}

	err = dpos.VoteCadidate(voter, cadidate, minvoterstake)
	if nil != err {
		t.Errorf("VoterCadidate failed --- %v", err)
	}

	//t.Log(dpos.isdpos())
	err = dpos.UnregCadidate(cadidate)
	if err.Error() != "already has voter" {
		t.Errorf("wrong err: %v", err)
	}

	err = dpos.UnvoteCadidate(voter)
	if err != nil {
		t.Errorf("wrong err: %v", err)
	}
}
