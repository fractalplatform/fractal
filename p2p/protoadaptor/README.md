### P2P功能概述
P2P模块用来实现P2P网络通信，包含了节点连接、加密握手、协议适配等功能。  
P2P协议适配主要对接Router模块，负责将收到的网络包转化成`Router.Event`，并投递到Router。也负责将Router递送来的`Event`转化成网络包发送到P2P网络。

### P2P使用
```go
cfg := p2p.Config{
	MaxPeers:    10,         // 最大连接数
	PrivateKey:  crypto.GenerateKey(), // ECC私钥
	Name:        "Fractal-P2P",      // 可选，节点名
	ListenAddr:  "127.0.0.1:12345", // 监听端口
	StaticNodes: []*enode.Node{},   // 可选，静态节点
}
srv := adaptor.NewProtoAdaptor(cfg)
srv.Start()
```

### 示例
```go
package main

import (
	"fmt"
	"os"

	"github.com/fractalplatform/fractal/crypto"
	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/p2p"
	"github.com/fractalplatform/fractal/p2p/enode"
	adaptor "github.com/fractalplatform/fractal/p2p/protoadaptor"
	//"github.com/fractalplatform/fractal/p2p/nat"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println(os.Args[0], " keyfile port [enode]")
		return
	}
	StationLoop(os.Args[1])
	if len(os.Args) == 3 {
		p2pTest(os.Args[1], os.Args[2], "")
	}
	if len(os.Args) == 4 {
		p2pTest(os.Args[1], os.Args[2], os.Args[3])
	}
}

func StationLoop(name string) {
	station := router.NewLocalStation(name, nil) // 新建名为name的站点
	ch := make(chan *router.Event)
	router.Subscribe(station, ch, router.RouterTestString, "")     // 订阅typecode为router.RouterTestString，消息类型为string的消息
	router.Subscribe(station, ch, router.RouterTestInt, uint32(0)) // router.RouterTestInt => uint32
	router.Subscribe(station, ch, router.P2pNewPeer, nil)          // 订阅NewPeer消息，无消息体
	router.Subscribe(station, ch, router.P2pDelPeer, nil)          // 订阅DelPeer消息，无消息体

	peerNum := uint32(0)
	broadcastNum := func(num uint32) {
		broadcast := router.GetStationByName("broadcast") // 从router获取broadcast站点
		fmt.Printf("%s >> %d\n", name, num)
		router.SendTo(station, broadcast, router.RouterTestInt, num) // 将消息发送至broadcast站点
	}
	go func() {
		for {
			e := <-ch
			switch e.Typecode {
			case router.P2pNewPeer:
				peerNum++
				fmt.Println("new peer!", peerNum)
				if peerNum > 2 {
					fmt.Println("drop new peer")
					router.SendTo(nil, nil, router.P2pDisconectPeer, e.From) // 断开连接
					continue
				}
				// 新入连接，e.From为RemoteStation
				hello := fmt.Sprintf("Hello, I'm %s. How many peers do you have?", name)
				fmt.Println(name, "> ", hello)
				router.SendTo(station, e.From, router.RouterTestString, hello) // 发送字符串到新入节点
			case router.P2pDelPeer:
				// 连接断开通知，e.From为RemoteStation
				peerNum--
				fmt.Println("deleted peer!", peerNum)
				broadcastNum(peerNum) // 广播
			case router.RouterTestString:
				fmt.Println(name, "< ", e.Data.(string))
				fmt.Printf("%s > peers: %d\n", name, peerNum)
				router.ReplyEvent(e, router.RouterTestInt, peerNum) //回应uint32消息
			case router.RouterTestInt:
				num := e.Data.(uint32)
				fmt.Println(name, "< peers:", num)
			}
		}
	}()
}

func p2pTest(keyfile string, port string, nodeUrl string) {
	nodekey, err := crypto.LoadECDSA(keyfile)
	if err != nil {
		nodekey, _ = crypto.GenerateKey()
		crypto.SaveECDSA(keyfile, nodekey)
	}

	cfg := p2p.Config{
		MaxPeers:    10,
		PrivateKey:  nodekey,
		Name:        "Fractal-P2P",
		ListenAddr:  "127.0.0.1:" + port,
		StaticNodes: []*enode.Node{},
	}

	if node, err := enode.ParseV4(nodeUrl); err == nil {
		cfg.StaticNodes = append(cfg.StaticNodes, node)
	}
	srv := adaptor.NewProtoAdaptor(cfg)
	srv.Start()
	fmt.Println(srv.NodeInfo().Enode)
	for {
	}
	srv.Stop()
}
```
