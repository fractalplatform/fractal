package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/consensus/dpos"
	colorable "github.com/mattn/go-colorable"
	isatty "github.com/mattn/go-isatty"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/sdk"
)

var (
	api      = sdk.NewAPI("http://127.0.0.1:8545")
	chainCfg *params.ChainConfig
)

// TTX
type TTX struct {
	Type    string      `json:"type,omitempty"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
	Gas     uint64      `json:"gas,omitempty"`
	AssetID uint64      `json:"id,omitempty"`
	Value   *big.Int    `json:"value,omitempty"`
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

	usecolor := (isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	ostream := log.StreamHandler(output, log.TerminalFormat(usecolor))
	glogger := log.NewGlogHandler(ostream)
	// logging
	log.PrintOrigins(false)
	glogger.Verbosity(log.LvlDebug)
	log.Root().SetHandler(glogger)
}

func runTx(api *sdk.API, tx *TTX, indent int) error {
	priv, err := crypto.HexToECDSA(tx.Priv)
	if err != nil {
		log.Error(strings.Repeat("*", indent), "hex priv err", err)
		return err
	}
	act := sdk.NewAccount(api, common.StrToName(tx.From), priv, chainCfg.SysTokenID, math.MaxUint64, true, chainCfg.ChainID)

	var hash common.Hash
	switch strings.ToLower(tx.Type) {
	case "createaccount":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &accountmanager.CreateAccountAction{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.CreateAccount(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.CreateAccountAction))
	case "updateaccount":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &accountmanager.UpdataAccountAction{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.UpdateAccount(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.UpdataAccountAction))
	case "updateaccountauthor":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &accountmanager.AccountAuthorAction{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.UpdateAccountAuthor(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.AccountAuthorAction))
	case "transfer":
		hash, err = act.Transfer(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas)
	case "issueasset":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &accountmanager.IssueAsset{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.IssueAsset(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.IssueAsset))
	case "updateasset":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &accountmanager.UpdateAsset{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.UpdateAsset(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.UpdateAsset))
	case "increaseasset":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &accountmanager.IncAsset{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.IncreaseAsset(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.IncAsset))
	case "destroyasset":
		hash, err = act.DestroyAsset(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas)
	case "setassetowner":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &accountmanager.UpdateAssetOwner{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.SetAssetOwner(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*accountmanager.UpdateAssetOwner))
	case "regcandidate":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &dpos.RegisterCandidate{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = arg
		}
		hash, err = act.RegCandidate(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*dpos.RegisterCandidate))
	case "updatecandidate":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &dpos.UpdateCandidate{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = act
		}
		hash, err = act.UpdateCandidate(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*dpos.UpdateCandidate))
	case "unregcandidate":
		hash, err = act.UnRegCandidate(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas)
	case "refundcandidate":
		hash, err = act.RefundCandidate(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas)
	case "votecandidate":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &dpos.VoteCandidate{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = act
		}
		hash, err = act.VoteCandidate(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*dpos.VoteCandidate))
	case "kickedcandidate":
		switch tx.Payload.(type) {
		case map[string]interface{}:
			arg := &dpos.KickedCandidate{}
			bts, _ := json.Marshal(tx.Payload)
			if err := json.Unmarshal(bts, arg); err != nil {
				return err
			}
			tx.Payload = act
		}
		hash, err = act.KickedCandidate(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*dpos.KickedCandidate))
	case "exittakeOver":
		hash, err = act.ExitTakeOver(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas)
	default:
		err = fmt.Errorf("unsupport type %v", tx.Type)
	}
	if bytes.Compare(hash.Bytes(), common.Hash{}.Bytes()) == 0 {
		log.Error(strings.Repeat("*", indent), "txpool err", err, "tx", tx)
		return fmt.Errorf("txpool error %v", err)
	}
	if tx.Succeed != (err == nil) {
		log.Error(strings.Repeat("*", indent), "succeed mismatch", err, "tx", tx)
		return fmt.Errorf("succeed mismatch %v", err)
	}
	if len(tx.Contain) > 0 && !strings.Contains(err.Error(), tx.Contain) {
		log.Error(strings.Repeat("*", indent), "contain mismatch", err, "tx", tx)
		return fmt.Errorf("contain mismatch %v", err)
	}
	log.Info(strings.Repeat("*", indent), "hash", hash.String())
	indent++
	for _, ctx := range tx.Childs {
		return runTx(api, ctx, indent)
	}
	return nil
}

func main() {
	total := 0
	failed := 0

	indent := 0
	txs := []*TTX{}
	for index, tx := range txs {
		log.Info(strings.Repeat("*", indent), "index", index)
		err := runTx(api, tx, indent)

		total++
		if err != nil {
			failed++
		}
	}
	log.Info("result", "total", total, "failed", failed)
}
