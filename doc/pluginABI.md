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

[consensus.sol]: ../plugin/consensus.sol
