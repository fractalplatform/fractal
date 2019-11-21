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
	"math/big"

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	ErrWrongAction = errors.New("action is invalid")
)

// Manager manage all plugins.
type Manager struct {
	stateDB *state.StateDB
	IAccount
	IAsset
	IConsensus
	IContract
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
			if err := pm.checkCreateAccount(param.Name, param.Pubkey, param.Desc); err != nil {
				return err
			}
		case IssueAsset:
			param := &IssueAssetAction{}
			if err := rlp.DecodeBytes(action.Data(), param); err != nil {
				return err
			}
			if err := pm.checkIssueAsset(action.Sender(), param.AssetName, param.Symbol, param.Amount, param.Decimals, param.Founder, param.Owner, param.UpperLimit, param.Description, pm.IAccount); err != nil {
				return err
			}
		case IncreaseAsset:
			param := &IncreaseAssetAction{}
			if err := rlp.DecodeBytes(action.Data(), param); err != nil {
				return err
			}
			if err := pm.checkIncreaseAsset(action.Sender(), param.To, param.AssetID, param.Amount, pm.IAccount); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func (pm *Manager) ExecTx(arg interface{}) ([]byte, error) {
	action, ok := arg.(*types.Action)
	if !ok {
		return nil, ErrWrongAction
	}
	switch action.Type() {
	case CreateAccount:
		param := &CreateAccountAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		return pm.CreateAccount(param.Name, param.Pubkey, param.Desc)
	case IssueAsset:
		param := &IssueAssetAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		return pm.IssueAsset(action.Sender(), param.AssetName, param.Symbol, param.Amount, param.Decimals, param.Founder, param.Owner, param.UpperLimit, param.Description, pm.IAccount)
	case IncreaseAsset:
		param := &IncreaseAssetAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		return pm.IncreaseAsset(action.Sender(), param.To, param.AssetID, param.Amount, pm.IAccount)
	case Transfer:
		err := pm.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value())
		return nil, err
	default:
		if action.Type() >= RegisterMiner && action.Type() < ConsensusEnd {
			snapshot := pm.stateDB.Snapshot()
			ret, err := pm.IConsensus.CallTx(action, pm)
			if err != nil {
				pm.stateDB.RevertToSnapshot(snapshot)
			}
			return ret, err
		}
		return nil, ErrWrongAction
	}
}

// NewPM create new plugin manager.
func NewPM(stateDB *state.StateDB) IPM {
	acm, _ := NewACM(stateDB)
	asm, _ := NewASM(stateDB)
	consensus := NewConsensus(stateDB)
	chainId := big.NewInt(1)
	signer, _ := NewSigner(chainId)
	fee, _ := NewFeeManager()
	return &Manager{
		IAccount:   acm,
		IAsset:     asm,
		IConsensus: consensus,
		ISigner:    signer,
		IFee:       fee,
		stateDB:    stateDB,
	}
}
