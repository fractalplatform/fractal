package common

import (
	"math"
	"math/big"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	tURL   = "www.test.com"
	tState = new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18))
)

func TestProducerReg(t *testing.T) {
	SkipConvey("types.RegProducer", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.RegProducer(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas, tURL, tState)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestProducerUpdate(t *testing.T) {
	SkipConvey("types.UpdateProducer", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.UpdateProducer(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas, tURL, tState)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestProducerUnreg(t *testing.T) {
	SkipConvey("types.UnregProducer", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.UnRegProducer(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestVoterVote(t *testing.T) {
	SkipConvey("types.VoteProducer", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.VoteProducer(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas, systemaccount, tState)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestVoterChangeVote(t *testing.T) {
	SkipConvey("types.ChangeProducer", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.ChangeProducer(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas, systemaccount)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestVoterUnvote(t *testing.T) {
	SkipConvey("types.UnvoteProducer", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.UnvoteProducer(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestProducerRemoveVoter(t *testing.T) {
	SkipConvey("types.RemoveVoter", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.UnvoteVoter(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas, systemaccount)
		_ = hash
		So(err, ShouldBeNil)
	})
}
