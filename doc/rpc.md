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

  - **item**

    - [item_getWorldByID](#item_getWorldByID)
    - [item_getWorldByName](#item_getWorldByName)
    - [item_getItemTypeByID](#item_getItemTypeByID)
    - [item_getItemTypeByName](#item_getItemTypeByName)
    - [item_getItemByID](#item_getItemByID)
    - [item_getItemByOwner](#item_getItemByOwner)
    - [item_getAccountItems](#item_getAccountItems)
    - [item_getItemTypeAttributeByID](#item_getItemTypeAttributeByID)
    - [item_getItemTypeAttributeByName](#item_getItemTypeAttributeByName)
    - [item_getItemAttributeByID](#item_getItemAttributeByID)
    - [item_getItemAttributeByName](#item_getItemAttributeByName)

  - **ft**

    - [ft_sendRawTransaction](#ft_sendRawTransaction)
    - [ft_getTransactionByHash](#ft_getTransactionByHash)
    - [ft_getInternalTxByHash](#ft_getInternalTxByHash)
    - [ft_getTransactionReceipt](#ft_getTransactionReceipt)
    - [ft_getInternalTxByAccount](#ft_getInternalTxByAccount)
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

  - **consensus**

    - [consensus_getAllCandidates](#consensus_getAllCandidates)
    - [consensus_getCandidateInfo](#consensus_getCandidateInfo)

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

#### item_getWorldByID

Get world info by world id

##### Parameters

- `Quantity` - id of the world.

##### Response

- `Object` - A world object,`world not exist` if not found
  - `worldID` - `Quantity` id of the world.
  - `name` - `String` name of the world.
  - `owner` - `String` who owns all permissions to the world.
  - `creator` - `String` who creator the world.
  - `description` - `String` description of the world.
  - `total` - `Quantity` issue itemType total in the world

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getWorldByID","params":[1],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "worldID": 1,
        "name": "menghuanxiyo",
        "owner": "testtest1",
        "creator": "testtest31",
        "description": "梦幻西游",
        "total": 2
    }
}

```
---

#### item_getWorldByName

Get world info by world name

##### Parameters

- `String` - name of the world.

##### Response

- `Object` - A world object,`world not exist` if not found
  - `worldID` - `Quantity` id of the world.
  - `name` - `String` name of the world.
  - `owner` - `String` who owns all permissions to the world.
  - `creator` - `String` who creator the world.
  - `description` - `String` description of the world.
  - `total` - `Quantity` issue itemType total in the world

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getWorldByName","params":["menghuanxiyo"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "worldID": 1,
        "name": "menghuanxiyo",
        "owner": "testtest1",
        "creator": "testtest31",
        "description": "梦幻西游",
        "total": 2
    }
}

```
---

#### item_getItemTypeByID

Get item type info by id

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.

##### Response

- `Object` - A itemType object,`itemType not exist` if not found
  - `worldID` - `Quantity` id of the world.
  - `itemTypeID` - `Quantity` id of the itemType.
  - `name` - `String` name of the world.
  - `merge` - `Bool` merge info.
  - `upperLimit` - `Quantity` maximum number of additional issues.
  - `addIssue` - `Quantity` total amount of issuance.
  - `description` - `String` description of the itemType.
  - `total` - `Quantity` issue item in the itemType.
  - `attrTotal` - `Quantity` item attribute total.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemTypeByID","params":[1,2],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "worldID": 1,
        "itemTypeID": 2,
        "name": "molishouzhuo",
        "merge": true,
        "upperLimit": 0,
        "addIssue": 100,
        "description": "魔力手镯",
        "total": 95,
        "attrTotal": 2
    }
}

```
---

#### item_getItemTypeByName

Get item type info by name

##### Parameters

- `Quantity` - id of the world.
- `String` - name of the item.

##### Response

- `Object` - A itemType object,`itemType not exist` if not found
  - `worldID` - `Quantity` id of the world.
  - `itemTypeID` - `Quantity` id of the itemType.
  - `name` - `String` name of the world.
  - `merge` - `Bool` merge info.
  - `upperLimit` - `Quantity` maximum number of additional issues.
  - `addIssue` - `Quantity` total amount of issuance.
  - `description` - `String` description of the itemType.
  - `total` - `Quantity` issue item in the itemType.
  - `attrTotal` - `Quantity` item attribute total.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemTypeByName","params":[1,"molishouzhuo"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "worldID": 1,
        "itemTypeID": 2,
        "name": "molishouzhuo",
        "merge": true,
        "upperLimit": 0,
        "addIssue": 100,
        "description": "魔力手镯",
        "total": 95,
        "attrTotal": 2
    }
}

```
---

#### item_getItemByID

Get item info by id

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.
- `Quantity` - id of the item.

##### Response

- `Object` - A item object,`item not exist` if not found
  - `worldID` - `Quantity` id of the world.
  - `itemTypeID` - `Quantity` id of the itemType.
  - `itemID` - `Quantity` id of the item.
  - `owner` - `String` owner of the item.
  - `description` - `String` description of the item.
  - `attrTotal` - `Quantity` item attribute total.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemByID","params":[1,1,1],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "worldID": 1,
        "itemTypeID": 1,
        "itemID": 1,
        "owner": "testtest3",
        "description": "💫",
        "destroy": false,
        "attrTotal": 4
    }
}

```
---

#### item_getItemByOwner

Get item info by owner

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.
- `String` - owner of the item.

##### Response

- `Object` - A item object,`item not exist` if not found
  - `worldID` - `Quantity` id of the world.
  - `itemTypeID` - `Quantity` id of the itemType.
  - `itemID` - `Quantity` id of the item.
  - `owner` - `String` owner of the item.
  - `description` - `String` description of the item.
  - `attrTotal` - `Quantity` item attribute total.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemByOwner","params":[1,1,"testtest1"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": [
        {
            "worldID": 1,
            "itemTypeID": 1,
            "itemID": 2,
            "owner": "testtest1",
            "description": "💫☺",
            "destroy": true,
            "attrTotal": 3
        }
    ]
}

```
---

#### item_getAccountItems

Get items info by owner

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.
- `String` - owner of the items.

##### Response

- `Object` - A items object,`item not exist` if not found
  - `worldID` - `Quantity` id of the world.
  - `itemTypeID` - `Quantity` id of the itemType.
  - `owner` - `String` owner of the items.
  - `total` - `Quantity` items total.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getAccountItems","params":[1,1,"testtest1"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "worldID": 1,
        "itemTypeID": 2,
        "owner": "testtest1",
        "total": 70
    }
}

```
---

#### item_getItemTypeAttributeByID

Get items info by owner

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.
- `Quantity` - id of the attribute.

##### Response

- `Object` - A attribute object,`attribute not exist` if not found
  - `modifyPermission` - `Quantity` modify permission.
  - `name` - `String` name of the attributes.
  - `description` - `String` description of the attribute.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemTypeAttributeByID","params":[1,1,2],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "modifyPermission": 0,
        "name": "gongjishanghai",
        "description": "攻击伤害+50"
    }
}

```
---

#### item_getItemTypeAttributeByName

Get attribute info by name

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.
- `String` - name of the attribute.

##### Response

- `Object` - A attribute object,`attribute not exist` if not found
  - `modifyPermission` - `Quantity` modify permission.
  - `name` - `String` name of the attributes.
  - `description` - `String` description of the attribute.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemTypeAttributeByName","params":[1,1,"gongjishanghai"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "modifyPermission": 0,
        "name": "gongjishanghai",
        "description": "攻击伤害+50"
    }
}

```
---

#### item_getItemAttributeByID

Get attribute info by name

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.
- `Quantity` - id of the item.
- `Quantity` - id of the attribute.

##### Response

- `Object` - A attribute object,`attribute not exist` if not found
  - `modifyPermission` - `Quantity` modify permission.
  - `name` - `String` name of the attributes.
  - `description` - `String` description of the attribute.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemAttributeByID","params":[1,1,1,1],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "modifyPermission": 0,
        "name": "gongjishanghai",
        "description": "攻击伤害+50"
    }
}

```
---

#### item_getItemAttributeByName

Get attribute info by name

##### Parameters

- `Quantity` - id of the world.
- `Quantity` - id of the itemType.
- `Quantity` - id of the item.
- `String` - name of the attribute.

##### Response

- `Object` - A attribute object,`attribute not exist` if not found
  - `modifyPermission` - `Quantity` modify permission.
  - `name` - `String` name of the attributes.
  - `description` - `String` description of the attribute.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"item_getItemAttributeByName","params":[1,1,1,"item"],"id":1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "modifyPermission": 0,
        "name": "gongjishanghai",
        "description": "攻击伤害+50"
    }
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
  - `type` - `Quantity` type of transaction.
  - `payloadType` - `Quantity` type of payload (only plugin transaction).
  - `nonce` - `Quantity` the number of actions made by the sender prior to this one.
  - `from` - `String` account name of the sender.
  - `to` - `String` account name of the sender.
  - `assetID` - `Quantity` which asset to transfer.
  - `gas` - `Quantity` gas provided by the sender.
  - `value` - `Quantity` value transferred.
  - `remark` - `String` the data of extra context.
  - `payload` - `String` the data send along with the action.
  - `gasAssetID` - `Quantity` which asset to pay gas.
  - `gasPrice` - `Quantity` gas price provided by the sender.
  - `gasCost` - `Quantity` the number of transaction gas cost.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"ft_getTransactionByHash","params":["0xeb701c7159e270f3ae9f209951ded94896ced101a0bb9c6b40729e3d57d1067b"],"id": 1}' http://localhost:8545

// Result contract transaction
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "blockNumber": 12,
        "transactionIndex": 0,
        "blockHash": "0x9d0b89d443c14dc6414d22c5983d7575d58cc546033a9f18d6488a5802597e04",
        "txHash": "0x3d12afc586f461c9b473fa984011db619e5c208abd106961ab396f7c96edf301",
        "type": 1,
        "nonce": 1,
        "from": "fractalfounder",
        "to": "fractalfounder",
        "assetID": 0,
        "gas": 20000000,
        "value": 0,
        "remark": "0x",
        "payload": "0x27e1b24f0000000000000000000000000000000000000000000000000000000000000000",
        "gasAssetID": 0,
        "gasPrice": 100000000000,
        "gasCost": 2000000000000000000
    }
}


// Result contract transaction
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "blockNumber": 12,
        "transactionIndex": 0,
        "blockHash": "0x9d0b89d443c14dc6414d22c5983d7575d58cc546033a9f18d6488a5802597e04",
        "txHash": "0x3d12afc586f461c9b473fa984011db619e5c208abd106961ab396f7c96edf301",
        "type": 1,
        "payloadType":255,
        "nonce": 1,
        "from": "fractalfounder",
        "to": "fractalfounder",
        "assetID": 0,
        "gas": 20000000,
        "value": 0,
        "remark": "0x",
        "payload": "0x27e1b24f0000000000000000000000000000000000000000000000000000000000000000",
        "gasAssetID": 0,
        "gasPrice": 100000000000,
        "gasCost": 2000000000000000000
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
  - `type` - `Quantity` type of transaction.
  - `status` - `Quantity` either 1 (success) or 0 (failure).
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
    - `transactionIndex` - `Quantity` integer of the transaction's index position in the block.
    - `logIndex` - `Quantity` integer of the log's index position in the action.

##### Example

```js

// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0", "method":"ft_getTransactionReceipt","params":["0x01c746c5a181088c14e7e6f43739d0dc3db423f6b2cfa102d4549f338514a804"],"id": 1}' http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "blockNumber": 16,
        "blockHash": "0x217d8d0db6cdf962a0a87bbabd5f49a83924f46bf9ddabb2641f8ecb92adcb48",
        "txHash": "0x97c30f65a686e12e3303fb0b3293eb3084a04278438d50983e2b7322b75cb948",
        "transactionIndex": 0,
        "postState": "0x4ef1d85be4afd8f44934a34473aa7e5e2772f90f668e485e51c0980b21076d55",
        "Type": 2,
        "status": 1,
        "gasUsed": 2170508,
        "gasAllot": [
            {
                "name": "fractalfounder",
                "gas": 899660,
                "typeID": 2
            }
        ],
        "error": "",
        "cumulativeGasUsed": 2170508,
        "totalGasUsed": 2170508,
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "logs": null
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

curl -X POST -H "Content-Type: application/json" -d `{"jsonrpc":"2.0", "method":"ft_getTransactions","params":[["0x3d12afc586f461c9b473fa984011db619e5c208abd106961ab396f7c96edf301"]],"id": 1}` http://localhost:8545

// Result
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": [
        {
            "blockNumber": 17,
            "transactionIndex": 0,
            "blockHash": "0xf4c0e11c9a629016e420d921a1317aa4d5c3c6861be79fd6a78c936f90523c51",
            "txHash": "0x3d12afc586f461c9b473fa984011db619e5c208abd106961ab396f7c96edf301",
            "type": 1,
            "nonce": 1,
            "from": "fractalfounder",
            "to": "fractalfounder",
            "assetID": 0,
            "gas": 20000000,
            "value": 0,
            "remark": "0x",
            "payload": "0x27e1b24f0000000000000000000000000000000000000000000000000000000000000000",
            "gasAssetID": 0,
            "gasPrice": 100000000000,
            "gasCost": 2000000000000000000
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
  - `gasLimit`- `Quantity` the maximum gas allowed in this block.
  - `gasUsed` - `Quantity` the total used gas by all transactions in this block.
  - `hash` - `String` hash of block.
  - `logsBloom` - `String` the bloom filter for the logs of the block.
  - `miner` - `String` the name of the beneficiary to whom the mining rewards were given.
  - `number` - `Quantity` the block number.
  - `proof` - `String`signature of consensus
  - `parentHash` - `String` hash of the parent block.
  - `receiptsRoot` - `String` the root of the receipts trie of the block.
  - `sign` signature of the block by miner
  - `size` - `String` integer the size of this block in bytes.
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
    "jsonrpc":"2.0",
    "id":1,
    "result":[
        {
            "difficulty":864001,
            "extraData":"0x",
            "gasLimit":30000000,
            "gasUsed":0,
            "hash":"0xbdb89ef1026b2d2ff67e4c85250f954b9760feef35f1c0f809729158e9bac79d",
            "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "miner":"fractalfounder",
            "number":19,
            "parentHash":"0xc5752461b4700e9bc38bab5c6df1cd3ab20f73ce11c35854292e8be8f1845e2a",
            "proof":"0x85a91e6e851c210d263cc2d06e02f9ab630960d2dc524bf1f906478749009341",
            "receiptsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000",
            "sign":"0xfccfb1b1cea8ced1b760111b1448299b1c14b5916a9fdb6539d335b2767942740204db9f275fd7ad40b9d2f79cb0f26d28a7edca29f419d4ee5f493447cb860b25",
            "size":530,
            "stateRoot":"0x9441c7f35ca79fcff592169047a196d36f32cbc93f851cd6c2957d30e2cda022",
            "timestamp":1576811018,
            "totalDifficulty":16265788,
            "transactions":[

            ],
            "transactionsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000"
        }
    ]
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

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getBlockByHash","params":["0xbdb89ef1026b2d2ff67e4c85250f954b9760feef35f1c0f809729158e9bac79d", true],"id":1}' http://localhost:8545

// Result
{
    "jsonrpc":"2.0",
    "id":1,
    "result":[
        {
            "difficulty":864001,
            "extraData":"0x",
            "gasLimit":30000000,
            "gasUsed":0,
            "hash":"0xbdb89ef1026b2d2ff67e4c85250f954b9760feef35f1c0f809729158e9bac79d",
            "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "miner":"fractalfounder",
            "number":19,
            "parentHash":"0xc5752461b4700e9bc38bab5c6df1cd3ab20f73ce11c35854292e8be8f1845e2a",
            "proof":"0x85a91e6e851c210d263cc2d06e02f9ab630960d2dc524bf1f906478749009341",
            "receiptsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000",
            "sign":"0xfccfb1b1cea8ced1b760111b1448299b1c14b5916a9fdb6539d335b2767942740204db9f275fd7ad40b9d2f79cb0f26d28a7edca29f419d4ee5f493447cb860b25",
            "size":530,
            "stateRoot":"0x9441c7f35ca79fcff592169047a196d36f32cbc93f851cd6c2957d30e2cda022",
            "timestamp":1576811018,
            "totalDifficulty":16265788,
            "transactions":[

            ],
            "transactionsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000"
        }
    ]
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

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_getBlockByNumber","params":[19, false],"id":1}' http://localhost:8545


// Result

{
    "jsonrpc":"2.0",
    "id":1,
    "result":[
        {
            "difficulty":864001,
            "extraData":"0x",
            "gasLimit":30000000,
            "gasUsed":0,
            "hash":"0xbdb89ef1026b2d2ff67e4c85250f954b9760feef35f1c0f809729158e9bac79d",
            "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "miner":"fractalfounder",
            "number":19,
            "parentHash":"0xc5752461b4700e9bc38bab5c6df1cd3ab20f73ce11c35854292e8be8f1845e2a",
            "proof":"0x85a91e6e851c210d263cc2d06e02f9ab630960d2dc524bf1f906478749009341",
            "receiptsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000",
            "sign":"0xfccfb1b1cea8ced1b760111b1448299b1c14b5916a9fdb6539d335b2767942740204db9f275fd7ad40b9d2f79cb0f26d28a7edca29f419d4ee5f493447cb860b25",
            "size":530,
            "stateRoot":"0x9441c7f35ca79fcff592169047a196d36f32cbc93f851cd6c2957d30e2cda022",
            "timestamp":1576811018,
            "totalDifficulty":16265788,
            "transactions":[

            ],
            "transactionsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000"
        }
    ]
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
            "difficulty": 864001,
            "extraData": "0x",
            "gasLimit": 30000000,
            "gasUsed": 410725,
            "hash": "0xf4c0e11c9a629016e420d921a1317aa4d5c3c6861be79fd6a78c936f90523c51",
            "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "miner": "fractalfounder",
            "number": 17,
            "parentHash": "0x217d8d0db6cdf962a0a87bbabd5f49a83924f46bf9ddabb2641f8ecb92adcb48",
            "proof": "0xbc205af6295c0b7721732691a078d9448fb3843f7cbe34e04beff7ceab2b3495",
            "receiptsRoot": "0x9c83514e9dcef16cbf98ecd5276bbdfaf697163437011d7ed46920cbd5bca9ff",
            "sign": "0x797cb826deb3090b5b01419f34e7a086c7166d41014a515f705099d8cf2f41fb47de37ab8e299a25cc174120c22150ad896e3f9c5ac061553c68e9980c3a3f3925",
            "size": 691,
            "stateRoot": "0x53aab00eb2ae705011a377640fe7ed1da461806665411e2367e04f7c66ad1930",
            "timestamp": 1576811012,
            "totalDifficulty": 14537786,
            "transactions": [
                {
                    "blockNumber": 17,
                    "transactionIndex": 0,
                    "blockHash": "0xf4c0e11c9a629016e420d921a1317aa4d5c3c6861be79fd6a78c936f90523c51",
                    "txHash": "0x3d12afc586f461c9b473fa984011db619e5c208abd106961ab396f7c96edf301",
                    "type": 1,
                    "nonce": 1,
                    "from": "fractalfounder",
                    "to": "fractalfounder",
                    "assetID": 0,
                    "gas": 20000000,
                    "value": 0,
                    "remark": "0x",
                    "payload": "0x27e1b24f0000000000000000000000000000000000000000000000000000000000000000",
                    "gasAssetID": 0,
                    "gasPrice": 100000000000,
                    "gasCost": 2000000000000000000
                }
            ],
            "transactionsRoot": "0x64eec3cb845d539b0cd130f30ebff74d8695b96352e4c6de0ce30b27f75adc13"
        },
        "receipts": [
            {
                "PostState": "U6qwDrKucFARo3dkD+ftHaRhgGZlQR4jZ+BPfGatGTA=",
                "Status": 0,
                "Index": 0,
                "GasUsed": 410725,
                "GasAllot": [
                    {
                        "name": "fractalfounder",
                        "gas": 208500,
                        "typeID": 2
                    }
                ],
                "Error": "evm: execution reverted",
                "CumulativeGasUsed": 410725,
                "Bloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
                "Logs": [],
                "TxHash": "0x3d12afc586f461c9b473fa984011db619e5c208abd106961ab396f7c96edf301",
                "TotalGasUsed": 410725
            }
        ],
        "detailTxs": null
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
    "jsonrpc":"2.0",
    "id":1,
    "result":[
        {
            "difficulty":864001,
            "extraData":"0x",
            "gasLimit":30000000,
            "gasUsed":0,
            "hash":"0xbdb89ef1026b2d2ff67e4c85250f954b9760feef35f1c0f809729158e9bac79d",
            "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "miner":"fractalfounder",
            "number":19,
            "parentHash":"0xc5752461b4700e9bc38bab5c6df1cd3ab20f73ce11c35854292e8be8f1845e2a",
            "proof":"0x85a91e6e851c210d263cc2d06e02f9ab630960d2dc524bf1f906478749009341",
            "receiptsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000",
            "sign":"0xfccfb1b1cea8ced1b760111b1448299b1c14b5916a9fdb6539d335b2767942740204db9f275fd7ad40b9d2f79cb0f26d28a7edca29f419d4ee5f493447cb860b25",
            "size":530,
            "stateRoot":"0x9441c7f35ca79fcff592169047a196d36f32cbc93f851cd6c2957d30e2cda022",
            "timestamp":1576811018,
            "totalDifficulty":16265788,
            "transactions":[

            ],
            "transactionsRoot":"0x0000000000000000000000000000000000000000000000000000000000000000"
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

  - `txType` - `Quantity` type of transaction. `0` creat contract,`1` call contract
  - `from` - `String` name of sender account.
  - `to` - `String` name of receipt account.
  - `assetId` - `Quantity` id of used asset.
  - `gas` - `Quantity` integer of the gas provided for the transaction execution. eth_call consumes zero gas, but this parameter may be needed by some executions.
  - `gasPrice` - `Quantity` integer of the gasPrice used for each paid gas.
  - `value` - `Quantity` integer of the value sent with this transaction.
  - `data` - `String` hash of the method signature and encoded parameters.
  - `remark` - `String` extra data with this transaction.

- `Quantity` - height of the block, or string value `latest` or `earliest`.

##### Response

`String` - the return value of executed contract.

##### Example

```js
// Request

curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ft_estimateGas","params":[{"txType":1,"from":"fractalfounder","to":"fractalfounder","assetID":0,"gas":200000000,"gasPrice":100000000000,"value":0,"data":"0x27e1b24f0000000000000000000000000000000000000000000000000000000000000000","remark":""}],"id":1},"latest"]}' http://localhost:8545


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

curl -X POST -H "Content-Type: application/json" -d `{"jsonrpc":"2.0","method":"ft_estimateGas","params":[{"txType":1,"from":"fractalfounder","to":"fractalfounder","assetID":0,"gas":200000000,"gasPrice":100000000000,"value":0,"data":"0x27e1b24f0000000000000000000000000000000000000000000000000000000000000000","remark":""}],"id":1}` http://localhost:8545

// Result

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": 203732
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
        "bootnodes": [],
        "chainId": 1,
        "chainName": "fractal",
        "chainUrl": "https://fractalproject.com",
        "snapshotInterval": 180000,
        "systemName": "fractalfounder",
        "accountName": "fractalaccount",
        "assetName": "fractalasset",
        "dposName": "fractaldpos",
        "feeName": "fractalfee",
        "systemToken": "ftoken",
        "sysTokenID": 0,
        "sysTokenDecimal": 0,
        "referenceTime": 1575967052
    }
}

```

---
