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

package txpool

import (
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// accountSet is simply a set of name to check for existence
type accountSet struct {
	accounts map[common.Name]struct{}
}

// newAccountSet creates a new name set with an associated signer for sender
// derivations.
func newAccountSet(signer types.Signer) *accountSet {
	return &accountSet{
		accounts: make(map[common.Name]struct{}),
	}
}

// contains checks if a given name is contained within the set.
func (as *accountSet) contains(name common.Name) bool {
	_, exist := as.accounts[name]
	return exist
}

// containsName checks if the sender of a given tx is within the set.
func (as *accountSet) containsName(tx *types.Transaction) bool {
	// todo every action
	return as.contains(tx.GetActions()[0].Sender())
}

// add inserts a new name into the set to track.
func (as *accountSet) add(name common.Name) {
	as.accounts[name] = struct{}{}
}
