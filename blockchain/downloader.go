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
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	mapset "github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	router "github.com/fractalplatform/fractal/event"
	adaptor "github.com/fractalplatform/fractal/p2p/protoadaptor"
	"github.com/fractalplatform/fractal/types"
)

// NewMinedBlockEvent is posted when a block has been imported.
type NewMinedBlockEvent struct{ Block *types.Block }

var (
	emptyHash = common.Hash{}
)

const (
	maxKnownBlocks = 1024 // Maximum block hashes to keep in the known list (prevent DOS)
)

type errorID int

const (
	other errorID = iota
	ioTimeout
	ioClose
	notFind
	sizeNotEqual
	insertError
)

// Error represent error by downloader
type Error struct {
	error
	eid errorID
}

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
	statusCh        chan *router.Event
	remotes         *simpleHeap //map[string]*stationStatus
	remotesMutex    sync.RWMutex
	blockchain      *BlockChain
	quit            chan struct{}
	loopWG          sync.WaitGroup
	downloadTrigger chan struct{}
	// bloom           HashBloom
	maxNumber   uint64
	knownBlocks mapset.Set
	subs        []router.Subscription
}

// NewDownloader create a new downloader
func NewDownloader(chain *BlockChain) *Downloader {
	dl := &Downloader{
		statusCh:   make(chan *router.Event),
		blockchain: chain,
		quit:       make(chan struct{}),
		remotes: &simpleHeap{cmp: func(a, b interface{}) int {
			ra, rb := a.(*stationStatus), b.(*stationStatus)
			return rb.getStatus().TD.Cmp(ra.getStatus().TD)
		}},
		downloadTrigger: make(chan struct{}, 1),
		knownBlocks:     mapset.NewSet(),
		subs:            make([]router.Subscription, 0, 2),
	}
	dl.loopWG.Add(2)
	go dl.syncstatus()
	go dl.loop()
	return dl
}

