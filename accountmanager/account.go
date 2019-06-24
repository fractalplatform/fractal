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

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
)

// AssetBalance asset and balance struct
type AssetBalance struct {
	AssetID uint64   `json:"assetID"`
	Balance *big.Int `json:"balance"`
}

type recoverActionResult struct {
	acctAuthors map[common.Name]*accountAuthor
}

type accountAuthor struct {
	threshold             uint64
	updateAuthorThreshold uint64
	version               common.Hash
	indexWeight           map[uint64]uint64
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
	AcctName  common.Name `json:"accountName"`
	Founder   common.Name `json:"founder"`
	AccountID uint64      `json:"accountID"`
	Number    uint64      `json:"number"`
	//ChargeRatio           uint64      `json:"chargeRatio"`
	Nonce                 uint64      `json:"nonce"`
	Code                  []byte      `json:"code"`
	CodeHash              common.Hash `json:"codeHash"`
	CodeSize              uint64      `json:"codeSize"`
	Threshold             uint64      `json:"threshold"`
	UpdateAuthorThreshold uint64      `json:"updateAuthorThreshold"`
	AuthorVersion         common.Hash `json:"authorVersion"`
	//sort by asset id asc
	Balances []*AssetBalance `json:"balances"`
	//realated account, pubkey and address
	Authors []*common.Author `json:"authors"`
	//code Suicide
	Suicide bool `json:"suicide"`
	//account destroy
	Destroy     bool   `json:"destroy"`
	Description string `json:"description"`
}

// NewAccount create a new account object.
func NewAccount(accountName common.Name, founderName common.Name, pubkey common.PubKey, description string) (*Account, error) {
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrCreateAccountError
	}

	auth := common.NewAuthor(pubkey, 1)
	acctObject := Account{
		AcctName:              accountName,
		Founder:               founderName,
		AccountID:             0,
		Number:                0,
		Nonce:                 0,
		Balances:              make([]*AssetBalance, 0),
		Code:                  make([]byte, 0),
		CodeHash:              crypto.Keccak256Hash(nil),
		Threshold:             1,
		UpdateAuthorThreshold: 1,
		Authors:               []*common.Author{auth},
		Suicide:               false,
		Destroy:               false,
		Description:           description,
	}
	acctObject.SetAuthorVersion()
	return &acctObject, nil
}

//HaveCode check account have code
func (a *Account) HaveCode() bool {
	return a.GetCodeSize() != 0
}

// IsEmpty check account empty
func (a *Account) IsEmpty() bool {
	return a.GetCodeSize() == 0 && len(a.Balances) == 0 && a.Nonce == 0
}

// GetName return account object name
func (a *Account) GetName() common.Name {
	return a.AcctName
}

//GetFounder return account object founder
func (a *Account) GetFounder() common.Name {
	return a.Founder
}

//SetFounder set account object founder
func (a *Account) SetFounder(f common.Name) {
	a.Founder = f
}

//GetAccountID return account object id
func (a *Account) GetAccountID() uint64 {
	return a.AccountID
}

//SetAccountID set account object id
func (a *Account) SetAccountID(id uint64) {
	a.AccountID = id
}

//GetAccountNumber return account object number
func (a *Account) GetAccountNumber() uint64 {
	return a.Number
}

//SetAccountNumber set account object number
func (a *Account) SetAccountNumber(number uint64) {
	a.Number = number
}

//GetChargeRatio return account charge ratio
// func (a *Account) GetChargeRatio() uint64 {
// 	return a.ChargeRatio
// }

//SetChargeRatio set account object charge ratio
// func (a *Account) SetChargeRatio(ra uint64) {
// 	a.ChargeRatio = ra
// }

// GetNonce get nonce
func (a *Account) GetNonce() uint64 {
	return a.Nonce
}

// SetNonce set nonce
func (a *Account) SetNonce(nonce uint64) {
	a.Nonce = nonce
}

