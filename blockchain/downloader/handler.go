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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	router "github.com/fractalplatform/fractal/event"
	adaptor "github.com/fractalplatform/fractal/p2p/protoadaptor"
	"github.com/fractalplatform/fractal/types"
)

type station struct {
	peerCh     chan *router.Event
	blockchain *BlockChain
	networkID  uint64
	quit       chan struct{}
	loopWG     sync.WaitGroup
	downloader *Downloader
	subs       []router.Subscription
}

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func newStation(bc *BlockChain, networkID uint64) *station {
	bs := &station{
		peerCh:     make(chan *router.Event),
		blockchain: bc,
		networkID:  networkID,
		quit:       make(chan struct{}),
		downloader: NewDownloader(bc),
		subs:       make([]router.Subscription, 6),
	}
	bs.subs[0] = router.Subscribe(nil, bs.peerCh, router.NewPeerNotify, nil)
	bs.subs[1] = router.Subscribe(nil, bs.peerCh, router.DelPeerNotify, nil)
	bs.subs[2] = router.Subscribe(nil, bs.peerCh, router.P2PGetStatus, "")
	bs.subs[3] = router.Subscribe(nil, bs.peerCh, router.P2PGetBlockHashMsg, &getBlockHashByNumber{})
	bs.subs[4] = router.Subscribe(nil, bs.peerCh, router.P2PGetBlockHeadersMsg, &getBlockHeadersData{})
	bs.subs[5] = router.Subscribe(nil, bs.peerCh, router.P2PGetBlockBodiesMsg, []common.Hash{})

	bs.loopWG.Add(1)
	go func() {
		bs.loop()
		bs.loopWG.Done()
	}()

	return bs
}

func (bs *station) chainStatus() *statusData {
	genesis := bs.blockchain.Genesis()
	head := bs.blockchain.CurrentHeader()
	hash := head.Hash()
	number := head.Number.Uint64()
	td := bs.blockchain.GetTd(hash, number)
	return &statusData{
		ProtocolVersion: uint32(1),
		NetworkID:       0,
		TD:              td,
		CurrentBlock:    hash,
		CurrentNumber:   number,
		GenesisBlock:    genesis.Hash(),
	}
}

func checkChainStatus(local *statusData, remote *statusData) error {
	if local.GenesisBlock != remote.GenesisBlock {
		return errResp(ErrGenesisBlockMismatch, "remote:%x (!= self:%x)", remote.GenesisBlock[:8], local.GenesisBlock[:8])
	}
	if local.NetworkID != remote.NetworkID {
		return errResp(ErrNetworkIDMismatch, "remote:%d (!= self:%d)", remote.NetworkID, local.NetworkID)
	}
	if local.ProtocolVersion != remote.ProtocolVersion {
		return errResp(ErrProtocolVersionMismatch, "remote:%d (!= self:%d)", remote.ProtocolVersion, local.ProtocolVersion)
	}
	return nil
}

func (bs *station) handshake(e *router.Event) {
	station := router.NewLocalStation("shake"+e.From.Name(), nil)
	ch := make(chan *router.Event)
	sub := router.Subscribe(station, ch, router.P2PStatusMsg, &statusData{})
	defer sub.Unsubscribe()
	router.StationRegister(station)
	defer router.StationUnregister(station)

	router.SendTo(station, e.From, router.P2PGetStatus, "")

	timer := time.After(5 * time.Second)
	select {
	case <-bs.quit:
	case e := <-ch:
		remote := e.Data.(*statusData)
		if err := checkChainStatus(bs.chainStatus(), remote); err != nil {
			router.SendTo(nil, nil, router.OneMinuteLimited, e.From) // disconnect and put into blacklist
			log.Warn("Handshake failure", "error", err, "station", fmt.Sprintf("%x", e.From.Name()))
			return
		}
		log.Info("Handshake complete", "station", fmt.Sprintf("%x", e.From.Name()))
		bs.downloader.AddStation(e.From, remote.TD, remote.CurrentNumber, remote.CurrentBlock)
		router.SendTo(e.From, nil, router.NewPeerPassedNotify, e.Data)
	case <-timer:
		log.Warn("Handshake timeout", "station", fmt.Sprintf("%x", e.From.Name()))
		router.SendTo(nil, nil, router.DisconectCtrl, e.From)
	}
}

