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

package common

import (
	"bytes"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// PubKeyLength of PubKey in bytes
const PubKeyLength = 65

var pubKeyT = reflect.TypeOf(PubKey{})

// HexToPubKey returns PubKey with byte values of s.
func HexToPubKey(s string) PubKey { return BytesToPubKey(FromHex(s)) }

// BytesToPubKey returns PubKey with value b.
func BytesToPubKey(b []byte) PubKey {
	var a PubKey
	a.SetBytes(b)
	return a
}

// IsHexPubKey verifies whether a string can represent a valid hex-encoded
//  PubKey or not.
func IsHexPubKey(s string) bool {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*(PubKeyLength) && isHex(s)
}

// PubKey represents the 64 byte public key
type PubKey [PubKeyLength]byte

//Bytes return bytes
func (p PubKey) Bytes() []byte { return p[:] }

// Big converts a hash to a big integer.
func (p PubKey) Big() *big.Int { return new(big.Int).SetBytes(p[:]) }

// Hex converts a hash to a hex string.
func (p PubKey) Hex() string { return hexutil.Encode(p[:]) }

//SetBytes set bytes to pubkey
func (p *PubKey) SetBytes(key []byte) {
	if len(key) > len(p) {
		key = key[len(key)-PubKeyLength:]
	}
	copy(p[PubKeyLength-len(key):], key)
}

// String implements fmt.Stringer.
func (p PubKey) String() string {
	return p.Hex()
}

// MarshalText returns the hex representation of a.
func (p PubKey) MarshalText() ([]byte, error) {
	return hexutil.Bytes(p[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (p *PubKey) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("PubKey", input, p[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (p *PubKey) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(pubKeyT, input, p[:])
}

// Compare returns an integer comparing two byte slices lexicographically.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
// A nil argument is equivalent to an empty slice.
func (p PubKey) Compare(x PubKey) int {
	return bytes.Compare(p.Bytes(), x.Bytes())
}
