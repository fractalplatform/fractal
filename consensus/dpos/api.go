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
	"time"

	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/rpc"
)

// API exposes dpos related methods for the RPC interface.
type API struct {
	dpos  *Dpos
	chain consensus.IChainReader
}

// Info get dpos info
func (api *API) Info() interface{} {
	return api.dpos.config
}

// IrreversibleRet result
type IrreversibleRet struct {
	ProposedIrreversible uint64 `json:"proposedIrreversible"`
	BftIrreversible      uint64 `json:"bftIrreversible"`
	Reversible           uint64 `json:"reversible"`
}

// Irreversible get irreversible info
func (api *API) Irreversible() interface{} {
	ret := &IrreversibleRet{}
	ret.Reversible = api.chain.CurrentHeader().Number.Uint64()
	ret.ProposedIrreversible = api.dpos.CalcProposedIrreversible(api.chain, nil, false)
	ret.BftIrreversible = api.dpos.CalcBFTIrreversible()
	return ret
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

// CandidateByHeight get candidate info of dpos
func (api *API) CandidateByHeight(height uint64, name string) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	epcho, err := api.epcho(height)
	if err != nil {
		return nil, err
	}
	return sys.GetCandidateInfoByTime(name, api.dpos.config.epochTimeStamp(epcho))
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

	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.Name)
	}
	return names, nil
}

// VotersByCandidate get voters info of candidate
func (api *API) VotersByCandidate(candidate string, detail bool) (interface{}, error) {
	number := api.chain.CurrentHeader().Number.Uint64()
	return api.VotersByCandidateByNumber(number, candidate, detail)
}

// VotersByCandidateByNumber get voters info of candidate
func (api *API) VotersByCandidateByNumber(number uint64, candidate string, detail bool) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	epcho, err := api.epcho(number)
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

	names := make([]string, 0, len(voters))
	for _, voter := range voters {
		names = append(names, voter.Name)
	}
	return names, nil
}

// VotersByVoter get voters info of voter
func (api *API) VotersByVoter(voter string, detail bool) (interface{}, error) {
	number := api.chain.CurrentHeader().Number.Uint64()
	return api.VotersByVoterByNumber(number, voter, detail)
}

// VotersByVoterByNumber get voters info of voter
func (api *API) VotersByVoterByNumber(number uint64, voter string, detail bool) (interface{}, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	epcho, err := api.epcho(number)
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
	number := api.chain.CurrentHeader().Number.Uint64()
	return api.AvailableStakeByNumber(number, voter)
}

// AvailableStakeByNumber get available stake
func (api *API) AvailableStakeByNumber(number uint64, voter string) (*big.Int, error) {
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	epcho, err := api.epcho(number)
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
	number := api.chain.CurrentHeader().Number.Uint64()
	return api.ValidCandidatesByNumber(number)
}

// ValidCandidatesByNumber valid candidates
func (api *API) ValidCandidatesByNumber(number uint64) (interface{}, error) {
	epcho, err := api.epcho(number)
	if err != nil {
		return nil, err
	}
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

// NextValidCandidates current valid candidates
func (api *API) NextValidCandidates() (interface{}, error) {
	number := api.chain.CurrentHeader().Number.Uint64()
	return api.NextValidCandidatesByNumber(number)
}

// NextValidCandidatesByNumber current valid candidates
func (api *API) NextValidCandidatesByNumber(number uint64) (interface{}, error) {
	epcho, err := api.epcho(number)
	if err != nil {
		return nil, err
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	return sys.GetState(epcho)
}

// SnapShotTime get snapshort
func (api *API) SnapShotTime() (interface{}, error) {
	number := api.chain.CurrentHeader().Number.Uint64()
	return api.SnapShotTimeByNumber(number)
}

// SnapShotTimeByNumber get snapshort by number
func (api *API) SnapShotTimeByNumber(number uint64) (interface{}, error) {
	epcho, err := api.epcho(number)
	if err != nil {
		return nil, err
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	timestamp := sys.config.epochTimeStamp(epcho)
	gstate, err := sys.GetState(epcho)
	if err != nil {
		return nil, err
	}
	if sys.config.epoch(sys.config.ReferenceTime) == gstate.PreEpcho {
		timestamp = sys.config.epochTimeStamp(gstate.PreEpcho)
	}
	res := map[string]interface{}{}
	res["timestamp"] = timestamp
	res["time"] = time.Unix(int64(timestamp/uint64(time.Second)), int64(timestamp%uint64(time.Second)))
	return res, nil
}

func (api *API) epcho(number uint64) (uint64, error) {
	header := api.chain.GetHeaderByNumber(number)
	if header == nil {
		return 0, fmt.Errorf("not found number %v", number)
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
