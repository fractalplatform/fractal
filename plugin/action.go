package plugin

import (
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

const (
	// CreateAccount repesents the create account.
	CreateAccount types.ActionType = 0x100 + iota
)

const (
	// CreateAccount repesents the create account.
	IssueAsset types.ActionType = 0x200 + iota
	IncreaseAsset
)

type CreateAccountAction struct {
	Name   string
	Pubkey common.PubKey
	Desc   string
}

type IncreaseAssetAction struct {
	AssetID uint64
	Amount  *big.Int
	To      string
}

type IssueAssetAction struct {
	AssetName   string
	Symbol      string
	Amount      *big.Int
	Owner       string
	Founder     string
	Decimals    uint64
	UpperLimit  *big.Int
	Contract    string
	Description string
}
