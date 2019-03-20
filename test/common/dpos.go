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

package common

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	args "github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	signer = types.NewSigner(params.DefaultChainconfig.ChainID)
)

// Account
type Account struct {
	name     common.Name
	priv     *ecdsa.PrivateKey
	feeid    uint64
	nonce    uint64
	getnonce func(common.Name) uint64
}

func NewAccount(name common.Name, priv *ecdsa.PrivateKey, feeid uint64, nonce uint64, getnonce func(common.Name) uint64) *Account {
	return &Account{
		name:     name,
		priv:     priv,
		feeid:    feeid,
		nonce:    nonce,
		getnonce: getnonce,
	}
}

func (acc *Account) PubKey() common.PubKey {
	return common.BytesToPubKey(crypto.FromECDSAPub(&acc.priv.PublicKey))
}

// CreateAccount
func (acc *Account) CreateAccount(to common.Name, value *big.Int, id uint64, gas uint64, pubkey common.PubKey) []byte {
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.CreateAccount, acc.name, to, acc.nonce, id, gas, value, pubkey.Bytes())
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

// Transfer
func (acc *Account) Transfer(to common.Name, value *big.Int, id uint64, gas uint64) []byte {
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.Transfer, acc.name, to, acc.nonce, id, gas, value, nil)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

// RegCadidate
func (acc *Account) RegCadidate(to common.Name, value *big.Int, id uint64, gas uint64, url string, state *big.Int) []byte {
	arg := &args.RegisterCadidate{
		Url:   url,
		Stake: state,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.RegCadidate, acc.name, to, acc.nonce, id, gas, value, payload)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

// UpdateCadidate
func (acc *Account) UpdateCadidate(to common.Name, value *big.Int, id uint64, gas uint64, url string, state *big.Int) []byte {
	arg := &args.UpdateCadidate{
		Url:   url,
		Stake: state,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.UpdateCadidate, acc.name, to, acc.nonce, id, gas, value, payload)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

// UnRegCadidate
func (acc *Account) UnRegCadidate(to common.Name, value *big.Int, id uint64, gas uint64) []byte {
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.UnregCadidate, acc.name, to, acc.nonce, id, gas, value, nil)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

// VoteCadidate
func (acc *Account) VoteCadidate(to common.Name, value *big.Int, id uint64, gas uint64, cadidate string, state *big.Int) []byte {
	arg := &args.VoteCadidate{
		Cadidate: cadidate,
		Stake:    state,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.VoteCadidate, acc.name, to, acc.nonce, id, gas, value, payload)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

// ChangeCadidate
func (acc *Account) ChangeCadidate(to common.Name, value *big.Int, id uint64, gas uint64, cadidate string) []byte {
	arg := &args.ChangeCadidate{
		Cadidate: cadidate,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.ChangeCadidate, acc.name, to, acc.nonce, id, gas, value, payload)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

func (acc *Account) UnvoteCadidate(to common.Name, value *big.Int, id uint64, gas uint64) []byte {
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.UnvoteCadidate, acc.name, to, acc.nonce, id, gas, value, nil)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}

// UnvoteVoter
func (acc *Account) UnvoteVoter(to common.Name, value *big.Int, id uint64, gas uint64, voter string) []byte {
	arg := &args.RemoveVoter{
		Voter: voter,
	}
	payload, err := rlp.EncodeToBytes(arg)
	if err != nil {
		panic(err)
	}
	if acc.getnonce != nil {
		acc.nonce = acc.getnonce(acc.name)
	}
	action := types.NewAction(types.RemoveVoter, acc.name, to, acc.nonce, id, gas, value, payload)
	if acc.getnonce == nil {
		acc.nonce++
	}

	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	if err := types.SignAction(action, tx, signer, acc.priv); err != nil {
		panic(err)
	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	return rawtx
}
