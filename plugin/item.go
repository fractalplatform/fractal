package plugin

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	itemRegExp              = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9]{1,31})$`)
	itemNameMaxLength       = uint64(32)
	itemManager             = "item"
	counterPrefix           = "itemTypeCounter"
	accountItemAmountPrefix = "accuntItemAmount"
	itemTypePrefix          = "itemType"
	itemInfoPrefix          = "itemInfo"
	itemTypeNamePrefix      = "itemTyepName"
	itemInfoNamePrefix      = "itemInfoName"
)

const UINT64_MAX uint64 = ^uint64(0)

type ItemManager struct {
	sdb *state.StateDB
}

type ItemType struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Creator     string `json:"creator"`
	CreateTime  uint64 `json:"createTime"`
	Description string `json:"description"`
	Total       uint64 `json:"total"`
}

type ItemInfo struct {
	TypeID      uint64       `json:"typeID"`
	ID          uint64       `json:"id"`
	Name        string       `json:"name"`
	CreateTime  uint64       `json:"createTime"`
	Total       uint64       `json:"total"`
	Description string       `json:"description"`
	UpperLimit  uint64       `json:"upperLimit"`
	Attributes  []*Attribute `json:"attributes"`
}

// NewIM new a ItemManager
func NewItemManage(sdb *state.StateDB) (*ItemManager, error) {
	if sdb == nil {
		return nil, ErrNewAssetManagerErr
	}

	itemManager := ItemManager{
		sdb: sdb,
	}
	itemManager.initItemCounter()
	return &itemManager, nil
}

func (im *ItemManager) AccountName() string {
	return "fractalitem"
}

func (im *ItemManager) CallTx(tx *envelope.PluginTx, pm IPM) ([]byte, error) {
	switch tx.PayloadType() {
	case IssueItemType:
		param := &IssueItemTypeAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.IssueItemType(tx.Sender(), param.Owner, param.Name, param.Description, pm)
	case UpdateItemTypeOwner:
		param := &UpdateItemTypeOwnerAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.UpdateItemTypeOwner(tx.Sender(), param.NewOwner, param.ItemTypeID, pm)
	case IssueItem:
		param := &IssueItemAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.IssueItem(tx.Sender(), param.ItemTypeID, param.Name, param.Description, param.UpperLimit, param.Total, param.Attributes, pm)
	case IncreaseItem:
		param := &IncreaseItemAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		return im.IncreaseItem(tx.Sender(), param.ItemTypeID, param.ItemInfoID, param.To, param.Amount, pm)
	case TransferItem:
		param := &TransferItemAction{}
		if err := rlp.DecodeBytes(tx.GetPayload(), param); err != nil {
			return nil, err
		}
		err := im.TransferItem(tx.Sender(), param.To, param.ItemTx)
		return nil, err
	}
	return nil, ErrWrongTransaction
}

// IssueItemType issue itemType
func (im *ItemManager) IssueItemType(creator, owner, name, description string, am IAccount) ([]byte, error) {
	if err := im.checkItemNameFormat(name); err != nil {
		return nil, err
	}
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}

	if _, err := am.getAccount(creator); err != nil {
		return nil, err
	}
	if _, err := am.getAccount(owner); err != nil {
		return nil, err
	}

	_, err := im.getItemTypeIDByName(creator, name)
	if err == nil {
		return nil, ErrItemTypeNameIsExist
	} else if err != ErrItemTypeNameNotExist {
		return nil, err
	}

	itemTypeID, err := im.getItemCounter()
	if err != nil {
		return nil, err
	}

	itemobj := ItemType{
		ID:          itemTypeID,
		Name:        name,
		Owner:       owner,
		Creator:     creator,
		CreateTime:  uint64(time.Now().Unix()),
		Description: description,
		Total:       0,
	}
	err = im.setNewItemType(&itemobj)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// UpdateItemTypeOwner update itemType owner
func (im *ItemManager) UpdateItemTypeOwner(from, newOwner string, itemTypeID uint64, am IAccount) ([]byte, error) {
	if _, err := am.getAccount(newOwner); err != nil {
		return nil, err
	}
	item, err := im.getItemTypeByID(itemTypeID)
	if err != nil {
		return nil, err
	}
	if from != item.Owner {
		return nil, ErrItemOwnerMismatch
	}
	item.Owner = newOwner

	err = im.setItemType(item)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// IssueItem issue new item
func (im *ItemManager) IssueItem(creator string, itemTypeID uint64, name string, description string, upperLimit uint64, total uint64, attributes []*Attribute, am IAccount) ([]byte, error) {
	if err := im.checkItemNameFormat(name); err != nil {
		return nil, err
	}
	if uint64(len(description)) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}
	if upperLimit > UINT64_MAX || total > UINT64_MAX {
		return nil, ErrAmountValueInvalid
	}
	if upperLimit != 0 {
		if total > upperLimit {
			return nil, ErrAmountValueInvalid
		}
	}
	for _, att := range attributes {
		if uint64(len(att.Name)) > MaxDescriptionLength || uint64(len(att.Description)) > MaxDescriptionLength {
			return nil, ErrItemAttributeDesTooLong
		}
	}

	if _, err := am.getAccount(creator); err != nil {
		return nil, err
	}

	itemType, err := im.getItemTypeByID(itemTypeID)
	if err != nil {
		return nil, err
	}

	if itemType.Owner != creator {
		return nil, ErrItemOwnerMismatch
	}

	_, err = im.getItemInfoIDByName(itemTypeID, name)
	if err == nil {
		return nil, ErrItemInfoNameIsExist
	} else if err != ErrItemInfoNameNotExist {
		return nil, err
	}

	itemInfo := ItemInfo{
		TypeID:      itemTypeID,
		ID:          itemType.Total + 1,
		Name:        name,
		CreateTime:  uint64(time.Now().Unix()),
		Total:       total,
		Description: description,
		UpperLimit:  upperLimit,
		Attributes:  attributes,
	}
	itemType.Total += 1
	err = im.setItemType(itemType)
	if err != nil {
		return nil, err
	}
	err = im.setNewItemInfo(&itemInfo)
	if err != nil {
		return nil, err
	}
	err = im.setAccountItemAmount(creator, itemTypeID, itemInfo.ID, itemInfo.Total)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// IncreaseItem increase item
func (im *ItemManager) IncreaseItem(from string, itemTypeID, itemInfoID uint64, to string, amount uint64, am IAccount) ([]byte, error) {
	itemInfo, err := im.getItemInfoByID(itemTypeID, itemInfoID)
	if err != nil {
		return nil, err
	}
	itemType, err := im.getItemTypeByID(itemTypeID)
	if err != nil {
		return nil, err
	}

	if itemType.Creator != from {
		return nil, ErrItemOwnerMismatch
	}

	if _, err := am.getAccount(to); err != nil {
		return nil, err
	}

	if itemInfo.UpperLimit > 0 {
		if amount+itemInfo.Total > itemInfo.UpperLimit {
			return nil, ErrItemUpperLimit
		}
	}

	itemInfo.Total += amount
	if itemInfo.Total > UINT64_MAX {
		return nil, ErrAmountValueInvalid
	}
	err = im.setItemInfo(itemInfo)
	if err != nil {
		return nil, err
	}
	err = im.addAccountItemAmount(to, itemTypeID, itemInfoID, amount)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (im *ItemManager) transferItemSingle(from, to string, itemTypeID, itemInfoID, amount uint64) error {
	n, err := im.getAccountItemAmount(from, itemTypeID, itemInfoID)
	if err != nil {
		return err
	}

	if n < amount {
		return ErrInsufficientItemAmount
	}

	if err = im.subAccountItemAmount(from, itemTypeID, itemInfoID, amount); err != nil {
		return err
	}

	if err = im.addAccountItemAmount(to, itemTypeID, itemInfoID, amount); err != nil {
		return err
	}

	return nil
}

// TransferItem transfer item
func (im *ItemManager) TransferItem(from, to string, ItemTx []*ItemTxParam) error {
	for _, tx := range ItemTx {
		if err := im.transferItemSingle(from, to, tx.ItemTypeID, tx.ItemInfoID, tx.Amount); err != nil {
			return err
		}
	}
	return nil
}

// GetItemAmount get account item amount
func (im *ItemManager) GetItemAmount(account string, itemTypeID, itemInfoID uint64) (uint64, error) {
	return im.getAccountItemAmount(account, itemTypeID, itemInfoID)
}

// GetItemTypeOwner get itemType owner
func (im *ItemManager) GetItemTypeOwner(itemTypeID uint64) (string, error) {
	obj, err := im.getItemTypeByID(itemTypeID)
	if err != nil {
		return "", err
	}
	return obj.Owner, nil
}

// GetItemAttribute get item attribute
func (im *ItemManager) GetItemAttribute(itemTypeID uint64, itemInfoID uint64, AttributeName string) (string, error) {
	obj, err := im.getItemInfoByID(itemTypeID, itemInfoID)
	if err != nil {
		return "", err
	}

	for _, att := range obj.Attributes {
		if att.Name == AttributeName {
			return att.Description, nil
		}
	}

	return "", fmt.Errorf("%s attribute not found", AttributeName)
}

func (im *ItemManager) initItemCounter() {
	_, err := im.getItemCounter()
	if err == ErrItemCounterNotExist {
		var itemTypeID uint64 = 1
		b, err := rlp.EncodeToBytes(&itemTypeID)
		if err != nil {
			panic(err)
		}
		im.sdb.Put(itemManager, counterPrefix, b)
	}
}

func (im *ItemManager) getItemCounter() (uint64, error) {
	b, err := im.sdb.Get(itemManager, counterPrefix)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, ErrItemCounterNotExist
	}
	var itemCounter uint64
	err = rlp.DecodeBytes(b, &itemCounter)
	if err != nil {
		return 0, err
	}
	return itemCounter, nil
}

func (im *ItemManager) checkItemNameFormat(name string) error {
	if uint64(len(name)) > itemNameMaxLength {
		return ErrAssetNameLengthErr
	}

	if itemRegExp.MatchString(name) != true {
		return ErrItemNameinvalid
	}
	return nil
}

func (im *ItemManager) getItemTypeIDByName(creator, name string) (uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemTypeNamePrefix, creator, name))
	if len(b) == 0 {
		return 0, ErrItemTypeNameNotExist
	}
	var itemTypeID uint64
	if err = rlp.DecodeBytes(b, &itemTypeID); err != nil {
		return 0, err
	}
	return itemTypeID, nil
}

func (im *ItemManager) getItemInfoIDByName(itemTypeID uint64, name string) (uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemInfoNamePrefix, strconv.FormatUint(itemTypeID, 10), name))
	if len(b) == 0 {
		return 0, ErrItemInfoNameNotExist
	}
	var itemInfoID uint64
	if err = rlp.DecodeBytes(b, &itemInfoID); err != nil {
		return 0, err
	}
	return itemInfoID, nil
}

func (im *ItemManager) setNewItemType(itemobj *ItemType) error {
	if itemobj == nil {
		return ErrItemObjectEmpty
	}
	itemTypeID := itemobj.ID

	b, err := rlp.EncodeToBytes(itemobj)
	if err != nil {
		return err
	}
	id, err := rlp.EncodeToBytes(&itemTypeID)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemTypePrefix, strconv.FormatUint(itemTypeID, 10)), b)
	im.sdb.Put(itemManager, dbKey(itemTypeNamePrefix, itemobj.Creator, itemobj.Name), id)

	itemTypeID += 1
	nid, err := rlp.EncodeToBytes(&itemTypeID)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, counterPrefix, nid)
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
	im.sdb.Put(itemManager, dbKey(itemTypePrefix, strconv.FormatUint(itemobj.ID, 10)), b)
	return nil
}

func (im *ItemManager) setNewItemInfo(itemobj *ItemInfo) error {
	if itemobj == nil {
		return ErrItemObjectEmpty
	}
	id := itemobj.ID
	b, err := rlp.EncodeToBytes(itemobj)
	if err != nil {
		return err
	}
	nid, err := rlp.EncodeToBytes(&id)
	if err != nil {
		return err
	}

	im.sdb.Put(itemManager, dbKey(itemInfoPrefix, strconv.FormatUint(itemobj.TypeID, 10), strconv.FormatUint(id, 10)), b)
	im.sdb.Put(itemManager, dbKey(itemInfoNamePrefix, strconv.FormatUint(itemobj.TypeID, 10), itemobj.Name), nid)
	return nil
}

func (im *ItemManager) setItemInfo(itemobj *ItemInfo) error {
	if itemobj == nil {
		return ErrItemObjectEmpty
	}
	b, err := rlp.EncodeToBytes(itemobj)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(itemInfoPrefix, strconv.FormatUint(itemobj.TypeID, 10), strconv.FormatUint(itemobj.ID, 10)), b)
	return nil
}

func (im *ItemManager) getItemTypeByID(itemTypeID uint64) (*ItemType, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemTypePrefix, strconv.FormatUint(itemTypeID, 10)))
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

func (im *ItemManager) getItemInfoByID(itemTypeID, itemInfoID uint64) (*ItemInfo, error) {
	b, err := im.sdb.Get(itemManager, dbKey(itemInfoPrefix, strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemInfoID, 10)))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, ErrItemInfoNotExist
	}

	var itemInfo ItemInfo
	if err = rlp.DecodeBytes(b, &itemInfo); err != nil {
		return nil, err
	}
	return &itemInfo, nil
}

func (im *ItemManager) setAccountItemAmount(account string, itemTypeID, itemInfoID, amount uint64) error {
	b, err := rlp.EncodeToBytes(&amount)
	if err != nil {
		return err
	}
	im.sdb.Put(itemManager, dbKey(accountItemAmountPrefix, strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemInfoID, 10), account), b)
	return nil
}

func (im *ItemManager) getAccountItemAmount(account string, itemTypeID, itemInfoID uint64) (uint64, error) {
	b, err := im.sdb.Get(itemManager, dbKey(accountItemAmountPrefix, strconv.FormatUint(itemTypeID, 10), strconv.FormatUint(itemInfoID, 10), account))
	if len(b) == 0 {
		return 0, ErrAccountNoItem
	}
	var amount uint64
	if err = rlp.DecodeBytes(b, &amount); err != nil {
		return 0, err
	}
	return amount, nil
}

func (im *ItemManager) addAccountItemAmount(account string, itemTypeID, itemInfoID, amount uint64) error {
	oldAmount, err := im.getAccountItemAmount(account, itemTypeID, itemInfoID)
	if err != nil && err != ErrAccountNoItem {
		return err
	}

	if err = im.setAccountItemAmount(account, itemTypeID, itemInfoID, amount+oldAmount); err != nil {
		return err
	}
	return nil
}

func (im *ItemManager) subAccountItemAmount(account string, itemTypeID, itemInfoID, amount uint64) error {
	oldAmount, err := im.getAccountItemAmount(account, itemTypeID, itemInfoID)
	if err != nil && err != ErrAccountNoItem {
		return err
	}
	if oldAmount < amount {
		return ErrInsufficientItemAmount
	}

	if err = im.setAccountItemAmount(account, itemTypeID, itemInfoID, oldAmount-amount); err != nil {
		return err
	}
	return nil
}

func dbKey(s ...string) string {
	return strings.Join(s, "_")
}

func (im *ItemManager) GetItemTypeByID(itemTypeID uint64) (*ItemType, error) {
	return im.getItemTypeByID(itemTypeID)
}

func (im *ItemManager) GetItemTypeByName(creator string, itemTypeName string) (*ItemType, error) {
	id, err := im.getItemTypeIDByName(creator, itemTypeName)
	if err != nil {
		return nil, err
	}
	return im.getItemTypeByID(id)
}

func (im *ItemManager) GetItemInfoByID(itemTypeID, itemInfoID uint64) (*ItemInfo, error) {
	return im.getItemInfoByID(itemTypeID, itemInfoID)
}

func (im *ItemManager) GetItemInfoByName(itemTypeID uint64, itemInfoName string) (*ItemInfo, error) {
	id, err := im.getItemInfoIDByName(itemTypeID, itemInfoName)
	if err != nil {
		return nil, err
	}
	return im.getItemInfoByID(itemTypeID, id)
}

func (im *ItemManager) Sol_IssueItemType(context *ContextSol, owner, name, description string) error {
	_, err := im.IssueItemType(context.tx.Sender(), owner, name, description, context.pm)
	return err
}

func (im *ItemManager) Sol_IssueItem(context *ContextSol, itemTypeID uint64, name string, description string, upperLimit uint64, total uint64, attName []string, attDes []string) error {
	if len(attName) != len(attDes) {
		return ErrParamErr
	}
	attributes := make([]*Attribute, len(attName))
	for i := 0; i < len(attName); i++ {
		temp := &Attribute{attName[i], attDes[i]}
		attributes[i] = temp
	}
	_, err := im.IssueItem(context.tx.Sender(), itemTypeID, name, description, upperLimit, total, attributes, context.pm)
	return err
}

func (im *ItemManager) Sol_IncreaseItem(context *ContextSol, itemTypeID uint64, itemInfoID uint64, amount uint64) error {
	_, err := im.IncreaseItem(context.tx.Sender(), itemTypeID, itemInfoID, context.tx.Recipient(), amount, context.pm)
	return err
}

func (im *ItemManager) Sol_TransferItem(context *ContextSol, itemTypeID []uint64, itemInfoID []uint64, amount []uint64) error {
	if len(itemTypeID) != len(itemInfoID) {
		return ErrParamErr
	}
	if len(itemTypeID) != len(amount) {
		return ErrParamErr
	}
	ItemTx := make([]*ItemTxParam, len(itemTypeID))
	for i := 0; i < len(itemTypeID); i++ {
		temp := &ItemTxParam{itemTypeID[i], itemInfoID[i], amount[i]}
		ItemTx[i] = temp
	}
	return im.TransferItem(context.tx.Sender(), context.tx.Recipient(), ItemTx)
}

func (im *ItemManager) Sol_GetItemAmount(context *ContextSol, itemTypeID, itemInfoID uint64) (uint64, error) {
	return im.GetItemAmount(context.tx.Sender(), itemTypeID, itemInfoID)
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
	ErrParamErr                = errors.New("param invalid")
	ErrInsufficientItemAmount  = errors.New("insufficient item amount")
)
