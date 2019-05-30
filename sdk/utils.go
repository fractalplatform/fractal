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

package sdk

import (
	"crypto/ecdsa"
	"math/rand"
	"time"

	"github.com/fractalplatform/fractal/crypto"
)

var (
	availableChars = "abcdefghijklmnopqrstuvwxyz0123456789"
)

// GenerateAccountName generate account name
func GenerateAccountName(namePrefix string, addStrLen int) string {
	newRandomName := namePrefix
	size := len(availableChars)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < addStrLen; i++ {
		index := rand.Intn(10000) % size
		newRandomName += string(availableChars[index])
	}
	return newRandomName
}

// GenerateKey generate pubkey and privkey
func GenerateKey() *ecdsa.PrivateKey {
	prikey, _ := crypto.GenerateKey()
	return prikey
}
