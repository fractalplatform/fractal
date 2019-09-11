package plugin

import (
	"errors"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/types"
)

var (
	errFuncNotExist      = errors.New("function not exist")
	errFuncParamNumWrong = errors.New("function param num wrong")
)

type operation struct {
	execute func(context *Context) ([]byte, error)
}

var methodSet = map[string]*operation{
	"NativeAccount_CreateAccount": &operation{
		execute: CreateAccount,
	},
}

type Context struct {
	account *accountmanager.AccountManager
	action  *types.Action
	params  []interface{}
}
