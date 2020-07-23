// Copyright 2018 The OEX Team Authors
// This file is part of the OEX project.
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

package oexservice

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/oexplatform/oexchain/accountmanager"
	"github.com/oexplatform/oexchain/blockchain"
	"github.com/oexplatform/oexchain/common"
	"github.com/oexplatform/oexchain/consensus"
	"github.com/oexplatform/oexchain/feemanager"
	"github.com/oexplatform/oexchain/oexservice/gasprice"
	"github.com/oexplatform/oexchain/p2p/enode"
	"github.com/oexplatform/oexchain/params"
	"github.com/oexplatform/oexchain/processor"
	"github.com/oexplatform/oexchain/processor/vm"
	"github.com/oexplatform/oexchain/rawdb"
	"github.com/oexplatform/oexchain/rpc"
	"github.com/oexplatform/oexchain/snapshot"
	"github.com/oexplatform/oexchain/state"
	"github.com/oexplatform/oexchain/txpool"
	"github.com/oexplatform/oexchain/types"
	"github.com/oexplatform/oexchain/utils/fdb"
)

// APIBackend implements oexservice api.Backend for full nodes
type APIBackend struct {
	ftservice *OEXService
	gpo       *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *APIBackend) ChainConfig() *params.ChainConfig {
	return b.ftservice.chainConfig
}
func (b *APIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *APIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.ftservice.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.ftservice.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *APIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.ftservice.txPool.AddLocal(signedTx)
}

func (b *APIBackend) TxPool() *txpool.TxPool {
	return b.ftservice.TxPool()
}

func (b *APIBackend) ChainDb() fdb.Database {
	return b.ftservice.chainDb
}

func (b *APIBackend) CurrentBlock() *types.Block {
	return b.ftservice.blockchain.CurrentBlock()
}

func (b *APIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.ftservice.blockchain.GetBlockByHash(hash), nil
}

