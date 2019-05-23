package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/fractalplatform/fractal/accountmanager"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/sdk"
)

var (
	api      = sdk.NewAPI("http://127.0.0.1:8545")
	decimals = big.NewInt(1)
	chainCfg *params.ChainConfig
)

// TTX
type TTX struct {
	Type    string      `json:"type,omitempty"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
	Gas     uint64      `json:"gas,omitempty"`
	AssetID uint64      `json:"id,omitempty"`
	Value   uint64      `json:"value,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
	Succeed bool        `json:"succeed,omitempty"`
	Contain string      `json:"contain,omitempty"`
	Priv    string      `json:"priv,omitempty"`
	Childs  []*TTX      `json:"childs,omitempty"`
}

func init() {
	// init chain config & decimals
	cfg, err := api.GetChainConfig()
	if err != nil {
		panic(fmt.Sprintf("init err %v", err))
	}
	chainCfg = cfg
	for i := uint64(0); i < chainCfg.SysTokenDecimals; i++ {
		decimals = new(big.Int).Mul(decimals, big.NewInt(10))
	}
}

func runTx(api *sdk.API, tx *TTX) error {
	priv, err := crypto.HexToECDSA(tx.Priv)
	if err != nil {
		return err
	}
	fmt.Println(tx.Priv)
	act := sdk.NewAccount(api, common.StrToName(tx.From), priv, chainCfg.SysTokenID, math.MaxInt64, true, chainCfg.ChainID)
	var hash common.Hash
	var err1 error
	switch strings.ToLower(tx.Type) {
	case "createaccount":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			act := &accountmanager.CreateAccountAction{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, act); err != nil {
				return err
			}
			tx.Payload = act
		}
		hash, err1 = act.CreateAccount(common.StrToName(tx.To), new(big.Int).Mul(big.NewInt(int64(tx.Value)), decimals), tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.CreateAccountAction))
	case "updateaccount":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			act := &accountmanager.UpdataAccountAction{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, act); err != nil {
				return err
			}
			tx.Payload = act
		}
		hash, err1 = act.UpdateAccount(common.StrToName(tx.To), new(big.Int).Mul(big.NewInt(int64(tx.Value)), decimals), tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.UpdataAccountAction))
	case "UpdateAccountAuthor":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			act := &accountmanager.AccountAuthorAction{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, act); err != nil {
				return err
			}
			tx.Payload = act
		}
		hash, err1 = act.UpdateAccountAuthor(common.StrToName(tx.To), new(big.Int).Mul(big.NewInt(int64(tx.Value)), decimals), tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.AccountAuthorAction))
	default:
	}
	if tx.Succeed == (err1 == nil) {
		return fmt.Errorf("mismatch %v - %v", hash.String(), tx)
	}
	if len(tx.Contain) > 0 && !strings.Contains(err1.Error(), tx.Contain) {
		return fmt.Errorf("mismatch %v - %v", hash.String(), tx)
	}
	for _, ctx := range tx.Childs {
		return runTx(api, ctx)
	}
	return nil
}

func main() {
	total := 0
	failed := 0
	txs := []*TTX{}
	for index, tx := range txs {
		total++
		err := runTx(api, tx)
		if err != nil {
			failed++
		}
		fmt.Println(fmt.Sprintf("%5d %v", index, err))
	}
	fmt.Println("total", total, "failed", failed)
}
