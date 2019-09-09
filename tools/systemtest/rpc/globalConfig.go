package rpc

import (
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
)

var (
	SystemAccountPriKey, _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	SystemAccountPubKey    = common.HexToPubKey("0x047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd")
	Gaslimit               = uint64(2000000)
	SystemAccount          = "ftsystemio"
	Minernonce             = uint64(0)
	MaxTxNumInTxpool       = 40960 + 4096
	StakeWeight            = int64(1000)
)