// Stop stop the downloader
func (dl *Downloader) Stop() {
	log.Info("Downloader stopping...")
	close(dl.quit)
	for _, sub := range dl.subs {
		sub.Unsubscribe()
	}
	for _, v := range dl.remotes.data {
		status := v.(*stationStatus)
		close(status.errCh)
	}
	dl.loopWG.Wait()
	log.Info("Downloader stopped.")
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
	defer dl.loopWG.Done()
	sub1 := router.Subscribe(nil, dl.statusCh, router.P2PNewBlockHashesMsg, &NewBlockHashesData{})
	sub2 := router.Subscribe(nil, dl.statusCh, router.NewMinedEv, NewMinedBlockEvent{})
	dl.subs = append(dl.subs, sub1, sub2)
	for {
		select {
		case <-dl.quit:
			return
		case e := <-dl.statusCh:
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
	dl.remotesMutex.RLock()
	defer dl.remotesMutex.RUnlock()
	if dl.remotes.Len() == 0 {
		return nil
	}
	return dl.remotes.min().(*stationStatus)
}

func waitEvent(errch chan struct{}, ch chan *router.Event, timeout time.Duration) (*router.Event, *Error) {
	timer := time.After(timeout)
	select {
	case e := <-ch:
		return e, nil
	case <-timer:
		return nil, &Error{errors.New("timeout"), ioTimeout}
	case <-errch:
		return nil, &Error{errors.New("channel closed"), ioClose}
	}
}

func syncReq(e *router.Event, recvCode int, recvType interface{}, timeout time.Duration, errch chan struct{}) (*router.Event, *Error) {
	start := time.Now()
	defer func() {
		router.AddAck(e.To, time.Since(start))
	}()
	ch := make(chan *router.Event)
	sub := router.Subscribe(e.From, ch, recvCode, recvType)
	defer sub.Unsubscribe()
	router.SendEvent(e)
	return waitEvent(errch, ch, timeout)
}

func getBlockHashes(from router.Station, to router.Station, req *getBlockHashByNumber, errch chan struct{}) ([]common.Hash, *Error) {
	se := &router.Event{
		From:     from,
		To:       to,
		Typecode: router.P2PGetBlockHashMsg,
		Data:     req,
	}
	timeout := time.Second + time.Duration(req.Amount)*(10*time.Millisecond)
	e, err := syncReq(se, router.P2PBlockHashMsg, []common.Hash{}, timeout, errch)
	if err != nil {
		return nil, err
	}
	hashes := e.Data.([]common.Hash)
	if len(hashes) != int(req.Amount) {
		return hashes, &Error{fmt.Errorf("wrong size, expected %d got %d", req.Amount, len(hashes)), sizeNotEqual}
	}
	return hashes, nil
}

func getHeaders(from router.Station, to router.Station, req *getBlockHeadersData, errch chan struct{}) ([]*types.Header, *Error) {
	se := &router.Event{
		From:     from,
		To:       to,
		Typecode: router.P2PGetBlockHeadersMsg,
		Data:     req,
	}
	timeout := time.Second + time.Duration(req.Amount)*(50*time.Millisecond)
	e, err := syncReq(se, router.P2PBlockHeadersMsg, []*types.Header{}, timeout, errch)
	if err != nil {
		return nil, err
	}
	headers := e.Data.([]*types.Header)
	if len(headers) != int(req.Amount) {
		return headers, &Error{fmt.Errorf("wrong size, expected %d got %d", req.Amount, len(headers)), sizeNotEqual}
	}
	return headers, nil
}

func getBlocks(from router.Station, to router.Station, req []common.Hash, errch chan struct{}) ([]*types.Body, *Error) {
	se := &router.Event{
		From:     from,
		To:       to,
		Typecode: router.P2PGetBlockBodiesMsg,
		Data:     req,
	}
	timeout := time.Second + time.Duration(len(req))*(100*time.Millisecond)
	e, err := syncReq(se, router.P2PBlockBodiesMsg, []*types.Body{}, timeout, errch)
	if err != nil {
		return nil, err
	}
	bodies := e.Data.([]*types.Body)
	if len(bodies) != len(req) {
		return bodies, &Error{fmt.Errorf("wrong size, expected %d got %d", len(req), len(bodies)), sizeNotEqual}
	}
	return bodies, nil
}

func (dl *Downloader) findAncestor(from router.Station, to router.Station, headNumber uint64, preAncestor uint64, errCh chan struct{}) (uint64, common.Hash, *Error) {
	if headNumber < 1 {
		return 0, dl.blockchain.Genesis().Hash(), nil
	}
	find := func(headNum, length uint64) (uint64, common.Hash, *Error) {
		hashes, err := getBlockHashes(from, to, &getBlockHashByNumber{headNumber, length, 0, true}, errCh)
		if err != nil {
			return 0, emptyHash, err
		}

		for i, hash := range hashes {
			if dl.blockchain.HasBlock(hash, headNum-uint64(i)) {
				log.Debug("downloader findAncestor", "hash", hash.Hex(), "number", headNum-uint64(i))
				return headNum - uint64(i), hash, nil
			}
		}
		return 0, emptyHash, &Error{errors.New("not find"), notFind}
	}

	irreversibleNumber := dl.blockchain.IrreversibleNumber()
	log.Debug("downloader findAncestor", "headNumber", headNumber, "preAncestor", preAncestor, "irreversibleNumber", irreversibleNumber)
	if preAncestor < irreversibleNumber {
		preAncestor = irreversibleNumber
	}
	searchLength := headNumber - preAncestor + 1
	if searchLength > 32 {
		searchLength = 32
	}
	for headNumber >= irreversibleNumber {
		ancestor, hash, err := find(headNumber, searchLength)
		if err == nil {
			return ancestor, hash, nil
		}
		if err != nil && err.eid != notFind {
			return 0, hash, err
		}
		headNumber -= searchLength
		searchLength = headNumber - irreversibleNumber + 1
		if searchLength > 32 {
			searchLength = 32
		}
	}
	return 0, emptyHash, &Error{fmt.Errorf("can not find ancestor after irreversibleNumber:%d", irreversibleNumber), notFind}
}

func (dl *Downloader) shortcutDownload(status *stationStatus, startNumber uint64, startHash common.Hash, endNumber uint64, endHash common.Hash) (uint64, *Error) {
	resultCh := make(chan *downloadTask)
	go (&downloadTask{
		worker:      status,
		startNumber: startNumber,
		startHash:   startHash,
		endNumber:   endNumber,
		endHash:     endHash,
		result:      resultCh,
	}).Do()

	task := <-resultCh

	if len(task.blocks) == 0 {
		return startNumber, task.err
	}

	if index, err := dl.blockchain.InsertChain(task.blocks); err != nil {
		return task.blocks[index].NumberU64() - 1, &Error{err, insertError}
	}
	return endNumber, nil
}

// return true means need call again
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

	log.Debug("downloader station:", "node", adaptor.GetFnode(status.station))
	log.Debug("downloader statusTD x ", "Local", dl.blockchain.GetTd(head.Hash(), head.NumberU64()), "Number", head.NumberU64(), "R", statusTD, "Number", statusNumber)

	headNumber := head.NumberU64()
	if headNumber < statusNumber && statusNumber < headNumber+6 {
		_, err := dl.shortcutDownload(status, headNumber, head.Hash(), statusNumber, statusHash)
		if err == nil { // download and insert completed
			head = dl.blockchain.CurrentBlock()
			dl.broadcastStatus(&NewBlockHashesData{
				Hash:      head.Hash(),
				Number:    head.NumberU64(),
				TD:        dl.blockchain.GetTd(head.Hash(), head.NumberU64()),
				Completed: true,
			})
			return false
		}
		if err.eid == insertError || err.eid == sizeNotEqual { // download failed because of the remote's error
			log.Warn("Disconnect because some error:", "node:", adaptor.GetFnode(status.station), "err", err)
			router.SendTo(nil, nil, router.OneMinuteLimited, status.station) // disconnect and put into blacklist
			return true
		}
		// download failed, continue download from other peers.
	}
	if headNumber > statusNumber {
		headNumber = statusNumber
	}

	rand.Seed(time.Now().UnixNano())
	stationSearch := router.NewLocalStation(fmt.Sprintf("downloaderSearch%d", rand.Int()), nil)
	router.StationRegister(stationSearch)
	defer router.StationUnregister(stationSearch)

	ancestor, ancestorHash, err := dl.findAncestor(stationSearch, status.station, headNumber, status.ancestor, status.errCh)
	if err != nil {
		log.Warn("ancestor err", "err", err, "errID:", err.eid)
		if err.eid == notFind {
			log.Warn("Disconnect because ancestor not find:", "node:", adaptor.GetFnode(status.station))
			router.SendTo(nil, nil, router.OneMinuteLimited, status.station) // disconnect and put into blacklist
		}
		return false
	}
	log.Debug("downloader ancestor:", "ancestor", ancestor)
	downloadStart := ancestor
	downloadStartHash := ancestorHash
	downloadAmount := statusNumber - ancestor
	if downloadAmount == 0 { // maybe the status of remote was changed
		return false
	}
	if downloadAmount > 1024 {
		downloadAmount = 1024
	}
	downloadEnd := downloadStart + downloadAmount
	downloadBulk := uint64(64)
	numbers := make([]uint64, 0, (downloadAmount+downloadBulk-1)/downloadBulk+1)
	hashes := make([]common.Hash, 0, (downloadAmount+downloadBulk-1)/downloadBulk+1)
	downloadSkip := downloadBulk - 1 // f(n+1) = f(n) + 1 + skip

	for i := downloadStart; i <= downloadEnd; i += downloadSkip + 1 {
		numbers = append(numbers, i)
	}
	hashes = append(hashes, downloadStartHash)

	if len(numbers[1:]) > 0 {
		hash, err := getBlockHashes(stationSearch, status.station, &getBlockHashByNumber{
			Number:  numbers[1],
			Amount:  uint64(len(numbers[1:])),
			Skip:    downloadSkip,
			Reverse: false}, status.errCh)
		if err != nil || len(hash) != len(numbers[1:]) {
			log.Debug("getBlockHashes 1 err", "err", err, "len(hash)", len(hash), "len(numbers)", len(numbers[1:]))
			return false
		}
		hashes = append(hashes, hash...)
	}

	if numbers[len(numbers)-1] != downloadEnd {
		numbers = append(numbers, downloadEnd)
		hash, err := getBlockHashes(stationSearch, status.station, &getBlockHashByNumber{
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

	n, err := dl.assignDownloadTask(hashes, numbers)
	status.ancestor = n
	if err != nil {
		log.Warn("Insert error:", "number:", n, "error", err)
		failedNum := numbers[len(numbers)-1] - n
		router.AddErr(status.station, failedNum)
		if failedNum > 32 {
			log.Warn("Disconnect because Insert error:", "node:", adaptor.GetFnode(status.station), "failedNum", failedNum)
			router.SendTo(nil, nil, router.OneMinuteLimited, status.station) // disconnect and put into blacklist
		}
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
	defer dl.loopWG.Done()
	download := func() {
		//for status := dl.bestStation(); dl.download(status); {
		for status := dl.bestStation(); dl.multiplexDownload(status); {
		}
	}
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-dl.quit:
			return
		case <-dl.downloadTrigger:
			download()
			timer.Stop()
			timer.Reset(10 * time.Second)
		case <-timer.C:
			dl.loopStart()
		}
	}
}

// Return the height of the last successfully inserted block and error
func (dl *Downloader) assignDownloadTask(hashes []common.Hash, numbers []uint64) (uint64, *Error) {
	log.Debug("assingDownloadTask:", "hashesLen", len(hashes), "numbersLen", len(numbers), "numbers", numbers)
	workers := &simpleHeap{cmp: dl.remotes.cmp}
	dl.remotesMutex.RLock()
	workers.data = append(workers.data, dl.remotes.data...)
	dl.remotesMutex.RUnlock()
	tasks := &simpleHeap{
		data: make([]interface{}, 0, len(numbers)-1),
		cmp: func(a, b interface{}) int {
			wa, wb := a.(*downloadTask), b.(*downloadTask)
			return int(wa.startNumber) - int(wb.startNumber)
		},
	}
	resultCh := make(chan *downloadTask)
	for i := len(numbers) - 1; i > 0; i-- {
		tasks.push(&downloadTask{
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
		task := tasks.pop()
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
				tasks.clear()
				continue
			}
			tasks.push(task)
		} else {
			workers.push(task.worker)
			insertList[task.startNumber] = task.blocks
		}
	}
	for _, start := range numbers[:len(numbers)-1] {
		blocks := insertList[start]
		if blocks == nil {
			return start, nil
		}
		if index, err := dl.blockchain.InsertChain(blocks); err != nil {
			return blocks[index].NumberU64() - 1, &Error{err, other}
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
	err         *Error
}

func (task *downloadTask) Do() {
	var err *Error
	var headers []*types.Header
	var bodies []*types.Body

	latestStatus := task.worker.getStatus()
	defer func() {
		task.err = err
		task.result <- task
		diff := latestStatus.Number - task.endNumber
		if latestStatus.Number < task.endNumber {
			diff = task.endNumber - latestStatus.Number
		}
		if len(task.blocks) == 0 && diff > 16 {
			task.errorTotal++
			router.AddErr(task.worker.station, 1)
		}
	}()
	if latestStatus.Number < task.endNumber {
		return
	}
	remote := task.worker.station
	rand.Seed(time.Now().UnixNano())
	station := router.NewLocalStation(fmt.Sprintf("dl%d%s", rand.Int(), remote.Name()), nil)
	router.StationRegister(station)
	defer router.StationUnregister(station)

	reqHash := &getBlockHashByNumber{task.startNumber, 2, task.endNumber - task.startNumber - 1, false}
	if task.endNumber == task.startNumber {
		reqHash.Skip = 0
		reqHash.Amount = 1
	}
	/*
		hashes, err := getBlockHashes(station, remote, reqHash, task.worker.errCh)
		if err != nil || len(hashes) != int(reqHash.Amount) ||
			hashes[0] != task.startHash || hashes[len(hashes)-1] != task.endHash {
			log.Debug(fmt.Sprint("err-1:", err, task.startNumber, task.endNumber, len(hashes)))
			if len(hashes) > 0 {
				log.Debug(fmt.Sprintf("0:%x\n0e:%x\ns:%x\nse:%x", hashes[0], hashes[len(hashes)-1], task.startHash, task.endHash))
			}
			return
		}
	*/
	downloadAmount := task.endNumber - task.startNumber
	headers, err = getHeaders(station, remote, &getBlockHeadersData{
		hashOrNumber{
			Number: task.startNumber + 1,
		}, downloadAmount, 0, false,
	}, task.worker.errCh)
	if err != nil || len(headers) != int(downloadAmount) {
		log.Debug("download header failed",
			"err", err,
			"recvAmount", len(headers),
			"taskAmount", downloadAmount,
		)
		return
	}
	if headers[0].Number.Uint64() != task.startNumber+1 || headers[0].ParentHash != task.startHash ||
		headers[len(headers)-1].Number.Uint64() != task.endNumber || headers[len(headers)-1].Hash() != task.endHash {
		log.Debug("download header don't match task",
			"recv.Start.Number", headers[0].Number.Uint64(),
			"recv.End.Number", headers[len(headers)-1].Number.Uint64(),
			"recv.Start.ParentHash", headers[0].ParentHash,
			"recv.End.Hash", headers[len(headers)-1].Hash(),
			"task.Start.Number", task.startNumber,
			"task.End.Number", task.endNumber,
			"task.Start.Hash", task.startHash,
			"task.End.Hash", task.endHash,
		)
		return
	}
	for i := 1; i < len(headers); i++ {
		if headers[i].ParentHash != headers[i-1].Hash() || headers[i].Number.Uint64() != headers[i-1].Number.Uint64()+1 {
			log.Debug("download headers are discontinuous",
				"parent.number", headers[i-1].Number.Uint64(),
				"parent.hash", headers[i-1].Hash(),
				"n.number", headers[i].Number.Uint64(),
				"n.parentHash", headers[i].ParentHash,
			)
			return
		}
	}

	reqHashes := make([]common.Hash, 0, len(headers))
	for _, header := range headers {
		if header.TxsRoot != emptyHash {
			reqHashes = append(reqHashes, header.Hash())
		}
	}

	bodies, err = getBlocks(station, remote, reqHashes, task.worker.errCh)
	if err != nil || len(bodies) != len(reqHashes) {
		log.Debug("download blocks failed",
			"err", err,
			"recvAmount", len(bodies),
			"taskAmount", len(reqHashes))
		return
	}

	blocks := make([]*types.Block, len(headers))
	bodyIndex := 0
	for i, header := range headers {
		if header.TxsRoot == emptyHash {
			blocks[i] = types.NewBlockWithHeader(header)
		} else {
			blocks[i] = types.NewBlockWithHeader(header).WithBody(bodies[bodyIndex].Transactions)
			bodyIndex++
		}
	}
	task.blocks = blocks
}
