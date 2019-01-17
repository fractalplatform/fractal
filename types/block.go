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

// Package types contains data types related to Fractal.
package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"sync/atomic"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
	"golang.org/x/crypto/sha3"
)

// Header represents a block header in the blockchain.
type Header struct {
	ParentHash   common.Hash `json:"parentHash"`
	Coinbase     common.Name `json:"miner"`
	Root         common.Hash `json:"stateRoot"`
	TxsRoot      common.Hash `json:"transactionsRoot"`
	ReceiptsRoot common.Hash `json:"receiptsRoot"`
	Bloom        Bloom       `json:"logsBloom"`
	Difficulty   *big.Int    `json:"difficulty"`
	Number       *big.Int    `json:"number"`
	GasLimit     uint64      `json:"gasLimit"`
	GasUsed      uint64      `json:"gasUsed"`
	Time         *big.Int    `json:"timestamp"`
	Extra        []byte      `json:"extraData"`
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashNoNonce() common.Hash {
	return rlpHash([]interface{}{
		h.ParentHash,
		h.Coinbase,
		h.Root,
		h.TxsRoot,
		h.ReceiptsRoot,
		h.Bloom,
		h.Difficulty,
		h.Number,
		h.GasLimit,
		h.GasUsed,
		h.Time,
		h.Extra,
	})
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// EncodeRLP serializes b into the  RLP block header format.
func (h *Header) EncodeRLP() ([]byte, error) { return rlp.EncodeToBytes(h) }

// DecodeRLP decodes the header
func (h *Header) DecodeRLP(input []byte) error { return rlp.Decode(bytes.NewReader(input), &h) }

// Marshal encodes the web3 RPC block header format.
func (h *Header) Marshal() ([]byte, error) { return json.Marshal(h) }

// Unmarshal decodes the web3 RPC block header format.
func (h *Header) Unmarshal(input []byte) error { return json.Unmarshal(input, h) }

// Block represents an entire block in the blockchain.
type Block struct {
	Head *Header
	Txs  []*Transaction

	// caches
	hash atomic.Value
	size atomic.Value

	// Td is used by package core to store the total difficulty
	// of the chain up to and including the block.
	td *big.Int

	// These fields are used by package eth to track
	// inter-peer block relay.
	receivedAt   time.Time
	receivedFrom interface{}
}

// NewBlock creates a new block. The input data is copied,
// changes to header and to the field values will not affect the
// block.
func NewBlock(header *Header, txs []*Transaction, receipts []*Receipt) *Block {
	b := &Block{Head: header, td: new(big.Int)}

	if len(txs) != len(receipts) {
		panic(fmt.Sprintf("txs len :%v!= receipts len :%v", len(txs), len(receipts)))
	}

	var txHashs, receiptHashs []common.Hash

	for i := 0; i < len(txs); i++ {
		txHashs, receiptHashs = append(txHashs, txs[i].Hash()), append(receiptHashs, receipts[i].Hash())
	}

	b.Head.TxsRoot = common.MerkleRoot(txHashs)
	b.Txs = make([]*Transaction, len(txs))
	copy(b.Txs, txs)

	b.Head.ReceiptsRoot = common.MerkleRoot(receiptHashs)
	b.Head.Bloom = CreateBloom(receipts)

	return b
}

// NewBlockWithHeader creates a block with the given header data. The
// header data is copied, changes to header and to the field values
// will not affect the block.
func NewBlockWithHeader(header *Header) *Block {
	return &Block{Head: CopyHeader(header)}
}
func (b *Block) Transactions() []*Transaction { return b.Txs }
func (b *Block) Number() *big.Int             { return new(big.Int).Set(b.Head.Number) }
func (b *Block) GasLimit() uint64             { return b.Head.GasLimit }
func (b *Block) GasUsed() uint64              { return b.Head.GasUsed }
func (b *Block) Difficulty() *big.Int         { return new(big.Int).Set(b.Head.Difficulty) }
func (b *Block) Time() *big.Int               { return new(big.Int).Set(b.Head.Time) }
func (b *Block) NumberU64() uint64            { return b.Head.Number.Uint64() }
func (b *Block) Coinbase() common.Name        { return b.Head.Coinbase }
func (b *Block) Root() common.Hash            { return b.Head.Root }
func (b *Block) ParentHash() common.Hash      { return b.Head.ParentHash }
func (b *Block) TxHash() common.Hash          { return b.Head.TxsRoot }
func (b *Block) ReceiptHash() common.Hash     { return b.Head.ReceiptsRoot }
func (b *Block) Extra() []byte                { return common.CopyBytes(b.Head.Extra) }
func (b *Block) Header() *Header              { return CopyHeader(b.Head) }
func (b *Block) Body() *Body                  { return &Body{b.Txs} }

func (b *Block) HashNoNonce() common.Hash {
	return b.Head.HashNoNonce()
}

// Size returns the true RLP encoded storage size of the block, either by encoding
// and returning it, or returning a previsouly cached value.
func (b *Block) Size() common.StorageSize {
	if size := b.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, b)
	b.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// EncodeRLP serializes b into the RLP block format.
func (b *Block) EncodeRLP() ([]byte, error) {
	return rlp.EncodeToBytes(b)
}

// DecodeRLP decodes the block
func (b *Block) DecodeRLP(input []byte) error {
	err := rlp.Decode(bytes.NewReader(input), &b)
	if err == nil {
		b.size.Store(common.StorageSize(len(input)))
	}
	return err
}

// Marshal encodes the web3 RPC block format.
func (b *Block) Marshal() ([]byte, error) {
	type Block struct {
		Header       *Header
		Transactions []*Transaction
	}
	var block Block
	block.Header = b.Head
	block.Transactions = b.Txs
	return json.Marshal(block)
}

// WithSeal returns a new block with the data from b but the header replaced with
// the sealed one.
func (b *Block) WithSeal(header *Header) *Block {
	cpy := *header

	return &Block{
		Head: &cpy,
		Txs:  b.Txs,
	}
}

// Hash returns the keccak256 hash of b's header.
// The hash is computed on the first call and cached thereafter.
func (b *Block) Hash() common.Hash {
	if hash := b.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := b.Head.Hash()
	b.hash.Store(v)
	return v
}

// WithBody returns a new block with the given transaction and uncle contents.
func (b *Block) WithBody(transactions []*Transaction) *Block {
	block := &Block{
		Head: CopyHeader(b.Head),
		Txs:  make([]*Transaction, len(transactions)),
	}
	copy(block.Txs, transactions)
	return block
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h
	if cpy.Time = new(big.Int); h.Time != nil {
		cpy.Time.Set(h.Time)
	}
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}

type Body struct {
	Transactions []*Transaction
}

type Blocks []*Block

type BlockBy func(b1, b2 *Block) bool

func (self BlockBy) Sort(blocks Blocks) {
	bs := blockSorter{
		blocks: blocks,
		by:     self,
	}
	sort.Sort(bs)
}

type blockSorter struct {
	blocks Blocks
	by     func(b1, b2 *Block) bool
}

func (self blockSorter) Len() int { return len(self.blocks) }
func (self blockSorter) Swap(i, j int) {
	self.blocks[i], self.blocks[j] = self.blocks[j], self.blocks[i]
}
func (self blockSorter) Less(i, j int) bool { return self.by(self.blocks[i], self.blocks[j]) }

func Number(b1, b2 *Block) bool { return b1.Head.Number.Cmp(b2.Head.Number) < 0 }

func DeriveTxMerkleRoot(txs []*Transaction) common.Hash {
	var txHashs []common.Hash
	for i := 0; i < len(txs); i++ {
		txHashs = append(txHashs, txs[i].Hash())
	}
	return common.MerkleRoot(txHashs)
}

func DeriveReceiPtMerkleRoot(receipts []*Receipt) common.Hash {
	var txHashs []common.Hash
	for i := 0; i < len(receipts); i++ {
		txHashs = append(txHashs, receipts[i].Hash())
	}
	return common.MerkleRoot(txHashs)
}
