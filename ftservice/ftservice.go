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

package ftservice

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/blockchain"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/consensus/miner"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/ftservice/gasprice"
	"github.com/fractalplatform/fractal/internal/api"
	"github.com/fractalplatform/fractal/node"
	"github.com/fractalplatform/fractal/p2p"
	adaptor "github.com/fractalplatform/fractal/p2p/protoadaptor"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/rpc"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/txpool"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/wallet"
)

// FtService implements the fractal service.
type FtService struct {
	config       *Config
	chainConfig  *params.ChainConfig
	shutdownChan chan bool // Channel for shutting down the service
	blockchain   *blockchain.BlockChain
	txPool       *txpool.TxPool
	chainDb      fdb.Database // Block chain database
	wallet       *wallet.Wallet
	engine       consensus.IEngine
	miner        *miner.Miner
	p2pServer    *adaptor.ProtoAdaptor
	gasPrice     *big.Int
	lock         sync.RWMutex // Protects the variadic fields (e.g. gas price)
	APIBackend   *APIBackend
}

// New creates a new ftservice object (including the initialisation of the common ftservice object)
func New(ctx *node.ServiceContext, config *Config) (*FtService, error) {
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}

	chainCfg, dposCfg, _, err := blockchain.SetupGenesisBlock(chainDb, config.Genesis)
	if err != nil {
		return nil, err
	}

	ctx.AppendBootNodes(chainCfg.BootNodes)

	ftservice := &FtService{
		config:       config,
		chainDb:      chainDb,
		chainConfig:  chainCfg,
		wallet:       ctx.Wallet,
		p2pServer:    ctx.P2P,
		shutdownChan: make(chan bool),
	}

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != blockchain.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d)", bcVersion, blockchain.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, blockchain.BlockChainVersion)
	}

	//blockchain
	ftservice.blockchain, err = blockchain.NewBlockChain(chainDb, vm.Config{}, ftservice.chainConfig, txpool.SenderCacher)
	if err != nil {
		return nil, err
	}
	ftservice.wallet.SetBlockChain(ftservice.blockchain)
	if config.Snapshot {
		go state.SnapShotblk(chainDb, 300, 3600)
	}

	statedb, err := ftservice.blockchain.State()
	if err != nil {
		panic(fmt.Sprintf("state db err %v", err))
	}
	accountManager, err := am.NewAccountManager(statedb)
	if err != nil {
		panic(fmt.Sprintf("genesis accountManager new err: %v", err))
	}
	if ok, err := accountManager.AccountIsExist(chainCfg.SysName); !ok {
		panic(fmt.Sprintf("system account is not exist %v", err))
	}

	assetInfo, err := accountManager.GetAssetInfoByName(chainCfg.SysToken)
	if err != nil {
		panic(fmt.Sprintf("genesis system asset err %v", err))
	}
	chainCfg.SysTokenID = assetInfo.AssetId
	chainCfg.SysTokenDecimals = assetInfo.Decimals

	// txpool
	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}

	ftservice.txPool = txpool.New(*config.TxPool, ftservice.chainConfig, ftservice.blockchain)

	engine := dpos.New(dposCfg, ftservice.blockchain)
	ftservice.engine = engine

	type bc struct {
		*blockchain.BlockChain
		consensus.IEngine
		*txpool.TxPool
		processor.Processor
	}

	bcc := &bc{
		ftservice.blockchain,
		ftservice.engine,
		ftservice.txPool,
		nil,
	}

	validator := processor.NewBlockValidator(bcc, ftservice.engine)
	txProcessor := processor.NewStateProcessor(bcc, ftservice.engine)

	ftservice.blockchain.SetValidator(validator)
	ftservice.blockchain.SetProcessor(txProcessor)

	bcc.Processor = txProcessor
	ftservice.miner = miner.NewMiner(bcc)
	if bts, err := hex.DecodeString(config.Miner.PrivateKey); err == nil {
		if !common.IsValidName(config.Miner.Name) {
			log.Error(fmt.Sprintf("miner name %v invalid", config.Miner.Name))
		} else if priv, err := crypto.ToECDSA(bts); err == nil {
			ftservice.miner.SetCoinbase(config.Miner.Name, priv)
		} else {
			log.Error("miner private error", err)
		}
	} else {
		log.Error("miner private error", err)
	}
	ftservice.miner.SetExtra([]byte(config.Miner.ExtraData))
	if config.Miner.Start {
		ftservice.miner.Start()
	}

	ftservice.APIBackend = &APIBackend{ftservice: ftservice}

	ftservice.SetGasPrice(ftservice.TxPool().GasPrice())

	return ftservice, nil
}

// APIs return the collection of RPC services the ftservice package offers.
func (fs *FtService) APIs() []rpc.API {
	apis := api.GetAPIs(fs.APIBackend)
	return apis
}

// Start implements node.Service, starting all internal goroutines.
func (fs *FtService) Start() error {
	log.Info("start fractal service...")
	return nil
}

// Stop implements node.Service, terminating all internal goroutine
func (fs *FtService) Stop() error {
	fs.blockchain.Stop()
	fs.txPool.Stop()
	fs.chainDb.Close()
	close(fs.shutdownChan)
	log.Info("ftservice stopped")
	return nil
}

func (fs *FtService) GasPrice() *big.Int {
	return fs.txPool.GasPrice()
}

func (fs *FtService) SetGasPrice(gasPrice *big.Int) bool {
	fs.config.GasPrice.Default = new(big.Int).SetBytes(gasPrice.Bytes())
	fs.APIBackend.gpo = gasprice.NewOracle(fs.APIBackend, fs.config.GasPrice)

	fs.txPool.SetGasPrice(new(big.Int).SetBytes(gasPrice.Bytes()))
	return true
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (fdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *FtService) BlockChain() *blockchain.BlockChain { return s.blockchain }
func (s *FtService) TxPool() *txpool.TxPool             { return s.txPool }
func (s *FtService) Engine() consensus.IEngine          { return s.engine }
func (s *FtService) ChainDb() fdb.Database              { return s.chainDb }
func (s *FtService) Wallet() *wallet.Wallet             { return s.wallet }
func (s *FtService) Protocols() []p2p.Protocol          { return nil }
