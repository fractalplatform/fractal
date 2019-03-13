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

package accountmanager

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var acctInfoPrefix = "AcctInfo"

var acctManagerName = "sysAccount"
var sysName string = ""

type AccountAction struct {
	AccountName common.Name   `json:"accountName,omitempty"`
	Founder     common.Name   `json:"founder,omitempty"`
	ChargeRatio uint64        `json:"chargeRatio,omitempty"`
	PublicKey   common.PubKey `json:"publicKey,omitempty"`
}

type IncAsset struct {
	AssetId uint64      `json:"assetId,omitempty"`
	Amount  *big.Int    `json:"amount,omitempty"`
	To      common.Name `json:"account,omitempty"`
}

//AccountManager represents account management model.
type AccountManager struct {
	sdb SdbIf
	ast *asset.Asset
}

//SetSysName set the global sys name
func SetSysName(name common.Name) bool {
	if common.IsValidName(name.String()) {
		sysName = name.String()
		return true
	}
	return false
}

//SetAcctMangerName  set the global account manager name
func SetAcctMangerName(name common.Name) bool {
	if common.IsValidName(name.String()) {
		acctManagerName = name.String()
		return true
	}
	return false
}

//NewAccountManager create new account manager
func NewAccountManager(db *state.StateDB) (*AccountManager, error) {
	if db == nil {
		return nil, ErrNewAccountErr
	}
	if len(acctManagerName) == 0 {
		log.Error("NewAccountManager error", "name", ErrAccountManagerNotExist, acctManagerName)
		return nil, ErrAccountManagerNotExist
	}

	return &AccountManager{
		sdb: db,
		ast: asset.NewAsset(db),
	}, nil
}

// AccountIsExist check account is exist.
func (am *AccountManager) AccountIsExist(accountName common.Name) (bool, error) {
	//check is exist
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if acct != nil {
		return true, nil
	}
	return false, nil
}

//AccountHaveCode check account have code
func (am *AccountManager) AccountHaveCode(accountName common.Name) (bool, error) {
	//check is exist
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if acct == nil {
		return false, ErrAccountNotExist
	}

	return acct.HaveCode(), nil
}

//AccountIsEmpty check account is empty
func (am *AccountManager) AccountIsEmpty(accountName common.Name) (bool, error) {
	//check is exist
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if acct == nil {
		return false, ErrAccountNotExist
	}

	if acct.IsEmpty() {
		return true, nil
	}
	return false, nil
}

//CreateAccount contract account
func (am *AccountManager) CreateAccount(accountName common.Name, founderName common.Name, chargeRatio uint64, pubkey common.PubKey) error {
	//check is exist
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct != nil {
		return ErrAccountIsExist
	}
	var fname common.Name
	if len(founderName.String()) > 0 && founderName != accountName {
		f, err := am.GetAccountByName(founderName)
		if err != nil {
			return err
		}
		if f == nil {
			return ErrAccountNotExist
		}
		fname.SetString(founderName.String())
	} else {
		fname.SetString(accountName.String())
	}

	acctObj, err := NewAccount(accountName, fname, pubkey)
	if err != nil {
		return err
	}
	if acctObj == nil {
		return ErrCreateAccountError
	}

	acctObj.SetChargeRatio(0)
	am.SetAccount(acctObj)
	return nil
}

//SetChargeRatio set the Charge Ratio of the accunt
func (am *AccountManager) SetChargeRatio(accountName common.Name, ra uint64) error {
	acct, err := am.GetAccountByName(accountName)
	if acct == nil {
		return ErrAccountNotExist
	}
	if err != nil {
		return err
	}
	acct.SetChargeRatio(ra)
	return am.SetAccount(acct)
}

