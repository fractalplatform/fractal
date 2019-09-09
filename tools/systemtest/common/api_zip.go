package common

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
