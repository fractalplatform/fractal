package native

import (
	"errors"

	"github.com/fractalplatform/fractal/common"
)

var (
	errFuncNotExist = errors.New("function not exist")
)

type NativeContract interface {
	Run(method string, params ...interface{}) ([]byte, error) // Run runs the precompiled contract
}

// PrecompiledContracts contains the default set of pre-compiled
var NativeContracts = map[common.Name]NativeContract{
	common.Name("native.asset"):    &NativeAsset{},
	common.Name("native.contract"): &Contract{},
}
