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
	IssueWorld envelope.PayloadType = 0x400 + iota
	UpdateWorldOwner
	IssueItemType
	IncreaseItem
	// IncreaseItem
	DestroyItem
	// IssueItems
	IncreaseItems
	DestroyItems
	TransferItem
	AddItemTypeAttributes
	DelItemTypeAttributes
	ModifyItemTypeAttributes
	AddItemAttributes
	DelItemAttributes
	ModifyItemAttributes
)

type CreateAccountAction struct {
	Name   string
	Pubkey string
	Desc   string
}

type ChangePubKeyAction struct {
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
	Description string
}

type IssueWorldAction struct {
	Owner       string
	Name        string
	Description string
}

type UpdateWorldOwnerAction struct {
	NewOwner string
	WorldID  uint64
}

type IssueItemTypeAction struct {
	WorldID     uint64
	Name        string
	Merge       bool
	UpperLimit  uint64
	Description string
	Attributes  []*Attribute
}

type IncreaseItemAction struct {
	WorldID     uint64
	ItemTypeID  uint64
	Owner       string
	Description string
	Attributes  []*Attribute
}

type DestroyItemAction struct {
	WorldID    uint64
	ItemTypeID uint64
	ItemID     uint64
}

type IncreaseItemsAction struct {
	WorldID    uint64
	ItemTypeID uint64
	Owner      string
	Count      uint64
}

type DestroyItemsAction struct {
	WorldID    uint64
	ItemTypeID uint64
	Count      uint64
}

type TransferItemAction struct {
	To     string
	ItemTx []*ItemTxParam
}

type AddItemTypeAttributesAction struct {
	WorldID    uint64
	ItemTypeID uint64
	Attributes []*Attribute
}

type DelItemTypeAttributesAction struct {
	WorldID    uint64
	ItemTypeID uint64
	AttrName   []string
}

type ModifyItemTypeAttributesAction struct {
	WorldID    uint64
	ItemTypeID uint64
	Attributes []*Attribute
}

type AddItemAttributesAction struct {
	WorldID    uint64
	ItemTypeID uint64
	ItemID     uint64
	Attributes []*Attribute
}

type DelItemAttributesAction struct {
	WorldID    uint64
	ItemTypeID uint64
	ItemID     uint64
	AttrName   []string
}

type ModifyItemAttributesAction struct {
	WorldID    uint64
	ItemTypeID uint64
	ItemID     uint64
	Attributes []*Attribute
}

// type ModifyPermission int

const (
	CannotModify uint64 = 0
	WorldOwner   uint64 = 1
	ItemOwner    uint64 = 2
)

type Attribute struct {
	Permission  uint64 `json:"modifyPermission"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ItemTxParam struct {
	WorldID    uint64
	ItemTypeID uint64
	ItemID     uint64
	Amount     uint64
}
