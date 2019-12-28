package plugin

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/abi"
)

// [p1, p2, p3][p1, p2]
// 1. 注册plugin函数
// 2.

// uintX/intX
// bytesX
// bytes
// string
// struct/truple
// []xx
// mapping x

// int 8 16 32 64 256      =
// uint  8 16 32 64 256    =
// bool bool               =
// address => array
// string string		   =
// bytes 1-32
// bytes

var errorType = reflect.TypeOf((*error)(nil)).Elem()

var mapstr = map[reflect.Type]string{
	reflect.TypeOf("string"):         "string",
	reflect.TypeOf(big.Int{}):        "uint256",
	reflect.TypeOf(common.Address{}): "address",
	reflect.TypeOf(true):             "bool",
	reflect.TypeOf([]byte{}):         "bytes",
	reflect.TypeOf(int(0)):           "int64",
	reflect.TypeOf(uint(0)):          "uint64",
	reflect.TypeOf(int8(0)):          "int8",
	reflect.TypeOf(uint8(0)):         "uint8",
	reflect.TypeOf(int16(0)):         "int16",
	reflect.TypeOf(uint16(0)):        "uint16",
	reflect.TypeOf(int32(0)):         "int32",
	reflect.TypeOf(uint32(0)):        "uint32",
	reflect.TypeOf(int64(0)):         "int64",
	reflect.TypeOf(uint64(0)):        "uint64",
	//reflect.TypeOf(struct{}{}): "tuple",
}

func getElem(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func getTypeStr(typ reflect.Type) (string, string) {
	if str, exist := mapstr[typ]; exist {
		return str, str
	}
	if typ.Kind() == reflect.Struct {
		return "tuple", "tuple"
	}
	if typ.Kind() == reflect.Array {
		elem := getElem(typ.Elem())
		base, truple := getTypeStr(elem)
		if elem.Kind() == reflect.Uint8 && typ.Len() <= 32 {
			return base, fmt.Sprintf("bytes%d", typ.Len())
		}
		return base, fmt.Sprintf("%s[%d]", truple, typ.Len())
	}
	if typ.Kind() == reflect.Slice {
		elem := getElem(typ.Elem())
		base, truple := getTypeStr(elem)
		return base, truple + "[]"
	}
	return "", ""
}

func goToComponents(structTyp reflect.Type) []abi.ArgumentMarshaling {
	if structTyp.Kind() == reflect.Slice || structTyp.Kind() == reflect.Array {
		elem := getElem(structTyp.Elem())
		return goToComponents(elem)
	}
	components := make([]abi.ArgumentMarshaling, 0, structTyp.NumField())
	for i := 0; i < structTyp.NumField(); i++ {
		// TODO: skip function
		field := structTyp.Field(i)
		typ := getElem(field.Type)
		base, truple := getTypeStr(typ)
		component := abi.ArgumentMarshaling{
			Name: field.Name,
			Type: truple,
		}
		if base == "tuple" {
			component.Components = goToComponents(typ)
		}
		components = append(components, component)
	}
	return components
}

func GoToArgument(in interface{}) (abi.Argument, error) {
	typ := reflect.TypeOf(in)
	return goToArgument(typ)
}

func goToArgument(typ reflect.Type) (abi.Argument, error) {
	typ = getElem(typ)

	var components []abi.ArgumentMarshaling
	base, typstr := getTypeStr(typ)

	if base == "tuple" {
		components = goToComponents(typ)
	}
	abityp, err := abi.NewType(typstr, components)
	return abi.Argument{
		Name: "_" + typ.String(),
		Type: abityp,
	}, err
}

type pluginMethod struct {
	abi.Method
	method reflect.Value
}
type pluginMethodes map[[4]byte]*pluginMethod

var pluginSolAPI = make(map[reflect.Type]pluginMethodes)

func isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}

func solParamsCheck(foo reflect.Method) error {
	funcType := foo.Func.Type()
	if funcType.NumIn() < 2 {
		return errors.New("SolAPI need at least two parameters")
	}
	if funcType.NumOut() < 1 {
		return errors.New("SolAPI need at least one return")
	}
	errRet := funcType.Out(funcType.NumOut() - 1)
	if !isErrorType(errRet) {
		return errors.New("SolAPI need at least one return")
	}
	return nil
}

func isSolMethod(foo reflect.Method) bool {
	return strings.HasPrefix(foo.Name, "Sol_")
}

func goMethodABI(foo reflect.Method) (*pluginMethod, error) {
	funcType := foo.Func.Type()
	input := make(abi.Arguments, 0, funcType.NumIn()-2)
	output := make(abi.Arguments, 0, funcType.NumOut()-1)
	for i := 2; i < funcType.NumIn(); i++ {
		in, err := goToArgument(funcType.In(i))
		if err != nil {
			return nil, err
		}
		input = append(input, in)
	}
	for i := 0; i < funcType.NumOut()-1; i++ {
		out, err := goToArgument(funcType.Out(i))
		if err != nil {
			return nil, err
		}
		output = append(output, out)
	}
	return &pluginMethod{
		abi.Method{
			Name:    foo.Name[4:],
			RawName: foo.Name[4:],
			Inputs:  input,
			Outputs: output,
		},
		foo.Func,
	}, nil
}

func PluginSolAPIRegister(o interface{}) error {
	typ := reflect.TypeOf(o)
	method := make(pluginMethodes)
	for i := 0; i < typ.NumMethod(); i++ {
		foo := typ.Method(i)
		if !isSolMethod(foo) {
			continue
		}
		if err := solParamsCheck(foo); err != nil {
			return err
		}
		mabi, err := goMethodABI(foo)
		if err != nil {
			return err
		}
		var key [4]byte
		copy(key[:], mabi.ID())
		method[key] = mabi
	}
	pluginSolAPI[typ] = method
	return nil
}

func PluginSolAPICall(o, p1 interface{}, data []byte) ([]byte, error) {
	mplugin, exist := pluginSolAPI[reflect.TypeOf(o)]
	if !exist {
		return nil, errors.New("plugin not exist")
	}
	if len(data) < 4 {
		return nil, errors.New("calldata must larger than 4 bytes")
	}
	var sigID [4]byte
	copy(sigID[:], data[:4])
	method, exist := mplugin[sigID]
	if !exist {
		return nil, errors.New("method is not exist")
	}
	end := len(data[4:]) / 32
	params, err := method.Inputs.UnpackValues(data[4 : end*32+4])
	if err != nil {
		return nil, err
	}
	callparams := make([]reflect.Value, len(params)+2)
	callparams[0] = reflect.ValueOf(o)
	callparams[1] = reflect.ValueOf(p1)
	for i, p := range params {
		callparams[i+2] = reflect.ValueOf(p)
	}
	out := method.method.Call(callparams)
	if callerr := out[len(out)-1].Interface(); callerr != nil {
		return nil, callerr.(error)
	}
	outInter := make([]interface{}, len(out)-1)
	for i, o := range out[:len(out)-1] {
		outInter[i] = o.Interface()
	}
	retbytes, err := method.Outputs.Pack(outInter...)
	if err != nil {
		return nil, err
	}
	if out[len(out)-1].IsNil() {
		return retbytes, nil
	}
	return retbytes, nil
}

func getPluginABI(pluginObj interface{}) *abi.ABI {
	typ := reflect.TypeOf(pluginObj)
	mplugin := pluginSolAPI[typ]
	ret := &abi.ABI{
		Methods: make(map[string]abi.Method, len(mplugin)),
	}
	for _, method := range mplugin {
		ret.Methods[method.Name] = method.Method
	}
	return ret
}
