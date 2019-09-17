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

package blockchain

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/txpool"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/utils/fdb/memdb"
)

var (
	systemPrivateKey, _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
)

type fakeEngine struct {
	*dpos.Dpos
}

func (fe *fakeEngine) VerifySeal(chain consensus.IChainReader, header *types.Header) error {
	log.Debug("blockchain uint test use fake engine VerifySeal function", "number", header.Number.Uint64())
	return nil
}

func newCanonical(t *testing.T, genesis *Genesis) *BlockChain {
	// Initialize a fresh chain with only a genesis block
	chainDb := rawdb.NewMemoryDatabase()

	chainCfg, dposCfg, _, err := SetupGenesisBlock(chainDb, genesis)
	if err != nil {
		t.Fatal(err)
	}

	blockchain, err := NewBlockChain(chainDb, false, vm.Config{}, chainCfg, nil, 0, txpool.SenderCacher)
	if err != nil {
		t.Fatal(err)
	}

	stateDB, err := blockchain.State()
	if err != nil {
		t.Fatalf("state db err %v", err)
	}
	accountManager, err := am.NewAccountManager(stateDB)
	if err != nil {
		t.Fatalf("genesis accountManager new err: %v", err)
	}
	if ok, err := accountManager.AccountIsExist(common.StrToName(chainCfg.SysName)); !ok {
		t.Fatalf("system account is not exist %v", err)
	}

	assetInfo, err := accountManager.GetAssetInfoByName(chainCfg.SysToken)
	if err != nil {
		t.Fatalf("genesis system asset err %v", err)
	}

	chainCfg.SysTokenID = assetInfo.AssetID
	chainCfg.SysTokenDecimals = assetInfo.Decimals

	engine := dpos.New(dposCfg, blockchain)

	bc := struct {
		*BlockChain
		consensus.IEngine
	}{blockchain, engine}

	validator := processor.NewBlockValidator(&bc, &fakeEngine{engine})
	txProcessor := processor.NewStateProcessor(&bc, engine)
	blockchain.SetValidator(validator)
	blockchain.SetProcessor(txProcessor)
	return blockchain
}

func makeNewChain(t *testing.T, genesis *Genesis, chain *BlockChain, n, seed int) (*BlockChain, []*types.Block) {
	tmpDB, err := deepCopyDB(chain.db)
	if err != nil {
		t.Fatal(err)
	}

	engine := dpos.New(dposConfig(genesis.Config), chain)

	newblocks, _ := generateChain(genesis.Config,
		chain.CurrentBlock(), engine, chain, tmpDB,
		n, seed, func(i int, b *blockGenerator) {

			name := common.StrToName(genesis.Config.SysName)
			b.SetCoinbase(name)

			engine.SetSignFn(func(content []byte, state *state.StateDB) ([]byte, error) {
				return crypto.Sign(content, systemPrivateKey)
			})
		})

	_, err = chain.InsertChain(newblocks)
	if err != nil {
		t.Fatal("makeNewChain func insert chain err", err)
	}
	return chain, newblocks
}

func deepCopyDB(db fdb.Database) (fdb.Database, error) {
	mdb, ok := db.(*memdb.MemDatabase)
	if !ok {
		return nil, errors.New("db must fdb.MemDatabase")
	}
	return mdb.Copy(), nil
}

