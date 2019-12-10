package plugin

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/fractalplatform/fractal/plugin/abi"
)

func TestABIEncode(t *testing.T) {
	testcase := []interface{}{"hello", big.NewInt(1)}
	for _, v := range testcase {
		b, err := encodeElem(v)
		fmt.Println("err:", err)
		fmt.Println("hex:", len(b))
		for i := 0; i < len(b); i += 32 {
			fmt.Printf("%04d:%x\n", i, b[i:i+32])
		}
	}
}

func TestABIDecode(t *testing.T) {
	testcase := [][]byte{
		[]byte{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x10,
		},
		[]byte{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10,
			'h', 'e', 'l', 'l', 'o', 'w', 'o', 'r', 'l', 'd', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		},
	}
	testobj := []interface{}{
		new(big.Int),
		new(string),
	}
	for i := range testobj {
		err := decodeElem(testcase[i], testobj[i])
		fmt.Println("err:", err)
		fmt.Println("obj:", reflect.ValueOf(testobj[i]).Elem().Interface())
	}
}

var big9 = big.NewInt(90)
var testcase = []interface{}{
	&[4]byte{1},
	&[33]byte{2},
	&[]byte{},
	new(string),
	&big9,
	&[]*big.Int{big.NewInt(1), big.NewInt(2)},
	&[...]*big.Int{big.NewInt(6), big.NewInt(2)},
	&struct {
		Lixp []string
	}{[]string{"hello", "world"}},
}

var checkcase = []interface{}{
	&[4]byte{},
	&[33]byte{},
	&[]byte{},
	new(string),
	&big9,
	&[]*big.Int{},
	&[2]*big.Int{},
	&struct {
		Lixp []string
	}{},
}

func TestABI(t *testing.T) {
	var inputs abi.Arguments
	for _, n := range testcase {
		arguement, err := GoToArgument(n)
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, arguement)
	}
	fmt.Println(inputs)
	b, err := inputs.Pack(testcase...)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%x\n", b)
	err = inputs.Unpack(&checkcase, b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(checkcase)
	for _, n := range checkcase {
		fmt.Println(n)
	}
}

func TestArgument(t *testing.T) {
	var inputs abi.Arguments
	arguement, err := GoToArgument([]byte{})
	inputs = append(inputs, arguement)
	fmt.Println(err)
	fmt.Println(arguement)
	b, err := inputs.Pack([]byte("hello"))
	fmt.Println(err)
	fmt.Println(b)
	var str []byte
	err = inputs.Unpack(&str, b)
	fmt.Println(err)
	fmt.Println(string(str))
}

func MulAPI(a, b *big.Int) *big.Int {
	return new(big.Int).Mul(a, b)
}

func StrAdd(a, b string) string {
	return a + b
}

type GetInfo struct {
	Name   string
	Age    uint64
	Firend []string
}

type Info2 struct {
	Age *big.Int
}

func CheckInfo(info *Info2) bool {
	return true
}

func GetInfoByName(name string) GetInfo {
	return GetInfo{
		Name:   name,
		Age:    21,
		Firend: []string{"lixp", "chengxu"},
	}
}

func ByteShow(str string, b []byte) {
	fmt.Println(str, len(b))
	if len(b)%32 == 4 {
		fmt.Printf("sig: %x\n", b[:4])
		b = b[4:]
	}
	for i := 0; i < len(b); i += 32 {
		fmt.Printf("%04x %x\n", i, b[i:i+32])
	}
}

func CallSimu(method string, args ...interface{}) ([]byte, error) {
	api := apimap[method]
	b, err := api.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	b = append(api.ID(), b...)
	ByteShow("call "+method+" "+api.Sig(), b)
	return SolAPICall(method, b)
}

func TestAPICallInt(t *testing.T) {
	method := "mul"
	err := SolAPIRegister(method, MulAPI)
	fmt.Println("reg err:", err)
	b, err := CallSimu(method, big.NewInt(9), big.NewInt(3))
	fmt.Println("mul err:", err)
	ByteShow("mul byte:", b)
}

func TestAPICallStr(t *testing.T) {
	method := "stradd"
	err := SolAPIRegister(method, StrAdd)
	fmt.Println("reg err:", err)
	b, err := CallSimu(method, "hello", "world")
	fmt.Println("add err:", err)
	ByteShow("add byte:", b)
}

func TestAPICallStruct(t *testing.T) {
	method := "getinfo"
	err := SolAPIRegister(method, GetInfoByName)
	fmt.Println("reg err:", err)
	b, err := CallSimu(method, "xiaoyu")
	fmt.Println("add err:", err)
	ByteShow("add byte:", b)
}

func TestAPICallStructIn(t *testing.T) {
	return
	method := "CheckInfo"
	err := SolAPIRegister(method, CheckInfo)
	fmt.Println("reg err:", err)
	b, err := CallSimu(method, Info2{big.NewInt(4)})
	fmt.Println("add err:", err)
	ByteShow("add byte:", b)
}

type pluginSimu struct {
	Name string
	Age  *big.Int
}

func (p *pluginSimu) Sol_getName(_ interface{}) (string, error) {
	return p.Name, nil
}

func (p *pluginSimu) Sol_setName(_ interface{}, name string) (string, error) {
	p.Name = name
	return p.Name, nil
}
func (p *pluginSimu) Sol_getAge(_ interface{}) (*big.Int, error) {
	return p.Age, nil
}
func (p *pluginSimu) Sol_setAge(_ interface{}, age *big.Int) (*big.Int, error) {
	p.Age = age
	return p.Age, nil
}
func (p *pluginSimu) Sol_set(_ interface{}, name string, age *big.Int) (*pluginSimu, error) {
	return p, nil
}

func PluginCallSimu(o interface{}, name string, args ...interface{}) ([]byte, error) {
	typ := reflect.TypeOf(o)
	apis := pluginSolAPI[typ]
	for k, v := range apis {
		if v.Name == name {
			b, err := v.Inputs.Pack(args...)
			if err != nil {
				return b, err
			}
			b = append(k[:], b...)
			ByteShow("call "+name+" "+v.Sig(), b)
			return PluginSolAPICall(o, struct{}{}, b)
		}
	}
	return nil, errors.New("method not exist")
}

func TestPluginAPICallInt(t *testing.T) {
	if err := PluginSolAPIRegister(&pluginSimu{}); err != nil {
		t.Fatal(err)
	}
	var a interface{} = &pluginSimu{"lixiaopeng", big.NewInt(27)}
	b, err := PluginCallSimu(a, "setName", "liuxiaoyu")
	if err != nil {
		t.Fatal(err)
	}
	ByteShow("ret", b)
}

func TestPluginAPICallStruct(t *testing.T) {
	if err := PluginSolAPIRegister(&pluginSimu{}); err != nil {
		t.Fatal(err)
	}
	b, err := PluginCallSimu(&pluginSimu{"lixiaopeng", big.NewInt(27)}, "set", "liuxiaoyu", big.NewInt(40))
	if err != nil {
		t.Fatal(err)
	}
	ByteShow("ret", b)
}
