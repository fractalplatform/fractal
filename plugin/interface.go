// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
)

// IPM plugin manager interface.
type IPM interface {
	IAccount
	IAsset
	IConsensus
	IFee
	ISigner
	ExecTx(tx *types.Transaction, ctx *Context, fromSol bool) ([]byte, error)
	IsPlugin(name string) bool
	InitChain(json.RawMessage, *types.Header, *params.ChainConfig) ([]*types.Transaction, error)
}

type IContract interface {
}

// IAccount account manager interface.
type IAccount interface {
	// external funcs
	GetNonce(accountName string) (uint64, error)
	SetNonce(accountName string, nonce uint64) error
	CreateAccount(accountName string, pubKey string, description string) ([]byte, error)
	GetCode(accountName string) ([]byte, error)
	GetCodeHash(accountName string) (common.Hash, error)
	SetCode(accountName string, code []byte) error
	GetBalance(accountName string, assetID uint64) (*big.Int, error)
	CanTransfer(accountName string, assetID uint64, value *big.Int) error
	TransferAsset(from, to string, assetID uint64, value *big.Int) error
	RecoverTx(signer ISigner, tx *types.Transaction) error
	AccountSign(accountName string, priv *ecdsa.PrivateKey, signer ISigner, signHash func(chainID *big.Int) common.Hash) ([]byte, error)
	AccountVerify(accountName string, signer ISigner, signature []byte, signHash func(chainID *big.Int) common.Hash) (*ecdsa.PublicKey, error)
	ChangePubKey(accountName string, pubKey string) error

	AccountIsExist(accountName string) error                  // for api
	GetAccountByName(accountName string) (interface{}, error) //for api
	// internal func
	checkCreateAccount(accountName string, pubKey string, description string) error
	getAccount(accountName string) (*Account, error)                          // for asset plugin
	addBalanceByID(accountName string, assetID uint64, amount *big.Int) error // for asset plugin
	subBalanceByID(accountName string, assetID uint64, amount *big.Int) error // for asset plugin

}

type IAsset interface {
	// external funcs
	IssueAsset(accountName string, assetName string, symbol string,
		amount *big.Int, decimals uint64, founder string, owner string,
		limit *big.Int, description string, am IAccount) ([]byte, error)
	IncreaseAsset(from, to string, assetID uint64, amount *big.Int, am IAccount) ([]byte, error)
	DestroyAsset(accountName string, assetID uint64, amount *big.Int, am IAccount) ([]byte, error)
	GetAssetID(assetName string) (uint64, error)
	GetAssetName(assetID uint64) (string, error)

	GetAssetInfoByName(assetName string) (*Asset, error) // for api
	GetAssetInfoByID(assetID uint64) (*Asset, error)     // for api

	// internal func
	checkIssueAsset(accountName string, assetName string, symbol string, amount *big.Int,
		decimals uint64, founder string, owner string, limit *big.Int, description string, am IAccount) error
	checkIncreaseAsset(from, to string, assetID uint64, amount *big.Int, am IAccount) error
}

type IConsensus interface {
	Init(genesisTime uint64, parent *types.Header)
	MineDelay(miner string) time.Duration
	Prepare(header *types.Header) error
	Finalize(header *types.Header, txs []*types.Transaction, receipts []*types.Receipt) (*types.Block, error)
	Seal(block *types.Block, priKey *ecdsa.PrivateKey, pm IPM) (*types.Block, error)
	//Difficult(header *types.Header) uint64
	Verify(header *types.Header) error
	VerifySeal(header *types.Header, pm IPM) error

	// RPC
	GetAllCandidates() []string
	GetCandidateInfo(miner string) *CandidateInfo
}

type IFee interface {
	DistributeGas(from string, gasMap map[types.DistributeKey]types.DistributeGas, assetID uint64, gasPrice *big.Int, am IAccount) ([]*types.GasDistribution, error)
}

type ISigner interface {
	Sign(signHash func(chainID *big.Int) common.Hash, prv *ecdsa.PrivateKey) ([]byte, error)
	Recover(signature []byte, signHash func(chainID *big.Int) common.Hash) ([]byte, error)
}
