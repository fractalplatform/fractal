package event

import (
	"sync"
	"sync/atomic"
	"time"
)

type stationEval struct {
	eval  map[string]*evaluate
	mutex sync.RWMutex
}

type evaluate struct {
	cpu    uint64
	netout uint64
	thread int64
	netin  uint64
	ack    uint64
	acknum uint64
	err    uint64
}

// higher score , poorer quality
func (e *evaluate) score() uint64 {
	s := e.err + e.cpu/uint64(time.Millisecond) + e.ack/uint64(time.Second)/(e.acknum+1) + uint64(e.thread) + 1000
	if e.netin > e.netout {
		return s * e.netin / (e.netout + 1)
	}
	return s * e.netout / (e.netin + 1)
}

func (e *evaluate) addThread(c int64) uint64 {
	return uint64(atomic.AddInt64(&e.thread, c))
}

func (e *evaluate) addCPU(dur time.Duration) time.Duration {
	return time.Duration(atomic.AddUint64(&e.cpu, uint64(dur)))
}

func (e *evaluate) addNetIn(pkg uint64) uint64 {
	return atomic.AddUint64(&e.netin, pkg)
}
func (e *evaluate) addNetOut(pkg uint64) uint64 {
	return atomic.AddUint64(&e.netout, pkg)
}

func (e *evaluate) addAck(dur time.Duration) (uint64, time.Duration) {
	return atomic.AddUint64(&e.acknum, 1), time.Duration(atomic.AddUint64(&e.ack, uint64(dur)))
}

func (e *evaluate) addErr(n uint64) uint64 {
	return atomic.AddUint64(&e.err, n)
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

func (se *stationEval) unregister(s Station) {
	id := s.Name()[:8]
	se.mutex.Lock()
	delete(se.eval, id)
	se.mutex.Unlock()
}

func (se *stationEval) getWorst() Station {
	se.mutex.RLock()
	score := uint64(0)
	wid := ""
	for id, e := range se.eval {
		if e.score() >= score {
			wid = id
		}
	}
	se.mutex.RUnlock()
	return GetStationByName(wid)
}

func (se *stationEval) addNetIn(s Station, pkg uint64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.addNetIn(pkg)
}
func (se *stationEval) addNetOut(s Station, pkg uint64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.addNetOut(pkg)
}

func (se *stationEval) addCPU(s Station, dur time.Duration) time.Duration {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.addCPU(dur)
}

func (se *stationEval) addAck(s Station, dur time.Duration) (uint64, time.Duration) {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0, 0
	}
	return e.addAck(dur)
}
func (se *stationEval) addErr(s Station, n uint64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.addErr(n)
}
func (se *stationEval) addThread(s Station, c int64) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.addThread(c)
}

func (se *stationEval) cpu(s Station) time.Duration {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return time.Duration(e.cpu)
}
func (se *stationEval) netin(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.netin
}
func (se *stationEval) netout(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.netout
}
func (se *stationEval) ack(s Station) (uint64, time.Duration) {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0, 0
	}
	return e.acknum, time.Duration(e.ack)
}

func (se *stationEval) err(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.err
}

func (se *stationEval) thread(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return uint64(e.thread)
}

func (se *stationEval) score(s Station) uint64 {
	se.mutex.RLock()
	e := se.eval[s.Name()[:8]]
	se.mutex.RUnlock()
	if e == nil {
		return 0
	}
	return e.score()
}
