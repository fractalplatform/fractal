package vm

import (
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

type PContext struct {
	*EVM
}

func NewPContext(evm *EVM) *PContext {
	return &PContext{EVM: evm}
}

// CurrentHeader retrieves the current header from the local chain.
func (p *PContext) CurrentHeader() *types.Header {
	return p.EVM.CurrentHeader()
}

// GetHeaderByNumber retrieves a block header from the database by number.
func (p *PContext) GetHeaderByNumber(number uint64) *types.Header {
	return p.EVM.GetHeaderByNumber(number)
}

// GetHeaderByHash retrieves a block header from the database by its hash.
func (p *PContext) GetHeaderByHash(hash common.Hash) *types.Header {
	return p.EVM.GetHeaderByHash(hash)
}
