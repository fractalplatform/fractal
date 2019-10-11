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
	"github.com/fractalplatform/fractal/params"
	"github.com/spf13/cobra"
)

var p2pCmd = &cobra.Command{
	Use:   "p2p",
	Short: "Offers and API for p2p networking",
	Long:  `Offers and API for p2p networking`,
	Args:  cobra.NoArgs,
}

var commonCall = func(method string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		params := make([]interface{}, len(args))
		for i, arg := range args {
			params[i] = arg
		}
		result := clientCallRaw(ipcEndpoint, method, params...)
		printJSON(result)
	}
}

var p2pSubCmds = []*cobra.Command{
	&cobra.Command{
		Use:   "add <url>",
		Short: "Connecting to a remote node.",
		Long:  `Connecting to a remote node.`,
		Args:  cobra.ExactArgs(1),
		Run:   commonCall("p2p_addPeer"),
	},

	&cobra.Command{
		Use:   "remove <url>",
		Short: "Disconnects from a remote node if the connection exists.",
		Long:  `Disconnects from a remote node if the connection exists.`,
		Args:  cobra.ExactArgs(1),
		Run:   commonCall("p2p_removePeer"),
	},

	&cobra.Command{
		Use:   "addtrusted <url>",
		Short: "Allows a remote node to always connect, even if slots are full.",
		Long:  `Allows a remote node to always connect, even if slots are full.`,
		Args:  cobra.ExactArgs(1),
		Run:   commonCall("p2p_addTrustedPeer"),
	},

	&cobra.Command{
		Use:   "removetrusted <url>",
		Short: "Removes a remote node from the trusted peer set, but it does not disconnect it automatically.",
		Long:  `Removes a remote node from the trusted peer set, but it does not disconnect it automatically.`,
		Args:  cobra.ExactArgs(1),
		Run:   commonCall("p2p_removeTrustedPeer"),
	},

	&cobra.Command{
		Use:   "addbad <url>",
		Short: "Add a bad node in black list.",
		Long:  `Add a bad node in black list..`,
		Args:  cobra.ExactArgs(1),
		Run:   commonCall("p2p_addBadNode"),
	},

	&cobra.Command{
		Use:   "removebad <url>",
		Short: "Removes a bad node from the black peer set.",
		Long:  `Removes a bad node from the black peer set.`,
		Args:  cobra.ExactArgs(1),
		Run:   commonCall("p2p_removeBadNode"),
	},

	&cobra.Command{
		Use:   "count",
		Short: "Return number of connected peers.",
		Long:  `Return number of connected peers.`,
		Args:  cobra.NoArgs,
		Run:   commonCall("p2p_peerCount"),
	},

	&cobra.Command{
		Use:   "list",
		Short: "Return connected peers list.",
		Long:  `Return connected peers list.`,
		Args:  cobra.NoArgs,
		Run:   commonCall("p2p_peers"),
	},

	&cobra.Command{
		Use:   "badcount",
		Short: "Return number of bad nodes .",
		Long:  `Return number of bad nodes .`,
		Args:  cobra.NoArgs,
		Run:   commonCall("p2p_badNodesCount"),
	},

	&cobra.Command{
		Use:   "badlist",
		Short: "Return bad nodes list.",
		Long:  `Return bad nodes list.`,
		Args:  cobra.NoArgs,
		Run:   commonCall("p2p_badNodes"),
	},

	&cobra.Command{
		Use:   "selfnode",
		Short: "Return self enode url.",
		Long:  `Return self enode url.`,
		Args:  cobra.NoArgs,
		Run:   commonCall("p2p_selfNode"),
	},

	&cobra.Command{
		Use:   "seednodes",
		Short: "Return seed enode url.",
		Long:  `Return seed enode url.`,
		Args:  cobra.NoArgs,
		Run:   commonCall("p2p_seedNodes"),
	},
}

func init() {
	RootCmd.AddCommand(p2pCmd)
	p2pCmd.AddCommand(p2pSubCmds...)
	p2pCmd.PersistentFlags().StringVarP(&ipcEndpoint, "ipcpath", "i", defaultIPCEndpoint(params.ClientIdentifier), "IPC Endpoint path")
}
