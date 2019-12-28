package plugin

import (
	"testing"
)

func Test_IssueWorld(t *testing.T) {
	type args struct {
		creator string
		owner   string
		name    string
		des     string
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"IssueWorld", args{"worldowner", "worldowner", "ab0123456789", "just for test"}, nil},
		{"IssueWorld2", args{"worldowner", "worldowner", "ab1123456789", "just for test"}, nil},
		{"IssueWorld3", args{"worldowner", "worldowner", "ab0123456789", "just for test"}, ErrWorldNameIsExist},
	}

	for _, item := range testItem {
		if _, err := pm.IssueWorld(item.arg.creator, item.arg.owner, item.arg.name, item.arg.des, pm); err != item.err {
			t.Errorf("%q. IssueItemType() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_IssueItemType(t *testing.T) {
	a := []*Attribute{&Attribute{0, "typeshudu", "100"}, &Attribute{1, "typelilang", "50"}}
	type args struct {
		creator    string
		worldID    uint64
		name       string
		merge      bool
		upperLimit uint64
		des        string
		attributes []*Attribute
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"IssueItemType", args{"worldowner", uint64(1), "tulongdao", false, uint64(10), "itemType1 tulongdao", a}, nil},
		{"IssueItemType2", args{"worldowner", uint64(1), "tulongdao2", false, uint64(5), "itemType2 tulongdao2", a}, nil},
		{"IssueItemType3", args{"worldowner", uint64(1), "tulongdao2", false, uint64(5), "itemType2 tulongdao2", a}, ErrItemTypeNameIsExist},
		{"IssueItemType4", args{"worldowner", uint64(1), "xueyao", true, uint64(0), "itemType1 xueyao", nil}, nil},
	}
	for _, item := range testItem {
		if _, err := pm.IssueItemType(item.arg.creator, item.arg.worldID, item.arg.name, item.arg.merge, item.arg.upperLimit, item.arg.des, item.arg.attributes, pm); err != item.err {
			t.Errorf("%q. IssueItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_IncreseItem(t *testing.T) {
	a := []*Attribute{&Attribute{0, "a11item", "不能修改"}, &Attribute{1, "a11item", "worldOwner 能修改"}, &Attribute{1, "a11item", "自己能修改"}}
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		owner      string
		des        string
		attributes []*Attribute
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"IncreaseItem1", args{"worldowner", uint64(1), uint64(1), "type1item111", "1_1_tulongdao1", a}, nil},
		{"IncreaseItem2", args{"worldowner", uint64(1), uint64(1), "type1item222", "1_1_tulongdao2", a}, nil},
	}
	for _, item := range testItem {
		if _, err := pm.IncreaseItem(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.owner, item.arg.des, item.arg.attributes, pm); err != item.err {
			t.Errorf("%q. IncreaseItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_DestroyItem(t *testing.T) {
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		itemID     uint64
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		// {"DestroyItem1", args{"type1item111", uint64(1), uint64(1), uint64(1)}, nil},
		{"DestroyItem1", args{"type1item111", uint64(1), uint64(1), uint64(2)}, ErrItemOwnerMismatch},
		// {"DestroyItem1", args{"type1item111", uint64(1), uint64(1), uint64(1)}, ErrItemIsDestroyed},
	}
	for _, item := range testItem {
		if _, err := pm.DestroyItem(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.itemID, pm); err != item.err {
			t.Errorf("%q. IssueItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}
func Test_IncreaseItems(t *testing.T) {
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		to         string
		amount     uint64
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"IncreaseItems1", args{"worldowner", uint64(1), uint64(3), "type1items333", uint64(15)}, nil},
		{"IncreaseItems2", args{"worldowner", uint64(1), uint64(1), "type1items333", uint64(15)}, ErrItemTypeMergeIsFalse},
		{"IncreaseItems3", args{"worldowner", uint64(1), uint64(3), "type1items444", uint64(20)}, nil},
	}
	for _, item := range testItem {
		if _, err := pm.IncreaseItems(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.to, item.arg.amount, pm); err != item.err {
			t.Errorf("%q. IssueItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_DestroyItems(t *testing.T) {
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		amount     uint64
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"DestroyItems1", args{"type1items333", uint64(1), uint64(3), uint64(5)}, nil},
		{"DestroyItems2", args{"type1items444", uint64(1), uint64(3), uint64(10)}, nil},
		{"DestroyItems3", args{"type1items333", uint64(1), uint64(3), uint64(150)}, ErrInsufficientItemAmount},
	}
	for _, item := range testItem {
		if _, err := pm.DestroyItems(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.amount, pm); err != item.err {
			t.Errorf("%q. IssueItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_TransferItem(t *testing.T) {
	tx1 := []*ItemTxParam{&ItemTxParam{uint64(1), uint64(1), uint64(1), uint64(1)}}
	tx2 := []*ItemTxParam{&ItemTxParam{uint64(1), uint64(3), uint64(0), uint64(15)}}
	tx3 := []*ItemTxParam{&ItemTxParam{uint64(1), uint64(3), uint64(0), uint64(10)}}
	// tx3 := []*ItemTxParam{&ItemTxParam{uint64(1), uint64(1), uint64(3), uint64(100)}}
	type args struct {
		from   string
		to     string
		itemTx []*ItemTxParam
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"TransferItem1", args{"type1item111", "itemtest1", tx1}, nil},
		{"TransferItem2", args{"type1items333", "itemtest2", tx2}, ErrInsufficientItemAmount},
		{"TransferItem2", args{"type1items333", "itemtest2", tx3}, nil},
		// {"TransferItem", args{"itemtype1", "itemtest2", tx3}, ErrAccountNoItem},
		// {"TransferItem", args{"itemtype1", "itemtest2", tx5}, ErrAccountNoItem},
	}
	for _, item := range testItem {
		if err := pm.TransferItem(item.arg.from, item.arg.to, item.arg.itemTx, pm); err != item.err {
			t.Errorf("%q. TransferItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_AddItemTypeAttributes(t *testing.T) {
	a := []*Attribute{&Attribute{0, "a11item1", "不能修改"}, &Attribute{1, "a11item2", "worldOwner 能修改"}, &Attribute{1, "a11item3", "自己能修改"}}
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		attributes []*Attribute
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"AddItemTypeAttributes1", args{"type1item111", uint64(1), uint64(1), a}, ErrItemOwnerMismatch},
		{"AddItemTypeAttributes2", args{"worldowner", uint64(1), uint64(1), a}, nil},
		{"AddItemTypeAttributes3", args{"worldowner", uint64(1), uint64(10), a}, ErrItemTypeNotExist},
	}
	for _, item := range testItem {
		if _, err := pm.AddItemTypeAttributes(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.attributes); err != item.err {
			t.Errorf("%q. AddItemTypeAttributes() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_DelItemTypeAttributes(t *testing.T) {
	a := []string{"a11item1"}
	a2 := []string{"a11item4"}
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		attrName   []string
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"DelItemTypeAttributes1", args{"type1item111", uint64(1), uint64(1), a}, ErrItemOwnerMismatch},
		{"DelItemTypeAttributes2", args{"worldowner", uint64(1), uint64(1), a}, nil},
		{"DelItemTypeAttributes3", args{"worldowner", uint64(1), uint64(1), a2}, ErrItemTypeAttrNotExist},
	}
	for _, item := range testItem {
		if _, err := pm.DelItemTypeAttributes(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.attrName); err != item.err {
			t.Errorf("%q. DelItemTypeAttributes() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_ModifyItemTypeAttributes(t *testing.T) {
	a := []*Attribute{&Attribute{0, "a11item1", "已修改"}}
	a2 := []*Attribute{&Attribute{0, "a11item4", "已修改"}}
	a3 := []*Attribute{&Attribute{2, "a11item4", "已修改"}}
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		attributes []*Attribute
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"ModifyItemTypeAttributes1", args{"type1item111", uint64(1), uint64(1), a}, ErrItemOwnerMismatch},
		{"ModifyItemTypeAttributes2", args{"worldowner", uint64(1), uint64(1), a}, nil},
		{"ModifyItemTypeAttributes3", args{"worldowner", uint64(1), uint64(1), a2}, ErrItemTypeAttrNotExist},
		{"ModifyItemTypeAttributes3", args{"worldowner", uint64(1), uint64(1), a3}, ErrInvalidPermission},
	}
	for _, item := range testItem {
		if _, err := pm.ModifyItemTypeAttributes(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.attributes); err != item.err {
			t.Errorf("%q. DelItemTypeAttributes() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_AddItemAttributes(t *testing.T) {
	a := []*Attribute{&Attribute{0, "a11item1", "不能修改"}, &Attribute{1, "a11item2", "worldOwner 能修改"}, &Attribute{1, "a11item3", "自己能修改"}}
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		itemID     uint64
		attributes []*Attribute
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"AddItemAttributes1", args{"type1item111", uint64(1), uint64(1), uint64(1), a}, ErrItemOwnerMismatch},
		{"AddItemAttributes2", args{"worldowner", uint64(1), uint64(1), uint64(1), a}, nil},
		{"AddItemAttributes3", args{"worldowner", uint64(1), uint64(10), uint64(1), a}, ErrItemTypeNotExist},
	}
	for _, item := range testItem {
		if _, err := pm.AddItemAttributes(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.itemID, item.arg.attributes); err != item.err {
			t.Errorf("%q. AddItemAttributes() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_DelItemAttributes(t *testing.T) {
	a := []string{"a11item1"}
	a2 := []string{"a11item4"}
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		itemID     uint64
		attrName   []string
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"DelItemAttributes1", args{"type1item111", uint64(1), uint64(1), uint64(1), a}, ErrItemOwnerMismatch},
		{"DelItemAttributes2", args{"worldowner", uint64(1), uint64(1), uint64(1), a}, nil},
		{"DelItemAttributes3", args{"worldowner", uint64(1), uint64(1), uint64(1), a2}, ErrItemTypeAttrNotExist},
	}
	for _, item := range testItem {
		if _, err := pm.DelItemAttributes(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.itemID, item.arg.attrName); err != item.err {
			t.Errorf("%q. DelItemAttributes() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_ModifyItemAttributes(t *testing.T) {
	a := []*Attribute{&Attribute{0, "a11item1", "已修改"}}
	a2 := []*Attribute{&Attribute{0, "a11item4", "已修改"}}
	a3 := []*Attribute{&Attribute{2, "a11item4", "已修改"}}
	type args struct {
		from       string
		worldID    uint64
		itemTypeID uint64
		itemID     uint64
		attributes []*Attribute
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"ModifyItemAttributes1", args{"type1item111", uint64(1), uint64(1), uint64(1), a}, ErrItemOwnerMismatch},
		{"ModifyItemAttributes2", args{"worldowner", uint64(1), uint64(1), uint64(1), a}, nil},
		{"ModifyItemAttributes3", args{"worldowner", uint64(1), uint64(1), uint64(1), a2}, ErrItemTypeAttrNotExist},
		{"ModifyItemAttributes4", args{"worldowner", uint64(1), uint64(1), uint64(1), a3}, ErrInvalidPermission},
	}
	for _, item := range testItem {
		if _, err := pm.ModifyItemAttributes(item.arg.from, item.arg.worldID, item.arg.itemTypeID, item.arg.itemID, item.arg.attributes); err != item.err {
			t.Errorf("%q. DelItemAttributes() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}
