package plugin

import (
	"errors"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/types"
)

var (
	errFuncNotExist      = errors.New("function not exist")
	errFuncParamNumWrong = errors.New("function param num wrong")
)

type operation struct {
	execute func(context *Context) ([]byte, uint64, error)
}

var methodSet = map[string]*operation{
	"account_CreateAccount": &operation{
		execute: CreateAccount,
	},
}

type Context struct {
	Account *accountmanager.AccountManager
	Action  *types.Action
	Evm     *vm.EVM
	Gas     uint64
	Params  []interface{}
}
