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
	"testing"
	"time"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/consensus"
	"github.com/fractalplatform/fractal/consensus/dpos"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/processor"
	"github.com/fractalplatform/fractal/processor/vm"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/txpool"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/fdb"
	"github.com/fractalplatform/fractal/utils/rlp"
)

type tdpos struct {
	*dpos.Dpos
}

func (tds *tdpos) VerifySeal(chain consensus.IChainReader, header *types.Header) error {
	return nil
}

func (tds *tdpos) Engine() consensus.IEngine {
	return tds
}

var (
	ds                   = dpos.New(dpos.DefaultConfig, nil)
	tengine              = &tdpos{ds}
	issurevalue          = "500000000000000000000000"
	sysnameprikey, _     = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	producerScheduleSize = dpos.DefaultConfig.ProducerScheduleSize
	blockInterval        = dpos.DefaultConfig.BlockInterval
	blockFrequency       = dpos.DefaultConfig.BlockFrequency
	epochInterval        = producerScheduleSize * blockFrequency * blockInterval * uint64(time.Millisecond)
)

type producerInfo struct {
	name   string
	prikey *ecdsa.PrivateKey
}

var producers []*producerInfo

func init() {
	producers = make([]*producerInfo, 3)
	pri1, _ := crypto.HexToECDSA("189c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	pri2, _ := crypto.HexToECDSA("9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658")
	pri3, _ := crypto.HexToECDSA("8605cf6e76c9fc8ac079d0f841bd5e99bd3ad40fdd56af067993ed14fc5bfca8")
	producers[0] = &producerInfo{"sysproducer1", pri1}
	producers[1] = &producerInfo{"sysproducer2", pri2}
	producers[2] = &producerInfo{"sysproducer3", pri3}

}

func makeProduceAndTime(st uint64, rounds int) ([]string, []uint64) {
	baseproducers := []string{"sysproducer1", "sysproducer2", "sysproducer3"}
	firsttime := blockInterval*1000*uint64(time.Microsecond) + st
	offset := firsttime % epochInterval
	offset = offset / (blockInterval * uint64(time.Millisecond))
	var pros []string
	for i := 0; i < rounds; i++ {
		for j := 0; j < len(baseproducers); j++ {
			for k := 0; k < int(blockFrequency); k++ {
				pros = append(pros, baseproducers[j])
			}
		}

	}
	pros = pros[offset:]
	ht := make([]uint64, len(pros))
	for i := 0; i < len(pros); i++ {
		ht[i] = blockInterval*1000*uint64(time.Microsecond)*uint64(i+1) + st
	}
	return pros, ht
}

func newCanonical(t *testing.T, engine consensus.IEngine) (*Genesis, fdb.Database, *BlockChain, uint64, error) {

	var (
		db         = fdb.NewMemDatabase()
		gspec      = DefaultGenesis()
		genesis, _ = gspec.Commit(db)
	)

	// Initialize a fresh chain with only a genesis block
	blockchain, _ := NewBlockChain(db, vm.Config{}, params.DefaultChainconfig, txpool.SenderCacher)

	type bc struct {
		*BlockChain
		consensus.IEngine
	}

	// Create validator and txProcessor
	cfg := blockchain.Config()
	cfg.SysTokenID = 1
	validator := processor.NewBlockValidator(&bc{blockchain, engine}, engine)
	txProcessor := processor.NewStateProcessor(&bc{blockchain, engine}, engine)

	blockchain.SetValidator(validator)
	blockchain.SetProcessor(txProcessor)

	tmpdb, err := deepCopyDB(db)
	if err != nil {
		return nil, nil, nil, uint64(0), fmt.Errorf("failed to deep copy db: %v", err)
	}
	starttime := uint64(0)
	blocks, _ := generateChain(gspec.Config, genesis, tengine, blockchain, tmpdb, 1, func(i int, block *BlockGenerator) {
		genesisname := genesis.Coinbase()
		block.SetCoinbase(genesisname)
		tengine.SetSignFn(func(content []byte, state *state.StateDB) ([]byte, error) {
			return crypto.Sign(content, sysnameprikey)
		})
		st := blockInterval*uint64(time.Millisecond) + block.parent.Head.Time.Uint64()
		starttime = st
		block.OffsetTime(int64(tengine.Slot(st)))
		state, err := state.New(genesis.Root(), state.NewDatabase(tmpdb))
		if err != nil {
			t.Error("new state failed")
		}
		txs := makeProducersTx(t, params.DefaultChainconfig.SysName.String(), sysnameprikey, producers, state)
		for _, tx := range txs {
			block.AddTx(tx)
		}
	})
	_, err = blockchain.InsertChain(blocks)
	if err != nil {
		t.Error("insert chain err", err)
	}
	cnt := int(blockFrequency*producerScheduleSize*3) - 2
	for index := 0; index < cnt; index++ {
		blocks1, _ := generateChain(gspec.Config, blockchain.CurrentBlock(), tengine, blockchain, tmpdb, 1, func(i int, block *BlockGenerator) {
			genesisname := genesis.Coinbase()
			block.SetCoinbase(genesisname)
			tengine.SetSignFn(func(content []byte, state *state.StateDB) ([]byte, error) {
				return crypto.Sign(content, sysnameprikey)
			})
			st := blockInterval*uint64(time.Millisecond) + starttime
			starttime = st
			block.OffsetTime(int64(tengine.Slot(st)))
		})
		_, err = blockchain.InsertChain(blocks1)
		if err != nil {
			t.Error("insert chain err", err)
		}
	}
	return gspec, db, blockchain, starttime, err
}

func makeNewChain(t *testing.T, gspec *Genesis, chain *BlockChain, db *fdb.Database, h int, headertime []uint64, miners []string, f MakeTransferTx) (*fdb.Database, *BlockChain, []*types.Block, error) {
	var newgenerateblocks []*types.Block
	tmpdb, err := deepCopyDB(*db)
	if err != nil {
		t.Error("copy db err", err)
		return nil, nil, nil, err
	}
	for j := 0; j < h; j++ {
		newblocks, _ := generateChain(gspec.Config, chain.CurrentBlock(), tengine, chain, tmpdb, 1, func(i int, b *BlockGenerator) {
			var minerInfo *producerInfo
			for k := 0; k < len(producers); k++ {
				if producers[k].name == miners[j] {
					minerInfo = producers[k]
				}
			}
			b.SetCoinbase(common.StrToName(minerInfo.name))
			tengine.SetSignFn(func(content []byte, state *state.StateDB) ([]byte, error) {
				return crypto.Sign(content, minerInfo.prikey)
			})
			b.OffsetTime(int64(tengine.Slot(headertime[j])))
			state, err := state.New(b.parent.Root(), state.NewDatabase(tmpdb))
			if err != nil {
				t.Error("new state failed", err)
			}
			if f != nil {
				tx := f(t, params.DefaultChainconfig.SysName.String(), minerInfo.name, sysnameprikey, state)
				b.AddTx(tx)
			}
		})
		_, err := chain.InsertChain(newblocks)
		if err != nil {
			t.Error("insert chain err", err)
		}
		newgenerateblocks = append(newgenerateblocks, newblocks...)
	}

	return db, chain, newgenerateblocks, err
}
func makeProducersTx(t *testing.T, from string, fromprikey *ecdsa.PrivateKey, newaccount []*producerInfo, state *state.StateDB) []*types.Transaction {
	var txs []*types.Transaction
	signer := types.NewSigner(params.DefaultChainconfig.ChainID)
	am, err := accountmanager.NewAccountManager(state)
	if err != nil {
		t.Error("new accountmanager failed")
	}
	nonce, err := am.GetNonce(common.StrToName(from))
	if err != nil {
		t.Errorf("get name %s nonce failed", from)
	}
	delegateValue := new(big.Int)
	delegateValue.SetString(issurevalue, 0)
	var actions []*types.Action
	for _, to := range newaccount {
		amount := new(big.Int).Mul(delegateValue, big.NewInt(2))
		pub := common.BytesToPubKey(crypto.FromECDSAPub(&to.prikey.PublicKey))
		action := types.NewAction(types.CreateAccount, common.StrToName(from), common.StrToName(to.name), nonce, uint64(1), uint64(210000), amount, pub[:])
		actions = append(actions, action)
		nonce++
	}
	tx := types.NewTransaction(uint64(1), big.NewInt(2), actions...)
	for _, action := range actions {
		err := types.SignAction(action, tx, signer, fromprikey)
		if err != nil {
			t.Errorf(fmt.Sprintf("SignAction err %v", err))
		}
	}
	txs = append(txs, tx)
	var actions1 []*types.Action
	for _, to := range newaccount {
		value := big.NewInt(1e5)
		url := "www." + to.name + ".io"
		arg := &dpos.RegisterProducer{
			Url:   url,
			Stake: delegateValue,
		}
		payload, _ := rlp.EncodeToBytes(arg)
		action := types.NewAction(types.RegProducer, common.StrToName(to.name), common.StrToName(to.name), 0, uint64(1), uint64(210000), value, payload)
		actions1 = append(actions1, action)
	}
	tx1 := types.NewTransaction(uint64(1), big.NewInt(2), actions1...)
	for i, action := range actions1 {
		err := types.SignAction(action, tx1, signer, newaccount[i].prikey)
		if err != nil {
			t.Errorf(fmt.Sprintf("SignAction err %v", err))
		}
	}
	txs = append(txs, tx1)
	return txs
}

type MakeTransferTx func(t *testing.T, from, to string, fromprikey *ecdsa.PrivateKey, state *state.StateDB) *types.Transaction

func makeTransferTx(t *testing.T, from, to string, fromprikey *ecdsa.PrivateKey, state *state.StateDB) *types.Transaction {
	signer := types.NewSigner(params.DefaultChainconfig.ChainID)
	am, err := accountmanager.NewAccountManager(state)
	if err != nil {
		t.Error("new accountmanager failed")
	}
	nonce, err := am.GetNonce(common.StrToName(from))
	if err != nil {
		t.Errorf("get name %s nonce failed", from)
	}
	action := types.NewAction(types.Transfer, common.StrToName(from), common.StrToName(to), nonce, uint64(1), uint64(210000), big.NewInt(1), nil)
	tx := types.NewTransaction(uint64(1), big.NewInt(2), action)
	err = types.SignAction(action, tx, signer, fromprikey)
	if err != nil {
		t.Errorf(fmt.Sprintf("SignAction err %v", err))
	}
	return tx
}

func deepCopyDB(db fdb.Database) (fdb.Database, error) {
	memdb, ok := db.(*fdb.MemDatabase)
	if !ok {
		return nil, errors.New("db must fdb.MemDatabase")
	}
	return memdb.Copy(), nil
}

func generateChain(config *params.ChainConfig, parent *types.Block, engine consensus.IEngine, chain *BlockChain, db fdb.Database, n int, gen func(int, *BlockGenerator)) ([]*types.Block, [][]*types.Receipt) {
	if config == nil {
		config = params.DefaultChainconfig
	}
	blocks, receipts := make(types.Blocks, n), make([][]*types.Receipt, n)
	genblock := func(i int, parent *types.Block, statedb *state.StateDB) (*types.Block, []*types.Receipt) {
		b := &BlockGenerator{i: i, parent: parent, statedb: statedb, config: config, engine: engine, bc: chain}
		b.header = makeHeader(b.bc, parent, statedb, b.engine)

		// Execute any user modifications to the block
		if gen != nil {
			gen(i, b)
		}

		if b.engine != nil {
			// Finalize and seal the block
			b.engine.Prepare(b.bc, b.header, b.txs, nil, nil)
			block, _ := b.engine.Finalize(b.bc, b.header, b.txs, b.receipts, statedb)
			block, err := b.engine.Seal(b.bc, block, nil)

			block.Head.ReceiptsRoot = types.DeriveReceiptsMerkleRoot(b.receipts)
			block.Head.TxsRoot = types.DeriveTxsMerkleRoot(b.txs)
			block.Head.Bloom = types.CreateBloom(b.receipts)

			batch := db.NewBatch()

			root, err := statedb.Commit(batch, block.Hash(), block.NumberU64())
			if err != nil {
				panic(fmt.Sprintf("state Commit error: %v", err))
			}
			triedb := statedb.Database().TrieDB()
			triedb.Commit(root, false)
			if batch.Write() != nil {
				panic(fmt.Sprintf("batch Write error: %v", err))

			}
			return block, b.receipts
		}
		return nil, nil
	}

	for i := 0; i < n; i++ {
		statedb, err := state.New(parent.Root(), state.NewDatabase(db))
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
		GasLimit:   params.CalcGasLimit(parent),
		Number:     new(big.Int).Add(parent.Number(), big.NewInt(1)),
		Time:       big.NewInt(0),
	}
}
