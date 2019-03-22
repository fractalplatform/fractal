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

package ftservice

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/ftservice/gasprice"
	"github.com/fractalplatform/fractal/p2p/enode"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/wallet"
)

// APIBackend implements ftserviceapi.Backend for full nodes
type APIBackend struct {
	ftservice *FtService
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

func (b *APIBackend) GetPoolTransactions() ([]*types.Transaction, error) {
	pending, err := b.ftservice.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs []*types.Transaction
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *APIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.ftservice.txPool.Get(hash)
}

func (b *APIBackend) Stats() (pending int, queued int) {
	return b.ftservice.txPool.Stats()
}

func (b *APIBackend) TxPoolContent() (map[common.Name][]*types.Transaction, map[common.Name][]*types.Transaction) {
	return b.ftservice.TxPool().Content()
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

func (b *APIBackend) GetBlockAndResult(ctx context.Context, blockNr rpc.BlockNumber) *types.BlockAndResult {
	hash := rawdb.ReadCanonicalHash(b.ftservice.chainDb, uint64(blockNr))
	if hash == (common.Hash{}) {
		return nil
	}
	block := b.ftservice.blockchain.GetBlockByNumber(uint64(blockNr))
	if block == nil {
		return nil
	}
	receipts := rawdb.ReadReceipts(b.ftservice.chainDb, hash, uint64(blockNr))
	txDetails := rawdb.ReadDetailTxs(b.ftservice.chainDb, hash, uint64(blockNr))
	return &types.BlockAndResult{
		Block:     block,
		Receipts:  receipts,
		DetailTxs: txDetails,
		Hash:      block.Hash(),
	}
}

func (b *APIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.ftservice.blockchain.GetTdByHash(blockHash)
}

func (b *APIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {

	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, _ := b.ftservice.miner.Pending()
		return block.Header(), nil
	}

	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ftservice.blockchain.CurrentBlock().Header(), nil
	}

	return b.ftservice.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *APIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, _ := b.ftservice.miner.Pending()
		return block, nil
	}

	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ftservice.blockchain.CurrentBlock(), nil
	}
	return b.ftservice.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

//
func (b *APIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.ftservice.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.ftservice.blockchain.StateAt(b.ftservice.blockchain.CurrentBlock().Root())
	return stateDb, header, err
}

func (b *APIBackend) GetEVM(ctx context.Context, account *accountmanager.AccountManager, state *state.StateDB, from common.Name, assetID uint64, gasPrice *big.Int, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	account.AddAccountBalanceByID(from, assetID, math.MaxBig256)
	vmError := func() error { return nil }

	evmcontext := &processor.EvmContext{
		ChainContext:  b.ftservice.BlockChain(),
		EgnineContext: b.ftservice.Engine(),
	}

	context := processor.NewEVMContext(from, assetID, gasPrice, header, evmcontext, nil)
	return vm.NewEVM(context, account, state, b.ChainConfig(), vmCfg), vmError, nil
}

func (b *APIBackend) SetGasPrice(gasPrice *big.Int) bool {
	b.ftservice.SetGasPrice(gasPrice)
	return true
}

func (b *APIBackend) Wallet() *wallet.Wallet {
	return b.ftservice.Wallet()
}

func (b *APIBackend) GetAccountManager() (*accountmanager.AccountManager, error) {
	sdb, err := b.ftservice.blockchain.State()
	if err != nil {
		return nil, err
	}
	acctm, err := accountmanager.NewAccountManager(sdb)
	if err != nil {
		return nil, err
	}
	return acctm, nil
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

// SelfNode returns the local node's endpoint information.
func (b *APIBackend) SelfNode() string {
	return b.ftservice.p2pServer.Self().String()
}

// APIs returns apis
func (b *APIBackend) Engine() consensus.IEngine {
	return b.ftservice.engine
}

// APIs returns apis
func (b *APIBackend) APIs() []rpc.API {
	return b.ftservice.miner.APIs(b.ftservice.blockchain)
}
