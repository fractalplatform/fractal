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
	"sort"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
type StateProcessor struct {
	bc ChainContext // Canonical block chain
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(bc ChainContext) *StateProcessor {
	return &StateProcessor{
		bc: bc,
	}
}

// Process processes the state changes according to the rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) .
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

	manager := plugin.NewPM(statedb)
	manager.Init(0, p.bc.GetHeaderByHash(header.ParentHash))

	// Prepare the block, applying any consensus engine specific extras (e.g. update last)
	if err := manager.Prepare(header); err != nil {
		return nil, nil, 0, err
	}

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
	manager.Finalize(header, block.Transactions(), receipts)
	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func (p *StateProcessor) ApplyTransaction(author *string, gp *common.GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	bc := p.bc
	config := bc.ChainConfig()
	pm := plugin.NewPM(statedb)
	pm.Init(0, p.bc.GetHeaderByHash(header.ParentHash))

	// todo for the moment，only system asset
	// assetID := tx.GasAssetID()
	assetID := p.bc.ChainConfig().SysTokenID
	if assetID != tx.GetGasAssetID() {
		return nil, 0, fmt.Errorf("only support system asset %d as tx fee", p.bc.ChainConfig().SysTokenID)
	}
	gasPrice := tx.GetGasPrice()
	//timer for vm exec overtime
	var t *time.Timer
	var totalGas uint64
	detailTx := &types.DetailTx{}

	nonce, err := pm.GetNonce(tx.Sender())
	if err != nil {
		return nil, 0, err
	}
	if nonce < tx.GetNonce() {
		return nil, 0, ErrNonceTooHigh
	} else if nonce > tx.GetNonce() {
		return nil, 0, ErrNonceTooLow
	}

	context := NewEVMContext(tx.Sender(), tx.Recipient(), assetID, tx.GetGasPrice(), header, p.bc, author)
	vmenv := vm.NewEVM(context, pm, statedb, config, cfg)

	//will abort the vm if overtime
	if false == cfg.EndTime.IsZero() {
		t = time.AfterFunc(cfg.EndTime.Sub(time.Now()), func() {
			vmenv.OverTimeAbort()
		})
	}

	pluginContext := plugin.NewContext(p.bc, header)

	_, gas, gasAllots, failed, err, vmerr := ApplyMessage(pm, vmenv, pluginContext, tx, gp, gasPrice, assetID, config)

	if false == cfg.EndTime.IsZero() {
		//close timer
		t.Stop()
	}

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
		log.Debug("processer apply transaction ", "hash", tx.Hash(), "err", vmerrstr)
	}

	internalTxs := make([]*types.InternalTx, 0, len(vmenv.InternalTxs))
	for _, itx := range vmenv.InternalTxs {
		internalTxs = append(internalTxs, itx)
	}

	sort.Sort(types.Distributes(gasAllots))

	root := statedb.ReceiptRoot()
	receipt := types.NewReceipt(root[:], *usedGas, totalGas)
	receipt.TxHash = tx.Hash()
	receipt.Status = status
	receipt.GasUsed = gas
	receipt.GasAllot = gasAllots
	receipt.Error = vmerrstr
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom([]*types.Receipt{receipt})

	detailTx.TxHash = receipt.TxHash
	detailTx.InternalTxs = internalTxs
	receipt.SetInternalTxsLog(detailTx)
	return receipt, totalGas, nil
}
