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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/blockchain"
	"github.com/fractalplatform/fractal/ftservice"
	"github.com/fractalplatform/fractal/metrics"
	"github.com/fractalplatform/fractal/metrics/influxdb"
	"github.com/fractalplatform/fractal/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ft",
	Short: "ft is a Leading High-performance Ledger",
	Long:  `ft is a Leading High-performance Ledger`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if viper.ConfigFileUsed() != "" {
			viperUmarshalConfig()
		}

		logConfig.Setup()

		node, err := makeNode()
		if err != nil {
			log.Error("ft make node failed.", "err", err)
			return
		}

		if err := registerService(node); err != nil {
			log.Error("ft start node failed.", "err", err)
			return
		}

		if err := startNode(node); err != nil {
			log.Error("ft start node failed.", "err", err)
			return
		}

		node.Wait()

	},
}

func viperUmarshalConfig() {
	err := viper.Unmarshal(logConfig)
	if err != nil {
		fmt.Println("Unmarshal logConfig err: ", err)
		os.Exit(-1)
	}

	err = viper.Unmarshal(ftconfig.NodeCfg)
	if err != nil {
		fmt.Println("Unmarshal NodeCfg err: ", err)
		os.Exit(-1)
	}

	err = viper.Unmarshal(ftconfig.FtServiceCfg.TxPool)
	if err != nil {
		fmt.Println("Unmarshal TxPool err: ", err)
		os.Exit(-1)
	}

	err = viper.Unmarshal(&ftconfig.FtServiceCfg.Miner)
	if err != nil {
		fmt.Println("Unmarshal miner err: ", err)
		os.Exit(-1)
	}

	err = viper.Unmarshal(ftconfig.FtServiceCfg)
	if err != nil {
		fmt.Println("Unmarshal FtServiceCfg err: ", err)
		os.Exit(-1)
	}

	err = viper.Unmarshal(ftconfig.FtServiceCfg.Miner)
	if err != nil {
		fmt.Println("Unmarshal MinerConfig err: ", err)
		os.Exit(-1)
	}

	err = viper.Unmarshal(ftconfig.NodeCfg.P2PConfig)
	if err != nil {
		fmt.Println("Unmarshal MinerConfig err: ", err)
		os.Exit(-1)
	}

}

func makeNode() (*node.Node, error) {
	// set miner config
	SetupMetrics()

	// Make sure we have a valid genesis JSON
	if len(ftconfig.GenesisFileFlag) != 0 {
		file, err := os.Open(ftconfig.GenesisFileFlag)
		if err != nil {
			return nil, fmt.Errorf("Failed to read genesis file: %v(%v)", ftconfig.GenesisFileFlag, err)
		}
		defer file.Close()

		genesis := new(blockchain.Genesis)
		if err := json.NewDecoder(file).Decode(genesis); err != nil {
			return nil, fmt.Errorf("invalid genesis file: %v(%v)", ftconfig.GenesisFileFlag, err)
		}
		ftconfig.FtServiceCfg.Genesis = genesis
	}
	return node.New(ftconfig.NodeCfg)
}

func SetupMetrics() {
	//need to set metrice.Enabled = true in metrics source code
	if ftconfig.FtServiceCfg.MetricsConf.MetricsFlag {
		log.Info("Enabling metrics collection")
		if ftconfig.FtServiceCfg.MetricsConf.InfluxDBFlag {
			log.Info("Enabling influxdb collection")
			go influxdb.InfluxDBWithTags(metrics.DefaultRegistry, 10*time.Second, ftconfig.FtServiceCfg.MetricsConf.Url,
				ftconfig.FtServiceCfg.MetricsConf.DataBase, ftconfig.FtServiceCfg.MetricsConf.UserName, ftconfig.FtServiceCfg.MetricsConf.PassWd,
				ftconfig.FtServiceCfg.MetricsConf.NameSpace, map[string]string{})
		}

	}
}

// start up the node itself
func startNode(stack *node.Node) error {
	if err := stack.Start(); err != nil {
		return err
	}
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		go stack.Stop()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Warn("Already shutting down, interrupt more to panic.", "times", i-1)
			}
		}
	}()
	return nil
}

