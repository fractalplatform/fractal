
> **所有交易的payload均由ABI进行编解码，ABI细节参考[这里][ethabi]**。

**账号**：

- [创建账号](#创建账号)
- [更新公钥](#更新公钥)

**资产**：

- [发行资产](#发行资产)
- [增发资产](#增发资产)

**转账**：

- [转账](#转账)

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

payload：

```
type ChangePubKeyAction struct {
	Pubkey string    //必填项 账号公钥
}
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

payload：

```
type IssueAssetAction struct {
	AssetName   string      //必填项 资产名首字符为小写字母开头，其余部分为小写字母和数字组合，账号长度为2-32位
	Symbol      string      //必填项 同资产名格式相同
	Amount      *big.Int    //必填项 可为0
	Owner       string      //必填项 不可为空 有效账号
	Founder     string      //必填项 可为空 为空默认为from
	Decimals    uint64      //必填项 大于等于0
	UpperLimit  *big.Int    //必填项  0为不设上限
	Description string      //资产描述字段 限长 255
}
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

payload：

```
type IncreaseAssetAction struct {
	AssetID uint64        //必填项 增发资产ID
	Amount  *big.Int      //增发数量 不可为0
	To      string        //接收资产账号
}
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

payload：

```
nil //填充此字段不影响交易结果
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