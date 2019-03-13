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
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/p2p/enode"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
)

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Name   common.Name   `json:"name,omitempty"`
	PubKey common.PubKey `json:"pubKey,omitempty"`
}

// Genesis specifies the header fields, state of a genesis block.
type Genesis struct {
	Config        *params.ChainConfig  `json:"config"`
	Dpos          *dpos.Config         `json:"dpos"`
	Timestamp     uint64               `json:"timestamp"`
	ExtraData     []byte               `json:"extraData"`
	GasLimit      uint64               `json:"gasLimit" `
	Difficulty    *big.Int             `json:"difficulty" `
	Coinbase      common.Name          `json:"coinbase"`
	AllocAccounts []*GenesisAccount    `json:"allocAccounts"`
	AllocAssets   []*asset.AssetObject `json:"allocAssets"`
}

// SetupGenesisBlock The returned chain configuration is never nil.
func SetupGenesisBlock(db fdb.Database, genesis *Genesis) (*params.ChainConfig, *dpos.Config, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.DefaultChainconfig, dpos.DefaultConfig, common.Hash{}, errGenesisNoConfig
	}
	if genesis != nil && genesis.Dpos == nil {
		return params.DefaultChainconfig, dpos.DefaultConfig, common.Hash{}, errGenesisNoDpos
	}

	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			genesis = DefaultGenesis()
		}
		block, err := genesis.Commit(db)
		log.Info("Writing genesis block", "hash", block.Hash().Hex())
		return genesis.Config, genesis.Dpos, block.Hash(), err
	}

	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := genesis.ToBlock(nil).Hash()
		if hash != stored {
			return genesis.Config, genesis.Dpos, hash, &GenesisMismatchError{stored, hash}
		}
	} else {
		genesis = new(Genesis)
		head := rawdb.ReadHeader(db, stored, 0)
		genesis.UnmarshalJSON(head.Extra)
	}

	// Get the existing dpos configuration.
	newdpos := genesis.dposOrDefault(stored)

	// Get the existing chain configuration.
	newcfg := genesis.configOrDefault(stored)

	//init account and asset manager
	// if !am.SetAcctMangerName(newcfg.SysName) {
	// 	return newcfg, newdpos, stored, fmt.Errorf("genesis set account manager fail")
	// }
	if !am.SetSysName(newcfg.SysName) {
		return newcfg, newdpos, stored, fmt.Errorf("genesis set account sys name fail")
	}
	// if !asset.SetAssetMangerName(newcfg.SysName) {
	// 	return newcfg, newdpos, stored, fmt.Errorf("genesis set asset manager fail")
	// }

	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {
		return newcfg, newdpos, stored, fmt.Errorf("missing block number for head header hash")
	}
	err := newdpos.Write(db, append([]byte("ft-dpos-"), stored.Bytes()...))
	rawdb.WriteChainConfig(db, stored, newcfg)
	return newcfg, newdpos, stored, err
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db fdb.Database) *types.Block {
	if db == nil {
		db = fdb.NewMemDatabase()
	}
	number := big.NewInt(0)
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if err != nil {
		panic(fmt.Sprintf("genesis statedb new err: %v", err))
	}
	accountManager, err := am.NewAccountManager(statedb)
	if err != nil {
		panic(fmt.Sprintf("genesis accountManager new err: %v", err))
	}
	//p2p
	for _, node := range g.Config.BootNodes {
		if _, err := enode.ParseV4(node); err != nil {
			panic(fmt.Sprintf("genesis bootnodes err: %v in %v", err, node))
		}
	}

	// dpos
	if !common.IsValidName(g.Dpos.SystemName) {
		panic(fmt.Sprintf("genesis invalid dpos account name %v", g.Dpos.SystemName))
	}
	g.AllocAccounts = append(g.AllocAccounts, &GenesisAccount{
		Name:   common.StrToName(g.Dpos.AccountName),
		PubKey: common.PubKey{},
	})
	if err := dpos.Genesis(g.Dpos, statedb, number.Uint64()); err != nil {
		panic(fmt.Sprintf("genesis dpos err %v", g.Dpos.SystemName))
	}

	for _, account := range g.AllocAccounts {
		if err := accountManager.CreateAccount(account.Name, common.Name(""), 0, account.PubKey); err != nil {
			panic(fmt.Sprintf("genesis create account err %v", err))
		}
	}

	for _, asset := range g.AllocAssets {
		if err := accountManager.IssueAsset(asset); err != nil {
			panic(fmt.Sprintf("genesis issue asset err %v", err))
		}
	}

	root := statedb.IntermediateRoot()
	gjson, _ := g.MarshalJSON()
	head := &types.Header{
		Number:     number,
		Time:       new(big.Int).SetUint64(g.Timestamp),
		ParentHash: common.Hash{},
		Extra:      gjson,
		GasLimit:   g.GasLimit,
		GasUsed:    0,
		Difficulty: g.Difficulty,
		Coinbase:   g.Coinbase,
		Root:       root,
	}

	block := types.NewBlock(head, nil, nil)
	batch := db.NewBatch()
	roothash, err := statedb.Commit(batch, block.Hash(), block.NumberU64())
	if err != nil {
		panic(fmt.Sprintf("genesis statedb commit err: %v", err))
	}
	statedb.Database().TrieDB().Commit(roothash, false)
	if err := batch.Write(); err != nil {
		panic(fmt.Sprintf("genesis batch write err: %v", err))
	}
	return block
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db fdb.Database) (*types.Block, error) {
	block := g.ToBlock(db)
	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty)
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())

	config := g.Config
	if config == nil {
		config = params.DefaultChainconfig
	}
	dposConfig := g.Dpos
	if dposConfig == nil {
		dposConfig = dpos.DefaultConfig
	}

	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}

