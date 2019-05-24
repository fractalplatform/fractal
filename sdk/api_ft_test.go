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

	"github.com/fractalplatform/fractal/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCurrentBlock(t *testing.T) {
	Convey("ft_getCurrentBlock", t, func() {
		block, err := api.GetCurrentBlock(false)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
	Convey("ft_getCurrentBlock", t, func() {
		block, err := api.GetCurrentBlock(true)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
}

func TestGetBlockByHash(t *testing.T) {
	Convey("ft_getBlockByHash", t, func() {
		block, err := api.GetCurrentBlock(false)
		So(err, ShouldBeNil)
		hash := common.HexToHash(block["hash"].(string))
		block, err = api.GetBlockByHash(hash, false)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
}

func TestGetBlockByNumber(t *testing.T) {
	Convey("ft_getBlockByNumber", t, func() {
		block, err := api.GetCurrentBlock(false)
		So(err, ShouldBeNil)
		block, err = api.GetBlockByNumber((int64(block["number"].(float64))), false)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
}
func TestGetTransactionByHash(t *testing.T) {
	Convey("ft_getTransactionByHash", t, func() {
		block, err := api.GetCurrentBlock(false)
		So(err, ShouldBeNil)
		for _, hash := range block["transactions"].([]interface{}) {
			tx, err := api.GetTransactionByHash(common.HexToHash(hash.(string)))
			So(err, ShouldBeNil)
			So(tx, ShouldNotBeNil)
		}
	})
}
func TestGetTransactionReceiptByHash(t *testing.T) {
	Convey("ft_getTransactionReceipt", t, func() {
		block, err := api.GetCurrentBlock(false)
		So(err, ShouldBeNil)
		for _, hash := range block["transactions"].([]interface{}) {
			receipt, err := api.GetTransactionReceiptByHash(common.HexToHash(hash.(string)))
			So(err, ShouldBeNil)
			So(receipt, ShouldNotBeNil)
		}

	})
}
func TestFTGasPrice(t *testing.T) {
	Convey("ft_gasPrice", t, func() {
		gasprice, err := api.GasPrice()
		So(err, ShouldBeNil)
		So(gasprice, ShouldNotBeNil)
	})
}

func TestGetChainConfig(t *testing.T) {
	Convey("ft_chainConfig", t, func() {
		cfg, err := api.GetChainConfig()
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
	})
}

// func TestGetGenesis(t *testing.T) {
// 	Convey("ft_chainConfig", t, func() {
// 		api := NewAPI(rpchost)
// 		cfg, err := api.GetGenesis()
// 		So(err, ShouldBeNil)
// 		So(cfg, ShouldNotBeNil)
// 	})
// }
