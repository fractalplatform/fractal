**账号**：

- [创建账号](#创建账号)
- [更新账号](#更新账号)
- [更新权限](#更新权限)

**资产**：

- [发行资产](#发行资产)
- [增发资产](#增发资产)
- [更改 owner](#更改owner)
- [更新资产](#更新资产)
- [销毁资产](#销毁资产)
- [更新资产合约](#更新资产合约)

**转账**：

- [转账](#转账)

**dpos**：

- [注册候选者](#注册候选者)
- [更新候选者](#更新候选者)
- [注销候选者](#注销候选者)
- [候选者取回抵押金](#候选者取回抵押金)
- [投票](#投票)
- [将候选者加入黑名单](#将候选者加入黑名单)
- [退出系统接管](#退出系统接管)
- [移除黑名单](#移除黑名单)

**合约**：

- [部署合约](#部署合约)
- [合约方法调用](#合约方法调用)

## 模块：账号

功能：

#### 创建账号

action：

```
AType:    actionType,  //必填项 CreateAccount 0x100
  Nonce:    nonce,  //必填项
  AssetID:  assetID, //必填项 转账资产ID
  From:     from,  //必填项
  To:       to,   //必填项，必为系统账号fractal.account
  GasLimit: gasLimit,  //必填项
  Amount:   new(big.Int), //转账时>0 不转账=0
  Payload:  payload,  //必填项
  Remark:   remark,   // 备注信息
```

payload：

```
type CreateAccountAction struct {
 AccountName common.Name   `json:"accountName,omitempty"`  //必填项 账户名首字符为小写字母开头，其余部分为小写字母和数字组合，主账户名长度12-16位，子账户名长度2-16位，子账户名全称为“主账户名.子账户名”
 Founder     common.Name   `json:"founder,omitempty"` //可填空 填空默认为账号本身
 PublicKey   common.PubKey `json:"publicKey,omitempty"` //账号公钥 填空为废账号
 Description string        `json:"description,omitempty"` //描述字段 限长 255
}
```

功能：

#### 更新账号

action：

```
AType:    actionType, //必填项 UpdateAccount 0x101
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 更新账号本身
  To:       to,   //必填项，必为系统账号fractal.account
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
type UpdataAccountAction struct {
 Founder     common.Name   `json:"founder,omitempty"` //必填项 可为空 为空默认为账号本身
}
```

功能：

#### 更新权限

action：

```
AType:    actionType, //必填项 UpdateAccountAuthor 0x103
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 更新账号本身
  To:       to, //必填项，必为系统账号fractal.account
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
type AccountAuthorAction struct {
 Threshold             uint64          `json:"threshold,omitempty"`
 UpdateAuthorThreshold uint64          `json:"updateAuthorThreshold,omitempty"`
 AuthorActions         []*AuthorAction `json:"authorActions,omitempty"`
}
```

data structure desc:

```
type AuthorAction struct {
 ActionType AuthorActionType //必填项 0添加 1更新 2删除
 Author     *common.Author //必填项
}

type Author struct {
 Owner  `json:"owner"`         //用户名，地址或公钥
 Weight uint64 `json:"weight"` //必填项 权重
}

Owner interface {
 String() string
}
```

## 模块：资产

功能：

#### 发行资产

action：

```
AType:    actionType, // 必填项 IssueAsset 0x201
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项
  To:       to, //必填项，必为系统账号fractal.asset
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
type IssueAsset struct {
 AssetName   string      `json:"assetName,omitempty"` //必填项 资产名由两部分组成，即“账户名:主资产名“　或　"账户名:主资产名.子资产名"。首字符必须为小写字母，其余部分为小写字母和数字组合，主资产名长度为2-16位，子资产名长度为1-8位。
 Symbol      string      `json:"symbol,omitempty"`//必填项 a-z0-9  2-16位 可重复
 Amount      *big.Int    `json:"amount"` //必填项 可为0
 Decimals    uint64      `json:"decimals"` //必填项 大于等于0
 Founder     common.Name `json:"founder,omitempty"` //必填项 可为空 为空默认为from
 Owner       common.Name `json:"owner,omitempty"` //必填项 不可为空 有效账号
 UpperLimit  *big.Int    `json:"upperLimit"` //必填项  0为不设上限
 Contract    common.Name `json:"contract"` //为空表示非协议资产。非空为协议资产账户
 Description string      `json:"detail"`   //资产描述字段 限长 255
}
```

功能：

#### 增发资产

action：

```
AType:    actionType, //必填项 IncreaseAsset 0x200
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 资产owner
  To:       to, //必填项，必为系统账号fractal.asset
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
type IncAsset struct {
 AssetId  uint64      `json:"assetId,omitempty"` //必填项 增发资产ID
 Amount   *big.Int    `json:"amount,omitempty"`  //增发数量 不可为0
 To common.Name `json:"acceptor,omitempty"` //接收资产账号
}
```

功能：

#### 更改 owner

action：

```
AType:    actionType，//必填项 SetAssetOwner 0x203
    Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 资产owner
  To:       to, //必填项，必为系统账号fractal.asset
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
type UpdateAssetOwner struct {
 AssetId    uint64      `json:"assetId"`  //必填项 更改owner资产ID
 Owner      common.Name `json:"owner,omitempty"`//必填项 有效账号 不可为空
}
```

功能：

#### 更新资产

action：

```
AType:    actionType,//必填项 UpdateAsset 0x204
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 资产owner
  To:       to, //必填项，必为系统账号fractal.asset
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
type UpdateAsset struct {
 AssetID    uint64      `json:"assetId"`  //必填项 更改owner资产ID
 Founder    common.Name `json:"founder,omitempty"` //必填项 有效账号
}
```

功能：

#### 销毁资产

action：

```
AType:    actionType,//必填项 DestroyAsset 0x202
  Nonce:    nonce,//必填项
  AssetID:  assetID, //必填项 销毁资产ID
  From:     from, //销毁资产账号
  To:       to,//必填项，必为系统账号fractal.asset
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int),//必填项 销毁资产数量
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
nil //填充此字段不影响交易结果
```

功能：

#### 更新资产合约

action：

```
AType:    actionType,//必填项 UpdateAssetContract 0x206
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 资产owner
  To:       to, //必填项，必为系统账号fractal.asset
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
type UpdateAssetContract struct {
	AssetID  uint64      `json:"assetId,omitempty"` //资产ID
	Contract common.Name `json:"contract"` //协议资产账户
}
```

## 模块：转账

功能：

#### 转账

action：

```
AType:    actionType, //必填项 Transfer 0x205
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 转账资产ID
  From:     from, //必填项
  To:       to, //必填项
  GasLimit: gasLimit,//必填项
  Amount:   new(big.Int), //转账金额
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
nil //填充此字段不影响交易结果
```

## 模块：dpos

[fractal 共识介绍](#https://github.com/fractalplatform/fractal/wiki/fractal%E5%85%B1%E8%AF%86%E4%BB%8B%E7%BB%8D)

功能：

#### 注册候选者

action：

```
AType:    actionType, //必填项 RegCandidate 0x300
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 注册候选者
  To:       to, //必填项 fractal.dpos
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 必须为50万FT(500000000000000000000000)
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
RegisterCandidate{
  Url:   url,  //必填项 可为空
 }
```

功能：

#### 更新候选者

action：

```
AType:    actionType, //必填项 UpdateCandidate 0x301
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 注册候选者
  To:       to, //必填项 fractal.dpos
  GasLimit: gasLimit, //必填项
  Amount:   new(big.Int), //必填项 0
  Payload:  payload,
  Remark:   remark,   // 备注信息
```

payload：

```
UpdateCandidate{
  Url:   url,  //必填项 可为空
 }
```

功能：

#### 注销候选者

action：

```
AType:    actionType, //必填项 UnregCandidate 0x302
  Nonce:    nonce, //必填项
  AssetID:  assetID, //必填项 有效资产ID
  From:     from, //必填项 注册候选者
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
