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

package sdk

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
)

// SendRawTransaction send tx
func (api *API) SendRawTransaction(rawTx []byte) (common.Hash, error) {
	hash := new(common.Hash)
	err := api.client.Call(hash, "ft_sendRawTransaction", hexutil.Bytes(rawTx))
	return *hash, err
}

// CurrentBlock current block info
func (api *API) CurrentBlock(fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := api.client.Call(&block, "ft_getCurrentBlock", fullTx)
	return block, err
}

// BlockByHash block info
func (api *API) BlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := api.client.Call(&block, "ft_getBlockByHash", hash, fullTx)
	return block, err
}

// BlockByNumber block info
func (api *API) BlockByNumber(number int64, fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := api.client.Call(&block, "ft_getBlockByNumber", rpc.BlockNumber(number), fullTx)
	return block, err
}

// TransactionByHash tx info
func (api *API) TransactionByHash(hash common.Hash) (*types.RPCTransaction, error) {
	tx := &types.RPCTransaction{}
	err := api.client.Call(tx, "ft_getTransactionByHash", hash)
	return tx, err
}

// TransactionReceiptByHash tx info
func (api *API) TransactionReceiptByHash(hash common.Hash) (*types.RPCReceipt, error) {
	receipt := &types.RPCReceipt{}
	err := api.client.Call(receipt, "ft_getTransactionReceipt", hash)
	return receipt, err
}

// GasPrice gas price
func (api *API) GasPrice() (*big.Int, error) {
	gasprice := big.NewInt(0)
	err := api.client.Call(gasprice, "ft_gasPrice")
	return gasprice, err
}
