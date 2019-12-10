package plugin

import (
	"math/big"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/magiconair/properties/assert"
)

func TestSign(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")

	signer, _ := NewSigner(big.NewInt(1))

	h := func(chainID *big.Int) common.Hash {
		return common.RlpHash("sign test")
	}
	d, err := signer.Sign(h, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	pub, err := signer.Recover(d, h)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, common.Bytes2Hex(pub), "047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd")

}
