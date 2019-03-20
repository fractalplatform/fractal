package sdk

import (
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
		block, _ := api.CurrentBlock(false)
		hash := common.HexToHash(block["hash"].(string))
		block, err := api.BlockByHash(hash, false)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
	})
}

func TestBlockByNumber(t *testing.T) {
	Convey("ft_getBlockByNumber", t, func() {
		api := NewAPI(rpchost)
		block, _ := api.CurrentBlock(false)
		block, err := api.BlockByNumber((int64(block["number"].(float64))), false)
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
