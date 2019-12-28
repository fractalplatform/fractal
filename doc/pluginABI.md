## consensus
```js
pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

interface ConsensusAPI {
    struct MinerInfo {
        address OwnerAccount; // 生产者账户
        address SignAccount;  // 签名账户
        uint256 RegisterNumber; // 最后注册时间
        uint256 Weight; // 权重
        uint256 Balance; // 抵押数量
        uint256 Epoch; // 轮数相关
    }
    function GetMinerInfo(address miner) external returns(MinerInfo memory);
    function UnregisterMiner() external;
    function RegisterMiner(address signer) external payable;
}
```
#### GetMinerInfo
说明: 获取生产者详情

参数: 
- `address miner`: 生产者账户名

返回:
- `struct MinerInfo`: 生产者信息

#### RegisterMiner
说明: 注册生产者

参数:
- `address signer`: 签名者账户, 留空则使用默认账户签名

返回: 无

#### UnregisterMiner
说明: 注销生产者，并将抵押金额退回至生产者账户

参数: 无

> 示例: [consensus.sol]

## account
```js
pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

interface AccountAPI {
    function GetBalance(address account, uint64 assetID) external returns(uint256);
    function Transfer(address to, uint64 assetid, uint256 value) external;
}
```
#### GetBalance
说明: 获取账户余额

参数: 
- `address account`: 账户名

返回:
- `uint256`: 账户余额

#### Transfer
说明: 转账

参数: 
- `address to`: 账户名
- `uint64 assetid`: 资产id
- `uint256 value`: 转账金额

返回:
- 无

## item
#### IssueWorld
说明: 发行world

参数: 
- `address owner`: 账户名
- `string name`: world名字
- `string description`: 描述

返回:
- 无

#### IssueItemType
说明: 发行道具类型

参数: 
- `uint64 worldID`: worldID
- `string name`: itemType名字
- `bool merge`: 同质标识
- `uint64 upperLimit`: 上限
- `string description`: 描述
- `uint64[] attrPermission`: 属性修改权限
- `string[] attrName`: 属性名字
- `string[] attrDes`: 属性描述

返回:
- 无

#### IncreaseItem
说明: 增发item

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `string name`: itemType名字
- `address owner`: 账户名
- `string description`: 描述
- `uint64[] attrPermission`: 属性修改权限
- `string[] attrName`: 属性名字
- `string[] attrDes`: 属性描述

返回:
- 无

#### DestroyItem
说明: 销毁item

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `uint64 itemID`: itemID

返回:
- 无

#### IncreaseItems
说明: 增发items

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `address to`: 账户名
- `uint64 amount`: 数量

返回:
- 无

#### DestroyItems
说明: 销毁items

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `uint64 amount`: 数量

返回:
- 无

#### TransferItem
说明: 交易

参数: 
- `address to`: 账户名
- `uint64[] worldID`: worldID
- `uint64[] itemTypeID`: itemTypeID
- `uint64[] itemiD`: itemID，如果交易items，填0。交易item，填1
- `uint64[] amount`: 数量，如果交易item，填1。

返回:
- 无

#### AddItemTypeAttributes
说明: 添加itemType属性

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `uint64[] attrPermission`: 属性修改权限
- `string[] attrName`: 属性名字
- `string[] attrDes`: 属性描述

返回:
- 无

#### DelItemTypeAttributes
说明: 删除itemType属性

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `string[] attrName`: 属性名字

返回:
- 无

#### ModifyItemTypeAttributes
说明: 修改itemType属性

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `uint64[] attrPermission`: 属性修改权限
- `string[] attrName`: 属性名字
- `string[] attrDes`: 属性描述

返回:
- 无

#### AddItemAttributes
说明: 添加item属性

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `uint64 itemID`: itemID
- `uint64[] attrPermission`: 属性修改权限
- `string[] attrName`: 属性名字
- `string[] attrDes`: 属性描述

返回:
- 无

#### DelItemAttributes
说明: 删除item属性

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `uint64 itemID`: itemID
- `string[] attrName`: 属性名字

返回:
- 无

#### ModifyItemAttributes
说明: 修改itemType属性

参数: 
- `uint64 worldID`: worldID
- `uint64 itemTypeID`: itemTypeID
- `uint64 itemID`: itemID
- `uint64[] attrPermission`: 属性修改权限
- `string[] attrName`: 属性名字
- `string[] attrDes`: 属性描述

返回:
- 无

> 示例: [item.sol]

[consensus.sol]: ../plugin/consensus.sol
[item.sol]: ../plugin/item.sol
