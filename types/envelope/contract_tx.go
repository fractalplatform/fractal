package envelope

import (
	"errors"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/common/hexutil"
)

type ContractTx struct {
	BasicTx
	AssetID   uint64
	Amount    *big.Int
	Payload   []byte
	Remark    []byte
	Signature []byte
}

func NewContractTx(typ Type, from, to string, nonce, assetID, gasAssetID, gasLimit uint64,
	gasPrice, amount *big.Int, payload, remark []byte) (*ContractTx, error) {
	if len(payload) == 0 {
		return nil, errors.New("contract transaction payload is empty")
	}
	payload = common.CopyBytes(payload)
	if len(remark) > 0 {
		remark = common.CopyBytes(remark)
	}

	tx := ContractTx{
		BasicTx: BasicTx{
			Typ:        typ,
			GasAssetID: gasAssetID,
			GasPrice:   gasPrice,
			GasLimit:   gasLimit,
			Nonce:      nonce,
			From:       from,
			To:         to,
		},
		AssetID: assetID,
		Amount:  new(big.Int),
		Payload: payload,
		Remark:  remark,
	}

	if amount != nil {
		tx.Amount.Set(amount)
	}
	return &tx, nil
}

func (ctx *ContractTx) Cost() *big.Int {
	total := new(big.Int)
	total.Add(total, new(big.Int).Mul(ctx.GasPrice, new(big.Int).SetUint64(ctx.GasLimit)))
	return total
}

func (ctx *ContractTx) Hash() common.Hash {
	return common.RlpHash(ctx)
}

func (p *ContractTx) GetAssetID() uint64 {
	return p.AssetID
}

func (p *ContractTx) Value() *big.Int {
	return p.Amount
}

func (p *ContractTx) GetPayload() []byte {
	return p.Payload
}

// SignHash hashes the action sign hash
func (ctx *ContractTx) SignHash(chainID *big.Int) common.Hash {
	return common.RlpHash([]interface{}{
		ctx.Typ,
		ctx.GasAssetID,
		ctx.GasPrice,
		ctx.GasLimit,
		ctx.Nonce,
		ctx.From,
		ctx.To,
		ctx.AssetID,
		ctx.Amount,
		ctx.Payload,
		ctx.Remark,
		chainID, uint(0), uint(0),
	})
}

func (ctx *ContractTx) GetSign() []byte { return ctx.Signature }

type RPCContract struct {
	BlockNumber      uint64        `json:"blockNumber"`
	TransactionIndex uint64        `json:"transactionIndex"`
	BlockHash        common.Hash   `json:"blockHash"`
	Hash             common.Hash   `json:"txHash"`
	Type             uint64        `json:"type"`
	Nonce            uint64        `json:"nonce"`
	From             string        `json:"from"`
	To               string        `json:"to"`
	AssetID          uint64        `json:"assetID"`
	GasLimit         uint64        `json:"gas"`
	Amount           *big.Int      `json:"value"`
	Remark           hexutil.Bytes `json:"remark"`
	Payload          hexutil.Bytes `json:"payload"`
	GasAssetID       uint64        `json:"gasAssetID"`
	GasPrice         *big.Int      `json:"gasPrice"`
	GasCost          *big.Int      `json:"gasCost"`
}

// NewRPCTransaction returns a transaction that will serialize to the RPC.
func (ctx *ContractTx) NewRPCTransaction(blockHash common.Hash, blockNumber uint64, index uint64) interface{} {
	result := &RPCContract{
		Hash:       ctx.Hash(),
		Type:       uint64(ctx.Typ),
		Nonce:      ctx.Nonce,
		From:       ctx.Sender(),
		To:         ctx.To,
		AssetID:    ctx.AssetID,
		GasLimit:   ctx.GasLimit,
		Amount:     ctx.Amount,
		Remark:     hexutil.Bytes(ctx.Remark),
		Payload:    hexutil.Bytes(ctx.Payload),
		GasAssetID: ctx.GasAssetID,
		GasPrice:   ctx.GasPrice,
		GasCost:    ctx.Cost(),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = blockNumber
		result.TransactionIndex = index
	}
	return result
}
