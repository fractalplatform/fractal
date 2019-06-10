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
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/rpc"
)

// API exposes dpos related methods for the RPC interface.
type API struct {
	miner *Miner
	chain consensus.IChainReader
}

func (api *API) Start() bool {
	return api.miner.Start(false)
}

func (api *API) Force() bool {
	return api.miner.Start(true)
}

func (api *API) Stop() bool {
	return api.miner.Stop()
}

func (api *API) Mining() bool {
	return api.miner.Mining()
}

func (api *API) SetCoinbase(name string, privKeys []string) error {
	return api.miner.SetCoinbase(name, privKeys)
}

func (api *API) SetDelay(delayDuration uint64) error {
	return api.miner.SetDelayDuration(delayDuration)
}

func (api *API) SetExtra(extra string) error {
	return api.miner.SetExtra([]byte(extra))
}

func (miner *Miner) APIs(chain consensus.IChainReader) []rpc.API {
	apis := []rpc.API{
		{
			Namespace: "miner",
			Version:   "1.0",
			Service: &API{
				miner: miner,
				chain: chain,
			},
		},
	}
	apis = append(apis, miner.worker.Engine().APIs(chain)...)
	return apis
}
