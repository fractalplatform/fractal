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
	"math"
	"math/big"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestIsHexAddress(t *testing.T) {
	tests := []struct {
		str string
		exp bool
	}{
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"0X5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"0XAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1", false},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beae", false},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed11", false},
		{"0xxaaeb6053f3e94c9b9a09f33669435e7ef1beaed", false},
	}

	for _, test := range tests {
		if result := IsHexAddress(test.str); result != test.exp {
			t.Errorf("IsHexAddress(%s) == %v; expected %v",
				test.str, result, test.exp)
		}
	}
}

func TestAddressUnmarshalJSON(t *testing.T) {
	var tests = []struct {
		Input     string
		ShouldErr bool
		Output    *big.Int
	}{
		{"", true, nil},
		{`""`, true, nil},
		{`"0x"`, true, nil},
		{`"0x00"`, true, nil},
		{`"0xG000000000000000000000000000000000000000"`, true, nil},
		{`"0x0000000000000000000000000000000000000000"`, false, big.NewInt(0)},
		{`"0x0000000000000000000000000000000000000010"`, false, big.NewInt(16)},
	}
	for i, test := range tests {
		var v Address
		err := json.Unmarshal([]byte(test.Input), &v)
		if err != nil && !test.ShouldErr {
			t.Errorf("test #%d: unexpected error: %v", i, err)
		}

		if err == nil {
			if test.ShouldErr {
				t.Errorf("test #%d: expected error, got none", i)
			}
			if v.Big().Cmp(test.Output) != 0 {
				t.Errorf("test #%d: address mismatch: have %v, want %v", i, v.Big(), test.Output)
			}
		}
	}
}

func TestAddressHexChecksum(t *testing.T) {
	var tests = []struct {
		Input  string
		Output string
	}{
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"},
		{"0xfb6916095ca1df60bb79ce92ce3ea74c37c5d359", "0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359"},
		{"0xdbf03b407c01e7cd3cbea99509d93f8dddc8c6fb", "0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB"},
		{"0xd1220a0cf47c7b9be7a2e6ba89f429762e7b9adb", "0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb"},
		// Ensure that non-standard length input values are handled correctly
		{"0xa", "0x000000000000000000000000000000000000000A"},
		{"0x0a", "0x000000000000000000000000000000000000000A"},
		{"0x00a", "0x000000000000000000000000000000000000000A"},
		{"0x000000000000000000000000000000000000000a", "0x000000000000000000000000000000000000000A"},
	}
	for i, test := range tests {
		output := HexToAddress(test.Input).Hex()
		if output != test.Output {
			t.Errorf("test #%d: failed to match when it should (%s != %s)", i, output, test.Output)
		}
	}
}

func TestAddressConvert(t *testing.T) {
	var tests = []struct {
		BigInt    *big.Int
		HexString string
		Bytes     []byte
	}{
		{big.NewInt(math.MaxInt64),
			"0x0000000000000000000000007FFfFFFffFfFfFff",
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 127, 255, 255, 255, 255, 255, 255, 255}},
		{big.NewInt(10),
			"0x000000000000000000000000000000000000000A",
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10}},
		{big.NewInt(15),
			"0x000000000000000000000000000000000000000F",
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 15}},
	}
	for i, test := range tests {
		addr := BigToAddress(test.BigInt)
		if addr.Hex() != test.HexString {
			t.Errorf("test #%d: failed to match when it should (%s != %s)", i, addr.Hex(), test.HexString)
		}
		if !bytes.Equal(addr.Bytes(), test.Bytes) {
			t.Errorf("test #%d: failed to match when it should (%x != %x)", i, addr.Bytes(), test.Bytes)
		}
	}
}

