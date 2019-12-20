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

package rpc

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
)

func TestBlockNumberJSONUnmarshal(t *testing.T) {
	tests := []struct {
		input    string
		mustFail bool
		expected BlockNumber
	}{
		0: {`"0x"`, true, BlockNumber(0)},
		1: {`0`, false, BlockNumber(0)},
		2: {`1`, false, BlockNumber(1)},
		3: {`9223372036854775807`, false, BlockNumber(math.MaxInt64)},
		4: {`9223372036854775808`, true, BlockNumber(0)},
		5: {`"latest"`, false, LatestBlockNumber},
		6: {`"earliest"`, false, EarliestBlockNumber},
		7: {`someString`, true, BlockNumber(0)},
		8: {`""`, true, BlockNumber(0)},
		9: {``, true, BlockNumber(0)},
	}

	for i, test := range tests {
		var num BlockNumber
		err := json.Unmarshal([]byte(test.input), &num)
		if test.mustFail && err == nil {
			t.Errorf("Test %d should fail", i)
			continue
		}
		if !test.mustFail && err != nil {
			t.Errorf("Test %d should pass but got err: %v", i, err)
			continue
		}
		if num != test.expected {
			t.Errorf("Test %d got unexpected value, want %d, got %d", i, test.expected, num)
		}
	}
}

// func TestBlockNumberOrHash_UnmarshalJSON(t *testing.T) {
// 	tests := []struct {
// 		input    string
// 		mustFail bool
// 		expected BlockNumberOrHash
// 	}{
// 		0:  {`"0x"`, true, BlockNumberOrHash{}},
// 		1:  {`"0x0"`, false, BlockNumberOrHashWithNumber(0)},
// 		2:  {`"0X1"`, false, BlockNumberOrHashWithNumber(1)},
// 		3:  {`"0x00"`, true, BlockNumberOrHash{}},
// 		4:  {`"0x01"`, true, BlockNumberOrHash{}},
// 		5:  {`"0x1"`, false, BlockNumberOrHashWithNumber(1)},
// 		6:  {`"0x12"`, false, BlockNumberOrHashWithNumber(18)},
// 		7:  {`"0x7fffffffffffffff"`, false, BlockNumberOrHashWithNumber(math.MaxInt64)},
// 		8:  {`"0x8000000000000000"`, true, BlockNumberOrHash{}},
// 		9:  {"0", true, BlockNumberOrHash{}},
// 		10: {`"ff"`, true, BlockNumberOrHash{}},
// 		11: {`"latest"`, false, BlockNumberOrHashWithNumber(LatestBlockNumber)},
// 		12: {`"earliest"`, false, BlockNumberOrHashWithNumber(EarliestBlockNumber)},
// 		13: {`someString`, true, BlockNumberOrHash{}},
// 		14: {`""`, true, BlockNumberOrHash{}},
// 		15: {``, true, BlockNumberOrHash{}},
// 		16: {`"0x0000000000000000000000000000000000000000000000000000000000000000"`, false, BlockNumberOrHashWithHash(common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"), false)},
// 		17: {`{"blockHash":"0x0000000000000000000000000000000000000000000000000000000000000000"}`, false, BlockNumberOrHashWithHash(common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"), false)},
// 		18: {`{"blockHash":"0x0000000000000000000000000000000000000000000000000000000000000000","requireCanonical":false}`, false, BlockNumberOrHashWithHash(common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"), false)},
// 		19: {`{"blockHash":"0x0000000000000000000000000000000000000000000000000000000000000000","requireCanonical":true}`, false, BlockNumberOrHashWithHash(common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"), true)},
// 		20: {`{"blockNumber":"1"}`, false, BlockNumberOrHashWithNumber(1)},
// 		21: {`{"blockNumber":"latest"}`, false, BlockNumberOrHashWithNumber(LatestBlockNumber)},
// 		22: {`{"blockNumber":"earliest"}`, false, BlockNumberOrHashWithNumber(EarliestBlockNumber)},
// 		23: {`{"blockNumber":"0x1", "blockHash":"0x0000000000000000000000000000000000000000000000000000000000000000"}`, true, BlockNumberOrHash{}},
// 	}

// 	for i, test := range tests {
// 		var bnh BlockNumberOrHash
// 		err := json.Unmarshal([]byte(test.input), &bnh)
// 		if test.mustFail && err == nil {
// 			t.Errorf("Test %d should fail", i)
// 			continue
// 		}
// 		if !test.mustFail && err != nil {
// 			t.Errorf("Test %d should pass but got err: %v", i, err)
// 			continue
// 		}
// 		hash, hashOk := bnh.Hash()
// 		expectedHash, expectedHashOk := test.expected.Hash()
// 		num, numOk := bnh.Number()
// 		expectedNum, expectedNumOk := test.expected.Number()
// 		if bnh.RequireCanonical != test.expected.RequireCanonical ||
// 			hash != expectedHash || hashOk != expectedHashOk ||
// 			num != expectedNum || numOk != expectedNumOk {
// 			t.Errorf("Test %d got unexpected value, want %v, got %v", i, test.expected, bnh)
// 		}
// 	}
// }
