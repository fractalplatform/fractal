## Command Line Options

### ft

ft is a Leading High-performance Ledger

```
ft is a Leading High-performance Ledger

Usage:
  ft [flags]
  ft [command]

Available Commands:
  chain       state裁剪使能或禁用
  debug       pprof 调试
  export      导出区块数据到文件
  help        显示帮助
  import      导入区块数据
  init        初始化创世块
  miner       控制挖矿命令
  p2p         P2P 网络设置和查询
  rpc         RPC调用
  txpool      查询txpool状态和设置txpool最小gas价格
  version     显示版本信息

Flags:
  -c, --config string                       TOML/YAML 配置文件
      --contractlog bool                    存储内部交易日志开关 (default false)
      --database_cache int                  数据库内存缓存分配(单位: MB) (default 768)
  -d, --datadir string                      数据存储目录 (default "$HOME/.ft_ledger")
      --debug_blockprofilerate int          按指定频率打开block profiling
      --debug_cpuprofile string             将CPU profile写入指定文件
      --debug_memprofilerate int            按指定频率打开memory profiling (default 524288)
      --debug_pprof                         启用pprof HTTP服务器
      --debug_pprof_addr string             pprof HTTP服务器监听地址 (default "localhost")
      --debug_pprof_port int                pprof HTTP服务器监听端口 (default 6060)
      --debug_trace string                  将execution trace写入指定文件
  -g, --genesis string                      创世块json文件
      --gpo_blocks int                      用于检查gas价格的最近块的个数 (default 20)
      --gpo_percentile int                  建议gas价参考最近交易的gas价的百分位数 (default 60)
  -h, --help                                显示帮助
      --http_cors strings                   允许跨域请求的域名列表 (default [*])
      --http_host string                    HTTP RPC服务器地址 (default "localhost")
      --http_modules strings                基于HTTP RPC接口提供的API (default [ft,dpos,fee,account])
      --http_port int                       HTTP RPC服务器端口 (default 8545)
      --http_vhosts strings                 HTTP RPC虚拟地址 (default [localhost])
      --ipcpath string                      ipc 文件存储路径 (default "ft.ipc")
      --log_backtrace string                请求特定日志记录堆栈跟踪 (e.g. "block.go:271")
      --log_debug                           突出显示调用位置日志 (文件名和行号)
      --log_dir string                      将日志记录写入指定文件
      --log_level int                       日志显示等级设置: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail (default 3)
      --log_module string                   每个模块的日志显示等级设置: comma-separated list of <pattern>=<level> (e.g. ft/*=5,p2p=4)
      --metrics_influxdb                    启动数据库存储统计指标
      --metrics_influxdb_URL string         metrics数据库 URL (default "http://localhost:8086")
      --metrics_influxdb_name string        metrics数据库名 (default "metrics")
      --metrics_influxdb_namespace string   metrics命名空间 (default "fractal/")
      --metrics_influxdb_passwd string      metrics数据库密码
      --metrics_influxdb_user string        metrics数据库用户名
      --metrics_start                       启用metrics收集和报告 flag that open statistical metrics
      --miner_extra string                  区块扩展信息 (default "system")
      --miner_delay uint                    矿工延迟广播间隔（default 0ms）
      --miner_name string                   挖矿奖励账户 (default "fractal.admin")
      --miner_private strings               挖矿奖励私钥 (default [289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032])
      --miner_start                         开启挖矿
      --p2p_bootnodes string                节点列表路径，用于与其他节点建立连接
      --p2p_dialratio int                   P2P 主动连接和被动连接的连接比率
      --p2p_id uint                         P2P 网络ID，节点的ID不同就不能互相通信
      --p2p_listenaddr string               P2P 监听端口 (default ":2018")
      --p2p_maxpeers int                    P2P 最大连接数量 (default 10)
      --p2p_maxpendpeers int                P2P 最大连接握手数量
      --p2p_name string                     P2P 节点名称 (default "Fractal-P2P")
      --p2p_nodedb string                   配置数据库路径，在网络中发现的节点写入数据库，默认保存在内存中
      --p2p_nodial bool                     True 不主动连接任何节点 (default False)
      --p2p_nodiscovery                     禁用节点发现机制 (手动添加节点)
      --p2p_staticnodes string              静态节点列表路径，节点在断开时自动重新连接
      --p2p_trustnodes string               信任节点列表路径  在达到节点最大连接数时允许节点连接
      --statepruning_enable bool            使能或禁止使用裁剪state (default true)
      --txpool_accountqueue uint            每个账户允许不可执行最大交易槽数量 (default 1280)
      --txpool_accountslots uint            每个账户保证可执行交易槽数量 (default 128)
      --txpool_globalqueue uint             所有账户不可执行最大交易槽数量 (default 4096)
      --txpool_globalslots uint             所有账户可执行最大交易槽数量 (default 4096)
      --txpool_journal string               存储本地交易日志，用于节点重启 (default "transactions.rlp")
      --txpool_lifetime duration            不可执行交易在队列中的最大生命周期 (default 3h0m0s)
      --txpool_nolocals                     禁用对本地交易价格豁免
      --txpool_pricebump uint               价格上涨百分比，取代已有交易 (default 10)
      --txpool_pricelimit uint              加入交易池的最小gas价格限制 (default 1)
      --txpool_rejournal duration           重新生成本地交易日志的时间间隔 (default 1h0m0s)
      --ws_exposeall                        基于websocket RPC 接口公开所有API模块，不仅仅是公共模块
      --ws_host string                      基于websocket RPC 服务器监听地址 (default "localhost")
      --ws_modules strings                  基于websocket RPC 接口提供的API (default [ft])
      --ws_origins strings                  基于websocket RPC 请求允许的源
      --ws_port int                         基于websocket RPC 服务器监听端口 (default 8546)

Use "ft [command] --help" for more information about a command.

```

