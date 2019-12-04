package plugin

import (
	"regexp"

	"github.com/fractalplatform/fractal/state"
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

type Attribute struct {
	Name        string
	Description string
}

type ItemTxParam struct {
	ItemTypeID uint64
	ItemInfoID uint64
	Amount     uint64
}

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

func (im *ItemManager) IssueItemType(creator, owner, name, description string, am IAccount) ([]byte, error) {
	return nil, nil
}
func (im *ItemManager) UpdateItemTypeOwner(from, newOwner string, itemTypeID uint64, am IAccount) ([]byte, error) {
	return nil, nil
}
func (im *ItemManager) IssueItem(creator string, itemTypeID uint64, name string, description string, upperLimit uint64, total uint64, attributes []*Attribute, am IAccount) ([]byte, error) {
	return nil, nil
}
func (im *ItemManager) IncreaseItem(from string, itemTypeID, itemInfoID uint64, to string, amount uint64, am IAccount) ([]byte, error) {
	return nil, nil
}
func (im *ItemManager) TransferItem(from, to string, ItemTx []*ItemTxParam) error {
	return nil
}
func (im *ItemManager) GetItemAmount(account string, itemTypeID, itemInfoID uint64) (uint64, error) {
	return 0, nil
}
func (im *ItemManager) GetItemAttribute(itemTypeID uint64, itemInfoID uint64, AttributeName string) (string, error) {
	return "", nil
}
