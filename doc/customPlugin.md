## 编写自定义Plugin
### interface.go
1. 在`interface.go`文件中定义新接口
```go
type ICustomPlugin interface {
    ...
    ...
}
```
2. 将新接口添加到`IPM`中
```go
type IPM interface {
	IAccount
	IAsset
	IConsensus
	IFee
	ISigner
	ICustomPlugin  // custom plugin
	ExecTx(tx *types.Transaction, fromSol bool) ([]byte, error)
	IsPlugin(name string) bool
	InitChain(json.RawMessage, *params.ChainConfig) ([]*types.Transaction, error)
}
```
### plugin.go
1. 在`Manager`结构体中添加自定义新接口
```go
type Manager struct {
	stateDB         *state.StateDB
	contracts       map[string]IContract
	contractsByType map[envelope.PayloadType]IContract
	IAccount
	IAsset
	IConsensus
	IFee
	ISigner
	ICustomPlugin  // custom plugin
}
```
2. 在`func NewPM(stateDB *state.StateDB) IPM`函数中添加自定义插件的`New`函数。根据自己的需求将自定义插件的实例存入`Manager`结构体中的`contracts`和`contractsByType`变量中。
```go
func NewPM(stateDB *state.StateDB) IPM {
    ...
    ...
	custom, _ := NewCustom()
	pm := &Manager{
	    ...
		ICustomPlugin:   custom,
		...
	}
	pm.contracts[custom.AccountName()] = custom
	//pm.contractsByType[PayloadType] = custom
	...
}
```
3. 自定义插件可以在`InitChain`函数中构建链初始化交易。
### 自定义插件文件：custom.go
1. 定义基础数据结构实现`ICustomPlugin`接口中的函数
```go
type Custom struct {
	sdb *state.StateDB  // 用于数据存储
}
```
2. 自定义插件能被交易调用，需实现`IContract`接口
```go
func (c *Custom) AccountName() string {
	return "fractalcustom"
}

func (c *Custom) CallTx(tx *envelope.PluginTx, pm IPM) ([]byte, error) {
	switch tx.PayloadType() {
	case envelope.PayloadType:
    	...
		return nil, err
	case envelope.PayloadType:
		...
		return nil, err
	}
	return nil, ErrWrongTransaction
}
```
### 插件实现api
参考[plugin.md](plugin.md)文档。  

注意：需在`plugin.go`中的`init`函数中注册。