package main

import (
	"crypto/ecdsa"
	"fmt"
	. "github.com/fractalplatform/systemtest/rpc"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"testing"
)

var accountInitAssetAmount = big.NewInt(0).Mul(big.NewInt(100), big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(StakeWeight)))

func getAssetAmountByStake(stake *big.Int) *big.Int {
	return big.NewInt(0).Mul(stake, big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(StakeWeight)))
}

func regAccountAndProducer(namePrefix string, url string, stake *big.Int) (string, *ecdsa.PrivateKey) {
	newAccountName, priKey := registerAccountAndTransfer(namePrefix, 8, 1, accountInitAssetAmount)
	fmt.Println("新账号：(new account:)" + newAccountName)
	So(RegProducer(newAccountName, priKey, url, big.NewInt(0).Mul(stake, big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(StakeWeight)))), ShouldBeNil)

	return newAccountName, priKey
}

func TestRegProducer_411(t *testing.T) {
	Convey("普通账户注册为生产者(Nomal account registers producer)", t, func() {
		regAccountAndProducer("411", "www.xxx.com", big.NewInt(10))
	})
}

func TestRegProducer_412(t *testing.T) {
	Convey("普通账户重复注册为生产者(Nomal account duplictely registers producer)", t, func() {
		newAccountName, priKey := regAccountAndProducer("412", "www.xxx.com", big.NewInt(15))
		So(RegProducer(newAccountName, priKey, "www.xxx.com", big.NewInt(15)), ShouldNotBeNil)
	})
}

func TestRegProducer_413(t *testing.T) {
	Convey("普通账户注册为生产者后注销再次注册(Nomal account registers producer, then log off, then register again)", t, func() {
		newAccountName, priKey := regAccountAndProducer("421", "www.xxx.com", big.NewInt(20))

		So(UnRegProducer(newAccountName, priKey), ShouldBeNil)
		So(RegProducer(newAccountName, priKey, "www.xxx.com", big.NewInt(20)), ShouldBeNil)
	})
}

func TestRegProducer_414(t *testing.T) {
	Convey("普通账户投票后再去注册生产者(Normal account registers producer after voting.)", t, func() {
		producer, _ := regAccountAndProducer("414", "www.xxx.com", big.NewInt(15))
		newAccountName, priKey := registerAccountAndTransfer("414", 8, 1, accountInitAssetAmount)

		So(VoteProducer(newAccountName, priKey, producer, big.NewInt(10)), ShouldBeNil)

		So(RegProducer(newAccountName, priKey, "www.xxx.com", big.NewInt(10)), ShouldNotBeNil)
	})
}

func TestUnRegProducer_421(t *testing.T) {
	Convey("普通账户注册为生产者后注销(Normal account registers producer, then log off.)", t, func() {
		newAccountName, priKey := regAccountAndProducer("421", "www.xxx.com", big.NewInt(10))

		So(UnRegProducer(newAccountName, priKey), ShouldBeNil)
	})
}

func TestUnRegProducer_422(t *testing.T) {
	Convey("普通账户未注册生产者，去注销(Normal account doesn't register producer, then log off.)", t, func() {
		newAccountName, priKey := registerAccountAndTransfer("422", 8, 1, accountInitAssetAmount)

		So(UnRegProducer(newAccountName, priKey), ShouldNotBeNil)
	})
}

func TestUnRegProducer_423(t *testing.T) {
	Convey("普通账户注册生产者后正常注销，再去注销(Normal account registers producer, then log off twice.)", t, func() {
		newAccountName, priKey := regAccountAndProducer("423", "www.xxx.com", big.NewInt(10))

		So(UnRegProducer(newAccountName, priKey), ShouldBeNil)
		So(UnRegProducer(newAccountName, priKey), ShouldNotBeNil)
	})
}

func TestUnRegProducer_424(t *testing.T) {
	Convey("普通账户注册生产者后正常注销，再去注册后，再注销(Normal account registers producer, then log off, then register, then log off.)", t, func() {
		newAccountName, priKey := regAccountAndProducer("424", "www.xxx.com", big.NewInt(10))
		So(UnRegProducer(newAccountName, priKey), ShouldBeNil)
		So(RegProducer(newAccountName, priKey, "www.xxx.com", big.NewInt(10)), ShouldBeNil)
		So(UnRegProducer(newAccountName, priKey), ShouldBeNil)
	})
}

