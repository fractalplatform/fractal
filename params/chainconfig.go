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

const DefaultPubkeyHex = "047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"

// ChainConfig is the core config which determines the blockchain settings.
// ChainConfig is stored in the database on a per block basis.
type ChainConfig struct {
	BootNodes      []string      `json:"bootnodes"` // enode URLs of the P2P bootstrap nodes
	ChainID        *big.Int      `json:"chainId"`   // chainId identifies the current chain and is used for replay protection
	ChainName      string        `json:"chainName"` // chain name
	ChainURL       string        `json:"chainUrl"`  // chain url
	AccountNameCfg *NameConfig   `json:"accountParams"`
	AssetNameCfg   *NameConfig   `json:"assetParams"`
	ChargeCfg      *ChargeConfig `json:"chargeParams"`
	ForkedCfg      *FrokedConfig `json:"upgradeParams"`
	DposCfg        *DposConfig   `json:"dposParams"`
	SysName        string        `json:"systemName"`  // system name
	AccountName    string        `json:"accountName"` // system name
	DposName       string        `json:"dposName"`    // system name
	SysToken       string        `json:"systemToken"` // system token

	SysTokenID       uint64 `json:"-"`
	SysTokenDecimals uint64 `json:"-"`
}

type ChargeConfig struct {
	AssetRatio    uint64 `json:"assetRatio"`
	ContractRatio uint64 `json:"contractRatio"`
}

type NameConfig struct {
	Level     uint64 `json:"level"`
	Length    uint64 `json:"length"`
	SubLength uint64 `json:"subLength"`
}

type FrokedConfig struct {
	ForkBlockNum   uint64 `json:"blockCnt"`
	Forkpercentage uint64 `json:"upgradeRatio"`
}

type DposConfig struct {
	MaxURLLen            uint64   `json:"maxURLLen"`            // url length
	UnitStake            *big.Int `json:"unitStake"`            // state unit
	CadidateMinQuantity  *big.Int `json:"cadidateMinQuantity"`  // min quantity
	VoterMinQuantity     *big.Int `json:"voterMinQuantity"`     // min quantity
	ActivatedMinQuantity *big.Int `json:"activatedMinQuantity"` // min active quantity
	BlockInterval        uint64   `json:"blockInterval"`
	BlockFrequency       uint64   `json:"blockFrequency"`
	CadidateScheduleSize uint64   `json:"cadidateScheduleSize"`
	DelayEcho            uint64   `json:"delayEcho"`
	ExtraBlockReward     *big.Int `json:"extraBlockReward"`
	BlockReward          *big.Int `json:"blockReward"`
}

var DefaultChainconfig = &ChainConfig{
	BootNodes: []string{},
	ChainID:   big.NewInt(1),
	ChainName: "fractal",
	ChainURL:  "https://fractalproject.com",
	AccountNameCfg: &NameConfig{
		Level:     1,
		Length:    16,
		SubLength: 8,
	},
	AssetNameCfg: &NameConfig{
		Level:     1,
		Length:    16,
		SubLength: 8,
	},
	ChargeCfg: &ChargeConfig{
		AssetRatio:    80,
		ContractRatio: 80,
	},
	ForkedCfg: &FrokedConfig{
		ForkBlockNum:   10000,
		Forkpercentage: 80,
	},
	DposCfg: &DposConfig{
		MaxURLLen:            512,
		UnitStake:            big.NewInt(1000),
		CadidateMinQuantity:  big.NewInt(10),
		VoterMinQuantity:     big.NewInt(1),
		ActivatedMinQuantity: big.NewInt(100),
		BlockInterval:        3000,
		BlockFrequency:       6,
		CadidateScheduleSize: 3,
		DelayEcho:            2,
		ExtraBlockReward:     big.NewInt(1),
		BlockReward:          big.NewInt(5),
	},
	SysName:     "fractal.admin",
	AccountName: "fractal.account",
	DposName:    "fractal.dpos",
	SysToken:    "ftoken",
}

func (cfg *ChainConfig) Copy() *ChainConfig {
	bts, _ := json.Marshal(cfg)
	c := &ChainConfig{}
	json.Unmarshal(bts, c)
	return c
}

const (
	// NextForkID is the id of next fork
	NextForkID uint64 = 0
)
