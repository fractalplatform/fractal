package plugin

import (
	"encoding/binary"
	"errors"
	"math/big"
	"strconv"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	acctManagerName = "sysAccount"
	acctInfoPrefix  = "acctInfo"
	accountIDPrefix = "accountId"
	counterPrefix   = "accountCounter"
	counterID       = uint64(4096)
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
	AccountID   uint64
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
	am := &AccountManager{db}
	am.initAccountCounter()
	return am, nil
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
func (am *AccountManager) CreateAccount(pubKey common.PubKey, description string) ([]byte, error) {
	//var acct CreateAccountAction
	//err := rlp.DecodeBytes(action.Data(), &acct)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if !common.IsHexPubKey(acct.PublicKey) {
	//	return nil, ErrInvalidPubKey
	//}
	//
	//pubKey := common.HexToPubKey(acct.PublicKey)
	tempKey, err := crypto.UnmarshalPubkey(pubKey.Bytes())
	if err != nil {
		return nil, err
	}

	newAddress := crypto.PubkeyToAddress(*tempKey)

	_, err = am.getAccount(newAddress)
	if err == nil {
		return nil, ErrAccountIsExist
	} else if err != ErrAccountNotExist {
		return nil, err
	}

	accountCounter, err := am.getAccountCounter()
	if err != nil {
		return nil, err
	}

	accountCounter += 1

	acctObject := Account{
		Address:     newAddress,
		AccountID:   accountCounter,
		Nonce:       0,
		Code:        make([]byte, 0),
		CodeHash:    crypto.Keccak256Hash(nil),
		CodeSize:    0,
		Balances:    nil,
		Suicide:     false,
		Destroy:     false,
		Description: description,
	}

	if err = am.setAccount(&acctObject); err != nil {
		return nil, err
	}

	aid, err := rlp.EncodeToBytes(&accountCounter)
	if err != nil {
		return nil, err
	}

	address, err := rlp.EncodeToBytes(&acctObject.Address)
	if err != nil {
		return nil, err
	}

	am.sdb.Put(acctManagerName, accountIDPrefix+strconv.FormatUint(accountCounter, 10), address)
	am.sdb.Put(acctManagerName, counterPrefix, aid)

	return newAddress.Bytes(), nil
}

// IssueAsset
// Pares Payload to issue a asset
func (am *AccountManager) IssueAsset(accountAddress common.Address, assetName string, symbol string, amount *big.Int, dec uint64, founder common.Address, owner common.Address, limit *big.Int, description string, asm IAsset) ([]byte, error) {
	//var issueAsset IssueAsset
	//err := rlp.DecodeBytes(action.Data(), &issueAsset)
	//if err != nil {
	//	return nil, err
	//}
	issueAsset := &IssueAsset{
		AssetName:   assetName,
		Symbol:      symbol,
		Amount:      amount,
		Decimals:    dec,
		Founder:     founder,
		Owner:       owner,
		UpperLimit:  limit,
		Description: description,
	}

	err := asm.CheckIssueAssetInfo(accountAddress, issueAsset)
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

	assetID, err := asm.IssueAssetForAccount(issueAsset.AssetName, issueAsset.Symbol, issueAsset.Amount, issueAsset.Decimals, issueAsset.Founder, issueAsset.Owner, issueAsset.UpperLimit, issueAsset.Description)
	if err != nil {
		return nil, err
	}
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(assetID))

	return buf, nil
}

