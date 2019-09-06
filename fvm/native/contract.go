package native

import (
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/types"
)

type Contract struct {
}

func (c *Contract) Run(method string, params ...interface{}) ([]byte, error) {
	return nil, nil
}

func (c *Contract) Call(evm *vm.EVM, action *types.Action, gas uint64) {

}
