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
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/consensus/dpos"

	"github.com/ethereum/go-ethereum/log"
	am "github.com/fractalplatform/fractal/accountmanager"
	at "github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	fm "github.com/fractalplatform/fractal/feemanager"
	"github.com/fractalplatform/fractal/p2p/enode"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	memdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Name   string        `json:"name,omitempty"`
	PubKey common.PubKey `json:"pubKey,omitempty"`
}

// GenesisCandidate is an cadicate in the state of the genesis block.
type GenesisCandidate struct {
	Name  string   `json:"name,omitempty"`
	URL   string   `json:"url,omitempty"`
	Stake *big.Int `json:"stake,omitempty"`
}

// GenesisAsset is an asset in the state of the genesis block.
type GenesisAsset struct {
	Name       string   `json:"name,omitempty"`
	Symbol     string   `json:"symbol,omitempty"`
	Amount     *big.Int `json:"amount,omitempty"`
	Decimals   uint64   `json:"decimals,omitempty"`
	Founder    string   `json:"founder,omitempty"`
	Owner      string   `json:"owner,omitempty"`
	UpperLimit *big.Int `json:"upperLimit,omitempty"`
}

// Genesis specifies the header fields, state of a genesis block.
type Genesis struct {
	Config          *params.ChainConfig `json:"config,omitempty"`
	Timestamp       uint64              `json:"timestamp,omitempty"`
	GasLimit        uint64              `json:"gasLimit,omitempty" `
	Difficulty      *big.Int            `json:"difficulty,omitempty" `
	AllocAccounts   []*GenesisAccount   `json:"allocAccounts,omitempty"`
	AllocCandidates []*GenesisCandidate `json:"allocCandidates,omitempty"`
	AllocAssets     []*GenesisAsset     `json:"allocAssets,omitempty"`
}

func dposConfig(cfg *params.ChainConfig) *dpos.Config {
	return &dpos.Config{
		MaxURLLen:             cfg.DposCfg.MaxURLLen,
		UnitStake:             cfg.DposCfg.UnitStake,
		CandidateMinQuantity:  cfg.DposCfg.CandidateMinQuantity,
		VoterMinQuantity:      cfg.DposCfg.VoterMinQuantity,
		ActivatedMinQuantity:  cfg.DposCfg.ActivatedMinQuantity,
		BlockInterval:         cfg.DposCfg.BlockInterval,
		BlockFrequency:        cfg.DposCfg.BlockFrequency,
		CandidateScheduleSize: cfg.DposCfg.CandidateScheduleSize,
		BackupScheduleSize:    cfg.DposCfg.BackupScheduleSize,
		EpchoInterval:         cfg.DposCfg.EpchoInterval,
		FreezeEpchoSize:       cfg.DposCfg.FreezeEpchoSize,
		AccountName:           cfg.DposName,
		SystemName:            cfg.SysName,
		SystemURL:             cfg.ChainURL,
		ExtraBlockReward:      cfg.DposCfg.ExtraBlockReward,
		BlockReward:           cfg.DposCfg.BlockReward,
		Decimals:              cfg.SysTokenDecimals,
		AssetID:               cfg.SysTokenID,
		ReferenceTime:         cfg.ReferenceTime,
	}
}

