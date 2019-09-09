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

//VM is a Virtual Machine based on Ethereum Virtual Machine
package fvm

import (
	"fmt"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/fvm/native"
)

func isNativeContract(contract common.Name) bool {
	return true
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.

func CallNative(contract common.Name, method string, params ...interface{}) ([]byte, error) {
	if c := native.NativeContracts[contract]; c != nil {
		return c.Run(method, params)
	}
	return nil, fmt.Errorf("native contract %s not exist", contract)
}
