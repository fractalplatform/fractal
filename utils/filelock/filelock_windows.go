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

import "syscall"

type windowsLock struct {
	fd syscall.Handle
}

func (fl *windowsLock) Release() error {
	return syscall.Close(fl.fd)
}

func newLock(fileName string) (Releaser, error) {
	pathp, err := syscall.UTF16PtrFromString(fileName)
	if err != nil {
		return nil, err
	}
	fd, err := syscall.CreateFile(pathp, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil, syscall.CREATE_ALWAYS, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	return &windowsLock{fd}, nil
}
