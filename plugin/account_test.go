package plugin

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
)

var sdb *state.StateDB
var pm IPM
var pubkey string

func getStateDB() *state.StateDB {
	db := rawdb.NewMemoryDatabase()
	trieDB := state.NewDatabase(db)
	stateDB, err := state.New(common.Hash{}, trieDB)
	if err != nil {
		fmt.Printf("test getStateDB() failure %v", err)
		return nil
	}
	return stateDB
}

func testInit() {
	sdb = getStateDB()
	pm = NewPM(sdb)
	pubkey = "047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"
	pm.CreateAccount("assetowner", pubkey, "just for test")
	pm.CreateAccount("assetfounder", pubkey, "just for test")
	pm.IssueAsset("assetowner", "ftoken", "ft", big.NewInt(1000), 10, "assetfounder", "assetowner", big.NewInt(10000), "issue for test", pm)
}

func Test_CreateAccount(t *testing.T) {
	invalidKey := "0x1234568"
	invalidKey2 := "0x047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd1111111"

	type args struct {
		name string
		key  string
		des  string
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"createAccount", args{"abc1231414", pubkey, "just for test"}, nil},
		{"createAccount", args{"abc111111", pubkey, "just for test"}, nil},
		{"createAccount", args{"testnonce", pubkey, "just for test"}, nil},
		{"createAccount", args{"testcode", pubkey, "just for test"}, nil},
		{"createAccount with invalid name", args{"a", pubkey, "just for test"}, ErrAccountNameInvalid},
		{"createAccount with same name", args{"abc1231414", pubkey, "just for test"}, ErrAccountIsExist},
		{"createAccount with invalid key1", args{"abc1231415", invalidKey, "just for test"}, ErrPubKey},
		{"createAccount with invalid key2", args{"abc1231416", invalidKey2, "just for test"}, ErrPubKey},
	}

	for _, item := range testItem {
		if _, err := pm.CreateAccount(item.arg.name, item.arg.key, item.arg.des); err != item.err {
			t.Errorf("%q. CreateAccount() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_SetNonce(t *testing.T) {
	type args struct {
		name  string
		nonce uint64
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"SetNonce account not exist ", args{"a", 10}, ErrAccountNotExist},
		{"SetNonce", args{"testnonce", 10}, nil},
	}

	for _, item := range testItem {
		if err := pm.SetNonce(item.arg.name, item.arg.nonce); err != item.err {
			t.Errorf("%q. SetNonce() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_GetNonce(t *testing.T) {
	type args struct {
		name string
	}
	testItem := []struct {
		testDes string
		arg     args
		nonce   uint64
		err     error
	}{
		{"GetNonce account not exist", args{"a"}, 0, ErrAccountNotExist},
		{"GetNonce", args{"testnonce"}, 10, nil},
	}

	for _, item := range testItem {
		get, err := pm.GetNonce(item.arg.name)
		if err != item.err {
			t.Errorf("%q. GetNonce() error = %v, wantErr %v", item.testDes, err, item.err)
		}

		if get != item.nonce {
			t.Errorf("%q. GetNonce() nonce = %d, wantNonce %d", get, err, item.nonce)
		}
	}
}

func Test_SetCode(t *testing.T) {
	bCode := []byte("contract code")

	type args struct {
		name string
		code []byte
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"SetCode", args{"testcode", bCode}, nil},
	}

	for _, item := range testItem {
		if err := pm.SetCode(item.arg.name, item.arg.code); err != item.err {
			t.Errorf("%q. SetNonce() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_GetCode(t *testing.T) {
	type args struct {
		name string
	}
	testItem := []struct {
		testDes  string
		arg      args
		wantCode []byte
		err      error
	}{
		//{"GetCode code is empty", args{"abc111111"}, nil, ErrCodeIsEmpty},
		{"GetCode", args{"testcode"}, []byte("contract code"), nil},
	}

	for _, item := range testItem {
		get, err := pm.GetCode(item.arg.name)
		if err != item.err {
			t.Errorf("%q. GetCode() error = %v, wantErr = %v", item.testDes, err, item.err)
		}
		if err == nil {
			if !bytes.Equal(get, item.wantCode) {
				t.Errorf("%q. GetCode() code no equal", item.testDes)
			}
		}
	}
}

func Test_GetCodeHash(t *testing.T) {
	type args struct {
		name string
	}
	testItem := []struct {
		testDes  string
		arg      args
		wantHash common.Hash
		err      error
	}{
		{"GetCodeHash hash is empty", args{"abc111111"}, common.Hash{}, ErrHashIsEmpty},
		{"GetCodeHash", args{"testcode"}, crypto.Keccak256Hash([]byte("contract code")), nil},
	}

	for _, item := range testItem {
		get, err := pm.GetCodeHash(item.arg.name)
		if err != item.err {
			t.Errorf("%q. GetCodeHash() error = %v, wantErr %v", item.testDes, err, item.err)
		}
		if err == nil {
			if !bytes.Equal(get.Bytes(), item.wantHash.Bytes()) {
				t.Errorf("%q. GetCodeHash() hash no equal", item.testDes)
			}
		}
	}
}

func Test_GetBalance(t *testing.T) {
	type args struct {
		name string
		id   uint64
	}
	testItem := []struct {
		testDes     string
		arg         args
		wantBalance *big.Int
		err         error
	}{
		{"GetBalance", args{"assetowner", params.SysTokenID()}, big.NewInt(1000), nil},
	}

	for _, item := range testItem {
		get, err := pm.GetBalance(item.arg.name, item.arg.id)
		if err != item.err {
			t.Errorf("%q. GetBalance() error = %v, wantErr %v", item.testDes, err, item.err)
		}
		if get.Cmp(item.wantBalance) != 0 {
			t.Errorf("%q. GetBalance() balance no equal", item.testDes)
		}
	}
}

func Test_CanTransfer(t *testing.T) {
	type args struct {
		name  string
		id    uint64
		value *big.Int
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"CanTransfer", args{"assetowner", params.SysTokenID(), big.NewInt(1000)}, nil},
		{"CanTransfer", args{"assetowner", params.SysTokenID(), big.NewInt(10000)}, ErrInsufficientBalance},
	}

	for _, item := range testItem {
		err := pm.CanTransfer(item.arg.name, item.arg.id, item.arg.value)
		if err != item.err {
			t.Errorf("%q. CanTransfer() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_TransferAsset(t *testing.T) {
	type args struct {
		from  string
		to    string
		id    uint64
		value *big.Int
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"TransferAsset", args{"assetowner", "assetfounder", params.SysTokenID(), big.NewInt(1000)}, nil},
	}

	for _, item := range testItem {
		err := pm.TransferAsset(item.arg.from, item.arg.to, item.arg.id, item.arg.value)
		if err != item.err {
			t.Errorf("%q. TransferAsset() error = %v, wantErr %v", item.testDes, err, item.err)
		}
		if err == nil {
			fromGet, _ := pm.GetBalance(item.arg.from, item.arg.id)
			if fromGet.Cmp(big.NewInt(0)) != 0 {
				t.Errorf("%q. TransferAsset() toAccount balance no equal", item.testDes)
			}
			toGet, _ := pm.GetBalance(item.arg.to, item.arg.id)
			if toGet.Cmp(item.arg.value) != 0 {
				t.Errorf("%q. TransferAsset() toAccount balance no equal", item.testDes)
			}
		}
	}
}

func TestMain(m *testing.M) {
	testInit()
	m.Run()
}
