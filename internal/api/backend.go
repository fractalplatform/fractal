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

// Package api implements the general API functions.
package api

import (
	"context"
	"math/big"

	"github.com/fractalplatform/fractal/consensus"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/wallet"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// ftservice API
	ChainDb() fdb.Database
	ChainConfig() *params.ChainConfig
	SuggestPrice(ctx context.Context) (*big.Int, error)

	// BlockChain API
	CurrentBlock() *types.Block
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error)
	BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error)
	GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) ([]*types.Receipt, error)
	GetDetailTxsLog(ctx context.Context, hash common.Hash) ([]*types.DetailTx, error)
	GetBlockDetailLog(ctx context.Context, blockNr rpc.BlockNumber) *types.BlockAndResult
	GetTd(blockHash common.Hash) *big.Int
	GetEVM(ctx context.Context, account *accountmanager.AccountManager, state *state.StateDB, from common.Name, assetID uint64, gasPrice *big.Int, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error)
	GetDetailTxByFilter(ctx context.Context, filterFn func(common.Name) bool, blockNr rpc.BlockNumber, lookbackNum uint64) []*types.DetailTx
	GetTxsByFilter(ctx context.Context, filterFn func(common.Name) bool, blockNr rpc.BlockNumber, lookbackNum uint64) []common.Hash

	// TxPool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetPoolTransactions() ([]*types.Transaction, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Name][]*types.Transaction, map[common.Name][]*types.Transaction)

	//Account API
	GetAccountManager() (*accountmanager.AccountManager, error)

	SetGasPrice(gasPrice *big.Int) bool

	//Wallet
	Wallet() *wallet.Wallet

	// P2P
	AddPeer(url string) error
	RemovePeer(url string) error
	AddTrustedPeer(url string) error
	RemoveTrustedPeer(url string) error
	PeerCount() int
	Peers() []string
	SelfNode() string

	Engine() consensus.IEngine

	APIs() []rpc.API
}

func GetAPIs(apiBackend Backend) []rpc.API {
	apis := []rpc.API{
		{
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "ft",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "ft",
			Version:   "1.0",
			Service:   NewPublicFractalAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "keystore",
			Version:   "1.0",
			Service:   NewPrivateKeyStoreAPI(apiBackend),
			Public:    true,
		},
		{
			Namespace: "account",
			Version:   "1.0",
			Service:   NewAccountAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "p2p",
			Version:   "1.0",
			Service:   NewPrivateP2pAPI(apiBackend),
			Public:    true,
		},
	}
	return append(apis, apiBackend.APIs()...)
}
