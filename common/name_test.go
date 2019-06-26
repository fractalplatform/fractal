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
	"fmt"
	"regexp"
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
		{"5aaeb6053f3e", false},
		{"测试名称", false},
		{"hello_world", false},
		{"hello world", false},
		{"Helloworld", false},
		{"short", false},
		{"longnamelongnamelongnamelongname", false},
	}
	length := uint64(31)
	reg := regexp.MustCompile(`^[a-z][a-z0-9]{6,16}(\.[a-z][a-z0-9]{0,16}){0,2}$`)
	for _, test := range tests {
		if result := StrToName(test.str).IsValid(reg, length); result != test.exp {
			t.Errorf("IsValidAccountName(%s) == %v; expected %v, len:%v",
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
		{"测试名称", false},
		{"hello_world", false},
		{"hello world", false},
		{"Helloworld", false},
		{"short", false},
		{"longnamelongnamelongnamelongname", false},
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

func TestIsChildren(t *testing.T) {
	acctRegExp := regexp.MustCompile(`^([a-z][a-z0-9]{6,15})(?:\.([a-z0-9]{1,8})){0,1}$`)

	type fields struct {
		from Name
		acct Name
		reg  *regexp.Regexp
	}

	tests := []struct {
		name   string
		fields fields
		exp    bool
	}{
		{"include", fields{StrToName("helloworld"), StrToName("helloworld.wert"), acctRegExp}, true},
		{"include2", fields{StrToName("helloworld.wert2"), StrToName("helloworld.wert2"), acctRegExp}, false},
		{"uninclude", fields{StrToName("helloworld"), StrToName("hellowordx.wert"), acctRegExp}, false},
		// {"longnamelongname", true},
		// {"5aaeb6053f3e", false},
		// {"测试名称", false},
		// {"hello_world", false},
		// {"hello world", false},
		// {"Helloworld", false},
		// {"short", false},
		// {"longnamelongnamelongnamelongname", false},
	}

	//eg := regexp.MustCompile("^[a-z][a-z0-9]{6,16}(\.[a-z][a-z0-9]{0,16}){0,2}$")
	for _, tt := range tests {

		if result := tt.fields.from.IsChildren(tt.fields.acct); result != tt.exp {
			t.Errorf("%q. Account.GetNonce() = %v, want %v", tt.name, result, tt.exp)

		}
	}
}

func TestIsChildren1(t *testing.T) {
	acctRegExp := regexp.MustCompile(`^([a-z][a-z0-9]{6,15})(?:\.([a-z0-9]{2,16})){0,1}(?:\.([a-z0-9]{2,16})){0,1}$`)

	type fields struct {
		from Name
		acct Name
		reg  *regexp.Regexp
	}

	tests := []struct {
		name   string
		fields fields
		exp    bool
	}{
		{"include1", fields{StrToName("abc5678.abc5678"), StrToName("abc4567.abc5678.abc5678"), acctRegExp}, false},
		{"include2", fields{StrToName("abc4567.abc5678.abc5678"), StrToName("abc5678.abc5678"), acctRegExp}, false},
		{"include3", fields{StrToName("abc4567.abc4567.abc4567"), StrToName("abc4567.abc4567.abc4567"), acctRegExp}, false},
		{"include4", fields{StrToName("abc4567.abc4567"), StrToName("abc4567.abc5678"), acctRegExp}, false},
		{"include5", fields{StrToName("abc4567"), StrToName("abc4567.abc"), acctRegExp}, true},
		{"include6", fields{StrToName("abc4567"), StrToName("abc4567.abc4567"), acctRegExp}, true},
		{"include7", fields{StrToName("abc4567"), StrToName("abc4567.a"), acctRegExp}, false},
		{"include8", fields{StrToName("abc5678.abc5678"), StrToName("abc456.abc5678.abc5678"), acctRegExp}, false},
		{"include9", fields{StrToName("abc5678.abc5678"), StrToName("abc4567.abc5678.abc5678"), acctRegExp}, false},
		{"include10", fields{StrToName("abc4567"), StrToName("abc4567"), acctRegExp}, false},
	}

	for _, tt := range tests {
		if len(FindStringSubmatch(tt.fields.reg, tt.fields.acct.String())) > 1 {
			if result := tt.fields.from.IsChildren(tt.fields.acct); result != tt.exp {
				t.Errorf("%q. Account.GetNonce() = %v, want %v", tt.name, result, tt.exp)
			}
		} else {
			t.Log(fmt.Sprintf("%q, %v, want %v", tt.name, tt.fields.acct.String(), tt.exp))
			if tt.exp == true {
				t.Errorf("%q, want %v", tt.name, tt.exp)
			}
		}
	}

}
