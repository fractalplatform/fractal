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
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/rpc"
)

// API exposes dpos related methods for the RPC interface.
type API struct {
	dpos  *Dpos
	chain consensus.IChainReader
}

// Info get dpos info
func (api *API) Info() (interface{}, error) {
	return api.dpos.config, nil
}

// IrreversibleRet result
type IrreversibleRet struct {
	ProposedIrreversible uint64 `json:"proposedIrreversible"`
	BftIrreversible      uint64 `json:"bftIrreversible"`
	Reversible           uint64 `json:"reversible"`
}

// Irreversible get irreversible info
func (api *API) Irreversible() (interface{}, error) {
	ret := &IrreversibleRet{}
	ret.Reversible = api.chain.CurrentHeader().Number.Uint64()
	ret.ProposedIrreversible = api.dpos.CalcProposedIrreversible(api.chain, nil, false)
	ret.BftIrreversible = api.dpos.CalcBFTIrreversible()
	return ret, nil
}

// Candidate get candidate info of dpos
func (api *API) Candidate(name string) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	if prod, err := sys.GetCandidate(name); err != nil {
		return nil, err
	} else if prod != nil {
		return prod, err
	}
	return nil, nil
}

// Candidates all candidates info
func (api *API) Candidates(detail bool) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	candidates, err := sys.GetCandidates()
	if err != nil || len(candidates) == 0 {
		return nil, err
	}
	if detail {
		return candidates, nil
	}

	names := []string{}
	for _, candidate := range candidates {
		names = append(names, candidate.Name)
	}
	return names, nil
}

// VotersByCandidate get voters info of candidate
func (api *API) VotersByCandidate(candidate string, detail bool) (interface{}, error) {
	height := api.chain.CurrentHeader().Number.Uint64()
	return api.VotersByCandidateByHeight(height, candidate, detail)
}

// VotersByCandidateByHeight get voters info of candidate
func (api *API) VotersByCandidateByHeight(height uint64, candidate string, detail bool) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	epcho, err := api.epcho(height)
	if err != nil {
		return nil, err
	}
	voters, err := sys.GetVotersByCandidate(epcho, candidate)
	if err != nil {
		return nil, err
	}
	if detail {
		return voters, nil
	}

	names := []string{}
	for _, voter := range voters {
		names = append(names, voter.Name)
	}
	return names, nil
}

// VotersByVoter get voters info of voter
func (api *API) VotersByVoter(voter string, detail bool) (interface{}, error) {
	height := api.chain.CurrentHeader().Number.Uint64()
	return api.VotersByVoterByHeight(height, voter, detail)
}

// VotersByVoterByHeight get voters info of voter
func (api *API) VotersByVoterByHeight(height uint64, voter string, detail bool) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	epcho, err := api.epcho(height)
	if err != nil {
		return nil, err
	}
	voters, err := sys.GetVotersByVoter(epcho, voter)
	if err != nil {
		return nil, err
	}
	if detail {
		return voters, nil
	}

	candidates := []string{}
	for _, voter := range voters {
		candidates = append(candidates, voter.Candidate)
	}
	return candidates, nil
}

// AvailableStake get available stake
func (api *API) AvailableStake(voter string) (*big.Int, error) {
	height := api.chain.CurrentHeader().Number.Uint64()
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	epcho, err := api.epcho(height)
	if err != nil {
		return nil, err
	}
	q, err := sys.getAvailableQuantity(epcho, voter)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Mul(q, sys.config.unitStake()), nil
}

// ValidCandidates current valid candidates
func (api *API) ValidCandidates() (interface{}, error) {
	height := api.chain.CurrentHeader().Number.Uint64()
	return api.ValidCandidatesByHeight(height)
}

// ValidCandidatesByHeight valid candidates
func (api *API) ValidCandidatesByHeight(height uint64) (interface{}, error) {
	epcho, err := api.epcho(height)
	if err != nil {
		return nil, err
	}
	return api.validCandidates(epcho)
}

func (api *API) validCandidates(epcho uint64) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	gstate, err := sys.GetState(epcho)
	if err != nil {
		return nil, err
	}
	return sys.GetState(gstate.PreEpcho)
}

func (api *API) epcho(height uint64) (uint64, error) {
	header := api.chain.GetHeaderByNumber(height)
	if header == nil {
		return 0, fmt.Errorf("not found height %v", height)
	}
	timestamp := header.Time.Uint64()
	return api.dpos.config.epoch(timestamp), nil
}

func (api *API) system() (*System, error) {
	state, err := api.chain.StateAt(api.chain.CurrentHeader().Root)
	if err != nil {
		return nil, err
	}
	sys := NewSystem(state, api.dpos.config)
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
