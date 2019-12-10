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
	"encoding/json"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
	"github.com/stretchr/testify/assert"
)

func TestContractTransactionEncodeAndDecode(t *testing.T) {
	ctx, err := envelope.NewContractTx(envelope.CreateContract, "sender", "receipt", 0, 0, 0, 100000,
		big.NewInt(100), big.NewInt(1000), []byte("payload"), []byte("remark"))
	assert.NoError(t, err)

	checkTxEncodeAndDecode(t, ctx)
}

func TestPluginTransactionEncodeAndDecode(t *testing.T) {
	ptx, err := envelope.NewPluginTx(10, "sender", "receipt", 0, 0, 0, 100000,
		big.NewInt(100), big.NewInt(1000), []byte("payload"), []byte("remark"))
	assert.NoError(t, err)

	checkTxEncodeAndDecode(t, ptx)
}

func checkTxEncodeAndDecode(t *testing.T, env envelope.Envelope) {
	tx := NewTransaction(env)
	txbytes, err := rlp.EncodeToBytes(tx)
	assert.NoError(t, err)

	newtx := new(Transaction)
	err = rlp.Decode(bytes.NewReader(txbytes), newtx)

	assert.NoError(t, err)
	newtxbytes, err := rlp.EncodeToBytes(newtx)
	assert.NoError(t, err)

	assert.Equal(t, newtxbytes, txbytes)

	newrpctx := newtx.NewRPCTransaction(common.Hash{}, 0, 0)
	testrpctx := tx.NewRPCTransaction(common.Hash{}, 0, 0)
	testrpctxbytes, err := json.Marshal(testrpctx)
	assert.NoError(t, err)
	newrpctxbytes, err := json.Marshal(newrpctx)
	assert.NoError(t, err)

	assert.Equal(t, newrpctxbytes, testrpctxbytes)
}