func (g *Genesis) dposOrDefault(ghash common.Hash) *dpos.Config {
	if g != nil {
		return g.Dpos
	}
	return dpos.DefaultConfig
}

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	if g != nil {
		return g.Config
	}
	return params.DefaultChainconfig
}

// DefaultGenesis returns the ft net genesis block.
func DefaultGenesis() *Genesis {
	gtime, _ := time.Parse("2006-01-02 15:04:05.999999999", "2019-01-16 00:00:00")
	return &Genesis{
		Config:        params.DefaultChainconfig,
		Dpos:          dpos.DefaultConfig,
		Timestamp:     uint64(gtime.UnixNano()),
		ExtraData:     hexutil.MustDecode(hexutil.Encode([]byte("ft Genesis Block"))),
		GasLimit:      params.GenesisGasLimit,
		Difficulty:    params.GenesisDifficulty,
		Coinbase:      params.DefaultChainconfig.SysName,
		AllocAccounts: DefaultGenesisAccounts(),
		AllocAssets:   DefaultGenesisAssets(),
	}
}

// DefaultGenesisAccounts returns the ft net genesis accounts.
func DefaultGenesisAccounts() []*GenesisAccount {
	pubKey := common.HexToPubKey(params.DefaultPubkeyHex)
	return []*GenesisAccount{
		&GenesisAccount{
			Name:   params.DefaultChainconfig.SysName,
			PubKey: pubKey,
		},
	}
}

// DefaultGenesisAssets returns the ft net genesis assets.
func DefaultGenesisAssets() []*asset.AssetObject {
	supply := new(big.Int)
	supply.SetString("100000000000000000000000000000", 10)
	return []*asset.AssetObject{
		&asset.AssetObject{
			AssetName:  params.DefaultChainconfig.SysToken,
			Symbol:     "ft",
			Amount:     supply,
			Decimals:   18,
			Owner:      params.DefaultChainconfig.SysName,
			Founder:    params.DefaultChainconfig.SysName,
			UpperLimit: supply,
		},
	}
}
