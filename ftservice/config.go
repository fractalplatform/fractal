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

package ftservice

import (
	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/blockchain"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/ftservice/gasprice"
	"github.com/fractalplatform/fractal/metrics"
	"github.com/fractalplatform/fractal/txpool"
)

// Config ftservice config
type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the main net block is used.
	Genesis *blockchain.Genesis `toml:",omitempty"`

	// Database options
	SkipBcVersionCheck bool `mapstructure:"ftservice-skipvcversioncheck"`
	DatabaseHandles    int  `mapstructure:"ftservice-databasehandles"`
	DatabaseCache      int  `mapstructure:"ftservice-databasecache"`

	// Transaction pool options
	TxPool *txpool.Config

	// Gas Price Oracle options
	GasPrice gasprice.Config

	// miner
	Miner *MinerConfig

	CoinBase    common.Address
	MetricsConf *metrics.Config

	// snapshot
	Snapshot        bool
	ContractLogFlag bool `mapstructure:"ftservice-ContractLogFlag"`
	AccountNameConf *am.Config
	AssetNameConf   *asset.Config
}

// MinerConfig miner config
type MinerConfig struct {
	Start       bool     `mapstructure:"miner-start"`
	Name        string   `mapstructure:"miner-name"`
	PrivateKeys []string `mapstructure:"miner-private"`
	ExtraData   string   `mapstructure:"miner-extra"`
}
