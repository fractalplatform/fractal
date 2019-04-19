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
	"crypto/ecdsa"
	"math"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	timeout = int64(time.Second) * 10
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
func (acc *Account) CreateAccount(to common.Name, value *big.Int, id uint64, gas uint64, newacct *accountmanager.AccountAction) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}
	payload, _ := rlp.EncodeToBytes(newacct)
	action := types.NewAction(types.CreateAccount, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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
func (acc *Account) UpdateAccount(to common.Name, value *big.Int, id uint64, gas uint64, newacct *accountmanager.AccountAction) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	bts, _ := rlp.EncodeToBytes(newacct)
	action := types.NewAction(types.UpdateAccount, acc.name, to, nonce, id, gas, value, bts)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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
	action := types.NewAction(types.Transfer, acc.name, to, nonce, id, gas, value, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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
func (acc *Account) IssueAsset(to common.Name, value *big.Int, id uint64, gas uint64, asset *asset.AssetObject) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(asset)
	action := types.NewAction(types.IssueAsset, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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
func (acc *Account) UpdateAsset(to common.Name, value *big.Int, id uint64, gas uint64, asset *asset.AssetObject) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(asset)
	action := types.NewAction(types.UpdateAsset, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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
	action := types.NewAction(types.IncreaseAsset, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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

	action := types.NewAction(types.DestroyAsset, acc.name, to, nonce, id, gas, value, nil)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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
func (acc *Account) SetAssetOwner(to common.Name, value *big.Int, id uint64, gas uint64, asset *asset.AssetObject) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(asset)
	action := types.NewAction(types.SetAssetOwner, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, acc.gasprice, []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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

// RegCadidate new cadidate
func (acc *Account) RegCadidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.RegisterCadidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.RegCadidate, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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

// UpdateCadidate update cadidate
func (acc *Account) UpdateCadidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.UpdateCadidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.UpdateCadidate, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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

// UnRegCadidate remove cadiate
func (acc *Account) UnRegCadidate(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	action := types.NewAction(types.UnregCadidate, acc.name, to, nonce, id, gas, value, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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

// VoteCadidate vote cadiate
func (acc *Account) VoteCadidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.VoteCadidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.VoteCadidate, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
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

// ChangeCadidate change cadidate
func (acc *Account) ChangeCadidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.ChangeCadidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.ChangeCadidate, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekChangeProdoucer(action)
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

// UnvoteCadidate unvote cadidate
func (acc *Account) UnvoteCadidate(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	action := types.NewAction(types.UnvoteCadidate, acc.name, to, nonce, id, gas, value, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekUnvoteProdoucer(action)
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

// UnvoteVoter remove voter
func (acc *Account) UnvoteVoter(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.RemoveVoter) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.RemoveVoter, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekRemoveVoter(action)
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

// KickedCadidate kicked cadidates
func (acc *Account) KickedCadidate(to common.Name, value *big.Int, id uint64, gas uint64, arg *dpos.KickedCadidate) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	payload, _ := rlp.EncodeToBytes(arg)
	action := types.NewAction(types.KickedCadidate, acc.name, to, nonce, id, gas, value, payload)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})

	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekKickedCadidate(action)
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

	action := types.NewAction(types.ExitTakeOver, acc.name, to, nonce, id, gas, value, nil)
	tx := types.NewTransaction(acc.feeid, big.NewInt(1e10), []*types.Action{action}...)
	key := types.MakeKeyPair(acc.priv, []uint64{0})
	err = types.SignActionWithMultiKey(action, tx, types.NewSigner(acc.chainID), []*types.KeyPair{key})
	if err != nil {
		return
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	checked := acc.checked || acc.nonce == math.MaxUint64
	var checkedfunc func() error
	if checked {
		// before
		checkedfunc, err = acc.chekKickedCadidate(action)
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