//UpdateAccount update the pubkey of the accunt
func (am *AccountManager) UpdateAccount(accountName common.Name, founderName common.Name, chargeRatio uint64, pubkey common.PubKey) error {
	acct, err := am.GetAccountByName(accountName)
	if acct == nil {
		return ErrAccountNotExist
	}
	if err != nil {
		return err
	}
	if len(founderName.String()) > 0 {
		f, err := am.GetAccountByName(founderName)
		if err != nil {
			return err
		}
		if f == nil {
			return ErrAccountNotExist
		}
	}
	if chargeRatio > 100 {
		return ErrChargeRatioInvalid
	}
	acct.SetFounder(founderName)
	acct.SetChargeRatio(chargeRatio)
	acct.SetPubKey(pubkey)
	return am.SetAccount(acct)
}

//GetAccountByTime get account by name and time
func (am *AccountManager) GetAccountByTime(accountName common.Name, time uint64) (*Account, error) {
	b, err := am.sdb.GetSnapshot(acctManagerName, acctInfoPrefix+accountName.String(), time)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}

	var acct Account
	if err := rlp.DecodeBytes(b, &acct); err != nil {
		return nil, err
	}

	return &acct, nil
}

//GetAccountByName get account by name
func (am *AccountManager) GetAccountByName(accountName common.Name) (*Account, error) {
	b, err := am.sdb.Get(acctManagerName, acctInfoPrefix+accountName.String())
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		log.Info("account not exist", "account", ErrAccountNotExist, accountName)
		return nil, nil
	}

	var acct Account
	if err := rlp.DecodeBytes(b, &acct); err != nil {
		return nil, err
	}

	return &acct, nil
}

//SetAccount store account object to db
func (am *AccountManager) SetAccount(acct *Account) error {
	if acct == nil {
		return ErrAccountIsNil
	}
	if acct.IsDestroyed() == true {
		return ErrAccountIsDestroy
	}
	b, err := rlp.EncodeToBytes(acct)
	if err != nil {
		return err
	}
	am.sdb.Put(acctManagerName, acctInfoPrefix+acct.GetName().String(), b)
	return nil
}

//DeleteAccountByName delete account
func (am *AccountManager) DeleteAccountByName(accountName common.Name) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return ErrAccountNotExist
	}
	if acct == nil {
		return ErrAccountNotExist
	}

	acct.SetDestroy()
	b, err := rlp.EncodeToBytes(acct)
	if err != nil {
		return err
	}
	am.sdb.Put(acct.GetName().String(), acctInfoPrefix, b)
	return nil
}

// GetNonce get nonce
func (am *AccountManager) GetNonce(accountName common.Name) (uint64, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return 0, err
	}
	if acct == nil {
		return 0, ErrAccountNotExist
	}
	return acct.GetNonce(), nil
}

// SetNonce set nonce
func (am *AccountManager) SetNonce(accountName common.Name, nonce uint64) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	acct.SetNonce(nonce)
	return am.SetAccount(acct)
}

//GetBalancesList get Balances return a list
//func (am *AccountManager) GetBalancesList(accountName common.Name) ([]*AssetBalance, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return nil, err
//	}
//	return acct.GetBalancesList(), nil
//}

//GetAllAccountBalance return all balance in map.
//func (am *AccountManager) GetAccountAllBalance(accountName common.Name) (map[uint64]*big.Int, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return nil, err
//	}
//	if acct == nil {
//		return nil, ErrAccountNotExist
//	}
//
//	return acct.GetAllBalances()
//}

//GetAcccountPubkey get account pub key
//func (am *AccountManager) GetAcccountPubkey(accountName common.Name) ([]byte, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return nil, err
//	}
//	if acct == nil {
//		return nil, ErrAccountNotExist
//	}
//	return acct.GetPubKey().Bytes(), nil
//}

// RecoverTx Make sure the transaction is signed properly and validate account authorization.
func (am *AccountManager) RecoverTx(signer types.Signer, tx *types.Transaction) error {
	for _, action := range tx.GetActions() {
		pub, err := types.Recover(signer, action, tx)
		if err != nil {
			return err
		}

		if err := am.IsValidSign(action.Sender(), action.Type(), pub); err != nil {
			return err
		}
	}
	return nil
}

