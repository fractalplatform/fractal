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
	"container/heap"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	// ErrOversizedData is returned if the input data of a transaction is greater
	// than some meaningful limit a user might use. This is not a consensus error
	// making the transaction invalid, rather a DOS protection.
	ErrOversizedData = errors.New("oversized data")

	// ErrEmptyActions transaction no actions
	ErrEmptyActions = errors.New("transaction no actions")
)

// Transaction represents an entire transaction in the block.
type Transaction struct {
	actions    []*Action
	gasAssetID uint64
	gasPrice   *big.Int
	// caches
	hash       atomic.Value
	extendHash atomic.Value
	size       atomic.Value
}

// NewTransaction initialize a transaction.
func NewTransaction(assetID uint64, price *big.Int, actions ...*Action) *Transaction {
	tx := &Transaction{
		actions:    actions,
		gasAssetID: assetID,
		gasPrice:   new(big.Int),
	}
	if price != nil {
		tx.gasPrice.Set(price)
	}
	return tx
}

// GasAssetID returns transaction gas asset id.
func (tx *Transaction) GasAssetID() uint64 { return tx.gasAssetID }

func (tx *Transaction) PayerExist() bool {
	return tx.gasPrice.Cmp(big.NewInt(0)) == 0 && tx.actions[0].fp != nil
}

// GasPrice returns transaction Higher gas price .
func (tx *Transaction) GasPrice() *big.Int {
	gasPrice := new(big.Int)
	if tx.gasPrice.Cmp(big.NewInt(0)) == 0 {
		if price := tx.actions[0].PayerGasPrice(); price == nil {
			return big.NewInt(0)
		}
		return gasPrice.Set(tx.actions[0].PayerGasPrice())
	}
	return gasPrice.Set(tx.gasPrice)
}

// Cost returns all actions gasprice * gaslimit.
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int)
	for _, a := range tx.actions {
		total.Add(total, new(big.Int).Mul(tx.gasPrice, new(big.Int).SetUint64(a.Gas())))
	}
	return total
}

// GetActions return transaction actons.
func (tx *Transaction) GetActions() []*Action {
	return tx.actions
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{tx.gasAssetID, tx.gasPrice, tx.actions})
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	var tmpTx struct {
		AssetID  uint64
		GasPrice *big.Int
		Actions  []*Action
	}

	_, size, _ := s.Kind()
	err := s.Decode(&tmpTx)
	if err == nil {
		tx.gasAssetID = tmpTx.AssetID
		tx.gasPrice = tmpTx.GasPrice
		tx.actions = tmpTx.Actions
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}
	return err
}

// Hash hashes the RLP encoding of tx.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	var acts [][]interface{}
	for _, a := range tx.actions {
		acts = append(acts, a.IgnoreExtend())
	}
	v := RlpHash([]interface{}{tx.gasAssetID, tx.gasPrice, acts})
	tx.hash.Store(v)
	return v
}

