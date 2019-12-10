package plugin

import (
	"math/big"

	"github.com/fractalplatform/fractal/types/envelope"
)

const (
	// CreateAccount repesents the create account.
	CreateAccount envelope.PayloadType = 0x100 + iota
)

const (
	IssueAsset envelope.PayloadType = 0x200 + iota
	IncreaseAsset
	Transfer
)

const (
	// RegisterMiner register msg.sender become a miner
	RegisterMiner envelope.PayloadType = 0x300 + iota
	UnregisterMiner
	ConsensusEnd
)

type CreateAccountAction struct {
	Name   string
	Pubkey string
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
