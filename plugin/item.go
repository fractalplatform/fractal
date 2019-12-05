package plugin

import (
	"errors"
	"regexp"

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	itemRegExp              = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9]{1,31})`)
	itemNameMaxLength       = uint64(32)
	itemManager             = "item"
	counterPrefix           = "itemTypeCounter"
	accountItemAmountPrefix = "accuntItemAmount"
	itemTypePrefix          = "itemType"
	itemInfoPrefix          = "itemInfo"
	itemTypeNamePrefix      = "itemTyepName"
	itemInfoNamePrefix      = "itemInfoName"
)

type ItemManager struct {
	sdb *state.StateDB
}

type ItemType struct {
	ID          uint64
	Name        string
	Owner       string
	Creator     string
	CreateTime  uint64
	Description string
	Total       uint64
}

type ItemInfo struct {
	TypeID      uint64
	ID          uint64
	Name        string
	CreateTime  uint64
	Total       uint64
	Description string
	UpperLimit  uint64
	Attributes  []*Attribute
}

// type Attribute struct {
// 	Name        string
// 	Description string
// }

// type ItemTxParam struct {
// 	ItemTypeID uint64
// 	ItemInfoID uint64
// 	Amount     uint64
// }

// NewIM new a ItemManager
func NewItemManage(sdb *state.StateDB) (*ItemManager, error) {
	if sdb == nil {
		return nil, ErrNewAssetManagerErr
	}

	itemManager := ItemManager{
		sdb: sdb,
	}
	return &itemManager, nil
}

func (im *ItemManager) AccountName() string {
	return "item"
}

func (im *ItemManager) CallTx(action *types.Action, pm IPM) ([]byte, error) {
	switch action.Type() {
	case IssueItemType:
		param := &IssueItemTypeAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		return im.IssueItemType(action.Sender(), param.Owner, param.Name, param.Description, pm)
	case UpdateItemTypeOwner:
		param := &UpdateItemTypeOwnerAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		return im.UpdateItemTypeOwner(action.Sender(), param.NewOwner, param.ItemTypeID, pm)
	case IssueItem:
		param := &IssueItemAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		return im.IssueItem(action.Sender(), param.ItemTypeID, param.Name, param.Description, param.UpperLimit, param.Total, param.Attributes, pm)
	case IncreaseItem:
		param := &IncreaseItemAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		return im.IncreaseItem(action.Sender(), param.ItemTypeID, param.ItemInfoID, param.To, param.Amount, pm)
	case TransferItem:
		param := &TransferItemAction{}
		if err := rlp.DecodeBytes(action.Data(), param); err != nil {
			return nil, err
		}
		err := im.TransferItem(action.Sender(), param.To, param.ItemTx)
		return nil, err
	}
	return nil, ErrWrongAction
}

// IssueItemType issue itemType
func (im *ItemManager) IssueItemType(creator, owner, name, description string, am IAccount) ([]byte, error) {
	return nil, nil
}

// UpdateItemTypeOwner update itemType owner
func (im *ItemManager) UpdateItemTypeOwner(from, newOwner string, itemTypeID uint64, am IAccount) ([]byte, error) {
	return nil, nil
}

// IssueItem issue new item
func (im *ItemManager) IssueItem(creator string, itemTypeID uint64, name string, description string, upperLimit uint64, total uint64, attributes []*Attribute, am IAccount) ([]byte, error) {
	return nil, nil
}

// IncreaseItem increase item
func (im *ItemManager) IncreaseItem(from string, itemTypeID, itemInfoID uint64, to string, amount uint64, am IAccount) ([]byte, error) {
	return nil, nil
}

func (im *ItemManager) transferItemSingle(from, to string, itemTypeID, itemInfoID, amount uint64) error {
	return nil
}

// TransferItem transfer item
func (im *ItemManager) TransferItem(from, to string, ItemTx []*ItemTxParam) error {
	return nil
}

// GetItemAmount get account item amount
func (im *ItemManager) GetItemAmount(account string, itemTypeID, itemInfoID uint64) (uint64, error) {
	return 0, nil
}

// GetItemTypeOwner get itemType owner
func (im *ItemManager) GetItemTypeOwner(itemTypeID uint64) (string, error) {
	return "", nil
}

// GetItemAttribute get item attribute
func (im *ItemManager) GetItemAttribute(itemTypeID uint64, itemInfoID uint64, AttributeName string) (string, error) {
	return "", nil
}

var (
	ErrItemCounterNotExist     = errors.New("item global counter not exist")
	ErrItemNameinvalid         = errors.New("item name invalid")
	ErrItemTypeNameNotExist    = errors.New("itemTypeName not exist")
	ErrItemTypeNameIsExist     = errors.New("itemTypeName is exist")
	ErrItemInfoNameNotExist    = errors.New("itemInfoName not exist")
	ErrItemInfoNameIsExist     = errors.New("itemInfoName is exist")
	ErrItemTypeNotExist        = errors.New("itemType not exist")
	ErrItemTypeIsExist         = errors.New("itemType is exist")
	ErrItemInfoNotExist        = errors.New("itemInfo not exist")
	ErrItemInfoIsExist         = errors.New("itemInfo is exist")
	ErrItemObjectEmpty         = errors.New("item object is empty")
	ErrItemOwnerMismatch       = errors.New("itemType owner mismatch")
	ErrItemAttributeDesTooLong = errors.New("item attribute description exceed max length")
	ErrItemUpperLimit          = errors.New("item amount over the issuance limit")
	ErrAccountNoItem           = errors.New("account not have item")
)
