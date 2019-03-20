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
	"container/heap"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	mapset "github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/types"
)

var (
	emptyHash = common.Hash{}
)

const (
	maxKnownBlocks = 1024 // Maximum block hashes to keep in the known list (prevent DOS)
)

type stationStatus struct {
	station  router.Station
	latest   unsafe.Pointer // *NewBlockHashesData
	ancestor uint64
	errCh    chan struct{}
}

func (status *stationStatus) updateStatus(news *NewBlockHashesData) {
	atomic.StorePointer(&status.latest, unsafe.Pointer(news))
}

func (status *stationStatus) getStatus() *NewBlockHashesData {
	latest := atomic.LoadPointer(&status.latest)
	return (*NewBlockHashesData)(latest)
}

// Downloader for blockchain sync block.
type Downloader struct {
	station         router.Station
	statusCh        chan *router.Event
	remotes         *stack //map[string]*stationStatus
	remotesMutex    sync.RWMutex
	blockchain      *BlockChain
	downloading     int32
	downloadTrigger chan struct{}
	// bloom           HashBloom
	maxNumber   uint64
	knownBlocks mapset.Set
}

// NewDownloader .
func NewDownloader(chain *BlockChain) *Downloader {
	dl := &Downloader{
		station:    router.NewLocalStation("downloader", nil),
		statusCh:   make(chan *router.Event),
		blockchain: chain,
		remotes: &stack{cmp: func(a, b interface{}) int {
			ra, rb := a.(*stationStatus), b.(*stationStatus)
			return rb.getStatus().TD.Cmp(ra.getStatus().TD)
		}},
		downloadTrigger: make(chan struct{}, 1),
		knownBlocks:     mapset.NewSet(),
	}
	go dl.syncstatus()
	go dl.loop()
	return dl
}

func (dl *Downloader) broadcastStatus(blockhash *NewBlockHashesData) {
	sign := struct {
		Hash     common.Hash
		Complete bool
	}{blockhash.Hash, blockhash.Completed}

	if blockhash.Number <= dl.maxNumber && dl.knownBlocks.Contains(sign) {
		return
	}
	if sign.Complete {
		sign.Complete = false
		dl.knownBlocks.Remove(sign)
		sign.Complete = true
	}

	for dl.knownBlocks.Cardinality() >= maxKnownBlocks {
		dl.knownBlocks.Pop()
	}
	dl.knownBlocks.Add(sign)

	dl.maxNumber = blockhash.Number
	go router.SendTo(nil, router.GetStationByName("broadcast"), router.P2PNewBlockHashesMsg, blockhash)
}

func (dl *Downloader) syncstatus() {
	router.Subscribe(nil, dl.statusCh, router.P2PNewBlockHashesMsg, &NewBlockHashesData{})
	router.Subscribe(nil, dl.statusCh, router.NewMinedEv, NewMinedBlockEvent{})
	for {
		e := <-dl.statusCh
		// NewMinedEv
		if e.Typecode == router.NewMinedEv {
			block := e.Data.(NewMinedBlockEvent).Block
			for dl.knownBlocks.Cardinality() >= maxKnownBlocks {
				dl.knownBlocks.Pop()
			}
			dl.knownBlocks.Add(block.Hash())
			dl.broadcastStatus(&NewBlockHashesData{
				Hash:      block.Hash(),
				Number:    block.NumberU64(),
				TD:        dl.blockchain.GetTd(block.Hash(), block.NumberU64()),
				Completed: true,
			})
			continue
		}
		// NewBlockHashesMsg
		hashdata := e.Data.(*NewBlockHashesData)
		if hashdata.Completed {
			dl.updateStationStatus(e.From.Name(), hashdata)
		}

		head := dl.blockchain.CurrentBlock()
		if hashdata.TD.Cmp(dl.blockchain.GetTd(head.Hash(), head.NumberU64())) > 0 {
			dl.loopStart()
			hashdata.Completed = false
			dl.broadcastStatus(hashdata)
		}
	}
}

