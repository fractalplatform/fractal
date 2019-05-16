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

package common

import (
	"hash"
	"sync"

	"golang.org/x/crypto/sha3"
)

// pool256 is a freelist for SHA3-256 hash objects.
var pool256 = &sync.Pool{New: func() interface{} { return sha3.NewLegacyKeccak256() }}

// pool512 is a freelist for SHA3-512 hash objects.
var pool512 = &sync.Pool{New: func() interface{} { return sha3.NewLegacyKeccak512() }}

// Get256 returns an initialized SHA3-256 hash ready to use.
// The caller should call Put256 when finished with the returned object.
func Get256() hash.Hash {
	return pool256.Get().(hash.Hash)
}

// Put256 resets h and puts it in the freelist.
func Put256(h hash.Hash) {
	h.Reset()
	pool256.Put(h)
}

// Get512 returns an initialized SHA3-512 hash ready to use.
// The caller should call Put512 when finished with the returned object.
func Get512() hash.Hash {
	return pool512.Get().(hash.Hash)
}

// Put512 resets h and puts it in the freelist.
func Put512(h hash.Hash) {
	h.Reset()
	pool512.Put(h)
}
