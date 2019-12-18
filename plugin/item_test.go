package plugin

import (
	"testing"
)

func Test_IssueItemType(t *testing.T) {
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
		{"IssueItemType", args{"itemtype1", "itemtype1", "book", "just for test"}, nil},
		{"IssueItemType", args{"itemtype2", "itemtype2", "car", "just for test"}, nil},
		{"IssueItemType", args{"itemtype3", "itemtype3", "pen", "just for test"}, nil},
		{"IssueItemType", args{"itemtype3", "itemtype3", "dog", "just for test"}, ErrItemTypeNameIsExist},
	}

	for _, item := range testItem {
		if _, err := pm.IssueItemType(item.arg.creator, item.arg.owner, item.arg.name, item.arg.des, pm); err != item.err {
			t.Errorf("%q. IssueItemType() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_IssueItem(t *testing.T) {
	a := []*Attribute{&Attribute{"shudu", "100"}, &Attribute{"lilang", "50"}}
	type args struct {
		creator    string
		itemTypeID uint64
		name       string
		des        string
		upperList  uint64
		total      uint64
		attributes []*Attribute
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"IssueItem", args{"itemtype1", uint64(1), "math", "just for test", uint64(10), uint64(5), a}, nil},
		{"IssueItem", args{"itemtype1", uint64(1), "english", "just for test", uint64(8), uint64(5), a}, nil},
		{"IssueItem", args{"itemtype1", uint64(1), "physics", "just for test", uint64(6), uint64(5), a}, nil},
		{"IssueItem", args{"itemtype1", uint64(1), "physics", "just for test", uint64(6), uint64(5), a}, ErrItemInfoNameIsExist},
		{"IssueItem", args{"itemtype3", uint64(1), "chemistry", "just for test", uint64(6), uint64(5), a}, ErrItemOwnerMismatch},
	}
	for _, item := range testItem {
		if _, err := pm.IssueItem(item.arg.creator, item.arg.itemTypeID, item.arg.name, item.arg.des, item.arg.upperList, item.arg.total, item.arg.attributes, pm); err != item.err {
			t.Errorf("%q. IssueItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}
func Test_IncreaseItem(t *testing.T) {
	type args struct {
		from       string
		itemTypeID uint64
		itemInfoID uint64
		to         string
		amount     uint64
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"IssueItem", args{"itemtype1", uint64(1), uint64(1), "itemtest1", uint64(5)}, nil},
		{"IssueItem", args{"itemtype1", uint64(1), uint64(1), "itemtest1", uint64(5)}, ErrItemUpperLimit},
		{"IssueItem", args{"itemtype3", uint64(1), uint64(1), "itemtest1", uint64(5)}, ErrItemOwnerMismatch},
	}

	for _, item := range testItem {
		if _, err := pm.IncreaseItem(item.arg.from, item.arg.itemTypeID, item.arg.itemInfoID, item.arg.to, item.arg.amount, pm); err != item.err {
			t.Errorf("%q. IncreaseItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_TransferItem(t *testing.T) {
	tx1 := []*ItemTxParam{&ItemTxParam{uint64(1), uint64(1), uint64(2)}}
	tx2 := []*ItemTxParam{&ItemTxParam{uint64(1), uint64(1), uint64(2)}, &ItemTxParam{uint64(1), uint64(2), uint64(2)}}
	tx3 := []*ItemTxParam{&ItemTxParam{uint64(2), uint64(1), uint64(2)}}
	tx4 := []*ItemTxParam{&ItemTxParam{uint64(1), uint64(1), uint64(10)}}
	// tx5 := []*ItemTxParam{&ItemTxParam{uint64(2), uint64(1), uint64(1)}, &ItemTxParam{uint64(1), uint64(1), uint64(10)}}
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
		{"TransferItem", args{"itemtype1", "itemtest2", tx1}, nil},
		{"TransferItem", args{"itemtype1", "itemtest2", tx2}, nil},
		{"TransferItem", args{"itemtype1", "itemtest2", tx3}, ErrAccountNoItem},
		{"TransferItem", args{"itemtype1", "itemtest2", tx4}, ErrInsufficientBalance},
		// {"TransferItem", args{"itemtype1", "itemtest2", tx5}, ErrAccountNoItem},
	}
	for _, item := range testItem {
		if err := pm.TransferItem(item.arg.from, item.arg.to, item.arg.itemTx); err != item.err {
			t.Errorf("%q. TransferItem() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}
