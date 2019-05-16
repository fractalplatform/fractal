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
	"math"
)

// MerkleRoot return merkle tree root hash.
func MerkleRoot(nodes []Hash) Hash {
	switch {
	case len(nodes) == 0:
		return Hash{}
	case len(nodes) == 1:
		return leafMerkleHash(nodes[0])
	default:
		k := prevPowerOfTwo(len(nodes))
		left := MerkleRoot(nodes[:k])
		right := MerkleRoot(nodes[k:])
		return interiorMerkleHash(left, right)
	}
}

func interiorMerkleHash(left, right Hash) (hash Hash) {
	d := Get256()
	defer Put256(d)
	d.Write(left.Bytes())
	d.Write(right.Bytes())
	d.Sum(hash[:0])
	return hash
}

func leafMerkleHash(node Hash) (hash Hash) {
	d := Get256()
	defer Put256(d)
	d.Write(node.Bytes())
	d.Sum(hash[:0])
	return hash
}

// prevPowerOfTwo returns the largest power of two that is smaller than a given number.
// In other words, for some input n, the prevPowerOfTwo k is a power of two such that
// k < n <= 2k. This is a helper function used during the calculation of a merkle tree.
func prevPowerOfTwo(n int) int {
	// If the number is a power of two, divide it by 2 and return.
	if n&(n-1) == 0 {
		return n / 2
	}

	// Otherwise, find the previous PoT.
	exponent := uint(math.Log2(float64(n)))
	return 1 << exponent // 2^exponent
}