//IsValidSign check the sign
func (am *AccountManager) IsValidSign(accountName common.Name, aType types.ActionType, pub common.PubKey) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	if acct.IsDestroyed() {
		return ErrAccountIsDestroy
	}
	//TODO action type verify

	if acct.GetPubKey().Compare(pub) != 0 {
		return fmt.Errorf("%v %v have %v excepted %v", acct.AcctName, ErrkeyNotSame, acct.GetPubKey().String(), pub.String())
	}
	return nil

}

//GetAssetInfoByName get asset info by asset name.
func (am *AccountManager) GetAssetInfoByName(assetName string) (*asset.AssetObject, error) {
	assetID, err := am.ast.GetAssetIdByName(assetName)
	if err != nil {
		return nil, err
	}
	return am.ast.GetAssetObjectById(assetID)
}

//GetAssetInfoByID get asset info by assetID
func (am *AccountManager) GetAssetInfoByID(assetID uint64) (*asset.AssetObject, error) {
	return am.ast.GetAssetObjectById(assetID)
}

//GetAccountBalanceByID get account balance by ID
func (am *AccountManager) GetAccountBalanceByID(accountName common.Name, assetID uint64) (*big.Int, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return big.NewInt(0), err
	}
	if acct == nil {
		return big.NewInt(0), ErrAccountNotExist
	}
	return acct.GetBalanceByID(assetID)
}

//GetAssetAmountByTime get asset amount by time
func (am *AccountManager) GetAssetAmountByTime(assetID uint64, time uint64) (*big.Int, error) {
	return am.ast.GetAssetAmountByTime(assetID, time)
}

//GetAccountLastChange account balance last change time
func (am *AccountManager) GetAccountLastChange(accountName common.Name) (uint64, error) {
	//TODO
	return 0, nil
}

//GetSnapshotTime get snapshot time
//num = 0  current snapshot time , 1 preview snapshot time , 2 next snapshot time
func (am *AccountManager) GetSnapshotTime(num uint64, time uint64) (uint64, error) {
	if num == 0 {
		return am.sdb.GetSnapshotLast()
	} else if num == 1 {
		return am.sdb.GetSnapshotPrev(time)
	} else if num == 2 {
		t, err := am.sdb.GetSnapshotLast()
		if err != nil {
			return 0, err
		}

		if t <= time {
			return 0, ErrSnapshotTimeNotExist
		} else {
			for {
				if t1, err := am.sdb.GetSnapshotPrev(t); err != nil {
					return t, nil
				} else if t1 <= time {
					return t, nil
				} else {
					t = t1
				}
			}
		}
	}
	return 0, ErrTimeTypeInvalid
}

//GetBalanceByTime get account balance by Time
func (am *AccountManager) GetBalanceByTime(accountName common.Name, assetID uint64, time uint64) (*big.Int, error) {
	acct, err := am.GetAccountByTime(accountName, time)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, ErrAccountNotExist
	}
	return acct.GetBalanceByID(assetID)
}

//GetFounder Get Account Founder
func (am *AccountManager) GetFounder(accountName common.Name) (common.Name, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return "", err
	}
	if acct == nil {
		return "", ErrAccountNotExist
	}
	return acct.GetFounder(), nil
}

//GetAssetFounder Get Asset Founder
func (am *AccountManager) GetAssetFounder(assetID uint64) (common.Name, error) {
	return am.ast.GetAssetFounderById(assetID)
}

//GetChargeRatio Get Account ChargeRatio
func (am *AccountManager) GetChargeRatio(accountName common.Name) (uint64, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return 0, err
	}
	if acct == nil {
		return 0, ErrAccountNotExist
	}
	return acct.GetChargeRatio(), nil
}

//GetAssetChargeRatio Get Asset ChargeRatio
func (am *AccountManager) GetAssetChargeRatio(assetID uint64) (uint64, error) {
	acctName, err := am.ast.GetAssetFounderById(assetID)
	if err != nil {
		return 0, err
	}
	if acctName == "" {
		return 0, ErrAccountNotExist
	}
	return am.GetChargeRatio(acctName)
}

