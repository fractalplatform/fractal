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
	"fmt"

	"github.com/fractalplatform/fractal/feemanager"
)

type FeeAPI struct {
	b Backend
}

const (
	MaxIDCount = uint64(1000)
)

func NewFeeAPI(b Backend) *FeeAPI {
	return &FeeAPI{b}
}

//GetObjectFeeByName get object fee by name
func (aapi *FeeAPI) GetObjectFeeByName(ctx context.Context, objectName string, objectType uint64) (*feemanager.ObjectFee, error) {
	fm, err := aapi.b.GetFeeManager()
	if err != nil {
		return nil, err
	}

	return fm.GetObjectFeeByName(objectName, objectType)
}

//GetObjectFeeByIDRange get object fee by name
func (aapi *FeeAPI) GetObjectFeeResult(ctx context.Context, startObjectFeeID uint64, count uint64) (*feemanager.ObjectFeeResult, error) {
	if count > MaxIDCount {
		return nil, fmt.Errorf("count can not over %d", MaxIDCount)
	}

	fm, err := aapi.b.GetFeeManager()
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

	objectFeeResult := &feemanager.ObjectFeeResult{Continue: false,
		ObjectFees: make([]*feemanager.ObjectFee, 0)}
	for index := uint64(0); index < count; index++ {
		objectFeeID := startObjectFeeID + index

		if objectFeeID <= feeCounter {
			objectFee, err := fm.GetObjectFeeByID(objectFeeID)

			if err != nil {
				return nil, err
			}
			if objectFee != nil {
				objectFeeResult.ObjectFees = append(objectFeeResult.ObjectFees, objectFee)
			}
		}
	}

	if (startObjectFeeID + count) <= feeCounter {
		objectFeeResult.Continue = true
	}

	return objectFeeResult, nil
}

//GetFeeManagerByTime
func (aapi *FeeAPI) GetObjectFeeResultByTime(ctx context.Context, time uint64, startObjectFeeID uint64, count uint64) (*feemanager.ObjectFeeResult, error) {
	if count > MaxIDCount {
		return nil, fmt.Errorf("count can not over %d", MaxIDCount)
	}

	fm, err := aapi.b.GetFeeManagerByTime(time)
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

	objectFeeResult := &feemanager.ObjectFeeResult{Continue: false,
		ObjectFees: make([]*feemanager.ObjectFee, 0)}
	for index := uint64(0); index < count; index++ {
		objectFeeID := startObjectFeeID + index

		if objectFeeID <= feeCounter {
			objectFee, err := fm.GetObjectFeeByID(objectFeeID)

			if err != nil {
				return nil, err
			}
			if objectFee != nil {
				objectFeeResult.ObjectFees = append(objectFeeResult.ObjectFees, objectFee)
			}
		}
	}

	if (startObjectFeeID + count) <= feeCounter {
		objectFeeResult.Continue = true
	}

	return objectFeeResult, nil
}
