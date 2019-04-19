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
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/fractalplatform/fractal/utils/rlp"
)

func TestDatabase(t *testing.T) {
	// gstate
	gstate := &globalState{
		Height:                          10,
		ActivatedCadidateScheduleUpdate: uint64(time.Now().UnixNano()),
		ActivatedCadidateSchedule:       []string{},
		ActivatedTotalQuantity:          big.NewInt(1000),
		TotalQuantity:                   big.NewInt(100000),
	}
	gval, err := rlp.EncodeToBytes(gstate)
	if err != nil {
		panic(fmt.Sprintf("gstate EncodeToBytes--%v", err))
	}

	ngstate := &globalState{}
	if err := rlp.DecodeBytes(gval, ngstate); err != nil {
		panic(fmt.Sprintf("gstate DecodeBytes--%v", err))
	}
	gjson, _ := json.Marshal(ngstate)
	t.Log("state     ", string(gjson))

	// voter
	vote := &voterInfo{
		Name:     "fos",
		Cadidate: "fos",
		Quantity: big.NewInt(1000),
	}

	vval, err := rlp.EncodeToBytes(vote)
	if err != nil {
		panic(fmt.Sprintf("vote EncodeToBytes--%v", err))
	}

	nvote := &voterInfo{}
	if err := rlp.DecodeBytes(vval, nvote); err != nil {
		panic(fmt.Sprintf("vote DecodeBytes--%v", err))
	}

	vjson, _ := json.Marshal(nvote)
	t.Log("voter     ", string(vjson))

	// cadidate
	prod := &cadidateInfo{
		Name:          "fos",
		URL:           "www.fractalproject.com",
		Quantity:      big.NewInt(1000),
		TotalQuantity: big.NewInt(1000),
	}

	pval, err := rlp.EncodeToBytes(prod)
	if err != nil {
		panic(fmt.Sprintf("prod EncodeToBytes--%v", err))
	}

	nprod := &cadidateInfo{}
	if err := rlp.DecodeBytes(pval, nprod); err != nil {
		panic(fmt.Sprintf("prod DecodeBytes--%v", err))
	}

	pjson, _ := json.Marshal(nprod)
	t.Log("prod     ", string(pjson))

	prod1 := &cadidateInfo{
		Name:          "fos1",
		URL:           "www.fractalproject.com",
		Quantity:      big.NewInt(1000),
		TotalQuantity: big.NewInt(2000),
	}

	prod2 := &cadidateInfo{
		Name:          "fos2",
		URL:           "www.fractalproject.com",
		Quantity:      big.NewInt(1000),
		TotalQuantity: big.NewInt(1000),
	}

	prods := cadidateInfoArray{}
	prods = append(prods, prod)
	prods = append(prods, prod1)
	prods = append(prods, prod2)

	psjson, _ := json.Marshal(prods)
	t.Log("prods     ", string(psjson))
	sort.Sort(prods)
	spsjson, _ := json.Marshal(prods)
	t.Log("prods sort ", string(spsjson))

}
