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

package snapshot

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var snapshotManagerName = "sysSnapshot"
var snapshotTime = "time"

// SnapshotManager snapshot manager object
type SnapshotManager struct {
	stateDB *state.StateDB
}

// BlockInfo store stateDB
type BlockInfo struct {
	Number    uint64
	BlockHash common.Hash
	Timestamp uint64 // previous snapshot timestamp
}

// NewSnapshotManager create manager
func NewSnapshotManager(state *state.StateDB) *SnapshotManager {
	return &SnapshotManager{
		stateDB: state,
	}
}

func (sn *SnapshotManager) SetSnapshot(time uint64, blockInfo BlockInfo) error {
	blockInfoEnc, err := rlp.EncodeToBytes(blockInfo)
	if err != nil {
		return err
	}

	timestampEnc, err := rlp.EncodeToBytes(time)
	if err != nil {
		return err
	}

	key := snapshotTime + strconv.FormatUint(time, 10)
	sn.stateDB.Put(snapshotManagerName, key, blockInfoEnc)
	sn.stateDB.Put(snapshotManagerName, snapshotTime, timestampEnc)
	return nil
}

func (sn *SnapshotManager) GetCurrentSnapshotHash() (uint64, common.Hash, error) {
	timestampEnc, err := sn.stateDB.Get(snapshotManagerName, snapshotTime)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	var timestamp uint64
	if err = rlp.DecodeBytes(timestampEnc, &timestamp); err != nil {
		return 0, common.Hash{}, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	key1 := snapshotTime + strconv.FormatUint(timestamp, 10)
	blockInfoEnc, err := sn.stateDB.Get(snapshotManagerName, key1)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Not snapshot info, error = %v", err)
	}
	var blockInfo BlockInfo
	if err = rlp.DecodeBytes(blockInfoEnc, &blockInfo); err != nil {
		return 0, common.Hash{}, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	return blockInfo.Number, blockInfo.BlockHash, nil
}

// GetLastSnapshot get last snapshot time
func (sn *SnapshotManager) GetLastSnapshotTime() (uint64, error) {
	timestampEnc, err := sn.stateDB.Get(snapshotManagerName, snapshotTime)
	if err != nil {
		return 0, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	var timestamp uint64
	if err = rlp.DecodeBytes(timestampEnc, &timestamp); err != nil {
		return 0, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	return timestamp, nil
}

// GetPrevSnapshot get previous snapshot time
func (sn *SnapshotManager) GetPrevSnapshotTime(time uint64) (uint64, error) {
	key := snapshotTime + strconv.FormatUint(time, 10)
	blockInfoEnc, err := sn.stateDB.Get(snapshotManagerName, key)
	if err != nil {
		return 0, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	var blockInfo BlockInfo
	if err = rlp.DecodeBytes(blockInfoEnc, &blockInfo); err != nil {
		return 0, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	return blockInfo.Timestamp, nil
}

func (sn *SnapshotManager) GetSnapshotMsg(account string, key string, time uint64) ([]byte, error) {
	if time == 0 {
		return nil, fmt.Errorf("Not snapshot info, time = %v", time)
	}

	key1 := snapshotTime + strconv.FormatUint(time, 10)
	blockInfoEnc, err := sn.stateDB.Get(snapshotManagerName, key1)
	if err != nil {
		return nil, fmt.Errorf("Not snapshot info, error = %v", err)
	}
	var blockInfo BlockInfo
	if err = rlp.DecodeBytes(blockInfoEnc, &blockInfo); err != nil {
		return nil, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	snapshotBlock := types.SnapshotBlock{
		Number:    blockInfo.Number,
		BlockHash: blockInfo.BlockHash,
	}

	db := sn.stateDB.Database().GetDB()
	snapshotInfo := rawdb.ReadSnapshot(db, snapshotBlock)
	if snapshotInfo == nil {
		return nil, errors.New("Not snapshot info, rawdb not exist")
	}

	dbCache := sn.stateDB.Database()
	statedb, err := state.New(snapshotInfo.Root, dbCache)
	if err != nil {
		return nil, fmt.Errorf("Not snapshot info, new state failed, error = %v", err)
	}

	value, err := statedb.Get(account, key)
	if err != nil {
		return nil, fmt.Errorf("Not snapshot info, msg not exist, error = %v", err)
	}
	return value, nil
}

//GetSnapshotState get snapshot state
func (sn *SnapshotManager) GetSnapshotState(time uint64) (*state.StateDB, error) {
	if time == 0 {
		return nil, fmt.Errorf("Not snapshot info, time = %v", time)
	}

	key1 := snapshotTime + strconv.FormatUint(time, 10)
	blockInfoEnc, err := sn.stateDB.Get(snapshotManagerName, key1)
	if err != nil {
		return nil, fmt.Errorf("Not snapshot info, error = %v", err)
	}
	var blockInfo BlockInfo
	if err = rlp.DecodeBytes(blockInfoEnc, &blockInfo); err != nil {
		return nil, fmt.Errorf("Not snapshot info, error = %v", err)
	}

	snapshotBlock := types.SnapshotBlock{
		Number:    blockInfo.Number,
		BlockHash: blockInfo.BlockHash,
	}

	db := sn.stateDB.Database().GetDB()
	snapshotInfo := rawdb.ReadSnapshot(db, snapshotBlock)
	if snapshotInfo == nil {
		return nil, errors.New("Not snapshot info, rawdb not exist")
	}

	dbCache := sn.stateDB.Database()
	statedb, err := state.New(snapshotInfo.Root, dbCache)
	if err != nil {
		return nil, fmt.Errorf("Not snapshot info, new state failed, error = %v", err)
	}

	return statedb, nil
}
