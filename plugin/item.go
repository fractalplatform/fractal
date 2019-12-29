package plugin

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	itemRegExp             = regexp.MustCompile(`^([a-z][a-z0-9]{1,31})$`)
	worldNameRegExp        = regexp.MustCompile(`^([a-z][a-z0-9]{11})$`)
	itemNameMaxLength      = uint64(32)
	itemManager            = "item"
	counterPrefix          = "worldCounter"
	worldNamePrefix        = "worldName"
	worldIDPrefix          = "worldID"
	itemTypeNamePrefix     = "itemTypeName"
	itemTypeIDPrefix       = "itemTypeID"
	itemIDPrefix           = "itemID"
	itemOwnerPrefix        = "itemOwner"
	itemsPrefix            = "items"
	itemTypeAttrIDPrefix   = "itemTypeAttrID"
	itemTypeAttrNamePrefix = "itemTypeAttrName"
	itemAttrIDPrefix       = "itemAttrID"
	itemAttrNamePrefix     = "itemAttrName"
)

const UINT64_MAX uint64 = ^uint64(0)

type ItemManager struct {
	sdb *state.StateDB
}

type World struct {
	ID      uint64 `json:"worldID"`
	Name    string `json:"name"`
	Owner   string `json:"owner"`
	Creator string `json:"creator"`
	// CreateTime  uint64 `json:"createTime"`
	Description string `json:"description"`
	Total       uint64 `json:"total"`
}

type ItemType struct {
	WorldID uint64 `json:"worldID"`
	ID      uint64 `json:"itemTypeID"`
	Name    string `json:"name"`
	// CreateTime  uint64 `json:"createTime"`
	Merge       bool   `json:"merge"`
	UpperLimit  uint64 `json:"upperLimit"`
	AddIssue    uint64 `json:"addIssue"`
	Description string `json:"description"`
	Total       uint64 `json:"total"`
	AttrTotal   uint64 `json:"attrTotal"`
}

