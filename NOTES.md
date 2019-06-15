### Fixed
- [DOWNLOADER] fixed bug that may casue dead loop 
- [BLOCKCHAIN] fixed state store irreversible number bug
- [DPOS] fixed replace rate for candiate
### Removed
- [TXPOOL] removed some unused variable in txpool/handler.go
- [RPC] removed invalid code
### Added
- [TXPOOL] limited the amount of gorouting not greater 1024
- [GENESIS] add use default block gaslimit and update genesis.json

