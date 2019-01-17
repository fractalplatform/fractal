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

// Package filelock provides portable file locking.
package filelock

import (
	"os"
	"path/filepath"
)

// Releaser provides the Release method to release a file lock.
type Releaser interface {
	Release() error
}

// New locks the file with the provided name. If the file does not exist, it is
// created. The returned Releaser is used to release the lock. existed is true
// if the file to lock already existed. A non-nil error is returned if the
// locking has failed. Neither this function nor the returned Releaser is
// goroutine-safe.
func New(fileName string) (r Releaser, existed bool, err error) {
	if err = os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
		return
	}

	_, err = os.Stat(fileName)
	existed = err == nil

	r, err = newLock(fileName)
	return
}
