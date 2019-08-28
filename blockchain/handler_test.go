package blockchain

import (
	"testing"
)

/*
type simuAdaptor struct{}

var client = router.NewRemoteStation("testClient", nil)
var server = router.NewRemoteStation("testServer", nil)

func (simuAdaptor) SendOut(e *router.Event) error {
	if e.To == server {
		e.To = nil
	} else {
		e.To = router.NewLocalStation(e.To.Name(), nil)
	}
	//e.From = router.NewLocalStation(e.From.Name(), nil)
	router.SendEvent(e)
	return nil
}
*/
func TestHandler(t *testing.T) {
	// printLog(log.LvlDebug)
	genesis := DefaultGenesis()
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	allCandidates, allHeaderTimes := genCanonicalCandidatesAndTimes(genesis)
	makeNewChain(t, genesis, chain, allCandidates, allHeaderTimes)

	//router.AdaptorRegister(simuAdaptor{})
	errCh := make(chan struct{})
	hash, err := getBlockHashes(nil, nil, &getBlcokHashByNumber{0, 1, 0, true}, errCh)
	if err != nil || len(hash) != 1 || hash[0] != chain.GetHeaderByNumber(0).Hash() {
		t.Fatal("genesis block not match")
	}
	if err != nil || len(hash) != 1 || hash[0] != chain.GetHeaderByNumber(0).Hash() {
		t.Fatal("genesis block hash not match")
	}

	headers, err := getHeaders(nil, nil, &getBlockHeadersData{
		hashOrNumber{
			Number: 0,
		}, 1, 0, false,
	}, errCh)
	if err != nil || len(headers) != 1 || headers[0].Number.Uint64() != 0 || headers[0].Hash() != chain.GetHeaderByNumber(headers[0].Number.Uint64()).Hash() {
		t.Fatal("genesis block header not match")
	}

}
