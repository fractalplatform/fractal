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

package types

import (
	"testing"

	"github.com/fractalplatform/fractal/utils/rlp"

	"github.com/stretchr/testify/assert"
)

func TestReceiptEncodeAndDecode(t *testing.T) {
	testR := NewReceipt([]byte("root"), 1000, 1000)
	testR.Logs = make([]*Log, 0)
	gasAllot := make([]*GasDistribution, 0)
	testR.ActionResults = append(testR.ActionResults, &ActionResult{Status: ReceiptStatusFailed, Index: uint64(0), GasAllot: gasAllot, GasUsed: uint64(100)})
	bytes, err := rlp.EncodeToBytes(testR)
	if err != nil {
		t.Fatal(err)
	}
	newR := &Receipt{}
	rlp.DecodeBytes(bytes, newR)
	assert.Equal(t, testR, newR)
}
