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
	"bytes"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
	"github.com/stretchr/testify/assert"
)

var (
	testAction = NewAction(
		Transfer,
		common.Name("fromname"),
		common.Name("totoname"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(1000),
		[]byte("test action"),
	)
)

func TestActionEncodeAndDecode(t *testing.T) {
	actionBytes, err := rlp.EncodeToBytes(testAction)
	if err != nil {
		t.Fatal(err)
	}

	actAction := &Action{}
	if err := rlp.Decode(bytes.NewReader(actionBytes), &actAction); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testAction, actAction)
}
