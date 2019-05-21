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

package blockchain

import "testing"

func TestMinHeap(t *testing.T) {
	mheap := &simpleHeap{
		cmp: func(a, b interface{}) int {
			return a.(int) - b.(int)
		},
	}
	for i := 0; i < 100; i++ {
		before := mheap.Len()
		mheap.push(i)
		if mheap.min().(int) != 0 {
			t.Fatal("top element not 0!")
		}
		if mheap.Len() != before+1 {
			t.Fatal("push failed!")
		}
	}
	for i := 0; i < 100; i++ {
		before := mheap.Len()
		if mheap.min().(int) != i {
			t.Fatal("top element is not min")
		}
		if mheap.pop().(int) != i {
			t.Fatal("pop is not min")
		}
		if mheap.Len() != before-1 {
			t.Fatal("push failed!")
		}
	}
}

func TestMaxHeap(t *testing.T) {
	mheap := &simpleHeap{
		cmp: func(a, b interface{}) int {
			return b.(int) - a.(int)
		},
	}
	for i := 0; i < 100; i++ {
		before := mheap.Len()
		mheap.push(i)
		if mheap.min().(int) != i {
			t.Fatal("top element not 0!")
		}
		if mheap.Len() != before+1 {
			t.Fatal("push failed!")
		}
	}
	for i := 100; i > 0; i-- {
		before := mheap.Len()
		if mheap.min().(int) != i-1 {
			t.Fatal("top element is not max")
		}
		if mheap.pop().(int) != i-1 {
			t.Fatal("pop is not min")
		}
		if mheap.Len() != before-1 {
			t.Fatal("push failed!")
		}
	}
}
