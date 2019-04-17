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

import "time"

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

	Lifetime time.Duration `mapstructure:"lifetime"` // Maximum amount of time non-executable transaction are queued

	GasAssetID uint64
}

func (c *Config) check() Config {
	conf := *c
	//todo check config
	return conf
}
