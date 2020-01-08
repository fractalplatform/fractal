> **所有交易的 payload 均由 ABI 进行编解码，ABI 细节参考[这里][ethabi]**。

**账号**：

- [创建账号](#创建账号)
- [更新公钥](#更新公钥)

**资产**：

- [发行资产](#发行资产)
- [增发资产](#增发资产)

**转账**：

- [转账](#转账)

**道具**

- [发行 World](#发行World)
- [更新 World 的 Owner](#更新World的Owner)
- [发行 Item 类型](#发行Item类型)
- [增发 item](#增发item)
- [销毁 item](#销毁item)
- [增发 items](#增发items)
- [销毁 items](#销毁items)
- [交易 item](#交易item)
- [增加 item 类型的属性](#增加item类型的属性)
- [删除 item 类型的属性](#删除item类型的属性)
- [修改 item 类型的属性](#修改item类型的属性)
- [增加 item 的属性](#增加item的属性)
- [删除 item 的属性](#删除item的属性)
- [修改 item 的属性](#修改item的属性)

**dpos**：

- [注册候选者](#注册候选者)
- [注销候选者](#注销候选者)

**合约**：

- [部署合约](#部署合约)
- [合约方法调用](#合约方法调用)

## 模块：账号

功能：

#### 创建账号

pluginTX：

```
PType: PayloadType      //必填项 CreateAccount envelope.PayloadType = 0x100
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  name: 账户名
//  pubkey: 账户公钥
//  desc: 描述
CreateAccount(string name, string pubKey, string desc)
```

功能：

#### 更新公钥

pluginTX：

```
PType: PayloadType      //必填项 ChangePubKey envelope.PayloadType = 0x101
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  pubkey: 账户公钥
function ChangePubKey(string pubKey)
```

## 模块：资产

功能：

#### 发行资产

pluginTX：

```
PType: PayloadType      //必填项 IssueAsset envelope.PayloadType = 0x200
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalasset
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  name: 资产名
//  symbol: 资产名简称
//  amount: 发行数量
//  decimals: 精度
//  founder: founder账户
//  owner: owner账户
//  limit: 上限
//  desc: 描述
function IssueAsset(string name ,string symbol ,uint256 amount, uint64 decimals, string founder, string owner, uint256 limit, string desc)
```

功能：

#### 增发资产

pluginTX：

```
PType: PayloadType      //必填项 IncreaseAsset envelope.PayloadType = 0x201
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalasset
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  to: 增发到账户
//  assetID: 资产id
//  amount: 增发数量
function IncreaseAsset(string to, uint64 assetID, uint256 amount)
```

## 模块：转账

功能：

#### 转账

pluginTX：

```
PType: PayloadType      //必填项 Transfer envelope.PayloadType = 0x202
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  to: 账户公钥
//  assetid: 资产id
//  value: 金额
function Transfer(string to, uint64 assetid, uint256 value)
```

## 模块：道具

功能：

#### 发行 World

pluginTX：

```
PType: PayloadType      //必填项 IssueWorld envelope.PayloadType = 0x400
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  owner: owner账户
//  name: world名字
//  description: 描述
function IssueWorld(string owner, string name, string description)
```

功能：

#### 更新 World 的 Owner

pluginTX：

```
PType: PayloadType      //必填项 UpdateWorldOwner envelope.PayloadType = 0x401
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  owner: owner账户
//  worldID: ID
function UpdateWorldOwner(string owner, uint64 worldID)
```

功能：

#### 发行 Item 类型

pluginTX：

```
PType: PayloadType      //必填项 IssueItemType envelope.PayloadType = 0x402
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  name: item类型名字
//  merge: 同质标识
//  upperLimit: 发行上限
//  description: 描述
//  attrPermission: 属性修改权限
//  attrName: 属性名称
//  attrDes: 属性描述
function IssueItemType(uint64 worldID, string name, bool merge, uint64 upperLimit, string description, uint64[] attrPermission, string[] attrName, string[] attrDes)
```

功能：

#### 增发 item

pluginTX：

```
PType: PayloadType      //必填项 IncreaseItem envelope.PayloadType = 0x403
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  owner: owner账户
//  description: 描述
//  attrPermission: 属性修改权限
//  attrName: 属性名称
//  attrDes: 属性描述
function IncreaseItem(uint64 worldID, uint64 itemTypeID, string owner, string description, uint64[] attrPermission, string[] attrName, string[] attrDes)
```

功能：

#### 销毁 item

pluginTX：

```
PType: PayloadType      //必填项 DestroyItem envelope.PayloadType = 0x404
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  itemID: item ID
function DestroyItem(uint64 worldID, uint64 itemTypeID, uint64 itemID)
```

功能：

#### 增发 items

pluginTX：

```
PType: PayloadType      //必填项 IncreaseItems envelope.PayloadType = 0x405
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  to: 增发到账户
//  amount: 增发数量
function IncreaseItems(uint64 worldID, uint64 itemTypeID, string to, uint64 amount)
```

功能：

#### 销毁 items

pluginTX：

```
PType: PayloadType      //必填项 DestroyItems envelope.PayloadType = 0x406
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  amount: 增发数量
function DestroyItems(uint64 worldID, uint64 itemTypeID, uint64 amount)
```

功能：

#### 交易 item

pluginTX：

```
PType: PayloadType      //必填项 TransferItem envelope.PayloadType = 0x407
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  to: 交易到账户
//  worldID: world ID
//  itemTypeID: item类型ID
//  itemID: item ID   如果交易同质道具，此处填0
//  amount: 交易数量   如果交易非同质道具，此处填0
function TransferItem(string to, uint64[] worldID, uint64[] itemTypeID, uint64[] itemID, uint64[] amount)
```

功能：

#### 增加 item 类型的属性

pluginTX：

```
PType: PayloadType      //必填项 AddItemTypeAttributes envelope.PayloadType = 0x408
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  attrPermission: 属性修改权限
//  attrName: 属性名称
//  attrDes: 属性描述
function AddItemTypeAttributes(uint64 worldID, uint64 itemTypeID, uint64[] attrPermission, string[] attrName, string[] attrDes)
```

功能：

#### 删除 item 类型的属性

pluginTX：

```
PType: PayloadType      //必填项 DelItemTypeAttributes envelope.PayloadType = 0x409
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  attrName: 属性名称
function DelItemTypeAttributes(uint64 worldID, uint64 itemTypeID, string[] attrName)
```

功能：

#### 修改 item 类型的属性

pluginTX：

```
PType: PayloadType      //必填项 ModifyItemTypeAttributes envelope.PayloadType = 0x40a
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  attrPermission: 属性修改权限
//  attrName: 属性名称
//  attrDes: 属性描述
function ModifyItemTypeAttributes(uint64 worldID, uint64 itemTypeID, uint64[] attrPermission, string[] attrName, string[] attrDes)
```

功能：

#### 增加 item 的属性

pluginTX：

```
PType: PayloadType      //必填项 AddItemAttributes envelope.PayloadType = 0x40b
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  itemTypeID: item ID
//  attrPermission: 属性修改权限
//  attrName: 属性名称
//  attrDes: 属性描述
function AddItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, uint64[] attrPermission, string[] attrName, string[] attrDes)
```

功能：

#### 删除 item 的属性

pluginTX：

```
PType: PayloadType      //必填项 DelItemAttributes envelope.PayloadType = 0x40c
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  itemID: item ID
//  attrName: 属性名称
function DelItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, string[] attrName)
```

功能：

#### 修改 item 的属性

pluginTX：

```
PType: PayloadType      //必填项 ModifyItemTypeAttributes envelope.PayloadType = 0x40d
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractalaccount
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

Payload ABI:

```
// 参数
//  worldID: world ID
//  itemTypeID: item类型ID
//  itemID: item ID
//  attrPermission: 属性修改权限
//  attrName: 属性名称
//  attrDes: 属性描述
function ModifyItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, uint64[] attrPermission, string[] attrName, string[] attrDes)
```

## 模块：dpos

pluginTX：

```
PType: PayloadType      //必填项 IssueAsset envelope.PayloadType = 0x200
From:     from,         //必填项
To:       to,           //必填项，必为系统账号fractaldpos
Nonce:    nonce,        //必填项
AssetID:  assetID,      //必填项 转账资产ID
GasAssetID:  assetID,   //必填项 gas消耗的资产ID
GasLimit: gasLimit,     //必填项
GasPrice:   new(big.Int), //转账时>0 不转账=0
Amount:   new(big.Int), //转账时>0 不转账=0
Payload:  payload,      //必填项
Remark:   remark,       // 备注信息
```

功能：

#### 注册候选者

- 抵押金额为`tx.Amount`
- 重复发送该交易可以追加抵押金额, 更新签名账户

Payload ABI：

```
// 参数
//  signer: 负责区块签名的账户
RegisterMiner(string signer);
}
```

功能：

#### 注销候选者

Payload ABI:

```
UnregisterMiner();
```

功能：

#### 候选者取回抵押金

action：

```
AType:    actionType, //必填项 RefundCandidate 0x303
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 投票人
  To:       to, //必填项 fractal.dpos
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
nil
```

功能：

#### 投票

action：

```
AType:    actionType, //必填项 VoteCandidate 0x304
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 投票人
  To:       to, //必填项 fractal.dpos
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
VoteCandidate{
  Candidate: common.Name, //必填项 有效已注册候选者
  Stake:    big.Int, //必填项 投票量
 }
```

功能：

#### 将候选者加入黑名单

action：

```
AType:    actionType,//必填项 KickedCandidate 0x400
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 系统账号 fractal.founder
  To:       to, //必填项 fractal.dpos
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
KickedCandidate{
  Candidates: []common.Name, //必填项 已注册候选人列表
 }
```

功能：

#### 退出系统接管

action：

```
AType:    actionType, //必填项 ExitTakeOver 0x401
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 系统账号 fractal.founder
  To:       to, //必填项 fractal.dpos
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
nil
```

功能：

#### 移除黑名单

action：

```
AType:    actionType, //必填项 RemoveKickedCandidate 0x402
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 系统账号 fractal.founder
  To:       to, //必填项 fractal.dpos
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
RemoveKickedCandidate{
  Candidates: []common.Name, //必填项 已加入黑名单的注册候选人列表
 }
```

## 模块：合约

功能：

#### 部署合约

action：

```
AType:    actionType,//必填项 CreateContract 0x1
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 转账资产ID
  From:     from, //必填项
  To:       to, //必填项 合约账号
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //转账数量
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
合约 bytes code
```

功能：

#### 合约方法调用

action：

```
AType:    actionType, //必填项 CallContract 0x0
  Nonce:    nonce, //必填项
  AssetID:  assetID, //转账资产ID
  From:     from, //必填项
  To:       to, //必填项 合约账号
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int),//转账数量
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
bytes 合约方法名，方法参数
```

[solidity 修改项](#https://github.com/fractalplatform/solidity/wiki/solidity%E4%BF%AE%E6%94%B9%E9%A1%B9)

[ethabi]: https://solidity.readthedocs.io/en/v0.6.0/abi-spec.html
