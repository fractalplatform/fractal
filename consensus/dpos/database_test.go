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

package dpos

import (
	"fmt"
	"math/big"
	"testing"
)

func TestCandidateType(t *testing.T) {
	type_list := []CandidateType{Normal, Freeze, Black, Jail, Unkown}
	for _, t := range type_list {
		if text, err := t.MarshalText(); err != nil {
			panic(fmt.Errorf("MarshalText --- %v", err))
		} else {
			var new_ct CandidateType
			if err := new_ct.UnmarshalText(text); err != nil {
				panic(fmt.Errorf("UnmarshalText --- %v", err))
			} else {
				if new_ct != t {
					panic(fmt.Errorf("Var mismatch --- %v", err))
				}
			}
		}
	}
}

func TestCandidateInfo(t *testing.T) {

	candidate_info := &CandidateInfo{
		Epoch:         1,
		Name:          "candidate1",
		URL:           "",
		Quantity:      big.NewInt(100),
		TotalQuantity: big.NewInt(100),
		Number:        1,
		Counter:       10,
		ActualCounter: 10,
		Type:          Normal,
		PrevKey:       "-",
		NextKey:       "-",
	}
	if candidate_info.invalid() {
		panic(fmt.Errorf("Candidate Info invalid"))
	}

	var new_ci *CandidateInfo
	candidate_array := CandidateInfoArray{}
	candidate_array = append(candidate_array, candidate_info)

	new_ci = candidate_info.copy()
	new_ci.Name = "candidate2"
	candidate_array = append(candidate_array, new_ci)

	new_ci = candidate_info.copy()
	new_ci.Name = "candidate3"
	candidate_array = append(candidate_array, new_ci)

	new_ci = candidate_info.copy()
	new_ci.Number = 2
	candidate_array = append(candidate_array, new_ci)

	if 4 != candidate_array.Len() {
		panic(fmt.Errorf("CandidateInfoArray Len mismatch"))
	}

	if candidate_array.Less(1, 2) {
		panic(fmt.Errorf("CandidateInfoArray Less mismatch"))
	}
	if !candidate_array.Less(2, 3) {
		panic(fmt.Errorf("CandidateInfoArray Less mismatch"))
	}

	candidate_array.Swap(0, 3)
	if candidate_info != candidate_array[3] {
		panic(fmt.Errorf("CandidateInfoArray Swap doesnt work"))
	}
}