type Item struct {
	WorldID     uint64 `json:"worldID"`
	TypeID      uint64 `json:"itemTypeID"`
	ID          uint64 `json:"itemID"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	Destroy     bool   `json:"destroy"`
	// Attributes  []*Attribute `json:"attributes"`
	AttrTotal uint64 `json:"attrTotal"`
}

type Items struct {
	WorldID uint64 `json:"worldID"`
	TypeID  uint64 `json:"itemTypeID"`
	Owner   string `json:"owner"`
	Amount  uint64 `json:"total"`
}

// NewIM new a ItemManager
func NewItemManage(sdb *state.StateDB) (*ItemManager, error) {
	if sdb == nil {
		return nil, ErrNewAssetManagerErr
	}

	itemManager := ItemManager{
		sdb: sdb,
	}
	itemManager.initWorldCounter()
	return &itemManager, nil
}

func (im *ItemManager) AccountName() string {
	return "fractalitem"
}

func (im *ItemManager) CallTx(tx *envelope.PluginTx, ctx *Context, pm IPM) ([]byte, error) {
	switch tx.PayloadType() {
	case IssueWorld:
		param := &IssueWorldAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.IssueWorld(tx.Sender(), param.Owner, param.Name, param.Description, pm)
	case UpdateWorldOwner:
		param := &UpdateWorldOwnerAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.UpdateWorldOwner(tx.Sender(), param.NewOwner, param.WorldID, pm)
	case IssueItemType:
		param := &IssueItemTypeAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.IssueItemType(tx.Sender(), param.WorldID, param.Name, param.Merge, param.UpperLimit, param.Description, param.Attributes, pm)
	case IncreaseItem:
		param := &IncreaseItemAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.IncreaseItem(tx.Sender(), param.WorldID, param.ItemTypeID, param.Owner, param.Description, param.Attributes, pm)
	case DestroyItem:
		param := &DestroyItemAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.DestroyItem(tx.Sender(), param.WorldID, param.ItemTypeID, param.ItemID, pm)
	case IncreaseItems:
		param := &IncreaseItemsAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.IncreaseItems(tx.Sender(), param.WorldID, param.ItemTypeID, param.Owner, param.Count, pm)
	case DestroyItems:
		param := &DestroyItemsAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.DestroyItems(tx.Sender(), param.WorldID, param.ItemTypeID, param.Count, pm)
	case TransferItem:
		param := &TransferItemAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		err := im.TransferItem(tx.Sender(), param.To, param.ItemTx, pm)
		return nil, err
	case AddItemTypeAttributes:
		param := &AddItemTypeAttributesAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.AddItemTypeAttributes(tx.Sender(), param.WorldID, param.ItemTypeID, param.Attributes)
	case DelItemTypeAttributes:
		param := &DelItemTypeAttributesAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.DelItemTypeAttributes(tx.Sender(), param.WorldID, param.ItemTypeID, param.AttrName)
	case ModifyItemTypeAttributes:
		param := &ModifyItemTypeAttributesAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.ModifyItemTypeAttributes(tx.Sender(), param.WorldID, param.ItemTypeID, param.Attributes)
	case AddItemAttributes:
		param := &AddItemAttributesAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.AddItemAttributes(tx.Sender(), param.WorldID, param.ItemTypeID, param.ItemID, param.Attributes)
	case DelItemAttributes:
		param := &DelItemAttributesAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.DelItemAttributes(tx.Sender(), param.WorldID, param.ItemTypeID, param.ItemID, param.AttrName)
	case ModifyItemAttributes:
		param := &ModifyItemAttributesAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.ModifyItemAttributes(tx.Sender(), param.WorldID, param.ItemTypeID, param.ItemID, param.Attributes)
	}

	return nil, ErrWrongTransaction
}

// IssueItemType issue itemType
func (im *ItemManager) IssueWorld(creator, owner, name, description string, am IAccount) ([]byte, error) {
	if err := im.checkWorldNameFormat(name); err != nil {
		return nil, err
	}
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}

	if _, err := am.getAccount(owner); err != nil {
		return nil, err
	}

	_, err := im.getWorldIDByName(name)
	if err == nil {
		return nil, ErrWorldNameIsExist
	} else if err != ErrWorldNameNotExist {
		return nil, err
	}

	worldID, err := im.getWorldCounter()
	if err != nil {
		return nil, err
	}

	worldObj := World{
		ID:      worldID,
		Name:    name,
		Owner:   owner,
		Creator: creator,
		// CreateTime:  uint64(time.Now().Unix()),
		Description: description,
		Total:       0,
	}
	err = im.setNewWorld(&worldObj)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (im *ItemManager) UpdateWorldOwner(from, newOwner string, worldID uint64, am IAccount) ([]byte, error) {
	if _, err := am.getAccount(newOwner); err != nil {
		return nil, err
	}
	world, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if from != world.Owner {
		return nil, ErrItemOwnerMismatch
	}
	world.Owner = newOwner

	err = im.setWorld(world)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (im *ItemManager) IssueItemType(creator string, worldID uint64, name string, merge bool, upperLimit uint64, description string, attributes []*Attribute, am IAccount) ([]byte, error) {
	if err := im.checkNameFormat(name); err != nil {
		return nil, err
	}
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}

	if upperLimit > UINT64_MAX {
		return nil, ErrAmountValueInvalid
	}

	if _, err := am.getAccount(creator); err != nil {
		return nil, err
	}

	err := im.checkAttribute(attributes)
	if err != nil {
		return nil, err
	}
	for _, attr := range attributes {
		if attr.Permission == ItemOwner {
			return nil, ErrInvalidPermission
		}
	}

	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}

	if creator != worldobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	_, err = im.getItemTypeIDByName(worldID, name)
	if err == nil {
		return nil, ErrItemTypeNameIsExist
	} else if err != ErrItemTypeNameNotExist {
		return nil, err
	}

	itemTypeID := worldobj.Total + 1

	itemTypeobj := ItemType{
		WorldID:    worldID,
		ID:         itemTypeID,
		Name:       name,
		Merge:      merge,
		UpperLimit: upperLimit,
		AddIssue:   0,
		// CreateTime:  uint64(time.Now().Unix()),
		Description: description,
		Total:       0,
		AttrTotal:   uint64(len(attributes)),
	}

	worldobj.Total += 1
	err = im.setWorld(worldobj)
	if err != nil {
		return nil, err
	}

	err = im.setNewItemType(&itemTypeobj)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attributes); i++ {
		err = im.setNewItemTypeAttr(worldID, itemTypeID, uint64(i+1), attributes[i])
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (im *ItemManager) IncreaseItem(from string, worldID uint64, itemTypeID uint64, owner string, description string, attributes []*Attribute, am IAccount) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if from != worldobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	if _, err := am.getAccount(owner); err != nil {
		return nil, err
	}

	itemTypeobj, err := im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return nil, err
	}
	if itemTypeobj.Merge != false {
		return nil, ErrItemTypeMergeIsTrue
	}

	if itemTypeobj.UpperLimit > 0 {
		if itemTypeobj.AddIssue+1 > itemTypeobj.UpperLimit {
			return nil, ErrItemUpperLimit
		}
	}
	if UINT64_MAX-itemTypeobj.AddIssue < 1 {
		return nil, ErrExceedMax
	}
	itemTypeobj.AddIssue += 1
	itemTypeobj.Total += 1

	err = im.checkAttribute(attributes)
	if err != nil {
		return nil, err
	}

	itemobj := Item{
		WorldID:     worldID,
		TypeID:      itemTypeID,
		ID:          itemTypeobj.AddIssue,
		Owner:       owner,
		Description: description,
		Destroy:     false,
		AttrTotal:   uint64(len(attributes)),
	}

	err = im.setItemType(itemTypeobj)
	if err != nil {
		return nil, err
	}
	err = im.setItem(&itemobj)
	if err != nil {
		return nil, err
	}
	err = im.setItemOwnerDB(&itemobj)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attributes); i++ {
		err = im.setNewItemAttr(worldID, itemTypeID, itemobj.ID, uint64(i+1), attributes[i])
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (im *ItemManager) DestroyItem(from string, worldID uint64, itemTypeID uint64, itemID uint64, am IAccount) ([]byte, error) {
	itemTypeobj, err := im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return nil, err
	}
	itemobj, err := im.getItemByID(worldID, itemTypeID, itemID)
	if err != nil {
		return nil, err
	}
	if from != itemobj.Owner {
		return nil, ErrItemOwnerMismatch
	}
	if itemobj.Destroy == true {
		return nil, ErrItemIsDestroyed
	}

	itemobj.Destroy = true
	err = im.setItem(itemobj)
	if err != nil {
		return nil, err
	}
	itemTypeobj.Total -= 1
	err = im.setItemType(itemTypeobj)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (im *ItemManager) IncreaseItems(from string, worldID uint64, itemTypeID uint64, to string, amount uint64, am IAccount) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if from != worldobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	if _, err := am.getAccount(to); err != nil {
		return nil, err
	}

	itemTypeobj, err := im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return nil, err
	}
	if itemTypeobj.Merge != true {
		return nil, ErrItemTypeMergeIsFalse
	}

	if itemTypeobj.UpperLimit > 0 {
		if itemTypeobj.AddIssue+amount > itemTypeobj.UpperLimit {
			return nil, ErrItemUpperLimit
		}
	}

	if UINT64_MAX-itemTypeobj.AddIssue < amount {
		return nil, ErrExceedMax
	}

	itemTypeobj.AddIssue += amount
	itemTypeobj.Total += amount

	itemsobj, err := im.getItemsByOwner(worldID, itemTypeID, to)
	if err != nil && err != ErrItemsNotExist {
		return nil, err
	}

	if err == ErrItemsNotExist {
		itemsobj = &Items{
			WorldID: worldID,
			TypeID:  itemTypeID,
			Owner:   to,
			Amount:  amount,
		}
	} else {
		itemsobj.Amount += amount
	}

	err = im.setItemType(itemTypeobj)
	if err != nil {
		return nil, err
	}
	err = im.setItems(itemsobj)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (im *ItemManager) DestroyItems(from string, worldID uint64, itemTypeID uint64, amount uint64, am IAccount) ([]byte, error) {
	itemTypeobj, err := im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return nil, err
	}
	itemsobj, err := im.getItemsByOwner(worldID, itemTypeID, from)
	if err != nil {
		return nil, err
	}
	if from != itemsobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	if itemsobj.Amount < amount {
		return nil, ErrInsufficientItemAmount
	}

	itemsobj.Amount -= amount

	if itemsobj.Amount != uint64(0) {
		err = im.setItems(itemsobj)
		if err != nil {
			return nil, err
		}
	} else {
		err = im.delItemsByOwner(worldID, itemTypeID, from)
		if err != nil {
			return nil, err
		}
	}
	itemTypeobj.Total -= amount
	err = im.setItemType(itemTypeobj)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (im *ItemManager) transferItemSingle(from, to string, worldID, itemTypeID, itemID, amount uint64) error {
	itemTypeobj, err := im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return err
	}
	if itemTypeobj.Merge == true {
		if itemID != uint64(0) {
			return ErrInvalidItemID
		}
		if err = im.subItemsAmount(from, worldID, itemTypeID, amount); err != nil {
			return err
		}
		if err = im.addItemsAmount(to, worldID, itemTypeID, amount); err != nil {
			return err
		}
	} else {
		if amount != uint64(1) {
			return ErrInvalidItemAmount
		}
		if err = im.txItem(from, to, worldID, itemTypeID, itemID); err != nil {
			return err
		}
	}
	return nil
}

func (im *ItemManager) txItem(from, to string, worldID, itemTypeID, itemID uint64) error {
	itemobj, err := im.getItemByID(worldID, itemTypeID, itemID)
	if err != nil {
		return err
	}

	if from != itemobj.Owner {
		return ErrItemOwnerMismatch
	}
	if itemobj.Destroy == true {
		return ErrItemIsDestroyed
	}
	err = im.delItem(itemobj)
	if err != nil {
		return err
	}
	err = im.delItemOwnerDB(itemobj)
	if err != nil {
		return err
	}
	itemobj.Owner = to
	err = im.setItem(itemobj)
	if err != nil {
		return err
	}
	err = im.setItemOwnerDB(itemobj)
	if err != nil {
		return err
	}
	return nil
}

func (im *ItemManager) TransferItem(from, to string, ItemTx []*ItemTxParam, am IAccount) error {
	if from == to {
		return nil
	}
	if _, err := am.getAccount(to); err != nil {
		return err
	}

	for _, tx := range ItemTx {
		if err := im.transferItemSingle(from, to, tx.WorldID, tx.ItemTypeID, tx.ItemID, tx.Amount); err != nil {
			return err
		}
	}
	return nil
}

func (im *ItemManager) AddItemTypeAttributes(from string, worldID, itemTypeID uint64, attributes []*Attribute) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if from != worldobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	itemTypeobj, err := im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return nil, err
	}

	for _, attr := range attributes {
		if attr.Permission == ItemOwner {
			return nil, ErrInvalidPermission
		}
	}

	err = im.checkAttribute(attributes)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attributes); i++ {
		_, err = im.getItemTypeAttrByName(worldID, itemTypeID, attributes[i].Name)
		if err == nil {
			return nil, ErrItemTypeAttrIsExist
		} else if err != ErrItemTypeAttrNotExist {
			return nil, err
		}

		err = im.setNewItemTypeAttr(worldID, itemTypeID, uint64(i+1)+itemTypeobj.AttrTotal, attributes[i])
		if err != nil {
			return nil, err
		}
	}

	itemTypeobj.AttrTotal += uint64(len(attributes))
	err = im.setItemType(itemTypeobj)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (im *ItemManager) DelItemTypeAttributes(from string, worldID, itemTypeID uint64, attrName []string) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if from != worldobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	_, err = im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attrName); i++ {
		err = im.delItemTypeAttrByName(worldID, itemTypeID, attrName[i])
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (im *ItemManager) ModifyItemTypeAttributes(from string, worldID, itemTypeID uint64, attributes []*Attribute) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}

	for _, attr := range attributes {
		if attr.Permission == ItemOwner {
			return nil, ErrInvalidPermission
		}
	}

	err = im.checkAttribute(attributes)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attributes); i++ {
		err = im.modifyitemTypeAttr(worldobj.Owner, from, worldID, itemTypeID, attributes[i])
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (im *ItemManager) AddItemAttributes(from string, worldID, itemTypeID, itemID uint64, attributes []*Attribute) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if from != worldobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	itemobj, err := im.getItemByID(worldID, itemTypeID, itemID)
	if err != nil {
		return nil, err
	}

	err = im.checkAttribute(attributes)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attributes); i++ {
		_, err = im.getItemAttrByName(worldID, itemTypeID, itemID, attributes[i].Name)
		if err == nil {
			return nil, ErrItemAttrIsExist
		} else if err != ErrItemAttrNotExist {
			return nil, err
		}

		err = im.setNewItemAttr(worldID, itemTypeID, itemID, uint64(i+1)+itemobj.AttrTotal, attributes[i])
		if err != nil {
			return nil, err
		}
	}

	itemobj.AttrTotal += uint64(len(attributes))
	err = im.setItem(itemobj)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (im *ItemManager) DelItemAttributes(from string, worldID, itemTypeID, itemID uint64, attrName []string) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if from != worldobj.Owner {
		return nil, ErrItemOwnerMismatch
	}

	_, err = im.getItemByID(worldID, itemTypeID, itemID)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attrName); i++ {
		err = im.delItemAttrByName(worldID, itemTypeID, itemID, attrName[i])
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (im *ItemManager) ModifyItemAttributes(from string, worldID, itemTypeID, itemID uint64, attributes []*Attribute) ([]byte, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}

	err = im.checkAttribute(attributes)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(attributes); i++ {
		err = im.modifyitemAttr(worldobj.Owner, from, worldID, itemID, itemTypeID, attributes[i])
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (im *ItemManager) initWorldCounter() {
	_, err := im.getWorldCounter()
	if err == ErrWorldCounterNotExist {
		var worldID uint64 = 1
		b, err := rlp.EncodeToBytes(&worldID)
		if err != nil {
			panic(err)
		}
		im.sdb.Put(itemManager, counterPrefix, b)
	}
}

func (im *ItemManager) getWorldCounter() (uint64, error) {
	b, err := im.sdb.Get(itemManager, counterPrefix)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, ErrWorldCounterNotExist
	}
	var worldCounter uint64
	err = rlp.DecodeBytes(b, &worldCounter)
	if err != nil {
		return 0, err
	}
	return worldCounter, nil
}

func (im *ItemManager) setNewWorld(worldobj *World) error {
	if worldobj == nil {
		return ErrWorldObjectEmpty
	}
	worldID := worldobj.ID

	b, err := rlp.EncodeToBytes(worldobj)
	if err != nil {
		return err
	}
	id, err := rlp.EncodeToBytes(&worldID)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(worldIDPrefix, strconv.FormatUint(worldID, 10)), b)
	im.sdb.Put(itemManager, dbKey(worldNamePrefix, worldobj.Name), id)

	worldID += 1
	nid, err := rlp.EncodeToBytes(&worldID)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, counterPrefix, nid)
	return nil
}

func (im *ItemManager) setWorld(worldobj *World) error {
	if worldobj == nil {
		return ErrWorldObjectEmpty
	}
	b, err := rlp.EncodeToBytes(worldobj)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(worldIDPrefix, strconv.FormatUint(worldobj.ID, 10)), b)
	return nil
}

func (im *ItemManager) getWorldByName(worldName string) (*World, error) {
	worldID, err := im.getWorldIDByName(worldName)
	if err != nil {
		return nil, err
	}
	return im.getWorldByID(worldID)
}

func (im *ItemManager) getWorldByID(worldID uint64) (*World, error) {
	b, err := im.sdb.Get(itemManager, dbKey(worldIDPrefix, strconv.FormatUint(worldID, 10)))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrWorldNotExist
	}

	var worldobj World
	if err = rlp.DecodeBytes(b, &worldobj); err != nil {
		return nil, err
	}
	return &worldobj, nil
}

func (im *ItemManager) getWorldIDByName(name string) (uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(worldNamePrefix, name))
	if len(b) == 0 {
		return 0, ErrWorldNameNotExist
	}
	var WorldID uint64
	if err = rlp.DecodeBytes(b, &WorldID); err != nil {
		return 0, err
	}
	return WorldID, nil
}

func (im *ItemManager) checkWorldNameFormat(name string) error {
	if worldNameRegExp.MatchString(name) != true {
		return ErrWorldNameinvalid
	}
	return nil
}

func (im *ItemManager) checkNameFormat(name string) error {
	if uint64(len(name)) > itemNameMaxLength {
		return ErrItemNameLengthErr
	}

	if itemRegExp.MatchString(name) != true {
		return ErrItemNameinvalid
	}
	return nil
}

func (im *ItemManager) checkAttribute(attributes []*Attribute) error {
	dict := map[string]int{}
	for i, attr := range attributes {
		if err := im.checkNameFormat(attr.Name); err != nil {
			return err
		}
		if uint64(len(attr.Description)) > MaxDescriptionLength {
			return ErrItemAttributeDesTooLong
		}
		if attr.Permission > ItemOwner {
			return ErrInvalidPermission
		}
		if _, ok := dict[attr.Name]; ok {
			return ErrDuplicateAttr
		}
		dict[attr.Name] = i
	}
	return nil
}

func (im *ItemManager) setNewItemType(itemTypeobj *ItemType) error {
	if itemTypeobj == nil {
		return ErrItemTypeObjectEmpty
	}
	id := itemTypeobj.ID
	b, err := rlp.EncodeToBytes(itemTypeobj)
	if err != nil {
		return err
	}
	nid, err := rlp.EncodeToBytes(&id)
	if err != nil {
		return err
	}

	im.sdb.Put(itemManager, dbKey(itemTypeIDPrefix, strconv.FormatUint(itemTypeobj.WorldID, 10), strconv.FormatUint(itemTypeobj.ID, 10)), b)
	im.sdb.Put(itemManager, dbKey(itemTypeNamePrefix, strconv.FormatUint(itemTypeobj.WorldID, 10), itemTypeobj.Name), nid)
	return nil
}

func (im *ItemManager) setItemType(itemobj *ItemType) error {
	if itemobj == nil {
		return ErrItemObjectEmpty
	}
	b, err := rlp.EncodeToBytes(itemobj)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemTypeIDPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.ID, 10)), b)
	return nil
}

func (im *ItemManager) getItemTypeByName(worldID uint64, name string) (*ItemType, error) {
	itemTypeID, err := im.getItemTypeIDByName(worldID, name)
	if err != nil {
		return nil, err
	}
	return im.getItemTypeByID(worldID, itemTypeID)
}

func (im *ItemManager) getItemTypeIDByName(worldID uint64, name string) (uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemTypeNamePrefix, strconv.FormatUint(worldID, 10), name))
	if len(b) == 0 {
		return 0, ErrItemTypeNameNotExist
	}
	var itemTypeID uint64
	if err = rlp.DecodeBytes(b, &itemTypeID); err != nil {
		return 0, err
	}
	return itemTypeID, nil
}

func dbKey(s ...string) string {
	return strings.Join(s, "_")
}

func (im *ItemManager) getItemTypeByID(worldID, itemTypeID uint64) (*ItemType, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemTypeIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10)))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrItemTypeNotExist
	}

	var itemType ItemType
	if err = rlp.DecodeBytes(b, &itemType); err != nil {
		return nil, err
	}
	return &itemType, nil
}

func (im *ItemManager) setNewItemTypeAttr(worldID, itemTypeID, attrID uint64, attrobj *Attribute) error {
	b, err := rlp.EncodeToBytes(attrobj)
	if err != nil {
		return err
	}
	id, err := rlp.EncodeToBytes(&attrID)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemTypeAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(attrID, 10)), b)
	im.sdb.Put(itemManager, dbKey(itemTypeAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), attrobj.Name), id)
	return nil
}

func (im *ItemManager) setItemTypeAttr(worldID, itemTypeID, attrID uint64, attrobj *Attribute) error {
	b, err := rlp.EncodeToBytes(attrobj)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemTypeAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(attrID, 10)), b)
	return nil
}

func (im *ItemManager) getItemTypeAttrByName(worldID, itemTypeID uint64, attrName string) (*Attribute, error) {
	attrID, err := im.getItemTypeAttrIDByName(worldID, itemTypeID, attrName)
	if err != nil {
		return nil, err
	}
	return im.getItemTypeAttrByID(worldID, itemTypeID, attrID)
}

func (im *ItemManager) getItemTypeAttrIDByName(worldID, itemTypeID uint64, attrName string) (uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemTypeAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), attrName))
	if len(b) == 0 {
		return 0, ErrItemTypeAttrNotExist
	}
	var attrID uint64
	err = rlp.DecodeBytes(b, &attrID)
	if err != nil {
		return 0, err
	}
	return attrID, nil
}

func (im *ItemManager) getItemTypeAttrByID(worldID, itemTypeID, attrID uint64) (*Attribute, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemTypeAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(attrID, 10)))
	if len(b) == 0 {
		return nil, ErrItemTypeAttrNotExist
	}
	var attrobj Attribute
	if err = rlp.DecodeBytes(b, &attrobj); err != nil {
		return nil, err
	}
	return &attrobj, nil
}

func (im *ItemManager) delItemTypeAttrByName(worldID, itemTypeID uint64, attrName string) error {
	b, err := im.sdb.Get(itemManager, dbKey(itemTypeAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), attrName))
	if len(b) == 0 {
		return ErrItemTypeAttrNotExist
	}
	var attrID uint64
	err = rlp.DecodeBytes(b, &attrID)
	if err != nil {
		return err
	}
	im.sdb.Delete(itemManager, dbKey(itemTypeAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), attrName))
	im.sdb.Delete(itemManager, dbKey(itemTypeAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(attrID, 10)))
	return nil
}

func (im *ItemManager) modifyitemTypeAttr(worldOwner, from string, worldID, itemTypeID uint64, attr *Attribute) error {
	attrID, err := im.getItemTypeAttrIDByName(worldID, itemTypeID, attr.Name)
	if err != nil {
		return err
	}

	attrobj, err := im.getItemTypeAttrByID(worldID, itemTypeID, attrID)
	if err != nil {
		return err
	}

	if attrobj.Permission != WorldOwner {
		return ErrNoPermission
	}
	if worldOwner != from {
		return ErrNoPermission
	}
	if attr.Permission == ItemOwner {
		return ErrInvalidPermission
	}
	attrobj.Permission = attr.Permission
	attrobj.Description = attr.Description
	err = im.setItemTypeAttr(worldID, itemTypeID, attrID, attrobj)
	if err != nil {
		return err
	}

	return nil
}

func (im *ItemManager) setItem(itemobj *Item) error {
	if itemobj == nil {
		return ErrItemObjectEmpty
	}
	b, err := rlp.EncodeToBytes(itemobj)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemIDPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), strconv.FormatUint(itemobj.ID, 10)), b)
	// im.sdb.Put(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), itemobj.Owner), b)
	return nil
}

func (im *ItemManager) setItemOwnerDB(itemobj *Item) error {
	itemIDarr, err := im.getItemOwnerDB(itemobj)
	if err != nil && err != ErrItemNotExist {
		return err
	}
	if err == ErrItemNotExist {
		temp := []uint64{itemobj.ID}
		b, err := rlp.EncodeToBytes(temp)
		if err != nil {
			return err
		}
		im.sdb.Put(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), itemobj.Owner), b)
		return nil
	}
	itemIDarr = append(itemIDarr, itemobj.ID)
	b2, err := rlp.EncodeToBytes(itemIDarr)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), itemobj.Owner), b2)
	return nil
}

func (im *ItemManager) getItemOwnerDB(itemobj *Item) ([]uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), itemobj.Owner))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrItemNotExist
	}
	var itemIDarr []uint64
	if err = rlp.DecodeBytes(b, &itemIDarr); err != nil {
		return nil, err
	}
	return itemIDarr, nil
}

func (im *ItemManager) delItemOwnerDB(itemobj *Item) error {
	itemIDarr, err := im.getItemOwnerDB(itemobj)
	if err != nil {
		return err
	}
	if len(itemIDarr) == 1 {
		im.sdb.Delete(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), itemobj.Owner))
		return nil
	}
	i := 0
	for ; i < len(itemIDarr); i++ {
		if itemIDarr[i] == itemobj.ID {
			break
		}
	}
	itemIDarr = append(itemIDarr[:i], itemIDarr[i+1:]...)
	b, err := rlp.EncodeToBytes(itemIDarr)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), itemobj.Owner), b)
	return nil
}

func (im *ItemManager) delItem(itemobj *Item) error {
	if itemobj == nil {
		return ErrItemObjectEmpty
	}
	im.sdb.Delete(itemManager, dbKey(itemIDPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), strconv.FormatUint(itemobj.ID, 10)))
	// im.sdb.Delete(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(itemobj.WorldID, 10), strconv.FormatUint(itemobj.TypeID, 10), itemobj.Owner))
	return nil
}

func (im *ItemManager) getItemByID(worldID, itemTypeID, itemID uint64) (*Item, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10)))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrItemNotExist
	}

	var item Item
	if err = rlp.DecodeBytes(b, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (im *ItemManager) getItemByOwner(worldID, itemTypeID uint64, owner string) ([]*Item, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemOwnerPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), owner))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrItemNotExist
	}

	var itemIDarr []uint64
	if err = rlp.DecodeBytes(b, &itemIDarr); err != nil {
		return nil, err
	}
	var item []*Item
	for _, id := range itemIDarr {
		temp, err := im.getItemByID(worldID, itemTypeID, id)
		if err != nil {
			return nil, err
		}
		item = append(item, temp)
	}
	return item, nil
}

func (im *ItemManager) setNewItemAttr(worldID, itemTypeID, itemID, attrID uint64, attrobj *Attribute) error {
	b, err := rlp.EncodeToBytes(attrobj)
	if err != nil {
		return err
	}
	id, err := rlp.EncodeToBytes(&attrID)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), strconv.FormatUint(attrID, 10)), b)
	im.sdb.Put(itemManager, dbKey(itemAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), attrobj.Name), id)
	return nil
}

func (im *ItemManager) setItemAttr(worldID, itemTypeID, itemID, attrID uint64, attrobj *Attribute) error {
	b, err := rlp.EncodeToBytes(attrobj)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), strconv.FormatUint(attrID, 10)), b)
	return nil
}

func (im *ItemManager) getItemAttrByName(worldID, itemTypeID, itemID uint64, attrName string) (*Attribute, error) {
	attrID, err := im.getItemAttrIDByName(worldID, itemTypeID, itemID, attrName)
	if err != nil {
		return nil, err
	}
	return im.getItemAttrByID(worldID, itemTypeID, itemID, attrID)
}

func (im *ItemManager) getItemAttrIDByName(worldID, itemTypeID, itemID uint64, attrName string) (uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), attrName))
	if len(b) == 0 {
		return 0, ErrItemAttrNotExist
	}
	var attrID uint64
	err = rlp.DecodeBytes(b, &attrID)
	if err != nil {
		return 0, err
	}
	return attrID, nil
}

func (im *ItemManager) getItemAttrByID(worldID, itemTypeID, itemID, attrID uint64) (*Attribute, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), strconv.FormatUint(attrID, 10)))
	if len(b) == 0 {
		return nil, ErrItemAttrNotExist
	}
	var attrobj Attribute
	if err = rlp.DecodeBytes(b, &attrobj); err != nil {
		return nil, err
	}
	return &attrobj, nil
}

func (im *ItemManager) delItemAttrByName(worldID, itemTypeID, itemID uint64, attrName string) error {
	b, err := im.sdb.Get(itemManager, dbKey(itemAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), attrName))
	if len(b) == 0 {
		return ErrItemAttrNotExist
	}
	var attrID uint64
	err = rlp.DecodeBytes(b, &attrID)
	if err != nil {
		return err
	}
	im.sdb.Delete(itemManager, dbKey(itemAttrNamePrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), attrName))
	im.sdb.Delete(itemManager, dbKey(itemAttrIDPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemID, 10), strconv.FormatUint(attrID, 10)))
	return nil
}

func (im *ItemManager) modifyitemAttr(worldOwner, from string, worldID, itemTypeID, itemID uint64, attr *Attribute) error {
	attrID, err := im.getItemAttrIDByName(worldID, itemTypeID, itemID, attr.Name)
	if err != nil {
		return err
	}

	attrobj, err := im.getItemAttrByID(worldID, itemTypeID, itemID, attrID)
	if err != nil {
		return err
	}
	itemobj, err := im.getItemByID(worldID, itemTypeID, itemID)
	if err != nil {
		return err
	}

	switch attrobj.Permission {
	case CannotModify:
		return ErrNoPermission
	case WorldOwner:
		if worldOwner != from {
			return ErrNoPermission
		}
	case ItemOwner:
		if itemobj.Owner != from {
			return ErrNoPermission
		}
	}
	attrobj.Permission = attr.Permission
	attrobj.Description = attr.Description
	err = im.setItemAttr(worldID, itemTypeID, itemID, attrID, attrobj)
	if err != nil {
		return err
	}

	return nil
}

func (im *ItemManager) setItems(itemsobj *Items) error {
	if itemsobj == nil {
		return ErrItemsObjectEmpty
	}
	b, err := rlp.EncodeToBytes(itemsobj)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemsPrefix, strconv.FormatUint(itemsobj.WorldID, 10), strconv.FormatUint(itemsobj.TypeID, 10), itemsobj.Owner), b)
	return nil
}

func (im *ItemManager) getItemsByOwner(worldID, itemTypeID uint64, owner string) (*Items, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemsPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), owner))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrItemsNotExist
	}

	var items Items
	if err = rlp.DecodeBytes(b, &items); err != nil {
		return nil, err
	}
	return &items, nil
}

func (im *ItemManager) delItemsByOwner(worldID, itemTypeID uint64, owner string) error {
	im.sdb.Delete(itemManager, dbKey(itemsPrefix, strconv.FormatUint(worldID, 10), strconv.FormatUint(itemTypeID, 10), owner))
	return nil
}

func (im *ItemManager) addItemsAmount(account string, worldID, itemTypeID, amount uint64) error {
	itemsobj, err := im.getItemsByOwner(worldID, itemTypeID, account)
	if err != nil && err != ErrItemsNotExist {
		return err
	}

	if err == ErrItemsNotExist {
		itemsobj = &Items{
			WorldID: worldID,
			TypeID:  itemTypeID,
			Owner:   account,
			Amount:  amount,
		}
	} else {
		itemsobj.Amount += amount
	}
	err = im.setItems(itemsobj)
	if err != nil {
		return err
	}
	return nil
}

func (im *ItemManager) subItemsAmount(account string, worldID, itemTypeID, amount uint64) error {
	itemsobj, err := im.getItemsByOwner(worldID, itemTypeID, account)
	if err != nil {
		return err
	}
	if itemsobj.Amount < amount {
		return ErrInsufficientItemAmount
	}
	itemsobj.Amount -= amount
	if itemsobj.Amount == 0 {
		im.delItemsByOwner(worldID, itemTypeID, account)
		return nil
	}
	err = im.setItems(itemsobj)
	if err != nil {
		return err
	}
	return nil
}

// for RPC

func (im *ItemManager) GetWorldByID(worldID uint64) (*World, error) {
	return im.getWorldByID(worldID)
}

func (im *ItemManager) GetWorldByName(worldName string) (*World, error) {
	return im.getWorldByName(worldName)
}

func (im *ItemManager) GetItemTypeByID(worldID, itemTypeID uint64) (*ItemType, error) {
	return im.getItemTypeByID(worldID, itemTypeID)
}
func (im *ItemManager) GetItemTypeByName(worldID uint64, itemTypeName string) (*ItemType, error) {
	return im.getItemTypeByName(worldID, itemTypeName)
}

func (im *ItemManager) GetItemByID(worldID, itemTypeID, itemID uint64) (*Item, error) {
	return im.getItemByID(worldID, itemTypeID, itemID)
}

func (im *ItemManager) GetItemByOwner(worldID, itemTypeID uint64, owner string) ([]*Item, error) {
	return im.getItemByOwner(worldID, itemTypeID, owner)
}

func (im *ItemManager) GetItemsByOwner(worldID, itemTypeID uint64, account string) (*Items, error) {
	return im.getItemsByOwner(worldID, itemTypeID, account)
}

func (im *ItemManager) GetItemTypeAttributeByID(worldID, itemTypeID, attrID uint64) (*Attribute, error) {
	return im.getItemTypeAttrByID(worldID, itemTypeID, attrID)
}

func (im *ItemManager) GetItemTypeAttributeByName(worldID, itemTypeID uint64, attrName string) (*Attribute, error) {
	return im.getItemTypeAttrByName(worldID, itemTypeID, attrName)
}

func (im *ItemManager) GetItemAttributeByID(worldID, itemTypeID, itemID, attrID uint64) (*Attribute, error) {
	return im.getItemAttrByID(worldID, itemTypeID, itemID, attrID)
}
func (im *ItemManager) GetItemAttributeByName(worldID, itemTypeID, itemID uint64, attrName string) (*Attribute, error) {
	return im.getItemAttrByName(worldID, itemTypeID, itemID, attrName)
}

// for API

func (im *ItemManager) Sol_IssueWorld(context *ContextSol, owner, name, description string) error {
	_, err := im.IssueWorld(context.tx.Sender(), owner, name, description, context.pm)
	return err
}

func (im *ItemManager) Sol_UpdateWorldOwner(context *ContextSol, owner string, worldID uint64) error {
	_, err := im.UpdateWorldOwner(context.tx.Sender(), owner, worldID, context.pm)
	return err
}

func (im *ItemManager) Sol_IssueItemType(context *ContextSol, worldID uint64, name string, merge bool, upperLimit uint64, des string, attrPermission []uint64, attrName []string, attrDes []string) error {
	if len(attrPermission) != len(attrName) {
		return ErrParamErr
	}
	if len(attrPermission) != len(attrDes) {
		return ErrParamErr
	}
	attr := make([]*Attribute, len(attrPermission))
	for i := 0; i < len(attrPermission); i++ {
		temp := &Attribute{attrPermission[i], attrName[i], attrDes[i]}
		attr[i] = temp
	}
	_, err := im.IssueItemType(context.tx.Sender(), worldID, name, merge, upperLimit, des, attr, context.pm)
	return err
}

func (im *ItemManager) Sol_IncreaseItem(context *ContextSol, worldID uint64, itemTypeID uint64, owner, description string, attrPermission []uint64, attrName []string, attrDes []string) error {
	if len(attrPermission) != len(attrName) {
		return ErrParamErr
	}
	if len(attrPermission) != len(attrDes) {
		return ErrParamErr
	}
	attr := make([]*Attribute, len(attrPermission))
	for i := 0; i < len(attrPermission); i++ {
		temp := &Attribute{attrPermission[i], attrName[i], attrDes[i]}
		attr[i] = temp
	}
	_, err := im.IncreaseItem(context.tx.Sender(), worldID, itemTypeID, owner, description, attr, context.pm)
	return err
}

func (im *ItemManager) Sol_DestroyItem(context *ContextSol, worldID uint64, itemTypeID uint64, itemID uint64) error {
	_, err := im.DestroyItem(context.tx.Sender(), worldID, itemTypeID, itemID, context.pm)
	return err
}

func (im *ItemManager) Sol_IncreaseItems(context *ContextSol, worldID uint64, itemTypeID uint64, to string, amount uint64) error {
	_, err := im.IncreaseItems(context.tx.Sender(), worldID, itemTypeID, to, amount, context.pm)
	return err
}

func (im *ItemManager) Sol_DestroyItems(context *ContextSol, worldID uint64, itemTypeID uint64, amount uint64) error {
	_, err := im.DestroyItems(context.tx.Sender(), worldID, itemTypeID, amount, context.pm)
	return err
}

func (im *ItemManager) Sol_TransferItem(context *ContextSol, to string, worldID []uint64, itemTypeID []uint64, itemID []uint64, amount []uint64) error {
	if len(worldID) != len(itemTypeID) {
		return ErrParamErr
	}
	if len(worldID) != len(itemID) {
		return ErrParamErr
	}
	if len(worldID) != len(amount) {
		return ErrParamErr
	}
	itemTx := make([]*ItemTxParam, len(worldID))
	for i := 0; i < len(worldID); i++ {
		temp := &ItemTxParam{worldID[i], itemTypeID[i], itemID[i], amount[i]}
		itemTx[i] = temp
	}
	return im.TransferItem(context.tx.Sender(), to, itemTx, context.pm)
}

func (im *ItemManager) Sol_AddItemTypeAttributes(context *ContextSol, worldID uint64, itemTypeID uint64, attrPermission []uint64, attrName []string, attrDes []string) error {
	if len(attrPermission) != len(attrName) {
		return ErrParamErr
	}
	if len(attrPermission) != len(attrDes) {
		return ErrParamErr
	}
	attr := make([]*Attribute, len(attrPermission))
	for i := 0; i < len(attrPermission); i++ {
		temp := &Attribute{attrPermission[i], attrName[i], attrDes[i]}
		attr[i] = temp
	}
	_, err := im.AddItemTypeAttributes(context.tx.Sender(), worldID, itemTypeID, attr)
	return err
}

func (im *ItemManager) Sol_DelItemTypeAttributes(context *ContextSol, worldID uint64, itemTypeID uint64, attrName []string) error {
	_, err := im.DelItemTypeAttributes(context.tx.Sender(), worldID, itemTypeID, attrName)
	return err
}

func (im *ItemManager) Sol_ModifyItemTypeAttributes(context *ContextSol, worldID uint64, itemTypeID uint64, attrPermission []uint64, attrName []string, attrDes []string) error {
	if len(attrPermission) != len(attrName) {
		return ErrParamErr
	}
	if len(attrPermission) != len(attrDes) {
		return ErrParamErr
	}
	attr := make([]*Attribute, len(attrPermission))
	for i := 0; i < len(attrPermission); i++ {
		temp := &Attribute{attrPermission[i], attrName[i], attrDes[i]}
		attr[i] = temp
	}
	_, err := im.ModifyItemTypeAttributes(context.tx.Sender(), worldID, itemTypeID, attr)
	return err
}

func (im *ItemManager) Sol_AddItemAttributes(context *ContextSol, worldID uint64, itemTypeID uint64, itemID uint64, attrPermission []uint64, attrName []string, attrDes []string) error {
	if len(attrPermission) != len(attrName) {
		return ErrParamErr
	}
	if len(attrPermission) != len(attrDes) {
		return ErrParamErr
	}
	attr := make([]*Attribute, len(attrPermission))
	for i := 0; i < len(attrPermission); i++ {
		temp := &Attribute{attrPermission[i], attrName[i], attrDes[i]}
		attr[i] = temp
	}
	_, err := im.AddItemAttributes(context.tx.Sender(), worldID, itemTypeID, itemID, attr)
	return err
}

func (im *ItemManager) Sol_DelItemAttributes(context *ContextSol, worldID uint64, itemTypeID uint64, itemID uint64, attrName []string) error {
	_, err := im.DelItemAttributes(context.tx.Sender(), worldID, itemTypeID, itemID, attrName)
	return err
}

func (im *ItemManager) Sol_ModifyItemAttributes(context *ContextSol, worldID uint64, itemTypeID uint64, itemID uint64, attrPermission []uint64, attrName []string, attrDes []string) error {
	if len(attrPermission) != len(attrName) {
		return ErrParamErr
	}
	if len(attrPermission) != len(attrDes) {
		return ErrParamErr
	}
	attr := make([]*Attribute, len(attrPermission))
	for i := 0; i < len(attrPermission); i++ {
		temp := &Attribute{attrPermission[i], attrName[i], attrDes[i]}
		attr[i] = temp
	}
	_, err := im.ModifyItemAttributes(context.tx.Sender(), worldID, itemTypeID, itemID, attr)
	return err
}

type SolWorldInfo struct {
	ID          uint64
	Name        string
	Owner       common.Address
	Creator     common.Address
	Description string
	Total       uint64
}

func (im *ItemManager) Sol_GetWorldInfo(context *ContextSol, worldID uint64) (*SolWorldInfo, error) {
	worldobj, err := im.getWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	solWorld := &SolWorldInfo{
		ID:          worldobj.ID,
		Name:        worldobj.Name,
		Owner:       common.StringToAddress(worldobj.Owner),
		Creator:     common.StringToAddress(worldobj.Creator),
		Description: worldobj.Description,
		Total:       worldobj.Total,
	}
	return solWorld, nil
}

type SolItemType struct {
	WorldID     uint64
	ID          uint64
	Name        string
	Merge       bool
	UpperLimit  uint64
	AddIssue    uint64
	Description string
	Total       uint64
	AttrTotal   uint64
}

func (im *ItemManager) Sol_GetItemType(context *ContextSol, worldID uint64, itemTypeID uint64) (*SolItemType, error) {
	obj, err := im.getItemTypeByID(worldID, itemTypeID)
	if err != nil {
		return nil, err
	}
	solItemTypeobj := &SolItemType{
		WorldID:     obj.WorldID,
		ID:          obj.ID,
		Name:        obj.Name,
		Merge:       obj.Merge,
		UpperLimit:  obj.UpperLimit,
		AddIssue:    obj.AddIssue,
		Description: obj.Description,
		Total:       obj.Total,
		AttrTotal:   obj.AttrTotal,
	}
	return solItemTypeobj, nil
}

type SolItem struct {
	WorldID     uint64
	TypeID      uint64
	ID          uint64
	Owner       common.Address
	Description string
	Destroy     bool
	AttrTotal   uint64
}

func (im *ItemManager) Sol_GetItem(context *ContextSol, worldID uint64, itemTypeID uint64, itemID uint64) (*SolItem, error) {
	obj, err := im.getItemByID(worldID, itemTypeID, itemID)
	if err != nil {
		return nil, err
	}
	sobj := &SolItem{
		WorldID:     obj.WorldID,
		TypeID:      obj.TypeID,
		ID:          obj.ID,
		Owner:       common.StringToAddress(obj.Owner),
		Description: obj.Description,
		Destroy:     obj.Destroy,
		AttrTotal:   obj.AttrTotal,
	}
	return sobj, nil
}

type SolItems struct {
	WorldID uint64
	TypeID  uint64
	Owner   common.Address
	Amount  uint64
}

func (im *ItemManager) Sol_GetItems(context *ContextSol, worldID uint64, itemTypeID uint64, owner string) (*SolItems, error) {
	o, err := im.getItemsByOwner(worldID, itemTypeID, owner)
	if err != nil {
		return nil, err
	}
	so := &SolItems{
		WorldID: o.WorldID,
		TypeID:  o.TypeID,
		Owner:   common.StringToAddress(owner),
		Amount:  o.Amount,
	}
	return so, nil
}

var (
	ErrWorldCounterNotExist    = errors.New("item global counter not exist")
	ErrItemNameinvalid         = errors.New("item name invalid")
	ErrWorldNameinvalid        = errors.New("world name invalid")
	ErrItemNameLengthErr       = errors.New("item name length err")
	ErrWorldNameNotExist       = errors.New("WorldName not exist")
	ErrWorldNameIsExist        = errors.New("WorldName is exist")
	ErrItemTypeNameNotExist    = errors.New("itemTypeName not exist")
	ErrItemTypeNameIsExist     = errors.New("itemTypeName is exist")
	ErrItemInfoNameNotExist    = errors.New("itemInfoName not exist")
	ErrItemInfoNameIsExist     = errors.New("itemInfoName is exist")
	ErrWorldNotExist           = errors.New("world not exist")
	ErrWorldIsExist            = errors.New("world is exist")
	ErrItemTypeNotExist        = errors.New("itemType not exist")
	ErrItemTypeIsExist         = errors.New("itemType is exist")
	ErrItemNotExist            = errors.New("item not exist")
	ErrItemIsExist             = errors.New("item is exist")
	ErrItemsNotExist           = errors.New("items not exist")
	ErrItemsIsExist            = errors.New("items is exist")
	ErrWorldObjectEmpty        = errors.New("world object is empty")
	ErrItemTypeObjectEmpty     = errors.New("itemType object is empty")
	ErrItemObjectEmpty         = errors.New("item object is empty")
	ErrItemsObjectEmpty        = errors.New("items object is empty")
	ErrItemOwnerMismatch       = errors.New("owner mismatch")
	ErrItemAttributeNameIsNull = errors.New("attribute name is null")
	ErrItemAttributeDesTooLong = errors.New("attribute description exceed max length")
	ErrItemUpperLimit          = errors.New("amount over the issuance limit")
	ErrAccountNoItem           = errors.New("account not have item")
	ErrParamErr                = errors.New("param invalid")
	ErrInsufficientItemAmount  = errors.New("insufficient amount")
	ErrInvalidItemAmount       = errors.New("invalid amount")
	ErrInvalidItemID           = errors.New("invalid id")
	ErrItemTypeMergeIsFalse    = errors.New("itemType merge is false")
	ErrItemTypeMergeIsTrue     = errors.New("itemType merge is true")
	ErrItemIsDestroyed         = errors.New("item is destroyed")
	ErrItemTypeAttrNotExist    = errors.New("itemType attribute not exist")
	ErrItemTypeAttrIsExist     = errors.New("itemType attribute is exist")
	ErrItemAttrNotExist        = errors.New("item attribute not exist")
	ErrItemAttrIsExist         = errors.New("item attribute is exist")
	ErrNoPermission            = errors.New("no permission to modify")
	ErrInvalidPermission       = errors.New("invalid permission")
	ErrDuplicateAttr           = errors.New("duplicate attribute name")
	ErrExceedMax               = errors.New("exceed max value")
)
