// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"math/big"

	"github.com/fractalplatform/fractal/types"
)

type FeeManager struct {
}

func NewFeeManager() (IFee, error) {
	return &FeeManager{}, nil
}

func (fm *FeeManager) DistributeGas(from string, gasMap map[types.DistributeKey]types.DistributeGas, assetID uint64, gasPrice *big.Int, am IAccount) ([]*types.GasDistribution, error) {
	var coinbase string
	var totalGas int64
	var gasAllots []*types.GasDistribution

	for key, gas := range gasMap {
		if key.ObjectType == types.CoinbaseFeeType {
			coinbase = key.ObjectName
		}
		totalGas += gas.Value
	}

	gasAllots = append(gasAllots, &types.GasDistribution{
		Account: coinbase,
		Gas:     uint64(totalGas),
		TypeID:  types.CoinbaseFeeType})

	gasBalance := new(big.Int).Mul(new(big.Int).SetInt64(totalGas), gasPrice)
	if err := am.TransferAsset(from, coinbase, assetID, gasBalance); err != nil {
		return nil, err
	}

	return gasAllots, nil
}
