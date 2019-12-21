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

package genesis

import (
	"bytes"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/utils/fdb"
)

var defaultGenesisBlockHash = common.HexToHash("0x5bb30c6dd70f4297f603e7f06ff929de1afd1ef18db20f8a4acee708fa180421")

func TestDefaultGenesisBlock(t *testing.T) {
	block, _, err := DefaultGenesis().ToBlock(nil)
	if err != nil {
		t.Fatal(err)
	}
	if block.Hash() != defaultGenesisBlockHash {
		t.Errorf("wrong mainnet genesis hash, got %v, want %v", block.Hash().Hex(), defaultGenesisBlockHash.Hex())
	}
}

func TestSetupGenesis(t *testing.T) {
	var (
		customGHash = common.HexToHash("a38b5027bb6d7637ee9007e70e0d3e72156b5ee5d4db8269d6f33d81d2a0ad3e")

		customG = Genesis{
			Config: params.DefaultChainconfig.Copy(),
		}
		oldCustomG = customG

		oldCustomGHash = common.HexToHash("9b994f2bd6aa177c97c715278c0b92417963d5724e4b314a2096f3b69f31815d")
	)
	customG.Config.ChainID = big.NewInt(5)
	oldCustomG.Config = customG.Config.Copy()
	oldCustomG.Config.ChainID = big.NewInt(6)

	tests := []struct {
		name       string
		fn         func(fdb.Database) (*params.ChainConfig, common.Hash, error)
		wantConfig *params.ChainConfig
		wantHash   common.Hash
		wantErr    error
	}{
		{
			name: "genesis without ChainConfig",
			fn: func(db fdb.Database) (*params.ChainConfig, common.Hash, error) {
				return SetupGenesisBlock(db, new(Genesis))
			},
			wantErr:    errGenesisNoConfig,
			wantConfig: params.DefaultChainconfig,
		},
		{
			name: "no block in DB, genesis == nil",
			fn: func(db fdb.Database) (*params.ChainConfig, common.Hash, error) {
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   defaultGenesisBlockHash,
			wantConfig: params.DefaultChainconfig,
		},
		{
			name: "mainnet block in DB, genesis == nil",
			fn: func(db fdb.Database) (*params.ChainConfig, common.Hash, error) {
				if _, err := DefaultGenesis().Commit(db); err != nil {
					return nil, common.Hash{}, err
				}
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   defaultGenesisBlockHash,
			wantConfig: params.DefaultChainconfig,
		},
		{
			name: "compatible config in DB",
			fn: func(db fdb.Database) (*params.ChainConfig, common.Hash, error) {
				if _, err := oldCustomG.Commit(db); err != nil {
					return nil, common.Hash{}, err
				}
				return SetupGenesisBlock(db, &customG)
			},
			wantErr: &GenesisMismatchError{
				oldCustomGHash,
				customGHash,
			},
			wantHash:   customGHash,
			wantConfig: customG.Config,
		},
	}

	for _, test := range tests {
		db := rawdb.NewMemoryDatabase()
		config, hash, err := test.fn(db)

		// Check the return values.
		if !reflect.DeepEqual(err, test.wantErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: 1 returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
		}

		bts, _ := json.Marshal(config)
		wbts, _ := json.Marshal(test.wantConfig)
		if !bytes.Equal(bts, wbts) {
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
