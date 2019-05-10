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

	"github.com/fractalplatform/fractal/feemanager"
	"github.com/fractalplatform/fractal/params"
)

type FeeAPI struct {
	b Backend
}

func NewFeeAPI(b Backend) *FeeAPI {
	return &FeeAPI{b}
}

//GetObjectFeeByName get object fee by name
//objectName: Asset Name, Contract Name, Coinbase Name
//objectType:  Asset Type(0),Contract Type(1),Coinbase Type(2)
func (fapi *FeeAPI) GetObjectFeeByName(ctx context.Context, objectName string, objectType uint64) (*feemanager.ObjectFee, error) {
	fm, err := fapi.b.GetFeeManager()
	if err != nil {
		return nil, err
	}

	return fm.GetObjectFeeByName(objectName, objectType)
}

//GetObjectFeeResult get object fee infomation
//startObjectFeeID: object fee id, start from 1
//count: The count of results obtained at one time, If it's more than 1,000, it's 1,000
//time: snapshot time
func (fapi *FeeAPI) GetObjectFeeResult(ctx context.Context, startObjectFeeID uint64, count uint64, time uint64) (*feemanager.ObjectFeeResult, error) {
	var fm *feemanager.FeeManager
	var err error
	var bContinue bool

	if count > params.MaxFeeResultCount {
		count = params.MaxFeeResultCount
	}

	if time == 0 {
		fm, err = fapi.b.GetFeeManager()
	} else {
		fm, err = fapi.b.GetFeeManagerByTime(time)
	}
	if err != nil {
		return nil, err
	}

	feeCounter, err := fm.GetFeeCounter()
	if err != nil {
		return nil, err
	}

	if feeCounter == 0 || feeCounter < startObjectFeeID {
		return nil, nil
	}

	if count <= (feeCounter - startObjectFeeID) {
		bContinue = true
	} else {
		count = feeCounter - startObjectFeeID + 1
		bContinue = false
	}

	objectFeeResult := &feemanager.ObjectFeeResult{Continue: bContinue,
		ObjectFees: make([]*feemanager.ObjectFee, 0, count)}
	for index := uint64(0); index < count; index++ {
		objectFeeID := startObjectFeeID + index
		objectFee, err := fm.GetObjectFeeByID(objectFeeID)
		if err != nil {
			return nil, err
		}
		if objectFee != nil {
			objectFeeResult.ObjectFees = append(objectFeeResult.ObjectFees, objectFee)
		}
	}

	return objectFeeResult, nil
}
