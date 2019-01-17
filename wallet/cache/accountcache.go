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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
)

const reloadInterval = 2 * time.Second

// AccountCache is a live index of all accounts in the keystore.
type AccountCache struct {
	sync.Mutex
	keydir      string
	cache       fileCache
	reloadTimer *time.Timer
	accountsMap map[common.Address]Account
}

// NewAccountCache creates a account cache.
func NewAccountCache(keydir string) *AccountCache {
	ac := &AccountCache{
		keydir:      keydir,
		cache:       fileCache{all: make(map[string]interface{})},
		accountsMap: make(map[common.Address]Account),
	}
	return ac
}

// Accounts returns all accounts in cacahe.
func (ac *AccountCache) Accounts() Accounts {
	ac.Reload()
	ac.Lock()
	defer ac.Unlock()
	var cpy Accounts
	for _, a := range ac.accountsMap {
		cpy = append(cpy, a)
	}
	sort.Sort(cpy)
	return cpy
}

// Reload reload cache.
func (ac *AccountCache) Reload() {
	ac.Lock()
	if ac.reloadTimer == nil {
		ac.reloadTimer = time.NewTimer(0)
	} else {
		select {
		case <-ac.reloadTimer.C:
		default:
			ac.Unlock()
			return // The cache was reloaded recently.
		}
	}
	ac.reloadTimer.Reset(reloadInterval)
	ac.Unlock()
	ac.updateAccounts()
}

// Find find account by address.
func (ac *AccountCache) Find(addr common.Address) *Account {
	ac.Reload()
	ac.Lock()
	defer ac.Unlock()
	a, ok := ac.accountsMap[addr]
	if !ok {
		return nil
	}
	return &a
}

// Close close account cache.
func (ac *AccountCache) Close() {
	ac.Lock()
	if ac.reloadTimer != nil {
		ac.reloadTimer.Stop()
	}
	ac.Unlock()
}

// Has check whether a key with the given address in cache.
func (ac *AccountCache) Has(addr common.Address) bool {
	ac.Reload()
	ac.Lock()
	defer ac.Unlock()
	if _, ok := ac.accountsMap[addr]; ok {
		return true
	}
	return false
}

// Add add a new acount in cache.
func (ac *AccountCache) Add(new Account) {
	ac.Lock()
	defer ac.Unlock()
	ac.accountsMap[new.Addr] = new
}

// Delete delete a account in cache.
func (ac *AccountCache) Delete(addr common.Address) {
	ac.Lock()
	defer ac.Unlock()
	delete(ac.accountsMap, addr)
}

func (ac *AccountCache) updateAccounts() error {
	creates, deletes, updates, err := ac.cache.scan(ac.keydir)
	if err != nil {
		return err
	}

	if len(creates) == 0 && len(deletes) == 0 && len(updates) == 0 {
		return nil
	}

	for _, path := range creates {
		if a := readAccount(path); a != nil {
			ac.Add(*a)
		}
	}
	for _, path := range deletes {
		if a := readAccount(path); a != nil {
			ac.Delete(a.Addr)
		}
	}
	for _, path := range updates {
		if a := readAccount(path); a != nil {
			ac.Add(*a)
		}
	}

	return nil
}

func readAccount(path string) *Account {
	if len(path) < 40 {
		log.Error("invalid path", "path", path)
		return nil
	}
	if common.IsHexAddress(path[len(path)-40:]) {
		return &Account{
			Addr: common.HexToAddress(path[len(path)-40:]),
			Path: path,
		}
	}
	return nil
}
