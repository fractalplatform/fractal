### Forked
- [DPOS] allow contract asset transfer (#525)(#528)
- [FEE] other people pay transaction fee (#531)(#533)(#536)
### Fixed
- [FEE]fee transfer internal record (#495)
- [BLOCKCHIAN] fixed export blockchain error (#498)
- [GAS] modify gas price (#501)
- [MINER] add setcoinbase check (#500) and fix miner bug (#499)(#511)(#512)(#513)(#514)(#516)
- [P2P] fixed bug that may close a nil channel (#503)and fixed ddos check error (#519)
- [DOWNLOAD]add node into blacklist if it had too much errors(#519)(#523)
### Added
- [CMD] add version cmd compile date info (#505)(#521)
- [CMD] cmd/ft: add method 'seednodes' into sub-cmd 'p2p' (#497)
- [CMD] add txpool cmd gettxsbyaccount (#502)
- [P2P] p2p,rpc: add rpc to query seed nodes from db(#496)
- [TEST] add each code module unit test or note (#492)(#493)(#504)(#508)(#507)(#509)

