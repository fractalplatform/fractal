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

package rpcapi

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *types.Transaction) (common.Hash, error) {
	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	log.Info("Submitted transaction", "fullhash", tx.Hash().Hex())
	return tx.Hash(), nil
}

// RPCMarshalBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func RPCMarshalBlock(chainID *big.Int, b *types.Block, inclTx bool, fullTx bool) map[string]interface{} {
	head := b.Header() // copies the header once
	fields := map[string]interface{}{
		"number":               head.Number,
		"hash":                 b.Hash(),
		"proposedIrreversible": head.ProposedIrreversible,
		"parentHash":           head.ParentHash,
		"logsBloom":            head.Bloom,
		"stateRoot":            head.Root,
		"miner":                head.Coinbase,
		"difficulty":           head.Difficulty,
		"extraData":            hexutil.Bytes(head.Extra),
		"size":                 b.Size(),
		"gasLimit":             head.GasLimit,
		"gasUsed":              head.GasUsed,
		"timestamp":            head.Time,
		"transactionsRoot":     head.TxsRoot,
		"receiptsRoot":         head.ReceiptsRoot,
		"forkID":               head.ForkID,
	}

	if inclTx {
		formatTx := func(tx *types.Transaction, index uint64) interface{} {
			return tx.Hash()
		}
		if fullTx {
			formatTx = func(tx *types.Transaction, index uint64) interface{} {
				return tx.NewRPCTransaction(b.Hash(), b.NumberU64(), index)
			}
		}
		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		for i, tx := range txs {
			transactions[i] = formatTx(tx, uint64(i))
		}
		fields["transactions"] = transactions
	}

	return fields
}

func RPCMarshalBlockWithPayer(chainID *big.Int, b *types.Block, inclTx bool, fullTx bool) map[string]interface{} {
	head := b.Header() // copies the header once
	fields := map[string]interface{}{
		"number":               head.Number,
		"hash":                 b.Hash(),
		"proposedIrreversible": head.ProposedIrreversible,
		"parentHash":           head.ParentHash,
		"logsBloom":            head.Bloom,
		"stateRoot":            head.Root,
		"miner":                head.Coinbase,
		"difficulty":           head.Difficulty,
		"extraData":            hexutil.Bytes(head.Extra),
		"size":                 b.Size(),
		"gasLimit":             head.GasLimit,
		"gasUsed":              head.GasUsed,
		"timestamp":            head.Time,
		"transactionsRoot":     head.TxsRoot,
		"receiptsRoot":         head.ReceiptsRoot,
		"forkID":               head.ForkID,
	}

	if inclTx {
		formatTx := func(tx *types.Transaction, index uint64) interface{} {
			return tx.Hash()
		}
		if fullTx {
			formatTx = func(tx *types.Transaction, index uint64) interface{} {
				return tx.NewRPCTransactionWithPayer(b.Hash(), b.NumberU64(), index)
			}
		}
		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		for i, tx := range txs {
			transactions[i] = formatTx(tx, uint64(i))
		}
		fields["transactions"] = transactions
	}

	return fields
}
