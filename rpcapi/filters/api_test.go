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

package filters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fractalplatform/fractal/common"
)

func TestUnmarshalJSONNewFilterArgs(t *testing.T) {
	var (
		account0 = common.Name("fractal.founder")
		account1 = common.Name("supportmytest")
		topic0   = common.HexToHash("3ac225168df54212a25c1c01fd35bebfea408fdac2e31ddd6f80a4bbf9a5f1ca")
		topic1   = common.HexToHash("9084a792d2f8b16a62b882fd56f7860c07bf5fa91dd8a2ae7e809e5180fef0b3")
		topic2   = common.HexToHash("6ccae1c4af4152f460ff510e573399795dfab5dcf1fa60d1f33ac8fdc1e480ce")
	)

	// default values
	var test0 FilterCriteria
	if err := json.Unmarshal([]byte("{}"), &test0); err != nil {
		t.Fatal(err)
	}
	if len(test0.Accounts) != 0 {
		t.Fatalf("expected 0 addresses, got %d", len(test0.Accounts))
	}
	if len(test0.Topics) != 0 {
		t.Fatalf("expected 0 topics, got %d topics", len(test0.Topics))
	}

	// single address
	var test2 FilterCriteria
	vector := fmt.Sprintf(`{"accounts": ["%s"]}`, account0.String())
	if err := json.Unmarshal([]byte(vector), &test2); err != nil {
		t.Fatal(err)
	}
	if len(test2.Accounts) != 1 {
		t.Fatalf("expected 1 address, got %d address(es)", len(test2.Accounts))
	}
	if test2.Accounts[0] != account0 {
		t.Fatalf("expected address %x, got %x", account0, test2.Accounts[0])
	}

	// multiple address
	var test3 FilterCriteria
	vector = fmt.Sprintf(`{"accounts": ["%s", "%s"]}`, account0.String(), account1.String())
	if err := json.Unmarshal([]byte(vector), &test3); err != nil {
		t.Fatal(err)
	}
	if len(test3.Accounts) != 2 {
		t.Fatalf("expected 2 addresses, got %d address(es)", len(test3.Accounts))
	}
	if test3.Accounts[0] != account0 {
		t.Fatalf("expected address %x, got %x", account0, test3.Accounts[0])
	}
	if test3.Accounts[1] != account1 {
		t.Fatalf("expected address %x, got %x", account1, test3.Accounts[1])
	}

	// single topic
	var test4 FilterCriteria
	vector = fmt.Sprintf(`{"topics": ["%s"]}`, topic0.Hex())
	if err := json.Unmarshal([]byte(vector), &test4); err != nil {
		t.Fatal(err)
	}
	if len(test4.Topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(test4.Topics))
	}
	if len(test4.Topics[0]) != 1 {
		t.Fatalf("expected len(topics[0]) to be 1, got %d", len(test4.Topics[0]))
	}
	if test4.Topics[0][0] != topic0 {
		t.Fatalf("got %x, expected %x", test4.Topics[0][0], topic0)
	}

	// test multiple "AND" topics
	var test5 FilterCriteria
	vector = fmt.Sprintf(`{"topics": ["%s", "%s"]}`, topic0.Hex(), topic1.Hex())
	if err := json.Unmarshal([]byte(vector), &test5); err != nil {
		t.Fatal(err)
	}
	if len(test5.Topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(test5.Topics))
	}
	if len(test5.Topics[0]) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(test5.Topics[0]))
	}
	if test5.Topics[0][0] != topic0 {
		t.Fatalf("got %x, expected %x", test5.Topics[0][0], topic0)
	}
	if len(test5.Topics[1]) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(test5.Topics[1]))
	}
	if test5.Topics[1][0] != topic1 {
		t.Fatalf("got %x, expected %x", test5.Topics[1][0], topic1)
	}

	// test optional topic
	var test6 FilterCriteria
	vector = fmt.Sprintf(`{"topics": ["%s", null, "%s"]}`, topic0.Hex(), topic2.Hex())
	if err := json.Unmarshal([]byte(vector), &test6); err != nil {
		t.Fatal(err)
	}
	if len(test6.Topics) != 3 {
		t.Fatalf("expected 3 topics, got %d", len(test6.Topics))
	}
	if len(test6.Topics[0]) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(test6.Topics[0]))
	}
	if test6.Topics[0][0] != topic0 {
		t.Fatalf("got %x, expected %x", test6.Topics[0][0], topic0)
	}
	if len(test6.Topics[1]) != 0 {
		t.Fatalf("expected 0 topic, got %d", len(test6.Topics[1]))
	}
	if len(test6.Topics[2]) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(test6.Topics[2]))
	}
	if test6.Topics[2][0] != topic2 {
		t.Fatalf("got %x, expected %x", test6.Topics[2][0], topic2)
	}

	// test OR topics
	var test7 FilterCriteria
	vector = fmt.Sprintf(`{"topics": [["%s", "%s"], null, ["%s", null]]}`, topic0.Hex(), topic1.Hex(), topic2.Hex())
	if err := json.Unmarshal([]byte(vector), &test7); err != nil {
		t.Fatal(err)
	}
	if len(test7.Topics) != 3 {
		t.Fatalf("expected 3 topics, got %d topics", len(test7.Topics))
	}
	if len(test7.Topics[0]) != 2 {
		t.Fatalf("expected 2 topics, got %d topics", len(test7.Topics[0]))
	}
	if test7.Topics[0][0] != topic0 || test7.Topics[0][1] != topic1 {
		t.Fatalf("invalid topics expected [%x,%x], got [%x,%x]",
			topic0, topic1, test7.Topics[0][0], test7.Topics[0][1],
		)
	}
	if len(test7.Topics[1]) != 0 {
		t.Fatalf("expected 0 topic, got %d topics", len(test7.Topics[1]))
	}
	if len(test7.Topics[2]) != 0 {
		t.Fatalf("expected 0 topics, got %d topics", len(test7.Topics[2]))
	}
}
