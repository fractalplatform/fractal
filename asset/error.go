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

import "errors"

var (
	ErrAccountNameNull      = errors.New("account name is null")
	ErrAssetIsExist         = errors.New("asset is exist")
	ErrAssetNotExist        = errors.New("asset not exist")
	ErrOwnerMismatch        = errors.New("asset owner mismatch")
	ErrAssetNameEmpty       = errors.New("asset name is empty")
	ErrAssetObjectEmpty     = errors.New("asset object is empty")
	ErrNewAssetObject       = errors.New("create asset object input invalid")
	ErrAssetAmountZero      = errors.New("asset amount is zero")
	ErrUpperLimit           = errors.New("asset amount over the issuance limit")
	ErrDestroyLimit         = errors.New("asset destroy exceeding the lower limit")
	ErrAssetCountNotExist   = errors.New("asset total count not exist")
	ErrAssetIDInvalid       = errors.New("asset id invalid")
	ErrAssetManagerNotExist = errors.New("asset manager name not exist")
	ErrDetailTooLong        = errors.New("detail info exceed maximum")
	ErrNegativeAmount       = errors.New("negative amount")
)