### ftfinder

ftfinder is a fractal node discoverer, for more information,see:

```
ftfinder is a fractal node discoverer

Usage:
  ftfinder [flags]
  ftfinder [command]

Available Commands:
  help        Help about any command
  version     Show current version

Flags:
  -d, --datadir string          数据存储目录 (default "$HOME/.ft_ledger")
  -h, --help                    显示帮助
      --p2p_bootnodes string    节点列表路径，用于与其他节点建立连接
      --p2p_id uint             P2P 网络ID，节点的ID不同就不能互相通信
      --p2p_listenaddr string   P2P 监听端口 (default ":2018")
      --p2p_nodedb string       配置数据库路径，在网络中发现的节点写入数据库，默认保存在内存中

Use "ftfinder [command] --help" for more information about a command.

```

### 命令行 IPC 访问节点

- **chain**：
  - [startpure](#startpure)
- **debug**
  - [blockprofile](#blockprofile)
  - [cpuprofile](#cpuprofile)
  - [freeosmemory](#freeosmemory)
  - [gcstats](#gcstats)
  - [gotrace](#gotrace)
  - [memstats](#memstats)
  - [mutexprofile](#mutexprofile)
  - [stacks](#stacks)
  - [writememprofile](#writememprofile)
- **miner**
  - [force](#force)
  - [mining](#mining)
  - [setdelay](#setdelay)
  - [setcoinbase](#setcoinbase)
  - [setextra](#setextra)
  - [start](#start)
  - [stop](#stop)
- **p2p**
  - [add](#add)
  - [addbad](#addbad)
  - [addtrusted](#addtrusted)
  - [badcount](#badcount)
  - [badlist](#badlist)
  - [count](#count)
  - [list](#list)
  - [remove](#remove)
  - [removebad](#removebad)
  - [removetrusted](#removetrusted)
  - [selfnode](#selfnode)
- **txpool**
  - [content](#content)
  - [setgasprice](#setgasprice)
  - [status](#status)
- **[rpc](#RPC)**
  - [account_accountIsExist](#RPC)
  - [account_getAccountBalanceByID](#RPC)
  - [account_getAccountByName](#RPC)
  - [account_getAssetInfoByID](#RPC)
  - [account_getAssetInfoByName](#RPC)
  - [account_getCode](#RPC)
  - [account_getNonce](#RPC)
  - [bc_setStatePruning](#RPC)
  - [consensus_getAllCandidates](#RPC)
  - [consensus_getCandidateInfo](#RPC)
  - [debug_blockProfile](#RPC)
  - [debug_cpuProfile](#RPC)
  - [debug_freeOSMemory](#RPC)
  - [debug_gcStats](#RPC)
  - [debug_goTrace](#RPC)
  - [debug_memStats](#RPC)
  - [debug_mutexProfile](#RPC)
  - [debug_setBlockProfileRate](#RPC)
  - [debug_setGCPercent](#RPC)
  - [debug_setMutexProfileFraction](#RPC)
  - [debug_stacks](#RPC)
  - [debug_startCPUProfile](#RPC)
  - [debug_startGoTrace](#RPC)
  - [debug_stopCPUProfile](#RPC)
  - [debug_stopGoTrace](#RPC)
  - [debug_writeBlockProfile](#RPC)
  - [debug_writeMemProfile](#RPC)
  - [debug_writeMutexProfile](#RPC)
  - [ft_call](#RPC)
  - [ft_estimateGas](#RPC)
  - [ft_gasPrice](#RPC)
  - [ft_getBadBlocks](#RPC)
  - [ft_getBlockAndResultByNumber](#RPC)
  - [ft_getBlockByHash](#RPC)
  - [ft_getBlockByNumber](#RPC)
  - [ft_getChainConfig](#RPC)
  - [ft_getCurrentBlock](#RPC)
  - [ft_getInternalTxByAccount](#RPC)
  - [ft_getInternalTxByBloom](#RPC)
  - [ft_getInternalTxByHash](#RPC)
  - [ft_getTransactionByHash](#RPC)
  - [ft_getTransactionReceipt](#RPC)
  - [ft_getTransactions](#RPC)
  - [ft_sendRawTransaction](#RPC)
  - [item_getAccountItemAmount](#RPC)
  - [item_getItemAttribute](#RPC)
  - [item_getItemInfoByID](#RPC)
  - [item_getItemInfoByName](#RPC)
  - [item_getItemTypeByID](#RPC)
  - [item_getItemTypeByName](#RPC)
  - [miner_force](#RPC)
  - [miner_mining](#RPC)
  - [miner_setCoinbase](#RPC)
  - [miner_setDelay](#RPC)
  - [miner_setExtra](#RPC)
  - [miner_start](#RPC)
  - [miner_stop](#RPC)
  - [p2p_addBadNode](#RPC)
  - [p2p_addPeer](#RPC)
  - [p2p_addTrustedPeer](#RPC)
  - [p2p_badNodes](#RPC)
  - [p2p_badNodesCount](#RPC)
  - [p2p_peerCount](#RPC)
  - [p2p_peerEvents](#RPC)
  - [p2p_peers](#RPC)
  - [p2p_removeBadNode](#RPC)
  - [p2p_removePeer](#RPC)
  - [p2p_removeTrustedPeer](#RPC)
  - [p2p_seedNodes](#RPC)
  - [p2p_selfNode](#RPC)
  - [txpool_content](#RPC)
  - [txpool_getTransactions](#RPC)
  - [txpool_getTransactionsByAccount](#RPC)
  - [txpool_pendingTransactions](#RPC)
  - [txpool_setGasPrice](#RPC)
  - [txpool_status](#RPC)

---

#### startpure

```
使能或禁用裁剪 state

Usage:
  ft chain startpure <enable/disable> [flags]

Flags:
  -h, --help             显示帮助
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft chain startpure enable

# Result
{
  "preStatePruning": true,
  "currentNumber": 111140
}
```

---

#### blockprofile

```
按指定频率打开block profiling，并将数据写入文件

Usage:
  ft debug blockprofile <file> <nsec>  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### cpuprofile

```
按指定频率打开CPU profile，并将数据写入文件

Usage:
  ft debug cpuprofile <file> <nsec>  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### freeosmemory

```
将未使用的内存返回给操作系统

Usage:
  ft debug freeosmemory  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### gcstats

```
获取垃圾回收统计信息

Usage:
  ft debug gcstats  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### gotrace

```
按指定频率打开tracing，并将数据写入文件

Usage:
  ft debug gotrace <file> <nsec>  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### memstats

```
获取内存状态信息

Usage:
  ft debug memstats  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### mutexprofile

```
按指定频率打开mutex profiling，并将数据写入文件

Usage:
  ft debug mutexprofile <file> <nsec>  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### stacks

```
获取所有goroutines堆栈信息

Usage:
  ft debug stacks [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### writememprofile

```
将内存分配信息写入指定文件

Usage:
  ft debug writememprofile <file> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

---

#### force

```
强制启动挖矿（系统节点自治时使用）

Usage:
  ft miner force  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft miner force

# Result
true
```

---

#### mining

```
获取挖矿状态

Usage:
  ft miner mining [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft miner mining -i $HOME/.ft_ledger/ft.ipc

# Result
true
```

---

#### setcoinbase

```
设置挖矿账户

Usage:
  ft miner setcoinbase <name> <privKeys file> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft miner setcoinbase "miner_account", ~/privKeyFilePath

# Result
true
```

---

#### setextra

```
设置由该节点打包的区块的额外数据

Usage:
  ft miner setextra <extra> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft miner setextra "hello world"

# Result
true
```

---

#### setdelay

```
矿工延时广播间隔，单位毫秒

Usage:
  ft miner setdelay <delay> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft miner setdlay 100

# Result
true
```

---

#### start

```
启动挖矿

Usage:
  ft miner start [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft miner start

# Result
true
```

---

#### stop

```
停止挖矿

Usage:
  ft miner stop [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft miner stop

# Result
true
```

---

#### add

```
连接节点

Usage:
  ft p2p add <url> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p add "fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"

# Result
true
```

---

#### addbad

```
把节点添加到黑名单中

Usage:
  ft p2p addbad <url> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p addbad "fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"

# Result
true
```

---

#### addtrusted

```
把节点添加到信任节点集

Usage:
  ft p2p addtrusted <url> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p addtrusted "fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"

# Result
true
```

---

#### badcount

```
获取黑名单节点数量

Usage:
  ft p2p badcount [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p badcount

# Result
2
```

---

#### badlist

```
获取黑名单节点列表

Usage:
  ft p2p badlist [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p badlist

# Result
"fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"
"fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"
```

---

#### count

```
获取连接的节点数量

Usage:
  ft p2p count [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p count

# Result
3
```

---

#### list

```
获取连接的节点列表

Usage:
  ft p2p list [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p list

# Result
"fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"
"fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"
```

---

#### remove

```
断开节点

Usage:
  ft p2p remove <url> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p remove "fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"

# Result
true
```

---

#### removebad

```
把节点从黑名单中移出

Usage:
  ft p2p removebad <url> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p removebad "fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"

# Result
true
```

---

#### removetrusted

```
把节点从信任节点集中移出，但不自动断开连接

Usage:
  ft p2p removetrusted <url> [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p removetrusted "fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"

# Result
true
```

---

#### selfnode

```
获取节点地址

Usage:
  ft p2p selfnode [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft p2p selfnode

# Result
"fnode://4cc261784f24f1f12b7b5251f2626c4e0749676f18ef03080d3fbdc71644bfa04b9868583f44c1c1a495afa211f4c37329a78c4ce6f04ec19ed5a3ff4a52820d@127.0.0.1:2015"
```

---

#### content

```
查询txpool的详细信息

Usage:
  ft txpool content  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft txpool content

# Result
{
    "pending": {
        "0": {
            "blockHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
            "blockNumber": 0,
            "txHash": "0x2e57d7358b41f862f3694bdb6a15b9cd09a6ea63ac620c9587b245419ccc1308",
            "transactionIndex": 0,
            "actions": [
                {
                    "type": 0,
                    "nonce": 97,
                    "from": "normalaccta59",
                    "to": "contracta59",
                    "assetID": 0,
                    "gas": 2000000,
                    "value": 0,
                    "remark": "0x",
                    "payload": "0x05f030e700000000000000000000000000000000000000000000000000000000000010030000000000000000000000000000000000000000000000000000000000000001",
                    "actionHash": "0x261bd3adce4be61e7d61f48cd200acc6ef53cd9e22d3923eb2f7c27d1807ff49",
                    "actionIndex": 0
                }
            ],
            "gasAssetID": 0,
            "gasPrice": 1,
            "gasCost": 2000000
        },
    "queued": {
        "0": {
            "blockHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
            "blockNumber": 0,
            "txHash": "0xb8ea35d8d2e6a342849f3ce74235a71fc0d8628f412c596a45e111ae67b560ec",
            "transactionIndex": 0,
            "actions": [
                {
                    "type": 0,
                    "nonce": 99,
                    "from": "normalaccta59",
                    "to": "contracta59",
                    "assetID": 0,
                    "gas": 2000000,
                    "value": 0,
                    "remark": "0x",
                    "payload": "0x05f030e700000000000000000000000000000000000000000000000000000000000010030000000000000000000000000000000000000000000000000000000000000001",
                    "actionHash": "0x58da7f9b770e2f5a02b376ad5b8f5d5382fbef6697ca83a763442111f5c30aca",
                    "actionIndex": 0
                }
            ],
            "gasAssetID": 0,
            "gasPrice": 1,
            "gasCost": 2000000
        }
    }
}
```

---

#### setgasprice

```
设置gas价格

Usage:
  ft txpool setgasprice <gasprice uint64>  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft txpool setgasprice 1000000000

# Result
true
```

---

#### status

```
查询txpool状态

Usage:
  ft txpool status  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft txpool status
# Result
{
    "pending": 0
    "queued": 0
}
```

#### RPC
```
调用RPC

Usage:
  ft rpc method  [flags]

Flags:
  -h, --help   显示帮助

Global Flags:
  -i, --ipcpath string   IPC 路径 (default "$HOME/.ft_ledger/ft.ipc")
```

Example:

```
# Request
./build/bin/ft rpc ft_getCurrentBlock false
# Result
{
  "extraData": "0x",
  "gasLimit": 30000000,
  "gasUsed": 0,
  "hash": "0xc3f2e94cb8ab8a3be3bc4bc52e864ae4bdd8d2d630e53ba488c2d6702b399b0d",
  "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "miner": "fractalfounder",
  "number": 477,
  "parentHash": "0x0370999a691446cd43459fe00d3648491d337587b5b73fc8ee94a50423aae5c7",
  "proposedIrreversible": 0,
  "receiptsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "size": 534,
  "stateRoot": "0x8d23a46925af49c7f02159c853171df17e53a3482883dfcecf6fe4d3775f614f",
  "timestamp": 1576748153,
  "totalDifficulty": 411999659,
  "transactions": [],
  "transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```
