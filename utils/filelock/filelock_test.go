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
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

var datadirInUseErrnos = map[uint]bool{11: true, 32: true, 35: true}

// Tests that if the dir is already in use, an error is returned.
func TestFileLock(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary data directory: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	release, _, err := New(filepath.Join(dir, "LOCK"))
	if err != nil {
		t.Fatal(err)
	}

	defer release.Release()

	_, _, err = New(filepath.Join(dir, "LOCK"))
	if err != nil {
		if errno, ok := err.(syscall.Errno); !ok && !datadirInUseErrnos[uint(errno)] {
			t.Fatal(err)
		}
	}
}
