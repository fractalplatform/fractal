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
	"context"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
)

type Backend interface {
	ChainDb() fdb.Database
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Header
	HeaderByHash(ctx context.Context, blockHash common.Hash) *types.Header
	GetReceipts(ctx context.Context, blockHash common.Hash) ([]*types.Receipt, error)
	GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error)
}

// Filter can be used to retrieve and filter logs.
type Filter struct {
	backend Backend

	db       fdb.Database
	accounts []common.Name
	topics   [][]common.Hash

	block      common.Hash // Block hash if filtering a single block
	begin, end int64       // Range interval if filtering multiple blocks

}

func includes(accounts []common.Name, a common.Name) bool {
	for _, acct := range accounts {
		if acct == a {
			return true
		}
	}

	return false
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*types.Log, accounts []common.Name, topics [][]common.Hash) []*types.Log {
	var ret []*types.Log
Logs:
	for _, log := range logs {
		if len(accounts) > 0 && !includes(accounts, log.Name) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue Logs
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}

func bloomFilter(bloom types.Bloom, accounts []common.Name, topics [][]common.Hash) bool {
	if len(accounts) > 0 {
		var included bool
		for _, addr := range accounts {
			if types.BloomLookup(bloom, addr) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, sub := range topics {
		included := len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if types.BloomLookup(bloom, topic) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}
