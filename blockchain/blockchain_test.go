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
	"testing"
	"time"
)

func TestTheLastBlock(t *testing.T) {
	// printLog(log.LvlDebug)
	genesis := DefaultGenesis()
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	allCandidates, allHeaderTimes := genCanonicalCandidatesAndTimes(genesis)
	_, blocks := makeNewChain(t, genesis, chain, allCandidates, allHeaderTimes)

	// check chain block hash
	checkBlocksInsert(t, chain, blocks)
}

func TestSystemForkChain(t *testing.T) {
	var (
		allCandidates, allCandidates1   []string
		allHeaderTimes, allHeaderTimes1 []uint64
	)
	// printLog(log.LvlTrace)
	genesis := DefaultGenesis()

	allCandidates, allHeaderTimes = genCanonicalCandidatesAndTimes(genesis)

	allCandidates1 = append(allCandidates1, allCandidates...)
	//allCandidates1 = append(allCandidates1, "syscandidate0")
	//allCandidates1 = append(allCandidates1, params.DefaultChainconfig.SysName)

	allHeaderTimes1 = append(allHeaderTimes1, allHeaderTimes...)
	//allHeaderTimes1 = append(allHeaderTimes1, allHeaderTimes[len(allHeaderTimes)-1]+1000*uint64(time.Millisecond)*3*7)
	//allHeaderTimes1 = append(allHeaderTimes1, allHeaderTimes1[len(allHeaderTimes1)-1]+1000*uint64(time.Millisecond)*3)

	testFork(t, allCandidates, allCandidates1, allHeaderTimes, allHeaderTimes1)
}

func genCanonicalCandidatesAndTimes(genesis *Genesis) ([]string, []uint64) {
	var (
		//dposEpochNum   uint64 = 1
		allCandidates  []string
		allHeaderTimes []uint64
	)

	// geaerate block's candidates and block header time
	// system's candidates headertimes
	sysCandidates, sysHeaderTimes := makeSystemCandidatesAndTime(genesis.Timestamp*uint64(time.Millisecond), genesis)
	allCandidates = append(allCandidates, sysCandidates...)
	allHeaderTimes = append(allHeaderTimes, sysHeaderTimes...)

	// elected candidates headertimes
	// candidates, headerTimes := makeCandidatesAndTime(sysHeaderTimes[len(sysHeaderTimes)-1], genesis, dposEpochNum)
	// allCandidates = append(allCandidates, candidates[:12]...)
	// allHeaderTimes = append(allHeaderTimes, headerTimes[:12]...)

	// // elected candidates headertimes
	// candidates, headerTimes = makeCandidatesAndTime(headerTimes[len(headerTimes)-1], genesis, dposEpochNum)
	// allCandidates = append(allCandidates, candidates[:12]...)
	// allHeaderTimes = append(allHeaderTimes, headerTimes[:12]...)

	return allCandidates, allHeaderTimes
}

func testFork(t *testing.T, candidates, forkCandidates []string, headerTimes, forkHeaderTimes []uint64) {
	genesis := DefaultGenesis()
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	chain, _ = makeNewChain(t, genesis, chain, candidates, headerTimes)

	// generate fork blocks
	blocks := generateForkBlocks(t, DefaultGenesis(), forkCandidates, forkHeaderTimes)

	_, err := chain.InsertChain(blocks)
	if err != nil {
		t.Fatal(err)
	}

	// check chain block hash
	checkBlocksInsert(t, chain, blocks)

	// check if is complete block chain
	checkCompleteChain(t, chain)
}
