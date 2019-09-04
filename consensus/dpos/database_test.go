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
	typeList := []CandidateType{Normal, Freeze, Black, Jail, Unkown}
	for _, t := range typeList {
		if text, err := t.MarshalText(); err != nil {
			panic(fmt.Errorf("MarshalText --- %v", err))
		} else {
			var newCt CandidateType
			if err := newCt.UnmarshalText(text); err != nil {
				panic(fmt.Errorf("UnmarshalText --- %v", err))
			} else {
				if newCt != t {
					panic(fmt.Errorf("Var mismatch --- %v", err))
				}
			}
		}
	}
}

func TestCandidateInfo(t *testing.T) {

	candidateInfo := &CandidateInfo{
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
	if candidateInfo.invalid() {
		panic(fmt.Errorf("Candidate Info invalid"))
	}

	var newCi *CandidateInfo
	candidateArray := CandidateInfoArray{}
	candidateArray = append(candidateArray, candidateInfo)

	newCi = candidateInfo.copy()
	newCi.Name = "candidate2"
	candidateArray = append(candidateArray, newCi)

	newCi = candidateInfo.copy()
	newCi.Name = "candidate3"
	candidateArray = append(candidateArray, newCi)

	newCi = candidateInfo.copy()
	newCi.Number = 2
	candidateArray = append(candidateArray, newCi)

	if 4 != candidateArray.Len() {
		panic(fmt.Errorf("CandidateInfoArray Len mismatch"))
	}

	if candidateArray.Less(1, 2) {
		panic(fmt.Errorf("CandidateInfoArray Less mismatch"))
	}
	if !candidateArray.Less(2, 3) {
		panic(fmt.Errorf("CandidateInfoArray Less mismatch"))
	}

	candidateArray.Swap(0, 3)
	if candidateInfo != candidateArray[3] {
		panic(fmt.Errorf("CandidateInfoArray Swap doesnt work"))
	}
}
