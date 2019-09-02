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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/accountmanager"
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
	memdb "github.com/fractalplatform/fractal/utils/fdb/memdb"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	issurevalue        = "10000000000000000000000000000"
	syscandidatePrefix = "syscandidate"
	systemPrikey, _    = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
)

type candidateInfo struct {
	name   string
	prikey *ecdsa.PrivateKey
}

func getCandidates() map[string]*candidateInfo {
	candidates := make(map[string]*candidateInfo)
	// pri0, _ := crypto.HexToECDSA("189c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	// pri1, _ := crypto.HexToECDSA("9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658")
	// pri2, _ := crypto.HexToECDSA("8605cf6e76c9fc8ac079d0f841bd5e99bd3ad40fdd56af067993ed14fc5bfca8")

	// candidates["syscandidate0"] = &candidateInfo{"syscandidate0", pri0}
	// candidates["syscandidate1"] = &candidateInfo{"syscandidate1", pri1}
	// candidates["syscandidate2"] = &candidateInfo{"syscandidate2", pri2}
	return candidates
}

func getDefaultGenesisAccounts() (gas []*GenesisAccount) {
	for i := 0; i < len(getCandidates()); i++ {
		candidate := getCandidates()[syscandidatePrefix+strconv.Itoa(i)]
		gas = append(gas, &GenesisAccount{
			Name:   candidate.name,
			PubKey: common.BytesToPubKey(crypto.FromECDSAPub(&candidate.prikey.PublicKey)),
		})
	}
	return
}

func calculateEpochInterval(dposCfg *dpos.Config) uint64 {
	return dposCfg.CandidateScheduleSize * dposCfg.BlockFrequency * dposCfg.BlockInterval * uint64(time.Millisecond)
}

func makeSystemCandidatesAndTime(parentTime uint64, genesis *Genesis) ([]string, []uint64) {
	var (
		candidates     []string
		baseCandidates = getCandidates()
	)
	dcfg := dposConfig(genesis.Config)
	for i := uint64(0); i < (dcfg.EpochInterval/dcfg.BlockInterval*dcfg.BlockFrequency)+1; i++ {
		for j := 0; j < len(baseCandidates); j++ {
			for k := 0; k < int(genesis.Config.DposCfg.BlockFrequency); k++ {
				candidates = append(candidates, genesis.Config.SysName)
			}
		}
	}
	candidates = candidates[0:]
	headerTimes := make([]uint64, len(candidates))
	for i := 0; i < len(candidates); i++ {
		headerTimes[i] = genesis.Config.DposCfg.BlockInterval*uint64(time.Millisecond)*uint64(i+1) + parentTime
	}
	return candidates, headerTimes
}

func makeCandidatesAndTime(parentTime uint64, genesis *Genesis, rounds uint64) ([]string, []uint64) {
	var (
		candidates     []string
		baseCandidates = getCandidates()
	)
	for i := 0; uint64(i) < rounds; i++ {
		for j := 0; j < len(baseCandidates); j++ {
			for k := 0; k < int(genesis.Config.DposCfg.BlockFrequency); k++ {
				candidates = append(candidates, baseCandidates[syscandidatePrefix+strconv.Itoa(j)].name)
			}
		}
	}
	headerTimes := make([]uint64, len(candidates))
	for i := 0; i < len(candidates); i++ {
		headerTimes[i] = genesis.Config.DposCfg.BlockInterval*uint64(time.Millisecond)*uint64(i+1) + parentTime
	}
	return candidates, headerTimes
}

func newCanonical(t *testing.T, genesis *Genesis) *BlockChain {
	// Initialize a fresh chain with only a genesis block
	chainDb := memdb.NewMemDatabase()
	chainCfg, dposCfg, _, err := SetupGenesisBlock(chainDb, genesis)
	if err != nil {
		t.Fatal(err)
	}

	blockchain, err := NewBlockChain(chainDb, false, vm.Config{}, chainCfg, nil, 0, txpool.SenderCacher)
	if err != nil {
		t.Fatal(err)
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatalf("state db err %v", err)
	}
	accountManager, err := am.NewAccountManager(statedb)
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

	validator := processor.NewBlockValidator(&bc, engine)
	txProcessor := processor.NewStateProcessor(&bc, engine)
	blockchain.SetValidator(validator)
	blockchain.SetProcessor(txProcessor)

	return blockchain
}

