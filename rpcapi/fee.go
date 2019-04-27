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

package rpcapi

import (
	"context"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/feemanager"
)

type FeeAPI struct {
	b Backend
}

func NewFeeAPI(b Backend) *FeeAPI {
	return &FeeAPI{b}
}

//GetObjectFeeByName get object fee by name
func (aapi *FeeAPI) GetObjectFeeByName(ctx context.Context, objectName common.Name) (*feemanager.ObjectFee, error) {
	fm, err := aapi.b.GetFeeManager()
	if err != nil {
		return nil, err
	}

	return fm.GetObjectFeeByName(objectName)
}
