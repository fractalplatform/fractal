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

package blockchain

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/params"
)

func (g Genesis) MarshalJSON() ([]byte, error) {
	type genesisJSON struct {
		Config        *params.ChainConfig   `json:"config"`
		Dpos          *dpos.Config          `json:"dpos"`
		Nonce         math.HexOrDecimal64   `json:"nonce"`
		Timestamp     math.HexOrDecimal64   `json:"timestamp"`
		ExtraData     hexutil.Bytes         `json:"extraData"`
		GasLimit      math.HexOrDecimal64   `json:"gasLimit"`
		Difficulty    *math.HexOrDecimal256 `json:"difficulty"`
		Mixhash       common.Hash           `json:"mixHash"`
		Coinbase      common.Name           `json:"coinbase"`
		AllocAccounts []*GenesisAccount     `json:"allocAccounts"`
		AllocAssets   []*asset.AssetObject  `json:"allocAssets"`
	}
	var enc genesisJSON
	enc.Config = g.Config
	enc.Dpos = g.Dpos
	enc.Timestamp = math.HexOrDecimal64(g.Timestamp)
	enc.ExtraData = g.ExtraData
	enc.GasLimit = math.HexOrDecimal64(g.GasLimit)
	enc.Difficulty = (*math.HexOrDecimal256)(g.Difficulty)
	enc.Coinbase = g.Coinbase
	enc.AllocAccounts = g.AllocAccounts
	enc.AllocAssets = g.AllocAssets
	return json.Marshal(&enc)
}

func (g *Genesis) UnmarshalJSON(input []byte) error {
	type genesisJSON struct {
		Config        *params.ChainConfig   `json:"config"`
		Dpos          *dpos.Config          `json:"dpos"`
		Nonce         *math.HexOrDecimal64  `json:"nonce"`
		Timestamp     *math.HexOrDecimal64  `json:"timestamp"`
		ExtraData     *hexutil.Bytes        `json:"extraData"`
		GasLimit      *math.HexOrDecimal64  `json:"gasLimit"`
		Difficulty    *math.HexOrDecimal256 `json:"difficulty"`
		Mixhash       *common.Hash          `json:"mixHash"`
		Coinbase      common.Name           `json:"coinbase"`
		AllocAccounts []*GenesisAccount     `json:"allocAccounts"`
		AllocAssets   []*asset.AssetObject  `json:"allocAssets"`
	}
	var dec genesisJSON
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Config != nil {
		g.Config = dec.Config
	}
	if dec.Dpos != nil {
		g.Dpos = dec.Dpos
	}
	if dec.Timestamp != nil {
		g.Timestamp = uint64(*dec.Timestamp)
	}
	if dec.ExtraData != nil {
		g.ExtraData = *dec.ExtraData
	}
	if dec.GasLimit != nil {
		g.GasLimit = uint64(*dec.GasLimit)
	}
	if dec.Difficulty != nil {
		g.Difficulty = (*big.Int)(dec.Difficulty)
	}
	if len(dec.Coinbase) > 0 {
		g.Coinbase = dec.Coinbase
	}
	if len(dec.AllocAccounts) > 0 {
		g.AllocAccounts = dec.AllocAccounts
	}

	if len(dec.AllocAssets) > 0 {
		g.AllocAssets = dec.AllocAssets
	}
	return nil
}
