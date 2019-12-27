package plugin

import (
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
)

type ChainContext interface {
	ChainConfig() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header
}

type Context struct {
	ChainContext

	Coinbase    string // Provides information for COINBASE
	GasLimit    uint64 // Provides information for GASLIMIT
	BlockNumber uint64 // Provides information for NUMBER
	Time        uint64 // Provides information for TIME
	Difficulty  uint64 // Provides information for DIFFICULTY
}

func NewContext(chain ChainContext, header *types.Header) *Context {
	return &Context{
		ChainContext: chain,
		Coinbase:     header.Coinbase,
		GasLimit:     header.GasLimit,
		BlockNumber:  header.Number,
		Time:         header.Time,
		Difficulty:   header.Difficulty,
	}
}