func (dl *Downloader) getStationStatus(nameID string) (int, *stationStatus) {
	dl.remotesMutex.RLock()
	defer dl.remotesMutex.RUnlock()
	for i, v := range dl.remotes.data {
		status := v.(*stationStatus)
		if status.station.Name() == nameID {
			return i, status
		}
	}
	return -1, nil
}

func (dl *Downloader) updateStationStatus(nameID string, news *NewBlockHashesData) {
	dl.remotesMutex.Lock()
	defer dl.remotesMutex.Unlock()
	for i, v := range dl.remotes.data {
		status := v.(*stationStatus)
		if status.station.Name() == nameID {
			dl.remotes.remove(i)
			status.updateStatus(news)
			dl.remotes.push(status)
			return
		}
	}
}

func (dl *Downloader) setStationStatus(status *stationStatus) {
	dl.remotesMutex.Lock()
	dl.remotes.push(status)
	dl.remotesMutex.Unlock()
}

// AddStation .
func (dl *Downloader) AddStation(station router.Station, td *big.Int, number uint64, hash common.Hash) {
	status := &stationStatus{
		station: station,
		errCh:   make(chan struct{}),
	}
	status.updateStatus(&NewBlockHashesData{
		Hash:      hash,
		TD:        td,
		Number:    number,
		Completed: true,
	})
	dl.setStationStatus(status)
	head := dl.blockchain.CurrentBlock()
	if td.Cmp(dl.blockchain.GetTd(head.Hash(), head.NumberU64())) > 0 {
		dl.loopStart()
	}
}

// DelStation .
func (dl *Downloader) DelStation(station router.Station) {
	dl.remotesMutex.Lock()
	defer dl.remotesMutex.Unlock()
	for i, v := range dl.remotes.data {
		status := v.(*stationStatus)
		if status.station.Name() == station.Name() {
			dl.remotes.remove(i)
			close(status.errCh)
			return
		}
	}
}

func (dl *Downloader) bestStation() *stationStatus {
	if dl.remotes.Len() == 0 {
		return nil
	}
	dl.remotesMutex.RLock()
	defer dl.remotesMutex.RUnlock()
	return dl.remotes.min().(*stationStatus)
}

func waitEvent(errch chan struct{}, ch chan *router.Event, timeout time.Duration) (*router.Event, error) {
	timer := time.After(timeout)
	select {
	case e := <-ch:
		return e, nil
	case <-timer:
		return nil, errors.New("timeout")
	case <-errch:
		return nil, errors.New("channel closed")
	}
}

func syncReq(e *router.Event, recvCode int, recvData interface{}, errch chan struct{}) (interface{}, error) {
	ch := make(chan *router.Event)
	sub := router.Subscribe(e.From, ch, recvCode, recvData)
	defer sub.Unsubscribe()
	router.SendEvent(e)
	return waitEvent(errch, ch, 2*time.Second)
}

func getBlockHashes(from router.Station, to router.Station, req *getBlcokHashByNumber, errch chan struct{}) ([]common.Hash, error) {
	ch := make(chan *router.Event)
	sub := router.Subscribe(from, ch, router.P2PBlockHashMsg, []common.Hash{})
	defer sub.Unsubscribe()
	router.SendTo(from, to, router.P2PGetBlockHashMsg, req)
	e, err := waitEvent(errch, ch, time.Second+time.Duration(req.Amount)*(10*time.Millisecond))
	if err != nil {
		return nil, err
	}
	return e.Data.([]common.Hash), nil
}

func getHeaders(from router.Station, to router.Station, req *getBlockHeadersData, errch chan struct{}) ([]*types.Header, error) {
	ch := make(chan *router.Event)
	sub := router.Subscribe(from, ch, router.P2PBlockHeadersMsg, []*types.Header{})
	defer sub.Unsubscribe()
	router.SendTo(from, to, router.P2PGetBlockHeadersMsg, req)
	e, err := waitEvent(errch, ch, time.Second+time.Duration(req.Amount)*(50*time.Millisecond))
	if err != nil {
		return nil, err
	}
	return e.Data.([]*types.Header), nil
}

