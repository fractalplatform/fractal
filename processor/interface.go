// Copyright 2018 The OEX Team Authors
// This file is part of the OEX project.
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
	"github.com/oexplatform/oexchain/common"
	"github.com/oexplatform/oexchain/processor/vm"
	"github.com/oexplatform/oexchain/state"
	"github.com/oexplatform/oexchain/types"
)

// Validator is an interface which defines the standard for block validation. It
// is only responsible for validating block contents, as the header validation is
// done by the specific consensus engines.
type Validator interface {
	// ValidateHeader validates the given header's content.
	ValidateHeader(header *types.Header, seal bool) error

	// ValidateBody validates the given block's content.
	ValidateBody(block *types.Block) error

	// ValidateState validates the given statedb and optionally the receipts and
	// gas used.
	ValidateState(block, parent *types.Block, state *state.StateDB, receipts []*types.Receipt, usedGas uint64) error
}

// Processor is an interface for processing blocks using a given initial state.
type Processor interface {
	Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) ([]*types.Receipt, []*types.Log, uint64, error)
	ApplyTransaction(author *common.Name, gp *common.GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error)
}
