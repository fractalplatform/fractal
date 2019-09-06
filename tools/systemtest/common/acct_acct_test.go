package common

import (
	"crypto/ecdsa"
	"math"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	tAccountName common.Name
	tPrivKey     *ecdsa.PrivateKey
	tValue       = big.NewInt(10)
	tGas         = uint64(90000)
)

func TestTransfer(t *testing.T) {
	SkipConvey("types.Transfer", t, func() {
		priv, _ := crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(NewAPI(rpchost), common.StrToName(systemaccount), priv, systemassetid, math.MaxUint64, true, chainid)
		hash, err := sysAcct.Transfer(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas)
		_ = hash
		So(err, ShouldBeNil)
	})
}

func TestAccountCreate(t *testing.T) {
	SkipConvey("types.CreateAccount", t, func() {
		priv, _ := crypto.HexToECDSA(systemprivkey)
		sysAcct := NewAccount(NewAPI(rpchost), common.StrToName(systemaccount), priv, systemassetid, math.MaxUint64, true, chainid)
		priv, pub := GenerateKey()
		accountName := common.StrToName(GenerateAccountName("test", 15))
		hash, err := sysAcct.CreateAccount(accountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas, pub)
		_ = hash
		So(err, ShouldBeNil)
		tPrivKey = priv
		tAccountName = accountName
	})
}

func TestAccountUpdate(t *testing.T) {
	SkipConvey("types.UpdateAccount", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		priv, _ := GenerateKey()
		hash, err := acct.UpdateAccount(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas, priv)
		_ = hash
		So(err, ShouldBeNil)
		tPrivKey = priv
	})
}

func TestAccountDelete(t *testing.T) {
	SkipConvey("types.DeleteAccount", t, func() {
		acct := NewAccount(NewAPI(rpchost), tAccountName, tPrivKey, systemassetid, math.MaxUint64, true, chainid)
		hash, err := acct.DeleteAccount(tAccountName, new(big.Int).Mul(tValue, big.NewInt(1e18)), systemassetid, tGas)
		_ = hash
		So(err, ShouldBeNil)
	})
}
