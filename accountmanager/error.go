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

import "errors"

var (
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrNewAccountErr        = errors.New("new account err")
	ErrAssetIDInvalid       = errors.New("asset id invalid")
	ErrCreateAccountError   = errors.New("create account error")
	ErrAccountIsExist       = errors.New("account is exist")
	ErrAccountIsDestroy     = errors.New("account is destory")
	ErrAccountNotExist      = errors.New("account not exist")
	ErrHashIsEmpty          = errors.New("hash is empty")
	ErrkeyNotSame           = errors.New("key not same")
	ErrAccountNameInvalid   = errors.New("account name is Invalid")
	ErrInvalidPubKey        = errors.New("invalid public key")
	ErrAccountIsNil         = errors.New("account object is empty")
	ErrCodeIsEmpty          = errors.New("code is empty")
	ErrAmountValueInvalid   = errors.New("amount value is invalid")
	ErrAccountAssetNotExist = errors.New("account asset not exist")
	ErrUnkownTxType         = errors.New("Not support action type")
)
