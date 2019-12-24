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

package miner

import (
	"github.com/fractalplatform/fractal/rpc"
)

// API exposes miner related methods for the RPC interface.
type API struct {
	miner *Miner
}

// Start start mining
func (api *API) Start() bool {
	return api.miner.Start()
}

// Stop stop mining
func (api *API) Stop() bool {
	return api.miner.Stop()
}

// Mining get miner's status
func (api *API) Mining() bool {
	return api.miner.Mining()
}

// SetCoinbase bind miner name & privkey of node
func (api *API) SetCoinbase(name string, privKey string) error {
	return api.miner.SetCoinbase(name, privKey)
}

// SetDelay delay broacast block when mint block
func (api *API) SetDelay(delayDuration uint64) error {
	return api.miner.SetDelayDuration(delayDuration)
}

// SetExtra set extra data for miner
func (api *API) SetExtra(extra string) error {
	return api.miner.SetExtra([]byte(extra))
}

// APIs provide the miner RPC API.
func (miner *Miner) APIs() []rpc.API {
	apis := []rpc.API{
		{
			Namespace: "miner",
			Version:   "1.0",
			Service: &API{
				miner: miner,
			},
		},
	}
	//apis = append(apis, miner.worker.Engine().APIs(chain)...)
	return apis
}
