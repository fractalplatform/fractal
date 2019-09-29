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
}

var contentCmd = &cobra.Command{
	Use:   "content <fullTx bool>",
	Short: "Returns the transactions contained within the transaction pool.",
	Long:  `Returns the transactions contained within the transaction pool.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			result interface{}
			fullTx bool
		)
		if len(args) == 1 {
			fullTx = parseBool(args[0])
		}

		clientCall(ipcEndpoint, &result, "txpool_content", fullTx)
		printJSON(result)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status ",
	Short: "Returns the number of pending and queued transaction in the pool.",
	Long:  `Returns the number of pending and queued transaction in the pool.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result := map[string]int{}
		clientCall(ipcEndpoint, &result, "txpool_status")
		printJSON(result)
	},
}

var setGasPriceCmd = &cobra.Command{
	Use:   "setgasprice <gasprice uint64> ",
	Short: "Set txpool the Minimum gas price ",
	Long:  `Set txpool the Minimum gas price `,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		clientCall(ipcEndpoint, &result, "txpool_setGasPrice", parseBigInt(args[0]))
		printJSON(result)
	},
}

var getTxsCmd = &cobra.Command{
	Use:   "gettxs <txhashes string array> ",
	Short: "Returns the transaction for the given hash",
	Long:  `Returns the transaction for the given hash`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var result []*types.RPCTransaction
		clientCall(ipcEndpoint, &result, "txpool_getTransactions", args)
		printJSONList(result)
	},
}

var getTxsByAccountCmd = &cobra.Command{
	Use:   "gettxsbyname <account string> <fullTx bool> ",
	Short: "Returns the transaction for the given account",
	Long:  `Returns the transaction for the given account`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			result interface{}
			fullTx bool
		)
		if len(args) > 1 {
			fullTx = parseBool(args[1])
		}

		clientCall(ipcEndpoint, &result, "txpool_getTransactionsByAccount", args[0], fullTx)
		printJSON(result)
	},
}

var getPendingTxsCmd = &cobra.Command{
	Use:   "getpending <fullTx bool>",
	Short: "Returns the pending transactions that are in the transaction pool",
	Long:  `Returns the pending transactions that are in the transaction pool`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var result interface{}
		clientCall(ipcEndpoint, &result, "txpool_pendingTransactions", parseBool(args[0]))
		printJSON(result)
	},
}

func init() {
	RootCmd.AddCommand(txpoolCommand)
	txpoolCommand.AddCommand(contentCmd, statusCmd, setGasPriceCmd, getTxsCmd, getTxsByAccountCmd, getPendingTxsCmd)
	txpoolCommand.PersistentFlags().StringVarP(&ipcEndpoint, "ipcpath", "i", defaultIPCEndpoint(params.ClientIdentifier), "IPC Endpoint path")
}
