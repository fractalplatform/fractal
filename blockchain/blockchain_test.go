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
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/txpool"
)

// So we can deterministically seed different blockchains
var (
	canonicalSeed = 1
	forkSeed      = 2
)

func TestTheLastBlock(t *testing.T) {
	printLog(log.LvlDebug)

	genesis := DefaultGenesis()
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	_, blocks := makeNewChain(t, genesis, chain, 10, canonicalSeed)

	// check chain block hash
	checkBlocksInsert(t, chain, blocks)
}

func TestSystemForkChain(t *testing.T) {
	printLog(log.LvlDebug)

	genesis := DefaultGenesis()
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	chain, _ = makeNewChain(t, genesis, chain, 10, canonicalSeed)

	// generate fork blocks
	forkChain := newCanonical(t, genesis)
	defer forkChain.Stop()

	_, forkBlocks := makeNewChain(t, genesis, forkChain, 11, forkSeed)
	_, err := chain.InsertChain(forkBlocks)
	if err != nil {
		t.Fatal(err)
	}

	// check chain block hash
	checkBlocksInsert(t, chain, forkBlocks)

	// check if is complete block chain
	checkCompleteChain(t, chain)
}

func TestBadBlockHashes(t *testing.T) {
	genesis := DefaultGenesis()
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	_, blocks := makeNewChain(t, genesis, chain, 10, canonicalSeed)

	chain.badHashes[blocks[2].Header().Hash()] = true

	_, err := chain.InsertChain(blocks)
	if err != ErrBlacklistedHash {
		t.Errorf("error mismatch: have: %v, want: %v", err, ErrBlacklistedHash)
	}

	// test NewBlockChain()badblock err
	delete(chain.badHashes, blocks[2].Header().Hash())

	_, err = chain.InsertChain(blocks)
	if err != nil {
		t.Fatal(err)
	}

	newChain, err := NewBlockChain(chain.db, false, vm.Config{}, chain.chainConfig,
		[]string{blocks[2].Header().Hash().String()}, 0, txpool.SenderCacher)
	if err != nil {
		t.Fatal(err)
	}
	defer newChain.Stop()

	// if db have bad block then block will be reset, newChain.CurrentBlock().Hash() must equal chain or newchain genesis block hash.
	if newChain.CurrentBlock().Hash() != chain.GetBlockByNumber(0).Hash() ||
		newChain.CurrentBlock().Hash() != newChain.GetBlockByNumber(0).Hash() {
		t.Fatalf("cur hash %x , genesis hash %v", newChain.CurrentBlock().Hash(),
			newChain.Genesis().Hash())
	}

	t.Log(newChain.CurrentBlock().Hash().String())
	t.Log(newChain.Genesis().Hash().String())
}
