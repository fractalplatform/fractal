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
	"runtime"
	"strings"

	"github.com/fractalplatform/fractal/params"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show ft current version",
	Long:  `Show ft current version`,
	Run: func(cmd *cobra.Command, args []string) {
		version()
	},
}

func version() {
	fmt.Println(strings.Title(params.ClientIdentifier))
	gitCommit := params.GitCommit()
	if gitCommit != "" {
		fmt.Println("Git Commit:", gitCommit)
		fmt.Println("Version:", params.ArchiveVersion(gitCommit))
	} else {
		fmt.Println("Version:", params.Version)
	}
	fmt.Println("Architecture:", runtime.GOARCH)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("Operating System:", runtime.GOOS)
	fmt.Printf("GOPATH=%s\n", os.Getenv("GOPATH"))
	fmt.Printf("GOROOT=%s\n", runtime.GOROOT())
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
