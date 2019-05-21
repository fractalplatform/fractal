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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestGenGenesis(t *testing.T) {
	genesis := DefaultGenesis()

	j, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(fmt.Sprintf("genesis marshal --- %v", err))
	}

	//fmt.Println(string(j))

	if err := json.Unmarshal(j, &genesis); err != nil {
		t.Fatal(fmt.Sprintf("genesis Unmarshal --- %v", err))
	}
	nj, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(fmt.Sprintf("genesis marshal --- %v", err))
	}

	if !bytes.Equal(j, nj) {
		t.Fatal(fmt.Sprintf("genesis mismatch --- %v", err))
	}
}
