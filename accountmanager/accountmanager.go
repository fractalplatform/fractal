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
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/snapshot"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	acctRegExp          = regexp.MustCompile(`^([a-z][a-z0-9]{6,15})(?:\.([a-z0-9]{1,8})){0,1}$`)
	acctRegExpFork1     = regexp.MustCompile(`^([a-z][a-z0-9]{11,15})(\.([a-z0-9]{2,16})){0,2}$`)
	accountNameLength   = uint64(31)
	acctManagerName     = "sysAccount"
	acctInfoPrefix      = "acctInfo"
	accountNameIDPrefix = "accountNameId"
	counterPrefix       = "accountCounter"
	counterID           = uint64(4096)
)

type AuthorActionType uint64

const (
	AddAuthor AuthorActionType = iota
	UpdateAuthor
	DeleteAuthor
)

type CreateAccountAction struct {
	AccountName common.Name   `json:"accountName,omitempty"`
	Founder     common.Name   `json:"founder,omitempty"`
	PublicKey   common.PubKey `json:"publicKey,omitempty"`
	Description string        `json:"description,omitempty"`
}

type UpdataAccountAction struct {
	Founder common.Name `json:"founder,omitempty"`
}

type AuthorAction struct {
	ActionType AuthorActionType
	Author     *common.Author
}

type AccountAuthorAction struct {
	Threshold             uint64          `json:"threshold,omitempty"`
	UpdateAuthorThreshold uint64          `json:"updateAuthorThreshold,omitempty"`
	AuthorActions         []*AuthorAction `json:"authorActions,omitempty"`
}

type IssueAsset struct {
	AssetName   string      `json:"assetName"`
	Symbol      string      `json:"symbol"`
	Amount      *big.Int    `json:"amount"`
	Decimals    uint64      `json:"decimals"`
	Founder     common.Name `json:"founder"`
	Owner       common.Name `json:"owner"`
	UpperLimit  *big.Int    `json:"upperLimit"`
	Contract    common.Name `json:"contract"`
	Description string      `json:"description"`
}

type IncAsset struct {
	AssetID uint64      `json:"assetId,omitempty"`
	Amount  *big.Int    `json:"amount,omitempty"`
	To      common.Name `json:"acceptor,omitempty"`
}

type UpdateAsset struct {
	AssetID uint64      `json:"assetId,omitempty"`
	Founder common.Name `json:"founder"`
}

type UpdateAssetOwner struct {
	AssetID uint64      `json:"assetId,omitempty"`
	Owner   common.Name `json:"owner"`
}

type UpdateAssetContract struct {
	AssetID  uint64      `json:"assetId,omitempty"`
	Contract common.Name `json:"contract"`
}

//AccountManager represents account management model.
type AccountManager struct {
	sdb *state.StateDB
	ast *asset.Asset
}

func SetAccountNameConfig(config *Config) bool {
	if config.AccountNameLevel < 1 || config.AccountNameMaxLength < config.MainAccountNameMinLength || config.MainAccountNameMinLength >= config.MainAccountNameMaxLength {
		panic("account name level config error 1")
	}

	if config.AccountNameLevel > 1 && (config.SubAccountNameMinLength < 1 || config.SubAccountNameMinLength >= config.SubAccountNameMaxLength) {
		panic("account name level config error 2")
	}

	regexpStr := fmt.Sprintf("([a-z][a-z0-9]{%v,%v})", config.MainAccountNameMinLength-1, config.MainAccountNameMaxLength-1)
	for i := 1; i < int(config.AccountNameLevel); i++ {
		regexpStr += fmt.Sprintf("(?:\\.([a-z0-9]{%v,%v})){0,1}", config.SubAccountNameMinLength, config.SubAccountNameMaxLength)
	}

	regexp, err := regexp.Compile(fmt.Sprintf("^%s$", regexpStr))
	if err != nil {
		panic(err)
	}
	acctRegExp = regexp
	accountNameLength = config.AccountNameMaxLength
	return true
}

func GetAccountNameRegExp() *regexp.Regexp {
	return acctRegExp
}

func GetAccountNameRegExpFork1() *regexp.Regexp {
	return acctRegExpFork1
}

func GetAccountNameLength() uint64 {
	return accountNameLength
}

//SetAcctMangerName  set the global account manager name
func SetAcctMangerName(name common.Name) {
	acctManagerName = name.String()
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
	am := &AccountManager{
		sdb: db,
		ast: asset.NewAsset(db),
	}

	am.initAccountCounter()
	return am, nil
}

//initAccountCounter init account manage counter
func (am *AccountManager) initAccountCounter() {
	_, err := am.getAccountCounter()
	if err == ErrCounterNotExist {
		//var counterID uint64
		//counterID = 0
		//store assetCount
		b, err := rlp.EncodeToBytes(&counterID)
		if err != nil {
			panic(err)
		}
		am.sdb.Put(acctManagerName, counterPrefix, b)
	}
}

//getAccountCounter get account counter cur value
func (am *AccountManager) getAccountCounter() (uint64, error) {
	b, err := am.sdb.Get(acctManagerName, counterPrefix)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, ErrCounterNotExist
	}
	var accountCounter uint64
	err = rlp.DecodeBytes(b, &accountCounter)
	if err != nil {
		return 0, err
	}
	return accountCounter, nil
}

