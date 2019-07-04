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

func (api *API) BrowserAllEpoch() (interface{}, error) {
	epochs := Epochs{}
	epochs.Data = make([]*Epoch, 0)
	epochNumber, _ := api.epoch(api.chain.CurrentHeader().Number.Uint64())
	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	for {
		data := &Epoch{}
		timestamp := sys.config.epochTimeStamp(epochNumber)
		gstate, err := sys.GetState(epochNumber)
		if err != nil {
			return nil, err
		}
		if sys.config.epoch(sys.config.ReferenceTime) == gstate.PreEpoch {
			timestamp = sys.config.epochTimeStamp(gstate.PreEpoch)
		}

		data.Start = timestamp / 1000000000
		data.Epoch = epochNumber
		epochs.Data = append(epochs.Data, data)
		if epochNumber == 1 {
			break
		}
		// log.Info("BrowserAllEpoch1", "number", epochNumber)
		epochNumber = gstate.PreEpoch
		// log.Info("BrowserAllEpoch2", "number", epochNumber)
	}
	return epochs, nil
}

func (api *API) BrowserEpochRecord(reqEpochNumber uint64) (interface{}, error) {
	var req, data uint64
	if reqEpochNumber == 0 {
		log.Warn("BrowserAccounting 0")
		return nil, fmt.Errorf("request:0")
	}

	vote, _ := api.epoch(api.chain.CurrentHeader().Number.Uint64())
	if reqEpochNumber > vote {
		log.Warn("BrowserAccounting", " request:", reqEpochNumber, "> vote:", vote)
		return nil, fmt.Errorf("request:%d > vote:%d", reqEpochNumber, vote)
	}
	req = reqEpochNumber

	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	log.Info("BrowserAccounting", "req epoch:", req)
	reqEpoch, err := sys.GetState(req)
	if err != nil {
		return nil, err
	}

	data = reqEpoch.PreEpoch
	log.Info("BrowserAccounting", "data epoch:", data)
	timestamp := sys.config.epochTimeStamp(data)
	dataEpoch, err := sys.GetState(data)
	if err != nil {
		return nil, err
	}

	candidateInfos := ArrayCandidateInfoForBrowser{}
	candidateInfos.Data = make([]*CandidateInfoForBrowser, len(dataEpoch.ActivatedCandidateSchedule))

	var spare = 7
	var activate = 21
	for i, activatedCandidate := range dataEpoch.ActivatedCandidateSchedule {
		candidateInfo := &CandidateInfoForBrowser{}
		candidateInfo.Candidate = activatedCandidate

		curtmp, err := sys.GetCandidate(req, activatedCandidate)
		if err != nil {
			log.Warn("BrowserAccounting Cur Candidate:", req, " Data not found")
			return nil, fmt.Errorf("Cur Candidate:%d Data not found", req)
		}
		candidateInfo.NowCounter = curtmp.Counter
		candidateInfo.NowActualCounter = curtmp.ActualCounter

		tmp, err := sys.GetCandidate(data, activatedCandidate)
		if err != nil {
			return nil, err
		}
		candidateInfo.Quantity = tmp.Quantity.Mul(tmp.Quantity, api.dpos.config.unitStake()).String()
		candidateInfo.TotalQuantity = tmp.TotalQuantity.String()
		candidateInfo.Counter = tmp.Counter
		candidateInfo.ActualCounter = tmp.ActualCounter
		if i < activate {
			candidateInfo.Status = 1
		} else {
			candidateInfo.Status = 2
		}
		// candidateInfo.Status
		if balance, err := sys.GetBalanceByTime(activatedCandidate, timestamp); err != nil {
			log.Warn("BrowserAccounting", "candidate", activatedCandidate, "ignore", err)
			return nil, err
		} else {
			candidateInfo.Holder = balance.String()
		}
		candidateInfos.Data[i] = candidateInfo
	}

	fmt.Println(dataEpoch.BadCandidateIndexSchedule, len(candidateInfos.Data))
	if len(dataEpoch.BadCandidateIndexSchedule) > 14 {
		log.Warn("BrowserAccounting BadCandidateIndexSchedule > 14", "epoch", data)
		return nil, fmt.Errorf("count %d > 14", len(dataEpoch.BadCandidateIndexSchedule))
	}
	for i := 0; i < len(dataEpoch.BadCandidateIndexSchedule); i++ {
		j := i + activate
		if i < spare && len(candidateInfos.Data) > j {
			// log.Info("***** i", "", i)
			// log.Info("***** j", "", j)
			// log.Info("***** len", "", len(dataEpoch.BadCandidateIndexSchedule))
			candidateInfos.Data[dataEpoch.BadCandidateIndexSchedule[i]].Status = 0
			// log.Info("*********")
			candidateInfos.Data[j].Status = 1
			// log.Info("&&&&&&&&&")
		} else {
			candidateInfos.Data[dataEpoch.BadCandidateIndexSchedule[i]].Status = 0
		}
	}
	return candidateInfos, nil
}

func (api *API) BrowserVote() (interface{}, error) {
	vote, _ := api.epoch(api.chain.CurrentHeader().Number.Uint64())

	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	log.Info("BrowserVote", "vote:", vote)
	reqEpoch, err := sys.GetState(vote)
	if err != nil {
		return nil, err
	}

	//history miner rate
	prereqEpoch, err := sys.GetState(reqEpoch.PreEpoch)
	if err != nil {
		return nil, err
	}

	history := prereqEpoch.PreEpoch
	log.Info("BrowserVote", "history:", history)
	timestamp := sys.config.epochTimeStamp(vote)

	candidateInfos := ArrayCandidateInfoForBrowser{}
	candidateInfos.Data = make([]*CandidateInfoForBrowser, len(reqEpoch.ActivatedCandidateSchedule))

	for i, activatedCandidate := range reqEpoch.ActivatedCandidateSchedule {
		candidateInfo := &CandidateInfoForBrowser{}
		candidateInfo.Candidate = activatedCandidate

		tmp, err := sys.GetCandidate(vote, activatedCandidate)
		if err != nil {
			return nil, err
		}
		// candidateInfo.NowCounter = tmp.Counter
		// candidateInfo.NowActualCounter = tmp.ActualCounter

		candidateInfo.Quantity = tmp.Quantity.Mul(tmp.Quantity, api.dpos.config.unitStake()).String()
		candidateInfo.TotalQuantity = tmp.TotalQuantity.String()

		// candidateInfo.Type
		if balance, err := sys.GetBalanceByTime(activatedCandidate, timestamp); err != nil {
			log.Warn("Accounting", "candidate", activatedCandidate, "ignore", err)
			return nil, err
		} else {
			candidateInfo.Holder = balance.String()
		}

		histmp, err := sys.GetCandidate(history, activatedCandidate)
		if err != nil {
			return nil, err
		}
		candidateInfo.Counter = histmp.Counter
		candidateInfo.ActualCounter = histmp.ActualCounter
		candidateInfos.Data[i] = candidateInfo
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
