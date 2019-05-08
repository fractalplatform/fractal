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

package common

import (
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// GetCurrentBlock returns cureent block.
func GetCurrentBlock(fullTx bool) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := ClientCall("ft_getCurrentBlock", &result, fullTx)
	return result, err
}

//GetBlockByNumber returns the requested block.
func GetBlockByNumber(number uint64, fullTx bool) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := ClientCall("ft_getBlockByNumber", &result, number, fullTx)
	return result, err
}

// GetTransactionByHash returns the transaction for the given hash
func GetTransactionByHash(hash common.Hash) (*types.RPCTransaction, error) {
	result := &types.RPCTransaction{}
	err := ClientCall("ft_getTransactionByHash", &result, hash)
	return result, err
}
