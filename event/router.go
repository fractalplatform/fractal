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

package event

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// ProtoAdaptor used to send out event
type ProtoAdaptor interface {
	SendOut(*Event) error
}

// Router Router all events
type Router struct {
	namedFeeds   map[string]map[int]*Feed
	namedMutex   sync.RWMutex
	unnamedFeeds map[int]*Feed
	adaptor      ProtoAdaptor
	unnamedMutex sync.RWMutex
	stations     map[string]Station
	stationMutex sync.RWMutex
	eval         *stationEval
}

var router *Router
var routerMutex sync.RWMutex

// InitRouter init router.
func init() {
	routerMutex.Lock()
	router = New()
	routerMutex.Unlock()
}

// New returns an initialized Router instance.
func New() *Router {
	return &Router{
		unnamedFeeds: make(map[int]*Feed),
		namedFeeds:   make(map[string]map[int]*Feed),
		stations:     make(map[string]Station),
		eval:         newStationEval(),
	}
}

// Reset ntended for testing。
func Reset() {
	routerMutex.Lock()
	router = New()
	routerMutex.Unlock()
}

// Event is including normal event and p2p event
type Event struct {
	From     Station
	To       Station
	Typecode int
	Data     interface{}
}

// Type enumerator
const (
	P2PRouterTestInt      int = iota // 0
	P2PRouterTestInt64               // 1
	P2PRouterTestString              // 2
	P2PGetStatus                     // 3 Status request
	P2PStatusMsg                     // 4 Status response
	P2PGetBlockHashMsg               // 5 BlockHash request
	P2PGetBlockHeadersMsg            // 6 BlockHeader request
	P2PGetBlockBodiesMsg             // 7 BlockBodies request
	P2PBlockHeadersMsg               // 8 BlockHeader response
	P2PBlockBodiesMsg                // 9 BlockBodies response
	P2PBlockHashMsg                  // 10 BlockHash response
	P2PNewBlockHashesMsg             // 11 NewBlockHash notify
	P2PTxMsg                         // 12 TxMsg notify
	P2PEndSize
	ChainHeadEv         = 1023 + iota - P2PEndSize // 1024
	NewPeerNotify                                  // 1025 emit when remote peer incoming but needed to check chainID and genesis block
	DelPeerNotify                                  // 1026 emit when remote peer disconnected
	DisconectCtrl                                  // 1027 emit if needed to let remote peer disconnect
	NewPeerPassedNotify                            // 1028 emit when remote peer had same chain ID and genesis block
	OneMinuteLimited                               // 1029 add peer to blacklist
	NewMinedEv                                     // 1030 emit when new block was mined
	NewTxs                                         // 1031 emit when new transactions needed to broadcast
	EndSize
)

var typeListMutex sync.RWMutex
var typeList = [EndSize]reflect.Type{}

var typeLimit = [P2PEndSize]int{
	P2PGetStatus:          1,
	P2PGetBlockHashMsg:    128,
	P2PGetBlockHeadersMsg: 64,
	P2PGetBlockBodiesMsg:  64,
	P2PNewBlockHashesMsg:  3,
}

// ReplyEvent is equivalent to `SendTo(e.To, e.From, typecode, data)`
func ReplyEvent(e *Event, typecode int, data interface{}) {
	SendEvent(&Event{
		From:     e.To,
		To:       e.From,
		Typecode: typecode,
		Data:     data,
	})
}

// GetTypeByCode return Type by typecode
func GetTypeByCode(typecode int) reflect.Type {
	if typecode < P2PEndSize {
		typeListMutex.RLock()
		defer typeListMutex.RUnlock()
		return typeList[typecode]
	}
	return nil
}

func bindTypeToCode(typecode int, data interface{}) {
	if typecode >= EndSize {
		panic("dataType greater than EndSize!")
	}
	if data == nil {
		return
	}
	typ := reflect.TypeOf(data)
	typeListMutex.RLock()
	etyp := typeList[typecode]
	typeListMutex.RUnlock()
	if etyp == nil {
		typeListMutex.Lock()
		etyp = typeList[typecode]
		if etyp == nil {
			typeList[typecode] = typ
		}
		typeListMutex.Unlock()
		return
	}
	if etyp != typ {
		panic(fmt.Sprintf("%s mismatch %s!", typ.String(), etyp.String()))
	}
}

// GetStationByName retrun Station by Station's name
func GetStationByName(name string) Station {
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.GetStationByName(name)
}
func (router *Router) GetStationByName(name string) Station {
	router.stationMutex.RLock()
	defer router.stationMutex.RUnlock()
	return router.stations[name]
}

// StationRegister register 'Station' to Router
func StationRegister(station Station) {
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	router.StationRegister(station)
}
func (router *Router) StationRegister(station Station) {
	router.stationMutex.Lock()
	router.stations[station.Name()] = station
	router.stationMutex.Unlock()
	if station.IsRemote() && !station.IsBroadcast() {
		router.eval.register(station)
	}
}

// StationUnregister unregister 'Station'
func StationUnregister(station Station) {
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	router.StationUnregister(station)
}
func (router *Router) StationUnregister(station Station) {
	router.stationMutex.Lock()
	delete(router.stations, station.Name())
	router.stationMutex.Unlock()
	if station.IsRemote() {
		router.eval.unregister(station)
	}
}

