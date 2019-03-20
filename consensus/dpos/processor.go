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

package dpos

import (
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/utils/rlp"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

type RegisterProducer struct {
	Url   string
	Stake *big.Int
}

type UpdateProducer struct {
	Url   string
	Stake *big.Int
}

type VoteProducer struct {
	Producer string
	Stake    *big.Int
}

type ChangeProducer struct {
	Producer string
}

type RemoveVoter struct {
	Voter string
}

func (dpos *Dpos) ProcessAction(chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) error {
	snap := state.Snapshot()
	err := dpos.processAction(chainCfg, state, action)
	if err != nil {
		state.RevertToSnapshot(snap)
	}
	return err
}

func (dpos *Dpos) processAction(chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) error {
	sys := &System{
		config: dpos.config,
		IDB: &LDB{
			IDatabase: &stateDB{
				name:    dpos.config.AccountName,
				assetid: chainCfg.SysTokenID,
				state:   state,
			},
		},
	}

	if action.Value().Cmp(big.NewInt(0)) > 0 {
		return fmt.Errorf("invalid action value, must be zero")
	}

	switch action.Type() {
	case types.RegProducer:
		arg := &RegisterProducer{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.RegProducer(action.Sender().String(), arg.Url, arg.Stake); err != nil {
			return err
		}
	case types.UpdateProducer:
		arg := &UpdateProducer{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.UpdateProducer(action.Sender().String(), arg.Url, arg.Stake); err != nil {
			return err
		}
	case types.UnregProducer:
		if err := sys.UnregProducer(action.Sender().String()); err != nil {
			return err
		}
	case types.RemoveVoter:
		arg := &RemoveVoter{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.UnvoteVoter(action.Sender().String(), arg.Voter); err != nil {
			return err
		}
	case types.VoteProducer:
		arg := &VoteProducer{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.VoteProducer(action.Sender().String(), arg.Producer, arg.Stake); err != nil {
			return err
		}
	case types.ChangeProducer:
		arg := &ChangeProducer{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.ChangeProducer(action.Sender().String(), arg.Producer); err != nil {
			return err
		}
	case types.UnvoteProducer:
		if err := sys.UnvoteProducer(action.Sender().String()); err != nil {
			return err
		}
	default:
		return accountmanager.ErrUnkownTxType
	}
	// accountDB, err := accountmanager.NewAccountManager(state)
	// if err != nil {
	// 	return err
	// }
	// if action.Value().Cmp(big.NewInt(0)) > 0 {
	// 	accountDB.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value())
	// }
	return nil
}