func generateChain(config *params.ChainConfig, parent *types.Block, engine *dpos.Dpos,
	chain *BlockChain, db fdb.Database, n, seed int, gen func(int, *blockGenerator)) ([]*types.Block, [][]*types.Receipt) {

	if config == nil {
		config = params.DefaultChainconfig
	}

	chain.db = db
	blocks, receipts := make(types.Blocks, n), make([][]*types.Receipt, n)
	genblock := func(i int, parent *types.Block, stateDB *state.StateDB) (*types.Block, []*types.Receipt) {
		b := &blockGenerator{
			i:          i,
			parent:     parent,
			stateDB:    stateDB,
			config:     config,
			engine:     engine,
			BlockChain: chain,
		}

		b.header = makeHeader(b, parent, b.stateDB, b.engine, seed)

		// Execute any user modifications to the block
		if gen != nil {
			gen(i, b)
		}

		if b.engine != nil {
			// Finalize and seal the block
			if err := b.engine.Prepare(b, b.header, b.txs, nil, b.stateDB); err != nil {
				panic(fmt.Sprintf("engine prepare error: %v", err))
			}

			name := common.StrToName(chain.chainConfig.SysName)

			tx := types.NewTransaction(uint64(0), big.NewInt(1), types.NewAction(types.Transfer, name, common.StrToName(chain.chainConfig.AccountName), b.TxNonce(name), uint64(0), 109000, big.NewInt(100), nil, nil))

			keyPair := types.MakeKeyPair(systemPrivateKey, []uint64{0})
			if err := types.SignActionWithMultiKey(tx.GetActions()[0], tx, types.NewSigner(params.DefaultChainconfig.ChainID), 0, []*types.KeyPair{keyPair}); err != nil {
				panic(err)
			}

			b.AddTxWithChain(tx)

			block, err := b.engine.Finalize(b, b.header, b.txs, b.receipts, b.stateDB)
			if err != nil {
				panic(fmt.Sprintf("engine finalize error: %v", err))
			}

			block, err = b.engine.Seal(b, block, nil)
			if err != nil {
				panic(fmt.Sprintf("engine seal error: %v", err))
			}

			block.Head.ReceiptsRoot = types.DeriveReceiptsMerkleRoot(b.receipts)
			block.Head.TxsRoot = types.DeriveTxsMerkleRoot(b.txs)
			block.Head.Bloom = types.CreateBloom(b.receipts)
			batch := db.NewBatch()

			root, err := b.stateDB.Commit(batch, block.Hash(), block.NumberU64())
			if err != nil {
				panic(fmt.Sprintf("state Commit error: %v", err))
			}

			if err := b.stateDB.Database().TrieDB().Commit(root, false); err != nil {
				panic(fmt.Sprintf("trie write error: %v", err))
			}

			if err := batch.Write(); err != nil {
				panic(fmt.Sprintf("batch Write error: %v", err))
			}

			rawdb.WriteHeader(db, block.Head)

			return block, b.receipts
		}
		return nil, nil
	}

	for i := 0; i < n; i++ {
		stateDB, err := chain.StateAt(parent.Root())
		if err != nil {
			panic(err)
		}
		block, receipt := genblock(i, parent, stateDB)
		blocks[i] = block
		receipts[i] = receipt
		parent = block
	}

	return blocks, receipts
}

func makeHeader(chain consensus.IChainReader, parent *types.Block, state *state.StateDB, engine *dpos.Dpos, seed int) *types.Header {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Coinbase:   parent.Coinbase(),
		GasLimit:   params.BlockGasLimit,
		Number:     new(big.Int).Add(parent.Number(), big.NewInt(1)),
		Time:       big.NewInt(0),
		Difficulty: big.NewInt(0),
		Extra:      big.NewInt(int64(seed)).Bytes(),
	}

	header.Time.Add(header.Time, big.NewInt(int64(engine.Config().BlockInterval*uint64(time.Millisecond))))
	header.Time.Add(header.Time, parent.Time())
	header.Time = big.NewInt(int64(engine.Slot(header.Time.Uint64())))

	if header.Time.Cmp(parent.Header().Time) <= 0 {
		panic(fmt.Sprintf("header time %d less than parent header time %v ", header.Time.Uint64(), parent.Time().Uint64()))
	}

	header.Difficulty = engine.CalcDifficulty(chain, header.Time.Uint64(), parent.Header())
	return header
}

func printLog(level log.Lvl) {
	logger := log.NewGlogHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(false)))
	logger.Verbosity(level)
	log.Root().SetHandler(log.Handler(logger))
}

func checkBlocksInsert(t *testing.T, chain *BlockChain, blocks []*types.Block) {
	block := chain.CurrentBlock()
	index := 0
	for block.NumberU64() != 0 {
		index++
		if blocks[len(blocks)-index].Hash() != block.Hash() {
			t.Fatalf("Write/Get HeadBlockHash failed")
		}
		block = chain.GetBlockByNumber(block.NumberU64() - 1)
	}

}

func checkCompleteChain(t *testing.T, chain *BlockChain) {
	block := chain.CurrentBlock()
	for block.NumberU64() != 0 {
		parentBlock := chain.GetBlockByNumber(block.NumberU64() - 1)
		if block.ParentHash() != parentBlock.Hash() {
			t.Fatalf("is not complete chain block num: %v,hash:%v, parent hash: %v \n parent block num: %v, hash: %v,",
				block.NumberU64(), block.Hash().Hex(), block.ParentHash().Hex(),
				parentBlock.NumberU64(), parentBlock.Hash().Hex())
		}
		block = parentBlock
	}
}
