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
	minproducerstake = big.NewInt(0).Mul(DefaultConfig.unitStake(), DefaultConfig.ProducerMinQuantity)
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

	// RegProducer
	producer := "testproducer"
	url := "testurl"
	stake := big.NewInt(0).Sub(minproducerstake, big1)

	if err := dpos.UnregProducer(producer); err == nil {
		t.Errorf("UnregProducer should failed --- %v", err)
	}

	err = dpos.RegProducer(producer, url, stake)
	if err == nil {
		t.Errorf("RegProducer should failed --- %v", err)
	}

	if err := dpos.UnregProducer(producer); err == nil {
		t.Errorf("UnregProducer should failed --- %v", err)
	}

	err = dpos.RegProducer(producer, url, minproducerstake)
	if nil != err {
		t.Errorf("RegProducer failed --- %v", err)
	}

	if gstate, err := dpos.GetState(LastBlockHeight); err != nil || gstate.TotalQuantity.Cmp(DefaultConfig.ProducerMinQuantity) != 0 {
		t.Errorf("gstate totalQuantity mismatch --- %v(%v, %v)", err, gstate.TotalQuantity, DefaultConfig.ProducerMinQuantity)
	}

	// GetProducer
	if prod, err := dpos.GetProducer(producer); err != nil {
		t.Errorf("GetProducer failed --- %v", err)
	} else if prod.Name != producer || prod.URL != url || prod.Quantity.Cmp(DefaultConfig.ProducerMinQuantity) != 0 || prod.Quantity.Cmp(prod.TotalQuantity) != 0 {
		t.Errorf("producer info not match")
	}

	// Producers
	if prods, err := dpos.Producers(); err != nil || len(prods) != 1 {
		t.Errorf("producers mismatch")
	}

	// ProducersSize
	if size, err := dpos.ProducersSize(); err != nil || size != 1 {
		t.Errorf("producers mismatch")
	}

	// VoteProducer
	voter := "testvoter"
	vstake := big.NewInt(0).Sub(minvoterstake, big1)
	err = dpos.VoteProducer(voter, producer, vstake)
	if err == nil {
		t.Errorf("VoteProducer should failed --- %v", err)
	}

	err = dpos.VoteProducer(voter, producer, minvoterstake)
	if nil != err {
		t.Errorf("VoterProducer failed --- %v", err)
	}

	prod, _ := dpos.GetProducer(producer)
	gstate, _ := dpos.GetState(LastBlockHeight)
	if prod.TotalQuantity.Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("gstate totalQuantity mismatch --- %v(%v, %v)", err, gstate.TotalQuantity, prod.TotalQuantity)
	} else if new(big.Int).Sub(prod.TotalQuantity, prod.Quantity).Cmp(DefaultConfig.VoterMinQuantity) != 0 {
		t.Errorf("producer totalQuantity mismatch --- %v(%v, %v)", err, prod.TotalQuantity, prod.Quantity)
	}

	// GetVoter
	if vote, err := dpos.GetVoter(voter); err != nil {
		t.Errorf("GetVoter failed --- %v", err)
	} else if vote.Name != voter || vote.Producer != producer || vote.Quantity.Cmp(DefaultConfig.VoterMinQuantity) != 0 {
		t.Errorf("voter info not match --- %v", err)
	}

	// voter cant reg producer
	err = dpos.RegProducer(voter, url, minproducerstake)
	if err.Error() != "invalid producer testvoter(alreay vote to testproducer)" {
		t.Errorf("wrong err type --- %v", err)
	}

	// test change
	producer2 := "testproducer2"
	url2 := "testurl2"
	dpos.RegProducer(producer2, url2, minproducerstake)
	// Producers
	if prods, err := dpos.Producers(); err != nil || len(prods) != 2 {
		t.Errorf("producers mismatch")
	}

	if err := dpos.ChangeProducer(voter, producer2); err != nil {
		t.Errorf("ChangeProducer failed --- %v", err)
	}

	vote, _ := dpos.GetVoter(voter)
	prod, _ = dpos.GetProducer(producer)
	prod2, _ := dpos.GetProducer(producer2)
	gstate, _ = dpos.GetState(LastBlockHeight)

	if vote.Producer != prod2.Name || vote.Quantity.Cmp(DefaultConfig.VoterMinQuantity) != 0 ||
		prod.Quantity.Cmp(prod.TotalQuantity) != 0 || new(big.Int).Add(prod.TotalQuantity, prod2.TotalQuantity).Cmp(gstate.TotalQuantity) != 0 {
		t.Log(prod2.TotalQuantity, gstate.TotalQuantity)
		t.Error("Change stake not work")
	}

	if err := dpos.UnvoteProducer(voter); err != nil {
		t.Errorf("UnvoteProducer failed --- %v", err)
	} else if vote, err := dpos.GetVoter(voter); err != nil || vote != nil {
		t.Errorf("UnvoteProducer failed --- %v", err)
	}
	prod2, _ = dpos.GetProducer(producer2)
	gstate, _ = dpos.GetState(LastBlockHeight)
	if prod.Quantity.Cmp(prod.TotalQuantity) != 0 ||
		prod2.Quantity.Cmp(prod2.TotalQuantity) != 0 || new(big.Int).Add(prod.TotalQuantity, prod2.TotalQuantity).Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("UnvoteProducer failed")
	}

	if err := dpos.UnregProducer(producer); err != nil {
		t.Errorf("UnregProducer failed --- %v", err)
	} else if prod, err := dpos.GetProducer(producer); err != nil || prod != nil {
		t.Errorf("UnregProducer failed --- %v", err)
	} else if gstate, _ = dpos.GetState(LastBlockHeight); prod2.TotalQuantity.Cmp(gstate.TotalQuantity) != 0 {
		t.Errorf("UnvoteProducer failed mismatch %v %v", prod2.TotalQuantity, gstate.TotalQuantity)
	}

	// activate dpos state
	DefaultConfig.safeSize.Store(uint64(2))
	pmq2 := big.NewInt(0).Mul(DefaultConfig.ProducerMinQuantity, big.NewInt(2))
	DefaultConfig.ActivatedMinQuantity = big.NewInt(0).Add(pmq2, big1)

	err = dpos.RegProducer(producer, url, minproducerstake)
	if err != nil {
		t.Errorf("RegProducer err %v", err)
	}

	// register again
	err = dpos.RegProducer(producer, url, minproducerstake)
	if err.Error() != "invalid producer testproducer(already exist)" {
		t.Errorf("wrong err: %v", err)
	}

	err = dpos.VoteProducer(voter, producer, minvoterstake)
	if nil != err {
		t.Errorf("VoterProducer failed --- %v", err)
	}

	//t.Log(dpos.isdpos())
	err = dpos.UnregProducer(producer)
	if err.Error() != "already has voter" {
		t.Errorf("wrong err: %v", err)
	}

	err = dpos.UnvoteProducer(voter)
	if err != nil {
		t.Errorf("wrong err: %v", err)
	}
}
