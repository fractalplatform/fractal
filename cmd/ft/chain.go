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

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
	"github.com/spf13/cobra"
)

var (
	chainCommand = &cobra.Command{
		Use:   "chain",
		Short: "Support blockchain pure state ",
		Long:  "Support blockchain pure state ",
		Args:  cobra.NoArgs,
	}

	statePureCommand = &cobra.Command{
		Use:   "startpure <enable/disable>",
		Short: "Start or stop pure state ",
		Long:  "Start or stop pure state",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := prueState(args[0]); err != nil {
				fmt.Println(err)
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(chainCommand)
	chainCommand.AddCommand(statePureCommand)
	statePureCommand.Flags().StringVarP(&ipcEndpoint, "ipcpath", "i", defaultIPCEndpoint(params.ClientIdentifier), "IPC Endpoint path")
}

func prueState(arg string) error {
	var enable bool

	switch arg {
	case "enable":
		enable = true
	case "disable":
	default:
		return fmt.Errorf("not support arg %v", arg)
	}

	result := new(types.BlockState)
	clientCall(ipcEndpoint, &result, "ft_setStatePruning", enable)
	printJSON(result)
	return nil
}
