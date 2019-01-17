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
)

// ProtoAdaptor used to send out event
type ProtoAdaptor interface {
	SendOut(*Event) error
}

// Router Router all events
type Router struct {
	unnamedFeeds map[int]*Feed
	namedFeeds   map[string]map[int]*Feed
	adaptor      ProtoAdaptor
	mutex        sync.RWMutex
	stations     map[string]Station
	stationMutex sync.RWMutex
}

var router *Router

// Event is including normal event and p2p event
type Event struct {
	From     Station
	To       Station
	Typecode int
	Data     interface{}
}

// Type enumerator
const (
	RouterTestInt                int = iota // 0
	RouterTestInt64                         // 1
	RouterTestString                        // 2
	P2pNewPeer                              // 3
	P2pDelPeer                              // 4
	P2pDisconectPeer                        // 5
	DownloaderGetStatus                     // 6
	DownloaderStatusMsg                     // 7
	DownloaderGetBlockHashMsg               // 8
	DownloaderGetBlockHeadersMsg            // 9
	DownloaderGetBlockBodiesMsg             // 10
	BlockHeadersMsg                         // 11
	BlockBodiesMsg                          // 12
	BlockHashMsg                            // 13
	NewBlockHashesMsg                       // 14
	TxMsg                                   // 15

	ChainEv     // 16
	ChainSideEv // 17
	ChainHeadEv // 18
	LogsEv      // 19
	TxEv        // 20

	NewMinedEv

	EndSize
)

var clear []Subscription

var typeList = [EndSize]reflect.Type{
	RouterTestInt:    nil,
	RouterTestInt64:  nil,
	RouterTestString: nil,
	P2pNewPeer:       nil,
	P2pDelPeer:       nil,
	P2pDisconectPeer: nil,
	ChainEv:          nil,
	ChainSideEv:      nil,
	ChainHeadEv:      nil,
	LogsEv:           nil,
	TxEv:             nil,
}

func InitRounter() {
	router = &Router{
		unnamedFeeds: make(map[int]*Feed),
		namedFeeds:   make(map[string]map[int]*Feed),
		stations:     make(map[string]Station),
	}
	clear = make([]Subscription, 0)
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
	if typecode < EndSize {
		return typeList[typecode]
	}
	return nil
}

func bindTypeToCode(typecode int, data interface{}) {
	if typecode >= EndSize {
		panic("dataType greater than EndSize!")
	}
	typ := reflect.TypeOf(data)
	if typeList[typecode] == nil {
		typeList[typecode] = typ
		return
	}
	if typeList[typecode] != typ {
		panic(fmt.Sprintf("%s mismatch %s!", typ.String(), typeList[typecode].String()))
	}
}

// GetStationByName retrun Station by Station's name
func GetStationByName(name string) Station {
	router.stationMutex.RLock()
	defer router.stationMutex.RUnlock()
	return router.stations[name]
}

// StationRegister register 'Station' to Router
func StationRegister(station Station) {
	router.stationMutex.Lock()
	router.stations[station.Name()] = station
	router.stationMutex.Unlock()
}

// StationUnregister unregister 'Station'
func StationUnregister(station Station) {
	router.stationMutex.Lock()
	delete(router.stations, station.Name())
	router.stationMutex.Unlock()
}

func bindChannelToStation(station Station, typecode int, channel chan *Event) Subscription {
	name := station.Name()
	_, ok := router.namedFeeds[name]
	if !ok {
		router.namedFeeds[name] = make(map[int]*Feed)
	}
	feed, ok := router.namedFeeds[name][typecode]
	if !ok {
		feed = &Feed{}
		router.namedFeeds[name][typecode] = feed
	}
	return feed.Subscribe(channel)
}

func bindChannelToTypecode(typecode int, channel chan *Event) Subscription {
	feed, ok := router.unnamedFeeds[typecode]
	if !ok {
		feed = &Feed{}
		router.unnamedFeeds[typecode] = feed
	}
	return feed.Subscribe(channel)
}

// Subscribe .
func Subscribe(station Station, channel chan *Event, typecode int, data interface{}) Subscription {
	router.mutex.Lock()
	defer router.mutex.Unlock()

	bindTypeToCode(typecode, data)

	var sub Subscription

	if station != nil {
		StationRegister(station)
		sub = bindChannelToStation(station, typecode, channel)
	} else {
		sub = bindChannelToTypecode(typecode, channel)
	}
	clear = append(clear, sub)
	return sub
}

// AdaptorRegister register P2P interface to Router
func AdaptorRegister(adaptor ProtoAdaptor) {
	router.mutex.Lock()
	defer router.mutex.Unlock()
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

	//if e.Typecode >= EndSize || (typeList[e.Typecode] != nil && reflect.TypeOf(e.Data) != typeList[e.Typecode]) {
	//	fmt.Println("SendEvent Err:", e.Typecode, EndSize, reflect.TypeOf(e.Data), typeList[e.Typecode])
	//	panic("-")
	//return
	//}

	router.mutex.RLock()
	defer router.mutex.RUnlock()
	if e.To != nil {
		if e.To.IsRemote() {
			sendToAdaptor(e)
			return 1
		}
		//if len(e.To.Name()) != 0 {
		feeds, ok := router.namedFeeds[e.To.Name()]
		if ok {
			feed, ok := feeds[e.Typecode]
			if ok {
				nsent = feed.Send(e)
			}
		}
		return
		//}
	}

	if feed, ok := router.unnamedFeeds[e.Typecode]; ok {
		nsent = feed.Send(e)
		return
	}
	return
}

func sendToAdaptor(e *Event) {
	if router.adaptor != nil {
		router.adaptor.SendOut(e)
	}
}

// SendEvents .
func SendEvents(es []*Event) (nsent int) {
	for _, e := range es {
		nsent += SendEvent(e)
	}
	return
}

// Clear .
func Clear() {
	for _, sub := range clear {
		sub.Unsubscribe()
	}
}