func (am *AccountManager) CanTransfer(accountAddress common.Address, assetID uint64, value *big.Int) (bool, error) {
	if assetID != SystemAssetID {
		return false, ErrAssetIDInvalid
	}

	if value.Cmp(big.NewInt(0)) < 0 {
		return false, ErrAmountValueInvalid
	}

	val, err := am.GetBalanceByAddress(accountAddress, assetID)
	if err != nil {
		return false, err
	}

	if val.Cmp(value) < 0 {
		return false, ErrInsufficientBalance
	}

	return true, nil
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

	if err = am.subBalanceByID(fromAcct, value); err != nil {
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

// RecoverTx
// Make sure the transaction is signed properly and validate account authorization.
func (am *AccountManager) RecoverTx(signer types.Signer, tx *types.Transaction) error {
	for _, action := range tx.GetActions() {
		pubs, err := types.RecoverMultiKey(signer, action, tx)
		if err != nil {
			return err
		}

		tempKey, err := crypto.UnmarshalPubkey(pubs[0].Bytes())
		if err != nil {
			return err
		}
		tempAddress := crypto.PubkeyToAddress(*tempKey)

		account, err := am.getAccount(action.Sender())
		if err != nil {
			return err
		}

		if tempAddress.Compare(account.Address) != 0 {
			return err
		}
	}

	return nil
}

func (am *AccountManager) GetNonce(accountAddress common.Address) (uint64, error) {
	account, err := am.getAccount(accountAddress)
	if err != nil {
		return 0, err
	}

	return account.Nonce, nil
}

func (am *AccountManager) SetNonce(accountAddress common.Address, nonce uint64) error {
	account, err := am.getAccount(accountAddress)
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

func (am *AccountManager) GetAccount(accountAddress common.Address) (*Account, error) {
	return am.getAccount(accountAddress)
}

func (am *AccountManager) AccountHaveCode(accountAddress common.Address) (bool, error) {
	account, err := am.getAccount(accountAddress)
	if err != nil {
		return false, err
	}

	if account.CodeSize == 0 {
		return false, nil
	} else {
		return true, nil
	}

}

func (am *AccountManager) GetCode(accountAddress common.Address) ([]byte, error) {
	account, err := am.getAccount(accountAddress)
	if err != nil {
		return nil, err
	}

	if account.CodeSize == 0 || account.Suicide {
		return nil, ErrCodeIsEmpty
	}

	return account.Code, nil
}

func (am *AccountManager) SetCode(accountAddress common.Address, code []byte) (bool, error) {
	account, err := am.getAccount(accountAddress)
	if err != nil {
		return false, err
	}

	if len(code) == 0 {
		return false, ErrCodeIsEmpty
	}
	account.Code = code
	account.CodeHash = crypto.Keccak256Hash(code)
	account.CodeSize = uint64(len(code))

	err = am.setAccount(account)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (am *AccountManager) GetBalanceByAddress(accountAddress common.Address, assetID uint64) (*big.Int, error) {
	if assetID != SystemAssetID {
		return big.NewInt(0), ErrAssetIDInvalid
	}

	account, err := am.getAccount(accountAddress)
	if err != nil {
		return big.NewInt(0), err
	}

	if account.Balances == nil {
		return big.NewInt(0), ErrAccountAssetNotExist
	}

	return account.Balances.Balance, nil
}

func (am *AccountManager) DeleteAccount(accountAddress common.Address) error {
	account, err := am.getAccount(accountAddress)
	if err != nil {
		return err
	}

	account.Destroy = true

	if err = am.setAccount(account); err != nil {
		return err
	}
	return nil
}

func (am *AccountManager) GetBalanceByID(accountID, assetID uint64) *big.Int {
	if assetID != SystemAssetID {
		return big.NewInt(0)
	}

	account, err := am.getAccountByID(accountID)
	if err != nil {
		return big.NewInt(0)
	}

	return account.Balances.Balance
}

func (am *AccountManager) GetAccountID(accountAddress string) uint64 {
	account, err := am.getAccount(common.HexToAddress(accountAddress))
	if err != nil {
		return 0
	}
	return account.AccountID
}

func (am *AccountManager) GetCodeSizeByID(accountID uint64) uint64 {
	account, err := am.getAccountByID(accountID)
	if err != nil {
		return 0
	}
	return account.CodeSize
}

func (am *AccountManager) GetCodeById(accountID uint64) []byte {
	account, err := am.getAccountByID(accountID)
	if err != nil {
		return nil
	}

	if account.CodeSize == 0 {
		return nil
	}

	return account.Code
}

func (am *AccountManager) GetAccountAddressByID(accountID uint64) string {
	if accountID == 0 {
		return ""
	}

	b, err := am.sdb.Get(acctManagerName, accountIDPrefix+strconv.FormatUint(accountID, 10))
	if err != nil {
		return ""
	}

	if len(b) == 0 {
		return ""
	}

	var address common.Address
	if err = rlp.DecodeBytes(b, &address); err != nil {
		return ""
	}

	return address.String()
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

func (am *AccountManager) getAccountByID(accountID uint64) (*Account, error) {
	if accountID == 0 {
		return nil, ErrAccountIDInvalid
	}

	b, err := am.sdb.Get(acctManagerName, accountIDPrefix+strconv.FormatUint(accountID, 10))
	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		//log.Debug("account not exist", "id", ErrAccountNotExist, accountID)
		return nil, ErrAccountNotExist
	}

	var address common.Address
	if err = rlp.DecodeBytes(b, &address); err != nil {
		return nil, err
	}

	return am.getAccount(address)
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
	ErrCounterNotExist        = errors.New("account global counter not exist")
	ErrAccountIDInvalid       = errors.New("account id invalid")
)
