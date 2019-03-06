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

var defaultgenesisBlockHash = common.HexToHash("0x311caf361ad997c6261ce0dfb2c3d1779f4f6cfba128bcab33c0c5f4717d9603")

func TestDefaultGenesisBlock(t *testing.T) {
	block := DefaultGenesis().ToBlock(nil)
	if block.Hash() != defaultgenesisBlockHash {
		t.Errorf("wrong mainnet genesis hash, got %v, want %v", block.Hash().Hex(), defaultgenesisBlockHash.Hex())
	}
}

func TestSetupGenesis(t *testing.T) {
	var (
		customghash = common.HexToHash("0x36214d7aec5916cddac2c073d7ac28544c664dbea2b962fb863a76ad23479179")
		customg     = Genesis{
			Config:        &params.ChainConfig{ChainID: big.NewInt(3), SysName: "systemio", SysToken: "fractalfoundation"},
			Dpos:          dpos.DefaultConfig,
			Coinbase:      "coinbase",
			AllocAccounts: DefaultGenesisAccounts(),
			AllocAssets:   DefaultGenesisAssets(),
		}
		oldcustomg     = customg
		oldcustomghash = common.HexToHash("0x066b2f98ea3cd44a6f8d2a9c5fb21f8776e4276c586425d46bfa4c687eb4de72")
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
			wantConfig: params.DefaultChainconfig,
			wantDpos:   dpos.DefaultConfig,
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
		fmt.Println("=====>", i, test.name)

		config, dpos, hash, err := test.fn(db)
		fmt.Println("=====>", i, test.name, err)

		// Check the return values.
		if !reflect.DeepEqual(err, test.wantErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
		}
		if !reflect.DeepEqual(config, test.wantConfig) {
			t.Errorf("%s:\n returned %v\nwant     %v", test.name, config, test.wantConfig)
		}

		if !reflect.DeepEqual(dpos, test.wantDpos) {
			t.Errorf("%s:\nreturned %v\nwant     %v", test.name, config, test.wantConfig)
		}

		if hash != test.wantHash {
			t.Errorf("%s: returned hash %s, want %s", test.name, hash.Hex(), test.wantHash.Hex())
		} else if err == nil {
			// Check database content.
			stored := rawdb.ReadBlock(db, test.wantHash, 0)
			if stored.Hash() != test.wantHash {
				t.Errorf("%s: block in DB has hash %s, want %s", test.name, stored.Hash(), test.wantHash)
			}
		}
	}
}
