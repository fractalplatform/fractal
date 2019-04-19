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
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

type RegisterCadidate struct {
	Url string
}

type UpdateCadidate struct {
	Url   string
	Stake *big.Int
}

type VoteCadidate struct {
	Cadidate string
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

func (dpos *Dpos) ProcessAction(chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) ([]*types.InternalLog, error) {
	snap := state.Snapshot()
	internalLogs, err := dpos.processAction(chainCfg, state, action)
	if err != nil {
		state.RevertToSnapshot(snap)
	}
	return internalLogs, err
}

func (dpos *Dpos) processAction(chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) ([]*types.InternalLog, error) {
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

	var internalLogs []*types.InternalLog

	// if action.CheckValue() {
	// 	return nil, fmt.Errorf("amount value is invalid")
	// }

	if action.AssetID() != chainCfg.SysTokenID {
		return nil, fmt.Errorf("dpos only support system token id %v", chainCfg.SysTokenID)
	}

	if strings.Compare(action.Recipient().String(), dpos.config.AccountName) != 0 {
		return nil, fmt.Errorf("recipient must be %v abount dpos contract", dpos.config.AccountName)
	}

	if action.Value().Cmp(big.NewInt(0)) > 0 {
		accountDB, err := accountmanager.NewAccountManager(state)
		if err != nil {
			return nil, err
		}
		if err := accountDB.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value()); err != nil {
			return nil, err
		}
	}

	switch action.Type() {
	case types.RegCadidate:
		arg := &RegisterCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		if err := sys.RegCadidate(action.Sender().String(), arg.Url, action.Value()); err != nil {
			return nil, err
		}
	case types.UpdateCadidate:
		arg := &UpdateCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		if arg.Stake.Sign() == 1 {
			return nil, fmt.Errorf("stake cannot be greater zero")
		}
		if action.Value().Sign() == 1 && arg.Stake.Sign() == -1 {
			return nil, fmt.Errorf("value & stake cannot allowed at the same time")
		}
		if err := sys.UpdateCadidate(action.Sender().String(), arg.Url, new(big.Int).Add(action.Value(), arg.Stake)); err != nil {
			return nil, err
		}
	case types.UnregCadidate:
		stake, err := sys.UnregCadidate(action.Sender().String())
		if err != nil {
			return nil, err
		}
		actionX := types.NewAction(action.Type(), action.Recipient(), action.Sender(), 0, 0, action.AssetID(), stake, nil)
		internalLog := &types.InternalLog{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
		internalLogs = append(internalLogs, internalLog)
	case types.RemoveVoter:
		arg := &RemoveVoter{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		for _, voter := range arg.Voters {
			stake, err := sys.UnvoteVoter(action.Sender().String(), voter)
			if err != nil {
				return nil, err
			}
			actionX := types.NewAction(action.Type(), action.Recipient(), common.Name(voter), 0, 0, action.AssetID(), stake, nil)
			internalLog := &types.InternalLog{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
			internalLogs = append(internalLogs, internalLog)
		}
	case types.VoteCadidate:
		arg := &VoteCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		if err := sys.VoteCadidate(action.Sender().String(), arg.Cadidate, action.Value()); err != nil {
			return nil, err
		}
	case types.ChangeCadidate:
		arg := &ChangeCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		if err := sys.ChangeCadidate(action.Sender().String(), arg.Cadidate); err != nil {
			return nil, err
		}
	case types.UnvoteCadidate:
		stake, err := sys.UnvoteCadidate(action.Sender().String())
		if err != nil {
			return nil, err
		}
		actionX := types.NewAction(action.Type(), action.Recipient(), action.Sender(), 0, 0, action.AssetID(), stake, nil)
		internalLog := &types.InternalLog{Action: actionX.NewRPCAction(0), ActionType: "", GasUsed: 0, GasLimit: 0, Depth: 0, Error: ""}
		internalLogs = append(internalLogs, internalLog)
	case types.KickedCadidate:
		if strings.Compare(action.Sender().String(), dpos.config.SystemName) != 0 {
			return nil, fmt.Errorf("no permission for kicking cadidates")
		}
		arg := &KickedCadidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		for _, cadicate := range arg.Cadidates {
			if err := sys.KickedCadidate(cadicate); err != nil {
				return nil, err
			}
		}
	case types.ExitTakeOver:
		if strings.Compare(action.Sender().String(), dpos.config.SystemName) != 0 {
			return nil, fmt.Errorf("no permission for exit take over")
		}
		sys.ExitTakeOver()
	default:
		return nil, accountmanager.ErrUnkownTxType
	}
	return internalLogs, nil
}