func (router *Router) bindChannelToStation(station Station, typecode int, channel chan *Event) Subscription {
	name := station.Name()
	router.namedMutex.Lock()
	_, ok := router.namedFeeds[name]
	if !ok {
		router.namedFeeds[name] = make(map[int]*Feed)
	}
	feed, ok := router.namedFeeds[name][typecode]
	if !ok {
		feed = &Feed{}
		router.namedFeeds[name][typecode] = feed
	}
	router.namedMutex.Unlock()
	return feed.Subscribe(channel)
}

func (router *Router) bindChannelToTypecode(typecode int, channel chan *Event) Subscription {
	router.unnamedMutex.Lock()
	feed, ok := router.unnamedFeeds[typecode]
	if !ok {
		feed = &Feed{}
		router.unnamedFeeds[typecode] = feed
	}
	router.unnamedMutex.Unlock()
	return feed.Subscribe(channel)
}

// Subscribe .
func Subscribe(station Station, channel chan *Event, typecode int, data interface{}) Subscription {
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.Subscribe(station, channel, typecode, data)
}
func (router *Router) Subscribe(station Station, channel chan *Event, typecode int, data interface{}) Subscription {

	bindTypeToCode(typecode, data)

	var sub Subscription

	if station != nil {
		sub = router.bindChannelToStation(station, typecode, channel)
	} else {
		sub = router.bindChannelToTypecode(typecode, channel)
	}
	return sub
}

// AdaptorRegister register P2P interface to Router
func AdaptorRegister(adaptor ProtoAdaptor) {
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	router.AdaptorRegister(adaptor)
}
func (router *Router) AdaptorRegister(adaptor ProtoAdaptor) {
	router.unnamedMutex.Lock()
	defer router.unnamedMutex.Unlock()
	if router.adaptor == nil {
		router.adaptor = adaptor
	}
}

// SendTo  is equivalent to SendEvent(&Event{From: from, To: to, Type: typecode, Data: data})
func SendTo(from, to Station, typecode int, data interface{}) int {
	return SendEvent(&Event{From: from, To: to, Typecode: typecode, Data: data})
}

// SendEvent send event
func SendEvent(e *Event) (nsent int) {
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.SendEvent(e)
}
func (router *Router) SendEvent(e *Event) (nsent int) {

	//if e.Typecode >= EndSize || (typeList[e.Typecode] != nil && reflect.TypeOf(e.Data) != typeList[e.Typecode]) {
	//	fmt.Println("SendEvent Err:", e.Typecode, EndSize, reflect.TypeOf(e.Data), typeList[e.Typecode])
	//	panic("-")
	//return
	//}

	if e.To != nil {
		if e.To.IsRemote() {
			router.sendToAdaptor(e)
			return 1
		}
		//if len(e.To.Name()) != 0 {
		router.namedMutex.RLock()
		feeds, ok := router.namedFeeds[e.To.Name()]
		if ok {
			feed, ok := feeds[e.Typecode]
			if ok {
				nsent = feed.Send(e)
			}
		}
		router.namedMutex.RUnlock()
		return
		//}
	}

	router.unnamedMutex.RLock()
	if feed, ok := router.unnamedFeeds[e.Typecode]; ok {
		nsent = feed.Send(e)
	}
	router.unnamedMutex.RUnlock()
	return
}

func (router *Router) sendToAdaptor(e *Event) {
	router.unnamedMutex.RLock()
	if router.adaptor != nil {
		router.adaptor.SendOut(e)
	}
	router.unnamedMutex.RUnlock()
}

// SendEvents .
func SendEvents(es []*Event) (nsent int) {
	for _, e := range es {
		nsent += SendEvent(e)
	}
	return
}

//GetDDosLimit get messagetype req limit per second
func GetDDosLimit(t int) int {
	return typeLimit[t]
}

func AddNetIn(s Station, pkg uint64) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.addNetIn(s, pkg)
}
func AddNetOut(s Station, pkg uint64) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.addNetOut(s, pkg)
}

func AddCPU(s Station, dur time.Duration) time.Duration {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.addCPU(s, dur)
}

func AddAck(s Station, dur time.Duration) (uint64, time.Duration) {
	if s == nil {
		return 0, 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.addAck(s, dur)
}
func AddErr(s Station, n uint64) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.addErr(s, n)
}
func AddThread(s Station, c int64) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.addThread(s, c)
}

func CPU(s Station) time.Duration {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.cpu(s)
}
func NetIn(s Station) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.netin(s)
}
func NetOut(s Station) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.netout(s)
}
func Ack(s Station) (uint64, time.Duration) {
	if s == nil {
		return 0, 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.ack(s)
}

func Err(s Station) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.err(s)
}

func Thread(s Station) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.thread(s)
}

func Score(s Station) uint64 {
	if s == nil {
		return 0
	}
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.score(s)
}

func WorstStation() Station {
	routerMutex.RLock()
	defer routerMutex.RUnlock()
	return router.eval.getWorst()
}
