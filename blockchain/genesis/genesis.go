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
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/p2p/enode"
	"github.com/fractalplatform/fractal/params"
	pm "github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/snapshot"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// Genesis specifies the header fields, state of a genesis block.
type Genesis struct {
	Config        *params.ChainConfig       `json:"config,omitempty"`
	Timestamp     uint64                    `json:"timestamp,omitempty"`
	GasLimit      uint64                    `json:"gasLimit,omitempty" `
	Difficulty    *big.Int                  `json:"difficulty,omitempty" `
	AllocAccounts []*pm.CreateAccountAction `json:"allocAccounts,omitempty"`
	AllocAssets   []*pm.IssueAssetAction    `json:"allocAssets,omitempty"`
	Remark        string                    `json:"remark,omitempty"`
	ForkID        uint64                    `json:"forkID,omitempty"`
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
	g.Timestamp = uint64(time.Now().Unix())
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

	timestamp := g.Timestamp
	g.Config.ReferenceTime = timestamp

	// create account and asset
	mananger := pm.NewPM(statedb)
	mananger.Init(g.Timestamp, g.Config.DposName, nil)
	actions, err := g.CreateAccount()
	if err != nil {
		return nil, nil, err
	}

	astActions, err := g.CreateAsset()
	if err != nil {
		return nil, nil, err
	}

	minerActions, err := g.RegisterMiner()
	if err != nil {
		return nil, nil, err
	}

	actions = append(actions, astActions...)
	actions = append(actions, minerActions...)

	for index, action := range actions {
		_, err := mananger.ExecTx(action)
		if err != nil {
			return nil, nil, fmt.Errorf("genesis index: %v,create %v,err %v", index, action.Type(), err)
		}
	}

	g.Config.SysTokenID, err = mananger.GetAssetID(g.Config.SysToken)
	if err != nil {
		return nil, nil, fmt.Errorf("genesis system asset err %v", err)
	}

	// snapshot
	currentTime := timestamp
	currentTimeFormat := (currentTime / g.Config.SnapshotInterval) * g.Config.SnapshotInterval
	snapshotManager := snapshot.NewSnapshotManager(statedb)
	err = snapshotManager.SetSnapshot(currentTimeFormat, snapshot.BlockInfo{Number: number, BlockHash: common.Hash{}, Timestamp: 0})
	if err != nil {
		return nil, nil, fmt.Errorf("genesis snapshot err %v", err)
	}

	root := statedb.IntermediateRoot()
	aBytes, err := json.Marshal(g)
	if err != nil {
		return nil, nil, fmt.Errorf("genesis json marshal json err %v", err)
	}

	head := &types.Header{
		Number:     number,
		Time:       timestamp,
		ParentHash: common.Hash{},
		Extra:      aBytes,
		GasLimit:   g.GasLimit,
		GasUsed:    0,
		Coinbase:   g.Config.ChainName,
		Root:       root,
	}

	tx := types.NewTransaction(g.Config.SysTokenID, big.NewInt(0), actions...)
	receipt := types.NewReceipt(root[:], 0, 0)
	receipt.TxHash = tx.Hash()
	for index := range actions {
		receipt.ActionResults = append(receipt.ActionResults, &types.ActionResult{
			Status:  1,
			Index:   uint64(index),
			GasUsed: 0,
		})
	}

	receipts := []*types.Receipt{receipt}
	block := types.NewBlock(head, []*types.Transaction{tx}, receipts)
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

// CreateAccount create account
func (g *Genesis) CreateAccount() ([]*types.Action, error) {
	actions := []*types.Action{}
	act := &pm.CreateAccountAction{
		Name:   g.Config.ChainName,
		Pubkey: common.PubKey{},
	}

	payload, err := rlp.EncodeToBytes(act)
	if err != nil {
		return nil, err
	}
	actions = append(actions, types.NewAction(
		pm.CreateAccount,
		"",
		g.Config.AccountName,
		0,
		0,
		0,
		big.NewInt(0),
		payload,
		[]byte(g.Remark),
	))

	for _, act := range g.AllocAccounts {
		payload, err := rlp.EncodeToBytes(act)
		if err != nil {
			return nil, err
		}
		actions = append(actions, types.NewAction(
			pm.CreateAccount,
			g.Config.ChainName,
			g.Config.AccountName,
			0,
			0,
			0,
			big.NewInt(0),
			payload,
			nil,
		))
	}

	return actions, nil
}

// CreateAsset create asset
func (g *Genesis) CreateAsset() ([]*types.Action, error) {
	actions := []*types.Action{}

	for _, ast := range g.AllocAssets {
		payload, err := rlp.EncodeToBytes(ast)
		if err != nil {
			return nil, err
		}
		actions = append(actions, types.NewAction(
			pm.IssueAsset,
			g.Config.ChainName,
			g.Config.AssetName,
			0,
			0,
			0,
			big.NewInt(0),
			payload,
			nil,
		))
	}

	return actions, nil

}

// RegisterMiner register Miner
func (g *Genesis) RegisterMiner() ([]*types.Action, error) {
	return []*types.Action{types.NewAction(
		pm.RegisterMiner,
		g.Config.SysName,
		g.Config.DposName,
		1,             // nonce
		0,             // assetID
		0,             // gasLimit
		big.NewInt(1), // amount
		nil,
		nil,
	)}, nil
}

// DefaultGenesis returns the ft net genesis block.
func DefaultGenesis() *Genesis {
	return &Genesis{
		Config:        params.DefaultChainconfig,
		Timestamp:     1555776000000, // 2019-04-21 00:00:00
		GasLimit:      params.BlockGasLimit,
		Difficulty:    params.GenesisDifficulty,
		AllocAccounts: DefaultGenesisAccounts(),
		// 	AllocCandidates: DefaultGenesisCandidates(),
		AllocAssets: DefaultGenesisAssets(),
	}
}

// DefaultGenesisAccounts returns the ft net genesis accounts.
func DefaultGenesisAccounts() []*pm.CreateAccountAction {
	return []*pm.CreateAccountAction{
		&pm.CreateAccountAction{
			Name:   params.DefaultChainconfig.SysName,
			Desc:   "system account",
			Pubkey: common.HexToPubKey("047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"),
		},
		&pm.CreateAccountAction{
			Name:   params.DefaultChainconfig.AccountName,
			Desc:   "account manager account",
			Pubkey: common.HexToPubKey(""),
		},
		&pm.CreateAccountAction{
			Name:   params.DefaultChainconfig.AssetName,
			Desc:   "asset manager account",
			Pubkey: common.HexToPubKey(""),
		},
		&pm.CreateAccountAction{
			Name:   params.DefaultChainconfig.DposName,
			Desc:   "consensus account",
			Pubkey: common.HexToPubKey(""),
		},
		&pm.CreateAccountAction{
			Name:   params.DefaultChainconfig.FeeName,
			Desc:   "fee manager account",
			Pubkey: common.HexToPubKey(""),
		},
	}
}

// DefaultGenesisCandidates returns the ft net genesis candidates.
// func DefaultGenesisCandidates() []*Candidate {
// 	return []*Candidate{}
// }

// DefaultGenesisAssets returns the ft net genesis assets.
func DefaultGenesisAssets() []*pm.IssueAssetAction {
	supply := new(big.Int)
	supply.SetString("10000000000000000000000000000", 10)
	return []*pm.IssueAssetAction{
		&pm.IssueAssetAction{
			AssetName:   params.DefaultChainconfig.SysToken,
			Symbol:      "ft",
			Amount:      supply,
			Decimals:    18,
			Owner:       params.DefaultChainconfig.SysName,
			Founder:     params.DefaultChainconfig.SysName,
			UpperLimit:  supply,
			Contract:    "",
			Description: "",
		},
	}
}
