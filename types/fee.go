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

//type for fee
const (
	AssetFeeType    = uint64(0)
	ContractFeeType = uint64(1)
	CoinbaseFeeType = uint64(2)
)

// DistributeGas
type DistributeGas struct {
	Value  int64
	TypeID uint64
}

type DistributeKey struct {
	ObjectName string
	ObjectType uint64
}

type DistributeKeys []DistributeKey

func (keys DistributeKeys) Len() int {
	return len(keys)
}

func (keys DistributeKeys) Less(i, j int) bool {
	if keys[i].ObjectName == keys[j].ObjectName {
		return keys[i].ObjectType < keys[j].ObjectType
	}
	return keys[i].ObjectName < keys[j].ObjectName
}

func (keys DistributeKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}
