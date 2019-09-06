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
	"fmt"
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
	genesis := DefaultGenesis()
	blockCount := 10
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	makeNewChain(t, genesis, chain, blockCount, canonicalSeed)

	//router.AdaptorRegister(simuAdaptor{})
	errCh := make(chan struct{})
	hash, err := getBlockHashes(nil, nil, &getBlockHashByNumber{0, 1, 0, true}, errCh)
	if err != nil || len(hash) != 1 || hash[0] != chain.GetHeaderByNumber(0).Hash() {
		t.Fatal("genesis block not match")
	}

	headers, err := getHeaders(nil, nil, &getBlockHeadersData{
		hashOrNumber{
			Number: 0,
		}, 1, 0, false,
	}, errCh)
	if err != nil || len(headers) != 1 || headers[0].Number.Uint64() != 0 || headers[0].Hash() != chain.GetHeaderByNumber(headers[0].Number.Uint64()).Hash() {
		t.Fatal(fmt.Sprint("genesis block header not match", err, len(headers)))
	}
	//t.Fatal("genesis block header not match", err, len(headers))
}
