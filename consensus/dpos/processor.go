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

// RegisterCandidate candidate info
type RegisterCandidate struct {
	URL string
}

// UpdateCandidate candidate info
type UpdateCandidate struct {
	URL string
}

// VoteCandidate vote info
type VoteCandidate struct {
	Candidate string
	Stake     *big.Int
}

// KickedCandidate kicked info
type KickedCandidate struct {
	Candidates []string
}

// RemoveKickedCandidate remove kicked info
type RemoveKickedCandidate struct {
	Candidates []string
}

// ProcessAction exec action
func (dpos *Dpos) ProcessAction(fid uint64, number uint64, chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) ([]*types.InternalAction, error) {
	snap := state.Snapshot()
	internalLogs, err := dpos.processAction(fid, number, chainCfg, state, action)
	if err != nil {
		state.RevertToSnapshot(snap)
	}
	return internalLogs, err
}

func (dpos *Dpos) processAction(fid uint64, number uint64, chainCfg *params.ChainConfig, state *state.StateDB, action *types.Action) ([]*types.InternalAction, error) {
	if err := action.Check(chainCfg); err != nil {
		return nil, err
	}
	sys := NewSystem(state, dpos.config)

	if action.Value().Cmp(big.NewInt(0)) > 0 {
		accountDB, err := accountmanager.NewAccountManager(state)
		if err != nil {
			return nil, err
		}
		if err := accountDB.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value()); err != nil {
			return nil, err
		}
	}
	epoch, err := sys.GetLastestEpoch()
	if err != nil {
		return nil, err
	}
	switch action.Type() {
	case types.RegCandidate:
		if fid >= params.ForkID2 {
			if val := new(big.Int).Mul(dpos.config.CandidateMinQuantity, dpos.config.unitStake()); action.Value().Cmp(val) != 0 {
				return nil, fmt.Errorf("value must be %v", val)
			}
		}
		arg := &RegisterCandidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		if err := sys.RegCandidate(epoch, action.Sender().String(), arg.URL, action.Value(), number, fid); err != nil {
			return nil, err
		}
	case types.UpdateCandidate:
		if fid >= params.ForkID2 {
			if action.Value().Sign() == 1 {
				return nil, fmt.Errorf("value must be zero")
			}
		}
		arg := &UpdateCandidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		if err := sys.UpdateCandidate(epoch, action.Sender().String(), arg.URL, action.Value(), number, fid); err != nil {
			return nil, err
		}
	case types.UnregCandidate:
		if strings.Compare(action.Sender().String(), dpos.config.SystemName) == 0 {
			return nil, fmt.Errorf("no permission")
		}
		err := sys.UnregCandidate(epoch, action.Sender().String(), number, fid)
		if err != nil {
			return nil, err
		}
	case types.RefundCandidate:
		err := sys.RefundCandidate(epoch, action.Sender().String(), number, fid)
		if err != nil {
			return nil, err
		}
	case types.VoteCandidate:
		arg := &VoteCandidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		if err := sys.VoteCandidate(epoch, action.Sender().String(), arg.Candidate, arg.Stake, number, fid); err != nil {
			return nil, err
		}
	case types.KickedCandidate:
		gstate, _ := sys.GetState(epoch)
		if gstate.TakeOver == false || strings.Compare(action.Sender().String(), dpos.config.SystemName) != 0 {
			return nil, fmt.Errorf("no permission for kicking candidates")
		}
		arg := &KickedCandidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		for _, cadicate := range arg.Candidates {
			if strings.Compare(cadicate, dpos.config.SystemName) == 0 {
				continue
			}
			if err := sys.KickedCandidate(epoch, cadicate, number, fid); err != nil {
				return nil, err
			}
		}
	case types.RemoveKickedCandidate:
		if strings.Compare(action.Sender().String(), dpos.config.SystemName) != 0 {
			return nil, fmt.Errorf("no permission for removing candidates")
		}
		arg := &RemoveKickedCandidate{}
		if err := rlp.DecodeBytes(action.Data(), &arg); err != nil {
			return nil, err
		}
		for _, cadicate := range arg.Candidates {
			if strings.Compare(cadicate, dpos.config.SystemName) == 0 {
				continue
			}
			if err := sys.RemoveKickedCandidate(epoch, cadicate, number, fid); err != nil {
				return nil, err
			}
		}

	case types.ExitTakeOver:
		gstate, _ := sys.GetState(epoch)
		if gstate.TakeOver == false || strings.Compare(action.Sender().String(), dpos.config.SystemName) != 0 {
			return nil, fmt.Errorf("no permission for exit take over")
		}
		if err := sys.ExitTakeOver(epoch, number, fid); err != nil {
			return nil, err
		}
	default:
		return nil, accountmanager.ErrUnkownTxType
	}
	return sys.internalActions, nil
}
