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

package sdk

import (
	"fmt"

	"github.com/fractalplatform/fractal/rpc"
)

// API rpc api
type API struct {
	rpchost string
	client  *rpc.Client
}

// NewAPI create api interface
func NewAPI(rpchost string) *API {
	client, err := rpc.DialHTTP(rpchost)
	if err != nil {
		panic(fmt.Sprintf("dial http %v err %v", rpchost, err))
	}
	api := &API{}
	api.rpchost = rpchost
	api.client = client
	return api
}
