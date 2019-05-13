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

package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/blockchain"
	"github.com/fractalplatform/fractal/ftservice"
	"github.com/fractalplatform/fractal/types"
	ldb "github.com/fractalplatform/fractal/utils/fdb/leveldb"
	"github.com/fractalplatform/fractal/utils/rlp"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	importBatchSize = 2500
)

var importCommand = &cobra.Command{
	Use:   "import -d <datadir> -g <genesis.json> <block file name>",
	Short: "Import a blockchain file",
	Long:  "Import a blockchain file",
	Run: func(cmd *cobra.Command, args []string) {
		ftCfgInstance.LogCfg.Setup()
		if err := importChain(args); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(importCommand)
	importCommand.Flags().StringVarP(&ftCfgInstance.NodeCfg.DataDir, "datadir", "d", ftCfgInstance.NodeCfg.DataDir, "Data directory for the databases ")
	importCommand.Flags().StringVarP(&ftCfgInstance.GenesisFile, "genesis", "g", "", "genesis json file")

}

func importChain(args []string) error {
	if len(args) < 1 {
		return errors.New("This command requires an argument")
	}

	stack, err := makeNode()
	if err != nil {
		return err
	}

	ctx := stack.GetNodeConfig()
	ftsrv, err := ftservice.New(ctx, ftCfgInstance.FtServiceCfg)
	if err != nil {
		return err
	}

	// Start periodically gathering memory profiles
	var peakMemAlloc, peakMemSys uint64
	go func() {
		stats := new(runtime.MemStats)
		for {
			runtime.ReadMemStats(stats)
			if atomic.LoadUint64(&peakMemAlloc) < stats.Alloc {
				atomic.StoreUint64(&peakMemAlloc, stats.Alloc)
			}
			if atomic.LoadUint64(&peakMemSys) < stats.Sys {
				atomic.StoreUint64(&peakMemSys, stats.Sys)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	start := time.Now()

	fp := args[0]
	if len(args) == 1 {
		if err := importBlockchain(ftsrv.BlockChain(), fp); err != nil {
			return fmt.Errorf("Import error: %v", err)
		}
	} else {
		for i := 0; i < len(args); i++ {
			if err := importBlockchain(ftsrv.BlockChain(), args[i]); err != nil {
				return fmt.Errorf("Import error: %v, %v", err, args[i])
			}
		}
	}

	log.Info("Import done in ", "time", time.Since(start))

	db := ftsrv.ChainDb().(*ldb.LDBDatabase)
	stats, err := db.LDB().GetProperty("leveldb.stats")
	if err != nil {
		return fmt.Errorf("Failed to read database stats: %v", err)
	}
	fmt.Println(stats)

	ioStats, err := db.LDB().GetProperty("leveldb.iostats")
	if err != nil {
		return fmt.Errorf("Failed to read database iostats: %v", err)
	}
	fmt.Println(ioStats)

	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	fmt.Printf("Object memory: %.3f MB current, %.3f MB peak\n", float64(mem.Alloc)/1024/1024, float64(atomic.LoadUint64(&peakMemAlloc))/1024/1024)
	fmt.Printf("System memory: %.3f MB current, %.3f MB peak\n", float64(mem.Sys)/1024/1024, float64(atomic.LoadUint64(&peakMemSys))/1024/1024)
	fmt.Printf("Allocations:   %.3f million\n", float64(mem.Mallocs)/1000000)
	fmt.Printf("GC pause:      %v\n\n", time.Duration(mem.PauseTotalNs))

	// Compact the entire database to more accurately measure disk io and print the stats
	start = time.Now()
	fmt.Println("Compacting entire database...")
	if err = db.LDB().CompactRange(util.Range{}); err != nil {
		return fmt.Errorf("Compaction failed: %v", err)
	}
	fmt.Printf("Compaction done in %v.\n\n", time.Since(start))

	stats, err = db.LDB().GetProperty("leveldb.stats")
	if err != nil {
		return fmt.Errorf("Failed to read database stats: %v", err)
	}
	fmt.Println(stats)

	ioStats, err = db.LDB().GetProperty("leveldb.iostats")
	if err != nil {
		return fmt.Errorf("Failed to read database iostats: %v", err)
	}
	fmt.Println(ioStats)

	return nil
}

func importBlockchain(chain *blockchain.BlockChain, fn string) error {
	// Watch for Ctrl-C while the import is running.
	// If a signal is received, the import will stop at the next batch.
	interrupt := make(chan os.Signal, 1)
	stop := make(chan struct{})
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	defer close(interrupt)
	go func() {
		if _, ok := <-interrupt; ok {
			log.Info("Interrupted during import, stopping at next batch")
		}
		close(stop)
	}()
	checkInterrupt := func() bool {
		select {
		case <-stop:
			return true
		default:
			return false
		}
	}

	log.Info("Importing blockchain", "file", fn)

	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close()

	var reader io.Reader = fh
	if strings.HasSuffix(fn, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return err
		}
	}
	stream := rlp.NewStream(reader, 0)

	// Run actual the import.
	blocks := make(types.Blocks, importBatchSize)
	n := 0
	for batch := 0; ; batch++ {
		// Load a batch of RLP blocks.
		if checkInterrupt() {
			return fmt.Errorf("interrupted")
		}
		i := 0
		for ; i < importBatchSize; i++ {
			var b types.Block
			if err := stream.Decode(&b); err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("at block %d: %v", n, err)
			}
			// don't import first block
			if b.NumberU64() == 0 {
				i--
				continue
			}
			blocks[i] = &b
			n++
		}
		if i == 0 {
			break
		}
		// Import the batch.
		if checkInterrupt() {
			return fmt.Errorf("interrupted")
		}
		missing := missingBlocks(chain, blocks[:i])
		if len(missing) == 0 {
			log.Info("Skipping batch as all blocks present", "batch", batch, "first", blocks[0].Hash(), "last", blocks[i-1].Hash())
			continue
		}
		if _, err := chain.InsertChain(missing); err != nil {
			return fmt.Errorf("invalid block %d: %v", n, err)
		}
	}
	return nil
}

func missingBlocks(chain *blockchain.BlockChain, blocks []*types.Block) []*types.Block {
	head := chain.CurrentBlock()
	for i, block := range blocks {
		// If we're behind the chain head, only check block, state is available at head
		if head.NumberU64() > block.NumberU64() {
			if !chain.HasBlock(block.Hash(), block.NumberU64()) {
				return blocks[i:]
			}
			continue
		}
		// If we're above the chain head, state availability is a must
		if !chain.HasBlockAndState(block.Hash(), block.NumberU64()) {
			return blocks[i:]
		}
	}
	return nil
}
