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

// Config Asset name level
type Config struct {
	AssetNameLevel     uint64 `json:"assetNameLevel"`
	AssetNameLength    uint64 `json:"assetNameLength"`
	SubAssetNameLength uint64 `json:"subAssetNameLength"`
}

// DefaultAssetNameConf return asset config
func DefaultAssetNameConf() *Config {
	return &Config{
		AssetNameLevel:     0,
		AssetNameLength:    16,
		SubAssetNameLength: 0,
	}
}

const MaxDetailLength uint64 = 255
