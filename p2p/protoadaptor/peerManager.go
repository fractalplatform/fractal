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

package protoadaptor

import (
	"sync"

	router "github.com/fractalplatform/fractal/event"
)

type peerMangaer struct {
	activePeers map[[8]byte]*remotePeer
	station     router.Station
	mutex       sync.RWMutex
}

func (pm *peerMangaer) addActivePeer(peer *remotePeer) {
	var key [8]byte
	copy(key[:], peer.peer.ID().Bytes()[:8])
	pm.mutex.Lock()
	pm.activePeers[key] = peer
	pm.mutex.Unlock()
}

func (pm *peerMangaer) delActivePeer(peer *remotePeer) {
	var key [8]byte
	copy(key[:], peer.peer.ID().Bytes()[:8])
	pm.mutex.Lock()
	delete(pm.activePeers, key)
	pm.mutex.Unlock()
}

func (pm *peerMangaer) mapActivePeer(handler func(*remotePeer)) {
	pm.mutex.RLock()
	for _, peer := range pm.activePeers {
		handler(peer)
	}
	pm.mutex.RUnlock()
}