func registerService(stack *node.Node) error {
	var err error
	// register ftservice
	err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ftservice.New(ctx, ftconfig.FtServiceCfg)
	})
	return err
}

func initConfig() {
	if ftconfig.ConfigFileFlag != "" {
		viper.SetConfigFile(ftconfig.ConfigFileFlag)
	} else {
		log.Info("No config file , use default configuration.")
		return
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Error("Can't read config: %v, use default configuration.", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	falgs := RootCmd.Flags()
	// logging
	falgs.BoolVar(&logConfig.PrintOrigins, "log_debug", logConfig.PrintOrigins, "Prepends log messages with call-site location (file and line number)")
	falgs.IntVar(&logConfig.Level, "log_level", logConfig.Level, "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail")
	falgs.StringVar(&logConfig.Vmodule, "log_vmodule", logConfig.Vmodule, "Per-module verbosity: comma-separated list of <pattern>=<level> (e.g. ft/*=5,p2p=4)")
	falgs.StringVar(&logConfig.BacktraceAt, "log_backtrace", logConfig.BacktraceAt, "Request a stack trace at a specific logging statement (e.g. \"block.go:271\")")

	// config file
	falgs.StringVarP(&ftconfig.ConfigFileFlag, "config", "c", "", "TOML configuration file")
	falgs.StringVarP(&ftconfig.GenesisFileFlag, "genesis", "g", "", "genesis json file")

	// node
	falgs.StringVarP(&ftconfig.NodeCfg.DataDir, "datadir", "d", ftconfig.NodeCfg.DataDir, "Data directory for the databases and keystore")
	falgs.BoolVar(&ftconfig.NodeCfg.UseLightweightKDF, "lightkdf", ftconfig.NodeCfg.UseLightweightKDF, "Reduce key-derivation RAM & CPU usage at some expense of KDF strength")
	falgs.StringVar(&ftconfig.NodeCfg.IPCPath, "ipcpath", ftconfig.NodeCfg.IPCPath, "RPC:ipc file name")
	falgs.StringVar(&ftconfig.NodeCfg.HTTPHost, "http_host", ftconfig.NodeCfg.HTTPHost, "RPC:http host address")
	falgs.IntVar(&ftconfig.NodeCfg.HTTPPort, "http_port", ftconfig.NodeCfg.HTTPPort, "RPC:http host port")
	falgs.StringSliceVar(&ftconfig.NodeCfg.HTTPModules, "http_api", ftconfig.NodeCfg.HTTPModules, "RPC:http api's offered over the HTTP-RPC interface")
	falgs.StringSliceVar(&ftconfig.NodeCfg.HTTPCors, "http_cors", ftconfig.NodeCfg.HTTPCors, "RPC:Which to accept cross origin")
	falgs.StringSliceVar(&ftconfig.NodeCfg.HTTPVirtualHosts, "http_vhosts", ftconfig.NodeCfg.HTTPVirtualHosts, "virtual hostnames from which to accept requests")
	falgs.StringVar(&ftconfig.NodeCfg.WSHost, "ws_host", ftconfig.NodeCfg.WSHost, "RPC:websocket host address")
	falgs.IntVar(&ftconfig.NodeCfg.WSPort, "ws_port", ftconfig.NodeCfg.WSPort, "RPC:websocket host port")
	falgs.StringSliceVar(&ftconfig.NodeCfg.WSModules, "ws_api", ftconfig.NodeCfg.HTTPModules, "RPC:ws api's offered over the WS-RPC interface")
	falgs.StringSliceVar(&ftconfig.NodeCfg.WSOrigins, "ws_origins", ftconfig.NodeCfg.WSOrigins, "RPC:ws origins from which to accept websockets requests")
	falgs.BoolVar(&ftconfig.NodeCfg.WSExposeAll, "ws_exposeall", ftconfig.NodeCfg.WSExposeAll, "RPC:ws exposes all API modules via the WebSocket RPC interface rather than just the public ones.")

	// ftservice
	falgs.IntVar(&ftconfig.FtServiceCfg.DatabaseCache, "databasecache", ftconfig.FtServiceCfg.DatabaseCache, "Megabytes of memory allocated to internal database caching")

	// txpool
	falgs.BoolVar(&ftconfig.FtServiceCfg.TxPool.NoLocals, "txpool_nolocals", ftconfig.FtServiceCfg.TxPool.NoLocals, "Disables price exemptions for locally submitted transactions")
	falgs.StringVar(&ftconfig.FtServiceCfg.TxPool.Journal, "txpool_journal", ftconfig.FtServiceCfg.TxPool.Journal, "Disk journal for local transaction to survive node restarts")
	falgs.DurationVar(&ftconfig.FtServiceCfg.TxPool.Rejournal, "txpool_rejournal", ftconfig.FtServiceCfg.TxPool.Rejournal, "Time interval to regenerate the local transaction journal")
	falgs.Uint64Var(&ftconfig.FtServiceCfg.TxPool.PriceBump, "txpool_pricebump", ftconfig.FtServiceCfg.TxPool.PriceBump, "Price bump percentage to replace an already existing transaction")
	falgs.Uint64Var(&ftconfig.FtServiceCfg.TxPool.PriceLimit, "txpool_pricelimit", ftconfig.FtServiceCfg.TxPool.PriceLimit, "Minimum gas price limit to enforce for acceptance into the pool")
	falgs.Uint64Var(&ftconfig.FtServiceCfg.TxPool.AccountSlots, "txpool_accountslots", ftconfig.FtServiceCfg.TxPool.AccountSlots, "Minimum number of executable transaction slots guaranteed per account")
	falgs.Uint64Var(&ftconfig.FtServiceCfg.TxPool.AccountQueue, "txpool_accountqueue", ftconfig.FtServiceCfg.TxPool.AccountQueue, "Maximum number of non-executable transaction slots permitted per account")
	falgs.Uint64Var(&ftconfig.FtServiceCfg.TxPool.GlobalSlots, "txpool_globalslots", ftconfig.FtServiceCfg.TxPool.GlobalSlots, "Maximum number of executable transaction slots for all accounts")
	falgs.Uint64Var(&ftconfig.FtServiceCfg.TxPool.GlobalQueue, "txpool_globalqueue", ftconfig.FtServiceCfg.TxPool.GlobalQueue, "Minimum number of non-executable transaction slots for all accounts")
	falgs.DurationVar(&ftconfig.FtServiceCfg.TxPool.Lifetime, "txpool_lifetime", ftconfig.FtServiceCfg.TxPool.Lifetime, "Maximum amount of time non-executable transaction are queued")

	// miner
	falgs.BoolVar(&ftconfig.FtServiceCfg.Miner.Start, "miner_start", false, "miner start")
	falgs.StringVar(&ftconfig.FtServiceCfg.Miner.Name, "miner_coinbase", ftconfig.FtServiceCfg.Miner.Name, "name for block mining rewards")
	falgs.StringVar(&ftconfig.FtServiceCfg.Miner.PrivateKey, "miner_private", ftconfig.FtServiceCfg.Miner.PrivateKey, "hex of private key for block mining rewards")
	falgs.StringVar(&ftconfig.FtServiceCfg.Miner.ExtraData, "miner_extra", ftconfig.FtServiceCfg.Miner.ExtraData, "Block extra data set by the miner")

	// gas price oracle
	falgs.IntVar(&ftconfig.FtServiceCfg.GasPrice.Blocks, "gpo_blocks", ftconfig.FtServiceCfg.GasPrice.Blocks, "Number of recent blocks to check for gas prices")
	falgs.IntVar(&ftconfig.FtServiceCfg.GasPrice.Percentile, "gpo_percentile", ftconfig.FtServiceCfg.GasPrice.Percentile, "Suggested gas price is the given percentile of a set of recent transaction gas prices")
	falgs.BoolVar(&ftconfig.FtServiceCfg.MetricsConf.MetricsFlag, "test_metricsflag", ftconfig.FtServiceCfg.MetricsConf.MetricsFlag, "flag that open statistical metrics")
	falgs.BoolVar(&ftconfig.FtServiceCfg.MetricsConf.InfluxDBFlag, "test_influxdbflag", ftconfig.FtServiceCfg.MetricsConf.InfluxDBFlag, "flag that open influxdb thad store statistical metrics")
	falgs.StringVar(&ftconfig.FtServiceCfg.MetricsConf.Url, "test_influxdburl", ftconfig.FtServiceCfg.MetricsConf.Url, "url that connect influxdb")
	falgs.StringVar(&ftconfig.FtServiceCfg.MetricsConf.DataBase, "test_influxdbname", ftconfig.FtServiceCfg.MetricsConf.DataBase, "influxdb database name")
	falgs.StringVar(&ftconfig.FtServiceCfg.MetricsConf.UserName, "test_influxdbuser", ftconfig.FtServiceCfg.MetricsConf.UserName, "indluxdb user name")
	falgs.StringVar(&ftconfig.FtServiceCfg.MetricsConf.PassWd, "test_influxdbpasswd", ftconfig.FtServiceCfg.MetricsConf.PassWd, "influxdb user passwd")
	falgs.StringVar(&ftconfig.FtServiceCfg.MetricsConf.NameSpace, "test_influxdbnamespace", ftconfig.FtServiceCfg.MetricsConf.NameSpace, "influxdb namespace")

	// p2p
	falgs.IntVar(&ftconfig.NodeCfg.P2PConfig.MaxPeers, "p2p_maxpeers", ftconfig.NodeCfg.P2PConfig.MaxPeers,
		"Maximum number of network peers (network disabled if set to 0)")
	falgs.IntVar(&ftconfig.NodeCfg.P2PConfig.MaxPendingPeers, "p2p_maxpendpeers", ftconfig.NodeCfg.P2PConfig.MaxPendingPeers,
		"Maximum number of pending connection attempts (defaults used if set to 0)")
	falgs.IntVar(&ftconfig.NodeCfg.P2PConfig.DialRatio, "p2p_dialratio", ftconfig.NodeCfg.P2PConfig.DialRatio,
		"DialRatio controls the ratio of inbound to dialed connections")
	falgs.StringVar(&ftconfig.NodeCfg.P2PConfig.ListenAddr, "p2p_listenaddr", ftconfig.NodeCfg.P2PConfig.ListenAddr,
		"Network listening address")
	falgs.StringVar(&ftconfig.NodeCfg.P2PConfig.NodeDatabase, "p2p_nodedb", ftconfig.NodeCfg.P2PConfig.NodeDatabase,
		"The path to the database containing the previously seen live nodes in the network")
	falgs.StringVar(&ftconfig.NodeCfg.P2PConfig.Name, "p2p_nodename", ftconfig.NodeCfg.P2PConfig.Name,
		"The node name of this server")
	falgs.BoolVar(&ftconfig.NodeCfg.P2PConfig.NoDiscovery, "p2p_nodiscover", ftconfig.NodeCfg.P2PConfig.NoDiscovery,
		"Disables the peer discovery mechanism (manual peer addition)")
	falgs.BoolVar(&ftconfig.NodeCfg.P2PConfig.NoDial, "p2p_nodial", ftconfig.NodeCfg.P2PConfig.NoDial,
		"The server will not dial any peers.")
	falgs.StringVar(&ftconfig.NodeCfg.P2PBootNodes, "p2p_bootnodes", ftconfig.NodeCfg.P2PBootNodes,
		"Node list file. BootstrapNodes are used to establish connectivity with the rest of the network")
	falgs.StringVar(&ftconfig.NodeCfg.P2PStaticNodes, "p2p_staticnodes", ftconfig.NodeCfg.P2PStaticNodes,
		"Node list file. Static nodes are used as pre-configured connections which are always maintained and re-connected on disconnects")
	falgs.StringVar(&ftconfig.NodeCfg.P2PTrustNodes, "p2p_trustnodes", ftconfig.NodeCfg.P2PStaticNodes,
		"Node list file. Trusted nodes are usesd as pre-configured connections which are always allowed to connect, even above the peer limit")
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
