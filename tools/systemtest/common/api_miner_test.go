package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMinerStart(t *testing.T) {
	Convey("miner_start", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerStart()
		_ = ret
		So(err, ShouldBeNil)
	})
}

func TestMinerStop(t *testing.T) {
	Convey("miner_stop", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerStop()
		_ = ret
		So(err, ShouldBeNil)
	})
}

// MinerMining mining
func TestMinerMining(t *testing.T) {
	Convey("miner_mining", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerMining()
		_ = ret
		So(err, ShouldBeNil)
	})
}

// MinerSetExtra extra
func TestMinerSetExtra(t *testing.T) {
	Convey("miner_mining", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerSetExtra([]byte("testextra"))
		_ = ret
		So(err, ShouldBeNil)
	})
}

// MinerSetCoinbase coinbase
func TestMinerSetCoinbase(t *testing.T) {
	Convey("miner_mining", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerSetCoinbase(systemaccount, systemprivkey)
		_ = ret
		So(err, ShouldBeNil)
	})
}
