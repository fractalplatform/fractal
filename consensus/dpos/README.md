## dpos生成者注册及投票者投票

### 参数说明
- `UnitStake`  一票所代表的权益
- `MaxURLLen`  生产者URL最大长度
- `ProducerMinQuantity` 生成者需要抵押的最低票数
- `VoterMinQuantity`   投票者需要抵押的最低票数
- `ActivatedMinQuantity` dpos启动需要的最低票数
- `BlockInterval`  每轮生产者出块间隔
- `BlockFrequency` 每轮生产者连续出块个数
- `ProducerScheduleSize`: 每轮生成者最大个数

### dpos启动
- 网络投票总数 >= `ActivatedMinQuantity`  && 已注册的生成者个数 >= `ProducerScheduleSize` * 2 / 3 + 1 
- 启动后, 启动条件必须一直满足

### 注册生产者/更新生产者 
> RegProducer(producer string, address string, url string, stake *big.Int) error
- 账户名必须存在
- 账户名与提供的公钥匹配
- 提供的URL长度 <= `MaxURLLen`
- 抵押stake 需满足`UnitStake`的整数倍， 且商 >= `ProducerMinQuantity`
- 未投票

### 注销生成者
> UnregProducer(producer string) error
- 有效生产者
- 注销后，已注册生产者个数 >= `ProducerScheduleSize` * 2 / 3 + 1
- 注销后，网络投票总数 >= `ActivatedMinQuantity`
- 未存在投票者投票

### 投票者投票
> VoteProducer(voter string, producer string, stake *big.Int) error
- 账户名必须存在
- 抵押stake 需满足`UnitStake`的整数倍， 且商 >= `VoterMinQuantity`
- 未注册生产者
- 未投票
- 有效的生产者

### 投票者改票
> ChangeProducer(voter string, producer string) error
- 已投票
- 不是相同的生产者
- 有效的生产者

### 投票者取消投票
> UnvoteProducer(voter string) error
- 已投票
- 取消后，网络投票总数 >= `ActivatedMinQuantity`

### 生产者取消投票者的投票
> UnvoteVoter(producer string, voter string) error
- 投票者已投票
- 投票的生产者确实是该生产者
- 取消后，网络投票总数 >= `ActivatedMinQuantity`
