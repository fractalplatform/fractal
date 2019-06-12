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

package sdk

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/abi"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	timeout = int64(time.Second) * 7
)

// Account account object
type Account struct {
	api      *API
	name     common.Name
	priv     *ecdsa.PrivateKey
	gasprice *big.Int
	feeid    uint64
	nonce    uint64 // nonce == math.MaxUint64, auto get
	checked  bool   // check result
	chainID  *big.Int
}

// NewAccount new account object
func NewAccount(api *API, name common.Name, priv *ecdsa.PrivateKey, feeid uint64, nonce uint64, checked bool, chainID *big.Int) *Account {
	return &Account{
		api:      api,
		name:     name,
		priv:     priv,
		gasprice: big.NewInt(1e10),
		feeid:    feeid,
		nonce:    nonce,
		checked:  checked,
		chainID:  chainID,
	}
}

// Pubkey account pub key
func (acc *Account) Pubkey() common.PubKey {
	return common.BytesToPubKey(crypto.FromECDSAPub(&acc.priv.PublicKey))
}

//=====================================================================================
//                       Transactions
//=====================================================================================

// CreateAccount new account
func (acc *Account) CreateAccount(to common.Name, value *big.Int, id uint64, gas uint64, newacct *accountmanager.CreateAccountAction) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}
	payload, _ := rlp.EncodeToBytes(newacct)
	action := types.NewAction(types.CreateAccount, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)

	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkCreateAccount(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// UpdateAccount update accout
func (acc *Account) UpdateAccount(to common.Name, value *big.Int, id uint64, gas uint64, newacct *accountmanager.UpdataAccountAction) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	bts, _ := rlp.EncodeToBytes(newacct)
	action := types.NewAction(types.UpdateAccount, acc.name, to, nonce, id, gas, value, bts, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkUpdateAccount(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// UpdateAccountAuthor update accout
func (acc *Account) UpdateAccountAuthor(to common.Name, value *big.Int, id uint64, gas uint64, newacct *accountmanager.AccountAuthorAction) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	bts, _ := rlp.EncodeToBytes(newacct)
	action := types.NewAction(types.UpdateAccountAuthor, acc.name, to, nonce, id, gas, value, bts, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkUpdateAccount(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// Transfer transfer tokens
func (acc *Account) Transfer(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	// transfer
	action := types.NewAction(types.Transfer, acc.name, to, nonce, id, gas, value, nil, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		//before
		checkedfunc, err = acc.checkTranfer(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// IssueAsset new asset
func (acc *Account) IssueAsset(to common.Name, value *big.Int, id uint64, gas uint64, asset *accountmanager.IssueAsset) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(asset)
	action := types.NewAction(types.IssueAsset, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkIssueAsset(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}
	if err != nil {
		return
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// UpdateAsset update asset
func (acc *Account) UpdateAsset(to common.Name, value *big.Int, id uint64, gas uint64, asset *accountmanager.UpdateAsset) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(asset)
	action := types.NewAction(types.UpdateAsset, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkUpdateAsset(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// IncreaseAsset update asset
func (acc *Account) IncreaseAsset(to common.Name, value *big.Int, id uint64, gas uint64, asset *accountmanager.IncAsset) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(asset)
	action := types.NewAction(types.IncreaseAsset, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkIncreaseAsset(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// DestroyAsset destory asset
func (acc *Account) DestroyAsset(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	action := types.NewAction(types.DestroyAsset, acc.name, to, nonce, id, gas, value, nil, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkDestroyAsset(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// SetAssetOwner update asset owner
func (acc *Account) SetAssetOwner(to common.Name, value *big.Int, id uint64, gas uint64, asset *accountmanager.UpdateAssetOwner) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(asset)
	action := types.NewAction(types.SetAssetOwner, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.checkSetAssetOwner(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// RegCandidate new candidate
func (acc *Account) RegCandidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.RegisterCandidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.RegCandidate, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekRegProdoucer(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		// after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// UpdateCandidate update candidate
func (acc *Account) UpdateCandidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.UpdateCandidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.UpdateCandidate, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekUpdateProdoucer(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		// after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// UnRegCandidate remove cadiate
func (acc *Account) UnRegCandidate(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	action := types.NewAction(types.UnregCandidate, acc.name, to, nonce, id, gas, value, nil, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		panic(err)
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekUnregProdoucer(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		// after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// RefundCandidate refund cadiate
func (acc *Account) RefundCandidate(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	action := types.NewAction(types.RefundCandidate, acc.name, to, nonce, id, gas, value, nil, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		panic(err)
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekRefundProdoucer(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		// after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// VoteCandidate vote cadiate
func (acc *Account) VoteCandidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.VoteCandidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.VoteCandidate, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		panic(err)
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekVoteProdoucer(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		// after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// KickedCandidate kicked candidates
func (acc *Account) KickedCandidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.KickedCandidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.KickedCandidate, acc.name, to, nonce, id, gas, value, payload, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekKickedCandidate(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		// after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// ExitTakeOver exit take over
func (acc *Account) ExitTakeOver(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	action := types.NewAction(types.ExitTakeOver, acc.name, to, nonce, id, gas, value, nil, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), 0, []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekKickedCandidate(action)
		if err != nil {
			return
		}
	}
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		// after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
		err = checkedfunc()
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// CreateContract create and send contract transaction
func (acc *Account) CreateContract(id uint64, gas uint64, input []byte) (hash common.Hash, err error) {
	acc.nonce, err = acc.api.AccountNonce(acc.name.String())
	if err != nil {
		return
	}

	action := types.NewAction(types.CreateContract, acc.name, acc.name, acc.nonce, id, gas, nil, input, nil)
	gasprice := big.NewInt(1)
	tx := types.NewTransaction(0, gasprice, action)
	signer := types.MakeSigner(big.NewInt(1))
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, signer, 0, []*types.KeyPair{key})
	if err != nil {
		return
	}

	rawtx, _ := rlp.EncodeToBytes(tx)

	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if acc.checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
	}
	return
}

// CallContract call contract transaction
func (acc *Account) CallContract(id uint64, gas uint64, input []byte) (hash common.Hash, err error) {
	acc.nonce, err = acc.api.AccountNonce(acc.name.String())
	if err != nil {
		return
	}

	action := types.NewAction(types.CallContract, acc.name, acc.name, acc.nonce, id, gas, nil, input, nil)
	gasprice := big.NewInt(1)
	tx := types.NewTransaction(0, gasprice, action)

	signer := types.MakeSigner(big.NewInt(1))
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, signer, 0, []*types.KeyPair{key})
	if err != nil {
		return
	}

	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return
	}

	checked := acc.checked || acc.nonce == math.MaxUint64
	hash, err = acc.api.SendRawTransaction(rawtx)
	if err != nil {
		return
	}
	if checked {
		//after
		err = acc.utilReceipt(hash, timeout)
		if err != nil {
			return
		}
	}

	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

func input(abifile string, method string, params ...interface{}) (string, error) {
	var abicode string
	hexcode, err := ioutil.ReadFile(abifile)
	if err != nil {
		fmt.Printf("Could not load code from file: %v\n", err)
		return "", err
	}
	abicode = string(bytes.TrimRight(hexcode, "\n"))
	parsed, err := abi.JSON(strings.NewReader(abicode))
	if err != nil {
		fmt.Println("abi.json error ", err)
		return "", err
	}
	input, err := parsed.Pack(method, params...)
	if err != nil {
		fmt.Println("parsed.pack error ", err)
		return "", err
	}
	return common.Bytes2Hex(input), nil
}

func formCreateContractInput(abifile string, binfile string) ([]byte, error) {
	hexcode, err := ioutil.ReadFile(binfile)
	if err != nil {
		return nil, err
	}
	code := common.Hex2Bytes(string(bytes.TrimRight(hexcode, "\n")))
	createInput, err := input(abifile, "")
	if err != nil {
		return nil, err
	}
	createCode := append(code, common.Hex2Bytes(createInput)...)
	return createCode, nil
}

func formIssueAssetInput(abifile string, desc string) ([]byte, error) {
	issueAssetInput, err := input(abifile, "issue", desc)
	if err != nil {
		return nil, err
	}
	return common.Hex2Bytes(issueAssetInput), nil
}

func formIncreaseAssetInput(abifile string, assetID *big.Int, to common.Address, value *big.Int) ([]byte, error) {
	increaseAssetInput, err := input(abifile, "increase", assetID, to, value)
	if err != nil {
		return nil, err
	}
	return common.Hex2Bytes(increaseAssetInput), nil
}

func formTransferAssetInput(abifile string, assetID *big.Int, toAddr common.Address, value *big.Int) ([]byte, error) {
	transferAssetInput, err := input(abifile, "transfer", assetID, toAddr, value)
	if err != nil {
		return nil, err
	}
	return common.Hex2Bytes(transferAssetInput), nil
}

func formChangeAssetOwner(abifile string, newOwner common.Address, assetID *big.Int) ([]byte, error) {
	changeOwnerInput, err := input(abifile, "changeowner", newOwner, assetID)
	if err != nil {
		return nil, err
	}
	return common.Hex2Bytes(changeOwnerInput), nil
}

func formDestroyAsset(abifile string, assetID, value *big.Int) ([]byte, error) {
	destroyAssetInput, err := input(abifile, "destroy", assetID, value)
	if err != nil {
		return nil, err
	}
	return common.Hex2Bytes(destroyAssetInput), nil
}
