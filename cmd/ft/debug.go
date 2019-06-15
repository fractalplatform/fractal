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
	"runtime"
	"runtime/debug"

	"github.com/fractalplatform/fractal/params"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Offers and API for debug pprof",
	Long:  `Offers and API for debug pprof`,
	Args:  cobra.NoArgs,
}

var memStatsCmd = &cobra.Command{
	Use:   "memstats ",
	Short: "Returns detailed runtime memory statistics.",
	Long:  `Returns detailed runtime memory statistics.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var result = new(runtime.MemStats)
		clientCall(ipcEndpoint, &result, "debug_memStats")
		printJSON(result)
	},
}

var gcStatsCmd = &cobra.Command{
	Use:   "gcstats ",
	Short: "Returns GC statistics.",
	Long:  `Returns GC statistics.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var result = new(debug.GCStats)
		clientCall(ipcEndpoint, &result, "debug_gcStats")
		printJSON(result)
	},
}

var cpuProfileCmd = &cobra.Command{
	Use:   "cpuprofile <file> <nsec> ",
	Short: "Turns on CPU profiling for nsec seconds and writesprofile data to file.",
	Long:  `Turns on CPU profiling for nsec seconds and writesprofile data to file.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clientCall(ipcEndpoint, nil, "debug_cpuProfile", args[0], parseUint64(args[1]))
		printJSON(true)
	},
}

var goTraceCmd = &cobra.Command{
	Use:   "gotrace <file> <nsec> ",
	Short: "Turns on CPU profiling for nsec seconds and writesprofile data to file.",
	Long:  `Turns on CPU profiling for nsec seconds and writesprofile data to file.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clientCall(ipcEndpoint, nil, "debug_goTrace", args[0], parseUint64(args[1]))
		printJSON(true)
	},
}

var blockProfileCmd = &cobra.Command{
	Use:   "blockprofile <file> <nsec> ",
	Short: "Turns on goroutine profiling for nsec seconds and writes profile data to file.",
	Long:  `Turns on goroutine profiling for nsec seconds and writes profile data to file.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clientCall(ipcEndpoint, nil, "debug_blockProfile", args[0], parseUint64(args[1]))
		printJSON(true)
	},
}

var mutexProfileCmd = &cobra.Command{
	Use:   "mutexprofile <file> <nsec> ",
	Short: "Turns on mutex profiling for nsec seconds and writes profile data to file.",
	Long:  `Turns on mutex profiling for nsec seconds and writes profile data to file.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clientCall(ipcEndpoint, nil, "debug_mutexProfile", args[0], parseUint64(args[1]))
		printJSON(true)
	},
}

var writeMemProfileCmd = &cobra.Command{
	Use:   "writememprofile <file>",
	Short: "Writes an allocation profile to the given file.",
	Long:  `Writes an allocation profile to the given file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clientCall(ipcEndpoint, nil, "debug_writeMemProfile", args[0])
		printJSON(true)
	},
}

var stacksCmd = &cobra.Command{
	Use:   "stacks",
	Short: "Returns a printed representation of the stacks of all goroutines.",
	Long:  `Returns a printed representation of the stacks of all goroutines.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var result []byte
		clientCall(ipcEndpoint, &result, "debug_stacks")
		fmt.Println(string(result))
	},
}

var freeOSMemoryCmd = &cobra.Command{
	Use:   "freeosmemory ",
	Short: "Returns unused memory to the OS.",
	Long:  `Returns unused memory to the OS.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientCall(ipcEndpoint, nil, "debug_freeOSMemory")
		printJSON(true)
	},
}

func init() {
	RootCmd.AddCommand(debugCmd)
	debugCmd.AddCommand(memStatsCmd, gcStatsCmd, cpuProfileCmd, goTraceCmd, blockProfileCmd,
		mutexProfileCmd, writeMemProfileCmd, stacksCmd, freeOSMemoryCmd)
	debugCmd.PersistentFlags().StringVarP(&ipcEndpoint, "ipcpath", "i", defaultIPCEndpoint(params.ClientIdentifier), "IPC Endpoint path")
}