// SetupGenesisBlock The returned chain configuration is never nil.
func SetupGenesisBlock(db fdb.Database, genesis *Genesis) (chainCfg *params.ChainConfig, dcfg *dpos.Config, hash common.Hash, err error) {
	chainCfg = params.DefaultChainconfig
	dcfg = dpos.DefaultConfig
	hash = common.Hash{}
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(e.(string))
		}
	}()

	if genesis != nil && genesis.Config == nil {
		return params.DefaultChainconfig, dposConfig(params.DefaultChainconfig), common.Hash{}, errGenesisNoConfig
	}

	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			genesis = DefaultGenesis()
		}
		block, err := genesis.Commit(db)
		log.Info("Writing genesis block", "hash", block.Hash().Hex())

		return genesis.Config, dposConfig(genesis.Config), block.Hash(), err
	}

	// Check whether the genesis block is already written.
	if genesis != nil {
		blk, _ := genesis.ToBlock(nil)
		hash := blk.Hash()
		if hash != stored {
			return genesis.Config, dposConfig(genesis.Config), hash, &GenesisMismatchError{stored, hash}
		}
	}
	newcfg := genesis.configOrDefault(stored)

	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {
		return newcfg, dposConfig(newcfg), common.Hash{}, fmt.Errorf("missing block number for head header hash")
	}

	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		return newcfg, dposConfig(newcfg), common.Hash{}, fmt.Errorf("Found genesis block without chain config")
	}
	am.SetAccountNameConfig(&am.Config{
		AccountNameLevel:     storedcfg.AccountNameCfg.Level,
		AccountNameLength:    storedcfg.AccountNameCfg.Length,
		SubAccountNameLength: storedcfg.AccountNameCfg.SubLength,
	})
	at.SetAssetNameConfig(&at.Config{
		AssetNameLength:    storedcfg.AssetNameCfg.Length,
		AssetNameLevel:     storedcfg.AssetNameCfg.Level,
		SubAssetNameLength: storedcfg.AssetNameCfg.SubLength,
	})
	am.SetSysName(common.StrToName(storedcfg.AccountName))
	fm.SetFeeManagerName(common.StrToName(storedcfg.FeeName))
	return storedcfg, dposConfig(storedcfg), stored, nil
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db fdb.Database) (*types.Block, []*types.Receipt) {
	if db == nil {
		db = memdb.NewMemDatabase()
	}
	detailTx := &types.DetailTx{}
	var internals []*types.DetailAction
	am.SetAccountNameConfig(&am.Config{
		AccountNameLevel:     g.Config.AccountNameCfg.Level,
		AccountNameLength:    g.Config.AccountNameCfg.Length,
		SubAccountNameLength: g.Config.AccountNameCfg.SubLength,
	})
	at.SetAssetNameConfig(&at.Config{
		AssetNameLength:    g.Config.AssetNameCfg.Length,
		AssetNameLevel:     g.Config.AssetNameCfg.Level,
		SubAssetNameLength: g.Config.AssetNameCfg.SubLength,
	})
	am.SetSysName(common.StrToName(g.Config.AccountName))
	fm.SetFeeManagerName(common.StrToName(g.Config.FeeName))
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
		if len(node) == 0 {
			continue
		}
		if _, err := enode.ParseV4(node); err != nil {
			panic(fmt.Sprintf("genesis bootnodes err: %v in %v", err, node))
		}
	}

	actActions := []*types.Action{}
	timestamp := g.Timestamp * uint64(time.Millisecond)
	g.Config.ReferenceTime = timestamp
	if err := dpos.Genesis(dposConfig(g.Config), statedb, timestamp, number.Uint64()); err != nil {
		panic(fmt.Sprintf("genesis dpos err %v", err))
	}

	chainName := common.Name(g.Config.ChainName)
	accoutName := common.Name(g.Config.AccountName)
	// chain name
	act := &am.AccountAction{
		AccountName: chainName,
		PublicKey:   common.PubKey{},
	}
	payload, _ := rlp.EncodeToBytes(act)
	actActions = append(actActions, types.NewAction(
		types.CreateAccount,
		common.Name(""),
		accoutName,
		0,
		0,
		0,
		big.NewInt(0),
		payload,
	))

	for _, account := range g.AllocAccounts {
		pname := common.Name("")
		slt := strings.Split(account.Name, ".")
		if len(slt) > 1 {
			pname = common.Name(slt[0])
		}
		act := &am.AccountAction{
			AccountName: common.StrToName(account.Name),
			PublicKey:   account.PubKey,
		}
		payload, _ := rlp.EncodeToBytes(act)
		actActions = append(actActions, types.NewAction(
			types.CreateAccount,
			pname,
			accoutName,
			0,
			0,
			0,
			big.NewInt(0),
			payload,
		))
	}

	for index, action := range actActions {
		internalLogs, err := accountManager.Process(&types.AccountManagerContext{Action: action, Number: 0})
		if err != nil {
			panic(fmt.Sprintf("genesis create account %v,err %v", index, err))
		}
		internals = append(internals, &types.DetailAction{InternalActions: internalLogs})
	}

	astActions := []*types.Action{}
	for _, asset := range g.AllocAssets {
		pname := common.Name("")
		slt := strings.Split(asset.Name, ".")
		if len(slt) > 1 {
			if ast, _ := accountManager.GetAssetInfoByName(slt[0]); ast == nil {
				panic(fmt.Sprintf("parent asset not exist %v", ast.AssetName))
			} else {
				pname = ast.Owner
			}
		}
		ast := &at.AssetObject{
			AssetName:  asset.Name,
			Symbol:     asset.Symbol,
			Amount:     asset.Amount,
			Decimals:   asset.Decimals,
			Founder:    common.StrToName(asset.Founder),
			Owner:      common.StrToName(asset.Owner),
			UpperLimit: asset.UpperLimit,
		}
		payload, _ := rlp.EncodeToBytes(ast)
		astActions = append(astActions, types.NewAction(
			types.IssueAsset,
			pname,
			accoutName,
			0,
			0,
			0,
			big.NewInt(0),
			payload,
		))
	}

	for index, action := range astActions {
		internalLogs, err := accountManager.Process(&types.AccountManagerContext{Action: action, Number: 0})
		if err != nil {
			panic(fmt.Sprintf("genesis create asset %v,err %v", index, err))
		}
		internals = append(internals, &types.DetailAction{InternalActions: internalLogs})
	}

	if ok, err := accountManager.AccountIsExist(common.StrToName(g.Config.SysName)); !ok {
		panic(fmt.Sprintf("system is not exist %v", err))
	}
	if ok, err := accountManager.AccountIsExist(common.StrToName(g.Config.AccountName)); !ok {
		panic(fmt.Sprintf("account is not exist %v", err))
	}
	if ok, err := accountManager.AccountIsExist(common.StrToName(g.Config.DposName)); !ok {
		panic(fmt.Sprintf("dpos is not exist %v", err))
	}
	if ok, err := accountManager.AccountIsExist(common.StrToName(g.Config.FeeName)); !ok {
		panic(fmt.Sprintf("fee is not exist %v", err))
	}
	assetInfo, err := accountManager.GetAssetInfoByName(g.Config.SysToken)
	if err != nil {
		panic(fmt.Sprintf("genesis system asset err %v", err))
	}

	g.Config.SysTokenID = assetInfo.AssetId
	g.Config.SysTokenDecimals = assetInfo.Decimals

	sys := dpos.NewSystem(statedb, dposConfig(g.Config))
	for _, candidate := range g.AllocCandidates {
		if err := sys.SetCandidate(&dpos.CandidateInfo{
			Name:          candidate.Name,
			URL:           candidate.URL,
			Quantity:      big.NewInt(0),
			TotalQuantity: big.NewInt(0),
			Height:        number.Uint64(),
		}); err != nil {
			panic(fmt.Sprintf("genesis create candidate err %v", err))
		}
	}
	if err := sys.UpdateElectedCandidates(dpos.LastEpcho, dpos.LastEpcho, number.Uint64()); err != nil {
		panic(fmt.Sprintf("genesis create candidate err %v", err))
	}

	root := statedb.IntermediateRoot()
	gjson, err := json.Marshal(g)
	if err != nil {
		panic(fmt.Sprintf("genesis json marshal json err %v", err))
	}

	head := &types.Header{
		Number:     number,
		Time:       new(big.Int).SetUint64(timestamp),
		ParentHash: common.Hash{},
		Extra:      gjson,
		GasLimit:   g.GasLimit,
		GasUsed:    0,
		Difficulty: g.Difficulty,
		Coinbase:   common.StrToName(g.Config.SysName),
		Root:       root,
	}

	actions := []*types.Action{}
	actions = append(actions, actActions...)
	actions = append(actions, astActions...)
	tx := types.NewTransaction(0, big.NewInt(0), actions...)
	receipt := types.NewReceipt(root[:], 0, 0)
	receipt.TxHash = tx.Hash()
	for index := range actions {
		receipt.ActionResults = append(receipt.ActionResults, &types.ActionResult{
			Status:  1,
			Index:   uint64(index),
			GasUsed: 0,
		})
	}

	detailTx.TxHash = receipt.TxHash
	detailTx.Actions = internals
	receipt.SetInternalTxsLog(detailTx)

	receipts := []*types.Receipt{receipt}
	block := types.NewBlock(head, []*types.Transaction{tx}, receipts)
	batch := db.NewBatch()
	roothash, err := statedb.Commit(batch, block.Hash(), block.NumberU64())
	if err != nil {
		panic(fmt.Sprintf("genesis statedb commit err: %v", err))
	}
	statedb.Database().TrieDB().Commit(roothash, false)
	if err := batch.Write(); err != nil {
		panic(fmt.Sprintf("genesis batch write err: %v", err))
	}
	return block, receipts
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db fdb.Database) (*types.Block, error) {
	block, receipts := g.ToBlock(db)
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

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	if g != nil {
		return g.Config
	}
	return params.DefaultChainconfig
}

