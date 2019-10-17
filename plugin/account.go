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
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	acctManagerName = "sysAccount"
	acctInfoPrefix  = "acctInfo"
)

const SystemAssetID uint64 = 0
const MaxDescriptionLength uint64 = 255

type CreateAccountAction struct {
	PublicKey   string `json:"publicKey,omitempty"`
	Description string `json:"description,omitempty"`
}

type IssueAsset struct {
	AssetName   string         `json:"assetName"`
	Symbol      string         `json:"symbol"`
	Amount      *big.Int       `json:"amount"`
	Decimals    uint64         `json:"decimals"`
	Founder     common.Address `json:"founder"`
	Owner       common.Address `json:"owner"`
	UpperLimit  *big.Int       `json:"upperLimit"`
	Description string         `json:"description"`
}

type AssetBalance struct {
	AssetID uint64   `json:"assetID"`
	Balance *big.Int `json:"balance"`
}

type Account struct {
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

// NewAM New a AccountManager
func NewAM(db *state.StateDB) (IAccount, error) {
	if db == nil {
		return nil, ErrNewAccountManagerErr
	}
	return &AccountManager{db}, nil
}

//func (am *AccountManager) Process(accountManagerContext *types.AccountManagerContext) ([]*types.InternalAction, error)  {
//	snap := am.sdb.Snapshot()
//	internalActions, err := am.process(accountManagerContext)
//	if err != nil {
//		am.sdb.RevertToSnapshot(snap)
//	}
//	return internalActions, err
//}
//
//func (am *AccountManager) process(accountManagerContext *types.AccountManagerContext) ([]*types.InternalAction, error) {
//	action := accountManagerContext.Action
//	number := accountManagerContext.Number
//	var fromAccountExtra []common.Name
//	fromAccountExtra = append(fromAccountExtra, accountManagerContext.FromAccountExtra...)
//
//	if err := action.Check(accountManagerContext.ChainConfig); err != nil {
//		return nil, err
//	}
//
//	var internalActions []*types.InternalAction
//
//	if err := am.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value(), fromAccountExtra...); err != nil {
//		return nil, err
//	}
//
//	return internalActions, nil
//}

// CreateAccount
// Prase Payload to create a account
func (am *AccountManager) CreateAccount(action *types.Action) ([]byte, error) {
	var acct CreateAccountAction
	err := rlp.DecodeBytes(action.Data(), &acct)
	if err != nil {
		return nil, err
	}

	// 验证公钥是否有效，如果格式不正确。那么转换的时候为空？
	if !common.IsHexPubKey(acct.PublicKey) {
		return nil, ErrInvalidPubKey
	}
	// 生成地址，判断地址是否已存在
	pubKey := common.HexToPubKey(acct.PublicKey)
	tempKey, err := crypto.UnmarshalPubkey(pubKey.Bytes())
	if err != nil {
		return nil, err
	}

	newAddress := crypto.PubkeyToAddress(*tempKey)

	// 判断账户是否已存在
	_, err = am.getAccount(newAddress)
	if err == nil {
		return nil, ErrAccountIsExist
	} else if err != ErrAccountNotExist {
		return nil, err
	}

	// new一个新对象，存储sdb
	acctObject := Account{
		Address:     newAddress,
		Nonce:       0,
		Code:        make([]byte, 0),
		CodeHash:    crypto.Keccak256Hash(nil),
		CodeSize:    0,
		Balances:    nil,
		Suicide:     false,
		Destroy:     false,
		Description: acct.Description,
	}

	am.setAccount(&acctObject)

	return newAddress.Bytes(), nil
}

// IssueAsset
// Pares Payload to issue a asset
func (am *AccountManager) IssueAsset(action *types.Action, asm IAsset) ([]byte, error) {
	var issueAsset IssueAsset
	err := rlp.DecodeBytes(action.Data(), &issueAsset)
	if err != nil {
		return nil, err
	}

	err = asm.CheckIssueAssetInfo(common.HexToAddress(action.Sender()), &issueAsset)
	if err != nil {
		return nil, err
	}

	//check owner
	if !common.IsHexAddress(issueAsset.Owner.String()) {
		return nil, ErrAccountAddressInvalid
	}
	_, err = am.getAccount(issueAsset.Owner)
	if err != nil {
		return nil, err
	}
	// check founder
	if len(issueAsset.Founder.String()) > 0 {
		_, err = am.getAccount(issueAsset.Owner)
		if err != nil {
			return nil, err
		}
	} else {
		issueAsset.Founder = issueAsset.Owner
	}

	assetID, err := asm.IssueAsset(issueAsset.AssetName, issueAsset.Symbol, issueAsset.Amount, issueAsset.Decimals, issueAsset.Founder, issueAsset.Owner, issueAsset.UpperLimit, issueAsset.Description)
	if err != nil {
		return nil, err
	}
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(assetID))

	return buf, nil
}

