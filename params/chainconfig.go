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

package params

import (
	"encoding/json"
	"math/big"
)

// ChainConfig is the core config which determines the blockchain settings.
type ChainConfig struct {
	BootNodes        []string `json:"bootnodes"` // enode URLs of the P2P bootstrap nodes
	ChainID          *big.Int `json:"chainId"`   // chainId identifies the current chain and is used for replay protection
	ChainName        string   `json:"chainName"` // chain name
	ChainURL         string   `json:"chainUrl"`  // chain url
	SnapshotInterval uint64   `json:"snapshotInterval"`
	SysName          string   `json:"systemName"`  // system name
	AccountName      string   `json:"accountName"` // account name
	AssetName        string   `json:"assetName"`   // asset name
	DposName         string   `json:"dposName"`    // system name
	FeeName          string   `json:"feeName"`     //fee name
	SysToken         string   `json:"systemToken"` // system token
	SysTokenID       uint64   `json:"sysTokenID"`
	SysTokenDecimals uint64   `json:"sysTokenDecimal"`
	ReferenceTime    uint64   `json:"referenceTime"`
}

var DefaultChainconfig = &ChainConfig{
	BootNodes:        []string{},
	ChainID:          big.NewInt(1),
	ChainName:        "fractal",
	ChainURL:         "https://fractalproject.com",
	SnapshotInterval: 180000,
	SysName:          "fractalfounder",
	AccountName:      "fractalaccount",
	AssetName:        "fractalasset",
	DposName:         "fractaldpos",
	FeeName:          "fractalfee",
	SysToken:         "ftoken",
}

func (cfg *ChainConfig) Copy() *ChainConfig {
	bts, _ := json.Marshal(cfg)
	c := &ChainConfig{}
	json.Unmarshal(bts, c)
	return c
}
