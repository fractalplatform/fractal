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
	"errors"
	"math/big"
	"regexp"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/common/hexutil"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	acctRegExp           = regexp.MustCompile(`^([a-z][a-z0-9]{6,19})$`)
	acctManagerName      = "sysAccount"
	acctInfoPrefix       = "acctInfo"
	accountNameMaxLength = uint64(20)
)

const MaxDescriptionLength uint64 = 255

type AssetBalance struct {
	AssetID uint64   `json:"assetID"`
	Balance *big.Int `json:"balance"`
}

type Account struct {
	Name        string
	Address     common.Address
	Nonce       uint64
	Code        []byte
	CodeHash    common.Hash
	CodeSize    uint64
	Balances    *AssetBalance
	Suicide     bool
	Destroy     bool
	Description string
}

type AccountManager struct {
	sdb *state.StateDB
}

// NewACM New a AccountManager
func NewACM(db *state.StateDB) (*AccountManager, error) {
	if db == nil {
		return nil, ErrNewAccountManagerErr
	}
	return &AccountManager{db}, nil
}

func (am *AccountManager) CallTx(tx *envelope.PluginTx, ctx *Context, pm IPM) ([]byte, error) {
	switch tx.PayloadType() {
	case CreateAccount:
		param := &CreateAccountAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return am.CreateAccount(param.Name, param.Pubkey, param.Desc)
	case ChangePubKey:
		param := &ChangePubKeyAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		err := am.ChangePubKey(tx.Sender(), param.Pubkey)
		return nil, err
	case Transfer:
		err := am.TransferAsset(tx.Sender(), tx.Recipient(), tx.GetAssetID(), tx.Value())
		return nil, err
	}
	return nil, ErrWrongTransaction
}

// CreateAccount Parse Payload to create a account
func (am *AccountManager) CreateAccount(accountName string, pubKey string, description string) ([]byte, error) {
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}

	if !common.IsHexPubKey(pubKey) {
		return nil, ErrPubKey
	}

	if err := am.checkAccountName(accountName); err != nil {
		return nil, err
	}

	_, err := am.getAccount(accountName)
	if err == nil {
		return nil, ErrAccountIsExist
	} else if err != ErrAccountNotExist {
		return nil, err
	}

	tempKey := common.HexToPubKey(pubKey)
	newAddress := common.BytesToAddress(crypto.Keccak256(tempKey.Bytes()[1:])[12:])
	balance := &AssetBalance{0, big.NewInt(0)}

	acctObject := Account{
		Name:        accountName,
		Address:     newAddress,
		Nonce:       0,
		Code:        make([]byte, 0),
		CodeHash:    crypto.Keccak256Hash(nil),
		CodeSize:    0,
		Balances:    balance,
		Suicide:     false,
		Destroy:     false,
		Description: description,
	}

	return nil, am.setAccount(&acctObject)
}

// CanTransfer check if can transfer.
func (am *AccountManager) CanTransfer(accountName string, assetID uint64, value *big.Int) error {

	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	val, err := am.GetBalance(accountName, assetID)
	if err != nil {
		return err
	}

	if val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	return nil
}

// Transaction designated asset to other account
func (am *AccountManager) TransferAsset(fromAccount, toAccount string, assetID uint64, value *big.Int) error {
	if value.Cmp(big.NewInt(0)) == 0 {
		return nil
	} else if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	if fromAccount == toAccount {
		return nil
	}

	// check fromAccount
	fromAcct, err := am.getAccount(fromAccount)
	if err != nil {
		return err
	}

	toAcct, err := am.getAccount(toAccount)
	if err != nil {
		return err
	}

	if err = am.subBalance(fromAcct, assetID, value); err != nil {
		return err
	}

	if err = am.addBalance(toAcct, assetID, value); err != nil {
		return err
	}

	if err = am.setAccount(fromAcct); err != nil {
		return err
	}

	if err = am.setAccount(toAcct); err != nil {
		return err
	}

	return nil
}

// RecoverTx Make sure the transaction is signed properly and validate account authorization.
func (am *AccountManager) RecoverTx(signer ISigner, tx *types.Transaction) error {
	_, err := am.AccountVerify(tx.Sender(), signer, tx.GetSign(), tx.SignHash)
	return err
}

func (am *AccountManager) AccountSign(accountName string, priv *ecdsa.PrivateKey, signer ISigner, signHash func(chainID *big.Int) common.Hash) ([]byte, error) {
	signerAccount, err := am.getAccount(accountName)
	if err != nil {
		return nil, err
	}
	keyAddress := crypto.PubkeyToAddress(priv.PublicKey)
	if signerAccount.Address.Compare(keyAddress) != 0 {
		return nil, ErrkeyNotSame
	}
	//block.Head.Proof = crypto.VRF_Proof(priKey, c.parent.Hash().Bytes())
	return signer.Sign(signHash, priv)
}

