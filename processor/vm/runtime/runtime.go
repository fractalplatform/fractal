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
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
)

// Config is a basic type specifying certain configuration flags for running
// the EVM.
type Config struct {
	ChainConfig *params.ChainConfig
	Difficulty  *big.Int
	Origin      common.Name
	FromPubkey  common.PubKey
	Coinbase    common.Name
	BlockNumber *big.Int
	Time        *big.Int
	GasLimit    uint64
	AssetID     uint64
	GasPrice    *big.Int
	Value       *big.Int
	Debug       bool
	EVMConfig   vm.Config
	Account     *accountmanager.AccountManager
	State       *state.StateDB
	GetHashFn   func(n uint64) common.Hash
}

// sets defaults on the config
func setDefaults(cfg *Config) {
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = params.DefaultChainconfig
		//cfg.ChainConfig = &params.ChainConfig{
		//	ChainID:        big.NewInt(1),
		//	HomesteadBlock: new(big.Int),
		//	DAOForkBlock:   new(big.Int),
		//	DAOForkSupport: false,
		//	EIP150Block:    new(big.Int),
		//	EIP155Block:    new(big.Int),
		//	EIP158Block:    new(big.Int),
		//}
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

//create a new evm env
func NewEnv(cfg *Config) *vm.EVM {
	fmt.Println("in NewEnv ...")
	context := vm.Context{
		//CanTransfer: vm.CanTransfer,
		//Transfer:    vm.Transfer,
		GetHash: func(uint64) common.Hash { return common.Hash{} },
		GetDelegatedByTime: func(*state.StateDB, string, uint64) (*big.Int, error) {
			return big.NewInt(0), nil
		},

		//GetEpoch
		GetEpoch: func(state *state.StateDB, t uint64, epoch uint64) (peoch uint64, time uint64, err error) {
			return 2, 0, nil
		},
		//GetActivedCandidateSize
		GetActivedCandidateSize: func(state *state.StateDB, epoch uint64) (size uint64, err error) {
			return 3, nil
		},
		//GetActivedCandidate
		GetActivedCandidate: func(state *state.StateDB, epoch uint64, index uint64) (name string, stake *big.Int, totalVote *big.Int, counter uint64, actualCounter uint64, replace uint64, isbad bool, err error) {
			return "testname", big.NewInt(0), big.NewInt(3), 3, 3, 3, false, nil
		},

		//GetVoterStake
		GetVoterStake: func(state *state.StateDB, epoch uint64, voter string, candidate string) (stake *big.Int, err error) {
			return big.NewInt(9), nil
		},
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

	return vm.NewEVM(context, cfg.Account, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}

// Create executes the code using the EVM create method
func Create(action *types.Action, cfg *Config) ([]byte, uint64, error) {
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
func Call(action *types.Action, cfg *Config) ([]byte, uint64, error) {
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
