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
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/wallet/cache"
	"github.com/fractalplatform/fractal/wallet/keystore"
	"bufio"
	"io"
	"strings"
	"errors"
	"io/ioutil"
	"encoding/json"
)

// Wallet represents a software wallet.
type Wallet struct {
	accounts cache.Accounts
	cache    *cache.AccountCache
	ks       *keystore.KeyStore
	bindingFilePath string
}

// NewWallet creates a wallet to sign transaction.
func NewWallet(keyStoredir string, scryptN, scryptP int) *Wallet {
	log.Info("Disk storage enabled for keystore", "dir", keyStoredir)
	w := &Wallet{
		cache: cache.NewAccountCache(keyStoredir),
		ks:    &keystore.KeyStore{DirPath: keyStoredir, ScryptN: scryptN, ScryptP: scryptP},
	}
	w.bindingFilePath = w.ks.JoinPath("acountAddrBindingInfo.txt")
	_, err := os.Stat(w.bindingFilePath)
	if err == nil {
		os.Create(w.bindingFilePath)
	}
	return w
}

// NewAccount generates a new key and stores it into the key directory.
func (w *Wallet) NewAccount(passphrase string) (cache.Account, error) {
	key, err := keystore.NewKey(crand.Reader)
	if err != nil {
		return cache.Account{}, err
	}

	a := cache.Account{Addr: key.Addr, Path: w.ks.JoinPath(keyFileName(key.Addr))}

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

// BindAccountNameAddr bind the account name and address,
// if account name has been bound to another address, it will fail,
// and you should use UpdateBindingAddr func to bind new address.
func (w *Wallet) BindAccountNameAddr(a cache.Account, passphrase string, accountName string) error {
	a, _, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}

	addrAccountsMap := make(map[string][]string)
	fileContent, err := ioutil.ReadFile(w.bindingFilePath)
	if len(fileContent) > 0 {
		json.Unmarshal(fileContent, &addrAccountsMap)
	}
	for _, accounts := range addrAccountsMap {
		for _, account := range accounts {
			if account == accountName {
				return errors.New("Account has been bound to another address.")
			}
		}
	}
	addrStr := a.Addr.String()
	if _, ok := addrAccountsMap[addrStr]; ok {
		addrAccountsMap[addrStr] = append(addrAccountsMap[addrStr], accountName)
	} else {
		accounts := make([]string, 1)
		accounts = append(accounts, accountName)
		addrAccountsMap[addrStr] = accounts
	}

	if fileObj,err := os.OpenFile(w.bindingFilePath, os.O_RDWR|os.O_CREATE,0644); err == nil {
		defer fileObj.Close()
		fileContent, err = json.Marshal(addrAccountsMap)
		if ioutil.WriteFile(w.bindingFilePath, fileContent,0666) == nil {
			log.Info("写入文件成功:", string(content))
		}
	}
	return nil
}

func (w *Wallet) DeleteBound(a cache.Account, passphrase string, accountName string) error {
	a, _, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wallet) UpdateBindingAddr(a cache.Account, passphrase string, accountName string, newAddress string) error {
	a, _, err := w.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wallet) GetAccountNameByAddr(address string) string {

	return nil
}

func (w *Wallet) BatchGetAccountName(address []string) []string {

	return nil
}