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
