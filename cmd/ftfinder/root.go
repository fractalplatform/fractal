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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/cmd/utils"
	"github.com/fractalplatform/fractal/node"
	"github.com/fractalplatform/fractal/p2p"
	"github.com/spf13/cobra"
)

var nodeConfig = node.Config{
	P2PConfig: &p2p.Config{},
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	//	Use:   "ftfinder",
	//	Short: "ftfinder is a fractal node discoverer",
	//	Long:  `ftfinder is a fractal node discoverer`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		nodeConfig.P2PConfig.PrivateKey = nodeConfig.NodeKey()
		nodeConfig.P2PConfig.BootstrapNodes = nodeConfig.BootNodes()
		srv := p2p.Server{
			Config: nodeConfig.P2PConfig,
		}
		for i, n := range srv.Config.BootstrapNodes {
			fmt.Println(i, n.String())
		}
		srv.DiscoverOnly()
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		srv.Stop()
	},
}

func init() {
	RootCmd.AddCommand(utils.VersionCmd)
	falgs := RootCmd.Flags()
	// p2p
	falgs.StringVarP(&nodeConfig.DataDir, "datadir", "d", nodeConfig.DataDir, "Data directory for the databases and keystore")
	falgs.StringVar(&nodeConfig.P2PConfig.ListenAddr, "p2p_listenaddr", nodeConfig.P2PConfig.ListenAddr,
		"Network listening address")
	falgs.StringVar(&nodeConfig.P2PConfig.NodeDatabase, "p2p_nodedb", nodeConfig.P2PConfig.NodeDatabase,
		"The path to the database containing the previously seen live nodes in the network")
	falgs.UintVar(&nodeConfig.P2PConfig.NetworkID, "p2p_id", nodeConfig.P2PConfig.NetworkID,
		"The ID of the p2p network. Nodes have different ID cannot communicate, even if they have same chainID and block data.")
	falgs.StringVar(&nodeConfig.P2PBootNodes, "p2p_bootnodes", nodeConfig.P2PBootNodes,
		"Node list file. BootstrapNodes are used to establish connectivity with the rest of the network")
	defaultLogConfig().Setup()
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
