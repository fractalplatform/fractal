package main

import (
	. "github.com/fractalplatform/systemtest/rpc"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"testing"
)

func TestVoteProducer_511(t *testing.T) {
	Convey("投票给已注册的生产者(Vote to an producer.)", t, func() {
		producer, _ := regAccountAndProducer("511", "www.xxx.com", big.NewInt(10))

		newAccountName, priKey := registerAccountAndTransfer("511", 8, 1, accountInitAssetAmount)

		So(VoteProducer(newAccountName, priKey, producer, big.NewInt(10)), ShouldBeNil)
	})
}

func TestVoteProducer_512(t *testing.T) {
	Convey("投票给未注册生产者的账号(Vote to an account which hasn't been registered as producer.)", t, func() {
		accountName, _ := registerAccount("512", 1)

		newAccountName, priKey := registerAccountAndTransfer("512", 8, 1, accountInitAssetAmount)

		So(VoteProducer(newAccountName, priKey, accountName, big.NewInt(10)), ShouldNotBeNil)
	})
}

func TestVoteProducer_513(t *testing.T) {
	Convey("投票给注册后又注销的生产者(Vote to an producer which has been logged off.)", t, func() {
		producer, producerPriKey := regAccountAndProducer("513", "www.xxx.com", big.NewInt(10))

		So(UnRegProducer(producer, producerPriKey), ShouldBeNil)

		newAccountName, priKey := registerAccountAndTransfer("513", 8, 1, accountInitAssetAmount)

		So(VoteProducer(newAccountName, priKey, producer, big.NewInt(10)), ShouldNotBeNil)

	})
}

func TestVoteProducer_514(t *testing.T) {
	Convey("注册为生产者后，投票给其它生产者(Register as producer, then vote to other producers.)", t, func() {
		producer, producerPriKey := regAccountAndProducer("514", "www.xxx.com", big.NewInt(10))
		newProducer, _ := regAccountAndProducer("514", "www.xxx.com", big.NewInt(10))

		So(VoteProducer(producer, producerPriKey, newProducer, big.NewInt(10)), ShouldNotBeNil)
	})
}

func TestVoteProducer_515(t *testing.T) {
	Convey("注册为生产者后，再注销，再投票给其它生产者(Register as producer, then log off, then vote to other producers.)", t, func() {
		producer, producerPriKey := regAccountAndProducer("515", "www.xxx.com", big.NewInt(10))
		So(UnRegProducer(producer, producerPriKey), ShouldBeNil)

		newProducer, _ := regAccountAndProducer("515", "www.xxx.com", big.NewInt(10))

		So(VoteProducer(producer, producerPriKey, newProducer, big.NewInt(10)), ShouldBeNil)
	})
}

func TestChangeVote_521(t *testing.T) {
	Convey("将投票转投给其它注册的生产者(Switch the vote to other producers.)", t, func() {
		producer, producerPriKey := regAccountAndProducer("515", "www.xxx.com", big.NewInt(10))
		So(UnRegProducer(producer, producerPriKey), ShouldBeNil)

		newProducer, _ := regAccountAndProducer("515", "www.xxx.com", big.NewInt(10))

		So(VoteProducer(producer, producerPriKey, newProducer, big.NewInt(10)), ShouldBeNil)
	})
}
