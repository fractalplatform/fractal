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
	"bytes"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/utils/fdb"
	memdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

var defaultgenesisBlockHash = common.HexToHash("0xf360838779f71259f2102dd97ce8f0e0553bbd2b3d734687634042ed7821bf44")

func TestDefaultGenesisBlock(t *testing.T) {
	block, _ := DefaultGenesis().ToBlock(nil)
	if block.Hash() != defaultgenesisBlockHash {
		t.Errorf("wrong mainnet genesis hash, got %v, want %v", block.Hash().Hex(), defaultgenesisBlockHash.Hex())
	}
}

func TestSetupGenesis(t *testing.T) {
	var (
		customghash = common.HexToHash("0xb10e60cb09835f02799ae89f4a30e7e60be6cea5d6b1e7c8511dcc42dfbf9c65")
		customg     = Genesis{
			Config:          params.DefaultChainconfig.Copy(),
			AllocAccounts:   DefaultGenesisAccounts(),
			AllocAssets:     DefaultGenesisAssets(),
			AllocCandidates: DefaultGenesisCandidates(),
		}
		oldcustomg     = customg
		oldcustomghash = common.HexToHash("d1b02e1545a6469e79b38309928c04044cf2e5a00cf9aca2576de454f6c1ca06")
	)
	customg.Config.ChainID = big.NewInt(5)
	oldcustomg.Config = customg.Config.Copy()
	oldcustomg.Config.ChainID = big.NewInt(6)

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
		},
		{
			name: "no block in DB, genesis == nil",
			fn: func(db fdb.Database) (*params.ChainConfig, *dpos.Config, common.Hash, error) {
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   defaultgenesisBlockHash,
			wantConfig: params.DefaultChainconfig,
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
		},
		{
			name: "compatible config in DB",
			fn: func(db fdb.Database) (*params.ChainConfig, *dpos.Config, common.Hash, error) {
				if _, err := oldcustomg.Commit(db); err != nil {
					return nil, nil, common.Hash{}, err
				}
				return SetupGenesisBlock(db, &customg)
			},
			wantErr: &GenesisMismatchError{
				oldcustomghash,
				customghash,
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
		},
	}

	for _, test := range tests {
		db := memdb.NewMemDatabase()

		config, _, hash, err := test.fn(db)

		// Check the return values.
		if !reflect.DeepEqual(err, test.wantErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: 1 returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
		}

		bts, _ := json.Marshal(config)
		wbts, _ := json.Marshal(test.wantConfig)
		if bytes.Compare(bts, wbts) != 0 {
			t.Errorf("%s:\n 2 returned %v\nwant     %v", test.name, config, test.wantConfig)
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
