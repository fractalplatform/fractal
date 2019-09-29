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
	"github.com/fractalplatform/fractal/cmd/utils"
	"github.com/fractalplatform/fractal/debug"
	"github.com/fractalplatform/fractal/ftservice"
	"github.com/fractalplatform/fractal/metrics"
	"github.com/fractalplatform/fractal/metrics/influxdb"
	"github.com/fractalplatform/fractal/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	errNoConfigFile    string
	errViperReadConfig error
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ft",
	Short: "ft is a Leading High-performance Ledger",
	Long:  `ft is a Leading High-performance Ledger`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if viper.ConfigFileUsed() != "" {
			err = viper.Unmarshal(ftCfgInstance)
		}
		ftCfgInstance.LogCfg.Setup()
		if errNoConfigFile != "" {
			log.Info(errNoConfigFile)
		}

		if errViperReadConfig != nil {
			log.Error("Can't read config, use default configuration", "err", errViperReadConfig)
		}

		if err != nil {
			log.Error("viper umarshal config file faild", "err", err)
		}

		if err := debug.Setup(ftCfgInstance.DebugCfg); err != nil {
			log.Error("debug setup faild", "err", err)
		}

		log.Info("fractal node", "version", utils.FullVersion())

		node, err := makeNode()
		if err != nil {
			log.Error("ft make node failed.", "err", err)
			return
		}

		if err := registerService(node); err != nil {
			log.Error("ft register service failed.", "err", err)
			return
		}

		if err := startNode(node); err != nil {
			log.Error("ft start node failed.", "err", err)
			return
		}

		node.Wait()
		debug.Exit()
	},
}

func makeNode() (*node.Node, error) {
	genesis := blockchain.DefaultGenesis()
	// set miner config
	SetupMetrics()
	// Make sure we have a valid genesis JSON
	if len(ftCfgInstance.GenesisFile) != 0 {
		log.Info("Reading read genesis file", "path", ftCfgInstance.GenesisFile)
		file, err := os.Open(ftCfgInstance.GenesisFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to read genesis file: %v(%v)", ftCfgInstance.GenesisFile, err)
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(genesis); err != nil {
			return nil, fmt.Errorf("invalid genesis file: %v(%v)", ftCfgInstance.GenesisFile, err)
		}
		ftCfgInstance.FtServiceCfg.Genesis = genesis

	}
	block, _, err := genesis.ToBlock(nil)
	if err != nil {
		return nil, err
	}
	// p2p used to generate MagicNetID
	ftCfgInstance.NodeCfg.P2PConfig.GenesisHash = block.Hash()
	return node.New(ftCfgInstance.NodeCfg)
}

// SetupMetrics set metrics
func SetupMetrics() {
	//need to set metrice.Enabled = true in metrics source code
	if ftCfgInstance.FtServiceCfg.MetricsConf.MetricsFlag {
		log.Info("Enabling metrics collection")
		if ftCfgInstance.FtServiceCfg.MetricsConf.InfluxDBFlag {
			log.Info("Enabling influxdb collection")
			go influxdb.InfluxDBWithTags(metrics.DefaultRegistry, 10*time.Second, ftCfgInstance.FtServiceCfg.MetricsConf.URL,
				ftCfgInstance.FtServiceCfg.MetricsConf.DataBase, ftCfgInstance.FtServiceCfg.MetricsConf.UserName, ftCfgInstance.FtServiceCfg.MetricsConf.PassWd,
				ftCfgInstance.FtServiceCfg.MetricsConf.NameSpace, map[string]string{})
		}

	}
}

// start up the node itself
func startNode(stack *node.Node) error {
	debug.Memsize.Add("node", stack)
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
		debug.Exit() // ensure trace and CPU profile data is flushed.
		debug.LoudPanic("boom")
	}()
	return nil
}

func registerService(stack *node.Node) error {
	return stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return ftservice.New(ctx, ftCfgInstance.FtServiceCfg)
	})
}

func initConfig() {
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
	} else {
		errNoConfigFile = "No config file , use default configuration."
		return
	}
	if err := viper.ReadInConfig(); err != nil {
		errViperReadConfig = err
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.AddCommand(utils.VersionCmd)
	addFlags(RootCmd.Flags())
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