func getBlocks(from router.Station, to router.Station, hashes []common.Hash, errch chan struct{}) ([]*types.Body, error) {
	ch := make(chan *router.Event)
	sub := router.Subscribe(from, ch, router.P2PBlockBodiesMsg, []*types.Body{})
	defer sub.Unsubscribe()
	router.SendTo(from, to, router.P2PGetBlockBodiesMsg, hashes)
	e, err := waitEvent(errch, ch, time.Second+time.Duration(len(hashes))*time.Second)
	if err != nil {
		return nil, err
	}
	return e.Data.([]*types.Body), nil
}

func (dl *Downloader) findAncestor(from router.Station, to router.Station, headNumber uint64, searchStart uint64, errCh chan struct{}) (uint64, error) {
	if headNumber < 1 {
		return 0, nil
	}
	searchLength := headNumber - searchStart + 1 + 1
	if searchLength > 32 {
		searchLength = 32
	}

	hashes, err := getBlockHashes(from, to, &getBlcokHashByNumber{headNumber, searchLength, 0, true}, errCh)
	if err != nil {
		return 0, err
	}

	for i, hash := range hashes {
		if dl.blockchain.HasBlock(hash, headNumber-uint64(i)) {
			return headNumber - uint64(i), nil
		}
	}
	headNumber -= uint64(len(hashes))
	searchStart /= 2
	// binary search
	for headNumber > 0 {
		var err error
		var luckResult uint64
		searchLength := headNumber - searchStart + 1
		searchResult := sort.Search(int(searchLength), func(n int) bool {
			if err != nil || luckResult != 0 {
				return false // doesn't matter true or false
			}
			targetNumber := uint64(n) + searchStart
			var hashes []common.Hash

			hashes, err = getBlockHashes(from, to, &getBlcokHashByNumber{targetNumber, 2, 0, false}, errCh)
			if err != nil {
				return false // doesn't matter true or false
			}
			if len(hashes) < 1 {
				err = errors.New("wrong length of block hash")
				return false // doesn't matter true or false
			}
			hasBlock0 := dl.blockchain.HasBlock(hashes[0], targetNumber)
			// maybe we're lucky
			if len(hashes) == 2 && hasBlock0 && !dl.blockchain.HasBlock(hashes[1], targetNumber+1) {
				luckResult = targetNumber
				return false // doesn't matter true or false
			}
			// return false: move to right/high block
			// return true:  move to left/low block
			return !hasBlock0
		})
		if err != nil {
			return 0, err
		}
		if luckResult != 0 {
			return luckResult, nil
		}
		if searchResult > 0 {
			return uint64(searchResult) + searchStart - 1, nil
		}
		headNumber = searchStart - 1
		searchStart /= 2
	}
	// genesis block are same
	return 0, nil
}

