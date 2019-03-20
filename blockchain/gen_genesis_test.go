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
	"encoding/json"
	"fmt"
	"testing"
)

func TestGenGenesis(t *testing.T) {
	genesis := DefaultGenesis()

	j, err := genesis.MarshalJSON()
	if err != nil {
		t.Fatal(fmt.Sprintf("genesis marshal --- %v", err))
	}

	if err := genesis.UnmarshalJSON(j); err != nil {
		t.Fatal(fmt.Sprintf("genesis Unmarshal --- %v", err))
	}
	_, err = genesis.MarshalJSON()
	if err != nil {
		t.Fatal(fmt.Sprintf("genesis marshal --- %v", err))
	}

	s := `{
		"config": {
			"chainId": 1,
			"sysName": "ftsystemio",
			"sysToken": "fractalfoundation"
		},
		"dpos": {
			"MaxURLLen": 512,
			"UnitStake": 1000,
			"CadidateMinQuantity": 10000,
			"VoterMinQuantity": 1,
			"ActivatedMinQuantity": 1000000,
			"BlockInterval": 3000000000,
			"BlockFrequency": 6,
			"CadidateScheduleSize": 21,
			"DelayEcho": 0,
			"AccountName": "fdpos",
			"SystemName": "ft",
			"ExtraBlockRewardUnit": 1,
			"BlockReward": 20
		},
		"timestamp": "0x0",
		"extraData": "0x5a302047656e6573697320426c6f636b",
		"gasLimit": "0x5f5e100",
		"difficulty": "0x20000",
		"coinbase": "systemio",
		"allocAccounts": [{
			"name": "ftsystemio",
			"pubKey": "0x047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"
		}],
		"allocAssets": [{
			"assetname": "ftperfoundation",
			"symbol": "ft",
			"amount": 100000000,
			"decimals": 18,
			"owner": "ftsystemio"
		}]
	}`
	if err := json.Unmarshal([]byte(s), genesis); err != nil {
		t.Fatal(fmt.Sprintf("genesis unmarshal --- %v", err))
	}

	j, err = json.Marshal(genesis)
	if err != nil {
		t.Fatal(fmt.Sprintf("genesis marshal --- %v", err))
	}

}
