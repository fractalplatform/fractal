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

package wallet

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/blockchain"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/wallet/cache"
	"github.com/fractalplatform/fractal/wallet/keystore"
)

// Wallet represents a software wallet.
type Wallet struct {
	accounts        cache.Accounts
	cache           *cache.AccountCache
	ks              *keystore.KeyStore
	bindingFilePath string
	blockchain      *blockchain.BlockChain
}

// NewWallet creates a wallet to sign transaction.
func NewWallet(keyStoredir string, scryptN, scryptP int) *Wallet {
	log.Info("Disk storage enabled for keystore", "dir", keyStoredir)
	w := &Wallet{
		cache: cache.NewAccountCache(keyStoredir),
		ks:    &keystore.KeyStore{DirPath: keyStoredir, ScryptN: scryptN, ScryptP: scryptP},
	}
	w.bindingFilePath = w.ks.JoinPath("acountKeyBindingInfo.txt")
	w.createFileIfNotExist(w.bindingFilePath)
	return w
}

func (w *Wallet) createFileIfNotExist(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		_, err = os.Create(filePath)
		if err != nil {
			log.Error("Create file fail:", "err=", err, "file=", filePath)
			return err
		}
		log.Info("Create file success:", "file=", filePath)
		return nil
	}
	return nil
}
func (w *Wallet) SetBlockChain(blockchain *blockchain.BlockChain) {
	w.blockchain = blockchain
}
func (w *Wallet) GetAccountManager() (*am.AccountManager, error) {
	statedb, err := w.blockchain.State()
	if err != nil {
		return nil, err
	}
	return am.NewAccountManager(statedb)
}

// NewAccount generates a new key and stores it into the key directory.
func (w *Wallet) NewAccount(passphrase string) (cache.Account, error) {
	key, err := keystore.NewKey(crand.Reader)
	if err != nil {
		return cache.Account{}, err
	}
	publicKey := hexutil.Bytes(crypto.FromECDSAPub(&key.PrivateKey.PublicKey)).String()
	a := cache.Account{Addr: key.Addr,
		Path:      w.ks.JoinPath(keyFileName(key.Addr)),
		PublicKey: publicKey}

	if err := w.ks.StoreKey(key, a.Path, passphrase); err != nil {
		return cache.Account{}, err
	}
	w.cache.Add(a)
	return a, nil
}

// Delete deletes a account by passsphrase.
func (w *Wallet) Delete(a cache.Account, passphrase string) error {
	a, _, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}

	if err := os.Remove(a.Path); err != nil {
		return err
	}
	w.cache.Delete(a.Addr)
	return nil
}

// Update changes the passphrase of an existing account.
func (w *Wallet) Update(a cache.Account, passphrase, newPassphrase string) error {
	a, key, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}
	return w.ks.StoreKey(key, a.Path, newPassphrase)
}

// Export exports as a JSON key, encrypted with newPassphrase.
func (w *Wallet) Export(a cache.Account, passphrase, newPassphrase string) (keyJSON []byte, err error) {
	_, key, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	N, P := keystore.StandardScryptN, keystore.StandardScryptP
	return keystore.EncryptKey(key, newPassphrase, N, P)
}

// Import stores the given encrypted JSON key into the key directory.
func (w *Wallet) Import(keyJSON []byte, passphrase, newPassphrase string) (cache.Account, error) {
	key, err := keystore.DecryptKey(keyJSON, passphrase)
	if err != nil {
		return cache.Account{}, err
	}
	return w.importKey(key, newPassphrase)

}

// ImportECDSA stores the given key into the key directory, encrypting it with the passphrase.
func (w *Wallet) ImportECDSA(priv *ecdsa.PrivateKey, passphrase string) (cache.Account, error) {
	key := &keystore.Key{
		Addr:       crypto.PubkeyToAddress(priv.PublicKey),
		PrivateKey: priv,
	}
	if w.cache.Has(key.Addr) {
		return cache.Account{}, ErrAccountExists
	}
	return w.importKey(key, passphrase)
}

// HasAddress reports whether a key with the given address is present.
func (w *Wallet) HasAddress(addr common.Address) bool {
	return w.cache.Has(addr)
}

// Accounts returns all key files
func (w *Wallet) Accounts() cache.Accounts {
	return w.cache.Accounts()
}

// Find resolves the given account into a unique entry in the keystore.
func (w *Wallet) Find(addr common.Address) (cache.Account, error) {
	account := w.cache.Find(addr)
	if account != nil {
		return *account, nil
	}
	return cache.Account{}, ErrNoMatch
}

// SignHashWithPassphrase signs hash if the private key matching the given address
// can be decrypted with the given passphrase.
func (w *Wallet) SignHashWithPassphrase(a cache.Account, passphrase string, hash []byte) (signature []byte, err error) {
	_, key, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(hash, key.PrivateKey)
}

// SignTxWithPassphrase signs the Action if the private key matching the given address
// can be decrypted with the given passphrase.
func (w *Wallet) SignTxWithPassphrase(a cache.Account, passphrase string, tx *types.Transaction, action *types.Action, chainID *big.Int) (*types.Transaction, error) {
	_, key, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	if err := types.SignAction(action, tx, types.NewSigner(chainID), key.PrivateKey); err != nil {
		return nil, err
	}
	return tx, nil
}