//GetAccountBalanceByName get account balance by name
//func (am *AccountManager) GetAccountBalanceByName(accountName common.Name, assetName string) (*big.Int, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return big.NewInt(0), err
//	}
//	if acct == nil {
//		return big.NewInt(0), ErrAccountNotExist
//	}
//
//	assetID, err := am.ast.GetAssetIdByName(assetName)
//	if err != nil {
//		return big.NewInt(0), err
//	}
//	if assetID == 0 {
//		return big.NewInt(0), asset.ErrAssetNotExist
//	}
//
//	ba := &big.Int{}
//	ba, err = acct.GetBalanceByID(assetID)
//	if err != nil {
//		return big.NewInt(0), err
//	}
//
//	return ba, nil
//}

//SubAccountBalanceByID sub balance by assetID
func (am *AccountManager) SubAccountBalanceByID(accountName common.Name, assetID uint64, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}

	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	//
	val, err := acct.GetBalanceByID(assetID)
	if err != nil {
		return err
	}
	if val.Cmp(big.NewInt(0)) < 0 || val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	acct.SetBalance(assetID, new(big.Int).Sub(val, value))
	return am.SetAccount(acct)
}

//AddAccountBalanceByID add balance by assetID
func (am *AccountManager) AddAccountBalanceByID(accountName common.Name, assetID uint64, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}

	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	val, err := acct.GetBalanceByID(assetID)
	if err == ErrAccountAssetNotExist {
		acct.AddNewAssetByAssetID(assetID, value)
	} else {
		acct.SetBalance(assetID, new(big.Int).Add(val, value))
	}
	return am.SetAccount(acct)
}

//AddAccountBalanceByName  add balance by name
func (am *AccountManager) AddAccountBalanceByName(accountName common.Name, assetName string, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	assetID, err := am.ast.GetAssetIdByName(assetName)
	if err != nil {
		return err
	}

	if assetID == 0 {
		return asset.ErrAssetNotExist
	}
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	val, err := acct.GetBalanceByID(assetID)
	if err == ErrAccountAssetNotExist {
		acct.AddNewAssetByAssetID(assetID, value)
	} else {
		acct.SetBalance(assetID, new(big.Int).Add(val, value))
	}
	return am.SetAccount(acct)
}

//
func (am *AccountManager) EnoughAccountBalance(accountName common.Name, assetID uint64, value *big.Int) error {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	return acct.EnoughAccountBalance(assetID, value)
}

//
func (am *AccountManager) GetCode(accountName common.Name) ([]byte, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, ErrAccountNotExist
	}
	return acct.GetCode()
}

////
//func (am *AccountManager) SetCode(accountName common.Name, code []byte) (bool, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return false, err
//	}
//	if acct == nil {
//		return false, ErrAccountNotExist
//	}
//	err = acct.SetCode(code)
//	if err != nil {
//		return false, err
//	}
//	err = am.SetAccount(acct)
//	if err != nil {
//		return false, err
//	}
//	return true, nil
//}

//
//GetCodeSize get code size
func (am *AccountManager) GetCodeSize(accountName common.Name) (uint64, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return 0, err
	}
	if acct == nil {
		return 0, ErrAccountNotExist
	}
	return acct.GetCodeSize(), nil
}

// GetCodeHash get code hash
//func (am *AccountManager) GetCodeHash(accountName common.Name) (common.Hash, error) {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	if acct == nil {
//		return common.Hash{}, ErrAccountNotExist
//	}
//	return acct.GetCodeHash()
//}

//GetAccountFromValue  get account info via value bytes
func (am *AccountManager) GetAccountFromValue(accountName common.Name, key string, value []byte) (*Account, error) {
	if len(value) == 0 {
		return nil, ErrAccountNotExist
	}
	if key != accountName.String()+acctInfoPrefix {
		return nil, ErrAccountNameInvalid
	}
	var acct Account
	if err := rlp.DecodeBytes(value, &acct); err != nil {
		return nil, ErrAccountNotExist
	}
	if acct.AcctName != accountName {
		return nil, ErrAccountNameInvalid
	}
	return &acct, nil
}

