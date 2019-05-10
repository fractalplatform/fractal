package event

import (
	"sync"
	"sync/atomic"
)

type stationEval struct {
	eval      map[string]*evaluate
	mutex     sync.RWMutex
	highest   Station
	highscore uint64
	highMutex sync.RWMutex
}

type evaluate struct {
	cpu    uint64
	netin  uint64
	netout uint64
	ack    uint64
	err    uint64
}

func (e *evaluate) score() uint64 {
	s := e.err*10e9 + e.cpu*10e6 + e.ack
	if e.netin > e.netout {
		return s * e.netin / (e.netout + 1)
	}
	return s * e.netout / (e.netin + 1)
}

func (e *evaluate) addCPU(us uint64) uint64 {
	return atomic.AddUint64(&e.cpu, us)
}

func (e *evaluate) addNetIn(pkg uint64) uint64 {
	return atomic.AddUint64(&e.netin, pkg)
}
func (e *evaluate) addNetOut(pkg uint64) uint64 {
	return atomic.AddUint64(&e.netout, pkg)
}

func (e *evaluate) addAck(us uint64) uint64 {
	return atomic.AddUint64(&e.ack, us)
}

func (e *evaluate) addErr() uint64 {
	return atomic.AddUint64(&e.ack, 1)
}

func (e *evaluate) resetCPU() {
	atomic.StoreUint64(&e.cpu, 0)
}
func (e *evaluate) resetAck() {
	atomic.StoreUint64(&e.ack, 0)
}
func (e *evaluate) resetNetIn() {
	atomic.StoreUint64(&e.netin, 0)
}

func (e *evaluate) resetNetOut() {
	atomic.StoreUint64(&e.netout, 0)
}

func newStationEval() *stationEval {
	return &stationEval{
		eval: make(map[string]*evaluate),
	}
}

func (se *stationEval) register(s Station) {
	id := s.Name()[:8]
	se.mutex.Lock()
	if se.eval[id] == nil {
		se.eval[id] = &evaluate{}
	}
	se.mutex.Unlock()
}

/*
func (se *stationEval) knockOut(){
	a := sync.Map

}*/

func (se *stationEval) unregister(s Station) {
	id := s.Name()[:8]
	se.mutex.Lock()
	delete(se.eval, id)
	se.mutex.Unlock()
	se.delete(s)
}

func (se *stationEval) delete(s Station) {

	se.highMutex.Lock()
	if se.highest != nil && se.highest.Name()[:8] == s.Name()[:8] {
		se.highest = nil
	}
	se.highMutex.Unlock()
}

func (se *stationEval) updateTop(s Station, score uint64) {
	if score <= se.highscore {
		return
	}
	se.highMutex.Lock()
	if se.highest == nil || s.Name()[:8] != se.highest.Name()[:8] {
		se.highscore = score
		se.highest = s
	}
	se.highMutex.Unlock()
}

func (se *stationEval) getHighest() Station {
	se.highMutex.RLock()
	defer se.highMutex.RUnlock()
	return se.highest
}

func (se *stationEval) addNetIn(s Station, pkg uint64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	defer se.updateTop(s, e.score())
	return e.addNetIn(pkg)
}
func (se *stationEval) addNetOut(s Station, pkg uint64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	defer se.updateTop(s, e.score())
	return e.addNetOut(pkg)
}

func (se *stationEval) addCPU(s Station, us uint64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	defer se.updateTop(s, e.score())
	return e.addCPU(us)
}

func (se *stationEval) addAck(s Station, us uint64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	defer se.updateTop(s, e.score())
	return e.addAck(us)
}
func (se *stationEval) addErr(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	defer se.updateTop(s, e.score())
	return e.addErr()
}

func (se *stationEval) cpu(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	return e.cpu
}
func (se *stationEval) netin(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	return e.netin
}
func (se *stationEval) netout(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	return e.netout
}
func (se *stationEval) ack(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	return e.ack
}

func (se *stationEval) err(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	return e.err
}
