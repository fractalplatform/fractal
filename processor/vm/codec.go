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

package vm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/fractalplatform/fractal/utils/abi"
)

//var AbifilePath = "./system.abi"

// OutputsPack
func OutputsPack(abifile string, method string, params ...interface{}) ([]byte, error) {
	var abicode string

	hexcode, err := ioutil.ReadFile(abifile)
	if err != nil {
		fmt.Printf("Could not load code from file: %v\n", err)
		return nil, err
	}
	abicode = string(bytes.TrimRight(hexcode, "\n"))

	parsed, err := abi.JSON(strings.NewReader(abicode))
	if err != nil {
		fmt.Println("abi.json error ", err)
		return nil, err
	}

	input, err := parsed.OutputsPack(method, params...)
	if err != nil {
		fmt.Println("parsed.pack error ", err)
		return nil, err
	}
	return input, nil
}

// InputsUnPack
func InputsUnPack(abifile string, method string, v interface{}, input []byte) error {
	var abicode string

	hexcode, err := ioutil.ReadFile(abifile)
	if err != nil {
		fmt.Printf("Could not load code from file: %v\n", err)
		return err
	}
	abicode = string(bytes.TrimRight(hexcode, "\n"))

	parsed, err := abi.JSON(strings.NewReader(abicode))
	if err != nil {
		fmt.Println("abi.json error ", err)
		return err
	}

	err = parsed.InputsUnPack(v, method, input)
	if err != nil {
		fmt.Println("parsed.pack error ", err)
		return err
	}
	return nil
}
