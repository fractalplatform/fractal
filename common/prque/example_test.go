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
package prque_test

import (
	"fmt"

	"github.com/fractalplatform/fractal/common/prque"
)

// Insert some data into a priority queue and pop them out in prioritized order.
func Example_usage() {
	// Define some data to push into the priority queue
	prio := []int64{77, 22, 44, 55, 11, 88, 33, 99, 0, 66}
	data := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

	// Create the priority queue and insert the prioritized data
	pq := prque.New(nil)
	for i := 0; i < len(data); i++ {
		pq.Push(data[i], prio[i])
	}
	// Pop out the data and print them
	for !pq.Empty() {
		val, prio := pq.Pop()
		fmt.Printf("%d:%s ", prio, val)
	}
	// Output:
	// 99:seven 88:five 77:zero 66:nine 55:three 44:two 33:six 22:one 11:four 0:eight
}
