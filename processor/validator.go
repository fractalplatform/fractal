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

package processor

import (
	"fmt"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

var allowedFutureBlockTime = 15 * time.Second

// BlockValidator is responsible for validating block headers, body and
// processed state.
//
// BlockValidator implements Validator.
type BlockValidator struct {
	bc     ChainContext         // Canonical block chain
	engine consensus.IValidator // Consensus engine used for validating
}

// NewBlockValidator returns a new block validator which is safe for re-use
func NewBlockValidator(blockchain ChainContext, engine consensus.IValidator) *BlockValidator {
	validator := &BlockValidator{
		engine: engine,
		bc:     blockchain,
	}
	return validator
}

// ValidateHeader checks whether a header conforms to the consensus rules of the
// stock engine.
func (v *BlockValidator) ValidateHeader(header *types.Header, seal bool) error {
	// Short circuit if the header is known, or it's parent not
	if v.bc.HasBlockAndState(header.Hash(), header.Number.Uint64()) {
		return ErrKnownBlock
	}

	number := header.Number.Uint64()
	parent := v.bc.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return errParentBlock
	}

	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}

	if header.Time.Cmp(big.NewInt(time.Now().Add(allowedFutureBlockTime).UnixNano())) > 0 {
		return ErrFutureBlock
	}

	if header.Time.Cmp(parent.Time) <= 0 {
		return errZeroBlockTime
	}
	// Verify the block's difficulty based in it's timestamp and parent's difficulty
	expected := v.engine.CalcDifficulty(v.bc, header.Time.Uint64(), parent)

	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.GasLimit / params.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return ErrInvalidNumber
	}

	if !v.bc.HasBlockAndState(header.ParentHash, header.Number.Uint64()-1) {
		if !v.bc.HasBlock(header.ParentHash, header.Number.Uint64()-1) {
			return ErrUnknownAncestor
		}
		return ErrPrunedAncestor
	}

	// Checks the validity of forkID
	if err := v.bc.CheckForkID(header); err != nil {
		return err
	}

	// Verify the engine specific seal securing the block
	if seal {
		if err := v.engine.VerifySeal(v.bc, header); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBody verifies the the block header's transaction roots.
// The headers are assumed to be already validated at this point.
func (v *BlockValidator) ValidateBody(block *types.Block) error {
	if err := block.Check(); err != nil {
		return err
	}

	if block.CurForkID() >= params.ForkID4 {
		// Header validity is known at this point, check the uncles and transactions
		if hash := types.DeriveExtensTxsMerkleRoot(block.Txs); hash != block.TxHash() {
			return fmt.Errorf("transaction root hash mismatch: have %x, want %x", hash, block.TxHash())
		}
	} else {
		// Header validity is known at this point, check the uncles and transactions
		if hash := types.DeriveTxsMerkleRoot(block.Txs); hash != block.TxHash() {
			return fmt.Errorf("transaction root hash mismatch: have %x, want %x", hash, block.TxHash())
		}
	}
	return nil
}

// ValidateState validates the various changes that happen after a state
// transition, such as amount of used gas, the receipt roots and the state root
// itself. ValidateState returns a database batch if the validation was a success
// otherwise nil and an error is returned.
func (v *BlockValidator) ValidateState(block, parent *types.Block, statedb *state.StateDB, receipts []*types.Receipt, usedGas uint64) error {
	header := block.Header()
	if block.GasUsed() != usedGas {
		return fmt.Errorf("invalid gas used (remote: %d local: %d)", block.GasUsed(), usedGas)
	}
	// Validate the received block's bloom with the one derived from the generated receipts.
	// For valid blocks this should always validate to true.
	rbloom := types.CreateBloom(receipts)
	if rbloom != header.Bloom {
		return fmt.Errorf("invalid bloom (remote: %x  local: %x)", header.Bloom, rbloom)
	}
	// Tre receipt Trie's root (R = (Tr [[H1, R1], ... [Hn, R1]]))
	receiptSha := types.DeriveReceiptsMerkleRoot(receipts)
	if receiptSha != header.ReceiptsRoot {
		return fmt.Errorf("invalid receipt root hash (remote: %x local: %x)", header.ReceiptsRoot, receiptSha)
	}
	// Validate the state root against the received state root and throw
	// an error if they don't match.
	if root := statedb.IntermediateRoot(); header.Root != root {
		return fmt.Errorf("invalid merkle root (remote: %x local: %x)", header.Root, root)
	}
	return nil
}
