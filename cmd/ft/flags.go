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
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// ConfigFile the Fractal config file
	ConfigFile string
)

func addFlags(flags *flag.FlagSet) {
	// debug
	flags.BoolVar(
		&ftCfgInstance.DebugCfg.Pprof,
		"debug_pprof",
		ftCfgInstance.DebugCfg.Pprof,
		"Enable the pprof HTTP server",
	)
	viper.BindPFlag("debug.pprof", flags.Lookup("debug_pprof"))

	flags.IntVar(
		&ftCfgInstance.DebugCfg.PprofPort,
		"debug_pprof_port",
		ftCfgInstance.DebugCfg.PprofPort,
		"Pprof HTTP server listening port",
	)
	viper.BindPFlag("debug.pprofport", flags.Lookup("debug_pprof_port"))

	flags.StringVar(
		&ftCfgInstance.DebugCfg.PprofAddr,
		"debug_pprof_addr",
		ftCfgInstance.DebugCfg.PprofAddr,
		"Pprof HTTP server listening interface",
	)
	viper.BindPFlag("debug.pprofaddr", flags.Lookup("debug_pprof_addr"))

	flags.IntVar(
		&ftCfgInstance.DebugCfg.Memprofilerate,
		"debug_memprofilerate",
		ftCfgInstance.DebugCfg.Memprofilerate,
		"Turn on memory profiling with the given rate",
	)
	viper.BindPFlag("debug.memprofilerate", flags.Lookup("debug_memprofilerate"))

	flags.IntVar(
		&ftCfgInstance.DebugCfg.Blockprofilerate,
		"debug_blockprofilerate",
		ftCfgInstance.DebugCfg.Blockprofilerate,
		"Turn on block profiling with the given rate",
	)
	viper.BindPFlag("debug.blockprofilerate", flags.Lookup("debug_blockprofilerate"))

	flags.StringVar(
		&ftCfgInstance.DebugCfg.Cpuprofile,
		"debug_cpuprofile",
		ftCfgInstance.DebugCfg.Cpuprofile,
		"Write CPU profile to the given file",
	)
	viper.BindPFlag("debug.cpuprofile", flags.Lookup("debug_cpuprofile"))

	flags.StringVar(
		&ftCfgInstance.DebugCfg.Trace,
		"debug_trace",
		ftCfgInstance.DebugCfg.Trace,
		"Write execution trace to the given file",
	)
	viper.BindPFlag("debug.trace", flags.Lookup("debug_trace"))

	// log
	flags.StringVar(
		&ftCfgInstance.LogCfg.Logdir,
		"log_dir",
		ftCfgInstance.LogCfg.Logdir,
		"Writes log records to file chunks at the given path",
	)
	viper.BindPFlag("log.dir", flags.Lookup("log_dir"))

	flags.BoolVar(
		&ftCfgInstance.LogCfg.PrintOrigins,
		"log_debug",
		ftCfgInstance.LogCfg.PrintOrigins,
		"Prepends log messages with call-site location (file and line number)",
	)
	viper.BindPFlag("log.debug", flags.Lookup("log_debug"))

	flags.IntVar(
		&ftCfgInstance.LogCfg.Level,
		"log_level",
		ftCfgInstance.LogCfg.Level,
		"Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
	)
	viper.BindPFlag("log.level", flags.Lookup("log_level"))

	flags.StringVar(
		&ftCfgInstance.LogCfg.Vmodule,
		"log_module",
		ftCfgInstance.LogCfg.Vmodule,
		"Per-module verbosity: comma-separated list of <pattern>=<level> (e.g. ft/*=5,p2p=4)",
	)
	viper.BindPFlag("log.module", flags.Lookup("log_module"))

	flags.StringVar(
		&ftCfgInstance.LogCfg.BacktraceAt,
		"log_backtrace",
		ftCfgInstance.LogCfg.BacktraceAt,
		"Request a stack trace at a specific logging statement (e.g. \"block.go:271\")",
	)
	viper.BindPFlag("log.backtrace", flags.Lookup("log_backtrace"))

	// config file
	flags.StringVarP(
		&ConfigFile,
		"config", "c",
		"",
		"TOML/YAML configuration file",
	)

	// Genesis File
	flags.StringVarP(
		&ftCfgInstance.GenesisFile,
		"genesis",
		"g", "",
		"Genesis json file",
	)
	viper.BindPFlag("genesis", flags.Lookup("genesis"))

	// node datadir
	flags.StringVarP(
		&ftCfgInstance.NodeCfg.DataDir,
		"datadir", "d",
		ftCfgInstance.NodeCfg.DataDir,
		"Data directory for the databases ",
	)
	viper.BindPFlag("node.datadir", flags.Lookup("datadir"))

	// node
	flags.StringVar(
		&ftCfgInstance.NodeCfg.IPCPath,
		"ipcpath",
		ftCfgInstance.NodeCfg.IPCPath,
		"RPC:ipc file name",
	)
	viper.BindPFlag("node.ipcpath", flags.Lookup("ipcpath"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.HTTPHost,
		"http_host",
		ftCfgInstance.NodeCfg.HTTPHost,
		"RPC:http host address",
	)
	viper.BindPFlag("node.httphost", flags.Lookup("http_host"))

	flags.IntVar(
		&ftCfgInstance.NodeCfg.HTTPPort,
		"http_port",
		ftCfgInstance.NodeCfg.HTTPPort,
		"RPC:http host port",
	)
	viper.BindPFlag("node.httpport", flags.Lookup("http_port"))

	flags.StringSliceVar(
		&ftCfgInstance.NodeCfg.HTTPModules,
		"http_modules",
		ftCfgInstance.NodeCfg.HTTPModules,
		"RPC:http api's offered over the HTTP-RPC interface",
	)
	viper.BindPFlag("node.httpmodules", flags.Lookup("http_modules"))

	flags.StringSliceVar(
		&ftCfgInstance.NodeCfg.HTTPCors,
		"http_cors",
		ftCfgInstance.NodeCfg.HTTPCors,
		"RPC:Which to accept cross origin",
	)
	viper.BindPFlag("node.httpcors", flags.Lookup("http_cors"))

	flags.StringSliceVar(
		&ftCfgInstance.NodeCfg.HTTPVirtualHosts,
		"http_vhosts",
		ftCfgInstance.NodeCfg.HTTPVirtualHosts,
		"RPC:http virtual hostnames from which to accept requests",
	)
	viper.BindPFlag("node.httpvirtualhosts", flags.Lookup("http_vhosts"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.WSHost,
		"ws_host",
		ftCfgInstance.NodeCfg.WSHost,
		"RPC:websocket host address",
	)
	viper.BindPFlag("node.wshost", flags.Lookup("ws_host"))

	flags.IntVar(
		&ftCfgInstance.NodeCfg.WSPort,
		"ws_port",
		ftCfgInstance.NodeCfg.WSPort,
		"RPC:websocket host port",
	)
	viper.BindPFlag("node.wsport", flags.Lookup("ws_port"))

	flags.StringSliceVar(
		&ftCfgInstance.NodeCfg.WSModules,
		"ws_modules",
		ftCfgInstance.NodeCfg.WSModules,
		"RPC:ws api's offered over the WS-RPC interface",
	)
	viper.BindPFlag("node.wsmodules", flags.Lookup("ws_modules"))

	flags.StringSliceVar(
		&ftCfgInstance.NodeCfg.WSOrigins,
		"ws_origins",
		ftCfgInstance.NodeCfg.WSOrigins,
		"RPC:ws origins from which to accept websockets requests",
	)
	viper.BindPFlag("node.wsorigins", flags.Lookup("ws_origins"))

	flags.BoolVar(
		&ftCfgInstance.NodeCfg.WSExposeAll,
		"ws_exposeall",
		ftCfgInstance.NodeCfg.WSExposeAll,
		"RPC:ws exposes all API modules via the WebSocket RPC interface rather than just the public ones.",
	)
	viper.BindPFlag("node.wsexposeall", flags.Lookup("ws_exposeall"))

	// ftservice database options
	flags.IntVar(
		&ftCfgInstance.FtServiceCfg.DatabaseCache,
		"database_cache",
		ftCfgInstance.FtServiceCfg.DatabaseCache,
		"Megabytes of memory allocated to internal database caching",
	)
	viper.BindPFlag("ftservice.databasecache", flags.Lookup("database_cache"))

	flags.BoolVar(
		&ftCfgInstance.FtServiceCfg.ContractLogFlag,
		"contractlog",
		ftCfgInstance.FtServiceCfg.ContractLogFlag,
		"flag for db to store contrat internal transaction log.",
	)
	viper.BindPFlag("ftservice.contractlog", flags.Lookup("contractlog"))

	// state pruning
	flags.BoolVar(
		&ftCfgInstance.FtServiceCfg.StatePruning,
		"statepruning_enable",
		ftCfgInstance.FtServiceCfg.StatePruning,
		"flag for enable/disable state pruning.",
	)
	viper.BindPFlag("ftservice.statepruning", flags.Lookup("statepruning_enable"))

	// start number
	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.StartNumber,
		"start_number",
		ftCfgInstance.FtServiceCfg.StartNumber,
		"start chain with a specified block number.",
	)
	viper.BindPFlag("ftservice.startnumber", flags.Lookup("start_number"))

	// add bad block hashs
	flags.StringSliceVar(
		&ftCfgInstance.FtServiceCfg.BadHashes,
		"bad_hashes",
		ftCfgInstance.FtServiceCfg.BadHashes,
		"blockchain refuse bad block hashes",
	)
	viper.BindPFlag("ftservice.badhashes", flags.Lookup("bad_hashes"))

	// txpool
	flags.BoolVar(
		&ftCfgInstance.FtServiceCfg.TxPool.NoLocals,
		"txpool_nolocals",
		ftCfgInstance.FtServiceCfg.TxPool.NoLocals,
		"Disables price exemptions for locally submitted transactions",
	)
	viper.BindPFlag("ftservice.txpool.nolocals", flags.Lookup("txpool_nolocals"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.TxPool.Journal,
		"txpool_journal",
		ftCfgInstance.FtServiceCfg.TxPool.Journal,
		"Disk journal for local transaction to survive node restarts",
	)
	viper.BindPFlag("ftservice.txpool.journal", flags.Lookup("txpool_journal"))

	flags.DurationVar(
		&ftCfgInstance.FtServiceCfg.TxPool.Rejournal,
		"txpool_rejournal",
		ftCfgInstance.FtServiceCfg.TxPool.Rejournal,
		"Time interval to regenerate the local transaction journal",
	)
	viper.BindPFlag("ftservice.txpool.rejournal", flags.Lookup("txpool_rejournal"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.PriceBump,
		"txpool_pricebump",
		ftCfgInstance.FtServiceCfg.TxPool.PriceBump,
		"Price bump percentage to replace an already existing transaction",
	)
	viper.BindPFlag("ftservice.txpool.pricebump", flags.Lookup("txpool_pricebump"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.PriceLimit,
		"txpool_pricelimit",
		ftCfgInstance.FtServiceCfg.TxPool.PriceLimit,
		"Minimum gas price limit to enforce for acceptance into the pool",
	)
	viper.BindPFlag("ftservice.txpool.pricelimit", flags.Lookup("txpool_pricelimit"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.AccountSlots,
		"txpool_accountslots",
		ftCfgInstance.FtServiceCfg.TxPool.AccountSlots,
		"Number of executable transaction slots guaranteed per account",
	)
	viper.BindPFlag("ftservice.txpool.accountslots", flags.Lookup("txpool_accountslots"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.AccountQueue,
		"txpool_accountqueue",
		ftCfgInstance.FtServiceCfg.TxPool.AccountQueue,
		"Maximum number of non-executable transaction slots permitted per account",
	)
	viper.BindPFlag("ftservice.txpool.accountqueue", flags.Lookup("txpool_accountqueue"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.GlobalSlots,
		"txpool_globalslots",
		ftCfgInstance.FtServiceCfg.TxPool.GlobalSlots,
		"Maximum number of executable transaction slots for all accounts",
	)
	viper.BindPFlag("ftservice.txpool.globalslots", flags.Lookup("txpool_globalslots"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.GlobalQueue,
		"txpool_globalqueue",
		ftCfgInstance.FtServiceCfg.TxPool.GlobalQueue,
		"Minimum number of non-executable transaction slots for all accounts",
	)
	viper.BindPFlag("ftservice.txpool.globalqueue", flags.Lookup("txpool_globalqueue"))

	flags.DurationVar(
		&ftCfgInstance.FtServiceCfg.TxPool.Lifetime,
		"txpool_lifetime",
		ftCfgInstance.FtServiceCfg.TxPool.Lifetime,
		"Maximum amount of time non-executable transaction are queued",
	)
	viper.BindPFlag("ftservice.txpool.lifetime", flags.Lookup("txpool_lifetime"))

	flags.DurationVar(
		&ftCfgInstance.FtServiceCfg.TxPool.ResendTime,
		"txpool_resendtime",
		ftCfgInstance.FtServiceCfg.TxPool.ResendTime,
		"Maximum amount of time  executable transaction are resended",
	)
	viper.BindPFlag("ftservice.txpool.resendtime", flags.Lookup("txpool_resendtime"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.MinBroadcast,
		"txpool_minbroadcast",
		ftCfgInstance.FtServiceCfg.TxPool.MinBroadcast,
		"Minimum number of nodes for the transaction broadcast",
	)
	viper.BindPFlag("ftservice.txpool.minbroadcast", flags.Lookup("txpool_minbroadcast"))

	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.TxPool.RatioBroadcast,
		"txpool_ratiobroadcast",
		ftCfgInstance.FtServiceCfg.TxPool.RatioBroadcast,
		"Ratio of nodes for the transaction broadcast",
	)
	viper.BindPFlag("ftservice.txpool.ratiobroadcast", flags.Lookup("txpool_ratiobroadcast"))

	// miner
	flags.BoolVar(
		&ftCfgInstance.FtServiceCfg.Miner.Start,
		"miner_start",
		ftCfgInstance.FtServiceCfg.Miner.Start,
		"Start miner generate block and process transaction",
	)
	viper.BindPFlag("ftservice.miner.start", flags.Lookup("miner_start"))

	// miner
	flags.Uint64Var(
		&ftCfgInstance.FtServiceCfg.Miner.Delay,
		"miner_delay",
		ftCfgInstance.FtServiceCfg.Miner.Delay,
		"delay duration for miner (ms)",
	)
	viper.BindPFlag("ftservice.miner.delay", flags.Lookup("miner_delay"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.Miner.Name,
		"miner_name",
		ftCfgInstance.FtServiceCfg.Miner.Name,
		"Name for block mining rewards",
	)
	viper.BindPFlag("ftservice.miner.name", flags.Lookup("miner_name"))

	flags.StringSliceVar(
		&ftCfgInstance.FtServiceCfg.Miner.PrivateKeys,
		"miner_private",
		ftCfgInstance.FtServiceCfg.Miner.PrivateKeys,
		"Hex of private key for block mining rewards",
	)
	viper.BindPFlag("ftservice.miner.private", flags.Lookup("miner_private"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.Miner.ExtraData,
		"miner_extra",
		ftCfgInstance.FtServiceCfg.Miner.ExtraData,
		"Block extra data set by the miner",
	)
	viper.BindPFlag("ftservice.miner.name", flags.Lookup("miner_extra"))

	// gas price oracle
	flags.IntVar(
		&ftCfgInstance.FtServiceCfg.GasPrice.Blocks,
		"gpo_blocks",
		ftCfgInstance.FtServiceCfg.GasPrice.Blocks,
		"Number of recent blocks to check for gas prices",
	)
	viper.BindPFlag("ftservice.gpo.blocks", flags.Lookup("gpo_blocks"))

	flags.IntVar(
		&ftCfgInstance.FtServiceCfg.GasPrice.Percentile,
		"gpo_percentile",
		ftCfgInstance.FtServiceCfg.GasPrice.Percentile,
		"Suggested gas price is the given percentile of a set of recent transaction gas prices",
	)
	viper.BindPFlag("ftservice.gpo.percentile", flags.Lookup("gpo_percentile"))

	// metrics
	flags.BoolVar(
		&ftCfgInstance.FtServiceCfg.MetricsConf.MetricsFlag,
		"metrics_start",
		ftCfgInstance.FtServiceCfg.MetricsConf.MetricsFlag,
		"flag that open statistical metrics",
	)
	viper.BindPFlag("ftservice.metrics.start", flags.Lookup("metrics_start"))

	flags.BoolVar(
		&ftCfgInstance.FtServiceCfg.MetricsConf.InfluxDBFlag,
		"metrics_influxdb",
		ftCfgInstance.FtServiceCfg.MetricsConf.InfluxDBFlag,
		"flag that open influxdb thad store statistical metrics",
	)
	viper.BindPFlag("ftservice.metrics.influxdb", flags.Lookup("metrics_influxdb"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.MetricsConf.URL,
		"metrics_influxdb_URL",
		ftCfgInstance.FtServiceCfg.MetricsConf.URL,
		"URL that connect influxdb",
	)
	viper.BindPFlag("ftservice.metrics.influxdbURL", flags.Lookup("metrics_influxdb_URL"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.MetricsConf.DataBase,
		"metrics_influxdb_name",
		ftCfgInstance.FtServiceCfg.MetricsConf.DataBase,
		"Influxdb database name",
	)
	viper.BindPFlag("ftservice.metrics.influxdbname", flags.Lookup("metrics_influxdb_name"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.MetricsConf.UserName,
		"metrics_influxdb_user",
		ftCfgInstance.FtServiceCfg.MetricsConf.UserName,
		"Indluxdb user name",
	)
	viper.BindPFlag("ftservice.metrics.influxdbuser", flags.Lookup("metrics_influxdb_user"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.MetricsConf.PassWd,
		"metrics_influxdb_passwd",
		ftCfgInstance.FtServiceCfg.MetricsConf.PassWd,
		"Influxdb user passwd",
	)
	viper.BindPFlag("ftservice.metrics.influxdbpasswd", flags.Lookup("metrics_influxdb_passwd"))

	flags.StringVar(
		&ftCfgInstance.FtServiceCfg.MetricsConf.NameSpace,
		"metrics_influxdb_namespace",
		ftCfgInstance.FtServiceCfg.MetricsConf.NameSpace,
		"Influxdb namespace",
	)
	viper.BindPFlag("ftservice.metrics.influxdbnamepace", flags.Lookup("metrics_influxdb_namespace"))

	// p2p
	flags.UintVar(
		&ftCfgInstance.NodeCfg.P2PConfig.NetworkID,
		"p2p_id",
		ftCfgInstance.NodeCfg.P2PConfig.NetworkID,
		"The ID of the p2p network. Nodes have different ID cannot communicate, even if they have same chainID and block data.",
	)
	viper.BindPFlag("ftservice.p2p.networkid", flags.Lookup("p2p_id"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.P2PConfig.Name,
		"p2p_name",
		ftCfgInstance.NodeCfg.P2PConfig.Name,
		"The name sets the p2p node name of this server",
	)
	viper.BindPFlag("ftservice.p2p.name", flags.Lookup("p2p_name"))

	flags.IntVar(
		&ftCfgInstance.NodeCfg.P2PConfig.MaxPeers,
		"p2p_maxpeers",
		ftCfgInstance.NodeCfg.P2PConfig.MaxPeers,
		"Maximum number of network peers ",
	)
	viper.BindPFlag("ftservice.p2p.maxpeers", flags.Lookup("p2p_maxpeers"))

	flags.IntVar(
		&ftCfgInstance.NodeCfg.P2PConfig.MaxPendingPeers,
		"p2p_maxpendpeers",
		ftCfgInstance.NodeCfg.P2PConfig.MaxPendingPeers,
		"Maximum number of pending connection attempts ",
	)
	viper.BindPFlag("ftservice.p2p.maxpendpeers", flags.Lookup("p2p_maxpendpeers"))

	flags.IntVar(
		&ftCfgInstance.NodeCfg.P2PConfig.DialRatio,
		"p2p_dialratio",
		ftCfgInstance.NodeCfg.P2PConfig.DialRatio,
		"DialRatio controls the ratio of inbound to dialed connections",
	)
	viper.BindPFlag("ftservice.p2p.dialratio", flags.Lookup("p2p_dialratio"))

	flags.IntVar(
		&ftCfgInstance.NodeCfg.P2PConfig.PeerPeriod,
		"p2p_peerperiod",
		ftCfgInstance.NodeCfg.P2PConfig.PeerPeriod,
		"Disconnect the worst peer every 'p2p_peerperiod' ms(if peer count equal p2p_maxpeers), 0 means disable.",
	)
	viper.BindPFlag("ftservice.p2p.peerperiod", flags.Lookup("p2p_peerperiod"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.P2PConfig.ListenAddr,
		"p2p_listenaddr",
		ftCfgInstance.NodeCfg.P2PConfig.ListenAddr,
		"Network listening address",
	)
	viper.BindPFlag("ftservice.p2p.listenaddr", flags.Lookup("p2p_listenaddr"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.P2PConfig.NodeDatabase,
		"p2p_nodedb",
		ftCfgInstance.NodeCfg.P2PConfig.NodeDatabase,
		"The path to the database containing the previously seen live nodes in the network",
	)
	viper.BindPFlag("ftservice.p2p.nodedb", flags.Lookup("p2p_nodedb"))

	flags.BoolVar(
		&ftCfgInstance.NodeCfg.P2PConfig.NoDiscovery,
		"p2p_nodiscovery",
		ftCfgInstance.NodeCfg.P2PConfig.NoDiscovery,
		"Disables the peer discovery mechanism (manual peer addition)",
	)
	viper.BindPFlag("ftservice.p2p.nodiscovery", flags.Lookup("p2p_nodiscovery"))

	flags.BoolVar(
		&ftCfgInstance.NodeCfg.P2PConfig.NoDial,
		"p2p_nodial",
		ftCfgInstance.NodeCfg.P2PConfig.NoDial,
		"The server will not dial any peers.",
	)
	viper.BindPFlag("ftservice.p2p.nodial", flags.Lookup("p2p_nodial"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.P2PBootNodes,
		"p2p_bootnodes",
		ftCfgInstance.NodeCfg.P2PBootNodes,
		"Node list file. BootstrapNodes are used to establish connectivity with the rest of the network",
	)
	viper.BindPFlag("ftservice.p2p.bootnodes", flags.Lookup("p2p_bootnodes"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.P2PStaticNodes,
		"p2p_staticnodes",
		ftCfgInstance.NodeCfg.P2PStaticNodes,
		"Node list file. Static nodes are used as pre-configured connections which are always maintained and re-connected on disconnects",
	)
	viper.BindPFlag("ftservice.p2p.staticnodes", flags.Lookup("p2p_staticnodes"))

	flags.StringVar(
		&ftCfgInstance.NodeCfg.P2PTrustNodes,
		"p2p_trustnodes",
		ftCfgInstance.NodeCfg.P2PStaticNodes,
		"Node list file. Trusted nodes are usesd as pre-configured connections which are always allowed to connect, even above the peer limit",
	)
	viper.BindPFlag("ftservice.p2p.trustnodes", flags.Lookup("p2p_trustnodes"))

}