// AccountIsExist check account is exist.
func (am *AccountManager) AccountIsExist(accountName common.Name) (bool, error) {
	//check is exist
	accountID, err := am.GetAccountIDByName(accountName)
	if err != nil {
		return false, err
	}
	if accountID > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// AccountIsExistByID check account is exist by ID.
func (am *AccountManager) AccountIsExistByID(accountID uint64) (bool, error) {
	//check is exist
	account, err := am.GetAccountById(accountID)
	if err != nil {
		return false, err
	}
	if account != nil {
		return true, nil
	} else {
		return false, nil
	}
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

const (
	unKnown uint64 = iota
	mainAccount
	subAccount
)

func GetAccountNameLevel(accountName common.Name) (uint64, error) {
	if !accountName.IsValid(acctRegExp, accountNameLength) {
		return unKnown, fmt.Errorf("account %s is invalid", accountName.String())
	}

	if len(strings.Split(accountName.String(), ".")) == 1 {
		return mainAccount, nil
	}

	return subAccount, nil
}

func (am *AccountManager) checkAccountNameValid(fromName common.Name, accountName common.Name) error {
	accountLevel, err := GetAccountNameLevel(accountName)
	if err != nil {
		return err
	}

	if accountLevel == mainAccount {
		if !accountName.IsValid(acctRegExpFork1, accountNameLength) {
			return fmt.Errorf("account %s is invalid", accountName.String())
		}
	}

	if accountLevel == subAccount {
		if !fromName.IsChildren(accountName) {
			return ErrAccountInvalid
		}
	}

	return nil
}

//CreateAccount create account
func (am *AccountManager) CreateAccount(fromName common.Name, accountName common.Name, founderName common.Name, number uint64, curForkID uint64, pubkey common.PubKey, detail string) error {
	if curForkID >= params.ForkID1 {
		if err := am.checkAccountNameValid(fromName, accountName); err != nil {
			return err
		}
	} else {
		if len(common.FindStringSubmatch(acctRegExp, accountName.String())) > 1 {
			if !fromName.IsChildren(accountName) {
				return ErrAccountInvalid
			}
		}

		if !accountName.IsValid(acctRegExp, accountNameLength) {
			return fmt.Errorf("account %s is invalid", accountName.String())
		}
	}

	//check is exist
	accountID, err := am.GetAccountIDByName(accountName)
	if err != nil {
		return err
	}
	if accountID > 0 {
		return ErrAccountIsExist
	}

	// asset and account name diff
	_, err = am.ast.GetAssetIDByName(accountName.String())
	if err == nil {
		return ErrNameIsExist
	}

	var fName common.Name
	if len(founderName.String()) > 0 && founderName != accountName {
		f, err := am.GetAccountByName(founderName)
		if err != nil {
			return err
		}
		if f == nil {
			return ErrAccountNotExist
		}
		fName.SetString(founderName.String())
	} else {
		fName.SetString(accountName.String())
	}

	acctObj, err := NewAccount(accountName, fName, pubkey, detail)
	if err != nil {
		return err
	}
	if acctObj == nil {
		return ErrCreateAccountError
	}

	//get accountCounter
	accountCounter, err := am.getAccountCounter()
	if err != nil {
		return err
	}
	accountCounter = accountCounter + 1
	//set account id
	acctObj.SetAccountID(accountCounter)

	//store account name with account id
	aid, err := rlp.EncodeToBytes(&accountCounter)
	if err != nil {
		return err
	}
	acctObj.SetAccountNumber(number)
	//acctObj.SetChargeRatio(0)
	am.SetAccount(acctObj)
	am.sdb.Put(acctManagerName, accountNameIDPrefix+accountName.String(), aid)
	am.sdb.Put(acctManagerName, counterPrefix, aid)
	return nil
}

//SetChargeRatio set the Charge Ratio of the account
// func (am *AccountManager) SetChargeRatio(accountName common.Name, ra uint64) error {
// 	acct, err := am.GetAccountByName(accountName)
// 	if acct == nil {
// 		return ErrAccountNotExist
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	acct.SetChargeRatio(ra)
// 	return am.SetAccount(acct)
// }

//UpdateAccount update the pubkey of the account
func (am *AccountManager) UpdateAccount(accountName common.Name, accountAction *UpdataAccountAction) error {
	acct, err := am.GetAccountByName(accountName)
	if acct == nil {
		return ErrAccountNotExist
	}
	if err != nil {
		return err
	}
	if len(accountAction.Founder.String()) > 0 {
		f, err := am.GetAccountByName(accountAction.Founder)
		if err != nil {
			return err
		}
		if f == nil {
			return ErrAccountNotExist
		}
	} else {
		accountAction.Founder.SetString(accountName.String())
	}

	acct.SetFounder(accountAction.Founder)
	return am.SetAccount(acct)
}

func (am *AccountManager) UpdateAccountAuthor(accountName common.Name, acctAuth *AccountAuthorAction) error {
	acct, err := am.GetAccountByName(accountName)
	if acct == nil {
		return ErrAccountNotExist
	}
	if err != nil {
		return err
	}
	if acctAuth.Threshold != 0 {
		acct.SetThreshold(acctAuth.Threshold)
	}
	if acctAuth.UpdateAuthorThreshold != 0 {
		acct.SetUpdateAuthorThreshold(acctAuth.UpdateAuthorThreshold)
	}
	for _, authorAct := range acctAuth.AuthorActions {
		actionTy := authorAct.ActionType
		switch actionTy {
		case AddAuthor:
			acct.AddAuthor(authorAct.Author)
		case UpdateAuthor:
			acct.UpdateAuthor(authorAct.Author)
		case DeleteAuthor:
			acct.DeleteAuthor(authorAct.Author)
		default:
			return fmt.Errorf("invalid account author operation type %d", actionTy)
		}
	}
	if uint64(len(acct.Authors)) > params.MaxAuthorNum {
		return fmt.Errorf("account author length can not exceed %d", params.MaxAuthorNum)
	}
	acct.SetAuthorVersion()
	return am.SetAccount(acct)
}

//GetAccountByTime get account by name and time
func (am *AccountManager) GetAccountByTime(accountName common.Name, time uint64) (*Account, error) {
	accountID, err := am.GetAccountIDByName(accountName)
	if err != nil {
		return nil, err
	}

	snapshotManager := snapshot.NewSnapshotManager(am.sdb)
	b, err := snapshotManager.GetSnapshotMsg(acctManagerName, acctInfoPrefix+strconv.FormatUint(accountID, 10), time)
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
	accountID, err := am.GetAccountIDByName(accountName)
	if err != nil {
		return nil, err
	}
	return am.GetAccountById(accountID)
}

//GetAccountIDByName get account id by account name
func (am *AccountManager) GetAccountIDByName(accountName common.Name) (uint64, error) {
	if accountName == "" {
		return 0, nil
	}
	b, err := am.sdb.Get(acctManagerName, accountNameIDPrefix+accountName.String())
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	var accountID uint64
	if err := rlp.DecodeBytes(b, &accountID); err != nil {
		return 0, err
	}
	return accountID, nil
}

//GetAccountById get account by account id
func (am *AccountManager) GetAccountById(id uint64) (*Account, error) {
	if id == 0 {
		return nil, nil
	}

	b, err := am.sdb.Get(acctManagerName, acctInfoPrefix+strconv.FormatUint(id, 10))

	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		log.Debug("account not exist", "id", ErrAccountNotExist, id)
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
	if acct.IsDestroyed() {
		return ErrAccountIsDestroy
	}
	b, err := rlp.EncodeToBytes(acct)
	if err != nil {
		return err
	}

	//am.sdb.Put(acctManagerName, acctInfoPrefix+acct.GetName().String(), b)
	am.sdb.Put(acctManagerName, acctInfoPrefix+strconv.FormatUint(acct.GetAccountID(), 10), b)
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

// GetAuthorVersion returns the account author version
func (am *AccountManager) GetAuthorVersion(accountName common.Name) (common.Hash, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return common.Hash{}, err
	}
	if acct == nil {
		return common.Hash{}, ErrAccountNotExist
	}
	return acct.GetAuthorVersion(), nil
}

func (am *AccountManager) getParentAccount(accountName common.Name, parentIndex uint64) (common.Name, error) {
	if parentIndex == 0 {
		return accountName, nil
	}

	list := strings.Split(accountName.String(), ".")
	if parentIndex > uint64(len(list)-1) {
		return common.Name(""), fmt.Errorf("invalid index, %s , %d", accountName.String(), parentIndex)
	}

	level := uint64(len(list)) - parentIndex
	an := list[0]
	for i := uint64(1); i < level; i++ {
		an = an + "." + list[i]
	}

	return common.Name(an), nil
}

// RecoverTx Make sure the transaction is signed properly and validate account authorization.
func (am *AccountManager) RecoverTx(signer types.Signer, tx *types.Transaction) error {
	authorVersion := make(map[common.Name]common.Hash)
	for _, action := range tx.GetActions() {
		pubs, err := types.RecoverMultiKey(signer, action, tx)
		if err != nil {
			return err
		}

		if uint64(len(pubs)) > params.MaxSignLength {
			return fmt.Errorf("exceed max sign length, want most %d, actual is %d", params.MaxSignLength, len(pubs))
		}

		parentIndex := action.GetSignParent()
		signSender, err := am.getParentAccount(action.Sender(), parentIndex)
		if err != nil {
			return err
		}
		recoverRes := &recoverActionResult{make(map[common.Name]*accountAuthor)}
		for i, pub := range pubs {
			index := action.GetSignIndex(uint64(i))
			if uint64(len(index)) > params.MaxSignDepth {
				return fmt.Errorf("exceed max sign depth, want most %d, actual is %d", params.MaxSignDepth, len(index))
			}

			if err := am.ValidSign(signSender, pub, index, recoverRes); err != nil {
				return err
			}
		}

		for name, acctAuthor := range recoverRes.acctAuthors {
			var count uint64
			for _, weight := range acctAuthor.indexWeight {
				count += weight
			}
			threshold := acctAuthor.threshold
			if name.String() == signSender.String() && (action.Type() == types.UpdateAccountAuthor || signSender != action.Sender()) {
				threshold = acctAuthor.updateAuthorThreshold
			}
			if count < threshold {
				return fmt.Errorf("account %s want threshold %d, but actual is %d", name, threshold, count)
			}
			authorVersion[name] = acctAuthor.version
		}

		types.StoreAuthorCache(action, authorVersion)
	}
	if tx.PayerExist() {
		for _, action := range tx.GetActions() {
			pubs, err := types.RecoverPayerMultiKey(signer, action, tx)
			if err != nil {
				return err
			}

			if uint64(len(pubs)) > params.MaxSignLength {
				return fmt.Errorf("exceed max sign length, want most %d, actual is %d", params.MaxSignLength, len(pubs))
			}

			sig := action.PayerSignature()
			if sig == nil {
				return fmt.Errorf("payer signature is nil")
			}
			parentIndex := sig.ParentIndex
			signSender, err := am.getParentAccount(action.Payer(), parentIndex)
			if err != nil {
				return err
			}
			recoverRes := &recoverActionResult{make(map[common.Name]*accountAuthor)}
			for i, pub := range pubs {
				index := sig.SignData[uint64(i)].Index
				if uint64(len(index)) > params.MaxSignDepth {
					return fmt.Errorf("exceed max sign depth, want most %d, actual is %d", params.MaxSignDepth, len(index))
				}

				if err := am.ValidSign(signSender, pub, index, recoverRes); err != nil {
					return err
				}
			}

			for name, acctAuthor := range recoverRes.acctAuthors {
				var count uint64
				for _, weight := range acctAuthor.indexWeight {
					count += weight
				}
				threshold := acctAuthor.threshold
				if count < threshold {
					return fmt.Errorf("account %s want threshold %d, but actual is %d", name, threshold, count)
				}
				authorVersion[name] = acctAuthor.version
			}

			types.StoreAuthorCache(action, authorVersion)
		}
	}
	return nil
}

// IsValidSign
func (am *AccountManager) IsValidSign(accountName common.Name, pub common.PubKey) error {
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

	for _, author := range acct.Authors {
		if author.String() == pub.String() && author.GetWeight() >= acct.GetThreshold() {
			return nil
		}
	}
	return fmt.Errorf("%v %v excepted %v", acct.AcctName, ErrkeyNotSame, pub.String())
}

//ValidSign check the sign
func (am *AccountManager) ValidSign(accountName common.Name, pub common.PubKey, index []uint64, recoverRes *recoverActionResult) error {
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

	var i int
	var idx uint64
	for i, idx = range index {
		if idx >= uint64(len(acct.Authors)) {
			return fmt.Errorf("acct authors modified")
		}
		if i == len(index)-1 {
			break
		}
		switch ownerTy := acct.Authors[idx].Owner.(type) {
		case common.Name:
			nextacct, err := am.GetAccountByName(ownerTy)
			if err != nil {
				return err
			}
			if nextacct == nil {
				return ErrAccountNotExist
			}
			if nextacct.IsDestroyed() {
				return ErrAccountIsDestroy
			}
			if recoverRes.acctAuthors[acct.GetName()] == nil {
				a := &accountAuthor{version: acct.AuthorVersion, threshold: acct.Threshold, updateAuthorThreshold: acct.UpdateAuthorThreshold, indexWeight: map[uint64]uint64{idx: acct.Authors[idx].GetWeight()}}
				recoverRes.acctAuthors[acct.GetName()] = a
			} else {
				recoverRes.acctAuthors[acct.GetName()].indexWeight[idx] = acct.Authors[idx].GetWeight()
			}
			acct = nextacct
		default:
			return ErrAccountNotExist
		}
	}
	return am.ValidOneSign(acct, idx, pub, recoverRes)
}

func (am *AccountManager) ValidOneSign(acct *Account, index uint64, pub common.PubKey, recoverRes *recoverActionResult) error {
	switch ownerTy := acct.Authors[index].Owner.(type) {
	case common.PubKey:
		if pub.Compare(ownerTy) != 0 {
			return fmt.Errorf("%v %v have %v excepted %v", acct.AcctName, ErrkeyNotSame, pub.String(), ownerTy.String())
		}
	case common.Address:
		addr := common.BytesToAddress(crypto.Keccak256(pub.Bytes()[1:])[12:])
		if addr.Compare(ownerTy) != 0 {
			return fmt.Errorf("%v %v have %v excepted %v", acct.AcctName, ErrkeyNotSame, addr.String(), ownerTy.String())
		}
	default:
		return fmt.Errorf("wrong sign type")
	}
	if recoverRes.acctAuthors[acct.GetName()] == nil {
		a := &accountAuthor{version: acct.AuthorVersion, threshold: acct.Threshold, updateAuthorThreshold: acct.UpdateAuthorThreshold, indexWeight: map[uint64]uint64{index: acct.Authors[index].GetWeight()}}
		recoverRes.acctAuthors[acct.GetName()] = a
		return nil
	}
	recoverRes.acctAuthors[acct.GetName()].indexWeight[index] = acct.Authors[index].GetWeight()
	return nil
}

//GetAssetInfoByName get asset info by asset name.
func (am *AccountManager) GetAssetInfoByName(assetName string) (*asset.AssetObject, error) {
	assetID, err := am.ast.GetAssetIDByName(assetName)
	if err != nil {
		return nil, err
	}
	return am.ast.GetAssetObjectByID(assetID)
}

//GetAssetInfoByID get asset info by assetID
func (am *AccountManager) GetAssetInfoByID(assetID uint64) (*asset.AssetObject, error) {
	return am.ast.GetAssetObjectByID(assetID)
}

// GetAllAssetByAssetID get accout asset and subAsset information
func (am *AccountManager) GetAllAssetByAssetID(acct *Account, assetID uint64) (map[uint64]*big.Int, error) {
	var ba = make(map[uint64]*big.Int)

	b, err := acct.GetBalanceByID(assetID)
	if err != nil {
		return nil, err
	}
	ba[assetID] = b

	assetObj, err := am.ast.GetAssetObjectByID(assetID)
	if err != nil {
		return nil, err
	}

	assetName := assetObj.GetAssetName()
	balances, err := acct.GetAllBalances()
	if err != nil {
		return nil, err
	}

	for id, balance := range balances {
		subAssetObj, err := am.ast.GetAssetObjectByID(id)
		if err != nil {
			return nil, err
		}

		if common.StrToName(assetName).IsChildren(common.StrToName(subAssetObj.GetAssetName())) {
			ba[id] = balance
		}
	}

	return ba, nil
}

// GetAllBalanceByAssetID get account balance, balance(asset) = asset + subAsset
func (am *AccountManager) GetAllBalanceByAssetID(acct *Account, assetID uint64) (*big.Int, error) {
	var ba *big.Int
	ba = big.NewInt(0)

	b, _ := acct.GetBalanceByID(assetID)
	ba = ba.Add(ba, b)

	assetObj, err := am.ast.GetAssetObjectByID(assetID)
	if err != nil {
		return big.NewInt(0), err
	}

	assetName := assetObj.GetAssetName()
	balances, err := acct.GetAllBalances()
	if err != nil {
		return big.NewInt(0), err
	}

	for id, balance := range balances {
		subAssetObj, err := am.ast.GetAssetObjectByID(id)
		if err != nil {
			return big.NewInt(0), err
		}

		if common.StrToName(assetName).IsChildren(common.StrToName(subAssetObj.GetAssetName())) {
			ba = ba.Add(ba, balance)
		}
	}

	return ba, nil
}

//GetBalanceByTime get account balance by Time
func (am *AccountManager) GetBalanceByTime(accountName common.Name, assetID uint64, typeID uint64, time uint64) (*big.Int, error) {
	acct, err := am.GetAccountByTime(accountName, time)
	if err != nil {
		return big.NewInt(0), err
	}
	if acct == nil {
		return big.NewInt(0), ErrAccountNotExist
	}

	if typeID == 0 {
		return acct.GetBalanceByID(assetID)
	} else if typeID == 1 {
		return am.GetAllBalanceByAssetID(acct, assetID)
	} else {
		return big.NewInt(0), fmt.Errorf("type ID %d invalid", typeID)
	}
}

//GetAccountBalanceByID get account balance by ID
func (am *AccountManager) GetAccountBalanceByID(accountName common.Name, assetID uint64, typeID uint64) (*big.Int, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return big.NewInt(0), err
	}
	if acct == nil {
		return big.NewInt(0), ErrAccountNotExist
	}
	if typeID == 0 {
		return acct.GetBalanceByID(assetID)
	} else if typeID == 1 {
		return am.GetAllBalanceByAssetID(acct, assetID)
	} else {
		return big.NewInt(0), fmt.Errorf("type ID %d invalid", typeID)
	}
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
	snapshotManager := snapshot.NewSnapshotManager(am.sdb)
	if num == 0 {
		return snapshotManager.GetLastSnapshotTime()
	} else if num == 1 {
		return snapshotManager.GetPrevSnapshotTime(time)
	} else if num == 2 {
		t, err := snapshotManager.GetLastSnapshotTime()
		if err != nil {
			return 0, err
		}

		if t <= time {
			return 0, ErrSnapshotTimeNotExist
		} else {
			for {
				if t1, err := snapshotManager.GetPrevSnapshotTime(t); err != nil {
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
	return am.ast.GetAssetFounderByID(assetID)
}

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

	err = acct.SubBalanceByID(assetID, value)
	if err != nil {
		return err
	}

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

	_, err = acct.AddBalanceByID(assetID, value)
	if err != nil {
		return err
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

	assetID, err := am.ast.GetAssetIDByName(assetName)
	if err != nil {
		return err
	}

	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrAmountValueInvalid
	}

	_, err = acct.AddBalanceByID(assetID, value)
	if err != nil {
		return err
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

//SetCode set contract code
func (am *AccountManager) SetCode(accountName common.Name, code []byte) (bool, error) {
	acct, err := am.GetAccountByName(accountName)
	if err != nil {
		return false, err
	}
	if acct == nil {
		return false, ErrAccountNotExist
	}
	err = acct.SetCode(code)
	if err != nil {
		return false, err
	}
	err = am.SetAccount(acct)
	if err != nil {
		return false, err
	}
	return true, nil
}

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
// func (am *AccountManager) GetAccountFromValue(accountName common.Name, key string, value []byte) (*Account, error) {
// 	if len(value) == 0 {
// 		return nil, ErrAccountNotExist
// 	}
// 	if key != accountName.String()+acctInfoPrefix {
// 		return nil, ErrAccountNameInvalid
// 	}
// 	var acct Account
// 	if err := rlp.DecodeBytes(value, &acct); err != nil {
// 		return nil, ErrAccountNotExist
// 	}
// 	if acct.AcctName != accountName {
// 		return nil, ErrAccountNameInvalid
// 	}
// 	return &acct, nil
// }

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
func (am *AccountManager) TransferAsset(fromAccount common.Name, toAccount common.Name, assetID uint64, value *big.Int, fromAccountExtra ...common.Name) error {
	if sign := value.Sign(); sign == 0 {
		return nil
	} else if sign == -1 {
		return ErrNegativeValue
	}

	fromAccountExtra = append(fromAccountExtra, fromAccount)
	fromAccountExtra = append(fromAccountExtra, toAccount)
	if !am.ast.HasAccess(assetID, fromAccountExtra...) {
		return fmt.Errorf("no permissions of asset %v", assetID)
	}
	// if !am.ast.HasAccess(assetID, fromAccount, toAccount) {
	// 	return fmt.Errorf("no permissions of asset %v", assetID)
	// }

	//check from account
	fromAcct, err := am.GetAccountByName(fromAccount)
	if err != nil {
		return err
	}
	if fromAcct == nil {
		return ErrAccountNotExist
	}

	//check from account balance
	val, err := fromAcct.GetBalanceByID(assetID)
	if err != nil {
		return err
	}

	if val.Cmp(big.NewInt(0)) < 0 || val.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	if fromAccount == toAccount || value.Cmp(big.NewInt(0)) == 0 {
		return nil
	}
	//sub from account balance
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
	//add to account balance
	bNew, err := toAcct.AddBalanceByID(assetID, value)
	if err != nil {
		return err
	}
	if bNew {
		err := am.ast.IncStats(assetID)
		if err != nil {
			return err
		}
	}
	if err = am.SetAccount(fromAcct); err != nil {
		return err
	}
	return am.SetAccount(toAcct)
}

func (am *AccountManager) CheckAssetContract(contract common.Name, owner common.Name, from ...common.Name) bool {
	from = append(from, owner)
	for _, name := range from {
		if name == contract {
			return true
		}
	}
	return false
}

func (am *AccountManager) checkAssetNameAndOwner(fromName common.Name, assetInfo *IssueAsset) error {
	var assetNames []string
	var assetPre string

	names := strings.Split(assetInfo.AssetName, ":")
	if len(names) == 2 {
		if !common.StrToName(names[1]).IsValid(asset.GetAssetNameRegExp(), asset.GetAssetNameLength()) {
			return fmt.Errorf("asset name is invalid, name: %v", assetInfo.AssetName)
		}
		assetNames = strings.Split(names[1], ".")
		if len(assetNames) == 1 && names[0] != fromName.String() {
			return fmt.Errorf("asset name not match from, name: %v, from:%v", assetInfo.AssetName, fromName)
		}
		assetPre = names[0] + ":"
	} else {
		if !common.StrToName(assetInfo.AssetName).IsValid(asset.GetAssetNameRegExp(), asset.GetAssetNameLength()) {
			return fmt.Errorf("asset name is invalid, name: %v", assetInfo.AssetName)
		}
		assetNames = strings.Split(assetInfo.AssetName, ".")
		if len(assetNames) < 2 {
			return fmt.Errorf("asset name is invalid, name: %v", assetInfo.AssetName)
		}
		assetPre = ""
	}

	if len(assetNames) == 1 {
		return nil
	}

	//check sub asset owner
	parentAssetID, isValid := am.ast.IsValidAssetOwner(fromName, assetPre, assetNames)
	if !isValid {
		return fmt.Errorf("asset owner is invalid, name: %v", assetInfo.AssetName)
	}
	assetObj, _ := am.ast.GetAssetObjectByID(parentAssetID)
	assetInfo.Decimals = assetObj.GetDecimals()

	return nil
}

func (am *AccountManager) checkAssetInfoValid(fromName common.Name, assetInfo *IssueAsset) error {
	if assetInfo.Owner == "" {
		return fmt.Errorf("asset owner invalid")
	}

	if assetInfo.Amount.Cmp(big.NewInt(0)) < 0 || assetInfo.UpperLimit.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("asset amount or limit invalid, amount:%v,limit:%v", assetInfo.Amount, assetInfo.UpperLimit)
	}

	if assetInfo.UpperLimit.Cmp(big.NewInt(0)) > 0 {
		if assetInfo.Amount.Cmp(assetInfo.UpperLimit) > 0 {
			return fmt.Errorf("asset amount greater than limit, amount:%v,limit:%v", assetInfo.Amount, assetInfo.UpperLimit)
		}
	}

	err := am.checkAssetNameAndOwner(fromName, assetInfo)
	if err != nil {
		return err
	}

	//symbol use asset reg
	if !common.StrToName(assetInfo.Symbol).IsValid(asset.GetAssetNameRegExp(), asset.GetAssetNameLength()) {
		return fmt.Errorf("asset symbol invalid, symbol:%v", assetInfo.Symbol)
	}
	if uint64(len(assetInfo.Description)) > MaxDescriptionLength {
		return fmt.Errorf("asset description invalid, description:%v", assetInfo.Description)
	}

	return nil
}

//IssueAsset issue asset
func (am *AccountManager) IssueAsset(fromName common.Name, asset IssueAsset, number uint64, curForkID uint64) (uint64, error) {
	//check owner valid
	if curForkID >= params.ForkID1 {
		err := am.checkAssetInfoValid(fromName, &asset)
		if err != nil {
			return 0, err
		}
	} else {
		if !am.ast.IsValidMainAssetBeforeFork(asset.AssetName) {
			parentAssetID, isValid := am.ast.IsValidSubAssetBeforeFork(fromName, asset.AssetName)
			if !isValid {
				return 0, fmt.Errorf("account %s can not create %s", fromName, asset.AssetName)
			}
			assetObj, _ := am.ast.GetAssetObjectByID(parentAssetID)
			asset.Decimals = assetObj.GetDecimals()
		}
	}

	//check owner
	acct, err := am.GetAccountByName(asset.Owner)
	if err != nil {
		return 0, err
	}
	if acct == nil {
		return 0, ErrAccountNotExist
	}
	//check founder
	if len(asset.Founder) > 0 {
		f, err := am.GetAccountByName(asset.Founder)
		if err != nil {
			return 0, err
		}
		if f == nil {
			return 0, ErrAccountNotExist
		}
	} else {
		asset.Founder = asset.Owner
	}

	// check asset contract
	if len(asset.Contract) > 0 {
		if curForkID < params.ForkID1 {
			if !asset.Contract.IsValid(acctRegExp, accountNameLength) {
				return 0, fmt.Errorf("account %s is invalid", asset.Contract.String())
			}
		}

		f, err := am.GetAccountByName(asset.Contract)
		if err != nil {
			return 0, err
		}
		if f == nil {
			return 0, ErrAccountNotExist
		}
	}

	// check asset name is not account name
	name := common.StrToName(asset.AssetName)
	accountID, _ := am.GetAccountIDByName(name)
	if accountID > 0 {
		return 0, ErrNameIsExist
	}

	assetID, err := am.ast.IssueAsset(asset.AssetName, number, curForkID, asset.Symbol,
		asset.Amount, asset.Decimals, asset.Founder, asset.Owner,
		asset.UpperLimit, asset.Contract, asset.Description)
	if err != nil {
		return 0, err
	}

	//add the asset to owner
	return assetID, nil
}

//IncAsset2Acct increase asset and add amount to accout balance
func (am *AccountManager) IncAsset2Acct(fromName common.Name, toName common.Name, assetID uint64, amount *big.Int, forkID uint64) error {
	if err := am.ast.CheckOwner(fromName, assetID); err != nil {
		return err
	}

	if err := am.ast.IncreaseAsset(fromName, assetID, amount, forkID); err != nil {
		return err
	}
	return nil
}

//Process account action
func (am *AccountManager) Process(accountManagerContext *types.AccountManagerContext) ([]*types.InternalAction, error) {
	snap := am.sdb.Snapshot()
	internalActions, err := am.process(accountManagerContext)
	if err != nil {
		am.sdb.RevertToSnapshot(snap)
	}
	return internalActions, err
}

func (am *AccountManager) process(accountManagerContext *types.AccountManagerContext) ([]*types.InternalAction, error) {
	action := accountManagerContext.Action
	number := accountManagerContext.Number
	curForkID := accountManagerContext.CurForkID
	var fromAccountExtra []common.Name
	fromAccountExtra = append(fromAccountExtra, accountManagerContext.FromAccountExtra...)

	if err := action.Check(curForkID, accountManagerContext.ChainConfig); err != nil {
		return nil, err
	}

	var internalActions []*types.InternalAction
	//transfer
	if err := am.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value(), fromAccountExtra...); err != nil {
		return nil, err
	}

	//transaction
	switch action.Type() {
	case types.CreateAccount:
		var acct CreateAccountAction
		err := rlp.DecodeBytes(action.Data(), &acct)
		if err != nil {
			return nil, err
		}

		if err := am.CreateAccount(action.Sender(), acct.AccountName, acct.Founder, number, curForkID, acct.PublicKey, acct.Description); err != nil {
			return nil, err
		}

		if action.Value().Cmp(big.NewInt(0)) > 0 {
			if err := am.TransferAsset(common.Name(accountManagerContext.ChainConfig.AccountName), acct.AccountName, action.AssetID(), action.Value(), fromAccountExtra...); err != nil {
				return nil, err
			}
			actionX := types.NewAction(types.Transfer, common.Name(accountManagerContext.ChainConfig.AccountName), acct.AccountName, 0, action.AssetID(), 0, action.Value(), nil, nil)
			internalAction := &types.InternalAction{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
			internalActions = append(internalActions, internalAction)
		}
	case types.UpdateAccount:
		var acct UpdataAccountAction
		err := rlp.DecodeBytes(action.Data(), &acct)
		if err != nil {
			return nil, err
		}

		if err := am.UpdateAccount(action.Sender(), &acct); err != nil {
			return nil, err
		}
	case types.UpdateAccountAuthor:
		var acctAuth AccountAuthorAction
		err := rlp.DecodeBytes(action.Data(), &acctAuth)
		if err != nil {
			return nil, err
		}
		if err := am.UpdateAccountAuthor(action.Sender(), &acctAuth); err != nil {
			return nil, err
		}
	case types.IssueAsset:
		var issueAsset IssueAsset
		err := rlp.DecodeBytes(action.Data(), &issueAsset)
		if err != nil {
			return nil, err
		}
		fromAccountExtra = append(fromAccountExtra, action.Sender())

		if len(issueAsset.Contract) != 0 {
			if !am.CheckAssetContract(issueAsset.Contract, issueAsset.Owner, fromAccountExtra...) && issueAsset.Amount.Sign() != 0 {
				return nil, ErrAmountMustBeZero
			}
		}

		assetID, err := am.IssueAsset(action.Sender(), issueAsset, number, curForkID)
		if err != nil {
			return nil, err
		}

		if err := am.AddAccountBalanceByID(common.Name(accountManagerContext.ChainConfig.AssetName), assetID, issueAsset.Amount); err != nil {
			return nil, err
		}
		actionX := types.NewAction(types.Transfer, common.Name(""), common.Name(accountManagerContext.ChainConfig.AssetName), 0, assetID, 0, issueAsset.Amount, nil, nil)
		internalAction := &types.InternalAction{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
		internalActions = append(internalActions, internalAction)

		if err := am.TransferAsset(common.Name(accountManagerContext.ChainConfig.AssetName), issueAsset.Owner, assetID, issueAsset.Amount, fromAccountExtra...); err != nil {
			return nil, err
		}
		actionX = types.NewAction(types.Transfer, common.Name(accountManagerContext.ChainConfig.AssetName), issueAsset.Owner, 0, assetID, 0, issueAsset.Amount, nil, nil)
		internalAction = &types.InternalAction{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
		internalActions = append(internalActions, internalAction)
	case types.IncreaseAsset:
		var inc IncAsset
		err := rlp.DecodeBytes(action.Data(), &inc)
		if err != nil {
			return nil, err
		}

		if inc.Amount.Cmp(big.NewInt(0)) < 0 {
			return nil, ErrNegativeAmount
		}

		if err := am.IncAsset2Acct(action.Sender(), inc.To, inc.AssetID, inc.Amount, curForkID); err != nil {
			return nil, err
		}

		if err := am.AddAccountBalanceByID(common.Name(accountManagerContext.ChainConfig.AssetName), inc.AssetID, inc.Amount); err != nil {
			return nil, err
		}
		actionX := types.NewAction(types.Transfer, common.Name(""), common.Name(accountManagerContext.ChainConfig.AssetName), 0, inc.AssetID, 0, inc.Amount, nil, nil)
		internalAction := &types.InternalAction{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
		internalActions = append(internalActions, internalAction)

		fromAccountExtra = append(fromAccountExtra, action.Sender())
		if err := am.TransferAsset(common.Name(accountManagerContext.ChainConfig.AssetName), inc.To, inc.AssetID, inc.Amount, fromAccountExtra...); err != nil {
			return nil, err
		}
		actionX = types.NewAction(types.Transfer, common.Name(accountManagerContext.ChainConfig.AssetName), inc.To, 0, inc.AssetID, 0, inc.Amount, nil, nil)
		internalAction = &types.InternalAction{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
		internalActions = append(internalActions, internalAction)
	case types.DestroyAsset:
		if err := am.SubAccountBalanceByID(common.Name(accountManagerContext.ChainConfig.AssetName), action.AssetID(), action.Value()); err != nil {
			return nil, err
		}

		if err := am.ast.DestroyAsset(common.Name(accountManagerContext.ChainConfig.AssetName), action.AssetID(), action.Value()); err != nil {
			return nil, err
		}
		actionX := types.NewAction(types.Transfer, common.Name(accountManagerContext.ChainConfig.AssetName), common.Name(""), 0, action.AssetID(), 0, action.Value(), nil, nil)
		internalAction := &types.InternalAction{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
		internalActions = append(internalActions, internalAction)
	case types.UpdateAsset:
		var asset UpdateAsset
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return nil, err
		}

		if len(asset.Founder.String()) > 0 {
			acct, err := am.GetAccountByName(asset.Founder)
			if err != nil {
				return nil, err
			}
			if acct == nil {
				return nil, ErrAccountNotExist
			}
		}

		if err := am.ast.CheckOwner(action.Sender(), asset.AssetID); err != nil {
			return nil, err
		}

		if err := am.ast.UpdateAsset(action.Sender(), asset.AssetID, asset.Founder, curForkID); err != nil {
			return nil, err
		}
	case types.SetAssetOwner:
		var asset UpdateAssetOwner
		err := rlp.DecodeBytes(action.Data(), &asset)
		if err != nil {
			return nil, err
		}
		acct, err := am.GetAccountByName(asset.Owner)
		if err != nil {
			return nil, err
		}
		if acct == nil {
			return nil, ErrAccountNotExist
		}

		// check owner
		if err := am.ast.CheckOwner(action.Sender(), asset.AssetID); err != nil {
			return nil, err
		}

		if err := am.ast.SetAssetNewOwner(action.Sender(), asset.AssetID, asset.Owner); err != nil {
			return nil, err
		}
	case types.UpdateAssetContract:
		var assetContract UpdateAssetContract
		err := rlp.DecodeBytes(action.Data(), &assetContract)
		if err != nil {
			return nil, err
		}

		if len(assetContract.Contract) != 0 {
			acct, err := am.GetAccountByName(assetContract.Contract)
			if err != nil {
				return nil, err
			}
			if acct == nil {
				return nil, ErrAccountNotExist
			}
		}

		if err := am.ast.CheckOwner(action.Sender(), assetContract.AssetID); err != nil {
			return nil, err
		}

		if err := am.ast.SetAssetNewContract(assetContract.AssetID, assetContract.Contract); err != nil {
			return nil, err
		}

	case types.Transfer:
	default:
		return nil, ErrUnKnownTxType
	}

	return internalActions, nil
}
