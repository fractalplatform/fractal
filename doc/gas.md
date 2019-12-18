## `GAS` 消耗值

执行普通交易或者合约交易消耗的`gas`数量。

| `gas`名称                 | `gas`    | 使用场景                                              |
| ------------------------- | -------- | ----------------------------------------------------- |
| `ActionGas`               | 100000   | 交易基础`gas`                                         |
| `ActionGasCallContract`   | 200000   | 调用合约基础`gas`                                     |
| `ActionGasCreation`       | 500000   | 创建账户及创建合约 交易基础`gas`                      |
| `ActionGasIssueAsset`     | 10000000 | 发行资产 交易基础`gas`                                |
| `TxDataNonZeroGas`        | 68       | 交易 payload/remark 非 0 的每字节`gas`                |
| `TxDataZeroGas`           | 4        | 交易 payload/remark 为 0 的每字节`gas`                |
| `SignGas`                 | 50000    | 交易签名如果超过一个，则每多出一个签名，需花费的`gas` |
|                           |          |                                                       |
| `gt.SetOwner`             | 200      | 执行`setowner` 所耗`gas`                              |
| `gt.WithdrawFee`          | 700      | 执行`WithdrawFee` 所耗的`gas`                         |
| `gt.GetAccountTime`       | 200      | 执行`GetAccountTime` 所耗的`gas`                      |
| `gt.GetSnapshotTime`      | 200      | 执行`GetSnapshotTime` 所耗的`gas`                     |
| `gt.GetAssetInfo`         | 200      | 执行`GetAssetInfo` 所耗的`gas`                        |
| `gt.SnapBalance`          | 200      | 执行`SnapBalance` 所耗的`gas`                         |
| `gt.IssueAsset`           | 10000000 | 执行`IssueAsset` 所耗的`gas`                          |
| `gt.DestroyAsset`         | 200      | 执行`DestroyAsset` 所耗的`gas`                        |
| `gt.AddAsset`             | 200      | 执行`AddAsset` 所耗的`gas`                            |
| `gt.GetAccountID`         | 200      | 执行`GetAccountID` 所耗的`gas`                        |
| `gt.GetAssetID`           | 200      | 执行`GetAssetID` 所耗的`gas`                          |
| `gt.CryptoCalc`           | 20000    | 执行`CryptoCalc` 所耗的`gas`                          |
| `gt.CryptoByte`           | 1000     | 执行`CryptoByte` 所耗的`gas`                          |
| `gt.DeductGas`            | 200      | 执行`DeductGas` 所耗的`gas`                           |
| `gt.GetEpoch`             | 200      | 执行`GetEpoch` 所耗的`gas`                            |
| `gt.GetCandidateNum`      | 200      | 执行`GetCandidateNum` 所耗的`gas`                     |
| `gt.Candidate`            | 200      | 执行`Candidate` 所耗的`gas`                           |
| `gt.VoterStake`           | 200      | 执行`VoterStake` 所耗的`gas`                          |
|                           |          |                                                       |
| `gt.ExtcodeSize`          | 700      | 执行`ExtCodeSize` 所耗`gas`                           |
| `gt.ExtcodeCopy`          | 700      | 执行`ExtCodeCopy` 所耗`gas`                           |
| `gt.Balance`              | 400      | 执行`balance`和`balanceex` 所耗`gas`                  |
| `gt.SLoad`                | 200      | 执行`SLoad` 所耗`gas`                                 |
| `gt.Calls`                | 700      | 执行`Calls` 所耗`gas`                                 |
| `gt.ExpByte`              | 50       | 执行`ExpByte` 所耗`gas`                               |
| `gt.CallValueTransferGas` | 9000     | 执行`Calls`附带 Transfer 所耗`gas`                    |
| `gt.SstoreSetGas`         | 20000    | 执行`Sstore` 第一次存储所耗`gas`                      |
| `gt.SstoreResetGas`       | 5000     | 执行`Sstore` 重新存储所耗`gas`                        |
| `gt.JumpdestGas`          | 1        | 执行`JumpDest`所耗`gas`                               |
| `gt.CreateDataGas`        | 200      | create data 每字节`所耗`gas`                          |
| `gt.CopyGas`              | 3        | copy 所耗`gas`/字                                     |
| `gt.LogGas`               | 375      | Log 基础`gas`                                         |
| `gt.LogTopicGas`          | 375      | Log Topic 每个所耗`gas`                               |
| `gt.LogDataGas`           | 8        | Log data 每字节所耗`gas`                              |
| `gt.CreateGas`            | 32000    | 执行`Create` 所耗`gas`                                |
| `gt.MemoryGas`            | 3        | memory 所耗`gas`/字                                   |
| `gt.Sha3Gas`              | 30       | 执行`Sha3` 基础`gas`                                  |
| `gt.Sha3WordGas`          | 6        | 执行`Sha3` 所耗`gas`/字                               |
|                           |          |                                                       |
