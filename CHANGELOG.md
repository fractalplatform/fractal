# [fractal](https://github.com/fractalplatform/fractal) Changelog
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


[0.0.9]: https://github.com/fractalplatform/fractal/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/fractalplatform/fractal/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/fractalplatform/fractal/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/fractalplatform/fractal/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/fractalplatform/fractal/commits/v0.0.5
