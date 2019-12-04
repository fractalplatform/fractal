// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	ErrWrongAction   = errors.New("action is invalid")
	ErrWrongContract = errors.New("contract is invalid")
)

// Manager manage all plugins.
type Manager struct {
	stateDB         *state.StateDB
	contracts       map[string]IContract
	contractsByType map[types.ActionType]IContract
	IAccount
	IAsset
	IConsensus
	IFee
	ISigner
}

func (pm *Manager) BasicCheck(tx *types.Transaction) error {
	for _, action := range tx.GetActions() {
		switch action.Type() {
		case CreateAccount:
			param := &CreateAccountAction{}
			if err := rlp.DecodeBytes(action.Data(), param); err != nil {
				return err
			}
			// if action.Recipient() != chainCfg.AccountName {
			// 	return fmt.Errorf("Receipt should is %v", chainCfg.AccountName)
			// }
			if action.Recipient() != "fractalaccount" {
				return fmt.Errorf("Receipt should is fractalaccount")
			}
			if err := pm.checkCreateAccount(param.Name, param.Pubkey, param.Desc); err != nil {
				return err
			}
		case IssueAsset:
			param := &IssueAssetAction{}
			if err := rlp.DecodeBytes(action.Data(), param); err != nil {
				return err
			}
			// if action.Recipient() != chainCfg.AssetName {
			// 	return fmt.Errorf("Receipt should is %v", chainCfg.AssetName)
			// }
			if action.Recipient() != "fractalasset" {
				return fmt.Errorf("Receipt should is fractalasset")
			}
			if err := pm.checkIssueAsset(action.Sender(), param.AssetName, param.Symbol, param.Amount, param.Decimals, param.Founder, param.Owner, param.UpperLimit, param.Description, pm.IAccount); err != nil {
				return err
			}
		case IncreaseAsset:
			param := &IncreaseAssetAction{}
			if err := rlp.DecodeBytes(action.Data(), param); err != nil {
				return err
			}
			// if action.Recipient() != chainCfg.AssetName {
			// 	return fmt.Errorf("Receipt should is %v", chainCfg.AssetName)
			// }
			if action.Recipient() != "fractalasset" {
				return fmt.Errorf("Receipt should is fractalasset")
			}
			if err := pm.checkIncreaseAsset(action.Sender(), param.To, param.AssetID, param.Amount, pm.IAccount); err != nil {
				return err
			}
		default:
			if action.Type() != Transfer && (action.Type() < RegisterMiner || action.Type() >= ConsensusEnd) {
				return ErrWrongAction
			}
		}
	}
	return nil
}

func (pm *Manager) selectContract(action *types.Action) IContract {
	if contract, exist := pm.contracts[action.Recipient()]; exist {
		return contract
	}
	return pm.contractsByType[action.Type()]
}

func (pm *Manager) ExecTx(arg interface{}) ([]byte, error) {
	action, ok := arg.(*types.Action)
	if !ok {
		return nil, ErrWrongAction
	}

	if contract := pm.selectContract(action); contract != nil {
		snapshot := pm.stateDB.Snapshot()
		ret, err := contract.CallTx(action, pm)
		if err != nil {
			pm.stateDB.RevertToSnapshot(snapshot)
		}
		return ret, err
	}
	return nil, ErrWrongContract
}

// NewPM create new plugin manager.
func NewPM(stateDB *state.StateDB) IPM {
	acm, _ := NewACM(stateDB)
	asm, _ := NewASM(stateDB)
	consensus := NewConsensus(stateDB)
	chainID := big.NewInt(1)
	signer, _ := NewSigner(chainID)
	fee, _ := NewFeeManager()
	pm := &Manager{
		contracts:       make(map[string]IContract),
		contractsByType: make(map[types.ActionType]IContract),
		IAccount:        acm,
		IAsset:          asm,
		IConsensus:      consensus,
		ISigner:         signer,
		IFee:            fee,
		stateDB:         stateDB,
	}
	pm.contracts[acm.AccountName()] = acm
	pm.contracts[asm.AccountName()] = asm
	pm.contracts[consensus.AccountName()] = consensus
	pm.contractsByType[Transfer] = acm
	return pm
}