func (am *AccountManager) AccountVerify(accountName string, signer ISigner, signature []byte, signHash func(chainID *big.Int) common.Hash) (*ecdsa.PublicKey, error) {
	account, err := am.getAccount(accountName)
	if err != nil {
		return nil, err
	}

	pub, err := signer.Recover(signature, signHash)
	if err != nil {
		return nil, err
	}
	tempAddress := common.BytesToAddress(crypto.Keccak256(pub[1:])[12:])

	if tempAddress.Compare(account.Address) != 0 {
		return nil, ErrkeyNotSame
	}
	return crypto.UnmarshalPubkey(pub)
}

// GetNonce
func (am *AccountManager) GetNonce(accountName string) (uint64, error) {
	account, err := am.getAccount(accountName)
	if err != nil {
		return 0, err
	}

	return account.Nonce, nil
}

// SetNonce
func (am *AccountManager) SetNonce(accountName string, nonce uint64) error {
	account, err := am.getAccount(accountName)
	if err != nil {
		return err
	}

	account.Nonce = nonce

	err = am.setAccount(account)
	if err != nil {
		return err
	}
	return nil
}

// GetCode
func (am *AccountManager) GetCode(accountName string) ([]byte, error) {
	account, err := am.getAccount(accountName)
	if err != nil {
		return nil, err
	}

	if account.Suicide {
		return nil, ErrCodeIsEmpty
	}

	return account.Code, nil
}

// GetCodeHash
func (am *AccountManager) GetCodeHash(accountName string) (common.Hash, error) {
	account, err := am.getAccount(accountName)
	if err != nil {
		return common.Hash{}, err
	}

	if account.CodeSize == 0 {
		return common.Hash{}, ErrHashIsEmpty
	}

	return account.CodeHash, nil
}

// SetCode
func (am *AccountManager) SetCode(accountName string, code []byte) error {
	account, err := am.getAccount(accountName)
	if err != nil {
		return err
	}

	if len(code) == 0 {
		return ErrCodeIsEmpty
	}
	account.Code = code
	account.CodeHash = crypto.Keccak256Hash(code)
	account.CodeSize = uint64(len(code))

	err = am.setAccount(account)
	if err != nil {
		return err
	}
	return nil
}

// GetBalance get account asset balance
func (am *AccountManager) GetBalance(accountName string, assetID uint64) (*big.Int, error) {

	account, err := am.getAccount(accountName)
	if err != nil {
		return big.NewInt(0), err
	}

	if account.Balances.AssetID != assetID {
		return big.NewInt(0), ErrAssetIDInvalid
	}

	return account.Balances.Balance, nil
}

func (am *AccountManager) AccountIsExist(accountName string) error {
	_, err := am.getAccount(accountName)
	if err != nil {
		return err
	}
	return nil
}

func (am *AccountManager) GetAccountByName(accountName string) (interface{}, error) {
	account, err := am.getAccount(accountName)
	if err != nil {
		return nil, err
	}

	obj := struct {
		Name        string         `json:"name"`
		Address     common.Address `json:"address"`
		Nonce       uint64         `json:"nonce"`
		Code        hexutil.Bytes  `json:"code"`
		CodeHash    common.Hash    `json:"codeHash"`
		CodeSize    uint64         `json:"codeSize"`
		Balances    *AssetBalance  `json:"balance"`
		Suicide     bool           `json:"suicide"`
		Destroy     bool           `json:"destroy"`
		Description string         `json:"description"`
	}{
		account.Name,
		account.Address,
		account.Nonce,
		(hexutil.Bytes)(account.Code),
		account.CodeHash,
		account.CodeSize,
		account.Balances,
		account.Suicide,
		account.Destroy,
		account.Description,
	}

	return obj, nil
}

func (am *AccountManager) ChangePubKey(accountName string, pubKey string) error {
	if !common.IsHexPubKey(pubKey) {
		return ErrPubKey
	}

	account, err := am.getAccount(accountName)
	if err != nil {
		return err
	}
	tempKey := common.HexToPubKey(pubKey)
	account.Address = common.BytesToAddress(crypto.Keccak256(tempKey.Bytes()[1:])[12:])
	if err = am.setAccount(account); err != nil {
		return err
	}
	return nil
}

