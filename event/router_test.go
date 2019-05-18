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
	"testing"
	"time"
)

func BenchmarkSubscribe(t *testing.B) {
	var done sync.WaitGroup
	quit := make(chan struct{})
	rwtest := func(station Station) {
		defer done.Done()
		channel := make(chan *Event)
		if station != nil {
			StationRegister(station)
			defer StationUnregister(station)
		}
		sub := Subscribe(station, channel, P2PRouterTestString, "")
		done.Add(1)
		go func() {
			defer done.Done()
			defer sub.Unsubscribe()
			for {
				select {
				case <-quit:
					return
				case <-channel:
				}
			}
		}()
		for {
			select {
			case <-quit:
				return
			default:
				SendTo(nil, station, P2PRouterTestString, "Cocurrent")
			}
		}
	}
	done.Add(1)
	go rwtest(nil)
	for i := 0; i < 10; i++ {
		done.Add(1)
		go rwtest(NewLocalStation(fmt.Sprint("TestStation", i), nil))
	}
	nTime := int64(0)
	totalTime := int64(0)
	maxTime := time.Nanosecond
	minTime := time.Hour
	var testTime = time.Now()
	for {
		nTime++
		station1 := NewLocalStation("TestStation1", nil)
		var channel1 chan *Event
		var start = time.Now()
		sub := Subscribe(station1, channel1, P2PRouterTestString, "")
		duration := time.Since(start)
		StationUnregister(station1)
		sub.Unsubscribe()
		if duration > maxTime {
			maxTime = duration
		}
		if duration < minTime {
			minTime = duration
		}
		totalTime += duration.Nanoseconds()
		if time.Since(testTime) > 15*time.Second {
			break
		}
	}
	close(quit)
	done.Wait()
	// go test -v
	t.Logf("total %d %d avg %d max %d min %d ns\n", nTime, totalTime, totalTime/nTime, maxTime.Nanoseconds(), minTime.Nanoseconds())
}

func TestSendEventToStation(t *testing.T) {
	type testStation struct {
		station Station
		channel chan *Event
	}
	var (
		done     sync.WaitGroup
		station1 = &testStation{NewLocalStation("TestStation1", nil), make(chan *Event)}
		station2 = &testStation{NewLocalStation("TestStation2", nil), make(chan *Event)}
		station3 = &testStation{NewLocalStation("TestStation3", nil), make(chan *Event)}
		sub1     Subscription
		sub2     Subscription
		sub3     Subscription
	)
	StationRegister(station1.station)
	defer StationUnregister(station1.station)
	StationRegister(station2.station)
	defer StationUnregister(station2.station)
	StationRegister(station3.station)
	defer StationUnregister(station3.station)
	sub1 = Subscribe(station1.station, station1.channel, P2PRouterTestString, "")
	sub2 = Subscribe(station2.station, station2.channel, P2PRouterTestString, "")
	sub3 = Subscribe(station3.station, station3.channel, P2PRouterTestString, "")

	errorList := []string{}
	recvAndCheck := func(station *testStation, expect interface{}) {
		timer := time.After(time.Second)
		select {
		case e := <-station.channel:
			if rstr := e.Data.(string); rstr != expect.(string) {
				errorList = append(errorList, fmt.Sprintf("wrong string '%s', want '%s'", rstr, expect.(string)))
			}
		case <-timer:
			if expect != nil {
				errorList = append(errorList, fmt.Sprintf("timeout! want '%s'", expect.(string)))
			}
		}
		done.Done()
	}

	done.Add(3)
	msg := "Hello Fractal!"
	go recvAndCheck(station1, nil)
	go recvAndCheck(station2, nil)
	go recvAndCheck(station3, nil)
	go SendTo(nil, nil, P2PRouterTestString, msg)
	done.Wait()

	done.Add(3)
	go recvAndCheck(station1, "1")
	go recvAndCheck(station2, "2")
	go recvAndCheck(station3, "3")
	go SendTo(nil, GetStationByName("TestStation1"), P2PRouterTestString, "1")
	go SendTo(nil, GetStationByName("TestStation2"), P2PRouterTestString, "2")
	go SendTo(nil, GetStationByName("TestStation3"), P2PRouterTestString, "3")
	done.Wait()
	sub1.Unsubscribe()
	sub2.Unsubscribe()
	sub3.Unsubscribe()
	errStr := "\n"
	for _, e := range errorList {
		errStr += fmt.Sprintln(e)
	}
	if errStr != "\n" {
		t.Fatal(errStr)
	}
}

func TestSendEvent(t *testing.T) {
	var (
		done    sync.WaitGroup
		nsubs   = 10
		mutex   sync.RWMutex
		receive = 0
	)
	subscriber := func(ch chan *Event) {
		<-ch
		mutex.Lock()
		receive++
		mutex.Unlock()
		done.Done()
	}
	done.Add(nsubs)

	num := int(1)
	for i := 0; i < nsubs; i++ {
		ch := make(chan *Event)
		Subscribe(nil, ch, P2PRouterTestInt, num)
		go subscriber(ch)
	}

	event := &Event{
		Typecode: P2PRouterTestInt,
		Data:     num,
	}
	nsend := SendEvent(event)

	done.Wait()

	if nsend != nsubs {
		t.Fatalf("wrong int %d, want %d", nsubs, nsend)
	}

	if receive != nsubs {
		t.Fatalf("wrong int %d, want %d", nsubs, receive)
	}

	typ := GetTypeByCode(P2PRouterTestInt)
	if reflect.TypeOf(event.Data) != typ {
		t.Fatalf("wrong type")
	}
}

func TestUnsubscribe(t *testing.T) {
	var (
		done  sync.WaitGroup
		nsubs = 1000
	)

	subscriber := func(sub Subscription) {
		sub.Unsubscribe()
		done.Done()
	}

	done.Add(nsubs)
	for i := 0; i < nsubs; i++ {
		ch := make(chan *Event)
		sub := Subscribe(nil, ch, P2PRouterTestInt64, reflect.Int64)
		go subscriber(sub)
	}

	event := &Event{
		Typecode: P2PRouterTestInt64,
		Data:     int64(1),
	}
	SendEvent(event)
	done.Wait()
}
