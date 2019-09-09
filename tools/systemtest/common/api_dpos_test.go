package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDposInfo(t *testing.T) {
	Convey("dpos_info", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposInfo()
		_ = info
		So(err, ShouldBeNil)
	})
}

func TestDposIrreversible(t *testing.T) {
	Convey("dpos_irreversible", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposIrreversible()
		_ = info
		So(err, ShouldBeNil)
	})
}

func TestDposEpcho(t *testing.T) {
	Convey("dpos_epcho", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposEpcho(0)
		_ = info
		So(err, ShouldBeNil)
	})
}

func TestDposLatestEpcho(t *testing.T) {
	Convey("dpos_latestEpcho", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposLatestEpcho()
		_ = info
		So(err, ShouldBeNil)
	})
}

func TestDposValidateEpcho(t *testing.T) {
	Convey("dpos_validateEpcho", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposValidateEpcho()
		_ = info
		So(err, ShouldBeNil)
	})
}

func TestDposProducers(t *testing.T) {
	Convey("dpos_producers", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposProducers()
		_ = info
		So(err, ShouldBeNil)
	})
}

func TestDposAccount(t *testing.T) {
	Convey("dpos_account", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposAccount(systemaccount)
		_ = info
		So(err, ShouldBeNil)
	})
}
