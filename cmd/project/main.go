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

package main

import (
	"fmt"
	"os"

	"github.com/fractalplatform/fractal/cmd/utils"
)

func main() {
	switch os.Args[1] {
	case "changelog":
		fmt.Println(utils.History.MustChangelog())
	case "notes":
		fmt.Println(utils.History.CurrentNotes())
	case "version":
		fmt.Println(utils.History.CurrentVersion().String())
	}
}