// TransferAsset
// Transaction designated asset to other account
func (am *AccountManager) TransferAsset(fromAccount, toAccount common.Address, assetID uint64, value *big.Int, asm IAsset, fromAccountExtra ...common.Address) error {
	if sign := value.Sign(); sign == 0 {
		return nil
	} else if sign == -1 {
		return ErrNegativeValue
	}

	fromAccountExtra = append(fromAccountExtra, fromAccount)
	fromAccountExtra = append(fromAccountExtra, toAccount)

	// check fromAccount
	fromAcct, err := am.getAccount(fromAccount)
	if err != nil {
		return err
	}
	if fromAcct == nil {
		return ErrAccountNotExist
	}

	// check fromAccount balance
	if assetID != SystemAssetID {
		return ErrAccountAssetNotExist
	}

	if fromAccount == toAccount || value.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	err = am.subBalanceByID(fromAcct, value)
	if err != nil {
		return err
	}

	toAcct, err := am.getAccount(toAccount)
	if err != nil {
		return err
	}
	if toAcct == nil {
		return ErrAccountNotExist
	}
	if toAcct.Destroy == true {
		return ErrAccountIsDestroy
	}

	isNew := am.addBalanceByID(toAcct, value)

	if isNew {
		err = asm.IncStats(assetID)
		if err != nil {
			return err
		}
	}

	if err = am.setAccount(fromAcct); err != nil {
		return err
	}

	return am.setAccount(toAcct)
}

func (am *AccountManager) GetNonce(arg interface{}) uint64 {
	return 0
}

func (am *AccountManager) getAccount(address common.Address) (*Account, error) {
	//if err := checkAddress(address); err != nil {
	//	return nil, err
	//}

	b, err := am.sdb.Get(acctManagerName, acctInfoPrefix+address.String())

	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		//log.Debug("account not exist", "address", ErrAccountNotExist, address)
		return nil, ErrAccountNotExist // 原先版本返回nil，改为直接返回error
	}

	var account Account
	if err = rlp.DecodeBytes(b, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

func (am *AccountManager) setAccount(account *Account) error {
	if account == nil {
		return ErrAccountIsNil
	}
	if account.Destroy == true {
		return ErrAccountIsDestroy
	}

	b, err := rlp.EncodeToBytes(account)
	if err != nil {
		return err
	}

	am.sdb.Put(acctManagerName, acctInfoPrefix+account.Address.String(), b)

	return nil
}

func (am *AccountManager) addBalanceByID(account *Account, value *big.Int) bool {
	if account.Balances == nil {
		account.Balances = &AssetBalance{
			AssetID: SystemAssetID,
			Balance: value,
		}
		return true
	}

	account.Balances.Balance = new(big.Int).Add(account.Balances.Balance, value)
	return false
}

func (am *AccountManager) subBalanceByID(account *Account, value *big.Int) error {
	if account.Balances == nil {
		return ErrAccountAssetNotExist
	}

	if account.Balances.Balance.Cmp(big.NewInt(0)) < 0 || account.Balances.Balance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	return nil
}

var (
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrNewAccountManagerErr   = errors.New("new accountManager err")
	ErrAssetIDInvalid         = errors.New("asset id invalid")
	ErrCreateAccountError     = errors.New("create account error")
	ErrAccountInvalid         = errors.New("account not permission")
	ErrAccountIsExist         = errors.New("account is exist")
	ErrAccountIsDestroy       = errors.New("account is destroy")
	ErrAccountNotExist        = errors.New("account not exist")
	ErrAccountAddressInvalid  = errors.New("account address is invalid")
	ErrHashIsEmpty            = errors.New("hash is empty")
	ErrkeyNotSame             = errors.New("key not same")
	ErrInvalidPubKey          = errors.New("invalid public key")
	ErrAccountIsNil           = errors.New("account object is empty")
	ErrCodeIsEmpty            = errors.New("code is empty")
	ErrAmountValueInvalid     = errors.New("amount value is invalid")
	ErrAccountAssetNotExist   = errors.New("account asset not exist")
	ErrUnKnownTxType          = errors.New("not support action type")
	ErrChargeRatioInvalid     = errors.New("charge ratio value invalid ")
	ErrAccountManagerNotExist = errors.New("account manager name not exist")
	ErrAmountMustZero         = errors.New("amount must be zero")
	ErrToNameInvalid          = errors.New("action to name(Recipient) invalid")
	ErrInvalidReceiptAsset    = errors.New("invalid receipt of asset")
	ErrInvalidReceipt         = errors.New("invalid receipt")
	ErrNegativeValue          = errors.New("negative value")
	ErrNegativeAmount         = errors.New("negative amount")
	ErrAssetOwnerInvalid      = errors.New("asset owner Invalid ")
)
