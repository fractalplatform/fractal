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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type fileCache struct {
	all          map[string]interface{}
	lastModified time.Time
	mu           sync.RWMutex
}

func (fc *fileCache) scan(keyDir string) (creates, deletes, updates []string, err error) {
	// List all the failes from the keystore folder
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return nil, nil, nil, err
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Iterate all the files and gather their metadata
	all := make(map[string]interface{})
	mods := make(map[string]interface{})
	createsmap := make(map[string]interface{})

	var newLastMod time.Time
	for _, fi := range files {
		path := filepath.Join(keyDir, fi.Name())
		if !isKeyFile(fi) {
			log.Trace("Ignoring file on account scan", "path", path)
			continue
		}

		all[path] = struct{}{}

		modified := fi.ModTime()
		if modified.After(fc.lastModified) {
			mods[path] = struct{}{}
		}
		if modified.After(newLastMod) {
			newLastMod = modified
		}
	}

	// Update the tracked files and return the three path slice
	deletes = difference(fc.all, all) // Deletes = previous - current
	creates = difference(all, fc.all) // Creates = current - previous

	for _, v := range creates {
		createsmap[v] = struct{}{}
	}

	updates = difference(mods, createsmap) // Updates = modified - creates

	fc.all, fc.lastModified = all, newLastMod

	return creates, deletes, updates, nil
}

// difference Returns the difference between x and y. The returned string slice
// will contain all elements of x that are not also elements of y.
func difference(x, y map[string]interface{}) []string {
	var result []string
	for elem := range x {
		if _, ok := y[elem]; !ok {
			result = append(result, elem)
		}
	}
	return result
}

func isKeyFile(fi os.FileInfo) bool {
	if strings.HasSuffix(fi.Name(), "~") || strings.HasPrefix(fi.Name(), ".") {
		return false
	}
	if fi.IsDir() || fi.Mode()&os.ModeType != 0 {
		return false
	}
	return true
}
