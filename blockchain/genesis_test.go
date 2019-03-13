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

package blockchain

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/utils/fdb"
)

var defaultgenesisBlockHash = common.HexToHash("0xa139f1d16692026cd921e437e45506f2dec998f8ea57a5075e878afd4b428cc3")

func TestDefaultGenesisBlock(t *testing.T) {
	block := DefaultGenesis().ToBlock(nil)
	if block.Hash() != defaultgenesisBlockHash {
		t.Errorf("wrong mainnet genesis hash, got %v, want %v", block.Hash().Hex(), defaultgenesisBlockHash.Hex())
	}
}

func TestSetupGenesis(t *testing.T) {
	var (
		customghash = common.HexToHash("0xf0ef187e3f23ac20aba4df4e4a6d6327f2eb56a0724c1a690be32d0af5ecb501")
		customg     = Genesis{
			Config:        &params.ChainConfig{ChainID: big.NewInt(3), SysName: "systemio", SysToken: "fractalfoundation"},
			Dpos:          dpos.DefaultConfig,
			Coinbase:      "coinbase",
			AllocAccounts: DefaultGenesisAccounts(),
			AllocAssets:   DefaultGenesisAssets(),
		}
		oldcustomg     = customg
		oldcustomghash = common.HexToHash("0x461543306818df6d262089961cb45becf06a551dd36f61ab4cee78af8fb47f64")

		dposConfig = &dpos.Config{
			MaxURLLen:            512,
			UnitStake:            big.NewInt(1000),
			ProducerMinQuantity:  big.NewInt(10),
			VoterMinQuantity:     big.NewInt(1),
			ActivatedMinQuantity: big.NewInt(100),
			BlockInterval:        3000,
			BlockFrequency:       6,
			ProducerScheduleSize: 3,
			DelayEcho:            2,
			AccountName:          "ftsystemdpos",
			SystemName:           "ftsystemio",
			SystemURL:            "www.fractalproject.com",
			ExtraBlockReward:     big.NewInt(1),
			BlockReward:          big.NewInt(5),
			Decimals:             18,
		}

		chainConfig = &params.ChainConfig{
			ChainID:             big.NewInt(1),
			SysName:             "ftsystemio",
			SysToken:            "ftoken",
			AssetChargeRatio:    80,
			ContractChargeRatio: 80,
		}
	)
	oldcustomg.Config = &params.ChainConfig{ChainID: big.NewInt(2), SysName: "ftsystem", SysToken: "ftoken"}

	tests := []struct {
		name       string
		fn         func(fdb.Database) (*params.ChainConfig, *dpos.Config, common.Hash, error)
		wantConfig *params.ChainConfig
		wantDpos   *dpos.Config
		wantHash   common.Hash
		wantErr    error
	}{
		{
			name: "genesis without ChainConfig",
			fn: func(db fdb.Database) (*params.ChainConfig, *dpos.Config, common.Hash, error) {
				return SetupGenesisBlock(db, new(Genesis))
			},
			wantErr:    errGenesisNoConfig,
			wantConfig: params.DefaultChainconfig,
			wantDpos:   dpos.DefaultConfig,
		},
		{
			name: "no block in DB, genesis == nil",
			fn: func(db fdb.Database) (*params.ChainConfig, *dpos.Config, common.Hash, error) {
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   defaultgenesisBlockHash,
			wantConfig: params.DefaultChainconfig,
			wantDpos:   dpos.DefaultConfig,
		},
		{
			name: "mainnet block in DB, genesis == nil",
			fn: func(db fdb.Database) (*params.ChainConfig, *dpos.Config, common.Hash, error) {
				if _, err := DefaultGenesis().Commit(db); err != nil {
					return nil, nil, common.Hash{}, err
				}
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   defaultgenesisBlockHash,
			wantConfig: chainConfig,
			wantDpos:   dposConfig,
		},
		{
			name: "compatible config in DB",
			fn: func(db fdb.Database) (*params.ChainConfig, *dpos.Config, common.Hash, error) {
				if _, err := oldcustomg.Commit(db); err != nil {
					return nil, nil, common.Hash{}, err
				}
				fmt.Println("=====>SetupGenesisBlock")
				return SetupGenesisBlock(db, &customg)
			},
			wantErr: &GenesisMismatchError{
				oldcustomghash,
				customghash,
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
			wantDpos:   customg.Dpos,
		},
	}

	for i, test := range tests {
		db := fdb.NewMemDatabase()
		fmt.Println("1=====>", i, test.name)

		config, dpos, hash, err := test.fn(db)
		fmt.Println("2=====>", i, test.name, err)

		// Check the return values.
		if !reflect.DeepEqual(err, test.wantErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: 1 returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
		}

		if !reflect.DeepEqual(config, test.wantConfig) {
			t.Errorf("%s:\n 2 returned %v\nwant     %v", test.name, config, test.wantConfig)
		}

		if !reflect.DeepEqual(dpos, test.wantDpos) {
			t.Errorf("%s:\n 3returned %v\nwant     %v", test.name, config, test.wantConfig)
		}

		if hash != test.wantHash {
			t.Errorf("%s: 4 returned hash %s, want %s", test.name, hash.Hex(), test.wantHash.Hex())
		} else if err == nil {
			// Check database content.
			stored := rawdb.ReadBlock(db, test.wantHash, 0)
			if stored.Hash() != test.wantHash {
				t.Errorf("%s: 5 block in DB has hash %s, want %s", test.name, stored.Hash(), test.wantHash)
			}
		}
	}
}
