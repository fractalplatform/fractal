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

package types

import "github.com/fractalplatform/fractal/common"

type DetailTx struct {
	TxHash  common.Hash     `json:"txhash"`
	Actions []*DetailAction `json:"actions"`
}

type DetailAction struct {
	InternalActions []*InternalAction `json:"internalActions"`
}

type InternalAction struct {
	Action     *RPCAction `json:"action"`
	ActionType string     `json:"actionType"`
	GasUsed    uint64     `json:"gasUsed"`
	GasLimit   uint64     `json:"gasLimit"`
	Depth      uint64     `json:"depth"`
	Error      string     `json:"error"`
}

type BlockAndResult struct {
	Block     map[string]interface{} `json:"block"`
	Receipts  []*Receipt             `json:"receipts"`
	DetailTxs []*DetailTx            `json:"detailTxs"`
}
