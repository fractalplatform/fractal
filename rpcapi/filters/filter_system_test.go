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
	"math/big"
	"testing"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
)

type testBackend struct {
	db         fdb.Database
	sections   uint64
	txFeed     *event.Feed
	rmLogsFeed *event.Feed
	logsFeed   *event.Feed
	chainFeed  *event.Feed
}

func (b *testBackend) ChainDb() fdb.Database {
	return b.db
}

func (b *testBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) *types.Header {
	var (
		hash common.Hash
		num  uint64
	)
	if blockNr == rpc.LatestBlockNumber {
		hash = rawdb.ReadHeadBlockHash(b.db)
		number := rawdb.ReadHeaderNumber(b.db, hash)
		if number == nil {
			return nil
		}
		num = *number
	} else {
		num = uint64(blockNr)
		hash = rawdb.ReadCanonicalHash(b.db, num)
	}
	return rawdb.ReadHeader(b.db, hash, num)
}

func (b *testBackend) HeaderByHash(ctx context.Context, hash common.Hash) *types.Header {
	number := rawdb.ReadHeaderNumber(b.db, hash)
	if number == nil {
		return nil
	}
	return rawdb.ReadHeader(b.db, hash, *number)
}

func (b *testBackend) GetReceipts(ctx context.Context, hash common.Hash) ([]*types.Receipt, error) {
	if number := rawdb.ReadHeaderNumber(b.db, hash); number != nil {
		return rawdb.ReadReceipts(b.db, hash, *number), nil
	}
	return nil, nil
}

func (b *testBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.db, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.db, hash, *number)

	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

// TestBlockSubscription tests if a block subscription returns block hashes for posted chain events.
// It creates multiple subscriptions:
// - one at the start and should receive all posted chain events and a second (blockHashes)
// - one that is created after a cutoff moment and uninstalled after a second cutoff moment (blockHashes[cutoff1:cutoff2])
// - one that is created after the second cutoff moment (blockHashes[cutoff2:])
func TestBlockSubscription(t *testing.T) {
	t.Parallel()

	var (
		db         = rawdb.NewMemoryDatabase()
		txFeed     = new(event.Feed)
		rmLogsFeed = new(event.Feed)
		logsFeed   = new(event.Feed)
		chainFeed  = new(event.Feed)
		backend    = &testBackend{db, 0, txFeed, rmLogsFeed, logsFeed, chainFeed}
		api        = NewPublicFilterAPI(backend)

		testHeader = &types.Header{
			ParentHash: common.HexToHash("0a5843ac1cb04865017cb35a57b50b07084e5fcee39b5acadade33149f4fff9e"),
			Coinbase:   common.Name("cpinbase"),
			Root:       common.HexToHash("ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017"),
			Difficulty: big.NewInt(131072),
			Number:     big.NewInt(100),
			GasLimit:   uint64(3141592),
			GasUsed:    uint64(21000),
			Time:       big.NewInt(1426516743),
			Extra:      []byte("test Header"),
		}

		blocks = []*types.Block{
			types.NewBlockWithHeader(testHeader),
		}
	)

	chan0 := make(chan *types.Header)
	sub0 := api.events.SubscribeNewHeads(chan0)
	chan1 := make(chan *types.Header)
	sub1 := api.events.SubscribeNewHeads(chan1)

	go func() { // simulate client
		i1, i2 := 0, 0
		for i1 != len(blocks) || i2 != len(blocks) {
			select {
			case header := <-chan0:
				if blocks[i1].Hash() != header.Hash() {
					t.Errorf("sub0 received invalid hash on index %d, want %x, got %x", i1, blocks[i1].Hash(), header.Hash())
				}
				i1++
			case header := <-chan1:
				if blocks[i2].Hash() != header.Hash() {
					t.Errorf("sub1 received invalid hash on index %d, want %x, got %x", i2, blocks[i2].Hash(), header.Hash())
				}
				i2++
			}
		}

		sub0.Unsubscribe()
		sub1.Unsubscribe()
	}()

	time.Sleep(1 * time.Second)
	for _, block := range blocks {
		event.SendEvent(&event.Event{Typecode: event.ChainHeadEv, Data: block})
	}

	<-sub0.Err()
	<-sub1.Err()
}

