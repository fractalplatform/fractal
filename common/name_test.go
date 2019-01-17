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
	"encoding/json"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		str string
		exp bool
	}{
		{"helloworld", true},
		{"shortnam", true},
		{"longnamelongname", true},
		{"5aaeb6053f3e", true},
		{"测试名称", false},
		{"hello_world", false},
		{"hello world", false},
		{"Helloworld", false},
		{"short", false},
		{"longnamelongnamelongnamelongname", false},
	}

	for _, test := range tests {
		if result := IsValidName(test.str); result != test.exp {
			t.Errorf("IsValidName(%s) == %v; expected %v, len:%v",
				test.str, result, test.exp, len(test.str))
		}
	}
}

func TestNameUnmarshalJSON(t *testing.T) {
	var tests = []struct {
		Input     string
		ShouldErr bool
	}{
		{"helloworld", false},
		{"shortnam", false},
		{"longnamelongname", false},
		{"5aaeb6053f3e", false},
		{"测试名称", true},
		{"hello_world", true},
		{"hello world", true},
		{"Helloworld", true},
		{"short", true},
		{"longnamelongnamelongnamelongname", true},
	}
	for i, test := range tests {

		bytes, err := json.Marshal(test.Input)
		if err != nil {
			t.Fatal(err)
		}

		var v Name

		err = json.Unmarshal(bytes, &v)
		if err != nil && !test.ShouldErr {
			t.Errorf("test #%d: unexpected error: %v", i, err)
		}

		if err == nil {
			if test.ShouldErr {
				t.Errorf("test #%d: expected error, got none", i)
			}
			if v.String() != test.Input {
				t.Errorf("test #%d: Name mismatch: have %v, want %v", i, v.String(), test.Input)
			}
		}
	}
}