// CanTransfer check if can transfer.
func (am *AccountManager) CanTransfer(accountName common.Name, assetID uint64, value *big.Int) (bool, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if err = acct.EnoughAccountBalance(assetID, value); err == nil {
		return true, nil
	}
	return false, err
}

//TransferAsset transfer asset
func (am *AccountManager) TransferAsset(fromAccount common.Name, toAccount common.Name, assetID uint64, value *big.Int) error {
	//check from account
	fromAcct, err := am.GetAccountByName(fromAccount)
	if err != nil {
		return err
	}
	if fromAcct == nil {
		return ErrAccountNotExist
	}
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	if fromAccount == toAccount || value.Cmp(big.NewInt(0)) == 0 {
		return nil
	}
	//check from account balance
	val, err := fromAcct.GetBalanceByID(assetID)
	if err != nil {
		return err
	}
	if val.Cmp(big.NewInt(0)) < 0 || val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	fromAcct.SetBalance(assetID, new(big.Int).Sub(val, value))
	//check to account
	toAcct, err := am.GetAccountByName(toAccount)
	if err != nil {
		return err
	}
	if toAcct == nil {
		return ErrAccountNotExist
	}
	if toAcct.IsDestroyed() {
		return ErrAccountIsDestroy
	}
	val, err = toAcct.GetBalanceByID(assetID)
	if err == ErrAccountAssetNotExist {
		toAcct.AddNewAssetByAssetID(assetID, value)
	} else {
		toAcct.SetBalance(assetID, new(big.Int).Add(val, value))
	}
	if err = am.SetAccount(fromAcct); err != nil {
		return err
	}
	return am.SetAccount(toAcct)
}

//IssueAsset issue asset
func (am *AccountManager) IssueAsset(asset *asset.AssetObject) error {
	//check owner
	acct, err := am.GetAccountByName(asset.GetAssetOwner())
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotExist
	}
	//check founder
	if len(asset.GetAssetFounder()) > 0 {
		f, err := am.GetAccountByName(asset.GetAssetFounder())
		if err != nil {
			return err
		}
		if f == nil {
			return ErrAccountNotExist
		}
	} else {
		asset.SetAssetFounder(asset.GetAssetOwner())
	}

	if err := am.ast.IssueAsset(asset.GetAssetName(), asset.GetSymbol(), asset.GetAssetAmount(), asset.GetDecimals(), asset.GetAssetFounder(), asset.GetAssetOwner(), asset.GetUpperLimit()); err != nil {
		return err
	}
	//add the asset to owner
	return am.AddAccountBalanceByName(asset.GetAssetOwner(), asset.GetAssetName(), asset.GetAssetAmount())
}

//IncAsset2Acct increase asset and add amount to accout balance
func (am *AccountManager) IncAsset2Acct(fromName common.Name, toName common.Name, assetID uint64, amount *big.Int) error {
	if err := am.ast.IncreaseAsset(fromName, assetID, amount); err != nil {
		return err
	}
	return am.AddAccountBalanceByID(toName, assetID, amount)
}

//AddBalanceByName add balance to account
//func (am *AccountManager) AddBalanceByName(accountName common.Name, assetID uint64, amount *big.Int) error {
//	acct, err := am.GetAccountByName(accountName)
//	if err != nil {
//		return err
//	}
//	if acct == nil {
//		return ErrAccountNotExist
//	}
//	return acct.AddBalanceByID(assetID, amount)
//	rerturn
//}

//Process account action
func (am *AccountManager) Process(action *types.Action) error {
	snap := am.sdb.Snapshot()
	err := am.process(action)
	if err != nil {
		am.sdb.RevertToSnapshot(snap)
	}
	return err
}

