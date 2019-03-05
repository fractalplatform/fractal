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
	"math/big"

	"github.com/fractalplatform/fractal/common"
)

const DefaultPubkeyHex = "047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"

// ChainConfig is the core config which determines the blockchain settings.
// ChainConfig is stored in the database on a per block basis.
type ChainConfig struct {
	ChainID             *big.Int    `json:"chainId"`        // chainId identifies the current chain and is used for replay protection
	BootNodes           []string    `json:"bootnodes"`      // enode URLs of the P2P bootstrap nodes
	SysName             common.Name `json:"sysName"`        // system name
	AssetManager        common.Name `json:"ftassetmanager"` // global asset manager name
	AccountManager      common.Name `json:"ftacctmanager"`  // global account manager name
	SysToken            string      `json:"sysToken"`       // system token
	AssetChargeRatio    uint64      `json:"assetChargeRatio"`
	ContractChargeRatio uint64      `json:"contractChargeRatio"`
	SysTokenID          uint64      `json:"-"`
	SysTokenDecimals    uint64      `json:"-"`
	UpperLimit          *big.Int    `json:"upperlimit"` //
}

var DefaultChainconfig = &ChainConfig{
	ChainID:             big.NewInt(1),
	SysName:             "ftsystemio",
	AssetManager:        "ftassetmanager",
	AccountManager:      "ftacctmanager",
	SysToken:            "ftoken",
	AssetChargeRatio:    80,
	ContractChargeRatio: 80,
}