func (dl *Downloader) multiplexDownload(status *stationStatus) bool {
	log.Debug("multiplexDownload start")
	defer log.Debug("multiplexDownload end")
	if status == nil {
		log.Debug("status == nil")
		return false
	}
	latestStatus := status.getStatus()
	statusHash, statusNumber, statusTD := latestStatus.Hash, latestStatus.Number, latestStatus.TD
	head := dl.blockchain.CurrentBlock()
	if statusTD.Cmp(dl.blockchain.GetTd(head.Hash(), head.NumberU64())) <= 0 {
		log.Debug("statusTD < ", "Local", dl.blockchain.GetTd(head.Hash(), head.NumberU64()), "Number", head.NumberU64(), "R", statusTD, "Number", statusNumber)
		return false
	}

	stationSearch := router.NewLocalStation("downloaderSearch", nil)
	router.StationRegister(stationSearch)
	defer router.StationUnregister(stationSearch)

	headNumber := head.NumberU64()
	if headNumber > statusNumber {
		headNumber = statusNumber
	}
	ancestor, err := dl.findAncestor(stationSearch, status.station, headNumber, status.ancestor+1, status.errCh)
	if err != nil {
		log.Debug(fmt.Sprint("ancestor err", err))
		return false
	}
	downloadStart := ancestor + 1
	downloadAmount := statusNumber - ancestor
	if downloadAmount == 0 {
		log.Debug(fmt.Sprintf("Why-1?:number: head:%d headNumber:%d statusNumber: %d", head.NumberU64(), headNumber, statusNumber))
		log.Debug(fmt.Sprintf("Why-2?:hash: head %x status %x", head.Hash(), statusHash))
		log.Debug(fmt.Sprintf("Why-3?:td: head:%d status: %d", dl.blockchain.GetTd(head.Hash(), head.NumberU64()).Uint64(), statusTD.Uint64()))
		return false
	}
	if downloadAmount > 1024 {
		downloadAmount = 1024
	}
	downloadEnd := ancestor + downloadAmount
	downloadBulk := uint64(64)
	var numbers []uint64
	var hashes []common.Hash
	downloadSkip := downloadBulk
	for i := downloadStart; i <= downloadEnd; i += downloadSkip + 1 {
		numbers = append(numbers, i)
	}
	hashes, err = getBlockHashes(stationSearch, status.station, &getBlcokHashByNumber{
		Number:  downloadStart,
		Amount:  uint64(len(numbers)),
		Skip:    downloadSkip,
		Reverse: false}, status.errCh)
	if err != nil || len(hashes) != len(numbers) {
		log.Debug("getBlockHashes 1 err", "err", err, "len(hashes)", len(hashes), "len(numbers)", len(numbers))
		return false
	}
	if numbers[len(numbers)-1] != downloadEnd {
		numbers = append(numbers, downloadEnd)
		hash, err := getBlockHashes(stationSearch, status.station, &getBlcokHashByNumber{
			Number:  downloadEnd,
			Amount:  1,
			Skip:    0,
			Reverse: false}, status.errCh)
		if err != nil || len(hash) != 1 {
			log.Debug("getBlockHashes 2 err", "len(hash)", len(hash), "err", err)
			return false
		}
		hashes = append(hashes, hash...)
	}
	if len(numbers) == 1 {
		numbers = append(numbers, numbers[0])
		hashes = append(hashes, hashes[0])
	}
	// info1 := fmt.Sprintf("1 head:%d headNumber:%d statusNumber:%d ancestor:%d\n", head.NumberU64(), headNumber, statusNumber, ancestor)
	// log.Debug(info1)
	// info2 := fmt.Sprintf("2 head diff:%d status diff:%d\n", dl.blockchain.GetTd(head.Hash(), head.NumberU64()).Uint64(), statusTD.Uint64())
	// log.Debug(info2)
	// info3 := fmt.Sprintf("3 download start:%d end:%d amount:%d bluk:%d\n", downloadStart, downloadEnd, downloadAmount, downloadBulk)
	// log.Debug(info3)
	// info4 := fmt.Sprintf("4 numbers:%d hashes:%d\n", len(numbers), len(hashes))
	// log.Debug(info4)
	n, err := dl.assignDownloadTask(hashes, numbers)
	status.ancestor = n
	if err != nil {
		log.Warn(fmt.Sprint("Insert error:", n, err))
	}

	head = dl.blockchain.CurrentBlock()
	if statusTD.Cmp(dl.blockchain.GetTd(head.Hash(), head.NumberU64())) <= 0 {
		dl.broadcastStatus(&NewBlockHashesData{
			Hash:      head.Hash(),
			Number:    head.NumberU64(),
			TD:        dl.blockchain.GetTd(head.Hash(), head.NumberU64()),
			Completed: true,
		})
		return false
	}
	return true
}

func (dl *Downloader) loopStart() {
	select {
	// dl.downloadTrigger's cache is 1
	case dl.downloadTrigger <- struct{}{}:
	default:
	}
}

