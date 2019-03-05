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
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
)

// AssetBalance asset and balance struct
type AssetBalance struct {
	AssetID uint64   `json:"assetID"`
	Balance *big.Int `json:"balance"`
}

func newAssetBalance(assetID uint64, amount *big.Int) *AssetBalance {
	ab := AssetBalance{
		AssetID: assetID,
		Balance: amount,
	}
	return &ab
}

//Account account object
type Account struct {
	//LastTime *big.Int
	AcctName    common.Name   `json:"accountName"`
	Founder     common.Name   `json:"founder"`
	ChargeRatio uint64        `json:"chargeRatio"`
	Nonce       uint64        `json:"nonce"`
	PublicKey   common.PubKey `json:"publicKey"`
	Code        []byte        `json:"code"`
	CodeHash    common.Hash   `json:"codeHash"`
	CodeSize    uint64        `json:"codeSize"`
	//sort by asset id asc
	Balances []*AssetBalance `json:"balances"`
	//code Suicide
	Suicide bool `json:"suicide"`
	//account destroy
	Destroy bool `json:"destroy"`
}

// NewAccount create a new account object.
func NewAccount(accountName common.Name, founderName common.Name, pubkey common.PubKey) (*Account, error) {
	if !common.IsValidName(accountName.String()) {
		return nil, ErrAccountNameInvalid
	}

	acctObject := Account{
		AcctName:    accountName,
		Founder:     founderName,
		ChargeRatio: 0,
		PublicKey:   pubkey,
		Nonce:       0,
		Balances:    make([]*AssetBalance, 0),
		Code:        make([]byte, 0),
		CodeHash:    crypto.Keccak256Hash(nil),
		Suicide:     false,
		Destroy:     false,
	}
	return &acctObject, nil
}

func (a *Account) HaveCode() bool {
	if a.GetCodeSize() == 0 {
		return false
	}
	return true
}

func (a *Account) IsEmpty() bool {
	if a.GetCodeSize() == 0 && len(a.Balances) == 0 && a.Nonce == 0 {
		return true
	}
	return false
}

// GetName return account object name
func (a *Account) GetName() common.Name {
	return a.AcctName
}

func (a *Account) GetFounder() common.Name {
	return a.Founder
}

func (a *Account) SetFounder(f common.Name) {
	a.Founder = f
}

func (a *Account) GetChargeRatio() uint64 {
	return a.ChargeRatio
}

func (a *Account) SetChargeRatio(ra uint64) {
	a.ChargeRatio = ra
}

// GetNonce get nonce
func (a *Account) GetNonce() uint64 {
	return a.Nonce
}

// SetNonce set nonce
func (a *Account) SetNonce(nonce uint64) {
	a.Nonce = nonce
}

//GetPubKey get bugkey
func (a *Account) GetPubKey() common.PubKey {
	return a.PublicKey
}

//SetPubKey set pub key
func (a *Account) SetPubKey(pubkey common.PubKey) {
	a.PublicKey.SetBytes(pubkey.Bytes())
}

//GetCode get code
func (a *Account) GetCode() ([]byte, error) {
	if a.CodeSize == 0 || a.Suicide {
		return nil, ErrCodeIsEmpty
	}
	return a.Code, nil
}

// GetCodeSize get code size
func (a *Account) GetCodeSize() uint64 {
	return a.CodeSize
}

// SetCode set code
func (a *Account) SetCode(code []byte) error {
	if len(code) == 0 {
		return ErrCodeIsEmpty
	}
	a.Code = code
	a.CodeHash = crypto.Keccak256Hash(code)
	a.CodeSize = uint64(len(code))
	return nil
}

// GetCodeHash get code hash
func (a *Account) GetCodeHash() (common.Hash, error) {
	if len(a.CodeHash) == 0 {
		return common.Hash{}, ErrHashIsEmpty
	}
	return a.CodeHash, nil
}

//GetBalanceByID get balance by asset id
func (a *Account) GetBalanceByID(assetID uint64) (*big.Int, error) {
	if assetID == 0 {
		return big.NewInt(0), ErrAssetIDInvalid
	}
	if p, find := a.binarySearch(assetID); find == true {
		return a.Balances[p].Balance, nil
	}
	log.Info("asset not exist", "account", "asset", ErrAccountAssetNotExist, a.AcctName, assetID)
	return big.NewInt(0), ErrAccountAssetNotExist
}

