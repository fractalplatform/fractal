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

package blockchain

import (
	"errors"
	"fmt"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

const (
	forkInfo = "forkInfo"
)

// ForkConfig fork config
type ForkConfig struct {
	ForkBlockNum   uint64
	Forkpercentage uint64
}

// ForkInfo store in state db.
type ForkInfo struct {
	CurForkID          uint64
	NextForkID         uint64
	CurForkIDBlockNum  uint64
	NextForkIDBlockNum uint64
}

// ForkController control the hard forking.
type ForkController struct {
	cfg      *ForkConfig
	chainCfg *params.ChainConfig
}

// NewForkController return a new fork controller.
func NewForkController(cfg *ForkConfig, chaincfg *params.ChainConfig) *ForkController {
	return &ForkController{cfg: cfg, chainCfg: chaincfg}
}

func initForkController(chainName string, statedb *state.StateDB, curforkID uint64) error {
	if curforkID > params.NextForkID {
		return fmt.Errorf("not support fork ID: %v,the last fork ID: %v", curforkID, params.NextForkID)
	}

	var info = ForkInfo{CurForkID: curforkID, NextForkID: curforkID}
	infoBytes, err := statedb.Get(chainName, forkInfo)
	if err != nil {
		return err
	}

	if len(infoBytes) == 0 {
		infoRlp, err := rlp.EncodeToBytes(info)
		if err != nil {
			return err
		}
		statedb.Put(chainName, forkInfo, infoRlp)
	}
	return nil
}

func (fc *ForkController) getForkInfo(statedb *state.StateDB) (ForkInfo, error) {
	info := ForkInfo{}

	infoBytes, err := statedb.Get(fc.chainCfg.ChainName, forkInfo)
	if err != nil {
		return info, err
	}

	if len(infoBytes) == 0 {
		return info, errors.New("not found info in statedb")
	}

	if err := rlp.DecodeBytes(infoBytes, &info); err != nil {
		return info, err
	}
	return info, nil
}

func (fc *ForkController) putForkInfo(info ForkInfo, statedb *state.StateDB) error {
	infoBytes, err := rlp.EncodeToBytes(info)
	if err != nil {
		return err
	}

	statedb.Put(fc.chainCfg.ChainName, forkInfo, infoBytes)
	return nil
}

func (fc *ForkController) update(block *types.Block, statedb *state.StateDB, getHeader func(number uint64) *types.Header) error {
	info, err := fc.getForkInfo(statedb)
	if err != nil {
		return err
	}

	// treat older version as oldest version
	if block.CurForkID() != block.NextForkID() && info.NextForkID <= block.NextForkID() {
		if info.NextForkID < block.NextForkID() {
			// update next forkID
			info.NextForkID = block.NextForkID()
			info.NextForkIDBlockNum = 0
		}

		info.NextForkIDBlockNum++
		if info.CurForkIDBlockNum+info.NextForkIDBlockNum >= fc.cfg.ForkBlockNum {
			header := getHeader(block.NumberU64() + 1 - fc.cfg.ForkBlockNum)
			if header.NextForkID() == info.NextForkID {
				info.NextForkIDBlockNum--
			} else {
				info.CurForkIDBlockNum--
			}
		}
	} else {
		info.CurForkIDBlockNum++
		if info.CurForkIDBlockNum+info.NextForkIDBlockNum >= fc.cfg.ForkBlockNum {
			header := getHeader(block.NumberU64() + 1 - fc.cfg.ForkBlockNum)
			if header.NextForkID() == info.NextForkID {
				if info.NextForkIDBlockNum != 0 {
					info.NextForkIDBlockNum--
				} else {
					info.CurForkIDBlockNum--
				}
			} else {
				info.CurForkIDBlockNum--
			}
		}
	}

	if info.NextForkIDBlockNum*100/fc.cfg.ForkBlockNum >= fc.cfg.Forkpercentage {
		info.CurForkID = block.NextForkID()
		info.CurForkIDBlockNum = info.NextForkIDBlockNum
		info.NextForkIDBlockNum = 0
	}

	return fc.putForkInfo(info, statedb)
}

func (fc *ForkController) currentForkID(statedb *state.StateDB) (uint64, uint64, error) {
	info, err := fc.getForkInfo(statedb)
	if err != nil {
		return 0, 0, err
	}
	return info.CurForkID, params.NextForkID, nil
}

func (fc *ForkController) checkForkID(header *types.Header, state *state.StateDB) error {
	// check current fork id and next fork id
	if curForkID, _, err := fc.currentForkID(state); err != nil {
		return err
	} else if header.CurForkID() != curForkID || header.NextForkID() < curForkID {
		return fmt.Errorf("invild header curForkID: %v, header nextForkID: %v,actual curForkID %v, header hash: %v, header number: %v",
			header.CurForkID(), header.NextForkID(), curForkID, header.Hash().Hex(), header.Number.Uint64())
	}
	return nil
}

func (fc *ForkController) fillForkID(header *types.Header, state *state.StateDB) error {
	// check current fork id and next fork id
	curForkID, nextForkID, err := fc.currentForkID(state)
	if err != nil {
		return err
	}
	header.WithForkID(curForkID, nextForkID)
	return nil
}
