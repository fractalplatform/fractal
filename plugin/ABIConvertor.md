# 如何在合约中调用Plugin

以共识获取生产者信息为例，说明如何使用。

#### 示例
1. 在plugin中定义API
```go
type SolMinerInfo struct {
	OwnerAccount   common.Address // 生产者账户
	SignAccount    common.Address // 签名账户
	RegisterNumber uint64   // 注册高度
	Weight         uint64   // 权重
	Balance        *big.Int // 抵押余额
	Epoch          uint64   // 出块轮数相关
}
func (c *Consensus) Sol_GetMinerInfo(context *ContextSol, miner common.Address) (*SolMinerInfo, error) {
    info := &SolMinerInfo{}
    ...
    return info,nil
}
```
2. 在solidty中定义类似的方法
```js
contract ConsensusAPI {
    struct MinerInfo {
        address OwnerAccount;
        address SignAccount;
        uint256 RegisterNumber;
        uint256 Weight;
        uint256 Balance;
        uint256 Epoch;
    }
    function GetMinerInfo(address miner) external returns(MinerInfo memory);
    //function UnregisterMiner() public;
    //function RegisterMiner(address miner) external payable;
}
```
3. 在solidity中调用
```js
contract TestRead {
    ConsensusAPI constant consensus = ConsensusAPI(address(bytes20("fractaldpos")));
    event InfoLog(address,uint256,uint256);
    function testRead(address miner) public {
        ConsensusAPI.MinerInfo memory info = consensus.GetMinerInfo(miner);
        emit InfoLog(info.OwnerAccount, info.Weight, info.Balance);
    }
}
```
#### 插件API约定:
1. 插件定义的API方法名需要以`Sol_`开头
2. API第一个参数为`*ContextSol`类型，包含了交易信息以及插件访问接口
3. API最后一个返回值为`error`类型，如果不是`nil`，则表示执行失败。同时API内修改的`state`会被回退
4. 入参支持大部分基础类型(非`map`,`struct`,`float`)和`big.Int`,`common.Address`
5. 返回参数不支持`map`,`float`，并且可返回多个参数
6. 账户名在合约内为`address`类型, 访问是需要做转换
    - 使用`common.Address.AccountName()`来获取地址对应的账户名
    - 使用`common.StringToAddress(String)`来将账户名转换为地址

#### 合约API约定:
1. 需要定义和插件相同参数以及返回值的方法，且类型必须为`external`

#### 合约<=>插件 类型映射:
- 整型: `uintX,intX` <=> `uintX,intX` X = [8,16,32,64]
- 字符串: `string` <=> `string`
- byte流: `bytes` <=> `[]byte`
- 定长byte: `bytesX` <=> `[X]byte` 1 <= X <= 32
- 地址: `address` <=> `common.Address`
- 切片: `type[]` <=> `[]type`
- 数组: `type[X]` <=> `[X]type`