//GetBalancesList get all balance list
func (a *Account) GetBalancesList() []*AssetBalance {
	return a.Balances
}

//GetAllBalances get all balance list
func (a *Account) GetAllBalances() (map[uint64]*big.Int, error) {
	var ba = make(map[uint64]*big.Int, 0)
	for _, ab := range a.Balances {
		ba[ab.AssetID] = ab.Balance
	}
	return ba, nil
}

// BinarySearch binary search
func (a *Account) binarySearch(assetID uint64) (int64, bool) {
	if len(a.Balances) == 0 {
		return 0, false
	}
	low := int64(0)
	high := int64(len(a.Balances)) - 1

	for low <= high {
		mid := (low + high) / 2
		if a.Balances[mid].AssetID < assetID {
			low = mid + 1
		} else if a.Balances[mid].AssetID > assetID {
			high = mid - 1
		} else if a.Balances[mid].AssetID == assetID {
			return mid, true
		}
	}
	if high < 0 {
		high = 0
	}
	return high, false
}

//AddNewAssetByAssetID add a new asset to balance list
func (a *Account) AddNewAssetByAssetID(assetID uint64, amount *big.Int) {
	p, find := a.binarySearch(assetID)
	if find {
		a.Balances[p].Balance = amount
	} else {
		//append
		if len(a.Balances) == 0 || ((a.Balances[p].AssetID < assetID) && (p+1 == int64(len(a.Balances)))) {
			a.Balances = append(a.Balances, newAssetBalance(assetID, amount))
		} else {
			//insert
			if a.Balances[p].AssetID < assetID {
				//insert back
				p = p + 1
				tail := append([]*AssetBalance{}, a.Balances[p:]...)
				a.Balances = append(a.Balances[:p], newAssetBalance(assetID, amount))
				a.Balances = append(a.Balances, tail...)
			} else {
				//insert front
				if len(a.Balances) > 1 {
					if a.Balances[p].AssetID < assetID {
						p = p + 1
					}
					tail := append([]*AssetBalance{}, a.Balances[p:]...)
					a.Balances = append(a.Balances[:p], newAssetBalance(assetID, amount))
					a.Balances = append(a.Balances, tail...)
				} else {
					tail := append([]*AssetBalance{}, a.Balances[p:]...)
					a.Balances = append([]*AssetBalance{}, newAssetBalance(assetID, amount))
					a.Balances = append(a.Balances, tail...)
				}
			}
		}
	}
	return
}

//SetBalance set amount to balance
func (a *Account) SetBalance(assetID uint64, amount *big.Int) error {
	p, find := a.binarySearch(assetID)
	if find {
		a.Balances[p].Balance = amount
		return nil
	}
	return asset.ErrAssetNotExist
}

func (a *Account) SubBalanceByID(assetID uint64, value *big.Int) error {
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	val, err := a.GetBalanceByID(assetID)
	if err != nil {
		return err
	}
	if val.Cmp(big.NewInt(0)) < 0 || val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	a.SetBalance(assetID, new(big.Int).Sub(val, value))
	return nil
}

//AddAccountBalanceByID add balance by assetID
func (a *Account) AddBalanceByID(assetID uint64, value *big.Int) error {
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	val, err := a.GetBalanceByID(assetID)
	if err == ErrAccountAssetNotExist {
		a.AddNewAssetByAssetID(assetID, value)
	} else {
		a.SetBalance(assetID, new(big.Int).Add(val, value))
	}
	return nil
}

func (a *Account) EnoughAccountBalance(assetID uint64, value *big.Int) error {
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	val, err := a.GetBalanceByID(assetID)
	if err != nil {
		return err
	}
	if val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	return nil
}

// IsSuicided suicide
func (a *Account) IsSuicided() bool {
	return a.Suicide
}

// SetSuicide set setSuicide
func (a *Account) SetSuicide() {
	//just make a sign now
	a.CodeSize = 0
	a.Suicide = true
}

//IsDestroyed is destroyed
func (a *Account) IsDestroyed() bool {
	return a.Destroy
}

//SetDestroy set destroy
func (a *Account) SetDestroy() {
	//just make a sign now
	a.Destroy = true
}
