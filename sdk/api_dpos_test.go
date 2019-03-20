package sdk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDposInfo(t *testing.T) {
	Convey("dpos_info", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposInfo()
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposIrreversible(t *testing.T) {
	Convey("dpos_irreversible", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposIrreversible()
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposEpcho(t *testing.T) {
	Convey("dpos_epcho", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposEpcho(0)
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposLatestEpcho(t *testing.T) {
	Convey("dpos_latestEpcho", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposLatestEpcho()
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposValidateEpcho(t *testing.T) {
	Convey("dpos_validateEpcho", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposValidateEpcho()
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposCadidates(t *testing.T) {
	Convey("dpos_cadidates", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposCadidates()
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposAccount(t *testing.T) {
	Convey("dpos_account", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposAccount(systemaccount)
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
