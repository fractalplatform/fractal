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

