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
	"encoding/json"
	"testing"

	"github.com/fractalplatform/fractal/utils/rlp"
	"github.com/stretchr/testify/assert"
)

func TestAuthorEncodeAndDecodeRLP(t *testing.T) {
	var tests = []struct {
		inputAuthor *Author
	}{
		{&Author{Owner: Name("test"), Weight: 1}},
		{&Author{Owner: HexToPubKey("test"), Weight: 1}},
		{&Author{Owner: HexToAddress("test"), Weight: 1}},
	}
	for _, test := range tests {
		authorBytes, err := rlp.EncodeToBytes(test.inputAuthor)
		if err != nil {
			t.Fatal(err)
		}
		outputAuthor := &Author{}
		if err := rlp.Decode(bytes.NewReader(authorBytes), outputAuthor); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, test.inputAuthor, outputAuthor)
	}
}

func TestAuthorMarshalAndUnMarshal(t *testing.T) {
	var tests = []struct {
		inputAuthor *Author
	}{
		{&Author{Owner: Name("test"), Weight: 1}},
		{&Author{Owner: HexToPubKey("123455"), Weight: 1}},
		{&Author{Owner: HexToAddress("13123123123"), Weight: 1}},
	}
	for _, test := range tests {
		authorBytes, err := json.Marshal(test.inputAuthor)

		if err != nil {
			t.Fatal(err)
		}
		outputAuthor := &Author{}
		if err := json.Unmarshal(authorBytes, outputAuthor); err != nil {
			t.Fatal(err)
		}
		if test.inputAuthor.Owner.String() != outputAuthor.Owner.String() {
			t.Fatal("not equal")
		}
	}
}
