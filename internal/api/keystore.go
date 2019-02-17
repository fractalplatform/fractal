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
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

type PrivateKeyStoreAPI struct {
	b Backend
}

func NewPrivateKeyStoreAPI(b Backend) *PrivateKeyStoreAPI {
	return &PrivateKeyStoreAPI{b}
}

// NewAccount generates a new key and stores it into the key directory.
func (api *PrivateKeyStoreAPI) NewAccount(ctx context.Context, passphrase string) (map[string]interface{}, error) {
	a, err := api.b.Wallet().NewAccount(passphrase)
	if err != nil {
		return nil, err
	}

	key, err := api.b.Wallet().GetPrivateKey(a, passphrase)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"address":    a.Addr,
		"path":       a.Path,
		"publicKey":  hexutil.Bytes(crypto.FromECDSAPub(&key.PrivateKey.PublicKey)),
		"privateKey": hexutil.Bytes(crypto.FromECDSA(key.PrivateKey)),
	}, nil
}

// Delete deletes a account by passsphrase.
func (api *PrivateKeyStoreAPI) Delete(ctx context.Context, addr common.Address, passphrase string) error {
	a, err := api.b.Wallet().Find(addr)
	if err != nil {
		return err
	}
	return api.b.Wallet().Delete(a, passphrase)
}

// Update changes the passphrase of an existing account.
func (api *PrivateKeyStoreAPI) Update(ctx context.Context, addr common.Address, passphrase, newPassphrase string) error {
	a, err := api.b.Wallet().Find(addr)
	if err != nil {
		return err
	}
	return api.b.Wallet().Update(a, passphrase, newPassphrase)
}

// ImportRawKey stores the given key into the key directory, encrypting it with the passphrase.
func (api *PrivateKeyStoreAPI) ImportRawKey(ctx context.Context, privkey string, passphrase string) (map[string]interface{}, error) {
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return nil, err
	}
	a, err := api.b.Wallet().ImportECDSA(key, passphrase)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"address": a.Addr,
		"path":    a.Path,
	}, nil
}

// ExportRawKey export account private key .
func (api *PrivateKeyStoreAPI) ExportRawKey(ctx context.Context, addr common.Address, passphrase string) (hexutil.Bytes, error) {
	a, err := api.b.Wallet().Find(addr)
	if err != nil {
		return nil, err
	}
	key, err := api.b.Wallet().GetPrivateKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(crypto.FromECDSA(key.PrivateKey)), nil
}

// ListAccount returns all key files
func (api *PrivateKeyStoreAPI) ListAccount(ctx context.Context) ([]map[string]interface{}, error) {
	accounts := api.b.Wallet().Accounts()
	ret := make([]map[string]interface{}, 0)
	for _, account := range accounts {
		tmpa := map[string]interface{}{
			"address": account.Addr,
			"path":    account.Path,
		}
		ret = append(ret, tmpa)
	}
	return ret, nil
}

// SignTransaction sign transaction and return raw hex .
func (api *PrivateKeyStoreAPI) SignTransaction(ctx context.Context, addr common.Address, passphrase string, tx *types.Transaction) (hexutil.Bytes, error) {
	a, err := api.b.Wallet().Find(addr)
	if err != nil {
		return nil, err
	}
	signed, err := api.b.Wallet().SignTxWithPassphrase(a, passphrase, tx, tx.GetActions()[0], api.b.ChainConfig().ChainID)
	if err != nil {
		return nil, err
	}
	rawtx, err := rlp.EncodeToBytes(signed)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(rawtx), nil
}

// SignData sign data and return raw hex
func (api *PrivateKeyStoreAPI) SignData(ctx context.Context, addr common.Address, passphrase string, data hexutil.Bytes) (hexutil.Bytes, error) {
	a, err := api.b.Wallet().NewAccount(passphrase)
	if err != nil {
		return nil, err
	}
	key, err := api.b.Wallet().GetPrivateKey(a, passphrase)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(data[:], key.PrivateKey)
	if err != nil {
		return nil, err
	}

	return hexutil.Bytes(sig), nil
}

func (api *PrivateKeyStoreAPI) BindAccountName(ctx context.Context, addr common.Address, passphrase string, accountName string) error {
	a, err := api.b.Wallet().Find(addr)
	if err != nil {
		return err
	}
	return api.b.Wallet().BindAccountNameAddr(a, passphrase, accountName)
}