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

import "math/big"

// DposInfo dpos info
func (api *API) DposInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_info")
	return info, err
}

// DposIrreversible dpos irreversible info
func (api *API) DposIrreversible() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_irreversible")
	return info, err
}

// DposNextValidCandidates dpos candidate info
func (api *API) DposNextValidCandidates() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_nextValidCandidates")
	return info, err
}

// DposEpoch dpos epoch info
func (api *API) DposEpoch(height uint64) (uint64, error) {
	epoch := uint64(0)
	err := api.client.Call(&epoch, "dpos_epoch", height)
	return epoch, err
}

// DposPrevEpoch dpos epoch info
func (api *API) DposPrevEpoch(epoch uint64) (uint64, error) {
	pepoch := uint64(0)
	err := api.client.Call(&pepoch, "dpos_prevEpoch", epoch)
	return epoch, err
}

// DposNextEpoch dpos epoch info
func (api *API) DposNextEpoch(epoch uint64) (uint64, error) {
	nepoch := uint64(0)
	err := api.client.Call(&nepoch, "dpos_nextEpoch", epoch)
	return epoch, err
}

// DposValidCandidates dpos candidate info
func (api *API) DposValidCandidates(epoch uint64) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_validCandidates", epoch)
	return info, err
}

// DposCandidatesSize candidates size
func (api *API) DposCandidatesSize(epoch uint64) (uint64, error) {
	size := uint64(0)
	err := api.client.Call(&size, "dpos_candidatesSize", epoch)
	return size, err
}

// DposCandidates candidate info by name
func (api *API) DposCandidates(epoch uint64, detail bool) ([]map[string]interface{}, error) {
	info := []map[string]interface{}{}
	err := api.client.Call(&info, "dpos_candidates", epoch, detail)
	return info, err
}

// DposCandidate candidate info by name
func (api *API) DposCandidate(epoch uint64, name string) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_candidate", epoch, name)
	return info, err
}

// DposVotersByCandidate get voters info of candidate
func (api *API) DposVotersByCandidate(epoch uint64, candidate string, detail bool) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_votersByCandidate", epoch, candidate, detail)
	return info, err
}

// DposVotersByVoter get voters info of voter
func (api *API) DposVotersByVoter(epoch uint64, voter string, detail bool) (interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_votersByVoter", epoch, voter, detail)
	return info, err
}

// DposAvailableStake state info
func (api *API) DposAvailableStake(epoch uint64, name string) (*big.Int, error) {
	stake := big.NewInt(0)
	err := api.client.Call(&stake, "dpos_availableStake", epoch, name)
	return stake, err
}

// DposSnapShotTime dpos snapshot time info
func (api *API) DposSnapShotTime(epoch uint64) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_snapShotTime", epoch)
	return info, err
}
