// Copyright 2019 LiuBan Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	commit    = ""
	date      = ""
	goversion = ""
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current version",
	Long:  `Show current version`,
	Run: func(cmd *cobra.Command, args []string) {
		version()
	},
}

func version() {
	if commit != "" {
		fmt.Println("Git Commit:", commit)
	}
	fmt.Println("Version:", FullVersion())
	fmt.Println("Architecture:", runtime.GOARCH)
	if goversion != "" {
		fmt.Println("Go Version:", goversion)
	}
	fmt.Println("Operating System:", runtime.GOOS)
	if goPath := os.Getenv("GOPATH"); goPath != "" {
		fmt.Printf("GOPATH=%s\n", goPath)
	}
	fmt.Printf("GOROOT=%s\n", runtime.GOROOT())
}

// FullVersion returns the version.
func FullVersion() string {
	version := History.CurrentVersion().String()
	if commit != "" {
		version += "+commit." + commit
	}
	if date != "" {
		version += "+date." + date
	}
	return version
}
