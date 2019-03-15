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
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	bc     ChainContext      // Canonical block chain
	engine consensus.IEngine // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(bc ChainContext, engine consensus.IEngine) *StateProcessor {
	return &StateProcessor{
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) ([]*types.Receipt, []*types.Log, uint64, error) {
	var (
		receipts []*types.Receipt
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(common.GasPool).AddGas(block.GasLimit())
	)

	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, _, err := p.ApplyTransaction(nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, nil, 0, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}

	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, block.Transactions(), receipts, statedb)

	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func (p *StateProcessor) ApplyTransaction(author *common.Name, gp *common.GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	bc := p.bc
	config := bc.Config()
	accountDB, err := accountmanager.NewAccountManager(statedb)
	if err != nil {
		return nil, 0, err
	}
	if err := accountDB.RecoverTx(types.NewSigner(config.ChainID), tx); err != nil {
		return nil, 0, err
	}
	assetID := tx.GasAssetID()
	gasPrice := tx.GasPrice()

	var totalGas uint64
	var ios []*types.ActionResult
	detailTx := &types.DetailTx{}
	var internals []*types.InternalTx

	for i, action := range tx.GetActions() {
		if !action.CheckValue() {
			return nil, 0, ErrActionInvalidValue
		}

		nonce, err := accountDB.GetNonce(action.Sender())
		if err != nil {
			return nil, 0, err
		}
		if nonce < action.Nonce() {
			return nil, 0, ErrNonceTooHigh
		} else if nonce > action.Nonce() {
			return nil, 0, ErrNonceTooLow
		}

		evmcontext := &EvmContext{
			ChainContext:  p.bc,
			EgnineContext: p.engine,
		}
		context := NewEVMContext(action.Sender(), assetID, tx.GasPrice(), header, evmcontext, author)
		vmenv := vm.NewEVM(context, accountDB, statedb, config, cfg)

		_, gas, failed, err, vmerr := ApplyMessage(accountDB, vmenv, action, gp, gasPrice, assetID, config, p.engine)
		if err != nil {
			return nil, 0, err
		}

		*usedGas += gas
		totalGas += gas

		var status uint64
		if failed {
			status = types.ReceiptStatusFailed
		} else {
			status = types.ReceiptStatusSuccessful

		}
		vmerrstr := ""
		if vmerr != nil {
			vmerrstr = vmerr.Error()
		}
		var gasAllot []*types.GasDistribution
		for account, gas := range vmenv.FounderGasMap {
			gasAllot = append(gasAllot, &types.GasDistribution{Account: account, Gas: uint64(gas)})
		}
		ios = append(ios, &types.ActionResult{Status: status, Index: uint64(i), GasUsed: gas, GasAllot: gasAllot, Error: vmerrstr})
		internals = append(internals, &types.InternalTx{vmenv.InternalTxs})
	}
	root := statedb.ReceiptRoot()
	receipt := types.NewReceipt(root[:], *usedGas, totalGas)
	receipt.TxHash = tx.Hash()
	receipt.ActionResults = ios
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom([]*types.Receipt{receipt})

	detailTx.TxHash = receipt.TxHash
	detailTx.InternalTxs = internals
	receipt.SetInternalTxsLog(detailTx)
	return receipt, totalGas, nil
}
