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
	"errors"
	"math/big"
	"regexp"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	acctRegExp           = regexp.MustCompile(`^([a-z][a-z0-9]{6,31})`)
	acctManagerName      = "sysAccount"
	acctInfoPrefix       = "acctInfo"
	accountNameMaxLength = uint64(32)
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
func NewACM(db *state.StateDB) (IAccount, error) {
	if db == nil {
		return nil, ErrNewAccountManagerErr
	}
	return &AccountManager{db}, nil
}

// CreateAccount Parse Payload to create a account
func (am *AccountManager) CreateAccount(accountName string, pubKey common.PubKey, description string) ([]byte, error) {
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
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

	newAddress := common.BytesToAddress(crypto.Keccak256(pubKey.Bytes()[1:])[12:])
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

	if err = am.setAccount(&acctObject); err != nil {
		return nil, err
	}
	return nil, nil
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

	snap := am.sdb.Snapshot()

	if err = am.setAccount(fromAcct); err != nil {
		return err
	}

	if err = am.setAccount(toAcct); err != nil {
		am.sdb.RevertToSnapshot(snap)
		return err
	}

	return nil
}

// RecoverTx Make sure the transaction is signed properly and validate account authorization.
func (am *AccountManager) RecoverTx(signer ISigner, tx *types.Transaction) error {
	for _, action := range tx.GetActions() {
		pubs, err := signer.Recover(action)
		if err != nil {
			return err
		}

		tempAddress := common.BytesToAddress(crypto.Keccak256(pubs[1:])[12:])

		account, err := am.getAccount(action.Sender())
		if err != nil {
			return err
		}

		if tempAddress.Compare(account.Address) != 0 {
			return ErrkeyNotSame
		}
	}

	return nil
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

func (am *AccountManager) GetAccountByName(accountName string) (*Account, error) {
	account, err := am.getAccount(accountName)
	if err != nil {
		return nil, err
	}
	return account, nil
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
		return ErrAccountNameinvalid
	}
	return nil
}

// func (am *AccountManager) GetAccount(accountName string) (*Account, error) {
// 	return am.getAccount(accountName)
// }

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

var (
	ErrAccountNameLengthErr = errors.New("account name length err")
	ErrAccountNameinvalid   = errors.New("account name invalid")
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
)
