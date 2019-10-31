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

package downloader

import (
	"testing"

	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/log"
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

	chain := newCanonical(t, genesis)
	defer chain.Stop()

	chain, _ = makeNewChain(t, genesis, chain, 10, canonicalSeed)

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
