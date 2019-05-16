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
	"fmt"
	"io"
	"math/big"
	"sort"
	"sync/atomic"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// ForkID  represents a blockchain fork
type ForkID struct {
	Cur  uint64 `json:"cur"`
	Next uint64 `json:"next"`
}

// Header represents a block header in the blockchain.
type Header struct {
	ParentHash           common.Hash
	Coinbase             common.Name
	ProposedIrreversible uint64
	Root                 common.Hash
	TxsRoot              common.Hash
	ReceiptsRoot         common.Hash
	Bloom                Bloom
	Difficulty           *big.Int
	Number               *big.Int
	GasLimit             uint64
	GasUsed              uint64
	Time                 *big.Int
	Extra                []byte
	ForkID               ForkID
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *Header) Hash() common.Hash { return RlpHash(h) }

// WithForkID store fork id
func (h *Header) WithForkID(cur, next uint64) {
	h.ForkID = ForkID{Cur: cur, Next: next}
}

// CurForkID returns the header's current fork ID.
func (h *Header) CurForkID() uint64 { return h.ForkID.Cur }

// NextForkID returns the header's next fork ID.
func (h *Header) NextForkID() uint64 { return h.ForkID.Next }

// Block represents an entire block in the blockchain.
type Block struct {
	Head *Header
	Txs  []*Transaction

	// caches
	hash atomic.Value
	size atomic.Value
}

// "external" block encoding. used protocol, etc.
type extblock struct {
	Header *Header
	Txs    []*Transaction
}

// NewBlock creates a new block. The input data is copied,
// changes to header and to the field values will not affect the
// block.
func NewBlock(header *Header, txs []*Transaction, receipts []*Receipt) *Block {
	if len(txs) != len(receipts) {
		panic(fmt.Sprintf("txs len :%v!= receipts len :%v", len(txs), len(receipts)))
	}

	b := &Block{Head: header}
	b.Head.TxsRoot = DeriveTxsMerkleRoot(txs)
	b.Head.ReceiptsRoot = DeriveReceiptsMerkleRoot(receipts)

	b.Txs = make([]*Transaction, len(txs))
	copy(b.Txs, txs)
	b.Head.Bloom = CreateBloom(receipts)

	return b
}

// NewBlockWithHeader creates a block with the given header data. The
// header data is copied, changes to header and to the field values
// will not affect the block.
func NewBlockWithHeader(header *Header) *Block {
	return &Block{Head: CopyHeader(header)}
}

// Transactions returns the block's txs.
func (b *Block) Transactions() []*Transaction { return b.Txs }

// Number returns the block's Number.
func (b *Block) Number() *big.Int { return new(big.Int).Set(b.Head.Number) }

// GasLimit returns the block's GasLimit.
func (b *Block) GasLimit() uint64 { return b.Head.GasLimit }

// GasUsed returns the block's GasUsed.
func (b *Block) GasUsed() uint64 { return b.Head.GasUsed }

// Difficulty returns the block's Difficulty.
func (b *Block) Difficulty() *big.Int { return new(big.Int).Set(b.Head.Difficulty) }

// Time returns the block's Time.
func (b *Block) Time() *big.Int { return new(big.Int).Set(b.Head.Time) }

// NumberU64 returns the block's NumberU64.
func (b *Block) NumberU64() uint64 { return b.Head.Number.Uint64() }

// Coinbase returns the block's Coinbase.
func (b *Block) Coinbase() common.Name { return b.Head.Coinbase }

// Root returns the block's Root.
func (b *Block) Root() common.Hash { return b.Head.Root }

// ParentHash returns the block's ParentHash.
func (b *Block) ParentHash() common.Hash { return b.Head.ParentHash }

// TxHash returns the block's TxHash.
func (b *Block) TxHash() common.Hash { return b.Head.TxsRoot }

// ReceiptHash returns the block's ReceiptHash.
func (b *Block) ReceiptHash() common.Hash { return b.Head.ReceiptsRoot }

