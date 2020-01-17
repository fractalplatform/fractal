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

func (api *ItemAPI) GetWorldByID(worldID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetWorldByID(worldID)
}

func (api *ItemAPI) GetWorldByName(worldName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetWorldByName(worldName)
}

func (api *ItemAPI) GetItemTypeByID(worldID, itemTypeID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemTypeByID(worldID, itemTypeID)
}

func (api *ItemAPI) GetItemTypeByName(worldID uint64, itemTypeName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemTypeByName(worldID, itemTypeName)
}

func (api *ItemAPI) GetItemByID(worldID, itemTypeID, itemID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemByID(worldID, itemTypeID, itemID)
}

func (api *ItemAPI) GetItemByOwner(worldID, itemTypeID uint64, owner string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemByOwner(worldID, itemTypeID, owner)
}

func (api *ItemAPI) GetAccountItems(worldID, itemTypeID uint64, account string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemsByOwner(worldID, itemTypeID, account)
}

func (api *ItemAPI) GetItemTypeAttributeByID(worldID, itemTypeID, attrID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemTypeAttributeByID(worldID, itemTypeID, attrID)
}

func (api *ItemAPI) GetItemTypeAttributeByName(worldID, itemTypeID uint64, attrName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemTypeAttributeByName(worldID, itemTypeID, attrName)
}

func (api *ItemAPI) GetItemAttributeByID(worldID, itemTypeID, itemID, attrID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemAttributeByID(worldID, itemTypeID, itemID, attrID)
}

func (api *ItemAPI) GetItemAttributeByName(worldID, itemTypeID, itemID uint64, attrName string) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetItemAttributeByName(worldID, itemTypeID, itemID, attrName)
}

func (api *ItemAPI) GetItemApprove(from, to string, worldID, itemTypeID, itemID uint64) (interface{}, error) {
	pm, err := api.b.GetPM()
	if err != nil {
		return 0, err
	}
	return pm.GetAuthorize(from, to, worldID, itemTypeID, itemID)
}
