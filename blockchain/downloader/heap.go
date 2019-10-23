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

import "container/heap"

type simpleHeap struct {
	cmp  func(interface{}, interface{}) int
	data []interface{}
}

func (s *simpleHeap) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *simpleHeap) Less(i, j int) bool {
	return s.cmp(s.data[i], s.data[j]) < 0
}

func (s *simpleHeap) Push(v interface{}) {
	s.data = append(s.data, v)
}

func (s *simpleHeap) Pop() interface{} {
	v := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return v
}

func (s *simpleHeap) Len() int {
	return len(s.data)
}

func (s *simpleHeap) pop() interface{} {
	if len(s.data) == 0 {
		return nil
	}
	return heap.Pop(s)
}

func (s *simpleHeap) push(v interface{}) {
	heap.Push(s, v)
}

func (s *simpleHeap) remove(i int) {
	if len(s.data) > i {
		heap.Remove(s, i)
	}
}

func (s *simpleHeap) min() interface{} {
	if len(s.data) == 0 {
		return nil
	}
	return s.data[0]
}

func (s *simpleHeap) clear() {
	s.data = nil
}
