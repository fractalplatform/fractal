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

package runtime

import (
	"math"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/plugin"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types/envelope"
)

// Config is a basic type specifying certain configuration flags for running
// the EVM.
type Config struct {
	ChainConfig *params.ChainConfig
	Difficulty  *big.Int
	Origin      string
	FromPubkey  common.PubKey
	Coinbase    string
	BlockNumber *big.Int
	Time        *big.Int
	GasLimit    uint64
	AssetID     uint64
	GasPrice    *big.Int
	Value       *big.Int
	Debug       bool
	EVMConfig   vm.Config
	PM          plugin.IPM
	State       *state.StateDB
	GetHashFn   func(n uint64) common.Hash
}

// sets defaults on the config
func setDefaults(cfg *Config) {
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = params.DefaultChainconfig
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.Time == nil {
		cfg.Time = big.NewInt(time.Now().Unix())
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = math.MaxUint64
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = new(big.Int)
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	//cfg.Origin = common.BytesToAddress([]byte("sender"))
	if cfg.GetHashFn == nil {
		cfg.GetHashFn = func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		}
	}
}

//NewEnv create a new evm env
func NewEnv(cfg *Config) *vm.EVM {
	context := vm.Context{
		//CanTransfer: vm.CanTransfer,
		//Transfer:    vm.Transfer,
		GetHash: func(uint64) common.Hash { return common.Hash{} },

		Origin:      cfg.Origin,
		From:        cfg.Origin,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		AssetID:     cfg.AssetID,
		Difficulty:  cfg.Difficulty,
		GasLimit:    cfg.GasLimit,
		GasPrice:    cfg.GasPrice,
	}

	return vm.NewEVM(context, cfg.PM, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}

// Create executes the code using the EVM create method
func Create(action *envelope.ContractTx, cfg *Config) ([]byte, uint64, error) {
	if cfg == nil {
		cfg = new(Config)
	}
	setDefaults(cfg)

	var (
		vmenv  = NewEnv(cfg)
		sender = vm.AccountRef(cfg.Origin)
	)

	// Call the code with the given configuration.
	code, leftOverGas, err := vmenv.Create(
		sender,
		action,
		cfg.GasLimit,
	)
	return code, leftOverGas, err
}

// Call executes the code given by the contract's address. It will return the
// EVM's return value or an error if it failed.
//
// Call, unlike Execute, requires a config and also requires the State field to
// be set.
func Call(action *envelope.ContractTx, cfg *Config) ([]byte, uint64, error) {
	setDefaults(cfg)

	vmenv := NewEnv(cfg)

	sender := vm.AccountRef(cfg.Origin)
	// Call the code with the given configuration.
	ret, leftOverGas, err := vmenv.Call(
		sender,
		action,
		cfg.GasLimit,
	)

	return ret, leftOverGas, err
}
