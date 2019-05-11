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
	"bufio"
	"io"
	"os"
	"regexp"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var minerCmd = &cobra.Command{
	Use:   "miner",
	Short: "control miner start or stop else",
	Long:  `control miner start or stop else`,
	Args:  cobra.NoArgs,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start mint new block.",
	Long:  `start mint new block.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		clientCall(ipcEndpoint, &result, "miner_start")
		printJSON(result)
	},
}

var forceCmd = &cobra.Command{
	Use:   "force ",
	Short: "force start mint new block.",
	Long:  `force start mint new block.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		clientCall(ipcEndpoint, &result, "miner_force")
		printJSON(result)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop mint block.",
	Long:  `Stop mint block.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		clientCall(ipcEndpoint, &result, "miner_stop")
		printJSON(result)
	},
}

var miningCmd = &cobra.Command{
	Use:   "mining",
	Short: "Return the miner is mining",
	Long:  `Return the miner is mining.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		clientCall(ipcEndpoint, &result, "miner_mining")
		printJSON(result)
	},
}

var setCoinbaseCmd = &cobra.Command{
	Use:   "setcoinbase <name> <privKeys file>",
	Short: "Set the coinbase of the miner.",
	Long:  `Set the coinbase of the miner.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		name := common.Name(args[0])
		if !name.IsValid(regexp.MustCompile("^([a-z][a-z0-9]{6,15})(?:\\.([a-z0-9]{1,8})){0,1}$")) {
			jww.ERROR.Println("valid name: " + name)
			return
		}
		path := args[1]
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			jww.ERROR.Println("file is not exist.", "path", path)
			return
		}

		fi, err := os.Open(path)
		if err != nil {
			jww.ERROR.Println("read failed.", "path", path, "err", err)
			return
		}
		defer fi.Close()

		var keys []string
		br := bufio.NewReader(fi)
		for {
			line, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			keys = append(keys, string(line))
		}

		if len(keys) == 0 {
			jww.ERROR.Println("keys is empty ", "path", path)
			return
		}

		clientCall(ipcEndpoint, &result, "miner_mining", name, keys)
		printJSON(result)
	},
}

var setExtraCmd = &cobra.Command{
	Use:   "setextra <extra>",
	Short: "Set the extra of the miner.",
	Long:  `Set the extra of the miner.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var result bool
		clientCall(ipcEndpoint, &result, "miner_setExtra", args[0])
		printJSON(result)
	},
}

func init() {
	RootCmd.AddCommand(minerCmd)
	minerCmd.AddCommand(startCmd, forceCmd, stopCmd, miningCmd, setCoinbaseCmd, setExtraCmd)
	minerCmd.PersistentFlags().StringVarP(&ipcEndpoint, "ipcpath", "i", defaultIPCEndpoint(params.ClientIdentifier), "IPC Endpoint path")
}
