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

	"github.com/fractalplatform/fractal/common"
)

var (
	ErrSideBlock = errors.New("write side block and state")

	ErrNoGenesis = errors.New("Genesis not found in chain")

	errReorgSystemBlock = errors.New("not reorg system block")

	errGenesisNoConfig = errors.New("genesis has no chain configuration")

	// ErrBlacklistedHash is returned if a block to import is on the blacklist.
	ErrBlacklistedHash = errors.New("blacklisted hash")
)

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database already contains an incompatible genesis block (have %x, new %x)", e.Stored[:], e.New[:])
}
