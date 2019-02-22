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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"io"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/fractalplatform/fractal/crypto"
	"golang.org/x/crypto/scrypt"
)

const (
	defaultCipher = "aes-128-ctr"
	defaultKDF    = "scrypt"
)

type keyJSON struct {
	Address    string `json:"address"`
	PublicKey  string `json:"publickey"`
	Cipher     string `json:"cipher"`
	CipherText string `json:"ciphertext"`
	CipherIV   string `json:"cipheriv"`
	KDF        kdf    `json:"kdf"`
	MAC        string `json:"mac"`
}

type kdf struct {
	Type   string `json:"type"`
	KeyLen int    `json:"keylen"`
	N      int    `json:"N"`
	R      int    `json:"R"`
	P      int    `json:"P"`
	Salt   string `json:"salt"`
}

func (k *kdf) getKey(passphrase string) ([]byte, error) {
	salt, err := hex.DecodeString(k.Salt)
	if err != nil {
		return nil, err
	}
	return scrypt.Key([]byte(passphrase), salt, k.N, k.R, k.P, k.KeyLen)
}

func (kj *keyJSON) encryptKey(key *Key, passphrase string, scryptN, scryptP int) error {
	keyBytes := math.PaddedBigBytes(key.PrivateKey.D, 32)

	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	derivedKey, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return err
	}

	encryptKey := derivedKey[:16]

	iv := make([]byte, aes.BlockSize) // 16
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	cipherText, err := aesCTRXOR(encryptKey, keyBytes, iv)
	if err != nil {
		return err
	}

	mac := crypto.Keccak256(derivedKey[16:32], cipherText)

	kj.Address = key.Addr.Hex()
	kj.PublicKey = hexutil.Bytes(crypto.FromECDSAPub(&key.PrivateKey.PublicKey)).String()
	kj.Cipher = defaultCipher
	kj.CipherIV = hex.EncodeToString(iv)
	kj.CipherText = hex.EncodeToString(cipherText)
	kj.MAC = hex.EncodeToString(mac)
	kj.KDF.Type = defaultKDF
	kj.KDF.KeyLen = scryptDKLen
	kj.KDF.N = scryptN
	kj.KDF.R = scryptR
	kj.KDF.P = scryptP
	kj.KDF.Salt = hex.EncodeToString(salt)

	return nil
}

func (kj *keyJSON) decryptKey(passphrase string) (*Key, error) {
	mac, err := hex.DecodeString(kj.MAC)
	if err != nil {
		return nil, err
	}

	iv, err := hex.DecodeString(kj.CipherIV)
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(kj.CipherText)
	if err != nil {
		return nil, err
	}

	derivedKey, err := kj.KDF.getKey(passphrase)
	if err != nil {
		return nil, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, ErrDecrypt
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, err
	}

	key := crypto.ToECDSAUnsafe(plainText)
	return &Key{
		Addr:       crypto.PubkeyToAddress(key.PublicKey),
		PrivateKey: key,
	}, nil

}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}
