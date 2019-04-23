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

// MinerStart start
func (api *API) MinerStart() (bool, error) {
	ret := false
	err := api.client.Call(&ret, "miner_start")
	return ret, err
}

// MinerStop stop
func (api *API) MinerStop() (bool, error) {
	ret := false
	err := api.client.Call(&ret, "miner_stop")
	return ret, err
}

// MinerMining mining
func (api *API) MinerMining() (bool, error) {
	ret := false
	err := api.client.Call(&ret, "miner_mining")
	return ret, err
}

// MinerSetExtra extra
func (api *API) MinerSetExtra(extra []byte) (bool, error) {
	ret := true
	err := api.client.Call(&ret, "miner_setExtra", extra)
	return ret, err
}

// MinerSetCoinbase coinbase
func (api *API) MinerSetCoinbase(name string, privKeys []string) (bool, error) {
	ret := true
	err := api.client.Call(&ret, "miner_setCoinbase", name, privKeys)
	return ret, err
}
