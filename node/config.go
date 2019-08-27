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
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/p2p"
	"github.com/fractalplatform/fractal/p2p/enode"
)

const (
	datadirPrivateKey   = "nodekey"      // Path within the datadir to the node's private key
	datadirBootNodes    = "bootnodes"    // Path within the datadir to the boot node list
	datadirStaticNodes  = "staticnodes"  // Path within the datadir to the static node list
	datadirTrustedNodes = "trustednodes" // Path within the datadir to the trusted node list
)

// Config represents a small collection of configuration values to fine tune the
// P2P network layer of a protocol stack.
type Config struct {
	Name    string
	DataDir string `mapstructure:"datadir"`
	IPCPath string `mapstructure:"ipcpath"`

	HTTPHost         string   `mapstructure:"httphost"`
	HTTPPort         int      `mapstructure:"httpport"`
	HTTPModules      []string `mapstructure:"httpmodules"`
	HTTPCors         []string `mapstructure:"httpcors"`
	HTTPVirtualHosts []string `mapstructure:"httpvirtualhosts"`

	WSHost      string   `mapstructure:"wshost"`
	WSPort      int      `mapstructure:"wsport"`
	WSModules   []string `mapstructure:"wsmodules"`
	WSOrigins   []string `mapstructure:"wsorigins"`
	WSExposeAll bool     `mapstructure:"wsexposall"`

	// p2p
	P2PBootNodes    string `mapstructure:"bootnodes"`
	P2PStaticNodes  string `mapstructure:"staticnodes"`
	P2PTrustNodes   string `mapstructure:"trustnodes"`
	P2PNodeDatabase string `mapstructure:"nodedb"`

	P2PConfig *p2p.Config `mapstructure:"p2p"`

	// Logger is a custom logger to use with the p2p.Server.
	Logger log.Logger `toml:",omitempty"`
}

// NewConfig initialize config
func NewConfig(name, datadir string) *Config {
	return &Config{
		Name:    name,
		DataDir: datadir,
		Logger:  log.New(),
	}
}

// HTTPEndpoint resolves an HTTP endpoint based on the configured host interface
// and port parameters.
func (c *Config) HTTPEndpoint() string {
	if c.HTTPHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

// WSEndpoint resolves a websocket endpoint based on the configured host interface
// and port parameters.
func (c *Config) WSEndpoint() string {
	if c.WSHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.WSHost, c.WSPort)
}

// IPCEndpoint resolves an RPC endpoint based on a configured value, taking into
// account the set data folders as well as the designated platform we're currently
// running on.
func (c *Config) IPCEndpoint() string {
	// Short circuit if IPC has not been enabled
	if c.IPCPath == "" {
		return ""
	}
	// On windows we can only use plain top-level pipes
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(c.IPCPath, `\\.\pipe\`) {
			return c.IPCPath
		}
		return `\\.\pipe\` + c.IPCPath
	}
	// Resolve names into the data directory full paths otherwise
	if filepath.Base(c.IPCPath) == c.IPCPath {
		if c.DataDir == "" {
			return filepath.Join(os.TempDir(), c.IPCPath)
		}
		return filepath.Join(c.DataDir, c.IPCPath)
	}
	return c.IPCPath
}

// resolvePath resolves path in the instance directory.
func (c *Config) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(filepath.Join(c.DataDir, c.Name), path)
}

func (c *Config) NodeKey() *ecdsa.PrivateKey {
	// Use any specifically configured key.
	if c.P2PConfig.PrivateKey != nil {
		return c.P2PConfig.PrivateKey
	}

	// Generate ephemeral key if no datadir is being used.
	if c.DataDir == "" {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Crit(fmt.Sprintf("Failed to generate ephemeral node key: %v", err))
		}
		return key
	}

	keyfile := c.resolvePath(datadirPrivateKey)

	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}
	// No persistent key found, generate and store a new one.
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Crit(fmt.Sprintf("Failed to generate node key: %v", err))
	}
	instanceDir := filepath.Join(c.DataDir, c.Name)
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
		return key
	}

	keyfile = filepath.Join(instanceDir, datadirPrivateKey)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
	}

	return key
}

// BootNodes returns a list of node enode URLs configured as boot nodes.
func (c *Config) BootNodes() []*enode.Node {
	if len(c.P2PBootNodes) != 0 {
		return c.readEnodes(c.P2PBootNodes)
	}
	return c.readEnodes(c.resolvePath(datadirBootNodes))
}

// StaticNodes returns a list of node enode URLs configured as static nodes.
func (c *Config) StaticNodes() []*enode.Node {
	if len(c.P2PStaticNodes) != 0 {
		return c.readEnodes(c.P2PStaticNodes)
	}
	return c.readEnodes(c.resolvePath(datadirStaticNodes))
}

// TrustedNodes returns a list of node enode URLs configured as trusted nodes.
func (c *Config) TrustedNodes() []*enode.Node {
	if len(c.P2PTrustNodes) != 0 {
		return c.readEnodes(c.P2PTrustNodes)
	}
	return c.readEnodes(c.resolvePath(datadirTrustedNodes))
}

// NodeDB returns the path of nodedatabase
func (c *Config) NodeDB() string {
	if len(c.P2PNodeDatabase) == 0 {
		return ""
	}
	return filepath.Join(c.DataDir, c.P2PNodeDatabase)
}

func (c *Config) readEnodes(path string) []*enode.Node {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Debug("file is not exist.", "path", path)
		return nil
	}
	// Load the nodes from the config file.W

	fi, err := os.Open(path)
	if err != nil {
		log.Error("enodes config read failed.", "path", path, "err", err)
		return nil
	}
	defer fi.Close()

	var nodes []*enode.Node
	br := bufio.NewReader(fi)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		if node, err := enode.ParseV4(string(line)); err == nil {
			nodes = append(nodes, node)
		} else {
			log.Error("enodes config node parseV4 failed.", "err", err, "line", string(line))
		}
	}
	return nodes
}