func (b *APIBackend) GetReceipts(ctx context.Context, hash common.Hash) ([]*types.Receipt, error) {
	if number := rawdb.ReadHeaderNumber(b.ftservice.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.ftservice.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *APIBackend) GetDetailTxsLog(ctx context.Context, hash common.Hash) ([]*types.DetailTx, error) {
	if number := rawdb.ReadHeaderNumber(b.ftservice.chainDb, hash); number != nil {
		return rawdb.ReadDetailTxs(b.ftservice.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *APIBackend) GetBlockDetailLog(ctx context.Context, blockNr rpc.BlockNumber) *types.BlockAndResult {
	hash := rawdb.ReadCanonicalHash(b.ftservice.chainDb, uint64(blockNr))
	if hash == (common.Hash{}) {
		return nil
	}
	receipts := rawdb.ReadReceipts(b.ftservice.chainDb, hash, uint64(blockNr))
	txDetails := rawdb.ReadDetailTxs(b.ftservice.chainDb, hash, uint64(blockNr))
	return &types.BlockAndResult{
		Receipts:  receipts,
		DetailTxs: txDetails,
	}
}

func (b *APIBackend) GetTxsByFilter(ctx context.Context, filterFn func(common.Name) bool, blockNr, lookforwardNum uint64) *types.AccountTxs {
	if lookforwardNum > 128 {
		lookforwardNum = 128
	}

	lastnum := int64(blockNr + lookforwardNum)
	txhhpairs := make([]*types.TxHeightHashPair, 0)
	for ublocknum := int64(blockNr); ublocknum <= lastnum; ublocknum++ {
		hash := rawdb.ReadCanonicalHash(b.ftservice.chainDb, uint64(ublocknum))
		if hash == (common.Hash{}) {
			continue
		}

		blockBody := rawdb.ReadBody(b.ftservice.chainDb, hash, uint64(ublocknum))
		if blockBody == nil {
			continue
		}
		batchTxs := blockBody.Transactions

		for _, tx := range batchTxs {
			for _, act := range tx.GetActions() {
				if filterFn(act.Sender()) || filterFn(act.Recipient()) {
					hhpair := &types.TxHeightHashPair{
						Hash:   tx.Hash(),
						Height: uint64(ublocknum),
					}
					txhhpairs = append(txhhpairs, hhpair)
					break
				}
			}
		}
	}

	accountTxs := &types.AccountTxs{
		Txs:                     txhhpairs,
		IrreversibleBlockHeight: b.ftservice.engine.CalcBFTIrreversible(),
		EndHeight:               uint64(lastnum),
	}

	return accountTxs
}

func (b *APIBackend) GetDetailTxByFilter(ctx context.Context, filterFn func(common.Name) bool, blockNr, lookbackNum uint64) []*types.DetailTx {
	var lastnum int64
	if lookbackNum > blockNr {
		lastnum = 0
	} else {
		lastnum = int64(blockNr - lookbackNum)
	}
	txdetails := make([]*types.DetailTx, 0)

	for ublocknum := int64(blockNr); ublocknum >= lastnum; ublocknum-- {
		hash := rawdb.ReadCanonicalHash(b.ftservice.chainDb, uint64(ublocknum))
		if hash == (common.Hash{}) {
			continue
		}

		batchTxdetails := rawdb.ReadDetailTxs(b.ftservice.chainDb, hash, uint64(ublocknum))
		for _, txd := range batchTxdetails {
			newIntxs := make([]*types.DetailAction, 0)
			for _, intx := range txd.Actions {
				newInactions := make([]*types.InternalAction, 0)
				for _, inlog := range intx.InternalActions {
					if filterFn(inlog.Action.From) || filterFn(inlog.Action.To) {
						newInactions = append(newInactions, inlog)
					}
				}
				if len(newInactions) > 0 {
					newIntxs = append(newIntxs, &types.DetailAction{InternalActions: newInactions})
				}
			}

			if len(newIntxs) > 0 {
				txdetails = append(txdetails, &types.DetailTx{TxHash: txd.TxHash, Actions: newIntxs})
			}
		}
	}

	return txdetails
}

func (b *APIBackend) GetBadBlocks(ctx context.Context) ([]*types.Block, error) {
	return b.ftservice.blockchain.BadBlocks(), nil
}

func (b *APIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.ftservice.blockchain.GetTdByHash(blockHash)
}

func (b *APIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Header {
	if blockNr == rpc.LatestBlockNumber {
		return b.ftservice.blockchain.CurrentBlock().Header()
	}
	return b.ftservice.blockchain.GetHeaderByNumber(uint64(blockNr))
}

func (b *APIBackend) HeaderByHash(ctx context.Context, hash common.Hash) *types.Header {
	return b.ftservice.blockchain.GetHeaderByHash(hash)
}

func (b *APIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Block {
	if blockNr == rpc.LatestBlockNumber {
		return b.ftservice.blockchain.CurrentBlock()
	}
	return b.ftservice.blockchain.GetBlockByNumber(uint64(blockNr))
}

func (b *APIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	header := b.HeaderByNumber(ctx, blockNr)
	if header == nil {
		return nil, nil, nil
	}
	stateDb, err := b.ftservice.blockchain.StateAt(b.ftservice.blockchain.CurrentBlock().Root())
	return stateDb, header, err
}

func (b *APIBackend) GetEVM(ctx context.Context, account *accountmanager.AccountManager, state *state.StateDB, from common.Name, to common.Name, assetID uint64, gasPrice *big.Int, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	account.AddAccountBalanceByID(from, assetID, math.MaxBig256)
	vmError := func() error { return nil }

	evmContext := &processor.EvmContext{
		ChainContext:  b.ftservice.BlockChain(),
		EngineContext: b.ftservice.Engine(),
	}

	context := processor.NewEVMContext(from, to, assetID, gasPrice, header, evmContext, nil)
	return vm.NewEVM(context, account, state, b.ChainConfig(), vmCfg), vmError, nil
}

func (b *APIBackend) SetGasPrice(gasPrice *big.Int) bool {
	return b.ftservice.SetGasPrice(gasPrice)
}

func (b *APIBackend) ForkStatus(statedb *state.StateDB) (*blockchain.ForkConfig, blockchain.ForkInfo, error) {
	return b.ftservice.BlockChain().ForkStatus(statedb)
}

func (b *APIBackend) GetAccountManager() (*accountmanager.AccountManager, error) {
	sdb, err := b.ftservice.blockchain.State()
	if err != nil {
		return nil, err
	}
	return accountmanager.NewAccountManager(sdb)
}

//GetFeeManager get fee manager
func (b *APIBackend) GetFeeManager() (*feemanager.FeeManager, error) {
	sdb, err := b.ftservice.blockchain.State()
	if err != nil {
		return nil, err
	}
	acctm, err := accountmanager.NewAccountManager(sdb)
	if err != nil {
		return nil, err
	}

	fm := feemanager.NewFeeManager(sdb, acctm)
	return fm, nil
}

//GetFeeManagerByTime get fee manager
func (b *APIBackend) GetFeeManagerByTime(time uint64) (*feemanager.FeeManager, error) {
	sdb, err := b.ftservice.blockchain.State()
	if err != nil {
		return nil, err
	}

	snapshotManager := snapshot.NewSnapshotManager(sdb)
	state, err := snapshotManager.GetSnapshotState(time)
	if err != nil {
		return nil, err
	}

	acctm, err := accountmanager.NewAccountManager(state)
	if err != nil {
		return nil, err
	}

	fm := feemanager.NewFeeManager(state, acctm)
	return fm, nil
}

// AddPeer add a P2P peer
func (b *APIBackend) AddPeer(url string) error {
	node, err := enode.ParseV4(url)
	if err == nil {
		b.ftservice.p2pServer.AddPeer(node)
	}
	return err
}

// RemovePeer remove a P2P peer
func (b *APIBackend) RemovePeer(url string) error {
	node, err := enode.ParseV4(url)
	if err == nil {
		b.ftservice.p2pServer.RemovePeer(node)
	}
	return err
}

// AddTrustedPeer allows a remote node to always connect, even if slots are full
func (b *APIBackend) AddTrustedPeer(url string) error {
	node, err := enode.ParseV4(url)
	if err == nil {
		b.ftservice.p2pServer.AddTrustedPeer(node)
	}
	return err
}

// RemoveTrustedPeer removes a remote node from the trusted peer set, but it
// does not disconnect it automatically.
func (b *APIBackend) RemoveTrustedPeer(url string) error {
	node, err := enode.ParseV4(url)
	if err == nil {
		b.ftservice.p2pServer.RemoveTrustedPeer(node)
	}
	return err
}

// SeedNodes returns all seed nodes.
func (b *APIBackend) SeedNodes() []string {
	nodes := b.ftservice.p2pServer.SeedNodes()
	ns := make([]string, len(nodes))
	for i, node := range nodes {
		ns[i] = node.String()
	}
	return ns
}

// PeerCount returns the number of connected peers.
func (b *APIBackend) PeerCount() int {
	return b.ftservice.p2pServer.PeerCount()
}

// Peers returns all connected peers.
func (b *APIBackend) Peers() []string {
	ps := b.ftservice.p2pServer.Peers()
	peers := make([]string, len(ps))
	for i, peer := range ps {
		peers[i] = peer.Node().String()
	}
	return peers
}

// BadNodesCount returns the number of bad nodes.
func (b *APIBackend) BadNodesCount() int {
	return b.ftservice.p2pServer.BadNodesCount()
}

// BadNodes returns all bad nodes.
func (b *APIBackend) BadNodes() []string {
	nodes := b.ftservice.p2pServer.BadNodes()
	ns := make([]string, len(nodes))
	for i, node := range nodes {
		ns[i] = node.String()
	}
	return ns
}

// AddBadNode add a bad Node and would cause the node disconnected
func (b *APIBackend) AddBadNode(url string) error {
	node, err := enode.ParseV4(url)
	if err == nil {
		b.ftservice.p2pServer.AddBadNode(node, nil)
	}
	return err
}

// RemoveBadNode add a bad Node and would cause the node disconnected
func (b *APIBackend) RemoveBadNode(url string) error {
	node, err := enode.ParseV4(url)
	if err == nil {
		b.ftservice.p2pServer.RemoveBadNode(node)
	}
	return err
}

// SelfNode returns the local node's endpoint information.
func (b *APIBackend) SelfNode() string {
	return b.ftservice.p2pServer.Self().String()
}

func (b *APIBackend) Engine() consensus.IEngine {
	return b.ftservice.engine
}

//SetStatePruning set state pruning
func (b *APIBackend) SetStatePruning(enable bool) (bool, uint64) {
	return b.ftservice.blockchain.StatePruning(enable)
}

// APIs returns apis
func (b *APIBackend) APIs() []rpc.API {
	return b.ftservice.miner.APIs(b.ftservice.blockchain)
}
