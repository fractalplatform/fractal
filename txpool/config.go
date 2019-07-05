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

package txpool

import (
	"time"

	"github.com/ethereum/go-ethereum/log"
)

// Config  are the configuration parameters of the transaction pool.
type Config struct {
	NoLocals  bool          `mapstructure:"nolocals"`  // Whether local transaction handling should be disabled
	Journal   string        `mapstructure:"journal"`   // Journal of local transactions to survive node restarts
	Rejournal time.Duration `mapstructure:"rejournal"` // Time interval to regenerate the local transaction journal

	PriceLimit uint64 `mapstructure:"pricelimit"` // Minimum gas price to enforce for acceptance into the pool
	PriceBump  uint64 `mapstructure:"pricebump"`  // Minimum price bump percentage to replace an already existing transaction (nonce)

	AccountSlots uint64 `mapstructure:"accountslots"` // Minimum number of executable transaction slots guaranteed per account
	GlobalSlots  uint64 `mapstructure:"globalslots"`  // Maximum number of executable transaction slots for all accounts
	AccountQueue uint64 `mapstructure:"accountqueue"` // Maximum number of non-executable transaction slots permitted per account
	GlobalQueue  uint64 `mapstructure:"globalqueue"`  // Maximum number of non-executable transaction slots for all accounts

	Lifetime   time.Duration `mapstructure:"lifetime"`   // Maximum amount of time non-executable transaction are queued
	ResendTime time.Duration `mapstructure:"resendtime"` // Maximum amount of time  executable transaction are resended

	MinBroadcast   uint64 `mapstructure:"minbroadcast"`   // Minimum number of nodes for the transaction broadcast
	RatioBroadcast uint64 `mapstructure:"ratiobroadcast"` // Ratio of nodes for the transaction broadcast
	GasAssetID     uint64
}

// DefaultTxPoolConfig default txpool config
var DefaultTxPoolConfig = &Config{
	Journal:        "transactions.rlp",
	Rejournal:      time.Hour,
	PriceLimit:     1000000000,
	PriceBump:      10,
	AccountSlots:   128,
	GlobalSlots:    4096,
	AccountQueue:   1280,
	GlobalQueue:    4096,
	Lifetime:       3 * time.Hour,
	ResendTime:     10 * time.Minute,
	MinBroadcast:   3,
	RatioBroadcast: 3,
}

// check checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *Config) check() Config {
	conf := *config
	if conf.Rejournal < time.Second {
		log.Warn("Sanitizing invalid txpool journal time", "provided", conf.Rejournal, "updated", time.Second)
		conf.Rejournal = time.Second
	}
	if conf.PriceLimit < 1 {
		log.Warn("Sanitizing invalid txpool price limit", "provided", conf.PriceLimit, "updated", DefaultTxPoolConfig.PriceLimit)
		conf.PriceLimit = DefaultTxPoolConfig.PriceLimit
	}
	if conf.PriceBump < 1 {
		log.Warn("Sanitizing invalid txpool price bump", "provided", conf.PriceBump, "updated", DefaultTxPoolConfig.PriceBump)
		conf.PriceBump = DefaultTxPoolConfig.PriceBump
	}
	if conf.AccountSlots < 1 {
		log.Warn("Sanitizing invalid txpool account slots", "provided", conf.AccountSlots, "updated", DefaultTxPoolConfig.AccountSlots)
		conf.AccountSlots = DefaultTxPoolConfig.AccountSlots
	}
	if conf.GlobalSlots < 1 {
		log.Warn("Sanitizing invalid txpool global slots", "provided", conf.GlobalSlots, "updated", DefaultTxPoolConfig.GlobalSlots)
		conf.GlobalSlots = DefaultTxPoolConfig.GlobalSlots
	}
	if conf.AccountQueue < 1 {
		log.Warn("Sanitizing invalid txpool account queue", "provided", conf.AccountQueue, "updated", DefaultTxPoolConfig.AccountQueue)
		conf.AccountQueue = DefaultTxPoolConfig.AccountQueue
	}
	if conf.GlobalQueue < 1 {
		log.Warn("Sanitizing invalid txpool global queue", "provided", conf.GlobalQueue, "updated", DefaultTxPoolConfig.GlobalQueue)
		conf.GlobalQueue = DefaultTxPoolConfig.GlobalQueue
	}
	if conf.Lifetime < 1 {
		log.Warn("Sanitizing invalid txpool lifetime", "provided", conf.Lifetime, "updated", DefaultTxPoolConfig.Lifetime)
		conf.Lifetime = DefaultTxPoolConfig.Lifetime
	}
	if conf.ResendTime < 1 {
		log.Warn("Sanitizing invalid txpool resendtime", "provided", conf.ResendTime, "updated", DefaultTxPoolConfig.ResendTime)
		conf.ResendTime = DefaultTxPoolConfig.ResendTime
	}
	if conf.RatioBroadcast < 1 {
		log.Warn("Sanitizing invalid txpool ratiobroadcast", "provided", conf.RatioBroadcast, "updated", DefaultTxPoolConfig.RatioBroadcast)
		conf.RatioBroadcast = DefaultTxPoolConfig.RatioBroadcast
	}
	return conf
}
