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
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/fractalplatform/fractal/ftservice"
	"github.com/fractalplatform/fractal/params"
	"github.com/spf13/cobra"
)

var rpcCmd = &cobra.Command{
	Use:   "rpc",
	Short: "rpc client",
	Long:  `rpc client`,
	Args:  cobra.NoArgs,
}

var paramsIn = make(map[string][]reflect.Type)

var rpcCall = func(method string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		params := make([]interface{}, len(args))
		paramsTyp := paramsIn[method]
		for i, arg := range args {
			if i < len(paramsTyp) {
				b := reflect.New(paramsTyp[i]).Interface()
				//var b bool
				if err := json.Unmarshal([]byte(arg), &b); err != nil {
					if err := json.Unmarshal([]byte("\""+arg+"\""), &b); err != nil {
						params[i] = arg
						continue
					}
				}
				params[i] = b
			} else {
				params[i] = arg
			}
		}

		result := clientCallRaw(ipcEndpoint, method, params...)
		printJSON(result)
	}
}

func init() {
	RootCmd.AddCommand(rpcCmd)
	ft := &ftservice.FtService{}
	ft.APIBackend = ftservice.NewAPIBackend(ft)
	apis := ft.APIs()
	contextTyp := reflect.TypeOf((*context.Context)(nil)).Elem()
	for _, api := range apis {
		srv := api.Service
		typ := reflect.TypeOf(srv)
		for i := 0; i < typ.NumMethod(); i++ {
			method := typ.Method(i)
			rpcMethod := strings.ToLower(api.Namespace) + "_" + strings.ToLower(method.Name[:1]) + method.Name[1:]
			rpcCmd.AddCommand(&cobra.Command{
				Use:   rpcMethod,
				Short: rpcMethod,
				Long:  rpcMethod,
				Run:   rpcCall(rpcMethod),
			})
			funcTyp := method.Type
			for i := 1; i < funcTyp.NumIn(); i++ {
				typ := funcTyp.In(i)
				if typ == contextTyp {
					continue
				}
				//fmt.Println(rpcMethod, typ)
				paramsIn[rpcMethod] = append(paramsIn[rpcMethod], typ)
			}
		}
	}
	rpcCmd.PersistentFlags().StringVarP(&ipcEndpoint, "ipcpath", "i", defaultIPCEndpoint(params.ClientIdentifier), "IPC Endpoint path")
}
