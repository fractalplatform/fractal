### 1：下载 fractal 代码并编译

```
>git clone https://github.com/fractalplatform/fractal.git
>cd fractal
>make all
```

注意：目前请同步 dev 分支代码，后续更新 master 后再同步 master

### 2:修改 genesis.json 文件中的矿工配置

默认出块账号是 fractal.admin,我们需要将其对应的公钥修改为我们自己的公钥（在这之前需要有自己的公私钥对）；

```
>cd build
>vim genesis.json
```

找到文件中的 fractal.admin，将其对应的 pubKey 修改为自己的公钥

### 3：启动自己的链

```
>cd bin
>nohup ./ft --genesis=../genesis.json --datadir=./data/ --miner_start --contractlog --http_modules=fee,miner,dpos,account,txpool,ft &
>tail -f nohup.out   // 查看输出日志
```

启动参数说明：

1. miner_start：启动挖矿功能
2. contractlog：指记录内部交易，可供 RPC 查询
3. http_modules=fee,miner,dpos,account,txpool,ft：指开放 fee,miner,dpos,account,txpool,ft 的 RPC 接口
4. genesis=../genesis.json：指定 genesis 文件
5. datadir=./data/: 指定区块数据存储目录

- 注意：此时链还不能正常工作，查看日志时会发现以下错误

```
ERROR[05-21|13:07:12.003] failed to mint the block
timestamp=1558415232000000000 err="illegal candidate pubkey"
```

这个错误的原因是我们还没配置矿工的私钥，只有配置私钥后矿工才能正确签名出块，所以还需要执行下面的命令才行：

```
>./ft miner -i ./data/ft.ipc setcoinbase "fractal.admin" privateKey.txt
>tail -f nohup.out  // 此时能看到开始正常出块了
```

- privateKey.txt 里存放的是私钥，注意私钥不需要带 0x 前缀，此命令执行完后记得删除 privateKey.txt 这个文件，以防私钥泄露
- 矿工重启后，以上配置私钥的命令需要重新执行才行，并且还需要通过 rpc 接口依次调用 miner_stop 和 miner_force 才能恢复正常出块

### 相关问题：

- 问题 1：私链如何修改一个挖矿周期的时长？
- 答：需要在创世文件中进行配置，公链源码目录 build 文件夹下有 genesis.json 这个文件，里面有配置信息："epchoInterval":10800000，这里单位是毫秒，即 3 小时是一个周期，修改这里的配置重启即可改变挖矿周期，需要注意的是，在启动链的时候必须指定此参数:--genesis=genesis.json。

* 问题 2：我想把原来的链数据清空，然后从 0 开始出块，该怎么做？
* 答：先 kill 原有的私链进程，然后清空数据目录下所有数据（如清空./data 目录下的所有数据），再按照“3：启动自己的私链”中描述的过程执行一遍即可。

* 问题 3：genesis.json 文件中的各个配置项有说明吗？
* 答：请参考文档:https://github.com/fractalplatform/fractal/wiki/%E5%88%9B%E4%B8%96%E6%96%87%E4%BB%B6genesis.json%E8%AF%B4%E6%98%8E
