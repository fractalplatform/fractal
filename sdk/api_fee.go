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

package sdk

import "github.com/fractalplatform/fractal/feemanager"

// FeeInfo get object fee by name
func (api *API) FeeInfo(name string, objectType uint64) (map[string]interface{}, error) {
	feeInfo := map[string]interface{}{}
	err := api.client.Call(feeInfo, "fee_getObjectFeeByName", name, objectType)
	return feeInfo, err
}

//FeeInfoByID get object fee by id
func (api *API) FeeInfoByID(startFeeID uint64, count uint64, time uint64) (*feemanager.ObjectFeeResult, error) {
	feeInfo := &feemanager.ObjectFeeResult{}
	err := api.client.Call(feeInfo, "fee_getObjectFeeResult", startFeeID, count, time)
	return feeInfo, err
}
