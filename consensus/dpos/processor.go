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
	"strings"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/utils/rlp"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

type RegisterCadidate struct {
	Url   string
	Stake *big.Int
}

type UpdateCadidate struct {
	Url   string
	Stake *big.Int
}

type VoteCadidate struct {
	Cadidate string
	Stake    *big.Int
}

type ChangeCadidate struct {
	Cadidate string
}

type RemoveVoter struct {
	Voters []string
}

type KickedCadidate struct {
	Cadidates []string
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
	case types.RegCadidate:
		arg := &RegisterCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.RegCadidate(action.Sender().String(), arg.Url, arg.Stake); err != nil {
			return err
		}
	case types.UpdateCadidate:
		arg := &UpdateCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.UpdateCadidate(action.Sender().String(), arg.Url, arg.Stake); err != nil {
			return err
		}
	case types.UnregCadidate:
		if err := sys.UnregCadidate(action.Sender().String()); err != nil {
			return err
		}
	case types.RemoveVoter:
		arg := &RemoveVoter{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		for _, voter := range arg.Voters {
			if err := sys.UnvoteVoter(action.Sender().String(), voter); err != nil {
				return err
			}
		}
	case types.VoteCadidate:
		arg := &VoteCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.VoteCadidate(action.Sender().String(), arg.Cadidate, arg.Stake); err != nil {
			return err
		}
	case types.ChangeCadidate:
		arg := &ChangeCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		if err := sys.ChangeCadidate(action.Sender().String(), arg.Cadidate); err != nil {
			return err
		}
	case types.UnvoteCadidate:
		if err := sys.UnvoteCadidate(action.Sender().String()); err != nil {
			return err
		}
	case types.KickedCadidate:
		if strings.Compare(action.Sender().String(), dpos.config.SystemName) != 0 {
			return fmt.Errorf("no permission for kicking cadidates")
		}
		arg := &KickedCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return err
		}
		for _, cadicate := range arg.Cadidates {
			if err := sys.KickedCadidate(cadicate); err != nil {
				return err
			}
		}
	case types.ExitTakeOver:
		if strings.Compare(action.Sender().String(), dpos.config.SystemName) != 0 {
			return fmt.Errorf("no permission for exit take over")
		}
		sys.ExitTakeOver()
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
