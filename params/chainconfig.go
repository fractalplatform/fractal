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
	BootNodes        []string      `json:"bootnodes"` // enode URLs of the P2P bootstrap nodes
	ChainID          *big.Int      `json:"chainId"`   // chainId identifies the current chain and is used for replay protection
	ChainName        string        `json:"chainName"` // chain name
	ChainURL         string        `json:"chainUrl"`  // chain url
	AccountNameCfg   *NameConfig   `json:"accountParams"`
	AssetNameCfg     *NameConfig   `json:"assetParams"`
	ChargeCfg        *ChargeConfig `json:"chargeParams"`
	ForkedCfg        *FrokedConfig `json:"upgradeParams"`
	DposCfg          *DposConfig   `json:"dposParams"`
	SysName          string        `json:"systemName"`  // system name
	AccountName      string        `json:"accountName"` // account name
	AssetName        string        `json:"assetName"`   // asset name
	DposName         string        `json:"dposName"`    // system name
	SnapshotInterval uint64        `json:"snapshotInterval"`
	FeeName          string        `json:"feeName"`     //fee name
	SysToken         string        `json:"systemToken"` // system token
	SysTokenID       uint64        `json:"sysTokenID"`
	SysTokenDecimals uint64        `json:"sysTokenDecimal"`
	ReferenceTime    uint64        `json:"referenceTime"`
}

type ChargeConfig struct {
	AssetRatio    uint64 `json:"assetRatio"`
	ContractRatio uint64 `json:"contractRatio"`
}

type NameConfig struct {
	Level         uint64 `json:"level"`
	AllLength     uint64 `json:"alllength"`
	MainMinLength uint64 `json:"mainminlength"`
	MainMaxLength uint64 `json:"mainmaxlength"`
	SubMinLength  uint64 `json:"subminLength"`
	SubMaxLength  uint64 `json:"submaxLength"`
}

type FrokedConfig struct {
	ForkBlockNum   uint64 `json:"blockCnt"`
	Forkpercentage uint64 `json:"upgradeRatio"`
}

type DposConfig struct {
	MaxURLLen                     uint64   `json:"maxURLLen"` // url length
	UnitStake                     *big.Int `json:"unitStake"` // state unit
	CandidateAvailableMinQuantity *big.Int `json:"candidateAvailableMinQuantity"`
	CandidateMinQuantity          *big.Int `json:"candidateMinQuantity"` // min quantity
	VoterMinQuantity              *big.Int `json:"voterMinQuantity"`     // min quantity
	ActivatedMinCandidate         uint64   `json:"activatedMinCandidate"`
	ActivatedMinQuantity          *big.Int `json:"activatedMinQuantity"` // min active quantity
	BlockInterval                 uint64   `json:"blockInterval"`
	BlockFrequency                uint64   `json:"blockFrequency"`
	CandidateScheduleSize         uint64   `json:"candidateScheduleSize"`
	BackupScheduleSize            uint64   `json:"backupScheduleSize"`
	EpochInterval                 uint64   `json:"epochInterval"`
	FreezeEpochSize               uint64   `json:"freezeEpochSize"`
	ExtraBlockReward              *big.Int `json:"extraBlockReward"`
	BlockReward                   *big.Int `json:"blockReward"`
}

var DefaultChainconfig = &ChainConfig{
	BootNodes: []string{},
	ChainID:   big.NewInt(1),
	ChainName: "fractal",
	ChainURL:  "https://fractalproject.com",
	AccountNameCfg: &NameConfig{
		Level:         1,
		AllLength:     31,
		MainMinLength: 7,
		MainMaxLength: 16,
		SubMinLength:  2,
		SubMaxLength:  8,
	},
	AssetNameCfg: &NameConfig{
		Level:         2,
		AllLength:     31,
		MainMinLength: 2,
		MainMaxLength: 16,
		SubMinLength:  1,
		SubMaxLength:  8,
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
		MaxURLLen:                     512,
		UnitStake:                     big.NewInt(1000),
		CandidateAvailableMinQuantity: big.NewInt(10),
		CandidateMinQuantity:          big.NewInt(10),
		VoterMinQuantity:              big.NewInt(2),
		ActivatedMinCandidate:         3,
		ActivatedMinQuantity:          big.NewInt(100),
		BlockInterval:                 3000,
		BlockFrequency:                6,
		CandidateScheduleSize:         3,
		BackupScheduleSize:            0,
		EpochInterval:                 1080000,
		FreezeEpochSize:               3,
		ExtraBlockReward:              big.NewInt(1),
		BlockReward:                   big.NewInt(5),
	},
	SnapshotInterval: 180000,
	SysName:          "fractal.founder",
	AccountName:      "fractal.account",
	AssetName:        "fractal.asset",
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
	//ForkID0 init
	ForkID0 = uint64(0)
	//ForkID1 account first name > 12, asset name contain account name
	ForkID1 = uint64(1)
	//ForkID2 dpos
	ForkID2 = uint64(2)

	// NextForkID is the id of next fork
	NextForkID uint64 = 2
)
