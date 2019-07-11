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
	cache    *[]common.Name
}

// newAccountSet creates a new name set with an associated signer for sender
// derivations.
func newAccountSet(signer types.Signer, names ...common.Name) *accountSet {
	as := &accountSet{accounts: make(map[common.Name]struct{})}
	for _, name := range names {
		as.add(name)
	}
	return as
}

// addTx adds the sender of tx into the set.
func (as *accountSet) addTx(tx *types.Transaction) {
	as.add(tx.GetActions()[0].Sender())
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
	as.cache = nil
}

// flatten returns the list of addresses within this set, also caching it for later
// reuse. The returned slice should not be changed!
func (as *accountSet) flatten() []common.Name {
	if as.cache == nil {
		accounts := make([]common.Name, 0, len(as.accounts))
		for account := range as.accounts {
			accounts = append(accounts, account)
		}
		as.cache = &accounts
	}
	return *as.cache
}

// merge adds all addresses from the 'other' set into 'as'.
func (as *accountSet) merge(other *accountSet) {
	for addr := range other.accounts {
		as.accounts[addr] = struct{}{}
	}
	as.cache = nil
}
