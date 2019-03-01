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

package keystore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/fractalplatform/fractal/common"
)

const (
	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = 1 << 18

	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptP = 1

	// LightScryptN is the N parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptN = 1 << 12

	// LightScryptP is the P parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptP = 6

	scryptR     = 8
	scryptDKLen = 32
)

// KeyStore manages a key storage directory on disk.
type KeyStore struct {
	DirPath          string
	ScryptN, ScryptP int
}

// GetKey load the key from the key file and decrypt it.
func (ks *KeyStore) GetKey(addr common.Address, filename, passphrase string) (*Key, error) {
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	key, err := DecryptKey(keyjson, passphrase)
	if err != nil {
		return nil, err
	}
	// check address
	if key.Addr != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Addr, addr)
	}
	return key, nil
}

// StoreKey encrypts with 'passphrase' and stores in the given directory
func (ks *KeyStore) StoreKey(key *Key, filename, passphrase string) error {
	keyjson, err := EncryptKey(key, passphrase, ks.ScryptN, ks.ScryptP)
	if err != nil {
		return err
	}
	return writeKeyFile(filename, keyjson)
}

func (ks *KeyStore) GetPublicKey(filename string) (string, error) {
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	keyj := new(keyJSON)
	if err := json.Unmarshal(keyjson, keyj); err != nil {
		return "", err
	}

	return keyj.PublicKey, nil
}

// JoinPath join file name into key dir path.
func (ks *KeyStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.DirPath, filename)
}

// DecryptKey decrypts a key from a json bytes.
func DecryptKey(keyjson []byte, passphrase string) (*Key, error) {
	keyj := new(keyJSON)
	if err := json.Unmarshal(keyjson, keyj); err != nil {
		return nil, err
	}
	return keyj.decryptKey(passphrase)
}

// EncryptKey encrypts a key using the specified scrypt parameters into a json bytes.
func EncryptKey(key *Key, passphrase string, scryptN, scryptP int) ([]byte, error) {
	keyj := new(keyJSON)
	if err := keyj.encryptKey(key, passphrase, scryptN, scryptP); err != nil {
		return nil, err
	}
	return json.Marshal(keyj)
}
