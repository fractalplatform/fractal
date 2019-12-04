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
	ID          uint64 // 不可修改
	Name        string // 保证不重复
	Owner       string // 允许修改
	Creator     string // 不可修改
	CreateTime  uint64 // 不可修改
	Description string // 不可修改
	Total       uint64 // 已发行总数
}

type ItemInfo struct {
	TypeID      uint64       // 不可修改
	ID          uint64       // 保证不重复
	Name        string       // 不可修改
	CreateTime  uint64       // 不可修改
	Total       uint64       // 增发时修改
	Description string       // 不可修改
	UpperLimit  uint64       // 上限，不可修改
	Attributes  []*Attribute // 发行时确定好，后续禁止修改
}

type Attribute struct {
	Name        string // 不可修改
	Description string // 不可修改
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

// em.sdb.Put("item", "accuntItemAmount"+类型id+道具id+账号名, 道具数量)
// em.sdb.Put("item", "itemType"+类型id, ItemType obj)
// em.sdb.Put("item", "itemInfo"+类型id+道具id, ItemInfo obj)
// em.sdb.Put("item", "itemTyepName"+类型创建者+类型名, 类型id)
// em.sdb.Put("item", "itemInfoName"+类型id+道具名, 道具id)
