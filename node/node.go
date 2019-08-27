// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	router "github.com/fractalplatform/fractal/event"
	adaptor "github.com/fractalplatform/fractal/p2p/protoadaptor"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/utils/filelock"
)

// Node is a container on which services can be registered.
type Node struct {
	config          *Config
	running         bool
	instanceDirLock filelock.Releaser        // prevents concurrent use of instance directory
	serviceFuncs    []ServiceConstructor     // Service constructors (in dependency order)
	services        map[reflect.Type]Service // Currently running services
	stop            chan struct{}            // Channel to wait for termination notifications
	lock            sync.RWMutex

	rpcAPIs []rpc.API // List of APIs currently provided by the node

	ipcEndpoint string       // RPC endpoint to listen at (empty = RPC disabled)
	ipcListener net.Listener // RPC RPC listener socket to serve API requests
	ipcHandler  *rpc.Server  // RPC RPC request handler to process the API requests

	httpEndpoint  string       // HTTP endpoint (interface + port) to listen at (empty = HTTP disabled)
	httpWhitelist []string     // HTTP RPC modules to allow through this endpoint
	httpListener  net.Listener // HTTP RPC listener socket to server API requests
	httpHandler   *rpc.Server  // HTTP RPC request handler to process the API requests

	wsEndpoint string       // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	wsListener net.Listener // Websocket RPC listener socket to server API requests
	wsHandler  *rpc.Server  // Websocket RPC request handler to process the API requests

	p2pServer *adaptor.ProtoAdaptor

	log log.Logger
}

// New creates a new P2P node, ready for protocol registration.
func New(conf *Config) (*Node, error) {
	if conf.Logger == nil {
		conf.Logger = log.New()
	}

	return &Node{
		config:       conf,
		running:      false,
		serviceFuncs: []ServiceConstructor{},
		services:     make(map[reflect.Type]Service),
		log:          conf.Logger,
		httpEndpoint: conf.HTTPEndpoint(),
		wsEndpoint:   conf.WSEndpoint(),
		ipcEndpoint:  conf.IPCEndpoint(),
	}, nil
}

// Register injects a new service into the node's stack. The service created by
// the passed constructor must be unique in its type with regard to sibling ones.
func (n *Node) Register(constructor ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.running {
		return ErrNodeRunning
	}
	n.serviceFuncs = append(n.serviceFuncs, constructor)
	return nil
}

// Start create a live node and starts running it.
func (n *Node) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.running {
		return ErrNodeRunning
	}

	if err := n.openDataDir(); err != nil {
		return err
	}

	// Initialize the p2p server. This creates the node key and
	// discovery databases.
	router.Reset()
	n.config.P2PConfig.PrivateKey = n.config.NodeKey()
	n.config.P2PConfig.Logger = n.log
	n.config.P2PConfig.BootstrapNodes = n.config.BootNodes()
	n.config.P2PConfig.StaticNodes = n.config.StaticNodes()
	n.config.P2PConfig.TrustedNodes = n.config.TrustedNodes()
	n.config.P2PConfig.NodeDatabase = n.config.NodeDB()

	n.p2pServer = adaptor.NewProtoAdaptor(n.config.P2PConfig)

	services := make(map[reflect.Type]Service)
	for _, constructor := range n.serviceFuncs {
		// Create a new context for the particular service
		ctx := &ServiceContext{
			config:   n.config,
			services: make(map[reflect.Type]Service),
			P2P:      n.p2pServer,
		}
		for kind, s := range services { // copy needed for threaded access
			ctx.services[kind] = s
		}
		// Construct and save the service
		service, err := constructor(ctx)
		if err != nil {
			return err
		}
		kind := reflect.TypeOf(service)
		if _, exists := services[kind]; exists {
			return fmt.Errorf("duplicate service: %v", kind)
		}
		services[kind] = service
	}

	if err := n.p2pServer.Start(); err != nil {
		n.log.Error("Starting peer-to-peer server", "err", err)
	}

	// Start each of the services
	started := []reflect.Type{}
	for kind, service := range services {
		// Start the next service, stopping all previous upon failure
		if err := service.Start(); err != nil {
			for _, kind := range started {
				services[kind].Stop()
			}
			return err
		}
		// Mark the service started for potential cleanup
		started = append(started, kind)
	}

	// Lastly start the configured RPC interfaces
	if err := n.startRPCServices(services); err != nil {
		for _, service := range services {
			service.Stop()
		}
		return err
	}

	n.services = services
	n.running = true
	n.stop = make(chan struct{})
	return nil
}

// Stop terminates a running node along with all it's services. In the node was
// not started, an error is returned.
func (n *Node) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	// Short circuit if the node's not running
	if !n.running {
		return ErrNodeStopped
	}

	n.stopWS()
	n.stopHTTP()
	n.stopIPC()

	n.rpcAPIs = nil

	failure := &StopError{
		Services: make(map[reflect.Type]error),
	}
	for kind, service := range n.services {
		if err := service.Stop(); err != nil {
			failure.Services[kind] = err
		}
	}

	n.p2pServer.Stop()

	n.p2pServer = nil
	n.services = nil
	n.releaseInstanceDir()
	close(n.stop)
	n.running = false
	if len(failure.Services) > 0 {
		return failure
	}
	return nil
}

// Wait blocks the thread until the node is stopped. If the node is not running
// at the time of invocation, the method immediately returns.
func (n *Node) Wait() {
	n.lock.RLock()
	if !n.running {
		n.lock.RUnlock()
		return
	}
	stop := n.stop
	n.lock.RUnlock()

	<-stop
}

// Restart terminates a running node and boots up a new one in its place. If the
// node isn't running, an error is returned.
func (n *Node) Restart() error {
	if err := n.Stop(); err != nil {
		return err
	}
	return n.Start()
}

