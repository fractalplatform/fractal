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
)

func Test_newAssetObject(t *testing.T) {
	type args struct {
		assetName   string
		symbol      string
		amount      *big.Int
		dec         uint64
		founder     common.Name
		owner       common.Name
		UpperLimit  *big.Int
		description string
	}
	tests := []struct {
		name    string
		args    args
		want    *AssetObject
		wantErr bool
	}{
		// TODO: Add test cases.
		{"normal", args{"ft", "ft", big.NewInt(2), 18, common.Name(""), common.Name("a123"), big.NewInt(999999), ""}, &AssetObject{0, 0, 0, "ft", "ft", big.NewInt(2), 18, common.Name(""), common.Name("a123"), big.NewInt(2), big.NewInt(999999), common.Name(""), ""}, false},
		{"shortname", args{"z", "z", big.NewInt(2), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"longname", args{"ftt0123456789ftt12", "zz", big.NewInt(2), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"emptyname", args{"", "z", big.NewInt(2), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"symbolempty", args{"ft", "", big.NewInt(2), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"amount==0", args{"ft", "z", big.NewInt(-1), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"ownerempty", args{"ft", "z", big.NewInt(2), 18, common.Name(""), common.Name(""), big.NewInt(999999), ""}, nil, true},
		{"shortsymbol", args{"ft", "z", big.NewInt(2), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"longsymbol", args{"ft", "ftt0123456789ftt1", big.NewInt(2), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"emptyname", args{"ft", "#ip0123456789ft", big.NewInt(2), 18, common.Name("a123"), common.Name("a123"), big.NewInt(999999), ""}, nil, true},
		{"limiterror", args{"ft", "ft", big.NewInt(101), 18, common.Name(""), common.Name("a123"), big.NewInt(100), ""}, nil, true},
		{"descerror", args{"ft", "ft", big.NewInt(100), 18, common.Name(""), common.Name("a123"), big.NewInt(101),
			"aaaaaaaaaabbbbbbbbbbaaaaaaaaaabbbbbbbbbbaaaaaaaaaa" +
				"bbbbbbbbbbaaaaaaaaaabbbbbbbbbbaaaaaaaaaabbbbbbbbbb" +
				"aaaaaaaaaabbbbbbbbbbaaaaaaaaaabbbbbbbbbbaaaaaaaaaa" +
				"bbbbbbbbbbaaaaaaaaaabbbbbbbbbbaaaaaaaaaabbbbbbbbbb" +
				"aaaaaaaaaabbbbbbbbbbaaaaaaaaaabbbbbbbbbbaaaaaaaaaa" +
				"bbbbbbbbbbaaaaaaaaaabbbbbbbbbbaaaaaaaaaabbbbbbbbbb"}, nil, true},
	}
	for _, tt := range tests {
		got, err := NewAssetObject(tt.args.assetName, 0, tt.args.symbol, tt.args.amount, tt.args.dec, tt.args.founder, tt.args.owner, tt.args.UpperLimit, common.Name(""), tt.args.description)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. newAssetObject() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. newAssetObject() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAssetObject_GetAssetID(t *testing.T) {
	type fields struct {
		AssetID    uint64
		AssetName  string
		Symbol     string
		Amount     *big.Int
		Decimals   uint64
		founder    common.Name
		Owner      common.Name
		AddIssue   *big.Int
		UpperLimit *big.Int
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		// TODO: Add test cases.
		{"normal", fields{1, "ft", "ft0123456789ft", big.NewInt(2), 18, common.Name(""), common.Name("a123"), big.NewInt(0), big.NewInt(999999)}, 1},
		{"max", fields{18446744073709551615, "ft", "ft0123456789ft", big.NewInt(2), 18, common.Name(""), common.Name("a123"), big.NewInt(0), big.NewInt(999999)}, 18446744073709551615},
		//{"min", fields{0, "ft", "ft0123456789ft", big.NewInt(2), 18, common.Name("a123")}, 0},
		//{">max", fields{18446744073709551616, "ft", "ft0123456789ft", big.NewInt(2), 18, common.Name("a123")}, 0},
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:    tt.fields.AssetID,
			AssetName:  tt.fields.AssetName,
			Symbol:     tt.fields.Symbol,
			Amount:     tt.fields.Amount,
			Decimals:   tt.fields.Decimals,
			Founder:    tt.fields.founder,
			Owner:      tt.fields.Owner,
			AddIssue:   tt.fields.AddIssue,
			UpperLimit: tt.fields.UpperLimit,
		}
		if got := ao.GetAssetID(); got != tt.want {
			t.Errorf("%q. AssetObject.GetAssetID() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAssetObject_SetAssetID(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	type args struct {
		AssetID uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		//{">max", fields{0, "ft", "ft0123456789ft", big.NewInt(2), 18, common.Name("a123")}, args{18446744073709551616}},
		{"max", fields{0, "ft", "ft0123456789ftft", big.NewInt(2), 18, common.Name("a123")}, args{18446744073709551615}},
		{"normal", fields{0, "ft", "ft0123456789ft", big.NewInt(2), 18, common.Name("a123")}, args{184467}},
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		ao.SetAssetID(tt.args.AssetID)
	}
}

func TestAssetObject_GetSymbol(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		//	{"getexist",fields{}}
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		if got := ao.GetSymbol(); got != tt.want {
			t.Errorf("%q. AssetObject.GetSymbol() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAssetObject_SetSymbol(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	type args struct {
		sym string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		ao.SetSymbol(tt.args.sym)
	}
}

func TestAssetObject_GetDecimals(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}

	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		if got := ao.GetDecimals(); got != tt.want {
			t.Errorf("%q. AssetObject.GetDecimals() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAssetObject_SetDecimals(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	type args struct {
		dec uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		ao.SetDecimals(tt.args.dec)
	}
}

func TestAssetObject_GetAssetName(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		if got := ao.GetAssetName(); got != tt.want {
			t.Errorf("%q. AssetObject.GetAssetName() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAssetObject_SetAssetName(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	type args struct {
		assetName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		ao.SetAssetName(tt.args.assetName)
	}
}

func TestAssetObject_GetAssetAmount(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	tests := []struct {
		name   string
		fields fields
		want   *big.Int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		if got := ao.GetAssetAmount(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AssetObject.GetAssetAmount() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAssetObject_SetAssetAmount(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	type args struct {
		amount *big.Int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		ao.SetAssetAmount(tt.args.amount)
	}
}

func TestAssetObject_GetAssetOwner(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	tests := []struct {
		name   string
		fields fields
		want   common.Name
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		if got := ao.GetAssetOwner(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. AssetObject.GetAssetOwner() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAssetObject_SetAssetOwner(t *testing.T) {
	type fields struct {
		AssetID   uint64
		AssetName string
		Symbol    string
		Amount    *big.Int
		Decimals  uint64
		Owner     common.Name
	}
	type args struct {
		owner common.Name
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		ao := &AssetObject{
			AssetID:   tt.fields.AssetID,
			AssetName: tt.fields.AssetName,
			Symbol:    tt.fields.Symbol,
			Amount:    tt.fields.Amount,
			Decimals:  tt.fields.Decimals,
			Owner:     tt.fields.Owner,
		}
		ao.SetAssetOwner(tt.args.owner)
	}
}

func TestAssetObject_AssetNumber(t *testing.T) {
	ao := NewAssetObjectNoCheck("", uint64(0), "", big.NewInt(0), 0, common.Name(""), common.Name(""), big.NewInt(1000), common.Name(""), "")
	// ao := &AssetObject{
	// 	UpperLimit: big.NewInt(1000),
	// }

	id := uint64(1)
	ao.SetAssetID(id)
	if got := ao.GetAssetID(); got != id {
		t.Errorf("AssetObject.GetAssetID() = %v, want %v", got, id)
	}

	number := uint64(2)
	ao.SetAssetNumber(number)
	if got := ao.GetAssetNumber(); got != number {
		t.Errorf("AssetObject.GetAssetNumber() = %v, want %v", got, number)
	}

	stats := uint64(3)
	ao.SetAssetStats(stats)
	if got := ao.GetAssetStats(); got != stats {
		t.Errorf("AssetObject.GetAssetStats() = %v, want %v", got, stats)
	}

	symbol := "symbol"
	ao.SetSymbol(symbol)
	if got := ao.GetSymbol(); got != symbol {
		t.Errorf("AssetObject.GetSymbol() = %v, want %v", got, symbol)
	}

	decimals := uint64(4)
	ao.SetDecimals(decimals)
	if got := ao.GetDecimals(); got != decimals {
		t.Errorf("AssetObject.GetDecimals() = %v, want %v", got, decimals)
	}

	name := "name"
	ao.SetAssetName(name)
	if got := ao.GetAssetName(); got != name {
		t.Errorf("AssetObject.GetAssetName() = %v, want %v", got, name)
	}

	amount := big.NewInt(5)
	ao.SetAssetAmount(amount)
	if got := ao.GetAssetAmount(); got != amount {
		t.Errorf("AssetObject.GetAssetAmount() = %v, want %v", got, amount)
	}

	issue := big.NewInt(6)
	ao.SetAssetAddIssue(issue)
	if got := ao.GetAssetAddIssue(); got != issue {
		t.Errorf("AssetObject.GetAssetAddIssue() = %v, want %v", got, issue)
	}

	if got := ao.GetUpperLimit(); got != ao.UpperLimit {
		t.Errorf("AssetObject.GetUpperLimit() = %v, want %v", got, ao.UpperLimit)
	}

	contract := common.Name("contract")
	ao.SetAssetContract(contract)
	if got := ao.GetAssetContract(); got != contract {
		t.Errorf("AssetObject.GetAssetContract() = %v, want %v", got, contract)
	}
	if got := ao.GetContract(); got != contract {
		t.Errorf("AssetObject.GetContract() = %v, want %v", got, contract)
	}

	founder := common.Name("founder")
	ao.SetAssetFounder(founder)
	if got := ao.GetAssetFounder(); got != founder {
		t.Errorf("AssetObject.GetAssetFounder() = %v, want %v", got, founder)
	}

	owner := common.Name("owner")
	ao.SetAssetOwner(owner)
	if got := ao.GetAssetOwner(); got != owner {
		t.Errorf("AssetObject.GetAssetFounder() = %v, want %v", got, owner)
	}

	desc := "desc"
	ao.SetAssetDescription(desc)
	if got := ao.GetAssetDescription(); got != desc {
		t.Errorf("AssetObject.GetAssetFounder() = %v, want %v", got, desc)
	}

}
