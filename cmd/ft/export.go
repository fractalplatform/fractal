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
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/blockchain"
	"github.com/fractalplatform/fractal/ftservice"
	"github.com/spf13/cobra"
)

var exportCommand = &cobra.Command{
	Use:   "export -d <datadir> <block file name> <start num> <end num>",
	Short: "Export blockchain to file",
	Long:  "Export blockchain to file",
	Run: func(cmd *cobra.Command, args []string) {
		ftCfgInstance.LogCfg.Setup()
		if err := exportChain(args); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(exportCommand)
	exportCommand.Flags().StringVarP(&ftCfgInstance.NodeCfg.DataDir, "datadir", "d", ftCfgInstance.NodeCfg.DataDir, "Data directory for the databases ")
}

func exportChain(args []string) error {
	if len(args) < 1 {
		return errors.New("This command requires an argument")
	}

	start := time.Now()

	stack, err := makeNode()
	if err != nil {
		return err
	}

	ctx := stack.GetNodeConfig()
	ftsrv, err := ftservice.New(ctx, ftCfgInstance.FtServiceCfg)
	if err != nil {
		return err
	}

	fp := args[0]
	if len(args) < 3 {
		err = exportBlockChain(ftsrv.BlockChain(), fp)
	} else {
		first, ferr := strconv.ParseInt(args[1], 10, 64)
		last, lerr := strconv.ParseInt(args[2], 10, 64)
		if ferr != nil || lerr != nil {
			return errors.New("Export error in parsing parameters: block number not an integer")
		}
		if first < 0 || last < 0 {
			return errors.New("Export error: block number must be greater than 0")
		}
		err = exportAppendBlockChain(ftsrv.BlockChain(), fp, uint64(first), uint64(last))
	}
	log.Info("Export done in ", "time", time.Since(start))
	return err
}

func exportBlockChain(b *blockchain.BlockChain, fn string) error {
	log.Info("Exporting blockchain", "file", fn)
	// Open the file handle and potentially wrap with a gzip stream
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()

	var writer io.Writer = fh
	if strings.HasSuffix(fn, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}
	// Iterate over the blocks and export them
	if err := b.Export(writer); err != nil {
		return err
	}
	log.Info("Exported blockchain", "file", fn)

	return nil
}

// ExportAppendChain exports a blockchain into the specified file, appending to
// the file if data already exists in it.
func exportAppendBlockChain(b *blockchain.BlockChain, fn string, first uint64, last uint64) error {
	log.Info("Exporting blockchain", "file", fn)
	// Open the file handle and potentially wrap with a gzip stream
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()

	var writer io.Writer = fh
	if strings.HasSuffix(fn, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}
	// Iterate over the blocks and export them
	if err := b.ExportN(writer, first, last); err != nil {
		return err
	}
	log.Info("Exported blockchain to", "file", fn)
	return nil
}