// Extra returns the block's Extra.
func (b *Block) Extra() []byte { return common.CopyBytes(b.Head.Extra) }

// Header returns the block's Header.
func (b *Block) Header() *Header { return CopyHeader(b.Head) }

// Body returns the block's Body.
func (b *Block) Body() *Body { return &Body{b.Txs} }

// CurForkID returns the block's current fork ID.
func (b *Block) CurForkID() uint64 { return b.Head.CurForkID() }

// NextForkID returns the block's current fork ID.
func (b *Block) NextForkID() uint64 { return b.Head.NextForkID() }

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

// EncodeRLP serializes b into RLP block format.
func (b *Block) ExtEncodeRLP(w io.Writer) error {
	return rlp.Encode(w, extblock{
		Header: b.Head,
		Txs:    b.Txs,
	})
}

// DecodeRLP decodes the block
func (b *Block) DecodeRLP(input []byte) error {
	err := rlp.Decode(bytes.NewReader(input), &b)
	if err == nil {
		b.size.Store(common.StorageSize(len(input)))
	}
	return err
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

// WithBody returns a new block with the given transaction .
func (b *Block) WithBody(transactions []*Transaction) *Block {
	block := &Block{
		Head: CopyHeader(b.Head),
		Txs:  make([]*Transaction, len(transactions)),
	}
	copy(block.Txs, transactions)
	return block
}

// Check the validity of all fields
func (b *Block) Check() error {
	for _, tx := range b.Txs {
		if len(tx.actions) == 0 {
			return ErrEmptyActions
		}
	}
	return nil
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

// Body represents an entire body of the block transactions.
type Body struct {
	Transactions []*Transaction
}

// Blocks represents the blocks.
type Blocks []*Block

// BlockBy represents the block sort by rule type.
type BlockBy func(b1, b2 *Block) bool

// Sort sort blocks by BlockBy.
func (bb BlockBy) Sort(blocks Blocks) {
	bs := blockSorter{
		blocks: blocks,
		by:     bb,
	}
	sort.Sort(bs)
}

type blockSorter struct {
	blocks Blocks
	by     func(b1, b2 *Block) bool
}

func (bs blockSorter) Len() int           { return len(bs.blocks) }
func (bs blockSorter) Swap(i, j int)      { bs.blocks[i], bs.blocks[j] = bs.blocks[j], bs.blocks[i] }
func (bs blockSorter) Less(i, j int) bool { return bs.by(bs.blocks[i], bs.blocks[j]) }

// Number represents block sort by number.
func Number(b1, b2 *Block) bool { return b1.Head.Number.Cmp(b2.Head.Number) < 0 }

// DeriveTxsMerkleRoot returns txs merkle tree root hash.
func DeriveTxsMerkleRoot(txs []*Transaction) common.Hash {
	var txHashs []common.Hash
	for i := 0; i < len(txs); i++ {
		txHashs = append(txHashs, txs[i].Hash())
	}
	return common.MerkleRoot(txHashs)
}

// DeriveReceiptsMerkleRoot returns receiptes merkle tree root hash.
func DeriveReceiptsMerkleRoot(receipts []*Receipt) common.Hash {
	var txHashs []common.Hash
	for i := 0; i < len(receipts); i++ {
		txHashs = append(txHashs, receipts[i].ConsensusReceipt().Hash())
	}
	return common.MerkleRoot(txHashs)
}

func RlpHash(x interface{}) (h common.Hash) {
	hw := common.Get256()
	defer common.Put256(hw)
	err := rlp.Encode(hw, x)
	if err != nil {
		panic(fmt.Sprintf("rlp hash encode err: %v", err))
	}
	hw.Sum(h[:0])
	return h
}

type BlockState struct {
	PreStatePruning bool   `json:"preStatePruning"`
	CurrentNumber   uint64 `json:"currentNumber"`
}
