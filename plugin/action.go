package plugin

import (
	"math/big"

	"github.com/fractalplatform/fractal/types/envelope"
)

const (
	// CreateAccount repesents the create account.
	CreateAccount envelope.PayloadType = 0x100 + iota
	ChangePubKey
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

const (
	IssueItemType envelope.PayloadType = 0x400 + iota
	UpdateItemTypeOwner
	IssueItem
	IncreaseItem
	TransferItem
)

type CreateAccountAction struct {
	Name   string
	Pubkey string
	Desc   string
}

type ChangePubKeyAction struct {
	Name   string
	Pubkey string
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

type IssueItemTypeAction struct {
	Owner       string
	Name        string
	Description string
}

type UpdateItemTypeOwnerAction struct {
	NewOwner   string
	ItemTypeID uint64
}

type IssueItemAction struct {
	ItemTypeID  uint64
	Name        string
	Description string
	UpperLimit  uint64
	Total       uint64
	Attributes  []*Attribute
}

type IncreaseItemAction struct {
	ItemTypeID uint64
	ItemInfoID uint64
	To         string
	Amount     uint64
}

type TransferItemAction struct {
	To     string
	ItemTx []*ItemTxParam
}

type Attribute struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ItemTxParam struct {
	ItemTypeID uint64
	ItemInfoID uint64
	Amount     uint64
}
