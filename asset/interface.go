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

	"github.com/fractalplatform/fractal/common"
)

type AssetIf interface {
	GetAssetIDByName(assetName common.Name) (uint64, error)
	GetAllAssetObject() []*AssetObject
	IssueAssetObject(ao *AssetObject) (uint64, error)
	IssueAsset(assetName common.Name, symbol string, amount *big.Int, owner common.Name) error
	IncreaseAsset(accountName common.Name, assetID uint64, amount *big.Int) error
	SetAssetNewOwner(accountName common.Name, assetID uint64, newOwner common.Name) error
}
