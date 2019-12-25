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
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/p2p/enode"
	"github.com/fractalplatform/fractal/params"
	pm "github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
)

// Genesis specifies the header fields, state of a genesis block.
type Genesis struct {
	Config     *params.ChainConfig `json:"config,omitempty"`
	Timestamp  uint64              `json:"timestamp,omitempty"`
	GasLimit   uint64              `json:"gasLimit,omitempty" `
	Difficulty *big.Int            `json:"difficulty,omitempty" `
	Remark     string              `json:"remark,omitempty"`
	ForkID     uint64              `json:"forkID,omitempty"`
	PluginDoc  json.RawMessage     `json:"pluginDoc,omitempty"`
}

// SetupGenesisBlock The returned chain configuration is never nil.
func SetupGenesisBlock(db fdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.DefaultChainconfig, common.Hash{}, errGenesisNoConfig
	}

	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			genesis = DefaultGenesis()
		}
		block, err := genesis.Commit(db)
		if err != nil {
			return nil, common.Hash{}, err
		}

		log.Info("Writing genesis block", "hash", block.Hash().Hex())
		return genesis.Config, block.Hash(), err
	}

	// Check whether the genesis block is already written.
	if genesis != nil {
		blk, _, err := genesis.ToBlock(nil)
		if err != nil {
			return nil, common.Hash{}, err

		}
		hash := blk.Hash()
		if hash != stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
	}

	number := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if number == nil {
		return nil, common.Hash{}, errors.New("missing block number for head header hash")
	}

	storedCfg := rawdb.ReadChainConfig(db, stored)
	if storedCfg == nil {
		return nil, common.Hash{}, errors.New("Found genesis block without chain config")
	}

	return storedCfg, stored, nil
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db fdb.Database) (*types.Block, []*types.Receipt, error) {
	if db == nil {
		db = rawdb.NewMemoryDatabase()
	}

	number := uint64(0)
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if err != nil {
		return nil, nil, fmt.Errorf("genesis statedb new err: %v", err)
	}

	//p2p
	for _, node := range g.Config.BootNodes {
		if len(node) == 0 {
			continue
		}
		if _, err := enode.ParseV4(node); err != nil {
			log.Warn("genesis bootnodes parse failed", "err", err, "node", node)
		}
	}

	//check timestamp
	str := strconv.FormatUint(g.Timestamp, 10)
	if len(str) < 9 || len(str) > 10 {
		return nil, nil, fmt.Errorf("genesis timestamp too late or too early")
	}

	timestamp := g.Timestamp
	g.Config.ReferenceTime = timestamp

	// create account and asset
	mananger := pm.NewPM(statedb)
	mananger.Init(g.Timestamp, nil)
	actTxs, err := mananger.InitChain(g.PluginDoc, g.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("plugin init chain err %v", err)
	}
	sysAsset, err := mananger.GetAssetInfoByName(g.Config.SysToken)
	if err != nil {
		return nil, nil, fmt.Errorf("genesis get system asset err %v", err)
	}
	g.Config.SysTokenID = sysAsset.AssetID
	g.Config.SysTokenDecimals = sysAsset.Decimals

	root := statedb.IntermediateRoot()
	aBytes, err := json.Marshal(g)
	if err != nil {
		return nil, nil, fmt.Errorf("genesis json marshal json err %v", err)
	}

	head := &types.Header{
		Number:     number,
		Time:       timestamp,
		ParentHash: common.Hash{},
		GasLimit:   g.GasLimit,
		GasUsed:    0,
		Coinbase:   g.Config.ChainName,
		Root:       root,
		Extra:      aBytes,
	}

	receipts := []*types.Receipt{}
	for k, tx := range actTxs {
		receipt := types.NewReceipt(root[:], 0, 0)
		receipt.TxHash = tx.Hash()
		receipt.Status = 1
		receipt.Index = uint64(k)
		receipts = append(receipts, receipt)
	}

	block := types.NewBlock(head, actTxs, receipts)

	batch := db.NewBatch()
	// write snapshot to db
	snapshotInfo := types.SnapshotInfo{
		Root: root,
	}
	key := types.SnapshotBlock{
		Number:    block.NumberU64(),
		BlockHash: block.ParentHash(),
	}
	rawdb.WriteSnapshot(db, key, snapshotInfo)

	roothash, err := statedb.Commit(batch, block.Hash(), block.NumberU64())
	if err != nil {
		return nil, nil, fmt.Errorf("genesis statedb commit err: %v", err)
	}
	statedb.Database().TrieDB().Commit(roothash, false)
	if err := batch.Write(); err != nil {
		return nil, nil, fmt.Errorf("genesis batch write err: %v", err)
	}
	return block, receipts, nil
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db fdb.Database) (*types.Block, error) {
	block, receipts, err := g.ToBlock(db)
	if err != nil {
		return nil, err
	}
	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty)
	rawdb.WriteBlock(db, block)
	rawdb.WriteTxLookupEntries(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), receipts)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())
	rawdb.WriteChainConfig(db, block.Hash(), g.Config)
	rawdb.WriteIrreversibleNumber(db, uint64(0))
	return block, nil
}

// DefaultGenesis returns the ft net genesis block.
func DefaultGenesis() *Genesis {
	return &Genesis{
		Config:     params.DefaultChainconfig,
		Timestamp:  1575967052,
		GasLimit:   params.BlockGasLimit,
		Difficulty: params.GenesisDifficulty,
	}
}
