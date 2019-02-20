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
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/node"
	"github.com/fractalplatform/fractal/p2p"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	//	Use:   "fkey",
	//	Short: "fkey is a fractal key manager",
	//	Long:  `fkey is a fractal key manager`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		nodekey, _ := crypto.GenerateKey()
		cfg := node.Config{
			P2PBootNodes: "./bootnodes",
		}
		srv := p2p.Server{
			Config: &p2p.Config{
				PrivateKey:     nodekey,
				Name:           "Finder",
				ListenAddr:     ":12345",
				BootstrapNodes: cfg.BootNodes(),
				NodeDatabase:   "",
			},
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
	falgs := RootCmd.Flags()
	_ = falgs
	/*
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
	*/
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
