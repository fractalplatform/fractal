// Copyright 2018 The OEX Team Authors
// This file is part of the OEX project.
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

package main

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/oexplatform/oexchain/cmd/utils"
	"github.com/oexplatform/oexchain/debug"
	"github.com/oexplatform/oexchain/metrics"
	"github.com/oexplatform/oexchain/node"
	"github.com/oexplatform/oexchain/oexservice"
	"github.com/oexplatform/oexchain/oexservice/gasprice"
	"github.com/oexplatform/oexchain/p2p"
	"github.com/oexplatform/oexchain/params"
	"github.com/oexplatform/oexchain/txpool"
)

var (
	//oex config instance
	ftCfgInstance = defaultFtConfig()
	ipcEndpoint   string
)

type ftConfig struct {
	GenesisFile  string             `mapstructure:"genesis"`
	DebugCfg     *debug.Config      `mapstructure:"debug"`
	LogCfg       *utils.LogConfig   `mapstructure:"log"`
	NodeCfg      *node.Config       `mapstructure:"node"`
	FtServiceCfg *oexservice.Config `mapstructure:"oexservice"`
}

func defaultFtConfig() *ftConfig {
	return &ftConfig{
		DebugCfg:     debug.DefaultConfig(),
		LogCfg:       utils.DefaultLogConfig(),
		NodeCfg:      defaultNodeConfig(),
		FtServiceCfg: defaultFtServiceConfig(),
	}
}

func defaultNodeConfig() *node.Config {
	return &node.Config{
		Name:             params.ClientIdentifier,
		DataDir:          defaultDataDir(),
		IPCPath:          params.ClientIdentifier + ".ipc",
		HTTPHost:         "localhost",
		HTTPPort:         8545,
		HTTPModules:      []string{"oex", "dpos", "fee", "account"},
		HTTPVirtualHosts: []string{"localhost"},
		HTTPCors:         []string{"*"},
		WSHost:           "localhost",
		WSPort:           8546,
		WSModules:        []string{"oex"},
		Logger:           log.New(),
		P2PNodeDatabase:  "nodedb",
		P2PConfig:        defaultP2pConfig(),
	}
}

func defaultP2pConfig() *p2p.Config {
	cfg := &p2p.Config{
		MaxPeers:   10,
		Name:       "oex-P2P",
		ListenAddr: ":2018",
	}
	return cfg
}

func defaultFtServiceConfig() *oexservice.Config {
	return &oexservice.Config{
		DatabaseHandles: makeDatabaseHandles(),
		DatabaseCache:   768,
		TxPool:          txpool.DefaultTxPoolConfig,
		Miner:           defaultMinerConfig(),
		GasPrice: gasprice.Config{
			Blocks: 20,
		},
		MetricsConf:     defaultMetricsConfig(),
		ContractLogFlag: false,
		StatePruning:    true,
	}
}

func defaultMinerConfig() *oexservice.MinerConfig {
	return &oexservice.MinerConfig{
		Name:        params.DefaultChainconfig.SysName,
		PrivateKeys: []string{"289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"},
		ExtraData:   "system",
		Delay:       0,
	}
}

func defaultMetricsConfig() *metrics.Config {
	return &metrics.Config{
		MetricsFlag:  false,
		InfluxDBFlag: false,
		URL:          "http://localhost:8086",
		DataBase:     "metrics",
		UserName:     "",
		PassWd:       "",
		NameSpace:    "oex/",
	}
}
