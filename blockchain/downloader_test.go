package blockchain

import (
	"testing"

	"github.com/ethereum/go-ethereum/log"
)

func TestDownloadTask(t *testing.T) {
	printLog(log.LvlDebug)
	genesis := DefaultGenesis()
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	allCandidates, allHeaderTimes := genCanonicalCandidatesAndTimes(genesis)
	chain, blocks := makeNewChain(t, genesis, chain, allCandidates, allHeaderTimes)
	dl := NewDownloader(chain)
	status := &stationStatus{
		station: nil,
		errCh:   make(chan struct{}),
	}
	if len(blocks) > 0 {
		curblock := blocks[len(blocks)-1]
		n, err := dl.shortcutDownload(status, 0, chain.GetHeaderByNumber(0).Hash(), curblock.NumberU64(), curblock.Hash())
		if err != nil || n != curblock.NumberU64() {
			t.Error("err", err, "get", n, "want", curblock.NumberU64())
		}
	}
}
