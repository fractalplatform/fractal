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

package txpool

import (
	"math"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
)

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(action *types.Action) (uint64, error) {
	// Bump the required gas by the amount of transactional data
	dataGasFunc := func(data []byte) (uint64, error) {
		var gas uint64
		if len(data) > 0 {
			// Zero and non-zero bytes are priced differently
			var nz uint64
			for _, byt := range data {
				if byt != 0 {
					nz++
				}
			}
			// Make sure we don't exceed uint64 for all data combinations
			if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
				return 0, ErrOutOfGas
			}
			gas += nz * params.TxDataNonZeroGas

			z := uint64(len(data)) - nz
			if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
				return 0, ErrOutOfGas
			}
			gas += z * params.TxDataZeroGas
		}
		return gas, nil
	}

	var gas uint64

	if action.Type() == types.CreateContract || action.Type() == types.CreateAccount || action.Type() == types.IssueAsset {
		gas += params.ActionGasContractCreation
	} else {
		gas += params.ActionGas
	}

	dataGas, err := dataGasFunc(action.Data())
	if err != nil {
		return 0, err
	}
	gas += dataGas

	gas += (uint64(len(action.GetSign()) - 1)) * params.ActionGas

	return gas, nil
}
