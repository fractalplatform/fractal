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

package types

import (
	"github.com/fractalplatform/fractal/common"
)

// OptInfo status option info.
type OptInfo struct {
	Key   string
	Value []byte
	Opt   uint // record modification status : add/delete/update
}

// KvNode represents a status data.
type KvNode struct {
	Key   string
	Value []byte
}

// StateOut represents a block exec status data.
type StateOut struct {
	ReadSet    []*KvNode   // replay
	Reverts    []*OptInfo  // rollback previous block
	Changes    []*OptInfo  // forward next block
	ParentHash common.Hash // block parent hash
	Number     uint64      // block num
	Hash       common.Hash // current block hash
}

type AccountInfo struct {
	Name  string
	Key   string
	Value []byte
}
