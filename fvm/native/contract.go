package native

import (
	"github.com/fractalplatform/fractal/fvm/native/vm"
	"github.com/fractalplatform/fractal/types"
)

type Contract struct{}

func (c *Contract) Run(method string, params ...interface{}) ([]byte, error) {
	switch method {
	case "Call":
		evm := params[0].(*vm.EVM)
		action := params[1].(*types.Action)
		gas := params[2].(uint64)
		return c.Call(evm, action, gas)
	case "Create":
		evm := params[0].(*vm.EVM)
		action := params[1].(*types.Action)
		gas := params[2].(uint64)
		return c.Create(evm, action, gas)
	default:
		return nil, errFuncNotExist
	}
}

func (c *Contract) Call(evm *vm.EVM, action *types.Action, gas uint64) (ret []byte, err error) {
	b, _, err := evm.Call(vm.AccountRef(action.Sender()), action, gas)
	return b, err
}

func (c *Contract) Create(evm *vm.EVM, action *types.Action, gas uint64) (ret []byte, err error) {
	b, _, err := evm.Create(vm.AccountRef(action.Sender()), action, gas)
	return b, err
}
