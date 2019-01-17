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

package cache

import (
	"sort"
	"strings"

	"github.com/fractalplatform/fractal/common"
)

// Account represents an fractal account.
type Account struct {
	Addr common.Address `json:"address"` // account address derived from the key.
	Path string         `json:"path"`    // key json file path.
}

// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
//
func (a Account) Cmp(account Account) int {
	return strings.Compare(a.Path, account.Path)
}

// Accounts  is a Account slice type.
type Accounts []Account

func (as Accounts) drop(accounts ...Account) Accounts {
	for _, account := range accounts {
		index := sort.Search(len(as), func(i int) bool {
			return as[i].Cmp(account) >= 0
		})
		if index == len(as) {
			as = append(as, account)
			continue
		}
		as = append(as[:index], append([]Account{account}, as[index:]...)...)
	}
	return as
}

func (as Accounts) merge(accounts ...Account) Accounts {
	for _, account := range accounts {
		index := sort.Search(len(as), func(i int) bool {
			return as[i].Cmp(account) >= 0
		})
		if index == len(as) {
			continue
		}
		as = append(as[:index], as[index+1:]...)
	}
	return as
}

func (as Accounts) Len() int           { return len(as) }
func (as Accounts) Less(i, j int) bool { return as[i].Cmp(as[j]) < 0 }
func (as Accounts) Swap(i, j int)      { as[i], as[j] = as[j], as[i] }