func (dl *Downloader) loop() {
	download := func() {
		//for status := dl.bestStation(); dl.download(status); {
		for status := dl.bestStation(); dl.multiplexDownload(status); {
		}
	}
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-dl.downloadTrigger:
			download()
			timer.Stop()
			timer.Reset(10 * time.Second)
		case <-timer.C:
			dl.loopStart()
		}
	}
}

func (dl *Downloader) assignDownloadTask(hashes []common.Hash, numbers []uint64) (uint64, error) {
	log.Debug(fmt.Sprint("assingDownloadTask:", len(hashes), len(numbers), numbers))
	workers := &stack{cmp: dl.remotes.cmp}
	dl.remotesMutex.RLock()
	workers.data = append(workers.data, dl.remotes.data...)
	dl.remotesMutex.RUnlock()
	taskes := &stack{
		data: make([]interface{}, 0, len(numbers)-1),
		cmp: func(a, b interface{}) int {
			wa, wb := a.(*downloadTask), b.(*downloadTask)
			return int(wa.startNumber) - int(wb.startNumber)
		},
	}
	resultCh := make(chan *downloadTask)
	for i := len(numbers) - 1; i > 0; i-- {
		taskes.push(&downloadTask{
			startNumber: numbers[i-1],
			startHash:   hashes[i-1],
			endNumber:   numbers[i],
			endHash:     hashes[i],
			result:      resultCh,
		})
	}
	getReadyTask := func() *downloadTask {
		worker := workers.pop()
		if worker == nil {
			return nil
		}
		task := taskes.pop()
		if task == nil {
			workers.push(worker)
			return nil
		}
		task.(*downloadTask).worker = worker.(*stationStatus)
		return task.(*downloadTask)
	}
	maxTask := 16
	taskCount := 0
	doTask := func() {
		for taskCount < maxTask {
			task := getReadyTask()
			if task == nil {
				break
			}
			taskCount++
			go task.Do()
		}
	}
	// todo new station to download
	//var insertWg sync.WaitGroup
	insertList := make(map[uint64][]*types.Block, len(numbers)-1)
	for doTask(); taskCount > 0; doTask() {
		task := <-resultCh
		taskCount--
		if len(task.blocks) == 0 {
			if task.errorTotal > 5 {
				taskes.clear()
				continue
			}
			taskes.push(task)
		} else {
			workers.push(task.worker)
			insertList[task.startNumber] = task.blocks
		}
	}
	for _, start := range numbers[:len(numbers)-1] {
		blocks := insertList[start]
		if blocks == nil {
			return start - 1, nil
		}
		if _, err := dl.blockchain.InsertChain(blocks); err != nil {
			// bug: try again...
			log.Error("bug: try again...")
			time.Sleep(time.Second)
			if index, err := dl.blockchain.InsertChain(blocks); err != nil {
				return blocks[index].NumberU64() - 1, err
			}
		}
	}
	return numbers[len(numbers)-1], nil
}

type downloadTask struct {
	worker      *stationStatus
	startNumber uint64
	startHash   common.Hash
	endNumber   uint64
	endHash     common.Hash
	blocks      []*types.Block     // result blocks, length == 0 means failed
	errorTotal  int                // total error amount
	result      chan *downloadTask // result channel
}