func TestAddressMarshal(t *testing.T) {
	testAddr := HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	if marshaltext, err := yaml.Marshal(testAddr); err != nil {
		t.Errorf("MarshalText err: %v", err)
	} else {
		target := []byte{48, 120, 53, 97, 97, 101, 98, 54, 48, 53, 51, 102, 51, 101, 57, 52, 99, 57, 98, 57, 97, 48, 57, 102, 51, 51, 54, 54, 57, 52, 51, 53, 101, 55, 101, 102, 49, 98, 101, 97, 101, 100, 10}
		if !bytes.Equal(marshaltext, target) {
			t.Errorf("MarshalText mismatch when it should (%x != %x)", marshaltext, target)
		}

		newAddress := Address{}
		if err := yaml.Unmarshal(marshaltext, &newAddress); err != nil {
			t.Errorf("UnmarshalText err: %v", err)
		}
		if 0 != newAddress.Compare(testAddr) {
			t.Errorf("UnmarshalText address mismatch")
		}
	}
}

func TestUnprefixedAddressMarshal(t *testing.T) {
	marshaltext := []byte{53, 97, 97, 101, 98, 54, 48, 53, 51, 102, 51, 101, 57, 52, 99, 57, 98, 57, 97, 48, 57, 102, 51, 51, 54, 54, 57, 52, 51, 53, 101, 55, 101, 102, 49, 98, 101, 97, 101, 100, 10}
	unprefixedAddr := UnprefixedAddress{}
	if err := yaml.Unmarshal(marshaltext, &unprefixedAddr); err != nil {
		t.Errorf("UnmarshalText err: %v", err)
	}

	if fetchedMarshaltext, err := yaml.Marshal(unprefixedAddr); err != nil {
		t.Errorf("MarshalText err: %v", err)
	} else {
		if !bytes.Equal(marshaltext, fetchedMarshaltext) {
			t.Errorf("MarshalText mismatch when it should (%s != %s)", marshaltext, fetchedMarshaltext)
		}
	}
}

func TestMixedcaseAddress(t *testing.T) {
	hexString := "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed"
	originalAddr := HexToAddress(hexString)
	testAddr := NewMixedcaseAddress(originalAddr)
	mixedcaseRes := "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed [chksum ok]"
	if mixedcaseRes != testAddr.String() {
		t.Errorf("MixedcaseAddress String mismatched")
	}

	mixedcaseRes = "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"
	if mixedcaseRes != testAddr.Original() {
		t.Errorf("MixedcaseAddress Original mismatched")
	}

	if 0 != testAddr.Address().Compare(originalAddr) {
		t.Errorf("MixedcaseAddress address mismatch")
	}
}

func BenchmarkAddressHex(b *testing.B) {
	testAddr := HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	for n := 0; n < b.N; n++ {
		testAddr.Hex()
	}
}

func TestMixedcaseAccount_Address(t *testing.T) {

	// Note: 0X{checksum_addr} is not valid according to spec above

	var res []struct {
		A     MixedcaseAddress
		Valid bool
	}
	if err := json.Unmarshal([]byte(`[
		{"A" : "0xae967917c465db8578ca9024c205720b1a3651A9", "Valid": false},
		{"A" : "0xAe967917c465db8578ca9024c205720b1a3651A9", "Valid": true},
		{"A" : "0XAe967917c465db8578ca9024c205720b1a3651A9", "Valid": false},
		{"A" : "0x1111111111111111111112222222222223333323", "Valid": true}
		]`), &res); err != nil {
		t.Fatal(err)
	}

	for _, r := range res {
		if got := r.A.ValidChecksum(); got != r.Valid {
			t.Errorf("Expected checksum %v, got checksum %v, input %v", r.Valid, got, r.A.String())
		}
	}

	//These should throw exceptions:
	var r2 []MixedcaseAddress
	for _, r := range []string{
		`["0x11111111111111111111122222222222233333"]`,     // Too short
		`["0x111111111111111111111222222222222333332"]`,    // Too short
		`["0x11111111111111111111122222222222233333234"]`,  // Too long
		`["0x111111111111111111111222222222222333332344"]`, // Too long
		`["1111111111111111111112222222222223333323"]`,     // Missing 0x
		`["x1111111111111111111112222222222223333323"]`,    // Missing 0
		`["0xG111111111111111111112222222222223333323"]`,   //Non-hex
	} {
		if err := json.Unmarshal([]byte(r), &r2); err == nil {
			t.Errorf("Expected failure, input %v", r)
		}

	}

}
