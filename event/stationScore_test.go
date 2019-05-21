package event

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDelete(t *testing.T) {
	s := NewRemoteStation("01234567", nil)
	StationRegister(s)
	AddThread(s, 1)
	StationUnregister(s)
	AddThread(s, -1)
}

func TestEval(t *testing.T) {
	doAddAsync := func(s Station, done *sync.WaitGroup) {
		done.Add(6)
		go func() { AddThread(s, 100); done.Done() }()
		go func() { AddAck(s, time.Second); done.Done() }()
		go func() { AddCPU(s, time.Second); done.Done() }()
		go func() { AddNetIn(s, 100); done.Done() }()
		go func() { AddNetOut(s, 100); done.Done() }()
		go func() { AddErr(s, 100); done.Done() }()
		done.Done()
	}
	s := NewRemoteStation("01234567", nil)
	StationRegister(s)
	defer StationUnregister(s)
	var done sync.WaitGroup
	n := uint64(0)
	for n = 0; n < 100; n++ {
		done.Add(1)
		go doAddAsync(s, &done)
	}
	done.Wait()
	if got := Thread(s); got != n*100 {
		t.Errorf("wrong thread, want %d got %d", n*100, got)
	}
	AddThread(s, -int64(n*100))
	if got := Thread(s); got != 0 {
		t.Errorf("wrong thread, want %d got %d", 0, got)
	}
	if num, got := Ack(s); uint64(got) != n*uint64(time.Second) || num != n {
		t.Errorf("wrong thread, want (%d, %d) got (%d, %d)", n, n*uint64(time.Second), num, got)
	}
	if got := CPU(s); uint64(got) != n*uint64(time.Second) {
		t.Errorf("wrong thread, want %d got %d", n*uint64(time.Second), got)
	}
	if got := NetIn(s); got != n*100 {
		t.Errorf("wrong thread, want %d got %d", n*100, got)
	}
	if got := NetOut(s); got != n*100 {
		t.Errorf("wrong thread, want %d got %d", n*100, got)
	}
	if got := Err(s); got != n*100 {
		t.Errorf("wrong thread, want %d got %d", n*100, got)
	}
}

func TestScore(t *testing.T) {
	var done sync.WaitGroup
	var done1 sync.WaitGroup
	slist := make([]Station, 20)
	for i := range slist {
		s := NewRemoteStation(fmt.Sprintf("%08x", i), nil)
		done.Add(1)
		done1.Add(8)
		go func(i int) {
			n := time.Duration(i)
			StationRegister(s)
			go func() { AddThread(s, 1); done1.Done() }()
			go func() { AddCPU(s, n*time.Microsecond); done1.Done() }()
			go func() { AddAck(s, n*time.Millisecond); done1.Done() }()
			go func() { AddErr(s, 1); done1.Done() }()
			go func() { AddNetIn(s, 1); done1.Done() }()
			go func() { AddNetOut(s, 1); done1.Done() }()
			go func() { WorstStation(); done1.Done() }()
			go func() { Score(s); done1.Done() }()
			done.Done()
		}(i)
		slist[i] = s
	}
	done.Wait()
	done.Add(len(slist))
	for _, s := range slist {
		go func(s Station) {
			StationUnregister(s)
			done.Done()
		}(s)
	}
	done.Wait()
	done1.Wait()
}
