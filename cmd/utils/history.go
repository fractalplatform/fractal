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

package utils

import "github.com/monax/relic"

// Use below as template for change notes, delete empty sections but keep order
/*
### Security

### Changed

### Fixed

### Added

### Removed

### Deprecated
*/

// History the releases described by version string and changes, newest release first.
// The current release is taken to be the first release in the slice, and its
// version determines the single authoritative version for the next release.
//
// To cut a new release add a release to the front of this slice then run the
// release tagging script: ./scripts/tag_release.sh
var History relic.ImmutableHistory = relic.NewHistory("fractal", "https://github.com/fractalplatform/fractal").
	MustDeclareReleases(
		"0.0.15 - 2019-05-21",
		`### Changed
- [VM] change withdraw type to transfer
### Add
- [P2P] add flow control,some quit channel
- [P2p] periodic remove the worst peer if peer connections is full, but default is disabled.
- [RPC] add dpos rpc api for info by epcho
### Fixed
- [DPOS] fix bug when dpos started
- [ALL] fixs some bugs
`,
		"0.0.14 - 2019-05-20",
		`### Fixed
- [GENESIS] fix genesis bootnodes prase failed not start node
`,
		"0.0.13 - 2019-05-18",
		`### Add
- [GPO] add add gas price oracle unit test 
- [VM] move gas to GasTableInstanse
### Fixed
- [PARAMS] change genesis gas limit to 30 million 
- [VM] opCreate doing nothing but push zero into stack and distributeGasByScale distribute right num
- [ACCOUNT] add check asset contract name, check account name length 
- [ALL] fixs some bugs
`,
		"0.0.12 - 2019-05-13",
		`### Add
- [CMD] add p2p miner txpool command.
### Deprecated
- [RPCAPI] modify account and blockchain return result
- [DOC] add jsonrpc, cmd, p2p docs in wiki
`,
		"0.0.11 - 2019-05-06",
		`### Deprecated
- [ASSET] modify asset and account action struct
- [ACCOUNT] modify account detail to description
- [DPOS] add dpos reward interface for solidity
`,
		"0.0.10 - 2019-05-06",
		`### Fixed
- [ASSET] get asset object return panic
- [VM] fix contract issue asset bug
`,
		"0.0.9 - 2019-05-06",
		`### Added
- [BLOCKCHAIN] add gensis block account
- [FEE] the distributed gas will add to fractal.fee's balance
- [COMMON] add json unmarshal for author
- [ASSET] check valid for modifing about contract asset
### Fixed
- [VM] execWithdrawFee return err when fm.WithdrawFeeFromSystem fail
- [BLOCKCHAIN] fix fork contracl init err
- [GENESIS] genesis block action repeat
- [DPOS] fix updateElectedCandidates bug when dpos is false
- [ALL] fixs some bugs
### Changed
- [COMMON] modify name for support more scenes and modify subaccount/subasset name
- [ASSET] modify issue asset return assetID
`,
		"0.0.8 - 2019-04-30",
		`### Added
- [DEBUG] add debug pprof,trace cmd flags and rpc
- [FEE] add fee manager and some rpc interface
- [TXPOOL] add bloom in transaction P2P message
- [TYPES] types/action.go add remark field
### Fixed
- [TXPOOL] fixed txpool queue and pending don't remove no permissions transactions
- [VM] fix bug that distribute more gas than given when internal call happens
- [BLOCKCHAIN] fixed restart node missmatch genesis block hash
- [ACCOUNTMANAGER] generate author version when account create
- [DPOS] solve infinite loop for getvoters
- [ALL] fixs some bugs
`,
		"0.0.7 - 2019-04-23",
		`### Removed
- [WALLET] removed wallet module，the local node not support store private key
### Added
- [VM] add opt for opSnapBalance and destroyasset for contract
- [BLOCKCHAIN] support import/export block
- [RPC] add get the contract internal transaction
### Fixed
- [VM] add opt for opSnapBalance
- [TYPES] fixs the base types
- [ALL] fixs some bugs
`,
		"0.0.6 - 2019-04-04",
		`### Added
- [CRYPTO] add btcd secp256k1 crypto
### Fixed
- [MAKEFILE] fixed cross platform
`,
		"0.0.5 - 2019-04-04",
		`### Added
- [README] add license badge
- [SCRIPTS] add is_checkout_dirty.sh release.sh tag_release.sh commit_hash.sh
### Fixed
- [MAKEFILE] add check fmt tag_release release command
`,
	)
