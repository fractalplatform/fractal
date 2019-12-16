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
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/common/hexutil"
	"github.com/fractalplatform/fractal/utils/rlp"
)

const (
	// ReceiptStatusFailed is the status code of a action if execution failed.
	ReceiptStatusFailed = uint64(0)

	// ReceiptStatusSuccessful is the status code of a action if execution succeeded.
	ReceiptStatusSuccessful = uint64(1)
)

type GasDistribution struct {
	Account string `json:"name"`
	Gas     uint64 `json:"gas"`
	TypeID  uint64 `json:"typeID"`
}

// Receipt represents the results of a transaction.
type Receipt struct {
	PostState         []byte
	Status            uint64
	Index             uint64
	GasUsed           uint64
	GasAllot          []*GasDistribution
	Error             string
	CumulativeGasUsed uint64
	Bloom             Bloom
	Logs              []*Log
	TxHash            common.Hash
	TotalGasUsed      uint64
	internalTxsLog    *DetailTx
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, cumulativeGasUsed, totalGasUsed uint64) *Receipt {
	return &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: cumulativeGasUsed, TotalGasUsed: totalGasUsed}
}

// Size returns the approximate memory used by all internal contents
func (r *Receipt) Size() common.StorageSize {
	bytes, _ := rlp.EncodeToBytes(r)
	return common.StorageSize(len(bytes))
}

// Hash hashes the RLP encoding of Receipt.
func (r *Receipt) Hash() common.Hash {
	result := &Receipt{
		CumulativeGasUsed: r.CumulativeGasUsed,
		TotalGasUsed:      r.TotalGasUsed,
		Bloom:             r.Bloom,
		Status:            r.Status,
		Index:             r.Index,
		GasUsed:           r.GasUsed,
		Error:             r.Error,
		TxHash:            r.TxHash,
		GasAllot:          r.GasAllot,
	}

	result.PostState = make([]byte, len(r.PostState))
	copy(result.PostState, r.PostState)

	var logs []*Log
	for _, l := range r.Logs {
		log := &Log{
			Name:   l.Name,
			Topics: l.Topics,
		}
		log.Data = make([]byte, len(l.Data))
		copy(log.Data, l.Data)
		logs = append(logs, log)
	}
	result.Logs = logs

	return common.RlpHash(result)
}

// RPCReceipt that will serialize to the RPC representation of a Receipt.
type RPCReceipt struct {
	BlockNumber       uint64             `json:"blockNumber"`
	BlockHash         common.Hash        `json:"blockHash"`
	Hash              common.Hash        `json:"txHash"`
	TransactionIndex  uint64             `json:"transactionIndex"`
	PostState         hexutil.Bytes      `json:"postState"`
	TxType            uint64             `json:"Type"`
	Status            uint64             `json:"status"`
	GasUsed           uint64             `json:"gasUsed"`
	GasAllot          []*GasDistribution `json:"gasAllot"`
	Error             string             `json:"error"`
	CumulativeGasUsed uint64             `json:"cumulativeGasUsed"`
	TotalGasUsed      uint64             `json:"totalGasUsed"`
	Bloom             Bloom              `json:"logsBloom"`
	Logs              []*RPCLog          `json:"logs"`
}

// NewRPCReceipt returns a Receipt that will serialize to the RPC.
func (r *Receipt) NewRPCReceipt(blockHash common.Hash, blockNumber uint64, index uint64, tx *Transaction) *RPCReceipt {
	result := &RPCReceipt{
		BlockNumber:       blockNumber,
		BlockHash:         blockHash,
		Hash:              tx.Hash(),
		TransactionIndex:  index,
		PostState:         hexutil.Bytes(r.PostState),
		TxType:            uint64(tx.Type()),
		Status:            r.Status,
		GasUsed:           r.GasUsed,
		GasAllot:          r.GasAllot,
		Error:             r.Error,
		CumulativeGasUsed: r.CumulativeGasUsed,
		TotalGasUsed:      r.TotalGasUsed,
		Bloom:             r.Bloom,
	}

	var rlogs []*RPCLog
	for _, l := range r.Logs {
		rlogs = append(rlogs, l.NewRPCLog())
	}
	result.Logs = rlogs

	return result
}

func (r *Receipt) GetInternalTxsLog() *DetailTx {
	return r.internalTxsLog
}

func (r *Receipt) SetInternalTxsLog(dtxs *DetailTx) {
	r.internalTxsLog = dtxs
}
