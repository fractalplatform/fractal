// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package asset

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/state"
	memdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
)

var astdb = getStateDB()

var ast = getAsset()

func getStateDB() *state.StateDB {
	db := memdb.NewMemDatabase()
	tridb := state.NewDatabase(db)
	statedb, err := state.New(common.Hash{}, tridb)
	if err != nil {
		//t.Fatal("test getStateDB failure ", err)
		return nil
	}
	return statedb
}
func getAsset() *Asset {
	return NewAsset(astdb)
}

func TestAsset_InitAssetCount(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	db := memdb.NewMemDatabase()
	tridb := state.NewDatabase(db)
	statedb, err := state.New(common.Hash{}, tridb)
	if err != nil {
		//t.Fatal("test getStateDB failure ", err)
	}
	tests := []struct {
		name   string
		fields fields
	}{
		//
		{"init", fields{statedb}},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		a.InitAssetCount()
	}
	ast1 := NewAsset(statedb)
	num, _ := ast1.getAssetCount()
	if num != 0 {
		t.Errorf("InitAssetCount err")
	}
}

func TestNewAsset(t *testing.T) {
	type args struct {
		sdb *state.StateDB
	}

	tests := []struct {
		name string
		args args
		want *Asset
	}{
		//
		//{"newnil", args{nil}, nil},
		{"new", args{astdb}, ast},
	}
	for _, tt := range tests {
		if got := NewAsset(tt.args.sdb); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. NewAsset() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
func TestAsset_GetAssetObjectByName(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		assetName string
	}

	ao, _ := NewAssetObject("ft", 0, "zz", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	//ao.SetAssetId(0)
	ast.addNewAssetObject(ao)
	ao1, _ := NewAssetObject("ft2", 0, "zz2", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	//ao1.SetAssetId(1)
	ast.addNewAssetObject(ao1)
	ao2, _ := NewAssetObject("ft0", 0, "zz0", big.NewInt(1000), 0, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	//ao1.SetAssetId(2)
	ast.addNewAssetObject(ao2)
	ao3, _ := NewAssetObject("ftc", 0, "zzc", big.NewInt(1000), 0, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name("a123456789aeee"), "")
	//ao3.SetAssetId(3)
	ast.addNewAssetObject(ao3)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AssetObject
		wantErr bool
	}{
		// TODO: Add test cases.
		{"getall", fields{astdb}, args{"ft"}, ao, false},
		{"getall2", fields{astdb}, args{"ft2"}, ao1, false},
		{"getall3", fields{astdb}, args{"ft0"}, ao2, false},
		{"getall4", fields{astdb}, args{"ftc"}, ao3, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		got, err := a.GetAssetObjectByName(tt.args.assetName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.GetAssetObjectByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Asset.GetAssetObjectByName() = %v, want %v", tt.name, got, tt.want)
		}
		t.Logf("GetAssetObjectByName asset dec=%v", got.Decimals)
	}
}

func TestAsset_addNewAssetObject(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		ao *AssetObject
	}

	ao3, _ := NewAssetObject("ft3", 0, "zz3", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	//ao1.SetAssetId(3)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
		{"addnil", fields{astdb}, args{nil}, 0, true},
		{"add", fields{astdb}, args{ao3}, 4, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		got, err := a.addNewAssetObject(tt.args.ao)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.addNewAssetObject() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. Asset.addNewAssetObject() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAsset_GetAssetIdByName(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		assetName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		//
		{"normal", fields{astdb}, args{""}, 0, true},
		{"normal", fields{astdb}, args{"ft"}, 0, false},
		{"wrong", fields{astdb}, args{"ft2"}, 1, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		got, err := a.GetAssetIdByName(tt.args.assetName)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.GetAssetIdByName() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. Asset.GetAssetIdByName() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAsset_GetAssetObjectById(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		id uint64
	}

	ao, _ := NewAssetObject("ft", 0, "zz", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	ao.SetAssetId(0)
	ast.IssueAssetObject(ao)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AssetObject
		wantErr bool
	}{
		//
		{"assetnotexist", fields{astdb}, args{222}, nil, true},
		{"normal2", fields{astdb}, args{0}, ao, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		got, err := a.GetAssetObjectById(tt.args.id)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.GetAssetObjectById() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Asset.GetAssetObjectById() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAsset_getAssetCount(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}

	tests := []struct {
		name    string
		fields  fields
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
		{"get", fields{astdb}, 5, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		got, err := a.getAssetCount()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.getAssetCount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. Asset.getAssetCount() = %v, want %v", tt.name, got, tt.want)
		}
	}
	ao, _ := NewAssetObject("ft2", 0, "zz2", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	//ao.SetAssetId(1)
	ast.IssueAssetObject(ao)
	num, err := ast.getAssetCount()
	if err != nil {
		t.Errorf("get asset count err")
	}
	if num != 5 {
		t.Errorf("test asset count err")
	}
}

// func TestAsset_GetAllAssetObject(t *testing.T) {
// 	type fields struct {
// 		sdb *state.StateDB
// 	}
// 	aslice := make([]*AssetObject, 0)
// 	ao, _ := ast.GetAssetObjectById(1)
// 	aslice = append(aslice, ao)
// 	ao, _ = ast.GetAssetObjectById(2)
// 	aslice = append(aslice, ao)
// 	ao, _ = ast.GetAssetObjectById(3)
// 	aslice = append(aslice, ao)

// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		want    []*AssetObject
// 		wantErr bool
// 	}{
// 		//
// 		//{"getall", fields{astdb}, aslice, false},
// 	}
// 	for _, tt := range tests {
// 		a := &Asset{
// 			sdb: tt.fields.sdb,
// 		}
// 		got, err := a.GetAllAssetObject()
// 		if (err != nil) != tt.wantErr {
// 			t.Errorf("%q. Asset.GetAllAssetObject() error = %v, wantErr %v", tt.name, err, tt.wantErr)
// 			continue
// 		}
// 		if !reflect.DeepEqual(got, tt.want) {
// 			t.Errorf("%q. Asset.GetAllAssetObject() = %v, want %v", tt.name, got, tt.want)
// 		}
// 	}
// }

func TestAsset_SetAssetObject(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		ao *AssetObject
	}

	ao4, _ := NewAssetObject("ft4", 0, "zz4", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	ao4.SetAssetId(54)
	ao5, _ := NewAssetObject("ft5", 0, "zz5", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	ao5.SetAssetId(55)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"setnil", fields{astdb}, args{nil}, true},
		{"add", fields{astdb}, args{ao4}, false},
		{"add2", fields{astdb}, args{ao5}, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		if err := a.SetAssetObject(tt.args.ao); (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.SetAssetObject() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAsset_IssueAssetObject(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		ao *AssetObject
	}
	ao6, _ := NewAssetObject("ft6", 0, "zz6", big.NewInt(1000), 10, common.Name(""), common.Name("a123456789aeee"), big.NewInt(9999999999), common.Name(""), "")
	ao6.SetAssetId(11)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"nil", fields{astdb}, args{nil}, true},
		{"add", fields{astdb}, args{ao6}, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		if _, err := a.IssueAssetObject(tt.args.ao); (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.IssueAssetObject() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAsset_IssueAsset(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		assetName string
		symbol    string
		amount    *big.Int
		dec       uint64
		founder   common.Name
		owner     common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"nilname", fields{astdb}, args{"", "z", big.NewInt(1), 2, common.Name(""), common.Name("11")}, true},
		{"nilsym", fields{astdb}, args{"22", "", big.NewInt(2), 2, common.Name(""), common.Name("11")}, true},
		{"exist", fields{astdb}, args{"ft", "3", big.NewInt(2), 2, common.Name(""), common.Name("11")}, true},
		{"normal", fields{astdb}, args{"ft22", "23", big.NewInt(2), 2, common.Name(""), common.Name("a112345698")}, true},
		// {"normal1", fields{astdb}, args{"ft22.ft33", "23", big.NewInt(2), 2, common.Name(""), common.Name("112345698")}, false},
		// {"normal2", fields{astdb}, args{"ft22.ft44.ft55", "23", big.NewInt(2), 2, common.Name(""), common.Name("112345698")}, false},
		// {"errorowner", fields{astdb}, args{"ft22.ft44.ft55", "23", big.NewInt(2), 2, common.Name(""), common.Name("11234512")}, true},
		// {"errorowner1", fields{astdb}, args{"ft23.ft34", "23", big.NewInt(2), 2, common.Name(""), common.Name("11234512")}, true},
		// {"errorowner2", fields{astdb}, args{"ft23", "24", big.NewInt(2), 2, common.Name(""), common.Name("11234512")}, false},
		// {"errorowner3", fields{astdb}, args{"ft23.ft34", "24", big.NewInt(2), 2, common.Name(""), common.Name("11234512")}, false},
		// {"errorowner4", fields{astdb}, args{"ft24.", "25", big.NewInt(2), 2, common.Name(""), common.Name("11234523")}, true},
		// {"errorowner5", fields{astdb}, args{"ft24..", "25", big.NewInt(2), 2, common.Name(""), common.Name("11234523")}, true},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		if _, err := a.IssueAsset(tt.args.assetName, 0, 0, tt.args.symbol, tt.args.amount, tt.args.dec, tt.args.founder, tt.args.owner, big.NewInt(9999999999), common.Name(""), ""); (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.IssueAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAsset_IncreaseAsset(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		accountName common.Name
		assetId     uint64
		amount      *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"nilname", fields{astdb}, args{common.Name(""), 1, big.NewInt(2)}, true},
		{"wrongid", fields{astdb}, args{common.Name("11"), 0, big.NewInt(2)}, false},
		{"wrongamount", fields{astdb}, args{common.Name("11"), 0, big.NewInt(-2)}, true},
		{"normal", fields{astdb}, args{common.Name("a123456789aeee"), 1, big.NewInt(50)}, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		if err := a.IncreaseAsset(tt.args.accountName, tt.args.assetId, tt.args.amount); (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.IncreaseAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAsset_SetAssetNewOwner(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		accountName common.Name
		assetId     uint64
		newOwner    common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases
		{"nilname", fields{astdb}, args{common.Name(""), 1, common.Name("")}, true},
		{"wrongid", fields{astdb}, args{common.Name("11"), 0, common.Name("")}, false},
		{"wrongamount", fields{astdb}, args{common.Name("11"), 123, common.Name("")}, true},
		{"normal", fields{astdb}, args{common.Name("a123456789aeee"), 1, common.Name("a123456789afff")}, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		if err := a.SetAssetNewOwner(tt.args.accountName, tt.args.assetId, tt.args.newOwner); (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.SetAssetNewOwner() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAsset_UpdateAsset(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		accountName common.Name
		assetId     uint64
		Owner       common.Name
		founder     common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases
		{"nilname", fields{astdb}, args{common.Name(""), 1, common.Name(""), common.Name("")}, true},
		{"wrongassetid", fields{astdb}, args{common.Name("11"), 0, common.Name(""), common.Name("")}, false},
		{"wrongamount", fields{astdb}, args{common.Name("11"), 123, common.Name(""), common.Name("")}, true},
		{"nilfounder", fields{astdb}, args{common.Name("a123456789afff"), 1, common.Name("a123456789aeee"), common.Name("")}, false},
		{"normal", fields{astdb}, args{common.Name("a123456789afff"), 1, common.Name("a123456789afff"), common.Name("a123456789afff")}, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		if err := a.UpdateAsset(tt.args.accountName, tt.args.assetId, tt.args.founder); (err != nil) != tt.wantErr {
			t.Errorf("%q. Asset.updateAsset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}
func TestAsset_HasAccesss(t *testing.T) {
	type fields struct {
		sdb *state.StateDB
	}
	type args struct {
		assetId uint64
		name    common.Name
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases
		{"a0", fields{astdb}, args{0, common.Name("")}, true},
		{"a1", fields{astdb}, args{1, common.Name("")}, true},
		{"a2", fields{astdb}, args{2, common.Name("")}, true},
		{"a3", fields{astdb}, args{3, common.Name("a123456789aeee")}, true},
		{"a3_1", fields{astdb}, args{3, common.Name("a123456789afff")}, false},
	}
	for _, tt := range tests {
		a := &Asset{
			sdb: tt.fields.sdb,
		}
		if has := a.HasAccess(tt.args.assetId, tt.args.name); has != tt.wantErr {
			t.Errorf("%q. Asset.HasAccess() error = %v, wantErr %v", tt.name, has, tt.wantErr)
		}
	}
}
