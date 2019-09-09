package common

import (
	"crypto/ecdsa"
	"math"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	args "github.com/fractalplatform/fractal/consensus/dpos"
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

// Transfer 转账
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
		// before
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

// CreateAccount 新建账户
func (acc *Account) CreateAccount(accountName common.Name, founder common.Name, pubkey common.PubKey, detail string, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}
	account := &accountmanager.CreateAccountAction{
		AccountName: accountName,
		Founder:     founder,
		PublicKey:   pubkey,
		Description: detail,
	}
	payload, err := rlp.EncodeToBytes(account)
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

// UpdateAccount 更新账户
func (acc *Account) UpdateAccount(founder common.Name, to common.Name, value *big.Int, id uint64, gas uint64, priv *ecdsa.PrivateKey) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}
	account := &accountmanager.UpdateAccountAction{
		Founder: founder,
	}
	payload, err := rlp.EncodeToBytes(account)
	action := types.NewAction(types.UpdateAccount, acc.name, to, nonce, id, gas, value, payload, nil)
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

	acc.priv = priv
	if acc.nonce != math.MaxUint64 {
		acc.nonce++
	}
	return
}

// DeleteAccount 删除账户
func (acc *Account) DeleteAccount(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	action := types.NewAction(types.DeleteAccount, acc.name, to, nonce, id, gas, value, nil, nil)
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
		checkedfunc, err = acc.checkDeleteAccount(action)
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

// IssueAsset 发行资产
func (acc *Account) IssueAsset(assetname string, symbol string, amount *big.Int, decimals uint64, founder common.Name, owner common.Name, uplimit *big.Int, contract common.Name, detail string, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	asset := &accountmanager.IssueAsset{
		AssetName:   assetname,
		Symbol:      symbol,
		Amount:      amount,
		Decimals:    decimals,
		Founder:     founder,
		Owner:       owner,
		UpperLimit:  uplimit,
		Contract:    contract,
		Description: detail,
	}
	payload, err := rlp.EncodeToBytes(asset)
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

// IncreaseAsset 增发资产
func (acc *Account) IncreaseAsset(assetid uint64, toaccount common.Name, amount *big.Int, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	ast := &accountmanager.IncAsset{
		AssetID: assetid,
		To:      toaccount,
		Amount:  amount,
	}
	payload, _ := rlp.EncodeToBytes(ast)
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

// SetAssetOwner 修改资产owner
func (acc *Account) SetAssetOwner(assetid uint64, newowner common.Name, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	ast := &accountmanager.UpdateAssetOwner{
		AssetID: assetid,
		Owner:   newowner,
	}
	payload, _ := rlp.EncodeToBytes(ast)
	if err != nil {
		panic("rlp payload err")
	}
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

// SetAsset 修改资产founder
func (acc *Account) updateAsset(assetid uint64, founder common.Name, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	ast := &accountmanager.UpdateAsset{
		AssetID: assetid,
		Founder: founder,
	}
	payload, _ := rlp.EncodeToBytes(ast)
	if err != nil {
		panic("rlp payload err")
	}
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

//  修改资产contract
func (acc *Account) updateContractAsset(assetid uint64, contract common.Name, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	ast := &accountmanager.UpdateAssetContract{
		AssetID:  assetid,
		Contract: contract,
	}
	payload, _ := rlp.EncodeToBytes(ast)
	if err != nil {
		panic("rlp payload err")
	}
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

//销毁资产
func (acc *Account) DestroyAsset(to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

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
		// before
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

// RegCandidate 注册生成者
func (acc *Account) RegCandidate(url string, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	arg := &args.RegisterCandidate{
		URL: url,
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

// UpdateCandidate 更新生产者信息
func (acc *Account) UpdateCandidate(url string, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	arg := &args.UpdateCandidate{
		URL: url,
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

// UnRegCandidate 生产者注销
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

// RefundCandidate 生产者取回抵押金
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

// VoteCandidate 投票者投票
func (acc *Account) VoteCandidate(candidate string, stake *big.Int, to common.Name, value *big.Int, id uint64, gas uint64) (hash common.Hash, err error) {
	nonce := acc.nonce
	if nonce == math.MaxUint64 {
		nonce, err = acc.api.AccountNonce(acc.name.String())
		if err != nil {
			return
		}
	}

	arg := &args.VoteCandidate{
		Candidate: candidate,
		Stake:     stake,
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
