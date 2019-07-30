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
	"github.com/fractalplatform/fractal/common"
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
		epochNumber = gstate.PreEpoch
	}
	return epochs, nil
}

func (api *API) BrowserEpochRecord(reqEpochNumber uint64) (interface{}, error) {
	bstart := time.Now()
	var req, data uint64
	if reqEpochNumber == 0 {
		log.Warn("BrowserEpochRecord 0")
		return nil, fmt.Errorf("request:0")
	}

	vote, _ := api.epoch(api.chain.CurrentHeader().Number.Uint64())
	if reqEpochNumber > vote {
		log.Warn("BrowserEpochRecord", " request:", reqEpochNumber, "> vote:", vote)
		return nil, fmt.Errorf("request:%d > vote:%d", reqEpochNumber, vote)
	}
	req = reqEpochNumber

	sys, err := api.system()
	if err != nil {
		return nil, err
	}
	log.Info("BrowserEpochRecord", "req epoch:", req)
	reqEpoch, err := sys.GetState(req)
	if err != nil {
		return nil, err
	}

	data = reqEpoch.PreEpoch
	log.Info("BrowserEpochRecord", "data epoch:", data)
	timestamp := sys.config.epochTimeStamp(data)
	dataEpoch, err := sys.GetState(data)
	if err != nil {
		return nil, err
	}

	candidateInfos := ArrayCandidateInfoForBrowser{}
	candidateInfos.Data = make([]*CandidateInfoForBrowser, 0)

	candidateInfos.TakeOver = dataEpoch.TakeOver
	candidateInfos.Dpos = dataEpoch.Dpos

	if dataEpoch.Dpos {
		for _, activatedCandidate := range dataEpoch.ActivatedCandidateSchedule {
			if activatedCandidate == "fractal.founder" {
				continue
			}
			candidateInfo := &CandidateInfoForBrowser{}
			candidateInfo.Candidate = activatedCandidate

			curtmp, err := sys.GetCandidate(req, activatedCandidate)
			if err != nil {
				log.Warn("BrowserEpochRecord Cur Candidate:", req, " Data not found")
				return nil, fmt.Errorf("Cur Candidate:%d Data not found", req)
			}
			if curtmp != nil {
				candidateInfo.NowCounter = curtmp.Counter
				candidateInfo.NowActualCounter = curtmp.ActualCounter
			}

			tmp, err := sys.GetCandidate(data, activatedCandidate)
			if err != nil {
				return nil, err
			}
			if tmp != nil {
				candidateInfo.Quantity = tmp.Quantity.Mul(tmp.Quantity, api.dpos.config.unitStake()).String()
				candidateInfo.TotalQuantity = tmp.TotalQuantity.String()
				candidateInfo.Counter = tmp.Counter
				candidateInfo.ActualCounter = tmp.ActualCounter
			}

			var balance *big.Int
			if balance, err = sys.GetBalanceByTime(activatedCandidate, timestamp); err != nil {
				if err.Error() != "Not snapshot info, error = EOF" {
					log.Warn("BrowserEpochRecord", "candidate", activatedCandidate, "ignore", err)
					return nil, err
				}
			}
			candidateInfo.Holder = balance.String()

			candidateInfos.Data = append(candidateInfos.Data, candidateInfo)
		}

		if len(dataEpoch.UsingCandidateIndexSchedule) == 0 {
			usingCandidateIndexSchedule := []uint64{}
			for index := range dataEpoch.ActivatedCandidateSchedule {
				if uint64(index) >= sys.config.CandidateScheduleSize {
					break
				}
				usingCandidateIndexSchedule = append(usingCandidateIndexSchedule, uint64(index))
			}
			for index, offset := range dataEpoch.BadCandidateIndexSchedule {
				usingCandidateIndexSchedule[int(offset)] = sys.config.CandidateScheduleSize + uint64(index)
			}
			dataEpoch.UsingCandidateIndexSchedule = usingCandidateIndexSchedule
		}

		candidateInfos.BadCandidateIndexSchedule = dataEpoch.BadCandidateIndexSchedule
		candidateInfos.UsingCandidateIndexSchedule = dataEpoch.UsingCandidateIndexSchedule
	}

	log.Info("BrowserEpochRecord", "elapsed", common.PrettyDuration(time.Since(bstart)).String())
	return candidateInfos, nil
}

func (api *API) BrowserVote(reqEpochNumber uint64) (interface{}, error) {
	var req, history uint64
	bstart := time.Now()
	if reqEpochNumber == 0 {
		log.Warn("BrowserVote 0")
		return nil, fmt.Errorf("request:0")
	}
	vote, _ := api.epoch(api.chain.CurrentHeader().Number.Uint64())
	if reqEpochNumber > vote {
		log.Warn("BrowserVote", " request:", reqEpochNumber, "> vote:", vote)
		return nil, fmt.Errorf("request:%d > vote:%d", reqEpochNumber, vote)
	}
	req = reqEpochNumber
	log.Info("BrowserVote", "req:", req)
	sys, err := api.system()
	if err != nil {
		return nil, err
	}

	reqEpoch, err := sys.GetState(req)
	if err != nil {
		return nil, err
	}
	history = reqEpoch.PreEpoch
	log.Info("BrowserVote", "history:", history)
	timestamp := sys.config.epochTimeStamp(req)
	candidateInfos := ArrayCandidateInfoForBrowser{}
	candidates, err := sys.GetCandidates(req)

	if err != nil {
		return nil, err
	}
	sort.Sort(candidates)

	var declims uint64 = 1000000000000000000
	minQuantity := big.NewInt(0).Mul(api.dpos.config.CandidateAvailableMinQuantity, big.NewInt(0).SetUint64(declims))
	candidateInfos.Data = make([]*CandidateInfoForBrowser, 0)
	for _, c := range candidates {
		if c.Name == "fractal.founder" {
			continue
		}

		if c.Type == Freeze || c.Type == Black {
			continue
		}
		balance, err := sys.GetBalanceByTime(c.Name, timestamp)
		if err != nil {
			if err.Error() != "Not snapshot info, error = EOF" {
				log.Warn("BrowserVote", "candidate", c.Name, "ignore", err)
				return nil, err
			}
		}

		if balance.Cmp(minQuantity) < 0 {
			continue
		}

		candidateInfo := &CandidateInfoForBrowser{}
		candidateInfo.Candidate = c.Name

		candidateInfo.Quantity = c.Quantity.Mul(c.Quantity, api.dpos.config.unitStake()).String()
		candidateInfo.TotalQuantity = c.TotalQuantity.String()

		candidateInfo.Holder = balance.String()

		tmp, err := sys.GetCandidate(history, c.Name)
		if err != nil {
			return nil, err
		}

		if tmp != nil {
			candidateInfo.Counter = tmp.Counter
			candidateInfo.ActualCounter = tmp.ActualCounter
		}
		candidateInfos.Data = append(candidateInfos.Data, candidateInfo)
	}
	log.Info("BrowserVote", "elapsed", common.PrettyDuration(time.Since(bstart)).String())
	return candidateInfos, nil
}

// SnapShotStake get snapshot stake
func (api *API) SnapShotStake(epoch uint64, name string) (interface{}, error) {
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
	return sys.GetBalanceByTime(name, timestamp)
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
