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

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	ErrWrongAction = errors.New("action is invalid")
)

// Manager manage all plugins.
type Manager struct {
	IAccount
	IAsset
	IConsensus
	IContract
	IFee
	ISigner
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
		} else {
			return pm.CreateAccount(param.Name, param.Pubkey, param.Desc)
		}
	case IssueAsset:
		param := &IssueAssetAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		} else {
			return pm.IssueAsset(action.Sender(), param.AssetName, param.Symbol, param.Amount, param.Decimals, param.Founder, param.Owner, param.UpperLimit, param.Description, pm.IAccount)
		}
	case IncreaseAsset:
		param := &IncreaseAssetAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		} else {
			return pm.IncreaseAsset(action.Sender(), param.To, param.AssetID, param.Amount, pm.IAccount)
		}
	default:
		return nil, nil
	}
}

// NewPM create new plugin manager.
func NewPM(stateDB *state.StateDB) IPM {
	acm, _ := NewACM(stateDB)
	asm, _ := NewASM(stateDB)
	return &Manager{
		IAccount: acm,
		IAsset:   asm,
	}
}
