package plugin

import (
	"math/big"
	"testing"
)

func Test_IssueAsset(t *testing.T) {
	type args struct {
		accountName string
		assetName   string
		symbol      string
		amount      *big.Int
		decimals    uint64
		founder     string
		owner       string
		limit       *big.Int
		description string
		aa          IAccount
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		// {"IssueAsset with err assetName", args{"assetowner", "FT", "ft", big.NewInt(1000), 10, "assetfounder", "assetowner", big.NewInt(10000), "issue for test", pm}, ErrAssetNameinvalid},
		// {"IssueAsset with nil assetName", args{"assetowner", "", "ft", big.NewInt(1000), 10, "assetfounder", "assetowner", big.NewInt(10000), "issue for test", pm}, ErrParamIsNil},
		// {"IssueAsset with err amount", args{"assetowner", "ftoken", "ft", big.NewInt(-1000), 10, "assetfounder", "assetowner", big.NewInt(10000), "issue for test", pm}, ErrAmountValueInvalid},
		{"IssueAsset another", args{"assetowner", "ftoken2", "ft2", big.NewInt(1000), 10, "assetfounder", "assetowner", big.NewInt(10000), "issue for test", pm}, ErrIssueAsset},
	}

	for _, item := range testItem {
		_, err := pm.IssueAsset(item.arg.accountName, item.arg.assetName, item.arg.symbol, item.arg.amount,
			item.arg.decimals, item.arg.founder, item.arg.owner, item.arg.limit, item.arg.description, item.arg.aa)
		if err != item.err {
			t.Errorf("%q. IssueAsset() error = %v, wantErr %v", item.testDes, err, item.err)
		}
	}
}

func Test_GetAssetID(t *testing.T) {
	type args struct {
		name string
	}
	testItem := []struct {
		testDes string
		arg     args
		wantID  uint64
		err     error
	}{
		{"GetAssetID", args{""}, 0, ErrAssetNotExist},
		{"GetAssetID", args{"ftoken2"}, 0, ErrAssetNotExist},
		{"GetAssetID", args{"ftoken"}, 0, nil},
	}

	for _, item := range testItem {
		get, err := pm.GetAssetID(item.arg.name)
		if err != item.err {
			t.Errorf("%q. GetAssetID() error = %v, wantErr %v", item.testDes, err, item.err)
		}
		if err == nil {
			if get != item.wantID {
				t.Errorf("%q. GetAssetID() assetID no equal", item.testDes)
			}
		}
	}
}

func Test_GetAssetName(t *testing.T) {
	type args struct {
		id uint64
	}
	testItem := []struct {
		testDes  string
		arg      args
		wantName string
		err      error
	}{
		{"GetAssetName", args{10}, "", ErrAssetNotExist},
		{"GetAssetName", args{0}, "ftoken", nil},
	}

	for _, item := range testItem {
		get, err := pm.GetAssetName(item.arg.id)
		if err != item.err {
			t.Errorf("%q. GetAssetName() error = %v, wantErr %v", item.testDes, err, item.err)
		}
		if err == nil {
			if get != item.wantName {
				t.Errorf("%q. GetAssetName() assetID no equal", item.testDes)
			}
		}
	}
}

func Test_IncreaseAsset(t *testing.T) {
	type args struct {
		from  string
		to    string
		id    uint64
		value *big.Int
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"IncreaseAsset exceed upper limit", args{"assetowner", "assetowner", SystemAssetID, big.NewInt(10000)}, ErrUpperLimit},
		{"IncreaseAsset", args{"assetowner", "assetowner", SystemAssetID, big.NewInt(1000)}, nil},
	}

	for _, item := range testItem {
		_, err := pm.IncreaseAsset(item.arg.from, item.arg.to, item.arg.id, item.arg.value, pm)
		if err != item.err {
			t.Errorf("%q. IncreaseAsset() error = %v, wantErr %v", item.testDes, err, item.err)
		}
		if err == nil {
			toGet, _ := pm.GetBalance(item.arg.to, item.arg.id)
			if toGet.Cmp(item.arg.value) != 0 {
				t.Errorf("%q. IncreaseAsset() toAccount balance no equal", item.testDes)
			}
		}
	}
}

func Test_DestroyAsset(t *testing.T) {
	type args struct {
		from  string
		id    uint64
		value *big.Int
	}
	testItem := []struct {
		testDes string
		arg     args
		err     error
	}{
		{"DestroyAsset with err amount", args{"assetowner", SystemAssetID, big.NewInt(10000)}, ErrDestroyLimit},
		{"DestroyAsset", args{"assetowner", SystemAssetID, big.NewInt(1000)}, nil},
	}

	for _, item := range testItem {
		_, err := pm.DestroyAsset(item.arg.from, item.arg.id, item.arg.value, pm)
		if err != item.err {
			t.Errorf("%q. DestroyAsset() error = %v, wantErr %v", item.testDes, err, item.err)
		}
		if err == nil {
			toGet, _ := pm.GetBalance(item.arg.from, item.arg.id)
			if toGet.Cmp(big.NewInt(0)) != 0 {
				t.Errorf("%q. DestroyAsset() fromAccount balance no equal", item.testDes)
			}
		}
	}
}
