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

package api

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// PublicFractalAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicFractalAPI struct {
	b Backend
}

// NewPublicFractalAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicFractalAPI(b Backend) *PublicFractalAPI {
	return &PublicFractalAPI{b}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicFractalAPI) GasPrice(ctx context.Context) (*big.Int, error) {
	return s.b.SuggestPrice(ctx)
}

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicFractalAPI) SendRawTransaction(ctx context.Context, encodedTx hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, tx)
}

type SendArgs struct {
	ChainID    *big.Int         `json:"chainID"`
	ActionType types.ActionType `json:"actionType"`
	GasAssetID uint64           `json:"gasAssetId"`
	From       common.Name      `json:"from"`
	To         common.Name      `json:"to"`
	Nonce      uint64           `json:"nonce"`
	AssetID    uint64           `json:"assetId"`
	Gas        uint64           `json:"gas"`
	GasPrice   *big.Int         `json:"gasPrice"`
	Value      *big.Int         `json:"value"`
	Data       hexutil.Bytes    `json:"data"`
	Passphrase string           `json:"password"`
}

func (s *PublicFractalAPI) SendTransaction(ctx context.Context, args SendArgs) (common.Hash, error) {
	acct, err := s.b.GetAccountManager()
	if err != nil {
		return common.Hash{}, err
	}
	if acct == nil {
		return common.Hash{}, ErrGetAccounManagerErr
	}
	fromAcct, err := acct.GetAccountByName(args.From)
	if err != nil {
		return common.Hash{}, err
	}
	if fromAcct == nil {
		return common.Hash{}, errors.New("invalid user")
	}

	pubByte, _ := crypto.UnmarshalPubkey(fromAcct.Authors[0].Owner.(common.PubKey).Bytes())
	if !s.b.Wallet().HasAddress(crypto.PubkeyToAddress(*pubByte)) {
		return common.Hash{}, errors.New("user not in local wallet")
	}
	cacheAcct, err := s.b.Wallet().Find(crypto.PubkeyToAddress(*pubByte))
	if err != nil {
		return common.Hash{}, err
	}

	assetID := uint64(args.AssetID)
	gas := uint64(args.Gas)
	action := types.NewAction(args.ActionType, args.From, args.To, args.Nonce, assetID, gas, args.Value, args.Data)
	tx := types.NewTransaction(args.GasAssetID, args.GasPrice, action)

	tx, err = s.b.Wallet().SignTxWithPassphrase(cacheAcct, args.Passphrase, tx, action, args.ChainID)
	if err != nil {
		return common.Hash{}, err
	}

	return submitTransaction(ctx, s.b, tx)
}