func (bs *station) loop() {
	for {
		select {
		case <-bs.quit:
			return
		case e := <-bs.peerCh:
			switch e.Typecode {
			case router.NewPeerNotify:
				bs.loopWG.Add(1)
				go func() {
					bs.handshake(e)
					bs.loopWG.Done()
				}()
			case router.DelPeerNotify:
				bs.loopWG.Add(1)
				go func() {
					bs.downloader.DelStation(e.From)
					bs.loopWG.Done()
				}()
			default:
				if router.Thread(e.From) > 3 {
					log.Warn("Disconnect because request too frequently:", "node:", adaptor.GetFnode(e.From), "thread", router.Thread(e.From))
					router.SendTo(nil, nil, router.OneMinuteLimited, e.From)
					continue
				}
				router.AddThread(e.From, 1)
				bs.loopWG.Add(1)
				go func() {
					bs.handleMsg(e)
					bs.loopWG.Done()
				}()
			}
		}
	}
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (bs *station) handleMsg(e *router.Event) error {
	start := time.Now()
	defer func() {
		router.AddCPU(e.From, time.Since(start))
		router.AddThread(e.From, -1)
	}()
	switch e.Typecode {
	case router.P2PGetStatus:
		status := bs.chainStatus()
		router.ReplyEvent(e, router.P2PStatusMsg, status)

	case router.P2PGetBlockHashMsg:
		query := e.Data.(*getBlockHashByNumber)
		hashes := make([]common.Hash, 0, query.Amount)
		for len(hashes) < int(query.Amount) {
			header := bs.blockchain.GetHeaderByNumber(query.Number)
			if header == nil {
				break
			}
			hashes = append(hashes, header.Hash())
			if query.Reverse {
				if query.Number < query.Skip+1 {
					break
				}
				query.Number -= query.Skip + 1
			} else {
				query.Number += query.Skip + 1
			}
		}
		router.ReplyEvent(e, router.P2PBlockHashMsg, hashes)
	// Block header query, collect the requested headers and reply
	case router.P2PGetBlockHeadersMsg:
		// Decode the complex header query
		query := e.Data.(*getBlockHeadersData)
		if query.Origin.Hash != (common.Hash{}) {
			header := bs.blockchain.GetHeaderByHash(query.Origin.Hash)
			if header == nil {
				router.ReplyEvent(e, router.P2PBlockHeadersMsg, []*types.Header{})
				return nil
			}
			query.Origin.Number = header.Number.Uint64()
		}

		// Gather headers until the fetch or network limits is reached
		var (
			headers []*types.Header
		)
		for len(headers) < int(query.Amount) {
			// Retrieve the next header satisfying the query
			origin := bs.blockchain.GetHeaderByNumber(query.Origin.Number)
			if origin == nil {
				break
			}
			headers = append(headers, origin)

			// Advance to the next header of the query
			if query.Reverse {
				// Number based traversal towards the genesis block
				if query.Origin.Number < query.Skip+1 {
					break
				}
				query.Origin.Number -= query.Skip + 1
			} else {
				// Number based traversal towards the leaf block
				query.Origin.Number += query.Skip + 1
			}
		}

		router.ReplyEvent(e, router.P2PBlockHeadersMsg, headers)
		return nil
	case router.P2PGetBlockBodiesMsg:
		// Decode the retrieval message
		hashes := e.Data.([]common.Hash)
		// Gather blocks until the fetch or network limits is reached
		var (
			bodies []*types.Body
		)
		for _, hash := range hashes {
			// Retrieve the requested block body, stopping if enough was found
			body := bs.blockchain.GetBody(hash)
			if body == nil {
				break
			}
			bodies = append(bodies, body)
		}
		router.ReplyEvent(e, router.P2PBlockBodiesMsg, bodies)
		return nil
	}
	return nil
}

func (bs *station) Stop() {
	log.Info("BlockchainHandler stopping...")
	close(bs.quit)
	for _, sub := range bs.subs {
		sub.Unsubscribe()
	}
	bs.loopWG.Wait()
	bs.downloader.Stop()
	log.Info("BlockchainHandler stopped.")
}
