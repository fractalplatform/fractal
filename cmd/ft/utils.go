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
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/ethereum/go-ethereum/common/fdlimit"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/node"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/naoina/toml"
	jww "github.com/spf13/jwalterweatherman"
)

// defaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func defaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "ft_ledger")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "ft_ledger")
		} else {
			return filepath.Join(home, ".ft_ledger")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// makeDatabaseHandles raises out the number of allowed file handles per process
// for ft and returns half of the allowance to assign to the database.
func makeDatabaseHandles() int {
	limit, err := fdlimit.Current()
	if err != nil {
		log.Error("Failed to retrieve file descriptor allowance: %v", err)
	}
	if limit < 2048 {
		if _, err := fdlimit.Raise(2048); err != nil {
			log.Error("Failed to raise file descriptor allowance: %v", err)
		}
	}
	if limit > 2048 { // cap database file descriptors even if more is available
		limit = 2048
	}
	return limit / 2 // Leave half for networking and other stuff
}

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

func clientCall(endpoint string, result interface{}, method string, args ...interface{}) {
	client, err := dialRPC(ipcEndpoint)
	if err != nil {
		jww.ERROR.Println(err)
		os.Exit(-1)
	}
	if err := client.Call(result, method, args...); err != nil {
		jww.ERROR.Println(err)
		os.Exit(-1)
	}
}

// dialRPC returns a RPC client which connects to the given endpoint.
func dialRPC(endpoint string) (*rpc.Client, error) {
	if endpoint == "" {
		endpoint = defaultIPCEndpoint(params.ClientIdentifier)
	}
	return rpc.Dial(endpoint)
}

// DefaultIPCEndpoint returns the IPC path used by default.
func defaultIPCEndpoint(clientIdentifier string) string {
	if clientIdentifier == "" {
		clientIdentifier = strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if clientIdentifier == "" {
			panic("empty executable name")
		}
	}
	config := &node.Config{DataDir: defaultDataDir(), IPCPath: clientIdentifier + ".ipc"}
	return config.IPCEndpoint()
}

func printJSON(data interface{}) {
	rawData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		jww.ERROR.Println(err)
		os.Exit(1)
	}
	jww.FEEDBACK.Println(string(rawData))
}

func printJSONList(data interface{}) {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Slice {
		jww.ERROR.Printf("invalid type %v assertion", value.Kind())
		os.Exit(1)
	}

	for idx := 0; idx < value.Len(); idx++ {
		jww.FEEDBACK.Println(idx, ":")
		rawData, err := json.MarshalIndent(value.Index(idx).Interface(), "", "  ")
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(1)
		}
		jww.FEEDBACK.Println(string(rawData))
	}
}

func parseUint64(arg string) uint64 {
	num, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		jww.ERROR.Printf("%v can not convert uint64,err: %v", arg, err)
		os.Exit(1)
	}
	return num
}

func parseBigInt(arg string) *big.Int {
	price, ok := big.NewInt(0).SetString(arg, 10)
	if !ok {
		jww.ERROR.Printf("%v can not convert big.Int", arg)
		os.Exit(1)
	}
	return price
}

func parseBool(arg string) bool {
	if arg == "true" {
		return true
	}
	return false
}
