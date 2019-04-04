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

package dpos

import (
	"encoding/json"
	"math/big"
	"sort"

	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/rpc"
)

// API exposes dpos related methods for the RPC interface.
type API struct {
	dpos  *Dpos
	chain consensus.IChainReader
}

type Irreversible_Ret struct {
	ProposedIrreversible uint64 `json:"proposedIrreversible"`
	BftIrreversible      uint64 `json:"bftIrreversible"`
	Reversible           uint64 `json:"reversible"`
}

func (api *API) Info() (interface{}, error) {
	return api.dpos.config, nil
}

func (api *API) Irreversible() (interface{}, error) {
	ret := &Irreversible_Ret{}
	ret.Reversible = api.chain.CurrentHeader().Number.Uint64()
	ret.ProposedIrreversible = api.dpos.CalcProposedIrreversible(api.chain, false)
	ret.BftIrreversible = api.dpos.CalcBFTIrreversible()

	return ret, nil
}

func (api *API) Account(name string) (interface{}, error) {

	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	if vote, err := sys.GetVoter(name); err != nil {
		return nil, err
	} else if vote != nil {
		return vote, err
	}

	if prod, err := sys.GetCadidate(name); err != nil {
		return nil, err
	} else if prod != nil {
		return prod, err
	}

	return nil, nil
}

func (api *API) Cadidates() ([]map[string]interface{}, error) {
	pfileds := []map[string]interface{}{}

	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	cadidates, err := sys.Cadidates()
	if err != nil || len(cadidates) == 0 {
		return pfileds, err
	}

	prods := cadidateInfoArray{}
	for _, prod := range cadidates {
		prods = append(prods, prod)
	}
	sort.Sort(prods)
	for _, prod := range prods {
		fields := map[string]interface{}{}
		cjson, err := json.Marshal(prod)
		if err != nil {
			return pfileds, err
		}
		if err := json.Unmarshal(cjson, &fields); err != nil {
			return pfileds, err
		}
		pfileds = append(pfileds, fields)
	}

	return pfileds, nil
}

func (api *API) Epcho(height uint64) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	if gstate, err := sys.GetState(height); err != nil {
		return nil, err
	} else if gstate != nil {
		return gstate, err
	}

	return nil, nil
}

func (api *API) LatestEpcho() (interface{}, error) {
	return api.Epcho(api.chain.CurrentHeader().Number.Uint64())
}

func (api *API) ValidateEpcho() (interface{}, error) {
	curHeader := api.chain.CurrentHeader()
	targetTS := big.NewInt(curHeader.Time.Int64() - int64(api.dpos.config.DelayEcho*api.dpos.config.epochInterval()))
	for curHeader.Number.Uint64() > 0 {
		if curHeader.Time.Cmp(targetTS) == -1 {
			break
		}
		curHeader = api.chain.GetHeaderByHash(curHeader.ParentHash)
	}
	return api.Epcho(curHeader.Number.Uint64() + 1)
}

func (api *API) system() (*System, error) {
	state, err := api.chain.StateAt(api.chain.CurrentHeader().Root)
	if err != nil {
		return nil, err
	}
	sys := &System{
		config: api.dpos.config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:  api.dpos.config.AccountName,
				state: state,
			},
		},
	}
	return sys, nil
}

// APIs returning the user facing RPC APIs.
func (dpos *Dpos) APIs(chain consensus.IChainReader) []rpc.API {
	return []rpc.API{
		{
			Namespace: "dpos",
			Version:   "1.0",
			Service: &API{
				dpos:  dpos,
				chain: chain,
			},
			Public: true,
		},
	}
}