func makeNewChain(t *testing.T, genesis *Genesis, chain *BlockChain, candidates []string, headerTimes []uint64) (*BlockChain, []*types.Block) {

	tmpdb, err := deepCopyDB(chain.db)
	if err != nil {
		t.Fatal(err)
	}

	engine := dpos.New(dposConfig(genesis.Config), chain)

	newblocks, _ := generateChain(genesis.Config, chain.CurrentBlock(), engine, chain, tmpdb,
		len(headerTimes), func(i int, b *BlockGenerator) {

			baseCandidates := getCandidates()

			baseCandidates[genesis.Config.SysName] = &candidateInfo{genesis.Config.SysName, systemPrikey}

			minerInfo := baseCandidates[candidates[i]]

			b.SetCoinbase(common.StrToName(minerInfo.name))
			engine.SetSignFn(func(content []byte, state *state.StateDB) ([]byte, error) {
				return crypto.Sign(content, minerInfo.prikey)
			})
			b.OffsetTime(int64(engine.Slot(headerTimes[i])))

			if i == 0 {
				txs := makeCandidatesTx(t, genesis.Config.SysName, systemPrikey, b.statedb)
				for _, tx := range txs {
					b.AddTx(tx)
				}
			}

		})

	_, err = chain.InsertChain(newblocks)
	if err != nil {
		t.Fatal("insert chain err", err)
	}

	return chain, newblocks
}

func generateForkBlocks(t *testing.T, genesis *Genesis, candidates []string, headerTimes []uint64) []*types.Block {
	genesis.AllocAccounts = append(genesis.AllocAccounts, getDefaultGenesisAccounts()...)
	chain := newCanonical(t, genesis)
	defer chain.Stop()

	tmpdb, err := deepCopyDB(chain.db)
	if err != nil {
		t.Fatal(err)
	}

	engine := dpos.New(dposConfig(genesis.Config), chain)

	newblocks, _ := generateChain(genesis.Config, chain.CurrentBlock(), engine, chain, tmpdb,
		len(headerTimes), func(i int, b *BlockGenerator) {
			baseCandidates := getCandidates()

			baseCandidates[genesis.Config.SysName] = &candidateInfo{genesis.Config.SysName, systemPrikey}
			minerInfo := baseCandidates[candidates[i]]

			b.SetCoinbase(common.StrToName(minerInfo.name))
			engine.SetSignFn(func(content []byte, state *state.StateDB) ([]byte, error) {
				return crypto.Sign(content, minerInfo.prikey)
			})

			b.OffsetTime(int64(engine.Slot(headerTimes[i])))

			if i == 0 {
				txs := makeCandidatesTx(t, genesis.Config.SysName, systemPrikey, b.statedb)
				for _, tx := range txs {
					b.AddTx(tx)
				}
			}
		})

	return newblocks
}

