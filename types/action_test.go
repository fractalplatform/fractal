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
	"github.com/fractalplatform/fractal/params"
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
		[]byte("test remark"),
	)

	testAction2 = NewAction(
		UpdateAccount,
		common.Name("fromname"),
		common.Name("fractal.account"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(1000),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction3 = NewAction(
		UpdateAccount,
		common.Name("fromname"),
		common.Name("fractal.account"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction4 = NewAction(
		CreateContract,
		common.Name("fromname"),
		common.Name("fromname"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction5 = NewAction(
		CreateAccount,
		common.Name("fromname"),
		common.Name("fractal.aaaaaa"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
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

func TestAction_Check(t *testing.T) {
	actionBytes, err := rlp.EncodeToBytes(testAction)
	if err != nil {
		t.Fatal(err)
	}
	//
	actAction := &Action{}
	if err := rlp.Decode(bytes.NewReader(actionBytes), &actAction); err != nil {
		t.Fatal(err)
	}

	if err := actAction.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction_CheckValue err, wantErr %v", true)
	}

	//test2
	actionBytes2, err := rlp.EncodeToBytes(testAction2)
	if err != nil {
		t.Fatal(err)
	}

	actAction2 := &Action{}
	if err := rlp.Decode(bytes.NewReader(actionBytes2), &actAction2); err != nil {
		t.Fatal(err)
	}

	if err := actAction2.Check(params.NextForkID, params.DefaultChainconfig); err == nil {
		t.Errorf("TestAction2_CheckValue err, wantErr %v", false)
	}

	//test3
	actionBytes3, err := rlp.EncodeToBytes(testAction3)
	if err != nil {
		t.Fatal(err)
	}

	actAction3 := &Action{}
	if err := rlp.Decode(bytes.NewReader(actionBytes3), &actAction3); err != nil {
		t.Fatal(err)
	}

	if err := actAction3.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	//test4
	actionBytes4, err := rlp.EncodeToBytes(testAction4)
	if err != nil {
		t.Fatal(err)
	}

	actAction4 := &Action{}
	if err := rlp.Decode(bytes.NewReader(actionBytes4), &actAction4); err != nil {
		t.Fatal(err)
	}

	if err := actAction3.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	//test5
	actionBytes5, err := rlp.EncodeToBytes(testAction5)
	if err != nil {
		t.Fatal(err)
	}

	actAction5 := &Action{}
	if err := rlp.Decode(bytes.NewReader(actionBytes5), &actAction5); err != nil {
		t.Fatal(err)
	}

	if err := actAction3.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}
}

var (
	testAction10 = NewAction(
		IncreaseAsset,
		common.Name("fromname"),
		common.Name("fractal.asset"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction11 = NewAction(
		IssueAsset,
		common.Name("fromname"),
		common.Name("fractal.asset"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction12 = NewAction(
		DestroyAsset,
		common.Name("fromname"),
		common.Name("fractal.asset"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction13 = NewAction(
		SetAssetOwner,
		common.Name("fromname"),
		common.Name("fractal.asset"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)
	testAction14 = NewAction(
		UpdateAsset,
		common.Name("fromname"),
		common.Name("fractal.asset"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)
)

func TestAction_Check2(t *testing.T) {

	// actionBytes10, err := rlp.EncodeToBytes(testAction10)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// actAction10 := &Action{}
	// if err := rlp.Decode(bytes.NewReader(actionBytes10), &actAction10); err != nil {
	// 	t.Fatal(err)
	// }

	if err := testAction10.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction11.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction12.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction13.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction14.Check(params.NextForkID, params.DefaultChainconfig); err != nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

}

var (
	testAction20 = NewAction(
		RegCandidate,
		common.Name("fromname"),
		common.Name("fractal.dpos"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction21 = NewAction(
		UpdateCandidate,
		common.Name("fromname"),
		common.Name("fractal.dpos"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction22 = NewAction(
		UnregCandidate,
		common.Name("fromname"),
		common.Name("fractal.dpos"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)

	testAction23 = NewAction(
		VoteCandidate,
		common.Name("fromname"),
		common.Name("fractal.dpos"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)
	testAction24 = NewAction(
		RefundCandidate,
		common.Name("fromname"),
		common.Name("fractal.dpos"),
		uint64(1),
		uint64(3),
		uint64(2000),
		big.NewInt(0),
		[]byte("test action"),
		[]byte("test remark"),
	)
)

func TestAction_Check3(t *testing.T) {

	// actionBytes10, err := rlp.EncodeToBytes(testAction10)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// actAction10 := &Action{}
	// if err := rlp.Decode(bytes.NewReader(actionBytes10), &actAction10); err != nil {
	// 	t.Fatal(err)
	// }

	if err := testAction20.Check(params.NextForkID, params.DefaultChainconfig); err == nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction21.Check(params.NextForkID, params.DefaultChainconfig); err == nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction22.Check(params.NextForkID, params.DefaultChainconfig); err == nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction23.Check(params.NextForkID, params.DefaultChainconfig); err == nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

	if err := testAction24.Check(params.NextForkID, params.DefaultChainconfig); err == nil {
		t.Errorf("TestAction3_CheckValue err, wantErr %v", false)
	}

}
