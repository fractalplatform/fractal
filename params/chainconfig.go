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

var globalChainConfig *ChainConfig

func init() {
	globalChainConfig = DefaultChainconfig
}

// ChainConfig is the core config which determines the blockchain settings.
type ChainConfig struct {
	BootNodes        []string `json:"bootnodes"`   // enode URLs of the P2P bootstrap nodes
	ChainID          *big.Int `json:"chainId"`     // chainId identifies the current chain and is used for replay protection
	ChainName        string   `json:"chainName"`   // chain name
	ChainURL         string   `json:"chainUrl"`    // chain url
	SysName          string   `json:"systemName"`  // system name
	AccountName      string   `json:"accountName"` // account name
	AssetName        string   `json:"assetName"`   // asset name
	ItemName         string   `json:"itemName"`    // item name
	DposName         string   `json:"dposName"`    // system name
	FeeName          string   `json:"feeName"`     //fee name
	SysToken         string   `json:"systemToken"` // system token
	SysTokenID       uint64   `json:"sysTokenID"`
	SysTokenDecimals uint64   `json:"sysTokenDecimal"`
	ReferenceTime    uint64   `json:"referenceTime"`
}

var DefaultChainconfig = &ChainConfig{
	BootNodes:   []string{},
	ChainID:     big.NewInt(1),
	ChainName:   "fractal",
	ChainURL:    "https://fractalproject.com",
	SysName:     "fractalfounder",
	AccountName: "fractalaccount",
	AssetName:   "fractalasset",
	DposName:    "fractaldpos",
	FeeName:     "fractalfee",
	SysToken:    "ftoken",
}

func SetGlobalChainConfig(cfg *ChainConfig) {
	globalChainConfig = cfg
}

func (cfg *ChainConfig) GetBootNodes() []string {
	return cfg.BootNodes
}

func BootNodes() []string {
	return globalChainConfig.GetBootNodes()
}

func (cfg *ChainConfig) GetChainID() *big.Int {
	return cfg.ChainID
}

func ChainID() *big.Int {
	return globalChainConfig.GetChainID()
}

func (cfg *ChainConfig) GetChainName() string {
	return cfg.ChainName
}

func ChainName() string {
	return globalChainConfig.GetChainName()
}

func (cfg *ChainConfig) GetChainURL() string {
	return cfg.ChainURL
}

func ChainURL() string {
	return globalChainConfig.GetChainURL()
}

func (cfg *ChainConfig) GetSysName() string {
	return cfg.SysName
}

func SysName() string {
	return globalChainConfig.GetSysName()
}

func (cfg *ChainConfig) GetAccountName() string {
	return cfg.AccountName
}

func AccountName() string {
	return globalChainConfig.GetAccountName()
}

func (cfg *ChainConfig) GetAssetName() string {
	return cfg.AssetName
}

func AssetName() string {
	return globalChainConfig.GetAssetName()
}

func (cfg *ChainConfig) GetDposName() string {
	return cfg.DposName
}

func DposName() string {
	return globalChainConfig.GetDposName()
}

func (cfg *ChainConfig) GetFeeName() string {
	return cfg.FeeName
}

func FeeName() string {
	return globalChainConfig.GetFeeName()
}

func (cfg *ChainConfig) GetSysToken() string {
	return cfg.SysToken
}

func SysToken() string {
	return globalChainConfig.GetSysToken()
}

func (cfg *ChainConfig) GetSysTokenID() uint64 {
	return cfg.SysTokenID
}

func SysTokenID() uint64 {
	return globalChainConfig.GetSysTokenID()
}

func (cfg *ChainConfig) GetSysTokenDecimals() uint64 {
	return cfg.SysTokenDecimals
}

func SysTokenDecimals() uint64 {
	return globalChainConfig.GetSysTokenDecimals()
}

func (cfg *ChainConfig) GetReferenceTime() uint64 {
	return cfg.ReferenceTime
}

func ReferenceTime() uint64 {
	return globalChainConfig.GetReferenceTime()
}

func (cfg *ChainConfig) Copy() *ChainConfig {
	bts, _ := json.Marshal(cfg)
	c := &ChainConfig{}
	json.Unmarshal(bts, c)
	return c
}
