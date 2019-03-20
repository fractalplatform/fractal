package sdk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMinerMining(t *testing.T) {
	Convey("miner_mining", t, func() {
		api := NewAPI(rpchost)
		_, err := api.MinerMining()
		So(err, ShouldBeNil)
	})
}
func TestMinerStart(t *testing.T) {
	Convey("miner_start", t, func() {
		api := NewAPI(rpchost)
		mining, _ := api.MinerMining()
		ret, err := api.MinerStart()
		So(err, ShouldBeNil)
		if mining {
			So(ret, ShouldBeFalse)
		} else {
			So(ret, ShouldBeTrue)
		}
	})
}
func TestMinerStop(t *testing.T) {
	Convey("miner_stop", t, func() {
		api := NewAPI(rpchost)
		mining, _ := api.MinerMining()
		ret, err := api.MinerStop()
		So(err, ShouldBeNil)
		if mining {
			So(ret, ShouldBeTrue)
		} else {
			So(ret, ShouldBeFalse)
		}
	})
}
func TestMinerSetExtra(t *testing.T) {
	Convey("miner_setExtra", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerSetExtra([]byte("testextra"))
		So(err, ShouldBeNil)
		So(ret, ShouldBeTrue)
	})
}
func TestMinerSetCoinbase(t *testing.T) {
	Convey("miner_setCoinbase", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerSetCoinbase(systemaccount, []string{systemprivkey})
		So(err, ShouldBeNil)
		So(ret, ShouldBeTrue)
	})
}
func TestMinerSetExtraTooLong(t *testing.T) {
	Convey("miner_setExtra", t, func() {
		api := NewAPI(rpchost)
		ret, err := api.MinerSetExtra([]byte("testextratestextratestextratestextratestextratestextra"))
		So(err, ShouldNotBeNil)
		So(ret, ShouldBeTrue)
	})
}
