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
	"math/big"
	"regexp"
	"strings"
)

// Name represents the account name
type Name string

// StrToName  returns Name with string of s.
func StrToName(s string) Name {
	return Name(s)
}

// BytesToName returns Name with value b.
func BytesToName(b []byte) Name {
	return StrToName(string(b))
}

// BigToName returns Name with byte values of b.
func BigToName(b *big.Int) Name {
	return BytesToName(b.Bytes())
}

// Big converts a name to a big integer.
func (n Name) Big() *big.Int {
	return new(big.Int).SetBytes(n.Bytes())
}

// Bytes converts a name to bytes.
func (n Name) Bytes() []byte {
	return []byte(n.String())
}

// String converts a name to string.
func (n Name) String() string {
	return string(n)
}

// SetString  sets the name to the value of b..
func (n *Name) SetString(s string) {
	*n = Name(s)
}

// IsValid verifies whether a string can represent a valid name or not.
func (n Name) IsValid(reg *regexp.Regexp, length uint64) bool {
	if uint64(len(n.String())) > length {
		return false
	}

	return reg.MatchString(n.String())
}

// IsChildren name children
func (n Name) IsChildren(name Name) bool {
	if strings.HasPrefix(name.String(), n.String()) {
		if len(name.String()) > len(n.String()) && name.String()[len(n.String())] == '.' {
			return true
		}
	}
	return false
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
	*n = StrToName(input)
	return nil
}

// FindStringSubmatch return segment
func FindStringSubmatch(reg *regexp.Regexp, name string) (ret []string) {
	list := reg.FindStringSubmatch(name)
	for i := 1; i < len(list); i++ {
		if len(list[i]) == 0 {
			continue
		}
		ret = append(ret, list[i])
	}
	return
}
