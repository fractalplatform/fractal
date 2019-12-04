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
	"fmt"
	"math"
	"testing"
)

func TestAddGas(t *testing.T) {
	gaspool := new(GasPool).AddGas(uint64(1))
	gas := gaspool.AddGas(uint64(1))
	if gas.Gas() != 2 {
		t.Fatalf("result is %s not equal 2.", gas.String())
	}

	defer func() {
		if r := recover(); r != nil {
			if r.(string) != "gas pool pushed above uint64" {
				t.Fatal(r.(string))
			}
			return
		}
		t.Fatal("recover nothing")
	}()
	gaspool = new(GasPool).AddGas(math.MaxUint64)
	gaspool.AddGas(1)
}

func TestSubGas(t *testing.T) {
	gaspool := new(GasPool).AddGas(math.MaxUint64)
	if err := gaspool.SubGas(uint64(1)); err != nil {
		t.Fatal(err)
	}

	if gaspool.Gas() != math.MaxUint64-1 {
		t.Fatalf("result is %s not equal.", gaspool.String())
	}

	gaspool = new(GasPool).AddGas(uint64(1))
	if err := gaspool.SubGas(uint64(2)); err == nil || err != ErrGasLimitReached {
		t.Fatalf("expect ErrGasLimitReached")
	}
	fmt.Println(gaspool)
}