// DefaultGenesis returns the ft net genesis block.
func DefaultGenesis() *Genesis {
	return &Genesis{
		Config:          params.DefaultChainconfig,
		Timestamp:       1555776000000, // 2019-04-21 00:00:00
		GasLimit:        params.GenesisGasLimit,
		Difficulty:      params.GenesisDifficulty,
		AllocAccounts:   DefaultGenesisAccounts(),
		AllocCandidates: DefaultGenesisCandidates(),
		AllocAssets:     DefaultGenesisAssets(),
	}
}

// DefaultGenesisAccounts returns the ft net genesis accounts.
func DefaultGenesisAccounts() []*GenesisAccount {
	return []*GenesisAccount{
		&GenesisAccount{
			Name:   params.DefaultChainconfig.SysName,
			PubKey: common.HexToPubKey("047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"),
		},
		&GenesisAccount{
			Name:   params.DefaultChainconfig.AccountName,
			PubKey: common.HexToPubKey(""),
		},
		&GenesisAccount{
			Name:   params.DefaultChainconfig.DposName,
			PubKey: common.HexToPubKey(""),
		},
		&GenesisAccount{
			Name:   params.DefaultChainconfig.FeeName,
			PubKey: common.HexToPubKey(""),
		},
	}
}

// DefaultGenesisCandidates returns the ft net genesis candidates.
func DefaultGenesisCandidates() []*GenesisCandidate {
	return []*GenesisCandidate{}
}

// DefaultGenesisAssets returns the ft net genesis assets.
func DefaultGenesisAssets() []*GenesisAsset {
	supply := new(big.Int)
	supply.SetString("100000000000000000000000000000", 10)
	return []*GenesisAsset{
		&GenesisAsset{
			Name:       params.DefaultChainconfig.SysToken,
			Symbol:     "ft",
			Amount:     supply,
			Decimals:   18,
			Owner:      params.DefaultChainconfig.SysName,
			Founder:    params.DefaultChainconfig.SysName,
			UpperLimit: supply,
		},
	}
}
