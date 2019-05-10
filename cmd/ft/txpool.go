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
	"github.com/fractalplatform/fractal/types"

	"github.com/spf13/cobra"
)

var txpoolCommand = &cobra.Command{
	Use:   "txpool",
	Short: "Query txpool state and change txpool the Minimum gas price",
	Long:  "Query txpool state and change txpool the Minimum gas price",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var contentCmd = &cobra.Command{
	Use:   "content ",
	Short: "Returns the transactions contained within the transaction pool.",
	Long:  `Returns the transactions contained within the transaction pool.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result := map[string]map[string]map[string]*types.RPCTransaction{}
		clientCall(ipcEndpoint, &result, "txpool_content")
		printJSON(result)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status ",
	Short: "returns the number of pending and queued transaction in the pool.",
	Long:  `returns the number of pending and queued transaction in the pool.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result := map[string]int{}
		clientCall(ipcEndpoint, &result, "txpool_status")
		printJSON(result)
	},
}

var setGasPriceCmd = &cobra.Command{
	Use:   "setgasprice <gasprice uint64> ",
	Short: "set txpool the Minimum gas price ",
	Long:  `set txpool the Minimum gas price `,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		clientCall(ipcEndpoint, &result, "txpool_setGasPrice", parseUint64(args[0]))
		printJSON(result)
	},
}

func init() {
	RootCmd.AddCommand(txpoolCommand)
	txpoolCommand.AddCommand(contentCmd, statusCmd, setGasPriceCmd)
	txpoolCommand.PersistentFlags().StringVarP(&ipcEndpoint, "ipcpath", "i", defaultIPCEndpoint(params.ClientIdentifier), "IPC Endpoint path")
}
