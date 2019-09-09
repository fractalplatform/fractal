package common

import (
	"testing"

	"github.com/fractalplatform/fractal/common"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCurrentBlock(t *testing.T) {
	Convey("ft_getCurrentBlock", t, func() {
		api := NewAPI(rpchost)
		block, err := api.CurrentBlock(false)
		_ = block
		So(err, ShouldBeNil)
	})
}

func TestBlockByHash(t *testing.T) {
	Convey("ft_getBlockByHash", t, func() {
		hash := common.HexToHash("")
		api := NewAPI(rpchost)
		block, err := api.BlockByHash(hash, false)
		_ = block
		So(err, ShouldBeNil)
	})
}

func TestBlockByNumber(t *testing.T) {
	Convey("ft_getBlockByNumber", t, func() {
		api := NewAPI(rpchost)
		block, err := api.BlockByNumber(0, false)
		_ = block
		So(err, ShouldBeNil)
	})
}

func TestTransactionByHash(t *testing.T) {
	Convey("ft_getTransactionByHash", t, func() {
		hash := common.HexToHash("")
		api := NewAPI(rpchost)
		tx, err := api.TransactionByHash(hash)
		_ = tx
		So(err, ShouldBeNil)
	})
}

func TestTransactionReceiptByHash(t *testing.T) {
	Convey("ft_getTransactionReceipt", t, func() {
		hash := common.HexToHash("")
		api := NewAPI(rpchost)
		receipt, err := api.TransactionReceiptByHash(hash)
		_ = receipt
		So(err, ShouldBeNil)
	})
}

func TestGasPrice(t *testing.T) {
	Convey("ft_getTransactionReceipt", t, func() {
		api := NewAPI(rpchost)
		gasprice, err := api.GasPrice()
		_ = gasprice
		So(err, ShouldBeNil)
	})
}
