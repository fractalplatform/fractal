package common

import (
	"math"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	tAssetName   = "testasset"
	tAssetSymbol = "tat"
	tAmount      = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e8))
	tDecimals    = uint64(8)
	tAssetID     uint64
	chainid      = big.NewInt(1)
)

func TestAssetIssue(t *testing.T) {
	SkipConvey("types.IssueAsset", t, func() {
		priv, _ := crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(NewAPI(rpchost), common.StrToName(systemaccount), priv, systemassetid, math.MaxUint64, true, chainid)
		hash, err := sysAcct.IssueAsset(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e8)), systemassetid, tGas, tAssetName, tAssetSymbol, tAmount, tDecimals, common.StrToName(systemaccount))
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestAssetIncrease(t *testing.T) {
	SkipConvey("types.IncreaseAsset", t, func() {
		priv, _ := crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(NewAPI(rpchost), common.StrToName(systemaccount), priv, systemassetid, math.MaxUint64, true, chainid)
		hash, err := sysAcct.IncreaseAsset(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e8)), systemassetid, tGas, tAssetID, tAmount)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestAssetOwner(t *testing.T) {
	SkipConvey("types.SetAssetOwner", t, func() {
		priv, _ := crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(NewAPI(rpchost), common.StrToName(systemaccount), priv, systemassetid, math.MaxUint64, true, chainid)
		hash, err := sysAcct.SetAssetOwner(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e8)), systemassetid, tGas, tAssetID, tAccountName)
		_ = hash
		So(err, ShouldBeNil)
	})
}
