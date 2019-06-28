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
	"sort"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/rpc"
)

// API exposes dpos related methods for the RPC interface.
type API struct {
	dpos  *Dpos
	chain consensus.IChainReader
}

// Info get dpos config info
func (api *API) Info() interface{} {
	return api.dpos.config
}

// Irreversible get irreversible info
func (api *API) Irreversible() interface{} {
	ret := map[string]interface{}{}
	ret["reversible"] = api.chain.CurrentHeader().Number.Uint64()
	ret["proposedIrreversible"] = api.dpos.CalcProposedIrreversible(api.chain, nil, false)
	ret["bftIrreversible"] = api.dpos.CalcBFTIrreversible()
	return ret
}

// NextValidCandidates next valid candidates
func (api *API) NextValidCandidates() (interface{}, error) {
	epoch, err := api.epoch(api.chain.CurrentHeader().Number.Uint64())
	if err != nil {
		return nil, err
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	return sys.GetState(epoch)
}

// Epoch get epoch number by height
func (api *API) Epoch(height uint64) (uint64, error) {
	return api.epoch(height)
}

// PrevEpoch get prev epoch number by epoch
func (api *API) PrevEpoch(epoch uint64) (uint64, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	state, err := api.chain.StateAt(api.chain.CurrentHeader().Root)
	if err != nil {
		return 0, err
	}
	pepoch, _, err := api.dpos.GetEpoch(state, 1, epoch)
	return pepoch, err
}

// NextEpoch get next epoch number by epoch
func (api *API) NextEpoch(epoch uint64) (uint64, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	state, err := api.chain.StateAt(api.chain.CurrentHeader().Root)
	if err != nil {
		return 0, err
	}
	nepoch, _, err := api.dpos.GetEpoch(state, 2, epoch)
	return nepoch, err
}

// CandidatesSize get candidates size
func (api *API) CandidatesSize(epoch uint64) (uint64, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return 0, err
	}
	return sys.CandidatesSize(epoch)
}

// Candidates get all candidates info
func (api *API) Candidates(epoch uint64, detail bool) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	candidates, err := sys.GetCandidates(epoch)
	if err != nil {
		return nil, err
	}
	sort.Sort(candidates)
	if detail {
		return candidates, nil
	}
	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.Name)
	}
	return names, nil
}

// Candidate get candidate info
func (api *API) Candidate(epoch uint64, name string) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	return sys.GetCandidate(epoch, name)
}

// VotersByCandidate get voters info of candidate
func (api *API) VotersByCandidate(epoch uint64, candidate string, detail bool) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	voters, err := sys.GetVotersByCandidate(epoch, candidate)
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
func (api *API) VotersByVoter(epoch uint64, voter string, detail bool) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	voters, err := sys.GetVotersByVoter(epoch, voter)
	if err != nil {
		return nil, err
	}
	if detail {
		return voters, nil
	}
	candidates := make([]string, 0, len(voters))
	for _, voter := range voters {
		candidates = append(candidates, voter.Candidate)
	}
	return candidates, nil
}

// AvailableStake get available stake that can vote candidate
func (api *API) AvailableStake(epoch uint64, voter string) (*big.Int, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	q, err := sys.getAvailableQuantity(epoch, voter)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Mul(q, sys.config.unitStake()), nil
}

// ValidCandidates get valid candidates
func (api *API) ValidCandidates(epoch uint64) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	gstate, err := sys.GetState(epoch)
	if err != nil {
		return nil, err
	}
	return sys.GetState(gstate.PreEpoch)
}

