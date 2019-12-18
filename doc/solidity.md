- [新增资产相关接口](#%e6%96%b0%e5%a2%9e%e8%b5%84%e4%ba%a7%e7%9b%b8%e5%85%b3%e6%8e%a5%e5%8f%a3)
  - [issueasset](#issueasset)
  - [addasset](#addasset)
  - [transfer](#transfer)
  - [setassetowner](#setassetowner)
  - [destoryasset](#destoryasset)
  - [balanceex](#balanceex)
  - [getassetid](#getassetid)
  - [assetinfo](#assetinfo)
- [新增快照相关接口](#%e6%96%b0%e5%a2%9e%e5%bf%ab%e7%85%a7%e7%9b%b8%e5%85%b3%e6%8e%a5%e5%8f%a3)
  - [snapshottime](#snapshottime)
  - [snapbalance](#snapbalance)
- [新增加解密接口](#%e6%96%b0%e5%a2%9e%e5%8a%a0%e8%a7%a3%e5%af%86%e6%8e%a5%e5%8f%a3)
  - [cryptocalc](#cryptocalc)
- [新增账户相关接口](#%e6%96%b0%e5%a2%9e%e8%b4%a6%e6%88%b7%e7%9b%b8%e5%85%b3%e6%8e%a5%e5%8f%a3)
  - [getaccountid](#getaccountid)
  - [getaccounttime](#getaccounttime)
- [新增手续费相关接口](#%e6%96%b0%e5%a2%9e%e6%89%8b%e7%bb%ad%e8%b4%b9%e7%9b%b8%e5%85%b3%e6%8e%a5%e5%8f%a3)
  - [withdrawfee](#withdrawfee)
  - [deductgas](#deductgas)
- [新增DPOS相关接口](#%e6%96%b0%e5%a2%9edpos%e7%9b%b8%e5%85%b3%e6%8e%a5%e5%8f%a3)
  - [getepoch](#getepoch)
  - [getcandidatenum](#getcandidatenum)
  - [getcandidate](#getcandidate)
  - [getvoterstake](#getvoterstake)
- [新增属性](#%e6%96%b0%e5%a2%9e%e5%b1%9e%e6%80%a7)
  - [assetid](#assetid)
  - [recipient](#recipient)
- [不可用特性](#%e4%b8%8d%e5%8f%af%e7%94%a8%e7%89%b9%e6%80%a7)
  - [不再支持合约中创建账户或者合约](#%e4%b8%8d%e5%86%8d%e6%94%af%e6%8c%81%e5%90%88%e7%ba%a6%e4%b8%ad%e5%88%9b%e5%bb%ba%e8%b4%a6%e6%88%b7%e6%88%96%e8%80%85%e5%90%88%e7%ba%a6)
  - [不再支持ether，wei等单位](#%e4%b8%8d%e5%86%8d%e6%94%af%e6%8c%81etherwei%e7%ad%89%e5%8d%95%e4%bd%8d)

## 新增资产相关接口

### issueasset

函数说明：发行资产

参数说明：

- `string desc，资产信息（由以下几项的字符串拼接而成，用逗号分隔）`
  - `name 资产名`
  - `symbol 资产标志名` 
  - `total 发行总量` 
  - `decimal 资产精度`
  - `owner 资产owner`
  - `limit 资产发行上限`
  - `founder 资产founder`
  - `contract 协议合约名(可为空）`
  - `detail 协议描述信息（可为空，长度不超过255）`

返回值：

- `uint256 assetId, 资产ID(错误返回0)`

函数原型：`uint256 assetId = issueasset(string desc)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    //string desc ="ethereum,eth,10000000000,10,owner,100,founder,contract,detail"; 
    function reg(string desc) public payable returns(uint256){
        return issueasset(desc);
    }
}
```

### addasset

函数说明：增发资产

参数说明：

- `uint256 assetId，资产ID`
- `address to，账户ID(增发的资产将增发到此账户）`
- `uint256 value，增发数量`

返回值：空

函数原型：`addasset(uint256 assetId, address to, uint256 value)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function add(uint256 assetId, address to, uint256 value) public {
        addasset(assetId, to, value);
    }
}
```

### transfer 

函数说明：转账（此方法将替换原先的transfer方法，并且合约中执行此方法时，不会再去调用接收方的回调函数，而是单纯的转账操作）

参数说明：

- `address to，账户ID（转账接收方）`
- `uint256 assetId，资产ID`
- `uint256 value，转账金额`

返回值：空

函数原型：`to.transfer(uint256 assetId, uint256 value)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function transfer(uint256 assetId, address to, uint256 value) public {
        to.transfer(assetId, value);
    }
}
```

### setassetowner

函数说明：修改资产owner

参数说明：

- `uint256 assetId，资产ID`
- `address newowner，账户ID（新的资产owner）`

返回值：空

函数原型：`setassetowner(uint256 assetId, address newowner)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function setowner(uint256 assetId, address newowner) public {
        setassetowner(assetId, newowner);
    }
}
```

### destoryasset

函数说明：销毁资产

参数说明：

- `uint256 assetId，资产ID`
- `uint256 value，销毁资产量`

返回值：

- `uint256 assetId，资产ID(错误返回0)`

函数原型：`uint256 assetId = destroyasset(uint256 assetId, uint256 value)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
   function destroyAsset(uint256 assetId, uint256 value) public returns(uint256){
        return destroyasset(assetId, value);
    }
}
```

### balanceex

函数说明：查询余额

参数说明：

- `address to，账户ID`
- `uint256 assetId，资产ID`

返回值：

- `uint256 balance，余额`

函数原型：`uint256 balance = to.balanceex(uint256 assetId)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function getBalance(uint256 assetId, address to) public returns(uint256) {
        return to.balanceex(assetId);
    }
}
```

### getassetid

函数说明：查询资产ID

参数说明：

- `string name，资产名`

返回值：

- `int64 assetID，资产ID`

函数原型：`int64 assetID = getassetid(string name)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function getAssetID(string name) public returns(int64) {
        return getassetid(name);
    }
}
```

### assetinfo

函数说明：获取快照时刻资产的总量及资产名称

参数说明：

- `uint256 assetId，资产ID`
- `uint256 time，快照时间`
- `(bytes memory) assetName，资产名（作为入参传入）`

返回值：

- `uint256 amount，资产总量`

函数原型：`uint256 amount = assetinfo(uint256 assetId, uint256 time, bytes assetName)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function getAssetInfo(uint256 assetId, uint256 t) public returns (bytes,uint256){
        bytes memory assetName = new bytes(20);
        uint256 assetAmount = assetinfo(assetId, t, assetName);
        return (assetName, assetAmount);
    }
}
```

## 新增快照相关接口

### snapshottime

函数说明：获取快照时间

参数说明：

- `uint256 typeId，当typeId为0的时候，获取最新的快照时间；当为1的时候，获取time的前一个快照时间，当为2的时候，获取time的后一个快照时间`
- `uint256 time，时间`

返回值：

- `uint256 snapshotTime，快照时间`

函数原型：`uint256 snapshotTime = snapshottime(uint256 time, uint256 typeId)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function getSnapshotTime(uint256 typeId, uint256 time) public returns (uint256){
        return snapshottime(typeId, time);
    }
}
```

### snapbalance

函数说明：获取快照按时间余额

参数说明：

- `address to，账户ID`
- `uint256 assetId，资产ID`
- `uint256 time，快照时间`
- `uint256 opt，当为0的时候，获取主资产的余额，为1的时候，获取主资产的余额加质押金的余额，为2的时候，获取主资产加子资产的余额，为3的时候，获取主资产的质押金，余额和子资产的余额`

返回值：

- `uint256 balance，余额`

函数原型：`uint256 balance = to.snapbalance(uint256 assetId, uint256 time, uint256 opt)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function getSnapBalance(address to, uint256 assetId, uint256 t, uint256 opt) public returns (uint256){
        return to.snapbalance(assetId, t, opt);
    }
}
```

## 新增加解密接口

### cryptocalc

函数说明：加解密计算

参数说明：

- `(bytes memory) sourcecode，输入字节`
- `(bytes memory) key，公钥/私钥`
- `(bytes memory) destcode，输出字节`
- `uint256 type，操作类型：加密 0  ,解密 1`

返回值：

- `uint256 len，输出字节长度`

函数原型：`uint256 len = cryptocalc( bytes sourcecode, bytes key, bytes destcode, uint256 type);`

示例：

```
pragma solidity ^0.4.24;

contract Test {
   function getcrypto(bytes sourcecode, bytes key) public returns(bytes) {
        bytes memory destcode = new bytes(20);
        uint256 len = cryptocalc(sourcecode, key, destcode, 0);
        return destcode;
    }
}
```

## 新增账户相关接口

### getaccountid

函数说明：获取账户ID

参数说明：

- `string name，账户名`

返回值：

- `uint256 accountID，账户ID(错误返回0)`

函数原型：`uint256 accountID = getaccountid(string name)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
   function getAccountID(string name) public returns(uint256) {
        return getaccountid(name);
    }
}
```



### getaccounttime

函数说明：获取账户创建时间

参数说明：

- `uint256 account，账户ID`

返回值：

- `uint256 time，时间(错误返回0)`

函数原型：`uint256 time = getaccounttime(address account)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
   function gettime(uint256 account) public returns(uint256) {
        return getaccounttime(account);
    }
}
```

## 新增手续费相关接口

### withdrawfee

函数说明：回收手续费

参数说明：

- `uint256 id，可以为资产ID，合约账户名ID，矿工账户名ID`
- `uint256 typeId, 用来表示上一个参数的类型，当为0的时候，id代表资产ID，当为1的时候，id代表合约账户ID，当为2的时候，id代表矿工账户ID`

返回值：无

函数原型：`withdrawfee(uint256 id,uint256 typeId)`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function withdrawFee(uint256 id, uint256 typeId) public payable {
        withdrawfee(id, typeId);
    }
}
```

### deductgas

函数说明：扣除gas

参数说明：

- `uint256 amount，要扣除的gas数量`

返回值：

- `uint256 remain，剩余gas数量`

函数原型：`uint256 remain = deductgas ( uint256 amount);`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function deductMoreGas(uint256 amount) public returns(uint256){
        return deductgas(amount);
    }
}
```

## 新增DPOS相关接口

### getepoch

函数说明：获取生产周期

参数说明：

- `uint256 arg，0表示获取最新周期；非0，获取制定周期的前一周期`

返回值：

- `uint256 epoch，周期`

函数原型：`uint256 epoch = getepoch( uint256 arg);`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function getlatestEpochAndCertainEpoch(uint256 arg, uint256 epochID) public returns(uint256,uint256){
        uint256 latestEpoch = getepoch(0,epochID);
        uint256 certainEpoch = getepoch(arg,epochID);
        return (latestEpoch, certainEpoch);
    }
}
```

### getcandidatenum

函数说明：获取指定周期生产者个数

参数说明：

- `uint256 epoch，指定周期`

返回值：

- `uint256 num，生产者个数`

函数原型：`uint256 num = getcandidatenum( uint256 epoch );`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function getNum(uint256 epoch) public returns(uint256){
        uint256 num = getcandidatenum(epoch);
        return num;
    }
}
```

### getcandidate

函数说明：获取指定周期生产者信息

参数说明：

- `uint256 epoch， 指定周期`
- `uint256 index，生产者序号`
- `(bytes memory) name， 生产者名称` 

返回值：

- `uint256 stake，抵押权益`
- `uint256 cunt，应出块个数`
- `uint256 actcunt，实际出块个数`
- `uint256 replace，备用替换索引`
- `uint256 len，生产者名称字符串长度`

函数原型：`(uint256 stake, uint256 cunt, uint256  actcunt , uint256 replace, uint256 len) = getcandidate( uint256 epoch ，uint256 index ，bytes name);`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function candidate(uint256 epoch, uint256 index) public returns(uint256,uint256,uint256,uint256){
        bytes memory asname = new bytes(20);
        (uint256 stake,uint256 cunt,uint256  actcunt ,uint256 replace, uint256 len) = getcandidate(epoch,index,asname);
        return (stake, cunt, actcunt, replace);
    }
}
```

### getvoterstake

函数说明：获取指定周期投票者权益

参数说明：

- `uint256 epoch，指定周期`
- `uint256 voterID，指定投票者账户ID`
- `uint256 candidateID，指定生产者账户ID`

返回值：

- `uint256 stake，投票权益`

函数原型：`uint256 stake  = getvoterstake( uint256 epoch ，uint256 voterID, uint256 candidateID);`

示例：

```
pragma solidity ^0.4.24;

contract Test {
    function voterStake(uint256 epoch, uint256 voterID, uint256 candidateID) public returns(uint256){
        return getvoterstake(epoch, voterID, candidateID);
    }
}
```

## 新增属性

### assetid

属性介绍：新增msg.assetid，可以在合约中获取当前合约交易的资产ID

### recipient

属性介绍：新增tx.recipient，可以在合约中获取当前最外层交易的接收方

## 不可用特性

### 不再支持合约中创建账户或者合约

由于创建账户需要指定账户名，所以不再支持通过new的方式创建账户

### 不再支持ether，wei等单位

删除ether，wei等单位，对应修改为ft，aft，编写合约时请注意，两者差值为10的18次方