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
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// insertStats tracks and reports on block insertion.
type insertStats struct {
	queued, processed, ignored, txsCnt int
	usedGas                            uint64
	startTime                          time.Time
}

// statsReportLimit is the time limit during import and export after which we
// always print out progress. This avoids the user wondering what's going on.
const statsReportLimit = 8 * time.Second

// report prints statistics if some number of blocks have been processed
// or more than a few seconds have passed since the last message.
func (st *insertStats) report(chain []*types.Block, index int) {
	// Fetch the timings for the batch
	var (
		now     = time.Now()
		elapsed = now.Sub(st.startTime)
	)
	// If we're at the last block of the batch or report period reached, log
	if index == len(chain)-1 || elapsed >= statsReportLimit {
		end := chain[index]

		// Assemble the log context and send it to the logger
		context := []interface{}{
			"blocks", st.processed, "txs", st.txsCnt, "gas", float64(st.usedGas),
			"elapsed", common.PrettyDuration(elapsed), "number", end.Number(), "hash", end.Hash(),
		}

		if st.queued > 0 {
			context = append(context, []interface{}{"queued", st.queued}...)
		}
		if st.ignored > 0 {
			context = append(context, []interface{}{"ignored", st.ignored}...)
		}
		log.Info("Imported new chain segment", context...)

		// Bump the stats reported to the next section
		*st = insertStats{startTime: now}
	}
}