// Service retrieves a currently running service registered of a specific type.
func (n *Node) Service(service interface{}) error {
	n.lock.RLock()
	defer n.lock.RUnlock()

	// Short circuit if the node's not running
	if !n.running {
		return ErrNodeStopped
	}
	// Otherwise try to find the service to return
	element := reflect.ValueOf(service).Elem()
	if running, ok := n.services[element.Type()]; ok {
		element.Set(reflect.ValueOf(running))
		return nil
	}
	return ErrServiceUnknown
}

func (n *Node) openDataDir() error {
	if n.config.DataDir == "" {
		return nil
	}

	instdir := filepath.Join(n.config.DataDir, n.config.Name)
	if err := os.MkdirAll(instdir, 0700); err != nil {
		return err
	}
	return n.lockInstanceDir(instdir)
}

// lockInstanceDir Lock the instance directory to prevent concurrent use by another instance as well as
// accidental use of the instance directory as a database.
func (n *Node) lockInstanceDir(path string) error {
	release, _, err := filelock.New(filepath.Join(path, "LOCK"))
	if err != nil {
		return convertFileLockError(err)
	}
	n.instanceDirLock = release
	return nil
}

// releaseInstanceDir Release instance directory lock.
func (n *Node) releaseInstanceDir() {
	if n.instanceDirLock != nil {
		if err := n.instanceDirLock.Release(); err != nil {
			n.log.Error("Can't release datadir lock", "err", err)
		}
		n.instanceDirLock = nil
	}
}

// startRPCServices is a helper method to start all the various RPC endpoint during node startup.
func (n *Node) startRPCServices(services map[reflect.Type]Service) error {
	// Gather all the possible APIs to surface
	apis := n.apis()
	for _, service := range services {
		apis = append(apis, service.APIs()...)
	}

	if err := n.startIPC(apis); err != nil {
		return err
	}
	if err := n.startHTTP(n.httpEndpoint, apis, n.config.HTTPModules, n.config.HTTPCors, n.config.HTTPVirtualHosts); err != nil {
		n.stopIPC()
		return err
	}
	if err := n.startWS(n.wsEndpoint, apis, n.config.WSModules, n.config.WSOrigins, n.config.WSExposeAll); err != nil {
		n.stopHTTP()
		n.stopIPC()
		return err
	}
	// All API endpoints started successfully
	n.rpcAPIs = apis
	return nil
}

// TODO apis returns the collection of RPC descriptors this node offers
func (n *Node) apis() []rpc.API {
	return []rpc.API{}
}

// startIPC initializes and starts the  IPC endpoint.
func (n *Node) startIPC(apis []rpc.API) error {
	if n.ipcEndpoint == "" {
		return nil // IPC disabled.
	}
	listener, handler, err := rpc.StartIPCEndpoint(n.ipcEndpoint, apis)
	if err != nil {
		return err
	}
	n.ipcListener = listener
	n.ipcHandler = handler
	n.log.Info("IPC endpoint opened", "url", n.ipcEndpoint)
	return nil
}

// stopIPC terminates the IPC RPC endpoint.
func (n *Node) stopIPC() {
	if n.ipcListener != nil {
		n.ipcListener.Close()
		n.ipcListener = nil
		n.log.Info("IPC endpoint closed", "endpoint", n.ipcEndpoint)
	}
	if n.ipcHandler != nil {
		n.ipcHandler.Stop()
		n.ipcHandler = nil
	}
}

// startHTTP initializes and starts the HTTP RPC endpoint.
func (n *Node) startHTTP(endpoint string, apis []rpc.API, modules []string, cors []string, vhosts []string) error {
	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := rpc.StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts)
	if err != nil {
		return err
	}
	n.log.Info("HTTP endpoint opened", "url", fmt.Sprintf("http://%s", endpoint), "cors", strings.Join(cors, ","), "vhosts", strings.Join(vhosts, ","))
	// All listeners booted successfully
	n.httpEndpoint = endpoint
	n.httpListener = listener
	n.httpHandler = handler
	return nil
}

// stopHTTP terminates the HTTP RPC endpoint.
func (n *Node) stopHTTP() {
	if n.httpListener != nil {
		n.httpListener.Close()
		n.httpListener = nil
		n.log.Info("HTTP endpoint closed", "url", fmt.Sprintf("http://%s", n.httpEndpoint))
	}
	if n.httpHandler != nil {
		n.httpHandler.Stop()
		n.httpHandler = nil
	}
}

// startWS initializes and starts the websocket RPC endpoint.
func (n *Node) startWS(endpoint string, apis []rpc.API, modules []string, wsOrigins []string, exposeAll bool) error {
	// Short circuit if the WS endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := rpc.StartWSEndpoint(endpoint, apis, modules, wsOrigins, exposeAll)
	if err != nil {
		return err
	}
	n.log.Info("WebSocket endpoint opened", "url", fmt.Sprintf("ws://%s", listener.Addr()))
	// All listeners booted successfully
	n.wsEndpoint = endpoint
	n.wsListener = listener
	n.wsHandler = handler
	return nil
}

// stopWS terminates the websocket RPC endpoint.
func (n *Node) stopWS() {
	if n.wsListener != nil {
		n.wsListener.Close()
		n.wsListener = nil

		n.log.Info("WebSocket endpoint closed", "url", fmt.Sprintf("ws://%s", n.wsEndpoint))
	}
	if n.wsHandler != nil {
		n.wsHandler.Stop()
		n.wsHandler = nil
	}
}

func (n *Node) GetNodeConfig() *ServiceContext {
	ctx := &ServiceContext{
		config:   n.config,
		services: make(map[reflect.Type]Service),
		P2P:      n.p2pServer,
	}
	return ctx
}
