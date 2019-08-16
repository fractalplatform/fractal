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
	"fmt"
	"sort"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/stretchr/testify/assert"
)

func TestDistributeKeys(t *testing.T) {
	key1 := DistributeKey{ObjectName: common.Name("ftoken"), ObjectType: 0}
	key2 := DistributeKey{ObjectName: common.Name("ftoken"), ObjectType: 1}
	key3 := DistributeKey{ObjectName: common.Name("contract"), ObjectType: 1}
	key4 := DistributeKey{ObjectName: common.Name("miner"), ObjectType: 2}

	founderGasMap := make(map[DistributeKey]DistributeGas, 0)
	founderGasMap[key1] = DistributeGas{0, 0}
	founderGasMap[key2] = DistributeGas{0, 1}
	founderGasMap[key3] = DistributeGas{0, 1}
	founderGasMap[key4] = DistributeGas{0, 2}

	var keys DistributeKeys
	for key, _ := range founderGasMap {
		keys = append(keys, key)
	}
	sort.Sort(keys)

	key4 = DistributeKey{ObjectName: common.Name("ftoken"), ObjectType: 0}
	key3 = DistributeKey{ObjectName: common.Name("ftoken"), ObjectType: 1}
	key2 = DistributeKey{ObjectName: common.Name("contract"), ObjectType: 1}
	key1 = DistributeKey{ObjectName: common.Name("miner"), ObjectType: 2}

	founderGasMap = make(map[DistributeKey]DistributeGas, 0)
	founderGasMap[key1] = DistributeGas{0, 0}
	founderGasMap[key2] = DistributeGas{0, 1}
	founderGasMap[key3] = DistributeGas{0, 1}
	founderGasMap[key4] = DistributeGas{0, 2}

	var keys1 DistributeKeys
	for key, _ := range founderGasMap {
		keys1 = append(keys1, key)
	}
	sort.Sort(keys1)
	for _, key := range keys1 {
		fmt.Println(key)
	}
	assert.Equal(t, keys, keys1)
}

func TestOverTimeAbort(t *testing.T) {
	evm := &EVM{}
	evm.OverTimeAbort()
	if evm.IsOverTime() != true {
		t.Error("IsOverTime test fail")
	}
}

func TestIsOverTime(t *testing.T) {
	evm := &EVM{}
	if evm.IsOverTime() == true {
		evm.OverTimeAbort()
	}

	if evm.IsOverTime() != false {
		t.Error("IsOverTime test fail")
	}
}

func TestCheckReceipt(t *testing.T) {
	//evm := &EVM{}
	//a := &types.Action{}
	//evm.CheckReceipt(a)
	return
}

func TestdistributeContractGas(t *testing.T) {
	evm := &EVM{}
	evm.distributeContractGas(0, common.Name(""), common.Name(""))
	return
}

func TestdistributeAssetGas(t *testing.T) {
	evm := &EVM{}
	evm.distributeAssetGas(0, common.Name(""), common.Name(""))
	return
}

func TestdistributeGasByScale(t *testing.T) {
	evm := &EVM{}
	evm.distributeAssetGas(0, common.Name(""), common.Name(""))
	return
}

func TestCallCode(t *testing.T) {
	//evm := &EVM{}
	//a := &types.Action{}
	//evm.CallCode(nil, a, 0)
	return
}

func TestChainConfig(t *testing.T) {
	evm := &EVM{}
	if evm.ChainConfig() != nil {
		t.Error("test ChainConfig fail")
	}
	return
}

func TestDistributeGasByScale(t *testing.T) {
	evm := &EVM{}
	evm.distributeGasByScale(0, 0)
	return
}
