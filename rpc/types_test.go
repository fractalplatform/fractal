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