func (w *Wallet) importKey(key *keystore.Key, passphrase string) (cache.Account, error) {
	a := cache.Account{Addr: key.Addr, Path: w.ks.JoinPath(keyFileName(key.Addr))}
	if err := w.ks.StoreKey(key, a.Path, passphrase); err != nil {
		return cache.Account{}, err
	}
	w.cache.Add(a)
	return a, nil
}
func (w *Wallet) GetPrivateKey(a cache.Account, passphrase string) (*keystore.Key, error) {
	_, key, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	return key, nil
}
func (w *Wallet) getDecryptedKey(a cache.Account, passphrase string) (cache.Account, *keystore.Key, error) {
	a, err := w.Find(a.Addr)
	if err != nil {
		return a, nil, err
	}
	key, err := w.ks.GetKey(a.Addr, a.Path, passphrase)
	return a, key, err
}

// keyFileName implements the naming convention for keyfiles:
// UTC--<created_at UTC ISO8601>-<address hex>
func keyFileName(keyAddr common.Address) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}

// BindAccountNameAddr bind the account name and publicKey, the current publicKey is got from account manager.
func (w *Wallet) BindAccountAndPublicKey(accountName string) error {
	publicKeyAccountsMap := w.getBindingInfo()

	// get account info firstly
	accountMgr, err := w.GetAccountManager()
	if err != nil {
		return err
	}
	account, err := accountMgr.GetAccountByName(common.Name(accountName))
	if err != nil {
		return err
	}
	curPublicKey := account.PublicKey.String()

	// We need consider 5 situations below:
	// 1:publicKey exist, account NOT exist
	// 2:publicKey NOT exist, account exist
	// 3:publicKey AND account exist，but NOT matched
	// 4:publicKey AND account exist，and matched
	// 5:publicKey AND account both NOT exist

	// below loop can resolve:
	// (1) when account exist, if 4, return, if 2 and 3, it can dismatch them, but we need match new publicKey then
	// (2) when account NOT exist, copy original info
	tmpPublicKeyAccountsMap := make(map[string][]string)
	for publicKey, accounts := range publicKeyAccountsMap {
		tmpAccounts := make([]string, 0)
		for _, account := range accounts {
			if account != accountName {
				tmpAccounts = append(tmpAccounts, account)
				continue
			} else if curPublicKey == publicKey {
				return nil
			}
		}
		if len(tmpAccounts) > 0 {
			tmpPublicKeyAccountsMap[publicKey] = tmpAccounts
		}
	}
	if _, ok := tmpPublicKeyAccountsMap[curPublicKey]; ok {
		tmpPublicKeyAccountsMap[curPublicKey] = append(tmpPublicKeyAccountsMap[curPublicKey], accountName)
	} else {
		accounts := make([]string, 0)
		accounts = append(accounts, accountName)
		tmpPublicKeyAccountsMap[curPublicKey] = accounts
	}

	return w.writeBindingInfo(tmpPublicKeyAccountsMap)
}

// ps: in this func, we can't use account's publicKey to get accounts info,
// because the account's publicKey has been updated by block chain.
func (w *Wallet) DeleteBound(accountName string) error {
	publicKeyAccountsMap := w.getBindingInfo()

	tmpPublicKeyAccountsMap := make(map[string][]string)
	for publicKey, accounts := range publicKeyAccountsMap {
		tmpAccounts := make([]string, 0)
		for _, account := range accounts {
			if account != accountName {
				tmpAccounts = append(tmpAccounts, account)
				continue
			}
		}
		if len(tmpAccounts) > 0 {
			tmpPublicKeyAccountsMap[publicKey] = tmpAccounts
		}
	}
	return w.writeBindingInfo(tmpPublicKeyAccountsMap)
}

func (w *Wallet) GetAccountsByPublicKey(publicKey string) ([]am.Account, error) {
	publicKeyAccountsMap := w.getBindingInfo()
	accounts := make([]am.Account, 0)
	if accountNames, ok := publicKeyAccountsMap[publicKey]; ok {
		accountMgr, err := w.GetAccountManager()
		if err != nil {
			return nil, err
		}
		for _, accountName := range accountNames {
			account, err := accountMgr.GetAccountByName(common.Name(accountName))
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, *account)
		}
	}
	return accounts, nil
}

func (w *Wallet) GetAllAccounts() ([]am.Account, error) {
	publicKeyAccountsMap := w.getBindingInfo()
	accountMgr, err := w.GetAccountManager()
	if err != nil {
		return nil, err
	}
	accounts := make([]am.Account, 0)
	for _, accountNames := range publicKeyAccountsMap {
		for _, accountName := range accountNames {
			account, err := accountMgr.GetAccountByName(common.Name(accountName))
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, *account)
		}
	}
	return accounts, nil
}

func (w *Wallet) getBindingInfo() map[string][]string {
	publicKeyAccountsMap := make(map[string][]string)
	fileContent, _ := ioutil.ReadFile(w.bindingFilePath)
	if len(fileContent) > 0 {
		json.Unmarshal(fileContent, &publicKeyAccountsMap)
	}
	log.Debug("getBindingInfo:", "binging info", publicKeyAccountsMap)
	return publicKeyAccountsMap
}

func (w *Wallet) writeBindingInfo(addrAccountsMap map[string][]string) error {
	log.Debug("writeBindingInfo:", "binging info", addrAccountsMap)
	fileContent, err := json.Marshal(addrAccountsMap)
	if err != nil {
		log.Error("fail to marshall map to json string:", addrAccountsMap)
		return err
	}
	if ioutil.WriteFile(w.bindingFilePath, fileContent, 0666) == nil {
		log.Info("success to write binding info:", string(fileContent))
		return nil
	} else {
		log.Error("fail to write binding info:", string(fileContent))
		return errors.New("fail to write binding info")
	}
}
