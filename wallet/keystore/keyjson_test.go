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
	"crypto/ecdsa"
	crand "crypto/rand"
	"testing"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/stretchr/testify/assert"
)

func TestEncryptAndDecryptKey(t *testing.T) {
	keyj := new(keyJSON)

	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), crand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	key := &Key{
		Addr:       crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}

	err = keyj.encryptKey(key, "pwd", StandardScryptN, StandardScryptP)
	if err != nil {
		t.Fatal(err)
	}

	newk, err := keyj.decryptKey("pwd")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, newk, key)
}
