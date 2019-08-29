package blockchain

import (
	"testing"

	"github.com/ethereum/go-ethereum/log"
	router "github.com/fractalplatform/fractal/event"
)

type simuAdaptor struct{}

func (simuAdaptor) SendOut(e *router.Event) error {
	e.To = nil
	//e.From = router.NewLocalStation(e.From.Name(), nil)
	router.SendEvent(e)
	return nil
}

func TestDownloadTask(t *testing.T) {
	printLog(log.LvlDebug)

	router.AdaptorRegister(simuAdaptor{})
	genesis := DefaultGenesis()
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	allCandidates, allHeaderTimes := genCanonicalCandidatesAndTimes(genesis)
	chain, _ = makeNewChain(t, genesis, chain, allCandidates, allHeaderTimes)
	dl := chain.station.downloader
	status := &stationStatus{
		station: router.NewRemoteStation("teststatus", nil),
		errCh:   make(chan struct{}),
	}
	head := dl.blockchain.CurrentBlock()
	if head.NumberU64() > 0 {
		status.updateStatus(&NewBlockHashesData{
			Hash:      head.Hash(),
			TD:        head.Difficulty(),
			Number:    head.NumberU64(),
			Completed: true,
		})
		n, err := dl.shortcutDownload(status, 0, chain.GetHeaderByNumber(0).Hash(), head.NumberU64(), head.Hash())
		if err != nil || n != head.NumberU64() {
			t.Error("err", err, "get", n, "want", head.NumberU64())
		}
	}
}
