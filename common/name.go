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
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// Name represents the account name
type Name string

var accountNameCheck *regexp.Regexp
var assetNameCheck *regexp.Regexp

var (
	AccountNameLevel  uint64 = 1
	AccountNameLen    uint64 = 16
	SubAccountNameLen uint64 = 8
	AssetNameLevel    uint64 = 1
	AssetNameLen      uint64 = 16
	SubAssetNameLen   uint64 = 8
)

//Name type for fee withdraw
const (
	AssetName    = 0
	ContractName = 1
	CoinbaseName = 2
)

func init() {
	SetAccountNameCheckRule(AccountNameLevel, AccountNameLen, SubAccountNameLen)
	SetAssetNameCheckRule(AssetNameLevel, AssetNameLen, SubAssetNameLen)
}

func SetAccountNameCheckRule(nameLevel, nameLen, subNameLen uint64) {
	var nameCheck string
	if nameLevel == 0 {
		nameCheck = fmt.Sprintf("^[a-z0-9]{7,%d}$", nameLen)
	} else {
		nameCheck = fmt.Sprintf("^[a-z0-9]{7,%d}(\\.[a-z0-9]{1,%d}){0,%d}$", nameLen, subNameLen, nameLevel)
	}
	log.Info("Account name level", "level", nameLevel, "name length", nameLen, "sub name length", subNameLen)
	accountNameCheck = regexp.MustCompile(nameCheck)
}

func SetAssetNameCheckRule(nameLevel, nameLen, subNameLen uint64) {
	var nameCheck string
	if nameLevel == 0 {
		nameCheck = fmt.Sprintf("^[a-z0-9]{2,%d}$", nameLen)
	} else {
		nameCheck = fmt.Sprintf("^[a-z0-9]{2,%d}(\\.[a-z0-9]{1,%d}){0,%d}$", nameLen, subNameLen, nameLevel)
	}
	log.Info("Asset name level", "level", nameLevel, "name length", nameLen, "sub name length", subNameLen)
	assetNameCheck = regexp.MustCompile(nameCheck)
}

// IsValidAccountName verifies whether a string can represent a valid name or not.
func IsValidAccountName(s string) bool {
	if accountNameCheck == nil {
		return false
	}
	return accountNameCheck.MatchString(s)
}

func IsValidAssetName(s string) bool {
	if assetNameCheck == nil {
		return false
	}
	return assetNameCheck.MatchString(s)
}

// StrToName  returns Name with string of s.
func StrToName(s string) Name {
	n, err := parseName(s)
	if err != nil {
		panic(err)
	}
	return n
}

// StrToName  returns Name with string of s.
func StringToName(s string) (Name, error) {
	var n Name
	if !n.SetString(s) {
		return n, fmt.Errorf("invalid name %v", s)
	}
	return n, nil
}

func parseName(s string) (Name, error) {
	var n Name
	if !n.SetString(s) {
		return n, fmt.Errorf("invalid name %v", s)
	}
	return n, nil
}

// BytesToName returns Name with value b.
func BytesToName(b []byte) (Name, error) {
	return parseName(string(b))
}

// BigToName returns Name with byte values of b.
func BigToName(b *big.Int) (Name, error) { return BytesToName(b.Bytes()) }

// SetString  sets the name to the value of b..
func (n *Name) SetString(s string) bool {
	if !IsValidAccountName(s) {
		return false
	}
	*n = Name(s)
	return true
}

// UnmarshalText parses a hash in hex syntax.
func (n *Name) UnmarshalText(input []byte) error {
	return n.UnmarshalJSON(input)
}

// UnmarshalJSON parses a hash in hex syntax.
func (n *Name) UnmarshalJSON(data []byte) error {
	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}
	if len(input) > 0 {
		dec, err := parseName(string(input))
		if err != nil {
			return err
		}
		*n = dec
	}
	return nil
}

// EncodeRLP implements rlp.Encoder
func (n Name) EncodeRLP(w io.Writer) error {
	str := n.String()
	if len(str) != 0 {
		if _, err := parseName(str); err != nil {
			return err
		}
	}
	rlp.Encode(w, str)
	return nil
}

// DecodeRLP implements rlp.Decoder
func (n *Name) DecodeRLP(s *rlp.Stream) error {
	var str string
	err := s.Decode(&str)
	if err == nil {
		if len(str) != 0 {
			name, err := parseName(str)
			if err != nil {
				return err
			}
			*n = name
		}
	}
	return err
}

// String implements fmt.Stringer.
func (n Name) String() string {
	return string(n)
}

// Big converts a name to a big integer.
func (n Name) Big() *big.Int { return new(big.Int).SetBytes([]byte(n.String())) }

func (n Name) IsValidCreator(name string) bool {

	creator := n.String()
	current := name
	trimName := strings.TrimPrefix(current, creator)

	if strings.Index(trimName, ".") != 0 {
		return false
	}

	return true
}

func (n Name) AccountNameLevel() int {
	return len(strings.Split(n.String(), "."))
}

func SplitString(s string) []string {
	return strings.Split(s, ".")
}

func IsValidCreator(name string, subName string) bool {

	creator := name
	current := subName
	trimName := strings.TrimPrefix(current, creator)

	if strings.Index(trimName, ".") != 0 {
		return false
	}

	return true
}
