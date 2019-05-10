package event

import (
	"fmt"
	"sync"
	"testing"
)

func TestScore(t *testing.T) {
	var done sync.WaitGroup
	slist := make([]Station, 20)
	for i := range slist {
		s := NewRemoteStation(fmt.Sprintf("%08x", i), nil)
		done.Add(1)
		go func(i int) {
			n := uint64(i)
			StationRegister(s)
			go AddCPU(s, n*1e2)
			go AddAck(s, n*1e4)
			go AddErr(s)
			go AddNetIn(s, 1e3)
			go AddNetOut(s, 1e3)
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
}
