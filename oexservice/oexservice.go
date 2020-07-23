// Copyright 2018 The OEX Team Authors
// This file is part of the OEX project.
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

package oexservice

import (
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/oexplatform/oexchain/blockchain"
	"github.com/oexplatform/oexchain/consensus"
	"github.com/oexplatform/oexchain/consensus/dpos"
	"github.com/oexplatform/oexchain/consensus/miner"
	"github.com/oexplatform/oexchain/node"
	"github.com/oexplatform/oexchain/oexservice/gasprice"
	"github.com/oexplatform/oexchain/p2p"
	adaptor "github.com/oexplatform/oexchain/p2p/protoadaptor"
	"github.com/oexplatform/oexchain/params"
	"github.com/oexplatform/oexchain/processor"
	"github.com/oexplatform/oexchain/processor/vm"
	"github.com/oexplatform/oexchain/rpc"
	"github.com/oexplatform/oexchain/rpcapi"
	"github.com/oexplatform/oexchain/txpool"
	"github.com/oexplatform/oexchain/utils/fdb"
)

// OEXService implements the oex service.
type OEXService struct {
	config       *Config
	chainConfig  *params.ChainConfig
	shutdownChan chan bool // Channel for shutting down the service
	blockchain   *blockchain.BlockChain
	txPool       *txpool.TxPool
	chainDb      fdb.Database // Block chain database
	engine       consensus.IEngine
	miner        *miner.Miner
	p2pServer    *adaptor.ProtoAdaptor
	APIBackend   *APIBackend
}

// New creates a new oexservice object (including the initialisation of the common oexservice object)
func New(ctx *node.ServiceContext, config *Config) (*OEXService, error) {
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}

	chainCfg, dposCfg, _, err := blockchain.SetupGenesisBlock(chainDb, config.Genesis)
	if err != nil {
		return nil, err
	}

	ctx.AppendBootNodes(chainCfg.BootNodes)

	oexService := &OEXService{
		config:       config,
		chainDb:      chainDb,
		chainConfig:  chainCfg,
		p2pServer:    ctx.P2P,
		shutdownChan: make(chan bool),
	}

	//blockchain
	vmconfig := vm.Config{
		ContractLogFlag: config.ContractLogFlag,
	}

	oexService.blockchain, err = blockchain.NewBlockChain(chainDb, config.StatePruning, vmconfig, oexService.chainConfig, config.BadHashes, config.StartNumber, txpool.SenderCacher)
	if err != nil {
		return nil, err
	}

	// txpool
	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}

	oexService.txPool = txpool.New(*config.TxPool, oexService.chainConfig, oexService.blockchain)

	engine := dpos.New(dposCfg, oexService.blockchain)
	oexService.engine = engine

	type bc struct {
		*blockchain.BlockChain
		consensus.IEngine
		*txpool.TxPool
		processor.Processor
	}

	bcc := &bc{
		oexService.blockchain,
		oexService.engine,
		oexService.txPool,
		nil,
	}

	validator := processor.NewBlockValidator(bcc, oexService.engine)
	txProcessor := processor.NewStateProcessor(bcc, oexService.engine)

	oexService.blockchain.SetValidator(validator)
	oexService.blockchain.SetProcessor(txProcessor)

	bcc.Processor = txProcessor
	oexService.miner = miner.NewMiner(bcc)
	oexService.miner.SetDelayDuration(config.Miner.Delay)
	oexService.miner.SetCoinbase(config.Miner.Name, config.Miner.PrivateKeys)
	oexService.miner.SetExtra([]byte(config.Miner.ExtraData))
	if config.Miner.Start {
		oexService.miner.Start(false)
	}

	oexService.APIBackend = &APIBackend{ftservice: oexService}

	oexService.SetGasPrice(oexService.TxPool().GasPrice())
	return oexService, nil
}

// APIs return the collection of RPC services the oexservice package offers.
func (fs *OEXService) APIs() []rpc.API {
	return rpcapi.GetAPIs(fs.APIBackend)
}

// Start implements node.Service, starting all internal goroutines.
func (fs *OEXService) Start() error {
	log.Info("start oex service...")
	return nil
}

// Stop implements node.Service, terminating all internal goroutine
func (fs *OEXService) Stop() error {
	fs.miner.Stop()
	fs.blockchain.Stop()
	fs.txPool.Stop()
	fs.chainDb.Close()
	close(fs.shutdownChan)
	log.Info("oexservice stopped")
	return nil
}

func (fs *OEXService) GasPrice() *big.Int {
	return fs.txPool.GasPrice()
}

func (fs *OEXService) SetGasPrice(gasPrice *big.Int) bool {
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

func (s *OEXService) BlockChain() *blockchain.BlockChain { return s.blockchain }
func (s *OEXService) TxPool() *txpool.TxPool             { return s.txPool }
func (s *OEXService) Engine() consensus.IEngine          { return s.engine }
func (s *OEXService) ChainDb() fdb.Database              { return s.chainDb }
func (s *OEXService) Protocols() []p2p.Protocol          { return nil }