func makeCandidatesTx(t *testing.T, from string, fromprikey *ecdsa.PrivateKey, state *state.StateDB) []*types.Transaction {
	var txs []*types.Transaction
	signer := types.NewSigner(params.DefaultChainconfig.ChainID)
	am, err := accountmanager.NewAccountManager(state)
	if err != nil {
		t.Fatal("new accountmanager failed")
	}
	nonce, err := am.GetNonce(common.StrToName(from))
	if err != nil {
		t.Fatalf("get name %s nonce failed", from)
	}

	delegateValue := new(big.Int)
	delegateValue.SetString(issurevalue, 10)

	var actions []*types.Action
	for i := 0; i < len(getCandidates()); i++ {
		amount := new(big.Int).Mul(delegateValue, big.NewInt(2))
		action := types.NewAction(types.Transfer, common.StrToName(from), common.StrToName(getCandidates()[syscandidatePrefix+strconv.Itoa(i)].name), nonce, uint64(0), uint64(210000), amount, nil, nil)
		actions = append(actions, action)
		nonce++
	}
	tx := types.NewTransaction(uint64(0), big.NewInt(2), actions...)
	keyPair := types.MakeKeyPair(fromprikey, []uint64{0})
	for _, action := range actions {
		err := types.SignActionWithMultiKey(action, tx, signer, 0, []*types.KeyPair{keyPair})
		if err != nil {
			t.Fatalf(fmt.Sprintf("SignAction err %v", err))
		}
	}

	txs = append(txs, tx)

	var actions1 []*types.Action
	for i := 0; i < len(getCandidates()); i++ {
		to := getCandidates()[syscandidatePrefix+strconv.Itoa(i)]
		url := "www." + to.name + ".io"
		arg := &dpos.RegisterCandidate{
			URL: url,
		}
		payload, _ := rlp.EncodeToBytes(arg)
		action := types.NewAction(types.RegCandidate, common.StrToName(to.name), common.StrToName(params.DefaultChainconfig.DposName), 0, uint64(0), uint64(210000), delegateValue, payload, nil)
		actions1 = append(actions1, action)
	}

	tx1 := types.NewTransaction(uint64(0), big.NewInt(2), actions1...)
	for _, action := range actions1 {
		keyPair = types.MakeKeyPair(getCandidates()[action.Sender().String()].prikey, []uint64{0})
		err := types.SignActionWithMultiKey(action, tx1, signer, 0, []*types.KeyPair{keyPair})
		if err != nil {
			t.Fatalf(fmt.Sprintf("SignAction err %v", err))
		}
	}

	txs = append(txs, tx1)

	return txs
}

func deepCopyDB(db fdb.Database) (fdb.Database, error) {
	mdb, ok := db.(*memdb.MemDatabase)
	if !ok {
		return nil, errors.New("db must fdb.MemDatabase")
	}
	return mdb.Copy(), nil
}

func generateChain(config *params.ChainConfig, parent *types.Block, engine consensus.IEngine, chain *BlockChain, db fdb.Database, n int, gen func(int, *BlockGenerator)) ([]*types.Block, [][]*types.Receipt) {
	if config == nil {
		config = params.DefaultChainconfig
	}

	chain.db = db

	blocks, receipts := make(types.Blocks, n), make([][]*types.Receipt, n)
	genblock := func(i int, parent *types.Block, statedb *state.StateDB) (*types.Block, []*types.Receipt) {
		b := &BlockGenerator{i: i, parent: parent, statedb: statedb, config: config, engine: engine, BlockChain: chain}
		b.header = makeHeader(b, parent, statedb, b.engine)

		// Execute any user modifications to the block
		if gen != nil {
			gen(i, b)
		}

		if b.engine != nil {
			// Finalize and seal the block
			if err := b.engine.Prepare(b, b.header, b.txs, nil, b.statedb); err != nil {
				panic(fmt.Sprintf("engine prepare error: %v", err))
			}

			block, err := b.engine.Finalize(b, b.header, b.txs, b.receipts, b.statedb)
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

			root, err := b.statedb.Commit(batch, block.Hash(), block.NumberU64())
			if err != nil {
				panic(fmt.Sprintf("state Commit error: %v", err))
			}

			if err := b.statedb.Database().TrieDB().Commit(root, false); err != nil {
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
		statedb, err := chain.StateAt(parent.Root())
		if err != nil {
			panic(err)
		}
		block, receipt := genblock(i, parent, statedb)
		blocks[i] = block
		receipts[i] = receipt
		parent = block
	}

	return blocks, receipts
}

func makeHeader(chain consensus.IChainReader, parent *types.Block, state *state.StateDB, engine consensus.IEngine) *types.Header {
	return &types.Header{
		ParentHash: parent.Hash(),
		Coinbase:   parent.Coinbase(),
		Difficulty: engine.CalcDifficulty(chain, 0, parent.Header()),
		GasLimit:   params.BlockGasLimit,
		Number:     new(big.Int).Add(parent.Number(), big.NewInt(1)),
		Time:       big.NewInt(0),
	}
}

func printLog(level log.Lvl) {
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(false)))
	glogger.Verbosity(level)
	log.Root().SetHandler(log.Handler(glogger))
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
