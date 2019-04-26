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
	BootNodes        []string      `json:"bootnodes,omitempty"` // enode URLs of the P2P bootstrap nodes
	ChainID          *big.Int      `json:"chainId,omitempty"`   // chainId identifies the current chain and is used for replay protection
	ChainName        string        `json:"chainName,omitempty"` // chain name
	ChainURL         string        `json:"chainUrl,omitempty"`  // chain url
	AccountNameCfg   *NameConfig   `json:"accountParams,omitempty"`
	AssetNameCfg     *NameConfig   `json:"assetParams,omitempty"`
	ChargeCfg        *ChargeConfig `json:"chargeParams,omitempty"`
	ForkedCfg        *FrokedConfig `json:"upgradeParams,omitempty"`
	DposCfg          *DposConfig   `json:"dposParams,omitempty"`
	SysName          string        `json:"systemName,omitempty"`  // system name
	AccountName      string        `json:"accountName,omitempty"` // system name
	DposName         string        `json:"dposName,omitempty"`    // system name
	SnapshotInterval uint64        `json:"snapshotInterval,omitempty"`
	FeeName          string        `json:"feeName,omitempty"`     //fee name
	SysToken         string        `json:"systemToken,omitempty"` // system token
	SysTokenID       uint64        `json:"sysTokenID,omitempty"`
	SysTokenDecimals uint64        `json:"sysTokenDecimal,omitempty"`
	ReferenceTime    uint64        `json:"referenceTime,omitempty"`
}

type ChargeConfig struct {
	AssetRatio    uint64 `json:"assetRatio,omitempty"`
	ContractRatio uint64 `json:"contractRatio,omitempty"`
}

type NameConfig struct {
	Level     uint64 `json:"level,omitempty"`
	Length    uint64 `json:"length,omitempty"`
	SubLength uint64 `json:"subLength,omitempty"`
}

type FrokedConfig struct {
	ForkBlockNum   uint64 `json:"blockCnt,omitempty"`
	Forkpercentage uint64 `json:"upgradeRatio,omitempty"`
}

type DposConfig struct {
	MaxURLLen             uint64   `json:"maxURLLen,omitempty"`            // url length
	UnitStake             *big.Int `json:"unitStake,omitempty"`            // state unit
	CandidateMinQuantity  *big.Int `json:"candidateMinQuantity,omitempty"` // min quantity
	VoterMinQuantity      *big.Int `json:"voterMinQuantity,omitempty"`     // min quantity
	ActivatedMinQuantity  *big.Int `json:"activatedMinQuantity,omitempty"` // min active quantity
	BlockInterval         uint64   `json:"blockInterval,omitempty"`
	BlockFrequency        uint64   `json:"blockFrequency,omitempty"`
	CandidateScheduleSize uint64   `json:"candidateScheduleSize,omitempty"`
	BackupScheduleSize    uint64   `json:"backupScheduleSize,omitempty"`
	EpchoInterval         uint64   `json:"epchoInterval,omitempty"`
	FreezeEpchoSize       uint64   `json:"freezeEpchoSize,omitempty"`
	ExtraBlockReward      *big.Int `json:"extraBlockReward,omitempty"`
	BlockReward           *big.Int `json:"blockReward,omitempty"`
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
		MaxURLLen:             512,
		UnitStake:             big.NewInt(1000),
		CandidateMinQuantity:  big.NewInt(10),
		VoterMinQuantity:      big.NewInt(1),
		ActivatedMinQuantity:  big.NewInt(100),
		BlockInterval:         3000,
		BlockFrequency:        6,
		CandidateScheduleSize: 3,
		BackupScheduleSize:    0,
		EpchoInterval:         540000,
		FreezeEpchoSize:       3,
		ExtraBlockReward:      big.NewInt(1),
		BlockReward:           big.NewInt(5),
	},
	SnapshotInterval: 180000,
	SysName:          "fractal.admin",
	AccountName:      "fractal.account",
	DposName:         "fractal.dpos",
	FeeName:          "fractal.fee",
	SysToken:         "ftoken",
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
