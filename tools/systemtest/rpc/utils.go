package rpc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	once              sync.Once
	clientInstance    *rpc.Client
	clientInstanceMap map[string]*rpc.Client
	hostIp            = "192.168.2.13" // "127.0.0.1" //"121.69.28.78"//
	port              = 10090          // 8545 //
	gasPrice          = big.NewInt(3000000)
	globalNonce       = make(map[string]uint64)
	lock              sync.Mutex
	SentTxNum         = int32(0)
)

func init() {
	globalNonce = make(map[string]uint64)
}

func GenerateStatInfo() {
	atomic.AddInt32(&SentTxNum, 1)
}

type GenAction struct {
	*types.Action
	PrivateKey *ecdsa.PrivateKey
}

// DefultURL default rpc url
func DefultURL() string {
	return fmt.Sprintf("http://%s:%d", hostIp, port)
}

func GeneratePubKey() (common.PubKey, *ecdsa.PrivateKey) {
	prikey, _ := crypto.GenerateKey()
	return common.BytesToPubKey(crypto.FromECDSAPub(&prikey.PublicKey)), prikey
}

func NewGeAction(at types.ActionType, from, to common.Name, nonce uint64, assetid uint64, gaslimit uint64, amount *big.Int, payload []byte, prikey *ecdsa.PrivateKey) *GenAction {
	action := types.NewAction(at, from, to, nonce, assetid, gaslimit, amount, payload)
	return &GenAction{
		Action:     action,
		PrivateKey: prikey,
	}
}
func SendTxTest(gcs []*GenAction) (common.Hash, error) {
	//nonce := GetNonce(sendaddr, "latest")
	signer := types.NewSigner(params.DefaultChainconfig.ChainID)
	var actions []*types.Action
	for _, v := range gcs {
		actions = append(actions, v.Action)
	}
	tx := types.NewTransaction(uint64(1), gasPrice, actions...)
	for _, v := range gcs {
		err := types.SignAction(v.Action, tx, signer, v.PrivateKey)
		if err != nil {
			return common.Hash{}, err
		}
	}
	rawtx, _ := rlp.EncodeToBytes(tx)
	hash, err := SendRawTx(rawtx)
	GenerateStatInfo()
	return hash, err
}

//SendRawTx send raw transaction
func SendRawTx(rawTx []byte) (common.Hash, error) {
	hash := new(common.Hash)
	err := ClientCall("ft_sendRawTransaction", hash, hexutil.Bytes(rawTx))
	return *hash, err
}

// MustRPCClient Wraper rpc's client
func MustRPCClient() (*rpc.Client, error) {
	once.Do(func() {
		client, err := rpc.DialHTTP(DefultURL())
		if err != nil {
			return
		}
		clientInstance = client
	})

	return clientInstance, nil
}

func MustRPCClientWithAddr(nodeIp string, nodePort int64) (*rpc.Client, error) {
	endPoint := fmt.Sprintf("http://%s:%d", nodeIp, nodePort)
	if clientInstanceMap[endPoint] != nil {
		return clientInstanceMap[endPoint], nil
	}
	client, err := rpc.DialHTTP(endPoint)
	if err != nil {
		return nil, err
	}
	clientInstanceMap[endPoint] = client
	return client, nil
}

// ClientCall Wrapper rpc call api.
func ClientCall(method string, result interface{}, args ...interface{}) error {
	client, err := MustRPCClient()
	if err != nil {
		return err
	}
	err = client.CallContext(context.Background(), result, method, args...)
	if err != nil {
		return err
	}
	return nil
}

// ClientCall Wrapper rpc call api.
func ClientCallWithAddr(nodeIp string, nodePort int64, method string, result interface{}, args ...interface{}) error {
	client, err := MustRPCClientWithAddr(nodeIp, nodePort)
	if err != nil {
		return err
	}
	err = client.CallContext(context.Background(), result, method, args...)
	if err != nil {
		return err
	}
	return nil
}

// GasPrice suggest gas price
func GasPrice() (*big.Int, error) {
	gp := big.NewInt(0)
	err := ClientCall("ft_gasPrice", gp)
	return gp, err
}

// GetNonce get nonce by address and block number.
func GetNonce(accountName common.Name) (uint64, error) {
	lock.Lock()
	defer lock.Unlock()
	nonce, ok := globalNonce[accountName.String()]
	if !ok || nonce == 0 {
		err := ClientCall("account_getNonce", &nonce, accountName.String())
		if err != nil {
			return nonce, err
		}
		globalNonce[accountName.String()] = nonce
	} else {
		globalNonce[accountName.String()] = globalNonce[accountName.String()] + 1
	}
	//fmt.Printf("%s's nonce = %d\n", accountName, globalNonce[accountName.String()])
	return globalNonce[accountName.String()], nil
}

// defaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := HomeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "pi_ledger")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "pi_ledger")
		} else {
			return filepath.Join(home, ".pi_ledger")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
func GenerateRandomName(namePrefix string, addStrLen int) string {
	newRandomName := namePrefix
	var str string = "abcdefghijklmnopqrstuvwxyz0123456789"
	size := len(str)
	now := int64(time.Now().Nanosecond()) + int64(GetGID())
	rand.Seed(now)
	for i := 0; i < addStrLen; i++ {
		index := rand.Intn(10000) % size
		newRandomName += string(str[index])
	}
	return newRandomName
}

func checkReceipt(fromAccount string, txHash common.Hash, maxTime uint) error {
	receipt, outOfTime, err := DelayGetReceiptByTxHash(txHash, maxTime)
	if err != nil {
		return errors.New("获取交易receipt失败：" + err.Error())
	}
	if outOfTime {
		bInPending, bInQueued, _ := CheckTxInPendingOrQueue(fromAccount, txHash)
		if !bInPending && !bInQueued {
			return errors.New("获取receipt超时，并且无法从pending和queued队列中查到txhash")
		}
		if bInPending && bInQueued {
			return errors.New("获取receipt超时，交易位于pending队列中和queued队列中")
		}
		if bInPending {
			return errors.New("获取receipt超时，交易位于pending队列中")
		}
		if bInQueued {
			return errors.New("获取receipt超时，交易位于queued队列中")
		}
	}
	if len(receipt.ActionResults) == 0 {
		return errors.New("交易执行结果为失败")
	}
	if receipt.ActionResults[0].Status == 0 {
		return errors.New("交易执行结果为失败：" + receipt.ActionResults[0].Error)
	}
	return nil
}

func runTxAndCheckReceipt(txFunc func() (common.Hash, error), fromAccount string) error {
	txHash, err := txFunc()
	if err != nil {
		return err
	}
	err = checkReceipt(fromAccount, txHash, 60)
	if err != nil {
		return err
	}
	return nil
}
