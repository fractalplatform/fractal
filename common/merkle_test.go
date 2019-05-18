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
	"testing"
)

func TestMerkleRoot(t *testing.T) {
	cases := []struct {
		hashs []Hash
		want  Hash
	}{{
		hashs: []Hash{
			BytesToHash([]byte{'0'}),
		},
		want: HexToHash("0x6ff97a59c90d62cc7236ba3a37cd85351bf564556780cf8c1157a220f31f0cbb"),
	}, {
		hashs: []Hash{
			BytesToHash([]byte{'0'}),
			BytesToHash([]byte{'1'}),
		},
		want: HexToHash("0x0a8c3b5073db80d0a934f3e9254ea0502cad1e2405c01b2289d78359d86f5077"),
	}, {
		hashs: []Hash{
			BytesToHash([]byte{'0'}),
			BytesToHash([]byte{'1'}),
			BytesToHash([]byte{'2'}),
		},
		want: HexToHash("0x0e9af40e80945a79d7f3df8e0d188b7907ade7b0d59faee00c2f1306c15d1fea"),
	}, {
		hashs: []Hash{
			BytesToHash([]byte{'0'}),
			BytesToHash([]byte{'1'}),
			BytesToHash([]byte{'2'}),
			BytesToHash([]byte{'4'}),
		},
		want: HexToHash("0x4339975bdf87d0bb81f8b7ced7a087526e3d2c419594b1343a6f80c5e95a6ab6"),
	}}

	for _, c := range cases {
		got := MerkleRoot(c.hashs)
		if got != c.want {
			t.Errorf("got merkle root = %v want %v", got.Hex(), c.want.Hex())
		}
	}
}
