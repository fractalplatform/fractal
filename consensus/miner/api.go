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
	"encoding/hex"
	"fmt"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rpc"
)

// API exposes dpos related methods for the RPC interface.
type API struct {
	miner *Miner
	chain consensus.IChainReader
}

func (api *API) Start() bool {
	api.miner.Start()
	return true
}

func (api *API) Stop() bool {
	api.miner.Stop()
	return true
}

func (api *API) Mining() bool {
	return api.miner.Mining()
}

func (api *API) SetCoinbase(name string, privKey string) error {
	bts, err := hex.DecodeString(privKey)
	if err != nil {
		return err
	}
	priv, err := crypto.ToECDSA(bts)
	if err != nil {
		return err
	}
	if !common.IsValidName(name) {
		return fmt.Errorf("invalid name %v", name)
	}
	api.miner.SetCoinbase(name, priv)
	return nil
}

func (api *API) SetExtra(extra []byte) error {
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("Extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize)
	}
	api.miner.SetExtra(extra)
	return nil
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
			Public: true,
		},
	}
	apis = append(apis, miner.worker.Engine().APIs(chain)...)
	return apis
}
