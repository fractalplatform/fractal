package main

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"

	"github.com/fractalplatform/fractal/sdk"
)

func TestTx(t *testing.T) {
	priv1 := sdk.GenerateKey()
	priv2 := sdk.GenerateKey()
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "createaccount",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   100,
		Payload: &accountmanager.CreateAccountAction{
			AccountName: "sdktest005",
			Founder:     "sdktest005",
			PublicKey:   common.BytesToPubKey(crypto.FromECDSAPub(&priv1.PublicKey)),
			Description: "descr sdktest001",
		},
		Succeed: true,
		Childs: []*TTX{
			&TTX{
				Priv:    hex.EncodeToString(crypto.FromECDSA(priv1)),
				Type:    "createaccount",
				From:    "sdktest005",
				To:      chainCfg.AccountName,
				Gas:     1000000,
				AssetID: chainCfg.SysTokenID,
				Value:   10,
				Payload: &accountmanager.CreateAccountAction{
					AccountName: "sdktest002",
					Founder:     "sdktest002",
					PublicKey:   common.BytesToPubKey(crypto.FromECDSAPub(&priv2.PublicKey)),
					Description: "descr sdktest002",
				},
				Succeed: true,
			},
		},
	}

	cjson, _ := json.Marshal(tx)
	ttx := &TTX{}
	json.Unmarshal(cjson, ttx)
	if err := runTx(api, ttx, 0); err != nil {
		panic(err)
	}
}
