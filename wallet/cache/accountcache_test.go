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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	cache := NewAccountCache(d)

	var accs Accounts
	for i := 0; i < 7; i++ {
		addr := crypto.CreateAddress(common.Address{}, uint64(i))
		account := Account{Addr: addr, Path: keyFileName(addr)}
		cache.Add(account)
		accs = append(accs, account)
	}

	wantAccounts := make(Accounts, len(accs))
	copy(wantAccounts, accs)
	sort.Sort(wantAccounts)

	// test cache.Accounts
	list := cache.Accounts()
	assert.Equal(t, wantAccounts, list)

	// test cache.Has
	for _, a := range accs {
		assert.Equal(t, true, cache.Has(a.Addr))
	}

	for i := 0; i < len(accs); i += 2 {
		cache.Delete(wantAccounts[i].Addr)
	}

	// Check content again after deletion.
	wantAccountsAfterDelete := Accounts{
		wantAccounts[1],
		wantAccounts[3],
		wantAccounts[5],
	}

	list = cache.Accounts()
	assert.Equal(t, wantAccountsAfterDelete, list)

}

func keyFileName(keyAddr common.Address) string {
	toISO8601 := func(t time.Time) string {
		var tz string
		name, offset := t.Zone()
		if name == "UTC" {
			tz = "Z"
		} else {
			tz = fmt.Sprintf("%03d00", offset/3600)
		}
		return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
	}

	return fmt.Sprintf("UTC--%s--%s", toISO8601(time.Now().UTC()), hex.EncodeToString(keyAddr[:]))
}
