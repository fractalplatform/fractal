package plugin

import (
	"encoding/json"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

type PluginDoc struct {
	Accounts []*CreateAccountAction
	Assets   []*IssueAssetAction
}

func PluginDocJsonUnMarshal(raw json.RawMessage) (pd *PluginDoc, err error) {
	pd = new(PluginDoc)
	err = json.Unmarshal(raw, pd)
	return
}

// CreateAccount create account
func (pd *PluginDoc) CreateAccount(chainName, accountName string) ([]*types.Transaction, error) {
	var txs []*types.Transaction

	act := &CreateAccountAction{
		Name:   chainName,
		Pubkey: common.HexToPubKey("").String(),
	}

	payload, err := rlp.EncodeToBytes(act)
	if err != nil {
		return nil, err
	}
	env, err := envelope.NewPluginTx(
		CreateAccount,
		chainName,
		accountName,
		0,
		0,
		0,
		0,
		big.NewInt(0),
		big.NewInt(0),
		payload, nil)

	if err != nil {
		return nil, err
	}

	txs = append(txs, types.NewTransaction(env))

	for _, act := range pd.Accounts {
		payload, err := rlp.EncodeToBytes(act)
		if err != nil {
			return nil, err
		}

		env, err := envelope.NewPluginTx(
			CreateAccount,
			chainName,
			accountName,
			0,
			0,
			0,
			0,
			big.NewInt(0),
			big.NewInt(0),
			payload,
			nil)
		if err != nil {
			return nil, err
		}

		txs = append(txs, types.NewTransaction(env))

	}

	return txs, nil
}

// CreateAsset create asset
func (pd *PluginDoc) CreateAsset(chainName, assetName string) ([]*types.Transaction, error) {
	var txs []*types.Transaction

	for _, ast := range pd.Assets {
		payload, err := rlp.EncodeToBytes(ast)
		if err != nil {
			return nil, err
		}

		env, err := envelope.NewPluginTx(
			IssueAsset,
			chainName,
			assetName,
			0,
			0,
			0,
			0,
			big.NewInt(0),
			big.NewInt(0),
			payload,
			nil,
		)
		if err != nil {
			return nil, err
		}

		txs = append(txs, types.NewTransaction(env))
	}

	return txs, nil
}

// RegisterMiner register Miner
func (pd *PluginDoc) RegisterMiner(sysName, dposName string) ([]*types.Transaction, error) {
	env, err := envelope.NewPluginTx(
		RegisterMiner,
		sysName,
		dposName,
		1,             // nonce
		0,             // assetID
		0,             // gasAssetID
		0,             // gasLimit
		big.NewInt(0), // gasprice
		big.NewInt(1), // amount
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return []*types.Transaction{types.NewTransaction(env)}, nil
}

func DefaultPluginDoc() json.RawMessage {
	defaultPD := &PluginDoc{
		Accounts: DefaulAccounts(),
		Assets:   DefaultAssets(),
	}

	raw, err := json.Marshal(defaultPD)
	if err != nil {
		panic(err)
	}
	return raw
}

func DefaulAccounts() []*CreateAccountAction {
	return []*CreateAccountAction{
		&CreateAccountAction{
			Name:   params.DefaultChainconfig.SysName,
			Desc:   "system account",
			Pubkey: "047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd",
		},
		&CreateAccountAction{
			Name:   params.DefaultChainconfig.AccountName,
			Desc:   "account manager account",
			Pubkey: common.HexToPubKey("").String(),
		},
		&CreateAccountAction{
			Name:   params.DefaultChainconfig.AssetName,
			Desc:   "asset manager account",
			Pubkey: common.HexToPubKey("").String(),
		},
		&CreateAccountAction{
			Name:   params.DefaultChainconfig.DposName,
			Desc:   "consensus account",
			Pubkey: common.HexToPubKey("").String(),
		},
		&CreateAccountAction{
			Name:   params.DefaultChainconfig.FeeName,
			Desc:   "fee manager account",
			Pubkey: common.HexToPubKey("").String(),
		},
	}
}

func DefaultAssets() []*IssueAssetAction {
	supply := new(big.Int)
	supply.SetString("10000000000000000000000000000", 10)
	return []*IssueAssetAction{
		&IssueAssetAction{
			AssetName:   params.DefaultChainconfig.SysToken,
			Symbol:      "ft",
			Amount:      supply,
			Decimals:    18,
			Owner:       params.DefaultChainconfig.SysName,
			Founder:     params.DefaultChainconfig.SysName,
			UpperLimit:  supply,
			Contract:    "",
			Description: "",
		},
	}
}
