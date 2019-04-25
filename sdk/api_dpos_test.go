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
func TestDposCandidate(t *testing.T) {
	Convey("dpos_candidate", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposCandidate(systemaccount)
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposCandidates(t *testing.T) {
	Convey("dpos_candidates", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposCandidates(true)
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
func TestDposVotersByVoter(t *testing.T) {
	Convey("dpos_votersByVoter", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposVotersByVoter(systemaccount, true)
		So(err, ShouldBeNil)
		_ = info
		//So(info, ShouldNotBeEmpty)
	})
}
func TestDposVotersByCandidate(t *testing.T) {
	Convey("dpos_votersByCandidate", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposVotersByCandidate(systemaccount, true)
		So(err, ShouldBeNil)
		_ = info
		//So(info, ShouldNotBeEmpty)
	})
}
func TestDposAvailableStake(t *testing.T) {
	SkipConvey("dpos_availableStake", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposAvailableStake(systemaccount)
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}

func TestDposValidCandidates(t *testing.T) {
	Convey("dpos_validCandidates", t, func() {
		api := NewAPI(rpchost)
		info, err := api.DposValidCandidates()
		So(err, ShouldBeNil)
		So(info, ShouldNotBeEmpty)
	})
}
