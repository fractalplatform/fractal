package sdk

import (
	"crypto/ecdsa"
	"math/rand"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
)

var (
	availableChars = "abcdefghijklmnopqrstuvwxyz0123456789"
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
