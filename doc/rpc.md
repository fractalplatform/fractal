# RPC interface

- **RPC HTTP default endpoint**: http://localhost:8545
- **RPC interface list**：

  - **account**

    - [account_accountIsExist](#account_accountIsExist)
    - [account_getAccountByName](#account_getAccountByName)
    - [account_getCode](#account_getCode)
    - [account_getNonce](#account_getNonce)
    - [account_getAssetInfoByName](#account_getAssetInfoByName)
    - [account_getAssetInfoByID](#account_getAssetInfoByID)
    - [account_getAccountBalanceByID](#account_getAccountBalanceByID)

  - **ft**

    - [ft_sendRawTransaction](#ft_sendRawTransaction)
    - [ft_getTransactionByHash](#ft_getTransactionByHash)
    - [ft_getInternalTxByHash](#ft_getInternalTxByHash)
    - [ft_getTransactionReceipt](#ft_getTransactionReceipt)
    - [ft_getTxsByAccount](#ft_getTxsByAccount)
    - [ft_getInternalTxByAccount](#ft_getInternalTxByAccount)
    - [ft_getTxsByBloom](#ft_getTxsByBloom)
    - [ft_getInternalTxByBloom](#ft_getInternalTxByBloom)
    - [ft_getTransactions](#ft_getTransactions)
    - [ft_getCurrentBlock](#ft_getCurrentBlock)
    - [ft_getBlockByHash](#ft_getBlockByHash)
    - [ft_getBlockByNumber](#ft_getBlockByNumber)
    - [ft_getBlockAndResultByNumber](#ft_getBlockAndResultByNumber)
    - [ft_getBadBlocks](#ft_getBadBlocks)
    - [ft_gasPrice](#ft_gasprice)
    - [ft_call](#ft_call)
    - [ft_estimateGas](#ft_estimateGas)
    - [ft_getChainConfig](#ft_getChainConfig)
    - [ft_newFilter](#ft_newFilter)
    - [ft_getFilterChanges](#ft_getFilterChanges)
    - [ft_newPendingTransactionFilter](#ft_newPendingTransactionFilter)
    - [ft_newBlockFilter](#ft_newBlockFilter)
    - [ft_uninstallFilter](#ft_uninstallFilter)

  - **consensus**

    - [consensus_getAllCandidates](#consensus_getAllCandidates)
    - [consensus_getCandidateInfo](#consensus_getCandidateInfo)

  - **fee**

    - [fee_getObjectFeeByName](#fee_getObjectFeeByName)
    - [fee_getObjectFeeResult](#fee_getObjectFeeResult)

---

#### account_accountIsExist

Returns whether the account exists.

##### Parameters

- `String` - Name of the account.

##### Response

- `Boolean` - `true` when exist, otherwise `false`.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"account_accountIsExist","params":["fractalaccount"],"id":1}' http://localhost:8545

// Result

{
	"jsonrpc": "2.0",
	"id": 1,
	"result": true
}

```

---

#### account_getAccountByName

Get account information by name.

##### Parameters

- `String` - Name of the account.

##### Response

- `Object` - A account object,or `account not exist` if not found.
  - `accountName` - `String` name of the account.
  - `nonce` - `Quantity` integer of the number of transactions send from this account.
  - `code` - `String` the contract code from the given account.
  - `codeHash` - `String` hash of the contract code,or `0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470` if not found.
  - `codeSize` - `Quantity` size of the contract.
  - `balance` - `Array` all asset balances under the account.
    - `assetID` - `Quantity` id of asset.
    - `balance` - `Quantity` balance of asset.
  - `suicide` - `Boolean` `true` when account contract suicide, otherwise `false`.
  - `destroy` - `Boolean` `true` when account destroy, otherwise `false`.
  - `description` - `String` description of the account.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"account_getAccountByName","params":["fractalaccount"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "name": "fractalaccount",
        "address": "0x3f17f1962b36e491b30a40b2405849e597ba5fb5",
        "nonce": 0,
        "code": "0x",
        "codeHash": "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        "codeSize": 0,
        "balance": {
            "assetID": 0,
            "balance": 0
        },
        "suicide": false,
        "destroy": false,
        "description": "account manager account"
    }
}
```

---

#### account_getCode

Returns contract code by account name

##### Parameters

- `String` - Name of the account.

##### Response

- `String` - the code from the given address,or `code is empty` if not found.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"account_getCode","params":["tcontract1"],"id":1}' http://localhost:8545

// Result

{
	"jsonrpc": "2.0",
	"id": 1,
	"result": "0x900463ffff......"
}

```

---

#### account_getNonce

Returns nonce by account name

##### Parameters

- `String` - Name of the account.

##### Response

- `Quantity` - integer of the number of transactions send from this account.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"account_getNonce","params":["testtransfer"],"id":1}' http://localhost:8545

// Result

{
	"jsonrpc": "2.0",
	"id": 1,
	"result": 598
}

```

---

#### account_getAssetInfoByName

Get asset information by asset name.

##### Parameters

- `String` - name of the asset.

##### Response

- `Object` - A asset object,`asset not exist` if not found
  - `assetId` - `Quantity` id of the asset.
  - `assetName` - `String` name of the asset.
  - `symbol` - `String` symbol of asset.
  - `amount` - `Quantity` total currency amount of asset.
  - `decimals` - `Quantity` minimum unit of asset.
  - `founder` - `String` who created the asset.
  - `owner` - `String` who owns all permissions to the asset.
  - `addIssue` - `Quantity` total amount of issuance.
  - `upperLimit` - `Quantity` maximum number of additional issues.
  - `description` - `String` description of the asset.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"account_getAssetInfoByName","params":["ftoken"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "assetId": 0,
        "assetName": "ftoken",
        "symbol": "ft",
        "amount": 10000000000000000000000000000,
        "decimals": 18,
        "founder": "fractalfounder",
        "owner": "fractalfounder",
        "addIssue": 10000000000000000000000000000,
        "upperLimit": 10000000000000000000000000000,
        "description": ""
    }
}

```

---

#### account_getAssetInfoByID

Get asset information by asset ID.

##### Parameters

- `Quantity` - id of the asset.

##### Response

See [account_getAssetInfoByName](#account_getAssetInfoByName) reponse.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"account_getAssetInfoByID","params":[0],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "assetId": 0,
        "assetName": "ftoken",
        "symbol": "ft",
        "amount": 10000000000000000000000000000,
        "decimals": 18,
        "founder": "fractalfounder",
        "owner": "fractalfounder",
        "addIssue": 10000000000000000000000000000,
        "upperLimit": 10000000000000000000000000000,
        "description": ""
    }
}

```

---

#### account_getAccountBalanceByID

Get account balance by account name and asset id.

##### Parameters

- `String` - Name of the asset.
- `Quantity` - id of the asset.

##### Response

- `Quantity` - account current balance.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"account_getAccountBalanceByID","params":["qqqqqqqqqqq1", 0],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": 9997000000336529599999999999
}

```

---

#### ft_sendRawTransaction

Creates new message call transaction for signed transactions.

##### Parameters

- `String` - The signed transaction data.

##### Response

- `String` - the transaction hash, or the zero hash if the transaction is not yet available.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_sendRawTransaction","params":["0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"],"id":1}' http://localhost:8545

// Result

{
    "id":1,
    "jsonrpc": "2.0",
    "result": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331"
}

```

---

#### ft_getTransactionByHash

Returns the information about a transaction requested by transaction hash.

##### Parameters

- `String` - hash of the transaction

##### Response

- `Object` A transaction object, or `null` when no transaction was found.
  - `blockHash` - `String` hash of the block where this transaction was in. `null` when its pending.
  - `blockNumber` - `Quantity` block number where this transaction was in. `0` when its pending.
  - `txHash` - `String` hash of the transaction.
  - `transactionIndex` - `Quantity` integer of the transaction's index position in the block. `0` when its pending.
  - `actions` - `Array` list of action.
    - `type` - `Quantity` type of action.
    - `nonce` - `Quantity` the number of actions made by the sender prior to this one.
    - `from` - `String` account name of the sender.
    - `to` - `String` account name of the sender.
    - `assetID` - `Quantity` which asset to transfer.
    - `gas` - `Quantity` gas provided by the sender.
    - `value` - `Quantity` value transferred.
    - `remark` - `String` the data of extra context.
    - `payload` - `String` the data send along with the action.
    - `actionHash` - `String` hash of the action.
    - `actionIndex` - `Quantity` integer of the action's index position in the transaction.
  - `gasAssetID` - `Quantity` which asset to pay gas.
  - `gasPrice` - `Quantity` gas price provided by the sender.
  - `gasCost` - `Quantity` the number of transaction gas cost.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"ft_getTransactionByHash","params":["0xeb701c7159e270f3ae9f209951ded94896ced101a0bb9c6b40729e3d57d1067b"],"id": 1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "blockHash": "0x8f0964a37ca83d9bb69fb2a6f4ad0c394d75d045dbd954435cea2ba645d679bd",
        "blockNumber": 2349339,
        "txHash": "0xeb701c7159e270f3ae9f209951ded94896ced101a0bb9c6b40729e3d57d1067b",
        "transactionIndex": 1,
        "actions": [
            {
                "type": 517,
                "nonce": 1617077,
                "from": "alexander",
                "to": "nicholas",
                "assetID": 0,
                "gas": 1100000,
                "value": 1000000000000000,
                "remark": "0x68656c6c6f206e6963686f6c6173207e",
                "payload": "0x",
                "actionHash": "0x1bb321a0afe65eed2ed7698ffd94bfd2c559a9c4494ebce4c648bdc379750d7c",
                "actionIndex": 0
            }
        ],
        "gasAssetID": 0,
        "gasPrice": 1000000000,
        "gasCost": 1100000000000000
    }
}

```

---

#### ft_getInternalTxByHash

Returns the information about a internal transaction requested by transaction hash.
_Note:_ That the internal transaction is available for use flag `--contractlog=true`.

##### Parameters

- `String` - hash of the transaction

##### Response

- `Object` - A internal transaction object, or `null` when no internal transaction was found.
  - `txHash` - `String` hash of the transaction.
  - `actions` - `Array` list of action.
    - `internalActions` - `Array` list of internal action.
      - `action` - `Object` a internal action object,or `null` when no internal action was found.
        - `type` - `Quantity` type of action.
        - `nonce` - `Quantity` the number of actions made by the sender prior to this one.
        - `from` - `String` account name of the sender.
        - `to` - `String` account name of the sender.
        - `assetID` - `Quantity` which asset to transfer.
        - `gas` - `Quantity` gas provided by the sender.
        - `value` - `Quantity` value transferred.
        - `remark` - `String` the data of extra context.
        - `payload` - `String` the data send along with the action.
        - `actionHash` - `String` hash of the action.
        - `actionIndex` - `Quantity` integer of the action's index position in the transaction.
      - `actionType` - `Quantity` type of internal action.
      - `gasAssetID` - `Quantity` which asset to pay gas.
      - `gasPrice` - `Quantity` gas price provided by the sender.
      - `error` - `Quantity` error of internal action.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getInternalTxByHash","params":["0x3b8d4a177058d16fe2324cd3c442da723470fff1889c563a27555c9cb5f14adb"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "txhash": "0x3b8d4a177058d16fe2324cd3c442da723470fff1889c563a27555c9cb5f14adb",
        "actions": [
            {
                "internalActions": [
                    {
                        "action": {
                            "type": 0,
                            "nonce": 0,
                            "from": "tcontract5",
                            "to": "tcontract5",
                            "assetID": 0,
                            "gas": 0,
                            "value": 100000000000000,
                            "remark": "0x",
                            "payload": "0x",
                            "actionHash": "0x95057f1c7cdd23eb1a9cef122f6cdd90d87527469d7417212784fd10b3790e95",
                            "actionIndex": 0
                        },
                        "actionType": "transferex",
                        "gasUsed": 0,
                        "gasLimit": 0,
                        "depth": 1,
                        "error": ""
                    }
                ]
            }
        ]
    }
}

```

---

#### ft_getTransactionReceipt

Returns the receipt of a transaction by transaction hash.

_Note:_ That the receipt is not available for pending transactions.

##### Parameters

- `String` - hash of the transaction

##### Response

- `Object` -A receipt object, or `null` when no receipt was found.
  - `blockHash` - `String` hash of the block where this transaction was in.
  - `blockNumber` - `Quantity` block number where this transaction was in.
  - `txHash` - `String` hash of the transaction.
  - `transactionIndex` - `Quantity` integer of the transaction's index position in the block.
  - `postState` - `String` post transaction state root.
  - `actionResults` - `Array` results of action.
    - `actionType` - `Quantity` type of action.
    - `status` - `Quantity` either 1 (success) or 0 (failure).
    - `index` - `Quantity` integer of the action's index position in the transaction.
    - `gasUsed` - `Quantity`the amount of gas used by this specific action alone.
    - `gasAllot` - `Array` information of gas allot.
      - `name` - `String` name of account.
      - `gas` - `Quantity` integer of gas reward.
      - `typeId` - `Quantity` `0` transfer fee ,`1` contract ,`2` block reward.
    - `error` - `String` error of action.
  - `cumulativeGasUsed` - `Quantity` the total amount of gas used when this transaction was executed in the block.
  - `totalGasUsed` - `Quantity` the amount of gas used by this specific transaction alone.
  - `logsBloom` - `String` bloom filter for light clients to quickly retrieve related logs.
  - `logs` - `Array` array of log objects, which this transaction generated，
    - `name` - `String` name from which this log originated.
    - `topics` - `Array` data of indexed log arguments. (In solidity: The first topic is the hash of the signature of the event.)
    - `data` - `String` contains the non-indexed arguments of the log.
    - `blockNumber` - `Quantity` block number where this log was in.
    - `blockHash` - `String` hash of the block where this log was in.
    - `transactionHash` - `String` hash of the transaction where this log was in.
    - `logIndex` - `Quantity` integer of the log's index position in the action.
    - `actionIndex` - `Quantity` integer of the action's index position in the transaction.
    - `transactionIndex` - `Quantity` integer of the transaction's index position in the block.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"ft_getTransactionReceipt","params":["0x01c746c5a181088c14e7e6f43739d0dc3db423f6b2cfa102d4549f338514a804"],"id": 1}' http://localhost:8545

// Result

{
    "jsonrpc":"2.0",
    "id":1,
    "result":{
        "blockHash":"0x5b033934948e94221cb14700107031d033a55797d24397d62edf04a35e43273e",
        "blockNumber":512352,
        "txHash":"0x01c746c5a181088c14e7e6f43739d0dc3db423f6b2cfa102d4549f338514a804",
        "transactionIndex":30,
        "postState":"0xecd76a303d72ab65ca8d6708e4c473d02b41d4315ef487780ff79424c4cd40be",
        "actionResults":[
            {
                "actionType":0,
                "status":1,
                "index":0,
                "gasUsed":342999,
                "gasAllot":[
                    {
                        "name":"testtesttest41",
                        "gas":171562,
                        "typeId":1
                    },
                    {
                        "name":"testtesttest1001",
                        "gas":88437,
                        "typeId":1
                    },
                    {
                        "name":"testtesttest13",
                        "gas":68600,
                        "typeId":2
                    },
                    {
                        "name":"testtesttest45:a1744",
                        "gas":14400,
                        "typeId":0
                    }
                ],
                "error":""
            }
        ],
        "cumulativeGasUsed":10632969,
        "totalGasUsed":342999,
        "logsBloom":"0x00000000000080000000000000000000000200000000100000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "logs":[
            {
                "name":"testtesttest1001",
                "topics":[
                    "0xe4ea926e8921624f76e136a0feb603b4cf0fa652df5b3752c00cc24f711b3f57"
                ],
                "data":"0x00000000000000000000000000000000000000000000000015bb067628b1600000000000000000000000000000000000000000000000000000000000000006d60000000000000000000000000000000000000000000000056bc75e2d6310000000000000000000000000000000000000000000000000043116cf411ae8ea569000000000000000000000000000000000000000001ffd168b615cf58e2c000000",
                "blockNumber":512352,
                "blockHash":"0x5b033934948e94221cb14700107031d033a55797d24397d62edf04a35e43273e",
                "transactionHash":"0x01c746c5a181088c14e7e6f43739d0dc3db423f6b2cfa102d4549f338514a804",
                "logIndex":30,
                "actionIndex":0,
                "transactionIndex":30
            }
        ]
    }
}

```

---

#### ft_getTxsByAccount

Returns an array of transaction hashes matching a given account in range of blocks.

##### Parameters

- `String` - name of the account.
- `Quantity` - integer of the block start search.
- `Quantity` - integer of backward blocks.

##### Response

- `Object` -A hashes object.
  - `tx` - `Array` list of transaction hash.
    - `hash` - `String` hash of transaction.
    - `height` - `Quantity` block number where this transaction was in.
  - `irreversibleBlockHeight` - `Quantity` dpos irreversible block number.
  - `endHeight` - `Quantity` end block number.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getTxsByAccount","params":["walletservice.u9","2351865",1048],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "txs": [
            {
                "hash": "0x33983ebaf2387850fd26637198a1cb29462a82853f42c045e24d86b09e67a293",
                "height": 2351865
            }
        ],
        "irreversibleBlockHeight": 2356865,
        "endHeight": 2351993
    }
}

```

---

#### ft_getInternalTxByAccount

Returns an array of internal transaction matching a given account in range of blocks.

_Note:_ That the internal transaction is available for use flag `--contractlog=true`.

##### Parameters

- `String` - name of the account.
- `Quantity` - integer of the block start search.
- `Quantity` - integer of forward blocks.

##### Response

- `Array` - array of internal actions

  - see [ft_getInternalTxByHash](#ft_getInternalTxByHash) response

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getInternalTxByAccount","params":["tcontract5",14563,10],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc":"2.0",
    "id":1,
    "result":[
        {
            "txhash":"0x3b8d4a177058d16fe2324cd3c442da723470fff1889c563a27555c9cb5f14adb",
            "actions":[
                {
                    "internalActions":[
                        {
                            "action":{
                                "type":0,
                                "nonce":0,
                                "from":"tcontract5",
                                "to":"tcontract5",
                                "assetID":0,
                                "gas":0,
                                "value":100000000000000,
                                "remark":"0x",
                                "payload":"0x",
                                "actionHash":"0x95057f1c7cdd23eb1a9cef122f6cdd90d87527469d7417212784fd10b3790e95",
                                "actionIndex":0
                            },
                            "actionType":"transferex",
                            "gasUsed":0,
                            "gasLimit":0,
                            "depth":1,
                            "error":""
                        }
                    ]
                }
            ]
        }
    ]
}

```

---

#### ft_getTxsByBloom

Returns an array of transaction hashes matching a given bloom in range of blocks.

##### Parameters

- `String` - bloom hex string.
- `Quantity` - integer of the block start search.
- `Quantity` - integer of backward blocks.

##### Response

See [ft_getTxsByAccount](#ft_getTxsByAccount) parameters

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getTxsByBloom","params":["0x013...",14563,10],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "txs": [
            {
                "hash": "0x33983ebaf2387850fd26637198a1cb29462a82853f42c045e24d86b09e67a293",
                "height": 2351865
            }
        ],
        "irreversibleBlockHeight": 2356865,
        "endHeight": 2351993
    }
}


```

---

#### ft_getInternalTxByBloom

Returns an array of internal transaction matching a given bloom in range of blocks.

_Note:_ That the internal transaction is available for use flag `--contractlog=true`.

##### Parameters

- `String` - bloom hex string.
- `Quantity` - integer of the block start search.
- `Quantity` - integer of forward blocks.

##### Response

- `Array` - array of internal actions

  - see [ft_getInternalTxByHash](#ft_getInternalTxByHash) response

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getInternalTxByBloom","params":["0x013...",10,5],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc":"2.0",
    "id":1,
    "result":[
        {
            "txhash":"0x3b8d4a177058d16fe2324cd3c442da723470fff1889c563a27555c9cb5f14adb",
            "actions":[
                {
                    "internalActions":[
                        {
                            "action":{
                                "type":0,
                                "nonce":0,
                                "from":"tcontract5",
                                "to":"tcontract5",
                                "assetID":0,
                                "gas":0,
                                "value":100000000000000,
                                "remark":"0x",
                                "payload":"0x",
                                "actionHash":"0x95057f1c7cdd23eb1a9cef122f6cdd90d87527469d7417212784fd10b3790e95",
                                "actionIndex":0
                            },
                            "actionType":"transferex",
                            "gasUsed":0,
                            "gasLimit":0,
                            "depth":1,
                            "error":""
                        }
                    ]
                }
            ]
        }
    ]
}

```

---

#### ft_getTransactions

Returns transactions by hashes.

##### Parameters

- `Array` - list of transaction hash,maximum 2048.

##### Response

- `Array` - list of transaction

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d `{"jsonrpc":"2.0","method":"ft_getTransactions","params":[["0x068f322f62edf4686c65060ce2d7c78d1404f391950e4ea93588d9d4cf4f8724",
"0x82f7482c529e05909792911911656cd9b52ccc35d1a2d3d93130fac960dd2b48"]],"id":1}` http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": [
        {
            "blockHash": "0x8826a262034ea5adcdc07b802999d759e80c64d4fcfa465ecc3cb5cd36b46cc3",
            "blockNumber": 2359076,
            "txHash": "0x068f322f62edf4686c65060ce2d7c78d1404f391950e4ea93588d9d4cf4f8724",
            "transactionIndex": 0,
            "actions": [
                {
                    "type": 517,
                    "nonce": 1626507,
                    "from": "nicholas",
                    "to": "alexander",
                    "assetID": 0,
                    "gas": 1100000,
                    "value": 1000000000000000,
                    "remark": "0x68656c6c6f20616c6578616e646572207e",
                    "payload": "0x",
                    "actionHash": "0xa10ee0ddd109c9e9f779642028c7e0d46d152d7868dacb9872941f871f03954a",
                    "actionIndex": 0
                }
            ],
            "gasAssetID": 0,
            "gasPrice": 1000000000,
            "gasCost": 1100000000000000
        },
        {
            "blockHash": "0x8826a262034ea5adcdc07b802999d759e80c64d4fcfa465ecc3cb5cd36b46cc3",
            "blockNumber": 2359076,
            "txHash": "0x82f7482c529e05909792911911656cd9b52ccc35d1a2d3d93130fac960dd2b48",
            "transactionIndex": 1,
            "actions": [
                {
                    "type": 517,
                    "nonce": 1625267,
                    "from": "alexander",
                    "to": "nicholas",
                    "assetID": 0,
                    "gas": 1100000,
                    "value": 1000000000000000,
                    "remark": "0x68656c6c6f206e6963686f6c6173207e",
                    "payload": "0x",
                    "actionHash": "0xe3295a7ffaa3812c2445ea9bbce71a4d1d84593aba2edeae8c0d8de5a2d73b97",
                    "actionIndex": 0
                }
            ],
            "gasAssetID": 0,
            "gasPrice": 1000000000,
            "gasCost": 1100000000000000
        }
    ]
}

```

---

#### ft_getCurrentBlock

Returns information about the last block.

##### Parameters

- `Boolean` - If `true` it returns the full transaction objects, if `false` only the hashes of the transactions.

##### Response

- `Object` A block object.
  - `difficulty` - `Quantity` integer of the difficulty for this block.
  - `extraData` - `String` the `extra data` field of this block.
  - `forkID` - `Object` a fock object.
    - `cur` - `Quantity` current fork id.
    - `next` - `Quantity` next
  - `gasLimit`- `Quantity` the maximum gas allowed in this block.
  - `gasUsed` - `Quantity` the total used gas by all transactions in this block.
  - `hash` - `String` hash of block.
  - `logsBloom` - `String` the bloom filter for the logs of the block.
  - `miner` - `String` the name of the beneficiary to whom the mining rewards were given.
  - `number` - `Quantity` the block number.
  - `parentHash` - `String` hash of the parent block.
  - `proposedIrreversible` - `Quantity` irreversible block number.
  - `receiptsRoot` - `String` the root of the receipts trie of the block.
  - `size` - `Quantity` integer the size of this block in bytes.
  - `stateRoot` - `String` the root of the transaction trie of the block.
  - `timestamp` - `Quantity` the unix timestamp for when the block was collated.
  - `totalDifficulty` - `Quantity` integer of the total difficulty of the chain until this block.
  - `transactions` - `Array` Array of transaction objects, or 32 Bytes transaction hashes depending on the last given parameter.
  - `transactionsRoot` - `String` the root of the transaction trie of the block.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getCurrentBlock","params":[true],"id":1}' http://localhost:8545


// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "difficulty": 2399545,
        "extraData": "0x73797374656d6ae7eb673de85c31d44fd4c50ae05baa453ac33207f07067442097fc068c37e554d1a42e5765eadae4a6070e8685ccbe9be21391eb388475487112219419991800",
        "forkID": {
            "cur": 3,
            "next": 3
        },
        "gasLimit": 30000000,
        "gasUsed": 202244,
        "hash": "0x8826a262034ea5adcdc07b802999d759e80c64d4fcfa465ecc3cb5cd36b46cc3",
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "miner": "vanessaconnie",
        "number": 2359076,
        "parentHash": "0x149ac9d8ca31671243cc471ee1830a59ad9bb0a8c5090fca4aba073cea98ee03",
        "proposedIrreversible": 2358991,
        "receiptsRoot": "0xb8468af0562104b0fa7f64ea3bfe0e195fe8f538efe4610d16e6e77cfd8c7e2d",
        "size": 813,
        "stateRoot": "0x83da82455c0cf63b3ac815a708efc5b6cb0d2af8ba6acbfd0494b6714ca42ac7",
        "timestamp": 1566819432000000000,
        "totalDifficulty": 2872716285712,
        "transactions": [
            "0x068f322f62edf4686c65060ce2d7c78d1404f391950e4ea93588d9d4cf4f8724",
            "0x82f7482c529e05909792911911656cd9b52ccc35d1a2d3d93130fac960dd2b48"
        ],
        "transactionsRoot": "0xc2a6e70010fedf16349719f0e246365ba41b951f7a8ee0218d9fa050434c4573"
    }
}
```

---

#### ft_getBlockByHash

Returns information about a block by hash.

##### Parameters

- `String` - hash of the block.

- `Boolean` - If `true` it returns the full transaction objects, if `false` only the hashes of the transactions.

##### Response

See [ft_getCurrentBlock](#ft_getCurrentBlock) response

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getBlockByHash","params":["0x8826a262034ea5adcdc07b802999d759e80c64d4fcfa465ecc3cb5cd36b46cc3", true],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "difficulty": 2399545,
        "extraData": "0x73797374656d6ae7eb673de85c31d44fd4c50ae05baa453ac33207f07067442097fc068c37e554d1a42e5765eadae4a6070e8685ccbe9be21391eb388475487112219419991800",
        "forkID": {
            "cur": 3,
            "next": 3
        },
        "gasLimit": 30000000,
        "gasUsed": 202244,
        "hash": "0x8826a262034ea5adcdc07b802999d759e80c64d4fcfa465ecc3cb5cd36b46cc3",
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "miner": "vanessaconnie",
        "number": 2359076,
        "parentHash": "0x149ac9d8ca31671243cc471ee1830a59ad9bb0a8c5090fca4aba073cea98ee03",
        "proposedIrreversible": 2358991,
        "receiptsRoot": "0xb8468af0562104b0fa7f64ea3bfe0e195fe8f538efe4610d16e6e77cfd8c7e2d",
        "size": 813,
        "stateRoot": "0x83da82455c0cf63b3ac815a708efc5b6cb0d2af8ba6acbfd0494b6714ca42ac7",
        "timestamp": 1566819432000000000,
        "totalDifficulty": 2872716285712,
        "transactions": [
            "0x068f322f62edf4686c65060ce2d7c78d1404f391950e4ea93588d9d4cf4f8724",
            "0x82f7482c529e05909792911911656cd9b52ccc35d1a2d3d93130fac960dd2b48"
        ],
        "transactionsRoot": "0xc2a6e70010fedf16349719f0e246365ba41b951f7a8ee0218d9fa050434c4573"
    }
}

```

---

#### ft_getBlockByNumber

Returns information about a block by number.

##### Parameters

- `Quantity` - number of the block.

- `Boolean` - If `true` it returns the full transaction objects, if `false` only the hashes of the transactions.

##### Response

See [ft_getCurrentBlock](#ft_getCurrentBlock) response

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getBlockByNumber","params":[2359076, false],"id":1}' http://localhost:8545


// Result


{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "difficulty": 2399545,
        "extraData": "0x73797374656d6ae7eb673de85c31d44fd4c50ae05baa453ac33207f07067442097fc068c37e554d1a42e5765eadae4a6070e8685ccbe9be21391eb388475487112219419991800",
        "forkID": {
            "cur": 3,
            "next": 3
        },
        "gasLimit": 30000000,
        "gasUsed": 202244,
        "hash": "0x8826a262034ea5adcdc07b802999d759e80c64d4fcfa465ecc3cb5cd36b46cc3",
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "miner": "vanessaconnie",
        "number": 2359076,
        "parentHash": "0x149ac9d8ca31671243cc471ee1830a59ad9bb0a8c5090fca4aba073cea98ee03",
        "proposedIrreversible": 2358991,
        "receiptsRoot": "0xb8468af0562104b0fa7f64ea3bfe0e195fe8f538efe4610d16e6e77cfd8c7e2d",
        "size": 813,
        "stateRoot": "0x83da82455c0cf63b3ac815a708efc5b6cb0d2af8ba6acbfd0494b6714ca42ac7",
        "timestamp": 1566819432000000000,
        "totalDifficulty": 2872716285712,
        "transactions": [
            "0x068f322f62edf4686c65060ce2d7c78d1404f391950e4ea93588d9d4cf4f8724",
            "0x82f7482c529e05909792911911656cd9b52ccc35d1a2d3d93130fac960dd2b48"
        ],
        "transactionsRoot": "0xc2a6e70010fedf16349719f0e246365ba41b951f7a8ee0218d9fa050434c4573"
    }
}

```

---

#### ft_getBlockAndResultByNumber

Returns information about a block and receipts and internal transaction by number.

##### Parameters

- `Quantity` - number of the block.

##### Response

- `Object` - a object contain block, transaction, receipts,internal transaction.
  - `block` see [ft_getCurrentBlock](#ft_getCurrentBlock) response
  - `transaction` see [ft_getTransactionByHash](#ft_getTransactionByHash) response
  - `receipts` see [ft_getTransactionReceipt](#ft_getTransactionReceipt) response
  - `internal transaction` see [ft_getInternalTxByHash](#ft_getInternalTxByHash) response

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"ft_getBlockAndResultByNumber","params":["14563"],"id": 1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "block": {
            "difficulty": 51848,
            "extraData": "0x73797374656d92b346a6fe5253d666ef8813b5a8e7991c46cfc8c29fe50d2b81416f03718d6443665ae174855a7a5966668b608db4389d8b07600b3b09411e6fc78805e166f900",
            "forkID": {
                "cur": 0,
                "next": 0
            },
            "gasLimit": 30000000,
            "gasUsed": 211180,
            "hash": "0x4529b9c8faf0723e3889a3bfabfc0e86c5365ab076283286ba192b814f826ef6",
            "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "miner": "fractal.founder",
            "number": 14563,
            "parentHash": "0x4a920a85cb113474edab5b06246c3dda6f3311309636239861f782bfae597e5a",
            "proposedIrreversible": 14562,
            "receiptsRoot": "0xdcdb947919c79275673912d2f37f74b851aa79ef646d40b15865b3767ce37d2a",
            "size": 742,
            "stateRoot": "0x102f61ad25684178fcca6a63f91ab130d5656fba6d46471f1e0b7f573cd08743",
            "timestamp": 1559776341000000000,
            "totalDifficulty": 649029222,
            "transactions": [
                {
                    "blockHash": "0x4529b9c8faf0723e3889a3bfabfc0e86c5365ab076283286ba192b814f826ef6",
                    "blockNumber": 14563,
                    "txHash": "0x3b8d4a177058d16fe2324cd3c442da723470fff1889c563a27555c9cb5f14adb",
                    "transactionIndex": 0,
                    "actions": [
                        {
                            "type": 0,
                            "nonce": 69,
                            "from": "tcontract5",
                            "to": "tcontract5",
                            "assetID": 0,
                            "gas": 20000000,
                            "value": 0,
                            "remark": "0x",
                            "payload": "0x68669bc8000000000000000000000000000000000000000000000000000000000000103e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005af3107a4000",
                            "actionHash": "0x37db3c063984b5dc8ece446b1a74ea328099a66901124abd45ffada61b27b9f4",
                            "actionIndex": 0
                        }
                    ],
                    "gasAssetID": 0,
                    "gasPrice": 10000000000,
                    "gasCost": 200000000000000000
                }
            ],
            "transactionsRoot": "0xf3dbde4ef4f96e3befd7c1fbbb31c19db0d0680f746bfc187f29d60247feadff"
        },
        "receipts": [
            {
                "PostState": "EC9hrSVoQXj8ympj+RqxMNVlb7ptRkcfHgt/VzzQh0M=",
                "ActionResults": [
                    {
                        "Status": 1,
                        "Index": 0,
                        "GasUsed": 211180,
                        "GasAllot": [
                            {
                                "name": "ftoken",
                                "gas": 7200,
                                "typeId": 0
                            },
                            {
                                "name": "tcontract5",
                                "gas": 161743,
                                "typeId": 1
                            },
                            {
                                "name": "fractal.founder",
                                "gas": 42237,
                                "typeId": 2
                            }
                        ],
                        "Error": ""
                    }
                ],
                "CumulativeGasUsed": 211180,
                "Bloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
                "Logs": [],
                "TxHash": "0x3b8d4a177058d16fe2324cd3c442da723470fff1889c563a27555c9cb5f14adb",
                "TotalGasUsed": 211180
            }
        ],
        "detailTxs": [
            {
                "txhash": "0x3b8d4a177058d16fe2324cd3c442da723470fff1889c563a27555c9cb5f14adb",
                "actions": [
                    {
                        "internalActions": [
                            {
                                "action": {
                                    "type": 0,
                                    "nonce": 0,
                                    "from": "tcontract5",
                                    "to": "tcontract5",
                                    "assetID": 0,
                                    "gas": 0,
                                    "value": 100000000000000,
                                    "remark": "0x",
                                    "payload": "0x",
                                    "actionHash": "0x95057f1c7cdd23eb1a9cef122f6cdd90d87527469d7417212784fd10b3790e95",
                                    "actionIndex": 0
                                },
                                "actionType": "transferex",
                                "gasUsed": 0,
                                "gasLimit": 0,
                                "depth": 1,
                                "error": ""
                            }
                        ]
                    }
                ]
            }
        ]
    }
}

```

---

#### ft_getBadBlocks

Returns dose not insert blockchain blocks.

##### Parameters

- `Boolean` - If `true` it returns the full transaction objects, if `false` only the hashes of the transactions.

##### Response

- `Array` - list of blocks, or `null` if not found.
  - `block` - see [ft_getCurrentBlock](#ft_getCurrentBlock) response

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"ft_getBadBlocks","params":[true],"id": 1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": [
        {
            "difficulty": 51970,
            "extraData": "0x73797374656daa789d465a5b636c6d0bb4707f18ab98c31f6e578d25d118c8311a739daa8f9e0a45192161542118a6a35f0983a9716678d0a33a7c29ac027800f17ecec1b2dd00",
            "forkID": {
                "cur": 0,
                "next": 0
            },
            "gasLimit": 30000000,
            "gasUsed": 0,
            "hash": "0x162e2d1d9fc5a2bf31955712aa7f6e7d9054b85dd8e75ee72a6ce798f98845b4",
            "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "miner": "fractal.founder",
            "number": 14685,
            "parentHash": "0x83587a8469365c9de64c74e46a2588bc323c58a83ad7021c3a5743610e06150e",
            "proposedIrreversible": 14684,
            "receiptsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
            "size": 514,
            "stateRoot": "0xd54de90b55e5b3f628dc32843a502e93489701edb166a09b8415e7142de61acd",
            "timestamp": 1559776707000000000,
            "totalDifficulty": 655362181,
            "transactions": [],
            "transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000"
        }
    ]
}

```

---

#### ft_gasPrice

Returns the current price per gas.

##### Parameters None

##### Response

`Quantity` - integer of the current gas price.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"ft_gasPrice","params":[],"id": 1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": 1011103930
}

```

---

#### ft_call

Executes a new message call immediately without creating a transaction on the block chain.

##### Parameters

- `Object` - The transaction call object
  - `actionType` - `Quantity` type of transaction.
  - `from` - `String` name of sender account.
  - `to` - `String` name of receipt account.
  - `assetId` - `Quantity` id of used asset.
  - `gas` - `Quantity` integer of the gas provided for the transaction execution. eth_call consumes zero gas, but this parameter may be needed by some executions.
  - `gasPrice` - `Quantity` integer of the gasPrice used for each paid gas.
  - `value` - `Quantity` integer of the value sent with this transaction.
  - `data` - `String` hash of the method signature and encoded parameters.
  - `remark` - `String` extra data with this transaction. -`Quantity` - height of the block, or string value `latest` or `earliest`.

##### Response

`String` - the return value of executed contract.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"id":1,"jsonrpc":"2.0","method":"ft_call","params":[{"actionType":0,"from":"testtest3","to":"testtest1","assetId":0,"gas":2000000,"gasPrice":1,"value":0,"data":"0x51c8c3f100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000001009"},"latest"]}' http://localhost:8545


// Result
{
    "id":1,
    "jsonrpc": "2.0",
    "result": "0x"
}
```

---

#### ft_estimateGas

Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.

##### Parameters

See [ft_call](#ft_call) parameters

##### Response

- `Quantity` - the amount of gas used.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d `{"jsonrpc":"2.0","method":"ft_estimateGas","params":[{"from":"tcontract5","to":"tcontract5","data":"","assetId":1,"actionType":517,"gas":20000000,"gasPrice":1,"value":0,"remark":"0x46620e39"}],"id":1}` http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": "20000000"
}

```

---

#### ft_getChainConfig

Returns information of blockchain config.

##### Parameters

None

##### Response

- `Object` - A blockchain config object.

See [chain config](https://github.com/fractalplatform/fractal/wiki/%E5%88%9B%E4%B8%96%E6%96%87%E4%BB%B6genesis.json%E8%AF%B4%E6%98%8E)

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getChainConfig","params":[],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "bootnodes": [
            "fnode://7f06dd5bf47453e153272ff32e45c110c050ae43481e0b684cb7aac240bfeb80c92592b3afc133fcdb37a16356cb5d0ded5ff973fbb0dc8d6dda9962da1ba861@boot.t.ft.im:30000",
            "fnode://cb49ad1ea7fd8e351603404394eb568a0c163195be4faf5730d609bf939fefb2105dec9e54828ade3623a92614c9d8a14a98030d2d2b7edd63896b36c6a42a5f@boot.t.ft.im:30001"
        ],
        "chainId": 100,
        "chainName": "fractal",
        "chainUrl": "https://fractalproject.com",
        "accountParams": {
            "level": 3,
            "alllength": 31,
            "mainminlength": 7,
            "mainmaxlength": 16,
            "subminLength": 2,
            "submaxLength": 16
        },
        "assetParams": {
            "level": 3,
            "alllength": 31,
            "mainminlength": 2,
            "mainmaxlength": 16,
            "subminLength": 1,
            "submaxLength": 8
        },
        "chargeParams": {
            "assetRatio": 80,
            "contractRatio": 80
        },
        "upgradeParams": {
            "blockCnt": 10000,
            "upgradeRatio": 80
        },
        "dposParams": {
            "maxURLLen": 512,
            "unitStake": 1,
            "candidateAvailableMinQuantity": 3000000,
            "candidateMinQuantity": 500000,
            "voterMinQuantity": 1,
            "activatedMinCandidate": 28,
            "activatedMinQuantity": 1350000000,
            "blockInterval": 3000,
            "blockFrequency": 6,
            "candidateScheduleSize": 21,
            "backupScheduleSize": 7,
            "epochInterval": 604800000,
            "freezeEpochSize": 2,
            "extraBlockReward": 0,
            "blockReward": 0
        },
        "systemName": "fractal.founder",
        "accountName": "fractal.account",
        "assetName": "fractal.asset",
        "dposName": "fractal.dpos",
        "snapshotInterval": 3600000,
        "feeName": "fractal.fee",
        "systemToken": "ftoken",
        "sysTokenID": 0,
        "sysTokenDecimal": 18,
        "referenceTime": 1559620800000000000
    }
}

```

---

#### ft_newFilter

Creates a filter - `Object`, based on filter options, to notify when the state changes (logs). To check if the state has changed, call [ft_getFilterChanges](#ft_getfilterchanges)

A note on specifying topic filters:

Topics are order-dependent. A transaction with a log with topics [A, B] will be matched by the following topic filters:

- `[]` "anything"
- `[A]` "A in first position (and anything after)"
- `[`null`, B]` "anything in first position AND B in second position (and anything after)"
- `[A, B]` "A in first position AND B in second position (and anything after)"
- `[[A, B], [A, B]]` "(A OR B) in first position AND (A OR B) in second position (and anything after)"

##### Parameters

- `accounts`: `Data|Array`, account name - (optional) Contract name or a list of accounts from which logs should originate.
- `topics`: `Array of Data`, - (optional) Array of 32 Bytes `DATA` topics. Topics are order-dependent. Each topic can also be an array of DATA with "or" options.

##### Response

- `Quantity` - A filter id

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_newFilter","params":[{"accounts":["fractal.founder"],"topics":["0x000000000000000000000000a94f5374fce5edbc8e2a8697c15331677e6ebf0b"]}],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": "0xfce5a886e3550604059a5f3e404f8767"
}

```

---

#### ft_getFilterChanges

Polling method for a filter, which returns an array of logs or hashes which occurred since last poll

A note on specifying topic filters:

Topics are order-dependent. A transaction with a log with topics [A, B] will be matched by the following topic filters:

- `[]` "anything"
- `[A]` "A in first position (and anything after)"
- `[`null`, B]` "anything in first position AND B in second position (and anything after)"
- `[A, B]` "A in first position AND B in second position (and anything after)"
- `[[A, B], [A, B]]` "(A OR B) in first position AND (A OR B) in second position (and anything after)"

##### Parameters

- `Quantity` - A filter id

##### Response

- For filters created with `ft_newBlockFilter` the return are block hashes (`DATA`, 32 Bytes), e.g. `["0x3454645634534..."]`.
- For filters created with `ft_newPendingTransactionFilter`the return are transaction hashes (`DATA`, 32 Bytes), e.g. `["0x6345343454645..."]`.
- For filters created with `ft_newFilter` logs are - `Object`s with following params:
  - `index`: `Quantity` - integer of the log index position in the block.
  - `txIndex`: `Quantity` - integer of the transactions index position log was created from.
  - `actionIndex:` `Quantity` - integer of the action index position log was created from.
  - `txHash`: `DATA`, 32 Bytes - hash of the transactions this log was created from.
  - `blockHash`: `DATA`, 32 Bytes - hash of the block where this log was in. `null` when its pending.
  - `blockNumber`: `Quantity` - the block number where this log was in. `null` when its pending.
  - `name`: `DATA`, account - account from which this log originated.
  - `data`: `DATA` - contains the non-indexed arguments of the log.
  - `topics`: `Array of DATA` - Array of 0 to 4 32 Bytes `DATA` of indexed log arguments. (In _solidity_: The first topic is the _hash_ of the signature of the event (e.g. `Deposit(address,bytes32,uint256)`), except you declared the event with the `anonymous` specifier.)

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getFilterChanges","params":["0xfce5a886e3550604059a5f3e404f8767"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc":"2.0",
    "id":1,
    "result":[
        {
            "name":"contracta4",
            "topics":[
                "0x6f72646572416d6f756e74202d2d000000000000000000000000000000000000"
            ],
            "data":"0x0000000000000000000000000000000000000000000000000000000000000001",
            "blockNumber":122,
            "blockHash":"0xc1812526374103893728e8810809260d83ee0fbdaf4cfdf474929ecacb1773bd",
            "transactionHash":"0xefb0e9553cc9d6ef8ba28464cf30b31116ecd0dab876158ca1d61782ea04efdc",
            "logIndex":0,
            "actionIndex":0,
            "transactionIndex":0
        }
    ]
}

```

---

#### ft_newPendingTransactionFilter

Creates a filter in the node, to notify when new pending transactions arrive. To check if the state has changed, call [ft_getFilterChanges](#ft_getfilterchanges)

##### Parameters

None

##### Response

- `Quantity` - A filter id

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_newPendingTransactionFilter","params":[],"id":1}' http://localhost:8545

// Result

{
"jsonrpc": "2.0",
"id": 1,
"result": "0x50dbd6380389d414e254ce64f736b97c"
}

```

---

#### ft_newBlockFilter

Creates a filter in the node, to notify when new block arrive. To check if the state has changed, call [ft_getFilterChanges](#ft_getfilterchanges)

##### Parameters

None

##### Response

- `Quantity` - A filter id

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_newBlockFilter","params":[],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": "0x5aeb8e569931053fb6a9b9422c701a93"
}

```

#### ft_uninstallFilter

Uninstalls a filter with given id. Should always be called when watch is no longer needed. Additonally Filters timeout when they aren't requested with [ft_getFilterChanges](#ft_getfilterchanges) for a period of time.

##### Parameters

- `Quantity` - The filter id

##### Response

- `Boolean` - `true` if the filter was successfully uninstalled, otherwise `false`

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_uninstallFilter","params":["0xfce5a886e3550604059a5f3e404f8767"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": true
}

```

##

#### consensus_getAllCandidates

Returns all candidates.

##### Parameters

None

##### Response

- `Array` - list of candidiates.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"consensus_getAllCandidates","params":[], "id":1}' http://localhost:8545

// Result

["fractalfounder"]

```

---

#### consensus_getCandidateInfo

Returns candidates information by a specific epoch

##### Parameters

- `account` - account name of candidate.

##### Response

- `Object` - candidate's info.
  - OwnerAccount - `String` candidate's account name
  - SignAccount - `String` signer's account name
  - RegisterNumber - `Number` registration block number
  - Weight - `Number` candidate's weight
  - Balance - `Number` locked balance of candidate
  - Epoch - `Number` candidate's current epoch

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"consensus_getCandidateInfo","params":["fractalfounder"],"id":1}' http://localhost:8545

// Result


{
   "result" : {
      "Weight" : 100,
      "OwnerAccount" : "fractalfounder",
      "Epoch" : 232,
      "SignAccount" : "",
      "Balance" : 1,
      "RegisterNumber" : 0
   },
   "id" : 1,
   "jsonrpc" : "2.0"
}

```

---

#### fee_getObjectFeeByName

Returns fee detail by account's name

##### Parameters

- `String` - name of the account (asset name, contract name or miner's name)
- `Quantity` - type of the (Asset Type(0),Contract Type(1),Coinbase Type(2))

##### Response

- `Object` - a fee detail or `null` if not found.
  - `objectFeeID` - `Quantity` fee id.
  - `objectType` - `Quantity` type of fee.
  - `objectName` - `String` name of fee account.
  - `assetFees` - `Object`
    - `assetID` - `Quantity` fee asset Id.
    - `totalFee` - `Quantity` total fees received so far.
    - `remainFee` - `Quantity` no handling fees have been drawn so far.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"fee_getObjectFeeByName","params":["account4test",2],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "objectFeeID": 427,
        "objectType": 2,
        "objectName": "account4test",
        "assetFees": [
            {
                "assetID": 0,
                "totalFee": 4574852809410122731,
                "remainFee": 4574852809410122731
            }
        ]
    }

```

---

#### fee_getObjectFeeResult

Returns fee result from a specific objectFeeID within count number at a given timestamp.

##### Parameters None

- `Quantity` - the start object fee ID (start from 1).
- `Quantity` - the count of results(max 1000).
- `Quantity` - snapshot timestamp.

##### Response

- `Array` - list of fee detail.
  see [fee_getObjectFeeByName](#fee_getObjectFeeByName) response

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"fee_getObjectFeeResult","params":[1,2,1565064000000000000],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "continue": true,
        "objectFees": [
            {
                "objectFeeID": 1,
                "objectType": 1,
                "objectName": "fractal.account",
                "assetFees": [
                    {
                        "assetID": 0,
                        "totalFee": 8316000569129297073,
                        "remainFee": 8316000569129297073
                    }
                ]
            },
            {
                "objectFeeID": 2,
                "objectType": 2,
                "objectName": "fractal.founder",
                "assetFees": [
                    {
                        "assetID": 0,
                        "totalFee": 1034054730430338100850,
                        "remainFee": 1034054730430338100850
                    }
                ]
            }
        ]
    }
}

```