// TestPendingTxFilter tests whether pending tx filters retrieve all pending transactions that are posted to the event mux.
func TestPendingTxFilter(t *testing.T) {
	t.Parallel()

	var (
		db         = rawdb.NewMemoryDatabase()
		txFeed     = new(event.Feed)
		rmLogsFeed = new(event.Feed)
		logsFeed   = new(event.Feed)
		chainFeed  = new(event.Feed)
		backend    = &testBackend{db, 0, txFeed, rmLogsFeed, logsFeed, chainFeed}
		api        = NewPublicFilterAPI(backend)

		transactions = []*types.Transaction{
			types.NewTransaction(0, big.NewInt(1), &types.Action{}),
			types.NewTransaction(0, big.NewInt(1), &types.Action{}),
			types.NewTransaction(0, big.NewInt(1), &types.Action{}),
			types.NewTransaction(0, big.NewInt(1), &types.Action{}),
			types.NewTransaction(0, big.NewInt(1), &types.Action{}),
		}
		hashes []common.Hash
	)

	fid0 := api.NewPendingTransactionFilter()

	time.Sleep(1 * time.Second)
	event.SendEvent(&event.Event{Typecode: event.NewTxs, Data: transactions})

	timeout := time.Now().Add(1 * time.Second)
	for {
		results, err := api.GetFilterChanges(fid0)
		if err != nil {
			t.Fatalf("Unable to retrieve logs: %v", err)
		}

		h := results.([]common.Hash)
		hashes = append(hashes, h...)
		if len(hashes) >= len(transactions) {
			break
		}
		// check timeout
		if time.Now().After(timeout) {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	if len(hashes) != len(transactions) {
		t.Errorf("invalid number of transactions, want %d transactions(s), got %d", len(transactions), len(hashes))
		return
	}
	for i := range hashes {
		if hashes[i] != transactions[i].Hash() {
			t.Errorf("hashes[%d] invalid, want %x, got %x", i, transactions[i].Hash(), hashes[i])
		}
	}
}

// TestLogFilter tests whether log filters match the correct logs that are posted to the event feed.
// func TestLogFilter(t *testing.T) {
// 	t.Parallel()

// 	var (
// 		db         = rawdb.NewMemoryDatabase()
// 		txFeed     = new(event.Feed)
// 		rmLogsFeed = new(event.Feed)
// 		logsFeed   = new(event.Feed)
// 		chainFeed  = new(event.Feed)
// 		backend    = &testBackend{db, 0, txFeed, rmLogsFeed, logsFeed, chainFeed}
// 		api        = NewPublicFilterAPI(backend)

// 		firstAccount   = common.Name("firstaccount")
// 		secondAccount  = common.Name("secondaccount")
// 		thirdAccount   = common.Name("thirdaccount")
// 		notUsedAccount = common.Name("notusedaccount")
// 		firstTopic     = common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
// 		secondTopic    = common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")
// 		notUsedTopic   = common.HexToHash("0x9999999999999999999999999999999999999999999999999999999999999999")

// 		// posted twice, once as vm.Logs and once as core.PendingLogsEvent
// 		allLogs = []*types.Log{
// 			{Name: firstAccount},
// 			{Name: firstAccount, Topics: []common.Hash{firstTopic}, BlockNumber: 1},
// 			{Name: secondAccount, Topics: []common.Hash{firstTopic}, BlockNumber: 1},
// 			{Name: thirdAccount, Topics: []common.Hash{secondTopic}, BlockNumber: 2},
// 			{Name: thirdAccount, Topics: []common.Hash{secondTopic}, BlockNumber: 3},
// 		}

// 		testCases = []struct {
// 			crit     FilterCriteria
// 			expected []*types.Log
// 			id       rpc.ID
// 		}{
// 			// match all
// 			0: {FilterCriteria{}, allLogs, ""},
// 			// match none due to no matching addresses
// 			1: {FilterCriteria{Accounts: []common.Name{notUsedAccount}, Topics: [][]common.Hash{nil}}, []*types.Log{}, ""},
// 			// match logs based on addresses, ignore topics
// 			2: {FilterCriteria{Accounts: []common.Name{firstAccount}}, allLogs[:2], ""},
// 			// match none due to no matching topics (match with address)
// 			3: {FilterCriteria{Accounts: []common.Name{secondAccount}, Topics: [][]common.Hash{{notUsedTopic}}}, []*types.Log{}, ""},
// 			// match logs based on addresses and topics
// 			4: {FilterCriteria{Accounts: []common.Name{thirdAccount}, Topics: [][]common.Hash{{firstTopic, secondTopic}}}, allLogs[3:5], ""},
// 			// match logs based on multiple addresses and "or" topics
// 			5: {FilterCriteria{Accounts: []common.Name{secondAccount, thirdAccount}, Topics: [][]common.Hash{{firstTopic, secondTopic}}}, allLogs[2:5], ""},
// 			// match all logs due to wildcard topic
// 			6: {FilterCriteria{Topics: [][]common.Hash{nil}}, allLogs[1:], ""},
// 		}
// 	)

// 	// create all filters
// 	for i := range testCases {
// 		testCases[i].id, _ = api.NewFilter(testCases[i].crit)
// 	}

// 	// raise events
// 	time.Sleep(1 * time.Second)
// 	if nsend := logsFeed.Send(allLogs); nsend == 0 {
// 		t.Fatal("Shoud have at least one subscription")
// 	}

// 	for i, tt := range testCases {
// 		var fetched []*types.Log
// 		timeout := time.Now().Add(1 * time.Second)
// 		for { // fetch all expected logs
// 			results, err := api.GetFilterChanges(tt.id)
// 			if err != nil {
// 				t.Fatalf("Unable to fetch logs: %v", err)
// 			}

// 			fetched = append(fetched, results.([]*types.Log)...)
// 			if len(fetched) >= len(tt.expected) {
// 				break
// 			}
// 			// check timeout
// 			if time.Now().After(timeout) {
// 				break
// 			}

// 			time.Sleep(100 * time.Millisecond)
// 		}

// 		if len(fetched) != len(tt.expected) {
// 			t.Errorf("invalid number of logs for case %d, want %d log(s), got %d", i, len(tt.expected), len(fetched))
// 			return
// 		}
// 	}
// }
