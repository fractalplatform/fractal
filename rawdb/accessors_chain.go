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

package rawdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// ReadCanonicalHash retrieves the hash assigned to a canonical block number.
func ReadCanonicalHash(db DatabaseReader, number uint64) common.Hash {
	data, _ := db.Get(headerHashKey(number))
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteCanonicalHash stores the hash assigned to a canonical block number.
func WriteCanonicalHash(db DatabaseWriter, hash common.Hash, number uint64) {
	if err := db.Put(headerHashKey(number), hash.Bytes()); err != nil {
		log.Crit("Failed to store number to hash mapping", "err", err)
	}
}

// DeleteCanonicalHash removes the number to hash canonical mapping.
func DeleteCanonicalHash(db DatabaseDeleter, number uint64) {
	if err := db.Delete(headerHashKey(number)); err != nil {
		log.Crit("Failed to delete number to hash mapping", "err", err)
	}
}

// ReadHeaderNumber returns the header number assigned to a hash.
func ReadHeaderNumber(db DatabaseReader, hash common.Hash) *uint64 {
	data, _ := db.Get(headerNumberKey(hash))
	if len(data) != 8 {
		return nil
	}
	number := binary.BigEndian.Uint64(data)
	return &number
}

// ReadHeadHeaderHash retrieves the hash of the current canonical head header.
func ReadHeadHeaderHash(db DatabaseReader) common.Hash {
	data, _ := db.Get(headHeaderKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadHeaderHash stores the hash of the current canonical head header.
func WriteHeadHeaderHash(db DatabaseWriter, hash common.Hash) {
	if err := db.Put(headHeaderKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last header's hash", "err", err)
	}
}

// ReadIrreversibleNumber retrieves the irreversible number of chain.
func ReadIrreversibleNumber(db DatabaseReader) uint64 {
	data, err := db.Get(irreversibleNumberKey)
	if err != nil {
		log.Crit("Failed to get irreversible number ", "err", err)
	}
	if len(data) == 0 {
		return 0
	}
	return decodeBlockNumber(data)
}

// WriteIrreversibleNumber stores the irreversible number of chain.
func WriteIrreversibleNumber(db DatabaseWriter, number uint64) {
	if err := db.Put(irreversibleNumberKey, encodeBlockNumber(number)); err != nil {
		log.Crit("Failed to store irreversible number ", "err", err)
	}
}

// ReadHeadBlockHash retrieves the hash of the current canonical head block.
func ReadHeadBlockHash(db DatabaseReader) common.Hash {
	data, _ := db.Get(headBlockKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadBlockHash stores the head block's hash.
func WriteHeadBlockHash(db DatabaseWriter, hash common.Hash) {
	if err := db.Put(headBlockKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last block's hash", "err", err)
	}
}

// ReadHeaderRLP retrieves a block header in its raw RLP database encoding.
func ReadHeaderRLP(db DatabaseReader, hash common.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(headerKey(number, hash))
	return data
}

// HasHeader verifies the existence of a block header corresponding to the hash.
func HasHeader(db DatabaseReader, hash common.Hash, number uint64) bool {
	if has, err := db.Has(headerKey(number, hash)); !has || err != nil {
		return false
	}
	return true
}

// ReadHeader retrieves the block header corresponding to the hash.
func ReadHeader(db DatabaseReader, hash common.Hash, number uint64) *types.Header {
	data := ReadHeaderRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	header := new(types.Header)
	if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
		log.Crit("Invalid block header RLP", "hash", hash, "err", err)
		return nil
	}
	return header
}

// WriteHeader stores a block header into the database and also stores the hash-
// to-number mapping.
func WriteHeader(db DatabaseWriter, header *types.Header) {
	// Write the hash -> number mapping
	var (
		hash    = header.Hash()
		number  = header.Number.Uint64()
		encoded = encodeBlockNumber(number)
	)
	key := headerNumberKey(hash)
	if err := db.Put(key, encoded); err != nil {
		log.Crit("Failed to store hash to number mapping", "err", err)
	}
	// Write the encoded header
	data, err := rlp.EncodeToBytes(header)
	if err != nil {
		log.Crit("Failed to RLP encode header", "err", err)
	}
	key = headerKey(number, hash)
	if err := db.Put(key, data); err != nil {
		log.Crit("Failed to store header", "err", err)
	}
}

// DeleteHeader removes all block header data associated with a hash.
func DeleteHeader(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(headerKey(number, hash)); err != nil {
		log.Crit("Failed to delete header", "err", err)
	}
	if err := db.Delete(headerNumberKey(hash)); err != nil {
		log.Crit("Failed to delete hash to number mapping", "err", err)
	}
}

// ReadBodyRLP retrieves the block body (transactions and uncles) in RLP encoding.
func ReadBodyRLP(db DatabaseReader, hash common.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(blockBodyKey(number, hash))
	return data
}

// WriteBodyRLP stores an RLP encoded block body into the database.
func WriteBodyRLP(db DatabaseWriter, hash common.Hash, number uint64, rlp rlp.RawValue) {
	if err := db.Put(blockBodyKey(number, hash), rlp); err != nil {
		log.Crit("Failed to store block body", "err", err)
	}
}

// HasBody verifies the existence of a block body corresponding to the hash.
func HasBody(db DatabaseReader, hash common.Hash, number uint64) bool {
	if has, err := db.Has(blockBodyKey(number, hash)); !has || err != nil {
		return false
	}
	return true
}

// ReadBody retrieves the block body corresponding to the hash.
func ReadBody(db DatabaseReader, hash common.Hash, number uint64) *types.Body {
	data := ReadBodyRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	body := new(types.Body)
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		log.Error("Invalid block body RLP", "hash", hash, "err", err)
		return nil
	}
	return body
}

// WriteBody storea a block body into the database.
func WriteBody(db DatabaseWriter, hash common.Hash, number uint64, body *types.Body) {
	data, err := rlp.EncodeToBytes(body)
	if err != nil {
		log.Crit("Failed to RLP encode body", "err", err)
	}
	WriteBodyRLP(db, hash, number, data)
}

// DeleteBody removes all block body data associated with a hash.
func DeleteBody(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(blockBodyKey(number, hash)); err != nil {
		log.Crit("Failed to delete block body", "err", err)
	}
}

// ReadBlock retrieves an entire block corresponding to the hash, assembling it
// back from the stored header and body. If either the header or body could not
// be retrieved nil is returned.
//
// Note, due to concurrent download of header and block body the header and thus
// canonical hash can be stored in the database but the body data not (yet).
func ReadBlock(db DatabaseReader, hash common.Hash, number uint64) *types.Block {
	header := ReadHeader(db, hash, number)
	if header == nil {
		return nil
	}
	body := ReadBody(db, hash, number)
	if body == nil {
		return nil
	}
	return types.NewBlockWithHeader(header).WithBody(body.Transactions)
}

// WriteBlock serializes a block into the database, header and body separately.
func WriteBlock(db DatabaseWriter, block *types.Block) {
	WriteBody(db, block.Hash(), block.NumberU64(), block.Body())
	WriteHeader(db, block.Header())
}

// DeleteBlock removes all block data associated with a hash.
func DeleteBlock(db DatabaseDeleter, hash common.Hash, number uint64) {
	DeleteReceipts(db, hash, number)
	DeleteDetailTxs(db, hash, number)
	DeleteHeader(db, hash, number)
	DeleteBody(db, hash, number)
	DeleteTd(db, hash, number)
}

// ReadTd retrieves a block's total difficulty corresponding to the hash.
func ReadTd(db DatabaseReader, hash common.Hash, number uint64) *big.Int {
	data, _ := db.Get(headerTDKey(number, hash))
	if len(data) == 0 {
		return nil
	}
	td := new(big.Int)
	if err := rlp.Decode(bytes.NewReader(data), td); err != nil {
		log.Crit("Invalid block total difficulty RLP", "hash", hash, "err", err)
		return nil
	}
	return td
}

// WriteTd stores the total difficulty of a block into the database.
func WriteTd(db DatabaseWriter, hash common.Hash, number uint64, td *big.Int) {
	data, err := rlp.EncodeToBytes(td)
	if err != nil {
		log.Crit("Failed to RLP encode block total difficulty", "err", err)
	}
	if err := db.Put(headerTDKey(number, hash), data); err != nil {
		log.Crit("Failed to store block total difficulty", "err", err)
	}
}

// DeleteTd removes all block total difficulty data associated with a hash.
func DeleteTd(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(headerTDKey(number, hash)); err != nil {
		log.Crit("Failed to delete block total difficulty", "err", err)
	}
}

// ReadReceipts retrieves all the transaction receipts belonging to a block.
func ReadReceipts(db DatabaseReader, hash common.Hash, number uint64) []*types.Receipt {
	// Retrieve the flattened receipt slice
	data, _ := db.Get(blockReceiptsKey(number, hash))
	if len(data) == 0 {
		return nil
	}
	// Convert the revceipts from their storage form to their internal representation
	storageReceipts := []*types.Receipt{}
	if err := rlp.DecodeBytes(data, &storageReceipts); err != nil {
		fmt.Println("Invalid receipt array RLP", "hash", hash.String(), "err", err)
		return nil
	}
	receipts := make([]*types.Receipt, len(storageReceipts))
	for i, receipt := range storageReceipts {
		receipts[i] = (*types.Receipt)(receipt)
	}
	return receipts
}

// WriteReceipts stores all the transaction receipts belonging to a block.
func WriteReceipts(db DatabaseWriter, hash common.Hash, number uint64, receipts []*types.Receipt) {
	// Convert the receipts into their storage form and serialize them
	storageReceipts := make([]*types.Receipt, len(receipts))
	for i, receipt := range receipts {
		storageReceipts[i] = (*types.Receipt)(receipt)
	}
	bytes, err := rlp.EncodeToBytes(storageReceipts)
	if err != nil {
		log.Crit("Failed to encode block receipts", "err", err)
	}
	// Store the flattened receipt slice
	if err := db.Put(blockReceiptsKey(number, hash), bytes); err != nil {
		log.Crit("Failed to store block receipts", "err", err)
	}
}

// DeleteReceipts removes all receipt data associated with a block hash.
func DeleteReceipts(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(blockReceiptsKey(number, hash)); err != nil {
		log.Crit("Failed to delete block receipts", "err", err)
	}
}

// ReadDetailTxs retrieves all the contract log belonging to a block.
func ReadDetailTxs(db DatabaseReader, hash common.Hash, number uint64) []*types.DetailTx {
	// Retrieve the flattened receipt slice
	data, _ := db.Get(blockDetailTxsKey(number, hash))
	if len(data) == 0 {
		return nil
	}
	// Convert the revceipts from their storage form to their internal representation
	storageDetailTxs := []*types.DetailTx{}
	if err := rlp.DecodeBytes(data, &storageDetailTxs); err != nil {
		fmt.Println("Invalid detailtxs array RLP", "hash", hash.String(), "err", err)
		return nil
	}
	detailtxs := make([]*types.DetailTx, len(storageDetailTxs))
	for i, detailtx := range storageDetailTxs {
		detailtxs[i] = (*types.DetailTx)(detailtx)
	}
	return detailtxs
}

// WriteDetailTxs stores all the contract log belonging to a block.
func WriteDetailTxs(db DatabaseWriter, hash common.Hash, number uint64, dtxs []*types.DetailTx) {
	// Convert the receipts into their storage form and serialize them
	storageDetailTxs := make([]*types.DetailTx, len(dtxs))
	for i, dtx := range dtxs {
		storageDetailTxs[i] = (*types.DetailTx)(dtx)
	}
	bytes, err := rlp.EncodeToBytes(storageDetailTxs)
	if err != nil {
		log.Crit("Failed to encode block detailtxs", "err", err)
	}
	// Store the flattened receipt slice
	if err := db.Put(blockDetailTxsKey(number, hash), bytes); err != nil {
		log.Crit("Failed to store block detailtxs", "err", err)
	}
}

// DeleteDetailTxs removes all contract log data associated with a block hash.
func DeleteDetailTxs(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(blockDetailTxsKey(number, hash)); err != nil {
		log.Crit("Failed to delete block detailtxs", "err", err)
	}
}

// FindCommonAncestor returns the last common ancestor of two block headers
func FindCommonAncestor(db DatabaseReader, a, b *types.Header) *types.Header {
	for bn := b.Number.Uint64(); a.Number.Uint64() > bn; {
		a = ReadHeader(db, a.ParentHash, a.Number.Uint64()-1)
		if a == nil {
			return nil
		}
	}
	for an := a.Number.Uint64(); an < b.Number.Uint64(); {
		b = ReadHeader(db, b.ParentHash, b.Number.Uint64()-1)
		if b == nil {
			return nil
		}
	}
	for a.Hash() != b.Hash() {
		a = ReadHeader(db, a.ParentHash, a.Number.Uint64()-1)
		if a == nil {
			return nil
		}
		b = ReadHeader(db, b.ParentHash, b.Number.Uint64()-1)
		if b == nil {
			return nil
		}
	}
	return a
}

//WriteBlockStateOut block state operate for db
func WriteBlockStateOut(db DatabaseWriter, hash common.Hash, stateOut *types.StateOut) {
	data, err := rlp.EncodeToBytes(stateOut)
	if err != nil {
		log.Crit("Failed to RLP encode state out", "err", err)
	}
	if err := db.Put(blockStateOutKey(hash), data); err != nil {
		log.Crit("Failed to store block state out", "err", err)
	}
}

func ReadBlockStateOut(db DatabaseReader, hash common.Hash) *types.StateOut {
	data, _ := db.Get(blockStateOutKey(hash))
	if len(data) == 0 {
		return nil
	}
	stateOut := new(types.StateOut)
	if err := rlp.Decode(bytes.NewReader(data), stateOut); err != nil {
		log.Crit("Invalid block state RLP", "hash", hash, "err", err)
		return nil
	}
	return stateOut
}

func DeleteBlockStateOut(db DatabaseDeleter, hash common.Hash) {
	if err := db.Delete(blockStateOutKey(hash)); err != nil {
		log.Crit("Failed to delete block state", "err", err)
	}
}

func WriteOptBlockHash(db DatabaseWriter, hash common.Hash) {
	if err := db.Put(blockOptHash, hash.Bytes()); err != nil {
		log.Crit("Failed to store last opt block's hash", "err", err)
	}
}

func ReadOptBlockHash(db DatabaseReader) common.Hash {
	data, _ := db.Get(blockOptHash)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

func WriteSnapshot(db DatabaseWriter, key types.SnapshotBlock, snapshotInfo types.SnapshotInfo) {
	keyEnc, err := rlp.EncodeToBytes(key)
	if err != nil {
		log.Crit("Failed to RLP encode snapshotBlock", "err", err)
	}

	snapshotInfoEnc, err := rlp.EncodeToBytes(snapshotInfo)
	if err != nil {
		log.Crit("Failed to RLP encode snapshotInfo", "err", err)
	}

	if err := db.Put(blockSnapshotKey(keyEnc), snapshotInfoEnc); err != nil {
		log.Crit("Failed to store block snapshotInfo", "err", err)
	}
}

func ReadSnapshot(db DatabaseReader, key types.SnapshotBlock) *types.SnapshotInfo {
	keyEnc, err := rlp.EncodeToBytes(key)
	if err != nil {
		log.Crit("Failed to RLP encode snapshotBlock", "err", err)
	}

	snapshotInfoEnc, _ := db.Get(blockSnapshotKey(keyEnc))
	if len(snapshotInfoEnc) == 0 {
		return nil
	}

	snapshotInfo := new(types.SnapshotInfo)
	if err := rlp.Decode(bytes.NewReader(snapshotInfoEnc), snapshotInfo); err != nil {
		log.Crit("Invalid block snapshotInfo RLP", "number", key.Number, "hash", key.BlockHash, "err", err)
		return nil
	}
	return snapshotInfo
}