func (am *AccountManager) process(action *types.Action) error {
	//transfer
	if action.Value().Cmp(big.NewInt(0)) > 0 {
		if action.Type() == types.CreateAccount || action.Type() == types.DestroyAsset {
			if action.Recipient() != common.Name(sysName) {
				return ErrToNameInvalid
			}
		}
		if err := am.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value()); err != nil {
			return err
		}
	}

	//transaction
	switch action.Type() {
	case types.CreateAccount:
		var acct AccountAction
		err := rlp.DecodeBytes(action.Data(), &acct)
		if err != nil {
			return err
		}

		if err := am.CreateAccount(acct.AccountName, acct.Founder, 0, acct.PublicKey); err != nil {
			return err
		}

		break
	case types.UpdateAccount:
		var acct AccountAction
		err := rlp.DecodeBytes(action.Data(), &acct)
		if err != nil {
			return err
		}

		if err := am.UpdateAccount(action.Sender(), acct.Founder, 0, acct.PublicKey); err != nil {
			return err
		}
		break
		//case types.DeleteAccount:
		//	if err := am.DeleteAccountByName(action.Sender()); err != nil {
		//		return err
		//	}
		//	break
	case types.IssueAsset:
		var asset asset.AssetObject
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return err
		}
		if err := am.IssueAsset(&asset); err != nil {
			return err
		}
		break
	case types.IncreaseAsset:
		var inc IncAsset
		err := rlp.DecodeBytes(action.Data(), &inc)
		if err != nil {
			return err
		}
		if err = am.IncAsset2Acct(action.Sender(), inc.To, inc.AssetId, inc.Amount); err != nil {
			return err
		}
		break

	case types.DestroyAsset:
		var asset asset.AssetObject
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return err
		}

		if err := am.TransferAsset(action.Sender(), common.Name(sysName), action.AssetID(), action.Value()); err != nil {
			return err
		}

		if err = am.SubAccountBalanceByID(common.Name(sysName), asset.GetAssetId(), asset.GetAssetAmount()); err != nil {
			return err
		}

		if err = am.ast.DestroyAsset(common.Name(sysName), asset.GetAssetId(), asset.GetAssetAmount()); err != nil {
			return err
		}
		break
	case types.UpdateAsset:
		var asset asset.AssetObject
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return err
		}
		acct, err := am.GetAccountByName(asset.GetAssetOwner())
		if err != nil {
			return err
		}
		if acct == nil {
			return ErrAccountNotExist
		}
		if len(asset.GetAssetFounder().String()) > 0 {
			acct, err := am.GetAccountByName(asset.GetAssetFounder())
			if err != nil {
				return err
			}
			if acct == nil {
				return ErrAccountNotExist
			}
		}
		if err := am.ast.UpdateAsset(action.Sender(), asset.GetAssetId(), asset.GetAssetOwner(), asset.GetAssetFounder()); err != nil {
			return err
		}
		break
	case types.SetAssetOwner:
		var asset asset.AssetObject
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return err
		}
		acct, err := am.GetAccountByName(asset.GetAssetOwner())
		if err != nil {
			return err
		}
		if acct == nil {
			return ErrAccountNotExist
		}
		if err := am.ast.SetAssetNewOwner(action.Sender(), asset.GetAssetId(), asset.GetAssetOwner()); err != nil {
			return err
		}
		break
	// case types.SetAssetFounder:
	// 	var asset asset.AssetObject
	// 	err := rlp.DecodeBytes(action.Data(), &asset)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if len(asset.GetAssetFounder().String()) > 0 {
	// 		acct, err := am.GetAccountByName(asset.GetAssetFounder())
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if acct == nil {
	// 			return ErrAccountNotExist
	// 		}
	// 	}
	// 	if err = am.ast.SetAssetFounder(action.Sender(), asset.GetAssetId(), asset.GetAssetFounder()); err != nil {
	// 		return err
	// 	}
	// 	break
	case types.Transfer:
		//return am.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value())
		break
	default:
		return ErrUnkownTxType
	}

	return nil
}
