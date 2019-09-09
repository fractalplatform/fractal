package common

import (
	"crypto/ecdsa"
	"math/rand"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
)

var (
	availableChars  = "abcdefghijklmnopqrstuvwxyz0123456789"
	rpchost         = "http://127.0.0.1:8545"
	systemaccount   = "ftsystemio"
	systemprivkey   = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
	systemassetname = "ftfoundation"
	systemassetid   = uint64(1)
)

// GenerateAccountName generate account name
func GenerateAccountName(namePrefix string, addStrLen int) string {
	newRandomName := namePrefix
	size := len(availableChars)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < addStrLen; i++ {
		index := rand.Intn(10000) % size
		newRandomName += string(availableChars[index])
	}
	return newRandomName
}

// GenerateKey generate pubkey and privkey
func GenerateKey() (*ecdsa.PrivateKey, common.PubKey) {
	prikey, _ := crypto.GenerateKey()
	return prikey, common.BytesToPubKey(crypto.FromECDSAPub(&prikey.PublicKey))
}
