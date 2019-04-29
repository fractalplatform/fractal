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
	"fmt"
	"testing"

	"github.com/fractalplatform/fractal/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCurrentBlock(t *testing.T) {
	Convey("ft_getCurrentBlock", t, func() {
		api := NewAPI(rpchost)
		block, err := api.CurrentBlock(false)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
}

func TestBlockByHash(t *testing.T) {
	Convey("ft_getBlockByHash", t, func() {
		api := NewAPI(rpchost)
		block, err := api.CurrentBlock(false)
		So(err, ShouldBeNil)
		hash := common.HexToHash(block["hash"].(string))
		block, err = api.BlockByHash(hash, false)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
}

func TestBlockByNumber(t *testing.T) {
	Convey("ft_getBlockByNumber", t, func() {
		api := NewAPI(rpchost)
		block, err := api.CurrentBlock(false)
		So(err, ShouldBeNil)
		block, err = api.BlockByNumber((int64(block["number"].(float64))), false)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
}
func TestTransactionByHash(t *testing.T) {
	Convey("ft_getTransactionByHash", t, func() {
		hash := common.HexToHash("")
		api := NewAPI(rpchost)
		tx, err := api.TransactionByHash(hash)
		So(err, ShouldBeNil)
		So(tx, ShouldNotBeNil)
	})
}
func TestTransactionReceiptByHash(t *testing.T) {
	Convey("ft_getTransactionReceipt", t, func() {
		hash := common.HexToHash("")
		api := NewAPI(rpchost)
		receipt, err := api.TransactionReceiptByHash(hash)
		So(err, ShouldBeNil)
		So(receipt, ShouldNotBeNil)
	})
}
func TestGasPrice(t *testing.T) {
	Convey("ft_gasPrice", t, func() {
		api := NewAPI(rpchost)
		gasprice, err := api.GasPrice()
		So(err, ShouldBeNil)
		So(gasprice, ShouldNotBeNil)
	})
}

func TestChainConfig(t *testing.T) {
	Convey("ft_chainConfig", t, func() {
		api := NewAPI(rpchost)
		cfg, err := api.ChainConfig()
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		fmt.Printf("%v", cfg)
	})
}