func TestUnRegProducer_431(t *testing.T) {
	Convey("普通账户注册生产者后正常更新信息(Normal account registers producer, then updates the info)", t, func() {
		newAccountName, priKey := regAccountAndProducer("431", "www.xxx.com", big.NewInt(10))
		So(UpdateProducer(newAccountName, priKey, "www.yyy.com", big.NewInt(20)), ShouldBeNil)
	})
}

func TestUnRegProducer_432(t *testing.T) {
	Convey("普通账户未注册生产者去更新信息(Normal account doesn't register producer, then updates the info)", t, func() {
		newAccountName, priKey := registerAccountAndTransfer("432", 8, 1, accountInitAssetAmount)
		So(UpdateProducer(newAccountName, priKey, "www.yyy.com", big.NewInt(20)), ShouldNotBeNil)
	})
}

func TestUnRegProducer_433(t *testing.T) {
	Convey("注销生产者后去更新信息(Normal account loges off producer, then updates the info)", t, func() {
		newAccountName, priKey := regAccountAndProducer("433", "www.xxx.com", big.NewInt(10))
		So(UnRegProducer(newAccountName, priKey), ShouldBeNil)
		So(UpdateProducer(newAccountName, priKey, "www.yyy.com", big.NewInt(20)), ShouldNotBeNil)
	})
}

func TestUnRegProducer_434(t *testing.T) {
	Convey("注销生产者后再次注册然后去更新信息(Normal account loges off producer, then registers producer, then updates the info)", t, func() {
		newAccountName, priKey := regAccountAndProducer("434", "www.xxx.com", big.NewInt(10))
		So(UnRegProducer(newAccountName, priKey), ShouldBeNil)
		So(RegProducer(newAccountName, priKey, "www.xxx.com", big.NewInt(10)), ShouldBeNil)
		So(UpdateProducer(newAccountName, priKey, "www.yyy.com", big.NewInt(20)), ShouldBeNil)
	})
}

func TestChangeProducerKey_435(t *testing.T) {
	Convey("更改生产者私钥(Update producer's private key)", t, func() {
		So(false, ShouldBeTrue)
	})
}

func TestRemoveVoteByProducerK_441(t *testing.T) {
	Convey("取消已投过票的投票者的票(Remove one voters's vote by producer)", t, func() {
		producer, producerPriKey := regAccountAndProducer("441", "www.xxx.com", big.NewInt(20))
		newAccountName, priKey := registerAccountAndTransfer("414", 8, 1, accountInitAssetAmount)

		So(VoteProducer(newAccountName, priKey, producer, big.NewInt(10)), ShouldBeNil)
		So(ProducerRemoveVoter(producer, producerPriKey, newAccountName), ShouldBeNil)
	})
}

func TestRemoveVoteByProducerK_442(t *testing.T) {
	Convey("取消未投过票的投票者的票(Remove an unvoted voter's vote by producer)", t, func() {
		producer, producerPriKey := regAccountAndProducer("441", "www.xxx.com", big.NewInt(18))
		newAccountName, _ := registerAccountAndTransfer("414", 8, 1, accountInitAssetAmount)

		So(ProducerRemoveVoter(producer, producerPriKey, newAccountName), ShouldNotBeNil)
	})
}

func TestSearchProducer_451(t *testing.T) {
	Convey("查询当前所有生产者信息(Get all producers' info)", t, func() {
		regAccountAndProducer("451", "www.xxx.com", big.NewInt(10))
		regAccountAndProducer("451", "www.xxx.com", big.NewInt(15))
		regAccountAndProducer("451", "www.xxx.com", big.NewInt(20))

		So(ValidateAllProducers(), ShouldBeNil)
	})
}

func TestSearchProducer_452(t *testing.T) {
	Convey("查询某生产者信息(Get info of one producer)", t, func() {
		producer, _ := regAccountAndProducer("452", "www.xxx.com", big.NewInt(0).Mul(big.NewInt(100000), big.NewInt(1e18)))
		_, err := GetProducerInfo(producer)
		So(err, ShouldBeNil)
	})
}

func TestSearchProducer_453(t *testing.T) {
	Convey("查询所有active生产者信息(Check all active producers' info)", t, func() {
		So(ValidateActiveProducer(), ShouldBeNil)
	})
}
