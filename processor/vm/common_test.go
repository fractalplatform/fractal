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
package vm

import (
	"bytes"
	"math/big"
	"testing"
)

func TestCalcMemSize(t *testing.T) {
	tests := []struct {
		off   *big.Int
		l     *big.Int
		which *big.Int
	}{
		{big.NewInt(0), big.NewInt(0), big.NewInt(0)},
		{big.NewInt(10), big.NewInt(0), big.NewInt(0)},
		{big.NewInt(0), big.NewInt(10), big.NewInt(10)},
		{big.NewInt(10), big.NewInt(100), big.NewInt(110)},
	}
	for _, test := range tests {
		ret := calcMemSize(test.off, test.l)
		if ret.Cmp(test.which) != 0 {
			t.Fatalf("expected %x, got %02x", test.which, ret)
		}
	}
}

func TestGetData(t *testing.T) {
	tests := []struct {
		data  []byte
		start uint64
		size  uint64
		res   []byte
	}{
		{[]byte{}, 0, 0, []byte{}},
		{[]byte{}, 1, 0, []byte{}},
		{[]byte{}, 0, 1, []byte{0}},
		{[]byte{}, 1, 1, []byte{0}},
		{[]byte{1, 2, 3, 4, 5}, 0, 0, []byte{}},
		{[]byte{1, 2, 3, 4, 5}, 1, 0, []byte{}},
		{[]byte{1, 2, 3, 4, 5}, 0, 1, []byte{1}},
		{[]byte{1, 2, 3, 4, 5}, 1, 1, []byte{2}},
		{[]byte{1, 2, 3, 4, 5}, 1, 10, []byte{2, 3, 4, 5, 0, 0, 0, 0, 0, 0}},
		{[]byte{1, 2, 3, 4, 5, 0, 0, 0}, 1, 10, []byte{2, 3, 4, 5, 0, 0, 0, 0, 0, 0}},
	}
	for _, test := range tests {
		ret := getData(test.data, test.start, test.size)
		if !bytes.Equal(ret, test.res) {
			t.Fatal("expected ,", test.res, " got ", ret)
		}
	}
}

func TestGetDataBig(t *testing.T) {
	tests := []struct {
		data  []byte
		start *big.Int
		size  *big.Int
		res   []byte
	}{
		{[]byte{}, big.NewInt(0), big.NewInt(0), []byte{}},
		{[]byte{}, big.NewInt(1), big.NewInt(0), []byte{}},
		{[]byte{}, big.NewInt(0), big.NewInt(1), []byte{0}},
		{[]byte{}, big.NewInt(1), big.NewInt(1), []byte{0}},
		{[]byte{1, 2, 3, 4, 5}, big.NewInt(0), big.NewInt(0), []byte{}},
		{[]byte{1, 2, 3, 4, 5}, big.NewInt(1), big.NewInt(0), []byte{}},
		{[]byte{1, 2, 3, 4, 5}, big.NewInt(0), big.NewInt(1), []byte{1}},
		{[]byte{1, 2, 3, 4, 5}, big.NewInt(1), big.NewInt(1), []byte{2}},
		{[]byte{1, 2, 3, 4, 5}, big.NewInt(1), big.NewInt(10), []byte{2, 3, 4, 5, 0, 0, 0, 0, 0, 0}},
		{[]byte{1, 2, 3, 4, 5, 0, 0, 0}, big.NewInt(1), big.NewInt(10), []byte{2, 3, 4, 5, 0, 0, 0, 0, 0, 0}},
	}
	for _, test := range tests {
		ret := getDataBig(test.data, test.start, test.size)
		if !bytes.Equal(ret, test.res) {
			t.Fatal("expected ,", test.res, " got ", ret)
		}
	}
}

func TestToWordSize(t *testing.T) {
	tests := []struct {
		size uint64
		res  uint64
	}{
		{31, 1},
		{32, 1},
		{33, 2},
		{0, 0},
	}
	for _, test := range tests {
		ret := toWordSize(test.size)
		if ret != test.res {
			t.Fatalf("expected %x, got %02x", test.res, ret)
		}
	}
}

func TestAllZero(t *testing.T) {
	tests := []struct {
		data []byte
		res  bool
	}{
		{[]byte{0, 0, 0}, true},
		{[]byte{0, 0, 1}, false},
	}
	for _, test := range tests {
		ret := allZero(test.data)
		if ret != test.res {
			t.Fatal("expected ,", test.res, " got ", ret)
		}
	}
}