//GetAuthorVersion get author version
func (a *Account) GetAuthorVersion() common.Hash {
	return a.AuthorVersion
}

//SetAuthorVersion set author version
func (a *Account) SetAuthorVersion() {
	a.AuthorVersion = types.RlpHash([]interface{}{
		a.Authors,
		a.Threshold,
		a.UpdateAuthorThreshold,
	})
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

func (a *Account) GetThreshold() uint64 {
	return a.Threshold
}

func (a *Account) SetThreshold(t uint64) {
	a.Threshold = t
}

func (a *Account) SetUpdateAuthorThreshold(t uint64) {
	a.UpdateAuthorThreshold = t
}

func (a *Account) GetUpdateAuthorThreshold() uint64 {
	return a.UpdateAuthorThreshold
}

func (a *Account) AddAuthor(author *common.Author) error {
	for _, auth := range a.Authors {
		if author.Owner.String() == auth.Owner.String() {
			return fmt.Errorf("%s already exist", auth.Owner.String())
		}
	}
	a.Authors = append(a.Authors, author)
	return nil
}

func (a *Account) UpdateAuthor(author *common.Author) error {
	for _, auth := range a.Authors {
		if author.Owner.String() == auth.Owner.String() {
			auth.Weight = author.Weight
			break
		}
	}
	return nil
}

func (a *Account) DeleteAuthor(author *common.Author) error {
	for i, auth := range a.Authors {
		if author.Owner.String() == auth.Owner.String() {
			a.Authors = append(a.Authors[:i], a.Authors[i+1:]...)
			break
		}
	}
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
	p, find := a.binarySearch(assetID)
	if find {
		return a.Balances[p].Balance, nil
	}
	return big.NewInt(0), ErrAccountAssetNotExist
}

//GetBalancesList get all balance list
func (a *Account) GetBalancesList() []*AssetBalance {
	return a.Balances
}

//GetAllBalances get all balance list
func (a *Account) GetAllBalances() (map[uint64]*big.Int, error) {
	var ba = make(map[uint64]*big.Int)
	for _, ab := range a.Balances {
		ba[ab.AssetID] = ab.Balance
	}
	return ba, nil
}

// BinarySearch binary search
func (a *Account) binarySearch(assetID uint64) (int64, bool) {

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
func (a *Account) AddNewAssetByAssetID(p int64, assetID uint64, amount *big.Int) {
	//append
	if len(a.Balances) == 0 || ((a.Balances[p].AssetID < assetID) && (p+1 == int64(len(a.Balances)))) {
		a.Balances = append(a.Balances, newAssetBalance(assetID, amount))
	} else {
		//insert
		if a.Balances[p].AssetID < assetID {
			//insert after p
			p = p + 1
		}
		tail := append([]*AssetBalance{}, a.Balances[p:]...)
		a.Balances = append(a.Balances[:p], newAssetBalance(assetID, amount))
		a.Balances = append(a.Balances, tail...)
	}
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

//SubBalanceByID sub balance by assetID
func (a *Account) SubBalanceByID(assetID uint64, value *big.Int) error {
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}
	p, find := a.binarySearch(assetID)
	if !find {
		return ErrAccountAssetNotExist
	}
	val := a.Balances[p].Balance
	if val.Cmp(big.NewInt(0)) < 0 || val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}
	a.Balances[p].Balance = new(big.Int).Sub(val, value)
	return nil
}

//AddBalanceByID add balance by assetID
func (a *Account) AddBalanceByID(assetID uint64, value *big.Int) (bool, error) {
	if value.Cmp(big.NewInt(0)) < 0 {
		return false, ErrAmountValueInvalid
	}
	isNew := false
	p, find := a.binarySearch(assetID)
	if !find {
		a.AddNewAssetByAssetID(p, assetID, value)
		isNew = true
	} else {
		a.Balances[p].Balance = new(big.Int).Add(a.Balances[p].Balance, value)
	}
	return isNew, nil
}

//EnoughAccountBalance check account have enough asset balance
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