func (am *AccountManager) addBalanceByID(accountName string, assetID uint64, amount *big.Int) error {
	account, err := am.getAccount(accountName)
	if err != nil {
		return err
	}

	if err = am.addBalance(account, assetID, amount); err != nil {
		return err
	}

	return am.setAccount(account)
}

func (am *AccountManager) subBalanceByID(accountName string, assetID uint64, amount *big.Int) error {
	account, err := am.getAccount(accountName)
	if err != nil {
		return err
	}

	if err = am.subBalance(account, assetID, amount); err != nil {
		return err
	}

	return am.setAccount(account)
}

//func (am *AccountManager) DeleteAccount(accountAddress common.Address) error {
//	account, err := am.getAccount(accountAddress)
//	if err != nil {
//		return err
//	}
//
//	account.Destroy = true
//
//	if err = am.setAccount(account); err != nil {
//		return err
//	}
//	return nil
//}

func (am *AccountManager) checkAccountName(accountName string) error {
	if uint64(len(accountName)) > accountNameMaxLength {
		return ErrAccountNameLengthErr
	}

	if acctRegExp.MatchString(accountName) != true {
		log.Info("check account name ", "name", accountName)
		return ErrAccountNameInvalid
	}
	return nil
}

func (am *AccountManager) getAccount(accountName string) (*Account, error) {
	b, err := am.sdb.Get(acctManagerName, acctInfoPrefix+accountName)
	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		return nil, ErrAccountNotExist
	}

	var account Account
	if err = rlp.DecodeBytes(b, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

func (am *AccountManager) setAccount(account *Account) error {
	if account == nil {
		return ErrAccountObjectIsNil
	}

	b, err := rlp.EncodeToBytes(account)
	if err != nil {
		return err
	}

	am.sdb.Put(acctManagerName, acctInfoPrefix+account.Name, b)

	return nil
}

func (am *AccountManager) addBalance(account *Account, assetID uint64, value *big.Int) error {
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	if account.Balances.AssetID != assetID {
		return ErrAssetIDInvalid
	}
	account.Balances.Balance = new(big.Int).Add(account.Balances.Balance, value)

	return nil
}

func (am *AccountManager) subBalance(account *Account, assetID uint64, value *big.Int) error {
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	if account.Balances.AssetID != assetID {
		return ErrAssetIDInvalid
	}

	if account.Balances.Balance.Cmp(big.NewInt(0)) < 0 || account.Balances.Balance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	account.Balances.Balance = new(big.Int).Sub(account.Balances.Balance, value)
	return nil
}

func (am *AccountManager) checkCreateAccount(accountName string, pubKey string, description string) error {
	if uint64(len(description)) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}

	if err := am.checkAccountName(accountName); err != nil {
		return err
	}

	if !common.IsHexPubKey(pubKey) {
		return ErrPubKey
	}
	return nil
}

func (am *AccountManager) Sol_CreateAccount(context *ContextSol, name string, pubKey string, desc string) error {
	_, err := am.CreateAccount(name, pubKey, desc)
	return err
}

func (am *AccountManager) Sol_ChangePubKey(context *ContextSol, pubKey string) error {
	return am.ChangePubKey(context.tx.Sender(), pubKey)
}

func (am *AccountManager) Sol_GetBalance(context *ContextSol, account string, assetID uint64) (*big.Int, error) {
	return am.GetBalance(account, assetID)
}

func (am *AccountManager) Sol_Transfer(context *ContextSol, to string, assetID uint64, value *big.Int) error {
	return am.TransferAsset(context.tx.Sender(), to, assetID, value)
}

func (am *AccountManager) Sol_AddressToString(context *ContextSol, name common.Address) (string, error) {
	return name.AccountName(), nil
}

func (am *AccountManager) Sol_StringToAddress(Context *ContextSol, name string) (common.Address, error) {
	return common.StringToAddress(name)
}

var (
	ErrAccountNameLengthErr = errors.New("account name length err")
	ErrAccountNameInvalid   = errors.New("account name invalid")
	ErrNewAccountManagerErr = errors.New("new account manager err")
	ErrAccountNotExist      = errors.New("account not exist")
	ErrAccountIsExist       = errors.New("account is exist")
	ErrAccountObjectIsNil   = errors.New("account object is nil")
	ErrAssetIDInvalid       = errors.New("assetID invalid")
	ErrAmountValueInvalid   = errors.New("amount value invalid")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrCodeIsEmpty          = errors.New("code is empty")
	ErrHashIsEmpty          = errors.New("hash is empty")
	ErrkeyNotSame           = errors.New("key not same")
	ErrPubKey               = errors.New("pubkey invalid")
)
