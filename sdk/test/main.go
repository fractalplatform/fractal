package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path"
	"path/filepath"
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
	api      *sdk.API
	chainCfg *params.ChainConfig
)

// TTX
type TTX struct {
	Comment string      `json:"comment,omitempty"`
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

func Init() {
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

func runTx(api *sdk.API, tx *TTX, file string) error {
	priv, err := crypto.HexToECDSA(tx.Priv)
	if err != nil {
		log.Error(file, "hex priv err", err)
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
			tx.Payload = arg
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
			tx.Payload = arg
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
			tx.Payload = arg
		}
		hash, err = act.KickedCandidate(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas, tx.Payload.(*dpos.KickedCandidate))
	case "exittakeover":
		hash, err = act.ExitTakeOver(common.StrToName(tx.To), tx.Value, tx.AssetID, tx.Gas)
	default:
		err = fmt.Errorf("unsupport type %v", tx.Type)
	}
	if tx.Succeed != (err == nil) {
		log.Error(file, "hash", hash.String(), "succeed mismatch", err, "comment", tx.Comment)
		return fmt.Errorf("succeed mismatch %v", err)
	}
	if len(tx.Contain) > 0 && !strings.Contains(err.Error(), tx.Contain) {
		log.Error(file, "hash", hash.String(), "contain mismatch", err, "except", tx.Contain, "comment", tx.Comment)
		return fmt.Errorf("contain mismatch %v", err)
	}
	if tx.Succeed && bytes.Compare(hash.Bytes(), common.Hash{}.Bytes()) == 0 {
		log.Error(file, "hash", hash.String(), "txpool err", err, "comment", tx.Comment)
		return fmt.Errorf("txpool error %v", err)
	}
	log.Info(file, "hash", hash.String(), "comment", tx.Comment)
	for _, ctx := range tx.Childs {
		if err := runTx(api, ctx, file); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	_rpchost := flag.String("u", "http://127.0.0.1:8545", "RPC地址")
	_dirfile := flag.String("d", "./testcase", "目录名/文件名")
	flag.Parse()
	api = sdk.NewAPI(*_rpchost)
	Init()
	f, _ := os.Stat(*_dirfile)
	if f.IsDir() {
		rd, err := ioutil.ReadDir(*_dirfile)
		if err != nil {
			panic(err)
		}
		run(*_dirfile, rd)
	} else {
		rd := []os.FileInfo{f}
		run(filepath.Dir(*_dirfile), rd)
	}
}

func run(dir string, rd []os.FileInfo) {
	for _, fi := range rd {
		if !fi.IsDir() {
			bts, err := ioutil.ReadFile(path.Join(dir, fi.Name()))
			if err != nil {
				panic(err)
			}
			txs := []*TTX{}
			d := json.NewDecoder(bytes.NewReader(bts))
			d.UseNumber()
			if err := d.Decode(&txs); err != nil {
				panic(err)
			}

			for _, tx := range txs {
				runTx(api, tx, path.Join(dir, fi.Name()))
			}
		} else {
			crd, err := ioutil.ReadDir(path.Join(dir, fi.Name()))
			if err != nil {
				panic(err)
			}
			run(path.Join(dir, fi.Name()), crd)
		}
	}
}
