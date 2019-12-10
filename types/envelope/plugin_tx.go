package envelope

import (
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/common/hexutil"
)

type PayloadType uint64

type PluginTx struct {
	BasicTx
	PType     PayloadType
	AssetID   uint64
	Amount    *big.Int
	Payload   []byte
	Remark    []byte
	Signature []byte
}

func NewPluginTx(pType PayloadType, from, to string, nonce, assetID, gasAssetID, gasLimit uint64,
	gasPrice, amount *big.Int, payload, remark []byte) (*PluginTx, error) {
	if len(payload) > 0 {
		payload = common.CopyBytes(payload)
	}
	if len(remark) > 0 {
		remark = common.CopyBytes(remark)
	}

	tx := PluginTx{
		BasicTx: BasicTx{
			Typ:        Plugin,
			GasAssetID: gasAssetID,
			GasPrice:   gasPrice,
			GasLimit:   gasLimit,
			Nonce:      nonce,
			From:       from,
			To:         to,
		},
		PType:   pType,
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

func (p *PluginTx) GetAssetID() uint64 {
	return p.AssetID
}

func (p *PluginTx) Value() *big.Int {
	return p.Amount
}

func (p *PluginTx) GetPayload() []byte {
	return p.Payload
}

func (p *PluginTx) PayloadType() PayloadType {
	return p.PType
}

func (p *PluginTx) Cost() *big.Int {
	total := new(big.Int)
	total.Add(total, new(big.Int).Mul(p.GasPrice, new(big.Int).SetUint64(p.GasLimit)))
	return total
}

func (p *PluginTx) Hash() common.Hash {
	return common.RlpHash(p)
}

// SignHash hashes the action sign hash
func (p *PluginTx) SignHash(chainID *big.Int) common.Hash {
	return common.RlpHash([]interface{}{
		p.Typ,
		p.GasAssetID,
		p.GasPrice,
		p.GasLimit,
		p.Nonce,
		p.From,
		p.To,
		p.PType,
		p.AssetID,
		p.Amount,
		p.Payload,
		p.Remark,
		chainID, uint(0), uint(0),
	})
}

func (p *PluginTx) GetSign() []byte { return p.Signature }

type RPCPlugin struct {
	BlockNumber      uint64        `json:"blockNumber"`
	TransactionIndex uint64        `json:"transactionIndex"`
	BlockHash        common.Hash   `json:"blockHash"`
	Hash             common.Hash   `json:"txHash"`
	Type             uint64        `json:"type"`
	PType            uint64        `json:"payloadType"`
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
func (p *PluginTx) NewRPCTransaction(blockHash common.Hash, blockNumber uint64, index uint64) interface{} {
	result := &RPCPlugin{
		Hash:       p.Hash(),
		Type:       uint64(p.Typ),
		PType:      uint64(p.PType),
		Nonce:      p.Nonce,
		From:       p.Sender(),
		To:         p.To,
		AssetID:    p.AssetID,
		GasLimit:   p.GasLimit,
		Amount:     p.Amount,
		Remark:     hexutil.Bytes(p.Remark),
		Payload:    hexutil.Bytes(p.Payload),
		GasAssetID: p.GasAssetID,
		GasPrice:   p.GasPrice,
		GasCost:    p.Cost(),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = blockNumber
		result.TransactionIndex = index
	}
	return result
}
