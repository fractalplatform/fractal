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

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/params"
)

func TestTheLastBlock(t *testing.T) {
	// printLog(log.LvlDebug)
	genesis := DefaultGenesis()
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	allCadidates, allHeaderTimes := genCanonicalCadidatesAndTimes(genesis)
	_, blocks := makeNewChain(t, genesis, chain, allCadidates, allHeaderTimes)

	// check chain block hash
	checkBlocksInsert(t, chain, blocks)
}

func TestSystemForkChain(t *testing.T) {
	var (
		allCadidates, allCadidates1 []string
		allHeaderTimes              []uint64
	)

	//printLog(log.LvlTrace)
	genesis := DefaultGenesis()

	allCadidates, allHeaderTimes = genCanonicalCadidatesAndTimes(genesis)

	allCadidates1 = append(allCadidates1, allCadidates...)

	allCadidates1[len(allCadidates1)-1] = params.DefaultChainconfig.SysName.String()

	testFork(t, allCadidates, allCadidates1, allHeaderTimes, allHeaderTimes)
}

func TestOtherCadidatesForkSystemChain(t *testing.T) {
	var (
		allCadidates, allCadidates1     []string
		allHeaderTimes, allHeaderTimes1 []uint64
	)

	printLog(log.LvlWarn)
	genesis := DefaultGenesis()

	allCadidates, allHeaderTimes = genCanonicalCadidatesAndTimes(genesis)
	allCadidates1 = append(allCadidates1, allCadidates...)
	allHeaderTimes1 = append(allHeaderTimes1, allHeaderTimes...)

	allCadidates = allCadidates[0 : len(allCadidates)-1]
	allHeaderTimes = allHeaderTimes[0 : len(allHeaderTimes)-1]

	allCadidates[len(allCadidates)-1] = params.DefaultChainconfig.SysName.String()

	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	chain, _ = makeNewChain(t, genesis, chain, allCadidates, allHeaderTimes)

	// generate fork blocks
	blocks := generateForkBlocks(t, DefaultGenesis(), allCadidates1, allHeaderTimes1)

	_, err := chain.InsertChain(blocks)
	if err != nil {
		t.Error(err)
	}

	// check if is complete block chain
	checkCompleteChain(t, chain)

	if chain.CurrentBlock().Coinbase().String() != genesis.Config.SysName.String() {
		t.Fatalf("other cadidate:%v  ,can't fork the system: %v ", chain.CurrentBlock().Coinbase(), genesis.Config.SysName)
	}

}

func genCanonicalCadidatesAndTimes(genesis *Genesis) ([]string, []uint64) {
	var (
		dposEpochNum   uint64 = 1
		allCadidates   []string
		allHeaderTimes []uint64
	)

	// geaerate block's cadidates and block header time
	// system's cadidates headertimes
	sysCadidates, sysHeaderTimes := makeSystemCadidatesAndTime(genesis.Timestamp, genesis)
	allCadidates = append(allCadidates, sysCadidates...)
	allHeaderTimes = append(allHeaderTimes, sysHeaderTimes...)

	// elected cadidates headertimes
	cadidates, headerTimes := makeCadidatesAndTime(sysHeaderTimes[len(sysHeaderTimes)-1], genesis, dposEpochNum)
	allCadidates = append(allCadidates, cadidates...)
	allHeaderTimes = append(allHeaderTimes, headerTimes...)

	return allCadidates, allHeaderTimes
}

func testFork(t *testing.T, cadidates, forkCadidates []string, headerTimes, forkHeaderTimes []uint64) {
	genesis := DefaultGenesis()
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	chain, _ = makeNewChain(t, genesis, chain, cadidates, headerTimes)
	// generate fork blocks
	blocks := generateForkBlocks(t, DefaultGenesis(), forkCadidates, forkHeaderTimes)

	_, err := chain.InsertChain(blocks)
	if err != nil {
		t.Error(err)
	}

	// check chain block hash
	checkBlocksInsert(t, chain, blocks)

	// check if is complete block chain
	checkCompleteChain(t, chain)
}
