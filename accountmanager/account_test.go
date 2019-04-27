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

package accountmanager

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/fractalplatform/fractal/common"
)

func Test_newAssetBalance(t *testing.T) {
	type args struct {
		assetID uint64
		amount  *big.Int
	}
	tests := []struct {
		name string
		args args
		want *AssetBalance
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := newAssetBalance(tt.args.assetID, tt.args.amount); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. newAssetBalance() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNewAccount(t *testing.T) {
	type args struct {
		accountName common.Name
		founderName common.Name
		pubkey      common.PubKey
		detail      string
	}
	tests := []struct {
		name    string
		args    args
		want    *Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := NewAccount(tt.args.accountName, tt.args.founderName, tt.args.pubkey, tt.args.detail)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. NewAccount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. NewAccount() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_GetName(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   common.Name
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if got := a.GetName(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Account.GetName() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_GetNonce(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if got := a.GetNonce(); got != tt.want {
			t.Errorf("%q. Account.GetNonce() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_SetNonce(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		nonce uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		a.SetNonce(tt.args.nonce)
	}
}

func TestAccount_GetCode(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		got, err := a.GetCode()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.GetCode() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Account.GetCode() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_GetCodeSize(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if got := a.GetCodeSize(); got != tt.want {
			t.Errorf("%q. Account.GetCodeSize() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_SetCode(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		code []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if err := a.SetCode(tt.args.code); (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.SetCode() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccount_GetCodeHash(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    common.Hash
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		got, err := a.GetCodeHash()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.GetCodeHash() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Account.GetCodeHash() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_GetBalanceByID(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		assetID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *big.Int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		got, err := a.GetBalanceByID(tt.args.assetID)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.GetBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Account.GetBalanceByID() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_GetBalancesList(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []*AssetBalance
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if got := a.GetBalancesList(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Account.GetBalancesList() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_GetAllBalances(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[uint64]*big.Int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		got, err := a.GetAllBalances()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.GetAllBalances() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Account.GetAllBalances() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_binarySearch(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		assetID uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		got, got1 := a.binarySearch(tt.args.assetID)
		if got != tt.want {
			t.Errorf("%q. Account.binarySearch() got = %v, want %v", tt.name, got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("%q. Account.binarySearch() got1 = %v, want %v", tt.name, got1, tt.want1)
		}
	}
}

func TestAccount_AddNewAssetByAssetID(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		assetID uint64
		amount  *big.Int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		a.AddNewAssetByAssetID(tt.args.assetID, tt.args.amount)
	}
}

func TestAccount_SetBalance(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		assetID uint64
		amount  *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if err := a.SetBalance(tt.args.assetID, tt.args.amount); (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.SetBalance() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccount_SubBalanceByID(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		assetID uint64
		value   *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if err := a.SubBalanceByID(tt.args.assetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.SubBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccount_AddBalanceByID(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		assetID uint64
		value   *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if err := a.AddBalanceByID(tt.args.assetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.AddBalanceByID() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccount_EnoughAccountBalance(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	type args struct {
		assetID uint64
		value   *big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if err := a.EnoughAccountBalance(tt.args.assetID, tt.args.value); (err != nil) != tt.wantErr {
			t.Errorf("%q. Account.EnoughAccountBalance() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestAccount_IsSuicided(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if got := a.IsSuicided(); got != tt.want {
			t.Errorf("%q. Account.IsSuicided() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_SetSuicide(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		a.SetSuicide()
	}
}

func TestAccount_IsDestroyed(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		if got := a.IsDestroyed(); got != tt.want {
			t.Errorf("%q. Account.IsDestroyed() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAccount_SetDestroy(t *testing.T) {
	type fields struct {
		AcctName  common.Name
		Nonce     uint64
		PublicKey common.PubKey
		Code      []byte
		CodeHash  common.Hash
		CodeSize  uint64
		Balances  []*AssetBalance
		Suicide   bool
		Destroy   bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		a := &Account{
			AcctName: tt.fields.AcctName,
			Nonce:    tt.fields.Nonce,
			Code:     tt.fields.Code,
			CodeHash: tt.fields.CodeHash,
			CodeSize: tt.fields.CodeSize,
			Balances: tt.fields.Balances,
			Suicide:  tt.fields.Suicide,
			Destroy:  tt.fields.Destroy,
		}
		a.SetDestroy()
	}
}
