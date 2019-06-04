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

// Config Account Level
type Config struct {
	AccountNameLevel         uint64 `json:"accountNameLevel"`
	AccountNameMaxLength     uint64 `json:"accountNameMaxLength"`
	MainAccountNameMinLength uint64 `json:"mainAccountNameMinLength"`
	MainAccountNameMaxLength uint64 `json:"mainAccountNameMaxLength"`
	SubAccountNameMinLength  uint64 `json:"subAccountNameMinLength"`
	SubAccountNameMaxLength  uint64 `json:"subAccountNameMaxLength"`
}

const MaxDescriptionLength uint64 = 255
