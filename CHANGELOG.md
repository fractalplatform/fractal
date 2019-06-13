# [fractal](https://github.com/fractalplatform/fractal) Changelog
## [0.0.20] - 2019-06-12
### Fixed
- [DOWNLOADER] fixed bug of find ancestor and use random station 
- [BLOCKCHAIN] fixed blockchain irreversible number
### Add
- [DPOS] add thread test for rand vote candidate
- [BLOCKCHAIN] add refuse bad block hashes
- [BLOCKCHAIN] sync block with a specified block number


## [0.0.19] - 2019-06-11
### Fixed
- [ASSET] modify subasset decimals


## [0.0.18] - 2019-06-06
### Fixed
- [ACCOUNT] modify children check function
### Add
- [CONTRACT] contract add getassetid api 
- [MINER] fix should counter & add delay duration for miner


## [0.0.17] - 2019-06-05
### Changed
- [GENESIS] modify blockchain sys account name
### Fixed
- [BLOCKCHAIN] modify blockchain.HasState function
- [RPC] fix GetDelegatedByTime rpc interface


## [0.0.16] - 2019-06-04
### Changed
- [MAKEFILE] fixed bug of target build_workspace
- [ACCOUNT] account author lenght should not exceed 10
- [VM] modify gas distribution
### Add
- [DPOS] add min available quantity of candidate for vote 
- [CMD] add read yaml and toml test 
- [SDK] add sdk contract test
- [TYPES] support parentIndex when sign
- [TXPOOL] add txpool resend pending txs
### Fixed
- [P2P] broadcast txs to atleast 3 peers 
- [BLOCKCHAIN] downloader disconnected peers which has to much wrong blocks
- [DPOS] fix calc should counter of candidate
- [ALL] fixs some bugs


## [0.0.15] - 2019-05-21
### Changed
- [VM] change withdraw type to transfer
### Add
- [P2P] add flow control,some quit channel
- [P2p] periodic remove the worst peer if peer connections is full, but default is disabled.
- [RPC] add dpos rpc api for info by epcho
### Fixed
- [DPOS] fix bug when dpos started
- [ALL] fixs some bugs


## [0.0.14] - 2019-05-20
### Fixed
- [GENESIS] fix genesis bootnodes prase failed not start node


## [0.0.13] - 2019-05-18
### Add
- [GPO] add add gas price oracle unit test 
- [VM] move gas to GasTableInstanse
### Fixed
- [PARAMS] change genesis gas limit to 30 million 
- [VM] opCreate doing nothing but push zero into stack and distributeGasByScale distribute right num
- [ACCOUNT] add check asset contract name, check account name length 
- [ALL] fixs some bugs


## [0.0.12] - 2019-05-13
### Add
- [CMD] add p2p miner txpool command.
### Deprecated
- [RPCAPI] modify account and blockchain return result
- [DOC] add jsonrpc, cmd, p2p docs in wiki


## [0.0.11] - 2019-05-06
### Deprecated
- [ASSET] modify asset and account action struct
- [ACCOUNT] modify account detail to description
- [DPOS] add dpos reward interface for solidity


## [0.0.10] - 2019-05-06
### Fixed
- [ASSET] get asset object return panic
- [VM] fix contract issue asset bug


## [0.0.9] - 2019-05-06
### Added
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


## [0.0.8] - 2019-04-30
### Added
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


## [0.0.7] - 2019-04-23
### Removed
- [WALLET] removed wallet module，the local node not support store private key
### Added
- [VM] add opt for opSnapBalance and destroyasset for contract
- [BLOCKCHAIN] support import/export block
- [RPC] add get the contract internal transaction
### Fixed
- [VM] add opt for opSnapBalance
- [TYPES] fixs the base types
- [ALL] fixs some bugs


## [0.0.6] - 2019-04-04
### Added
- [CRYPTO] add btcd secp256k1 crypto
### Fixed
- [MAKEFILE] fixed cross platform


## [0.0.5] - 2019-04-04
### Added
- [README] add license badge
- [SCRIPTS] add is_checkout_dirty.sh release.sh tag_release.sh commit_hash.sh
### Fixed
- [MAKEFILE] add check fmt tag_release release command


[0.0.20]: https://github.com/fractalplatform/fractal/compare/v0.0.19...v0.0.20
[0.0.19]: https://github.com/fractalplatform/fractal/compare/v0.0.18...v0.0.19
[0.0.18]: https://github.com/fractalplatform/fractal/compare/v0.0.17...v0.0.18
[0.0.17]: https://github.com/fractalplatform/fractal/compare/v0.0.16...v0.0.17
[0.0.16]: https://github.com/fractalplatform/fractal/compare/v0.0.15...v0.0.16
[0.0.15]: https://github.com/fractalplatform/fractal/compare/v0.0.14...v0.0.15
[0.0.14]: https://github.com/fractalplatform/fractal/compare/v0.0.13...v0.0.14
[0.0.13]: https://github.com/fractalplatform/fractal/compare/v0.0.12...v0.0.13
[0.0.12]: https://github.com/fractalplatform/fractal/compare/v0.0.11...v0.0.12
[0.0.11]: https://github.com/fractalplatform/fractal/compare/v0.0.10...v0.0.11
[0.0.10]: https://github.com/fractalplatform/fractal/compare/v0.0.9...v0.0.10
[0.0.9]: https://github.com/fractalplatform/fractal/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/fractalplatform/fractal/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/fractalplatform/fractal/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/fractalplatform/fractal/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/fractalplatform/fractal/commits/v0.0.5