func (task *downloadTask) Do() {

	defer func() {
		task.errorTotal++
		task.result <- task
	}()
	latestStatus := task.worker.getStatus()
	if latestStatus.Number < task.endNumber {
		return
	}
	remote := task.worker.station
	station := router.NewLocalStation("dl"+remote.Name(), nil)
	router.StationRegister(station)
	defer router.StationUnregister(station)

	reqHash := &getBlcokHashByNumber{task.startNumber, 2, task.endNumber - task.startNumber - 1, false}
	if task.endNumber == task.startNumber {
		reqHash.Skip = 0
		reqHash.Amount = 1
	}
	hashes, err := getBlockHashes(station, remote, reqHash, task.worker.errCh)
	if err != nil || len(hashes) != int(reqHash.Amount) ||
		hashes[0] != task.startHash || hashes[len(hashes)-1] != task.endHash {
		log.Debug(fmt.Sprint("err-1:", err, task.startNumber, task.endNumber, len(hashes)))
		if len(hashes) > 0 {
			log.Debug(fmt.Sprintf("0:%x\n0e:%x\ns:%x\nse:%x", hashes[0], hashes[len(hashes)-1], task.startHash, task.endHash))
		}
		return
	}
	downloadAmount := task.endNumber - task.startNumber + 1
	headers, err := getHeaders(station, remote, &getBlockHeadersData{
		hashOrNumber{
			Number: task.startNumber,
		}, downloadAmount, 0, false,
	}, task.worker.errCh)
	if err != nil || len(headers) != int(downloadAmount) {
		log.Debug(fmt.Sprint("err-2:", err, len(headers), downloadAmount))
		return
	}
	if headers[0].Number.Uint64() != task.startNumber || headers[0].Hash() != task.startHash ||
		headers[len(headers)-1].Number.Uint64() != task.endNumber || headers[len(headers)-1].Hash() != task.endHash {
		log.Debug(fmt.Sprintf("e2-1 0d:%d\n0ed:%d\nsd:%d\nsed:%d", headers[0].Number.Uint64(), headers[len(headers)-1].Number.Uint64(), task.startNumber, task.endNumber))
		log.Debug(fmt.Sprintf("e2-2 0:%x\n0e:%x\ns:%x\nse:%x", headers[0].Hash(), headers[len(headers)-1].Hash(), task.startHash, task.endHash))
		return
	}
	for i := 1; i < len(headers); i++ {
		if headers[i].ParentHash != headers[i-1].Hash() || headers[i].Number.Uint64() != headers[i-1].Number.Uint64()+1 {
			log.Debug(fmt.Sprintf("err-3: phash:%x n->phash:%x\npn+1:%d n:%d", headers[i-1].Hash(), headers[i].ParentHash, headers[i-1].Number.Uint64()+1, headers[i].Number.Uint64()))
			return
		}
	}

	reqHashes := make([]common.Hash, 0, len(headers))
	for _, header := range headers {
		if header.Hash() != emptyHash {
			reqHashes = append(reqHashes, header.Hash())
		}
	}

	bodies, err := getBlocks(station, remote, reqHashes, task.worker.errCh)
	if err != nil || len(bodies) != len(reqHashes) {
		log.Debug(fmt.Sprint("err-4:", err, len(bodies), len(reqHashes)))
		return
	}

	blocks := make([]*types.Block, len(headers))
	bodyIndex := 0
	for i, header := range headers {
		if header.Hash() == emptyHash {
			blocks[i] = types.NewBlockWithHeader(header)
		} else {
			blocks[i] = types.NewBlockWithHeader(header).WithBody(bodies[bodyIndex].Transactions)
			bodyIndex++
		}
	}
	task.blocks = blocks
	return
}

type stack struct {
	cmp  func(interface{}, interface{}) int
	data []interface{}
}

func (s *stack) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *stack) Less(i, j int) bool {
	return s.cmp(s.data[i], s.data[j]) < 0
}

func (s *stack) Push(v interface{}) {
	s.data = append(s.data, v)
}

func (s *stack) Pop() interface{} {
	v := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return v
}

func (s *stack) Len() int {
	return len(s.data)
}

func (s *stack) pop() interface{} {
	if len(s.data) == 0 {
		return nil
	}
	return heap.Pop(s)
}

func (s *stack) push(v interface{}) {
	heap.Push(s, v)
}

func (s *stack) remove(i int) {
	if len(s.data) > i {
		heap.Remove(s, i)
	}
}

func (s *stack) min() interface{} {
	if len(s.data) == 0 {
		return nil
	}
	return s.data[0]
}

func (s *stack) clear() {
	s.data = nil
}
