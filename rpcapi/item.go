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

type ItemAPI struct {
	b Backend
}

func NewItemAPI(b Backend) *ItemAPI {
	return &ItemAPI{b}
}

func (api *ItemAPI) GetAccountItemAmount(account string, itemTypeID, itemInfoID uint64) (uint64, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemAmount(account, itemTypeID, itemInfoID)
}

func (api *ItemAPI) GetItemAttribute(itemTypeID uint64, itemInfoID uint64, AttributeName string) (string, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return "", err
	}
	return pm.GetItemAttribute(itemTypeID, itemInfoID, AttributeName)
}

func (api *ItemAPI) GetItemTypeByID(itemTypeID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return "", err
	}
	return pm.GetItemTypeByID(itemTypeID)
}

func (api *ItemAPI) GetItemTypeByName(creator string, itemTypeName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return "", err
	}
	return pm.GetItemTypeByName(creator, itemTypeName)
}

func (api *ItemAPI) GetItemInfoByID(itemTypeID uint64, itemInfoID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return "", err
	}
	return pm.GetItemInfoByID(itemTypeID, itemInfoID)
}

func (api *ItemAPI) GetItemInfoByName(itemTypeID uint64, itemInfoName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return "", err
	}
	return pm.GetItemInfoByName(itemTypeID, itemInfoName)
}