func (api *API) CandidatesInfoForBrowser(epoch uint64) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	gstate, err := sys.GetState(epoch)
	if err != nil {
		return nil, err
	}

	timestamp := sys.config.epochTimeStamp(gstate.PreEpoch)
	preGstate, err := sys.GetState(gstate.PreEpoch)
	if err != nil {
		return nil, err
	}

	candidateInfos := ArrayCandidateInfoForBrowser{}
	candidateInfos.Data = make([]*CandidateInfoForBrowser, len(preGstate.ActivatedCandidateSchedule))

	copyCandidate := make([]string, len(preGstate.ActivatedCandidateSchedule))
	var spare = 7
	var activate = 21
	for i, activatedCandidate := range preGstate.ActivatedCandidateSchedule {
		//backup
		copyCandidate[i] = activatedCandidate

		candidateInfo := &CandidateInfoForBrowser{}
		candidateInfo.Candidate = activatedCandidate

		tmp, err := sys.GetCandidate(gstate.PreEpoch, activatedCandidate)
		if err != nil {
			return nil, err
		}
		candidateInfo.Quantity = strconv.FormatInt(tmp.Quantity.Int64(), 10)
		candidateInfo.TotalQuantity = strconv.FormatInt(tmp.TotalQuantity.Int64(), 10)
		candidateInfo.Counter = tmp.Counter
		candidateInfo.ActualCounter = tmp.ActualCounter
		if i < activate {
			candidateInfo.Type = 1
		} else {
			candidateInfo.Type = 2
		}
		// candidateInfo.Type
		if balance, err := sys.GetBalanceByTime(activatedCandidate, timestamp); err != nil {
			log.Warn("CandidatesInfoForBrowser", "candidate", activatedCandidate, "ignore", err)
			return nil, err
		} else {
			candidateInfo.Holder = strconv.FormatInt(balance.Int64(), 10)
		}
		candidateInfos.Data[i] = candidateInfo
	}

	if len(preGstate.OffCandidateSchedule) > 14 {
		return nil, fmt.Errorf("OffCandidateSchedule count %d > 14", len(preGstate.OffCandidateSchedule))
	}
	for i := 0; i < len(preGstate.OffCandidateSchedule); i++ {
		if i < spare {
			j := i + activate
			candidateInfos.Data[preGstate.OffCandidateSchedule[i]].Type = 0
			candidateInfos.Data[j].Type = 1
		} else {
			candidateInfos.Data[preGstate.OffCandidateSchedule[i]].Type = 0
		}
	}
	return candidateInfos, nil
}

// SnapShotTime get snapshot timestamp
func (api *API) SnapShotTime(epoch uint64) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	timestamp := sys.config.epochTimeStamp(epoch)
	gstate, err := sys.GetState(epoch)
	if err != nil {
		return nil, err
	}
	if sys.config.epoch(sys.config.ReferenceTime) == gstate.PreEpoch {
		timestamp = sys.config.epochTimeStamp(gstate.PreEpoch)
	}
	res := map[string]interface{}{}
	res["epoch"] = epoch
	res["timestamp"] = timestamp
	res["time"] = time.Unix(int64(timestamp/uint64(time.Second)), int64(timestamp%uint64(time.Second)))
	return res, nil
}

func (api *API) epoch(number uint64) (uint64, error) {
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

// GetActivedCandidateSize get actived candidate size
func (api *API) GetActivedCandidateSize(epoch uint64) (uint64, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	state, err := api.chain.StateAt(api.chain.CurrentHeader().Root)
	if err != nil {
		return 0, err
	}
	return api.dpos.GetActivedCandidateSize(state, epoch)
}

// GetActivedCandidate get actived candidate info
func (api *API) GetActivedCandidate(epoch uint64, index uint64) (interface{}, error) {
	if epoch == 0 {
		epoch, _ = api.epoch(api.chain.CurrentHeader().Number.Uint64())
	}
	state, err := api.chain.StateAt(api.chain.CurrentHeader().Root)
	if err != nil {
		return nil, err
	}
	candidate, delegated, voted, scounter, acounter, rindex, isbad, err := api.dpos.GetActivedCandidate(state, epoch, index)
	if err != nil {
		return nil, err
	}
	ret := map[string]interface{}{}
	ret["epoch"] = epoch
	ret["candidate"] = candidate
	ret["delegatedStake"] = delegated
	ret["votedStake"] = voted
	ret["shouldCount"] = scounter
	ret["actualCount"] = acounter
	ret["replaceIndex"] = rindex
	ret["bad"] = isbad
	return ret, nil
}
