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

	"github.com/fractalplatform/fractal/params"
	. "github.com/smartystreets/goconvey/convey"
)

var feeaccount = params.DefaultChainconfig.FeeName

func TestFeeInfo(t *testing.T) {
	Convey("fee_getObjectFeeByName", t, func() {
		objFee, err := api.FeeInfo(feeaccount, 1)
		So(err, ShouldBeNil)
		So(objFee, ShouldNotBeNil)
	})
}

func TestFeeInfoByID(t *testing.T) {
	Convey("fee_getObjectFeeResult", t, func() {
		objFee, err := api.FeeInfoByID(1, 1, 0)
		So(err, ShouldBeNil)
		So(objFee, ShouldNotBeNil)
	})
}