// ExtensHash hashes the RLP encoding of tx.
func (tx *Transaction) ExtensHash() common.Hash {
	if hash := tx.extendHash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	v := RlpHash(tx)
	tx.extendHash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction,
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	tx.EncodeRLP(&c)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// Check the validity of all fields
func (tx *Transaction) Check(fid uint64, conf *params.ChainConfig) error {
	if len(tx.actions) == 0 {
		return ErrEmptyActions
	}

	// Heuristic limit, reject transactions over 32KB to prfeed DOS attacks
	if tx.Size() > common.StorageSize(params.MaxTxSize) {
		return ErrOversizedData
	}

	if conf.SysTokenID != tx.gasAssetID {
		return fmt.Errorf("only support system asset %d as tx fee", conf.SysTokenID)
	}

	for _, action := range tx.actions {
		if err := action.Check(fid, conf); err != nil {
			return err
		}
	}
	return nil
}

// RPCTransaction that will serialize to the RPC representation of a transaction.
type RPCTransaction struct {
	BlockHash        common.Hash  `json:"blockHash"`
	BlockNumber      uint64       `json:"blockNumber"`
	Hash             common.Hash  `json:"txHash"`
	TransactionIndex uint64       `json:"transactionIndex"`
	RPCActions       []*RPCAction `json:"actions"`
	GasAssetID       uint64       `json:"gasAssetID"`
	GasPrice         *big.Int     `json:"gasPrice"`
	GasCost          *big.Int     `json:"gasCost"`
}

// NewRPCTransaction returns a transaction that will serialize to the RPC.
func (tx *Transaction) NewRPCTransaction(blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransaction {
	result := new(RPCTransaction)
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = blockNumber
		result.TransactionIndex = index
	}
	result.Hash = tx.Hash()
	ras := make([]*RPCAction, len(tx.GetActions()))
	for index, action := range tx.GetActions() {
		ras[index] = action.NewRPCAction(uint64(index))
	}
	result.RPCActions = ras
	result.GasAssetID = tx.gasAssetID
	result.GasPrice = tx.gasPrice
	result.GasCost = tx.Cost()
	return result
}

type RPCTransactionWithPayer struct {
	BlockHash           common.Hash           `json:"blockHash"`
	BlockNumber         uint64                `json:"blockNumber"`
	Hash                common.Hash           `json:"txHash"`
	TransactionIndex    uint64                `json:"transactionIndex"`
	RPCActionsWithPayer []*RPCActionWithPayer `json:"actions"`
	GasAssetID          uint64                `json:"gasAssetID"`
	GasPrice            *big.Int              `json:"gasPrice"`
	GasCost             *big.Int              `json:"gasCost"`
}

func (tx *Transaction) NewRPCTransactionWithPayer(blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransactionWithPayer {
	result := new(RPCTransactionWithPayer)
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = blockNumber
		result.TransactionIndex = index
	}
	result.Hash = tx.Hash()
	ras := make([]*RPCActionWithPayer, len(tx.GetActions()))
	for index, action := range tx.GetActions() {
		ras[index] = action.NewRPCActionWithPayer(uint64(index))
	}
	result.RPCActionsWithPayer = ras
	result.GasAssetID = tx.gasAssetID
	result.GasPrice = tx.gasPrice
	result.GasCost = tx.Cost()
	return result
}

// TxByNonce sort by transaction first action nonce
type TxByNonce []*Transaction

func (s TxByNonce) Len() int { return len(s) }
func (s TxByNonce) Less(i, j int) bool {
	return s[i].GetActions()[0].Nonce() < s[j].GetActions()[0].Nonce()
}
func (s TxByNonce) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// TxByPrice implements both the sort and the heap interface,
type TxByPrice []*Transaction

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].GasPrice().Cmp(s[j].GasPrice()) > 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Push implements heap push.
func (s *TxByPrice) Push(x interface{}) { *s = append(*s, x.(*Transaction)) }

// Pop implements heap pop.
func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

// TxDifference returns a new set which is the difference between a and b.
func TxDifference(a, b []*Transaction) []*Transaction {
	keep := make([]*Transaction, 0, len(a))
	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}
	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}
	return keep
}

// TransactionsByPriceAndNonce represents a set of transactions that can return
// transactions in a profit-maximizing sorted order, while supporting removing
// entire batches of transactions for non-executable accounts.
type TransactionsByPriceAndNonce struct {
	txs   map[common.Name][]*Transaction // Per account nonce-sorted list of transactions
	heads TxByPrice                      // Next transaction for each unique account (price heap)
}

// NewTransactionsByPriceAndNonce creates a transaction set that can retrieve
// price sorted transactions in a nonce-honouring way.
//
// Note, the input map is reowned so the caller should not interact any more with
// if after providing it to the constructor.
func NewTransactionsByPriceAndNonce(txs map[common.Name][]*Transaction) *TransactionsByPriceAndNonce {
	// Initialize a price based heap with the head transactions
	heads := make(TxByPrice, 0, len(txs))
	for from, accTxs := range txs {
		heads = append(heads, accTxs[0])
		// Ensure the sender name is from the signer
		action := accTxs[0].actions[0]
		acc := action.Sender()
		txs[acc] = accTxs[1:]
		if from != acc {
			delete(txs, from)
		}
	}
	heap.Init(&heads)

	// Assemble and return the transaction set
	return &TransactionsByPriceAndNonce{
		txs:   txs,
		heads: heads,
	}
}

// Peek returns the next transaction by price.
func (t *TransactionsByPriceAndNonce) Peek() *Transaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0]
}

// Shift replaces the current best head with the next one from the same account.
func (t *TransactionsByPriceAndNonce) Shift() {
	action := t.heads[0].actions[0]
	acc := action.Sender()
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		t.heads[0], t.txs[acc] = txs[0], txs[1:]
		heap.Fix(&t.heads, 0)
	} else {
		heap.Pop(&t.heads)
	}
}

// Pop removes the best transaction, *not* replacing it with the next one from
// the same account. This should be used when a transaction cannot be executed
// and hence all subsequent ones should be discarded from the same account.
func (t *TransactionsByPriceAndNonce) Pop() {
	heap.Pop(&t.heads)
}
