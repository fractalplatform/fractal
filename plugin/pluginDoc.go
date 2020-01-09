package plugin

import (
	"encoding/json"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/abi"
)

type PluginDoc struct {
	Accounts []*CreateAccountAction `json:"accounts"`
	Assets   []*IssueAssetAction    `json:"assets"`
}

func PluginDocJsonUnMarshal(raw json.RawMessage) (pd *PluginDoc, err error) {
	pd = new(PluginDoc)
	err = json.Unmarshal(raw, pd)
	return
}

// CreateAccount create account
func (pd *PluginDoc) CreateAccount(pabi *abi.ABI, chainName, accountName string) ([]*types.Transaction, error) {
	var txs []*types.Transaction

	// see account.Sol_CreateAccount for params detail
	payload, err := pabi.Pack("CreateAccount", chainName, common.HexToPubKey("").String(), "")
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
		payload, err := pabi.Pack("CreateAccount", act.Name, act.Pubkey, act.Desc)
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
func (pd *PluginDoc) IssueAsset(pabi *abi.ABI, chainName, assetName string) ([]*types.Transaction, error) {
	var txs []*types.Transaction

	for _, ast := range pd.Assets {
		// see asset.Sol_IssueAsset for params detail
		payload, err := pabi.Pack("IssueAsset", ast.AssetName, ast.Symbol, ast.Amount, ast.Decimals,
			ast.Founder, ast.Owner, ast.UpperLimit, ast.Description)
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
func (pd *PluginDoc) RegisterMiner(pabi *abi.ABI, sysName, dposName string) ([]*types.Transaction, error) {
	// see consensus.Sol_RegisterMiner for params detail
	payload, err := pabi.Pack("RegisterMiner", "")
	if err != nil {
		return nil, err
	}
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
		payload,
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
			Description: "",
		},
	}
}
