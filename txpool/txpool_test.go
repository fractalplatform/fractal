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

package txpool

import (
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	am "github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/rawdb"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/stretchr/testify/assert"
)

func TestConfigCheck(t *testing.T) {
	cfg := new(Config)
	cfg.Journal = DefaultTxPoolConfig.Journal
	assert.Equal(t, cfg.check(), *DefaultTxPoolConfig)
}

// func TestAddPayerTx(t *testing.T) {
// 	var (
// 		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
// 		manager, _ = am.NewAccountManager(statedb)
// 		fname      = common.Name("fromname")
// 		tname      = common.Name("totestname")
// 		fkey       = generateAccount(t, fname, manager)
// 		tkey       = generateAccount(t, tname, manager)
// 		asset      = asset.NewAsset(statedb)
// 	)

// 	// issue asset
// 	if _, err := asset.IssueAsset("ft", 0, 0, "zz", new(big.Int).SetUint64(params.Fractal), 10, common.Name(""), fname, new(big.Int).SetUint64(params.Fractal), common.Name(""), ""); err != nil {
// 		t.Fatal(err)
// 	}

// 	// add balance
// 	if err := manager.AddAccountBalanceByName(fname, "ft", new(big.Int).SetUint64(params.Fractal)); err != nil {
// 		t.Fatal(err)
// 	}

// 	if err := manager.AddAccountBalanceByName(tname, "ft", new(big.Int).SetUint64(params.Fractal)); err != nil {
// 		t.Fatal(err)
// 	}

// 	blockchain := &testBlockChain{statedb, 1000000000, new(event.Feed)}
// 	tx0 := pricedTransaction(0, fname, tname, 109000, big.NewInt(0), fkey)
// 	tx1 := extendTransaction(0, tname, fname, tname, 109000, fkey, tkey)

// 	params.DefaultChainconfig.SysTokenID = 0

// 	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
// 	defer pool.Stop()

// 	nonce, err := pool.State().GetNonce(fname)
// 	if err != nil {
// 		t.Fatal("Invalid getNonce ", err)
// 	}
// 	if nonce != 0 {
// 		t.Fatalf("Invalid nonce, want 0, got %d", nonce)
// 	}

// 	errs := pool.addRemotesSync([]*types.Transaction{tx0, tx1})

// 	t.Log(errs)
// 	nonce, err = pool.State().GetNonce(fname)
// 	if err != nil {
// 		t.Fatal("Invalid getNonce ", err)
// 	}

// 	if nonce != 1 {
// 		t.Fatalf("Invalid nonce, want 1, got %d", nonce)
// 	}

// 	result := pool.Get(tx1.Hash())

// 	if !result.PayerExist() {
// 		t.Fatal("add payer tx failed")
// 	}
// }

// This test simulates a scenario where a new block is imported during a
// state reset and tests whether the pending state is in sync with the
// block head event that initiated the resetState().
func TestStateChangeDuringTransactionPoolReset(t *testing.T) {
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
		manager, _ = am.NewAccountManager(statedb)
		fname      = common.Name("fromname")
		tname      = common.Name("totestname")
		fkey       = generateAccount(t, fname, manager)
		_          = generateAccount(t, tname, manager)
		asset      = asset.NewAsset(statedb)
	)

	// issue asset
	if _, err := asset.IssueAsset("ft", 0, 0, "zz", new(big.Int).SetUint64(params.Fractal), 10, common.Name(""), fname, new(big.Int).SetUint64(params.Fractal), common.Name(""), ""); err != nil {
		t.Fatal(err)
	}

	// add balance
	if err := manager.AddAccountBalanceByName(fname, "ft", new(big.Int).SetUint64(params.Fractal)); err != nil {
		t.Fatal(err)
	}
	blockchain := &testBlockChain{statedb, 1000000000, new(event.Feed)}

	tx0 := transaction(0, fname, tname, 109000, fkey)
	tx1 := transaction(1, fname, tname, 109000, fkey)
	params.DefaultChainconfig.SysTokenID = 0
	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	nonce, err := pool.State().GetNonce(fname)
	if err != nil {
		t.Fatal("Invalid getNonce ", err)
	}
	if nonce != 0 {
		t.Fatalf("Invalid nonce, want 0, got %d", nonce)
	}

	pool.addRemotesSync([]*types.Transaction{tx0, tx1})

	nonce, err = pool.State().GetNonce(fname)
	if err != nil {
		t.Fatal("Invalid getNonce ", err)
	}
	if nonce != 2 {
		t.Fatalf("Invalid nonce, want 2, got %d", nonce)
	}

	<-pool.requestReset(nil, nil)

	_, err = pool.Pending()
	if err != nil {
		t.Fatalf("Could not fetch pending transactions: %v", err)
	}
	nonce, err = pool.State().GetNonce(fname)
	if err != nil {
		t.Fatal("Invalid getNonce ", err)
	}
	if nonce != 2 {
		t.Fatalf("Invalid nonce, want 2, got %d", nonce)
	}
}

func TestInvalidTransactions(t *testing.T) {

	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	tx := transaction(0, fname, tname, 100, fkey)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(1))
	if err := pool.addRemoteSync(tx); err != ErrInsufficientFundsForGas {
		t.Fatal("expected: ", ErrInsufficientFundsForGas, "actual: ", err)
	}

	value := new(big.Int).Add(tx.Cost(), tx.GetActions()[0].Value())

	if err := pool.curAccountManager.AddAccountBalanceByID(fname, assetID, value); err != nil {
		t.Fatal(err)
	}

	if err := pool.addRemoteSync(tx); err != ErrIntrinsicGas {
		t.Fatal("expected", ErrIntrinsicGas, "actual: ", err)
	}

	pool.curAccountManager.SetNonce(fname, 1)
	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(0xffffffffffffff))
	tx = transaction(0, fname, tname, 109000, fkey)
	if err := pool.addRemoteSync(tx); err != ErrNonceTooLow {
		t.Fatal("expected", ErrNonceTooLow, "actual: ", err)
	}

	tx = transaction(1, fname, tname, 109000, fkey)
	pool.gasPrice = big.NewInt(1000)
	if err := pool.addRemoteSync(tx); err != ErrUnderpriced {
		t.Fatal("expected", ErrUnderpriced, "actual: ", err)
	}

	if err := pool.AddLocal(tx); err != nil {
		t.Fatal("expected", nil, "actual: ", err)
	}
}

func TestTransactionQueue(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager)
	generateAccount(t, tname, manager)

	tx := transaction(0, fname, tname, 100, fkey)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(1000))

	<-pool.requestReset(nil, nil)

	pool.enqueueTx(tx.Hash(), tx)
	<-pool.requestPromoteExecutables(newAccountSet(pool.signer, fname))

	if len(pool.pending) != 1 {
		t.Fatal("expected valid txs to be 1 is", len(pool.pending))
	}

	tx = transaction(1, fname, tname, 100, fkey)

	pool.curAccountManager.SetNonce(fname, 2)
	pool.enqueueTx(tx.Hash(), tx)
	<-pool.requestPromoteExecutables(newAccountSet(pool.signer, fname))

	if _, ok := pool.pending[fname].txs.items[tx.GetActions()[0].Nonce()]; ok {
		t.Fatal("expected transaction to be in tx pool")
	}

	if len(pool.queue) > 0 {
		t.Fatal("expected transaction queue to be empty. is", len(pool.queue))
	}

	pool, manager = setupTxPool(fname)
	defer pool.Stop()
	fkey = generateAccount(t, fname, manager)
	tkey := generateAccount(t, tname, manager)

	tx1 := transaction(0, fname, tname, 100, fkey)
	tx2 := transaction(10, fname, tname, 100, fkey)
	tx3 := transaction(11, fname, tname, 100, fkey)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(1000))
	<-pool.requestReset(nil, nil)

	pool.enqueueTx(tx1.Hash(), tx1)
	pool.enqueueTx(tx2.Hash(), tx2)
	pool.enqueueTx(tx3.Hash(), tx3)

	<-pool.requestPromoteExecutables(newAccountSet(pool.signer, fname))

	if len(pool.pending) != 1 {
		t.Fatal("expected tx pool to be 1, got", len(pool.pending))
	}
	if pool.queue[fname].Len() != 2 {
		t.Fatal("expected len(queue) == 2, got", pool.queue[fname].Len())
	}

	// test change permissions

	// add account author tpubkey
	tpubkey := common.BytesToPubKey(crypto.FromECDSAPub(&tkey.PublicKey))
	auther := common.NewAuthor(tpubkey, 1)
	authorAction := &am.AuthorAction{ActionType: am.AddAuthor, Author: auther}
	acctAuth := &am.AccountAuthorAction{AuthorActions: []*am.AuthorAction{authorAction}}
	if err := pool.curAccountManager.UpdateAccountAuthor(fname, acctAuth); err != nil {
		t.Fatal(err)
	}

	// delete account author fpubkey
	fpubkey := common.BytesToPubKey(crypto.FromECDSAPub(&fkey.PublicKey))
	auther = common.NewAuthor(fpubkey, 1)
	authorAction = &am.AuthorAction{ActionType: am.DeleteAuthor, Author: auther}
	acctAuth = &am.AccountAuthorAction{AuthorActions: []*am.AuthorAction{authorAction}}
	if err := pool.curAccountManager.UpdateAccountAuthor(fname, acctAuth); err != nil {
		t.Fatal(err)
	}

	<-pool.requestPromoteExecutables(newAccountSet(pool.signer, fname))
	if len(pool.queue) != 0 {
		t.Fatal("expected len(queue) == 0, got", pool.queue[fname].Len())
	}

	pool.demoteUnexecutables()

	if len(pool.pending) != 0 {
		t.Fatal("expected tx pool to be 0, got", len(pool.pending))
	}
}

func TestTransactionChainFork(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)

	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager)
	tkey := generateAccount(t, tname, manager)

	resetAsset := func() {
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
		newmanager, _ := am.NewAccountManager(statedb)

		if err := newmanager.CreateAccount(common.Name("fractal"), fname, common.Name(""), 0, 0, common.BytesToPubKey(crypto.FromECDSAPub(&fkey.PublicKey)), ""); err != nil {
			t.Fatal(err)
		}
		if err := newmanager.CreateAccount(common.Name("fractal"), tname, common.Name(""), 0, 0, common.BytesToPubKey(crypto.FromECDSAPub(&tkey.PublicKey)), ""); err != nil {
			t.Fatal(err)
		}
		asset := asset.NewAsset(statedb)

		asset.IssueAsset("ft", 0, 0, "zz", new(big.Int).SetUint64(params.Fractal), 10, fname, fname, big.NewInt(1000000), common.Name(""), "")
		newmanager.AddAccountBalanceByID(fname, assetID, big.NewInt(100000000000000))

		pool.chain = &testBlockChain{statedb, 1000000, new(event.Feed)}
		<-pool.requestReset(nil, nil)

	}

	resetAsset()
	tx := transaction(0, fname, tname, 109000, fkey)
	if _, err := pool.add(tx, false); err != nil {
		t.Fatal("didn't expect error", err)
	}
	pool.removeTx(tx.Hash(), true)

	// reset the pool's internal state
	resetAsset()
	if _, err := pool.add(tx, false); err != nil {
		t.Fatal("didn't expect error", err)
	}
}

func TestTransactionDoubleNonce(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager)
	tkey := generateAccount(t, tname, manager)

	resetAsset := func() {
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
		newmanager, _ := am.NewAccountManager(statedb)

		if err := newmanager.CreateAccount(common.Name("fractal"), fname, common.Name(""), 0, 0, common.BytesToPubKey(crypto.FromECDSAPub(&fkey.PublicKey)), ""); err != nil {
			t.Fatal(err)
		}
		if err := newmanager.CreateAccount(common.Name("fractal"), tname, common.Name(""), 0, 0, common.BytesToPubKey(crypto.FromECDSAPub(&tkey.PublicKey)), ""); err != nil {
			t.Fatal(err)
		}
		asset := asset.NewAsset(statedb)

		asset.IssueAsset("ft", 0, 0, "zz", new(big.Int).SetUint64(params.Fractal), 10, fname, fname, big.NewInt(1000000), common.Name(""), "")
		newmanager.AddAccountBalanceByID(fname, assetID, big.NewInt(100000000000000))

		pool.chain = &testBlockChain{statedb, 1000000, new(event.Feed)}
		<-pool.requestReset(nil, nil)

	}
	resetAsset()

	keyPair := types.MakeKeyPair(fkey, []uint64{0})
	tx1 := newTx(big.NewInt(1), newAction(0, fname, tname, big.NewInt(100), 109000, nil))
	if err := types.SignActionWithMultiKey(tx1.GetActions()[0], tx1, types.NewSigner(params.DefaultChainconfig.ChainID), 0, []*types.KeyPair{keyPair}); err != nil {
		panic(err)
	}

	tx2 := newTx(big.NewInt(2), newAction(0, fname, tname, big.NewInt(100), 109000, nil))
	if err := types.SignActionWithMultiKey(tx2.GetActions()[0], tx2, types.NewSigner(params.DefaultChainconfig.ChainID), 0, []*types.KeyPair{keyPair}); err != nil {
		panic(err)
	}

	tx3 := newTx(big.NewInt(1), newAction(0, fname, tname, big.NewInt(100), 109000, nil))
	if err := types.SignActionWithMultiKey(tx3.GetActions()[0], tx3, types.NewSigner(params.DefaultChainconfig.ChainID), 0, []*types.KeyPair{keyPair}); err != nil {
		panic(err)
	}

	// Add the first two transaction, ensure higher priced stays only
	if replace, err := pool.add(tx1, false); err != nil || replace {
		t.Fatalf("first transaction insert failed (%v) or reported replacement (%v)", err, replace)
	}
	if replace, err := pool.add(tx2, false); err != nil || !replace {
		t.Fatalf("second transaction insert failed (%v) or not reported replacement (%v)", err, replace)
	}
	<-pool.requestPromoteExecutables(newAccountSet(pool.signer, fname))
	if pool.pending[fname].Len() != 1 {
		t.Fatal("expected 1 pending transactions, got", pool.pending[fname].Len())
	}
	if tx := pool.pending[fname].txs.items[0]; tx.Hash() != tx2.Hash() {
		t.Fatalf("transaction mismatch: have %x, want %x", tx.Hash(), tx2.Hash())
	}
	// Add the third transaction and ensure it's not saved (smaller price)
	pool.add(tx3, false)
	<-pool.requestPromoteExecutables(newAccountSet(pool.signer, fname))

	if pool.pending[fname].Len() != 1 {
		t.Fatal("expected 1 pending transactions, got", pool.pending[fname].Len())
	}
	if tx := pool.pending[fname].txs.items[0]; tx.Hash() != tx2.Hash() {
		t.Fatalf("transaction mismatch: have %x, want %x", tx.Hash(), tx2.Hash())
	}
	// Ensure the total transaction count is correct
	if pool.all.Count() != 1 {
		t.Fatal("expected 1 total transactions, got", pool.all.Count())
	}
}

func TestTransactionMissingNonce(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager)
	generateAccount(t, tname, manager)

	manager.AddAccountBalanceByID(fname, assetID, big.NewInt(100000000000000))

	tx := transaction(1, fname, tname, 109000, fkey)

	if _, err := pool.add(tx, false); err != nil {
		t.Fatal("didn't expect error", err)
	}
	if len(pool.pending) != 0 {
		t.Fatal("expected 0 pending transactions, got", len(pool.pending))
	}
	if pool.queue[fname].Len() != 1 {
		t.Fatal("expected 1 queued transaction, got", pool.queue[fname].Len())
	}
	if pool.all.Count() != 1 {
		t.Fatal("expected 1 total transactions, got", pool.all.Count())
	}
}

func TestTransactionNonceRecovery(t *testing.T) {

	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager)
	generateAccount(t, tname, manager)

	const n = 10

	manager.AddAccountBalanceByID(fname, assetID, big.NewInt(100000000000000))
	pool.curAccountManager.SetNonce(fname, n)

	<-pool.requestReset(nil, nil)
	tx := transaction(n, fname, tname, 109000, fkey)
	if err := pool.addRemoteSync(tx); err != nil {
		t.Fatal(err)
	}
	// simulate some weird re-order of transactions and missing nonce(s)
	pool.curAccountManager.SetNonce(fname, n-1)
	<-pool.requestReset(nil, nil)
	if fn, _ := pool.pendingAccountManager.GetNonce(fname); fn != n-1 {
		t.Fatalf("expected nonce to be %d, got %d", n-1, fn)
	}
}

// Tests that if an account runs out of funds, any pending and queued transactions
// are dropped.
func TestTransactionDropping(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager)
	generateAccount(t, tname, manager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(1000))

	// Add some pending and some queued transactions
	var (
		tx0  = transaction(0, fname, tname, 100, fkey)
		tx1  = transaction(1, fname, tname, 200, fkey)
		tx2  = transaction(2, fname, tname, 300, fkey)
		tx10 = transaction(10, fname, tname, 100, fkey)
		tx11 = transaction(11, fname, tname, 200, fkey)
		tx12 = transaction(12, fname, tname, 300, fkey)
	)
	pool.promoteTx(fname, tx0.Hash(), tx0)
	pool.promoteTx(fname, tx1.Hash(), tx1)
	pool.promoteTx(fname, tx2.Hash(), tx2)
	pool.enqueueTx(tx10.Hash(), tx10)
	pool.enqueueTx(tx11.Hash(), tx11)
	pool.enqueueTx(tx12.Hash(), tx12)

	// Check that pre and post validations leave the pool as is
	if pool.pending[fname].Len() != 3 {
		t.Fatalf("pending transaction mismatch: have %d, want %d", pool.pending[fname].Len(), 3)
	}
	if pool.queue[fname].Len() != 3 {
		t.Fatalf("queued transaction mismatch: have %d, want %d", pool.queue[fname].Len(), 3)
	}
	if pool.all.Count() != 6 {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), 6)
	}
	<-pool.requestReset(nil, nil)
	if pool.pending[fname].Len() != 3 {
		t.Fatalf("pending transaction mismatch: have %d, want %d", pool.pending[fname].Len(), 3)
	}
	if pool.queue[fname].Len() != 3 {
		t.Fatalf("queued transaction mismatch: have %d, want %d", pool.queue[fname].Len(), 3)
	}
	if pool.all.Count() != 6 {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), 6)
	}
	// Reduce the balance of the account, and check that invalidated transactions are dropped
	pool.curAccountManager.SubAccountBalanceByID(fname, assetID, big.NewInt(750))
	<-pool.requestReset(nil, nil)

	if _, ok := pool.pending[fname].txs.items[tx0.GetActions()[0].Nonce()]; !ok {
		t.Fatalf("funded pending transaction missing: %v", tx0)
	}
	if _, ok := pool.pending[fname].txs.items[tx1.GetActions()[0].Nonce()]; !ok {
		t.Fatalf("funded pending transaction missing: %v", tx0)
	}
	if _, ok := pool.pending[fname].txs.items[tx2.GetActions()[0].Nonce()]; ok {
		t.Fatalf("out-of-fund pending transaction present: %v", tx1)
	}
	if _, ok := pool.queue[fname].txs.items[tx10.GetActions()[0].Nonce()]; !ok {
		t.Fatalf("funded queued transaction missing: %v", tx10)
	}
	if _, ok := pool.queue[fname].txs.items[tx11.GetActions()[0].Nonce()]; !ok {
		t.Fatalf("funded queued transaction missing: %v", tx10)
	}
	if _, ok := pool.queue[fname].txs.items[tx12.GetActions()[0].Nonce()]; ok {
		t.Fatalf("out-of-fund queued transaction present: %v", tx11)
	}
	if pool.all.Count() != 4 {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), 4)
	}
	// Reduce the block gas limit, check that invalidated transactions are dropped
	pool.chain.(*testBlockChain).gasLimit = 100
	<-pool.requestReset(nil, nil)

	if _, ok := pool.pending[fname].txs.items[tx0.GetActions()[0].Nonce()]; !ok {
		t.Fatalf("funded pending transaction missing: %v", tx0)
	}
	if _, ok := pool.pending[fname].txs.items[tx1.GetActions()[0].Nonce()]; ok {
		t.Fatalf("over-gased pending transaction present: %v", tx1)
	}
	if _, ok := pool.queue[fname].txs.items[tx10.GetActions()[0].Nonce()]; !ok {
		t.Fatalf("funded queued transaction missing: %v", tx10)
	}
	if _, ok := pool.queue[fname].txs.items[tx11.GetActions()[0].Nonce()]; ok {
		t.Fatalf("over-gased queued transaction present: %v", tx11)
	}
	if pool.all.Count() != 2 {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), 2)
	}
}

// Tests that if a transaction is dropped from the current pending pool (e.g. out
// of fund), all consecutive (still valid, but not executable) transactions are
// postponed back into the future queue to prevent broadcasting them.
func TestTransactionPostponing(t *testing.T) {
	// Create the pool to test the postponing with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 1000000, new(event.Feed)}
	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager)

	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	// Create two test accounts to produce different gap profiles with
	keys := make([]*ecdsa.PrivateKey, 2)
	accs := make([]common.Name, len(keys))

	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(501000))
	}
	// Add a batch consecutive pending transactions for validation
	txs := []*types.Transaction{}
	for i, key := range keys {
		for j := 0; j < 100; j++ {
			var tx *types.Transaction
			if (i+j)%2 == 0 {
				tx = transaction(uint64(j), accs[i], tname, 250000, key)
			} else {
				tx = transaction(uint64(j), accs[i], tname, 500000, key)
			}
			txs = append(txs, tx)
		}
	}
	for i, err := range pool.addRemotesSync(txs) {
		if err != nil {
			t.Fatalf("tx %d: failed to add transactions: %v", i, err)
		}
	}
	// Check that pre and post validations leave the pool as is
	if pending := pool.pending[accs[0]].Len() + pool.pending[accs[1]].Len(); pending != len(txs) {
		t.Fatalf("pending transaction mismatch: have %d, want %d", pending, len(txs))
	}
	if len(pool.queue) != 0 {
		t.Fatalf("queued accounts mismatch: have %d, want %d", len(pool.queue), 0)
	}
	if pool.all.Count() != len(txs) {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), len(txs))
	}
	<-pool.requestReset(nil, nil)
	if pending := pool.pending[accs[0]].Len() + pool.pending[accs[1]].Len(); pending != len(txs) {
		t.Fatalf("pending transaction mismatch: have %d, want %d", pending, len(txs))
	}
	if len(pool.queue) != 0 {
		t.Fatalf("queued accounts mismatch: have %d, want %d", len(pool.queue), 0)
	}
	if pool.all.Count() != len(txs) {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), len(txs))
	}
	// Reduce the balance of the account, and check that transactions are reorganised
	for _, name := range accs {
		pool.curAccountManager.SubAccountBalanceByID(name, assetID, big.NewInt(1010))
	}

	<-pool.requestReset(nil, nil)

	// The first account's first transaction remains valid, check that subsequent
	// ones are either filtered out, or queued up for later.
	if _, ok := pool.pending[accs[0]].txs.items[txs[0].GetActions()[0].Nonce()]; !ok {
		t.Fatalf("tx %d: valid and funded transaction missing from pending pool: %v", 0, txs[0])
	}
	if _, ok := pool.queue[accs[0]].txs.items[txs[0].GetActions()[0].Nonce()]; ok {
		t.Fatalf("tx %d: valid and funded transaction present in future queue: %v", 0, txs[0])
	}
	for i, tx := range txs[1:100] {
		if i%2 == 1 {
			if _, ok := pool.pending[accs[0]].txs.items[tx.GetActions()[0].Nonce()]; ok {
				t.Fatalf("tx %d: valid but future transaction present in pending pool: %v", i+1, tx)
			}
			if _, ok := pool.queue[accs[0]].txs.items[tx.GetActions()[0].Nonce()]; !ok {
				t.Fatalf("tx %d: valid but future transaction missing from future queue: %v", i+1, tx)
			}
		} else {
			if _, ok := pool.pending[accs[0]].txs.items[tx.GetActions()[0].Nonce()]; ok {
				t.Fatalf("tx %d: out-of-fund transaction present in pending pool: %v", i+1, tx)
			}
			if _, ok := pool.queue[accs[0]].txs.items[tx.GetActions()[0].Nonce()]; ok {
				t.Fatalf("tx %d: out-of-fund transaction present in future queue: %v", i+1, tx)
			}
		}
	}
	// The second account's first transaction got invalid, check that all transactions
	// are either filtered out, or queued up for later.
	if pool.pending[accs[1]] != nil {
		t.Fatalf("invalidated account still has pending transactions")
	}
	for i, tx := range txs[100:] {
		if i%2 == 1 {
			if _, ok := pool.queue[accs[1]].txs.items[tx.GetActions()[0].Nonce()]; !ok {
				t.Fatalf("tx %d: valid but future transaction missing from future queue: %v", 100+i, tx)
			}
		} else {
			if _, ok := pool.queue[accs[1]].txs.items[tx.GetActions()[0].Nonce()]; ok {
				t.Fatalf("tx %d: out-of-fund transaction present in future queue: %v", 100+i, tx)
			}
		}
	}
	if pool.all.Count() != len(txs)/2 {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), len(txs)/2)
	}
}

// Tests that if the transaction pool has both executable and non-executable
// transactions from an origin account, filling the nonce gap moves all queued
// ones into the pending pool.
func TestTransactionGapFilling(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	// Keep track of transaction events to ensure all executables get announced
	events := make(chan *event.Event, testTxPoolConfig.AccountQueue+5)
	sub := event.Subscribe(nil, events, event.NewTxs, []*types.Transaction{})
	defer sub.Unsubscribe()

	// Create a pending and a queued transaction with a nonce-gap in between
	pool.addRemotesSync([]*types.Transaction{
		transaction(0, fname, tname, 1000000, fkey),
		transaction(2, fname, tname, 1000000, fkey),
	})

	pending, queued := pool.Stats()
	if pending != 1 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 1)
	}
	if queued != 1 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 1)
	}

	if err := validateEvents(events, 1); err != nil {
		t.Fatalf("original event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Fill the nonce gap and ensure all transactions become pending
	if err := pool.addRemoteSync(transaction(1, fname, tname, 1000000, fkey)); err != nil {
		t.Fatalf("failed to add gapped transaction: %v", err)
	}
	pending, queued = pool.Stats()

	if pending != 3 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 3)
	}
	if queued != 0 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
	}
	if err := validateEvents(events, 2); err != nil {
		t.Fatalf("gap-filling event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that if the transaction count belonging to a single account goes above
// some threshold, the higher transactions are dropped to prevent DOS attacks.
func TestTransactionQueueAccountLimiting(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager)
	generateAccount(t, tname, manager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	// Keep queuing up transactions and make sure all above a limit are dropped
	for i := uint64(1); i <= testTxPoolConfig.AccountQueue+5; i++ {
		if err := pool.addRemoteSync(transaction(i, fname, tname, 1000000, fkey)); err != nil {
			t.Fatalf("tx %d: failed to add transaction: %v", i, err)
		}
		if len(pool.pending) != 0 {
			t.Fatalf("tx %d: pending pool size mismatch: have %d, want %d", i, len(pool.pending), 0)
		}
		if i <= testTxPoolConfig.AccountQueue {
			if pool.queue[fname].Len() != int(i) {
				t.Fatalf("tx %d: queue size mismatch: have %d, want %d", i, pool.queue[fname].Len(), i)
			}
		} else {
			if pool.queue[fname].Len() != int(testTxPoolConfig.AccountQueue) {
				t.Fatalf("tx %d: queue limit mismatch: have %d, want %d", i, pool.queue[fname].Len(), testTxPoolConfig.AccountQueue)
			}
		}
	}
	if pool.all.Count() != int(testTxPoolConfig.AccountQueue) {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), testTxPoolConfig.AccountQueue)
	}
}

// Tests that if the transaction count belonging to multiple accounts go above
// some threshold, the higher transactions are dropped to prevent DOS attacks.
//
// This logic should not hold for local transactions, unless the local tracking
// mechanism is disabled.
func TestTransactionQueueGlobalLimiting(t *testing.T) {
	testTransactionQueueGlobalLimiting(t, false)
}
func TestTransactionQueueGlobalLimitingNoLocals(t *testing.T) {
	testTransactionQueueGlobalLimiting(t, true)
}

func testTransactionQueueGlobalLimiting(t *testing.T, nolocals bool) {
	// Create the pool to test the limit enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	config := testTxPoolConfig
	config.NoLocals = nolocals
	config.GlobalQueue = config.AccountQueue*3 - 1 // reduce the queue limits to shorten test time (-1 to make it non divisible)

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager)

	pool := New(config, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	// Create a number of test accounts and fund them (last one will be the local)
	keys := make([]*ecdsa.PrivateKey, 5)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	}
	local := keys[len(keys)-1]
	localacc := accs[len(accs)-1]

	// Generate and queue a batch of transactions
	nonces := make(map[common.Name]uint64)

	txs := make([]*types.Transaction, 0, 3*config.GlobalQueue)
	for len(txs) < cap(txs) {
		randInt := rand.Intn(len(keys) - 1)
		key := keys[randInt] // skip adding transactions with the local account

		txs = append(txs, transaction(nonces[accs[randInt]]+1, accs[randInt], tname, 1000000, key))
		nonces[accs[randInt]]++
	}
	// Import the batch and verify that limits have been enforced
	pool.addRemotesSync(txs)

	queued := 0
	for name, list := range pool.queue {
		if list.Len() > int(config.AccountQueue) {
			t.Fatalf("name %x: queued accounts overflown allowance: %d > %d", name, list.Len(), config.AccountQueue)
		}
		queued += list.Len()
	}
	if queued > int(config.GlobalQueue) {
		t.Fatalf("total transactions overflow allowance: %d > %d", queued, config.GlobalQueue)
	}
	// Generate a batch of transactions from the local account and import them
	txs = txs[:0]
	for i := uint64(0); i < 3*config.GlobalQueue; i++ {
		txs = append(txs, transaction(i+1, localacc, tname, 1000000, local))
	}
	pool.AddLocals(txs)

	// If locals are disabled, the previous eviction algorithm should apply here too
	if nolocals {
		queued := 0
		for name, list := range pool.queue {
			if list.Len() > int(config.AccountQueue) {
				t.Fatalf("name %x: queued accounts overflown allowance: %d > %d", name, list.Len(), config.AccountQueue)
			}
			queued += list.Len()
		}
		if queued > int(config.GlobalQueue) {
			t.Fatalf("total transactions overflow allowance: %d > %d", queued, config.GlobalQueue)
		}
	} else {
		// Local exemptions are enabled, make sure the local account owned the queue
		if len(pool.queue) != 1 {
			t.Fatalf("multiple accounts in queue: have %v, want %v", len(pool.queue), 1)
		}
		// Also ensure no local transactions are ever dropped, even if above global limits
		if queued := pool.queue[localacc].Len(); uint64(queued) != 3*config.GlobalQueue {
			t.Fatalf("local account queued transaction count mismatch: have %v, want %v", queued, 3*config.GlobalQueue)
		}
	}
}

// Tests that if an account remains idle for a prolonged amount of time, any
// non-executable transactions queued up are dropped to prevent wasting resources
// on shuffling them around.
//
// This logic should not hold for local transactions, unless the local tracking
// mechanism is disabled.
func TestTransactionQueueTimeLimiting(t *testing.T)         { testTransactionQueueTimeLimiting(t, false) }
func TestTransactionQueueTimeLimitingNoLocals(t *testing.T) { testTransactionQueueTimeLimiting(t, true) }

func testTransactionQueueTimeLimiting(t *testing.T, nolocals bool) {

	// Reduce the eviction interval to a testable amount
	defer func(old time.Duration) { evictionInterval = old }(evictionInterval)
	evictionInterval = time.Second

	// Create the pool to test the non-expiration enforcement
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	config := testTxPoolConfig
	config.Lifetime = time.Second
	config.NoLocals = nolocals

	pool := New(config, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	var (
		localName  = common.Name("localname")
		remoteName = common.Name("remotename")

		tname   = common.Name("totestname")
		assetID = uint64(0)
	)

	manager, _ := am.NewAccountManager(statedb)
	local := generateAccount(t, localName, manager, pool.pendingAccountManager)
	remote := generateAccount(t, remoteName, manager, pool.pendingAccountManager)
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	pool.curAccountManager.AddAccountBalanceByID(localName, assetID, big.NewInt(10000000000))
	pool.curAccountManager.AddAccountBalanceByID(remoteName, assetID, big.NewInt(10000000000))

	// Add the two transactions and ensure they both are queued up
	if err := pool.AddLocal(pricedTransaction(1, localName, tname, 109000, big.NewInt(1), local)); err != nil {
		t.Fatalf("failed to add local transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(1, remoteName, tname, 109000, big.NewInt(1), remote)); err != nil {
		t.Fatalf("failed to add remote transaction: %v", err)
	}
	pending, queued := pool.Stats()
	if pending != 0 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 0)
	}
	if queued != 2 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 2)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Wait a bit for eviction to run and clean up any leftovers, and ensure only the local remains
	time.Sleep(2 * config.Lifetime)

	pending, queued = pool.Stats()
	if pending != 0 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 0)
	}
	if nolocals {
		if queued != 0 {
			t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
		}
	} else {
		if queued != 1 {
			t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 1)
		}
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that even if the transaction count belonging to a single account goes
// above some threshold, as long as the transactions are executable, they are
// accepted.
func TestTransactionPendingLimiting(t *testing.T) {
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	// Keep track of transaction events to ensure all executables get announced
	events := make(chan *event.Event, testTxPoolConfig.AccountQueue+5)
	sub := event.Subscribe(nil, events, event.NewTxs, []*types.Transaction{})
	defer sub.Unsubscribe()

	// Keep queuing up transactions and make sure all above a limit are dropped
	for i := uint64(0); i < testTxPoolConfig.AccountQueue+5; i++ {
		if err := pool.addRemoteSync(transaction(i, fname, tname, 1000000, fkey)); err != nil {
			t.Fatalf("tx %d: failed to add transaction: %v", i, err)
		}
		if pool.pending[fname].Len() != int(i)+1 {
			t.Fatalf("tx %d: pending pool size mismatch: have %d, want %d", i, pool.pending[fname].Len(), i+1)
		}
		if len(pool.queue) != 0 {
			t.Fatalf("tx %d: queue size mismatch: have %d, want %d", i, pool.queue[fname].Len(), 0)
		}
	}
	if pool.all.Count() != int(testTxPoolConfig.AccountQueue+5) {
		t.Fatalf("total transaction mismatch: have %d, want %d", pool.all.Count(), testTxPoolConfig.AccountQueue+5)
	}
	if err := validateEvents(events, int(testTxPoolConfig.AccountQueue+5)); err != nil {
		t.Fatalf("event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that setting the transaction pool gas price to a higher value correctly
// discards everything cheaper than that and moves any gapped transactions back
// from the pending pool to the queue.
//
// Note, local transactions are never allowed to be dropped.
func TestTransactionPoolRepricing(t *testing.T) {
	// Create the pool to test the pricing enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))

	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	// Keep track of transaction events to ensure all executables get announced
	events := make(chan *event.Event, 32)
	sub := event.Subscribe(nil, events, event.NewTxs, []*types.Transaction{})
	defer sub.Unsubscribe()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 4)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))
	}
	// Generate and queue a batch of transactions, both pending and queued
	txs := []*types.Transaction{}

	txs = append(txs, pricedTransaction(0, accs[0], tname, 1000000, big.NewInt(2), keys[0]))
	txs = append(txs, pricedTransaction(1, accs[0], tname, 1000000, big.NewInt(1), keys[0]))
	txs = append(txs, pricedTransaction(2, accs[0], tname, 1000000, big.NewInt(2), keys[0]))

	txs = append(txs, pricedTransaction(0, accs[1], tname, 1000000, big.NewInt(1), keys[1]))
	txs = append(txs, pricedTransaction(1, accs[1], tname, 1000000, big.NewInt(2), keys[1]))
	txs = append(txs, pricedTransaction(2, accs[1], tname, 1000000, big.NewInt(2), keys[1]))

	txs = append(txs, pricedTransaction(1, accs[2], tname, 1000000, big.NewInt(2), keys[2]))
	txs = append(txs, pricedTransaction(2, accs[2], tname, 1000000, big.NewInt(1), keys[2]))
	txs = append(txs, pricedTransaction(3, accs[2], tname, 1000000, big.NewInt(2), keys[2]))

	ltx := pricedTransaction(0, accs[3], tname, 1000000, big.NewInt(1), keys[3])

	// Import the batch and that both pending and queued transactions match up
	pool.addRemotesSync(txs)
	pool.AddLocal(ltx)

	pending, queued := pool.Stats()
	if pending != 7 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 7)
	}
	if queued != 3 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 3)
	}
	if err := validateEvents(events, 7); err != nil {
		t.Fatalf("original event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Reprice the pool and check that underpriced transactions get dropped
	pool.SetGasPrice(big.NewInt(2))

	pending, queued = pool.Stats()
	if pending != 2 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 2)
	}
	if queued != 5 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 5)
	}
	if err := validateEvents(events, 0); err != nil {
		t.Fatalf("reprice event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Check that we can't add the old transactions back
	if err := pool.addRemoteSync(pricedTransaction(1, accs[0], tname, 1000000, big.NewInt(1), keys[0])); err != ErrUnderpriced {
		t.Fatalf("adding underpriced pending transaction error mismatch: have %v, want %v", err, ErrUnderpriced)
	}
	if err := pool.addRemoteSync(pricedTransaction(0, accs[1], tname, 1000000, big.NewInt(1), keys[1])); err != ErrUnderpriced {
		t.Fatalf("adding underpriced pending transaction error mismatch: have %v, want %v", err, ErrUnderpriced)
	}
	if err := pool.addRemoteSync(pricedTransaction(2, accs[2], tname, 1000000, big.NewInt(1), keys[2])); err != ErrUnderpriced {
		t.Fatalf("adding underpriced queued transaction error mismatch: have %v, want %v", err, ErrUnderpriced)
	}
	if err := validateEvents(events, 0); err != nil {
		t.Fatalf("post-reprice event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// However we can add local underpriced transactions
	tx := pricedTransaction(1, accs[3], tname, 1000000, big.NewInt(1), keys[3])
	if err := pool.AddLocal(tx); err != nil {
		t.Fatalf("failed to add underpriced local transaction: %v", err)
	}
	if pending, _ = pool.Stats(); pending != 3 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 3)
	}
	if err := validateEvents(events, 1); err != nil {
		t.Fatalf("post-reprice local event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// And we can fill gaps with properly priced transactions
	if err := pool.addRemoteSync(pricedTransaction(1, accs[0], tname, 1000000, big.NewInt(2), keys[0])); err != nil {
		t.Fatalf("failed to add pending transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(0, accs[1], tname, 1000000, big.NewInt(2), keys[1])); err != nil {
		t.Fatalf("failed to add pending transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(2, accs[2], tname, 1000000, big.NewInt(2), keys[2])); err != nil {
		t.Fatalf("failed to add queued transaction: %v", err)
	}
	if err := validateEvents(events, 5); err != nil {
		t.Fatalf("post-reprice event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that setting the transaction pool gas price to a higher value does not
// remove local transactions.
func TestTransactionPoolRepricingKeepsLocals(t *testing.T) {
	// Create the pool to test the pricing enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 3)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(1000*10000000))
	}

	// Create transaction (both pending and queued) with a linearly growing gasprice
	for i := uint64(0); i < 500; i++ {
		// Add pending
		pendingtx := pricedTransaction(i, accs[2], tname, 1000000, big.NewInt(int64(i)), keys[2])
		if err := pool.AddLocal(pendingtx); err != nil {
			t.Fatal(err)
		}
		// Add queued
		queuetx := pricedTransaction(i+501, accs[2], tname, 1000000, big.NewInt(int64(i)), keys[2])
		if err := pool.AddLocal(queuetx); err != nil {
			t.Fatal(err)
		}
	}
	pending, queued := pool.Stats()
	expPending, expQueued := 500, 500
	validate := func() {
		pending, queued = pool.Stats()
		if pending != expPending {
			t.Fatalf("pending transactions mismatched: have %d, want %d", pending, expPending)
		}
		if queued != expQueued {
			t.Fatalf("queued transactions mismatched: have %d, want %d", queued, expQueued)
		}

		if err := validateTxPoolInternals(pool); err != nil {
			t.Fatalf("pool internal state corrupted: %v", err)
		}
	}
	validate()

	// Reprice the pool and check that nothing is dropped
	pool.SetGasPrice(big.NewInt(2))
	validate()

	pool.SetGasPrice(big.NewInt(2))
	pool.SetGasPrice(big.NewInt(4))
	pool.SetGasPrice(big.NewInt(8))
	pool.SetGasPrice(big.NewInt(100))
	validate()
}

// Tests that when the pool reaches its global transaction limit, underpriced
// transactions are gradually shifted out for more expensive ones and any gapped
// pending transactions are moved into the queue.
//
// Note, local transactions are never allowed to be dropped.
func TestTransactionPoolUnderpricing(t *testing.T) {
	// Create the pool to test the pricing enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))

	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	config := testTxPoolConfig
	config.GlobalSlots = 2
	config.GlobalQueue = 2

	pool := New(config, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Keep track of transaction events to ensure all executables get announced
	events := make(chan *event.Event, 32)
	sub := event.Subscribe(nil, events, event.NewTxs, []*types.Transaction{})
	defer sub.Unsubscribe()

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 4)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))
	}
	// Generate and queue a batch of transactions, both pending and queued
	txs := []*types.Transaction{}

	txs = append(txs, pricedTransaction(0, accs[0], tname, 1000000, big.NewInt(1), keys[0]))
	txs = append(txs, pricedTransaction(1, accs[0], tname, 1000000, big.NewInt(2), keys[0]))
	txs = append(txs, pricedTransaction(1, accs[1], tname, 1000000, big.NewInt(1), keys[1]))

	ltx := pricedTransaction(0, accs[2], tname, 1000000, big.NewInt(1), keys[2])

	// Import the batch and that both pending and queued transactions match up
	pool.addRemotesSync(txs)
	pool.AddLocal(ltx)

	pending, queued := pool.Stats()
	if pending != 3 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 3)
	}
	if queued != 1 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 1)
	}
	if err := validateEvents(events, 3); err != nil {
		t.Fatalf("original event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Ensure that adding an underpriced transaction on block limit fails
	if err := pool.addRemoteSync(pricedTransaction(0, accs[1], tname, 1000000, big.NewInt(1), keys[1])); err != ErrUnderpriced {
		t.Fatalf("adding underpriced pending transaction error mismatch: have %v, want %v", err, ErrUnderpriced)
	}
	// Ensure that adding high priced transactions drops cheap ones, but not own
	if err := pool.addRemoteSync(pricedTransaction(0, accs[1], tname, 1000000, big.NewInt(3), keys[1])); err != nil { // +K1:0 => -K1:1 => Pend K0:0, K0:1, K1:0, K2:0; Que -
		t.Fatalf("failed to add well priced transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(2, accs[1], tname, 1000000, big.NewInt(4), keys[1])); err != nil { // +K1:2 => -K0:0 => Pend K1:0, K2:0; Que K0:1 K1:2
		t.Fatalf("failed to add well priced transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(3, accs[1], tname, 1000000, big.NewInt(5), keys[1])); err != nil { // +K1:3 => -K0:1 => Pend K1:0, K2:0; Que K1:2 K1:3
		t.Fatalf("failed to add well priced transaction: %v", err)
	}
	pending, queued = pool.Stats()
	if pending != 2 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 2)
	}
	if queued != 2 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 2)
	}
	if err := validateEvents(events, 1); err != nil {
		t.Fatalf("additional event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Ensure that adding local transactions can push out even higher priced ones
	ltx = pricedTransaction(1, accs[2], tname, 1000000, big.NewInt(0), keys[2])
	if err := pool.AddLocal(ltx); err != nil {
		t.Fatalf("failed to append underpriced local transaction: %v", err)
	}
	ltx = pricedTransaction(0, accs[3], tname, 1000000, big.NewInt(0), keys[3])
	if err := pool.AddLocal(ltx); err != nil {
		t.Fatalf("failed to add new underpriced local transaction: %v", err)
	}
	pending, queued = pool.Stats()
	if pending != 3 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 3)
	}
	if queued != 1 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 1)
	}
	if err := validateEvents(events, 2); err != nil {
		t.Fatalf("local event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that more expensive transactions push out cheap ones from the pool, but
// without producing instability by creating gaps that start jumping transactions
// back and forth between queued/pending.
func TestTransactionPoolStableUnderpricing(t *testing.T) {
	// Create the pool to test the pricing enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	config := testTxPoolConfig
	config.GlobalSlots = 128
	config.GlobalQueue = 0

	pool := New(config, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Keep track of transaction events to ensure all executables get announced
	events := make(chan *event.Event, 32)
	sub := event.Subscribe(nil, events, event.NewTxs, []*types.Transaction{})
	defer sub.Unsubscribe()

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 2)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))
	}

	// Fill up the entire queue with the same transaction price points
	txs := []*types.Transaction{}
	for i := uint64(0); i < config.GlobalSlots; i++ {
		txs = append(txs, pricedTransaction(i, accs[0], tname, 1000000, big.NewInt(1), keys[0]))
	}
	pool.addRemotesSync(txs)

	pending, queued := pool.Stats()
	if pending != int(config.GlobalSlots) {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, config.GlobalSlots)
	}
	if queued != 0 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
	}
	if err := validateEvents(events, int(config.GlobalSlots)); err != nil {
		t.Fatalf("original event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Ensure that adding high priced transactions drops a cheap, but doesn't produce a gap
	if err := pool.addRemoteSync(pricedTransaction(0, accs[1], tname, 1000000, big.NewInt(3), keys[1])); err != nil {
		t.Fatalf("failed to add well priced transaction: %v", err)
	}
	pending, queued = pool.Stats()
	if pending != int(config.GlobalSlots) {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, config.GlobalSlots)
	}
	if queued != 0 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
	}
	if err := validateEvents(events, 1); err != nil {
		t.Fatalf("additional event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that the pool rejects replacement transactions that don't meet the minimum
// price bump required.
func TestTransactionReplacement(t *testing.T) {
	// Create the pool to test the pricing enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	fname := common.Name("fromname")
	fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000000))

	// Keep track of transaction events to ensure all executables get announced
	events := make(chan *event.Event, 32)
	sub := event.Subscribe(nil, events, event.NewTxs, []*types.Transaction{})
	defer sub.Unsubscribe()

	// Add pending transactions, ensuring the minimum price bump is enforced for replacement (for ultra low prices too)
	price := int64(100)
	threshold := (price * (100 + int64(testTxPoolConfig.PriceBump))) / 100

	if err := pool.addRemoteSync(pricedTransaction(0, fname, tname, 1000000, big.NewInt(1), fkey)); err != nil {
		t.Fatalf("failed to add original cheap pending transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(0, fname, tname, 1000010, big.NewInt(1), fkey)); err != ErrReplaceUnderpriced {
		t.Fatalf("original cheap pending transaction replacement error mismatch: have %v, want %v", err, ErrReplaceUnderpriced)
	}
	if err := pool.addRemoteSync(pricedTransaction(0, fname, tname, 1000000, big.NewInt(2), fkey)); err != nil {
		t.Fatalf("failed to replace original cheap pending transaction: %v", err)
	}
	if err := validateEvents(events, 2); err != nil {
		t.Fatalf("cheap replacement event firing failed: %v", err)
	}

	if err := pool.addRemoteSync(pricedTransaction(0, fname, tname, 1000000, big.NewInt(price), fkey)); err != nil {
		t.Fatalf("failed to add original proper pending transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(0, fname, tname, 1000010, big.NewInt(threshold-1), fkey)); err != ErrReplaceUnderpriced {
		t.Fatalf("original proper pending transaction replacement error mismatch: have %v, want %v", err, ErrReplaceUnderpriced)
	}
	if err := pool.addRemoteSync(pricedTransaction(0, fname, tname, 1000000, big.NewInt(threshold), fkey)); err != nil {
		t.Fatalf("failed to replace original proper pending transaction: %v", err)
	}
	if err := validateEvents(events, 2); err != nil {
		t.Fatalf("proper replacement event firing failed: %v", err)
	}

	// Add queued transactions, ensuring the minimum price bump is enforced for replacement (for ultra low prices too)
	if err := pool.addRemoteSync(pricedTransaction(2, fname, tname, 1000000, big.NewInt(1), fkey)); err != nil {
		t.Fatalf("failed to add original cheap queued transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(2, fname, tname, 1000010, big.NewInt(1), fkey)); err != ErrReplaceUnderpriced {
		t.Fatalf("original cheap queued transaction replacement error mismatch: have %v, want %v", err, ErrReplaceUnderpriced)
	}
	if err := pool.addRemoteSync(pricedTransaction(2, fname, tname, 1000000, big.NewInt(2), fkey)); err != nil {
		t.Fatalf("failed to replace original cheap queued transaction: %v", err)
	}

	if err := pool.addRemoteSync(pricedTransaction(2, fname, tname, 1000000, big.NewInt(price), fkey)); err != nil {
		t.Fatalf("failed to add original proper queued transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(2, fname, tname, 1000010, big.NewInt(threshold-1), fkey)); err != ErrReplaceUnderpriced {
		t.Fatalf("original proper queued transaction replacement error mismatch: have %v, want %v", err, ErrReplaceUnderpriced)
	}
	if err := pool.addRemoteSync(pricedTransaction(2, fname, tname, 1000000, big.NewInt(threshold), fkey)); err != nil {
		t.Fatalf("failed to replace original proper queued transaction: %v", err)
	}

	if err := validateEvents(events, 0); err != nil {
		t.Fatalf("queued replacement event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that if the transaction count belonging to multiple accounts go above
// some hard threshold, the higher transactions are dropped to prevent DOS
// attacks.
func TestTransactionPendingGlobalLimiting(t *testing.T) {

	// Create the pool to test the limit enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	config := testTxPoolConfig
	config.GlobalSlots = config.AccountSlots * 10

	pool := New(config, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 5)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	}
	// Generate and queue a batch of transactions
	nonces := make(map[common.Name]uint64)

	txs := []*types.Transaction{}
	for i, key := range keys {
		for j := 0; j < int(config.GlobalSlots)/len(keys)*2; j++ {
			txs = append(txs, transaction(nonces[accs[i]], accs[i], tname, 1000000, key))
			nonces[accs[i]]++
		}
	}
	// Import the batch and verify that limits have been enforced
	pool.addRemotesSync(txs)

	pending := 0
	for _, list := range pool.pending {
		pending += list.Len()
	}
	if pending > int(config.GlobalSlots) {
		t.Fatalf("total pending transactions overflow allowance: %d > %d", pending, config.GlobalSlots)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that if transactions start being capped, transactions are also removed from 'all'
func TestTransactionCapClearsFromAll(t *testing.T) {
	// Create the pool to test the limit enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	config := testTxPoolConfig
	config.AccountSlots = 2
	config.AccountQueue = 2
	config.GlobalSlots = 8

	pool := New(config, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	fname := common.Name("fromname")
	fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Create a number of test accounts and fund them

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	txs := []*types.Transaction{}
	for j := 0; j < int(config.GlobalSlots)*2; j++ {
		txs = append(txs, transaction(uint64(j), fname, tname, 1000000, fkey))
	}
	// Import the batch and verify that limits have been enforced
	pool.addRemotesSync(txs)
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that if the transaction count belonging to multiple accounts go above
// some hard threshold, if they are under the minimum guaranteed slot count then
// the transactions are still kept.
func TestTransactionPendingMinimumAllowance(t *testing.T) {
	// Create the pool to test the limit enforcement with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))

	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	event.Reset()
	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	pool.config.GlobalSlots = 0
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 5)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)

		keys[i] = fkey
		accs[i] = fname

		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))
	}
	// Generate and queue a batch of transactions
	nonces := make(map[common.Name]uint64)

	txs := []*types.Transaction{}
	for i, key := range keys {
		for j := 0; j < int(testTxPoolConfig.AccountSlots)*2; j++ {
			txs = append(txs, transaction(nonces[accs[i]], accs[i], tname, 1000000, key))
			nonces[accs[i]]++
		}
	}
	// Import the batch and verify that limits have been enforced
	pool.addRemotesSync(txs)

	for name, list := range pool.pending {
		if list.Len() != int(testTxPoolConfig.AccountSlots) {
			t.Fatalf("name %s: total pending transactions mismatch: have %d, want %d", name, list.Len(), testTxPoolConfig.AccountSlots)
		}
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that local transactions are journaled to disk, but remote transactions
// get discarded between restarts.
func TestTransactionJournaling(t *testing.T)         { testTransactionJournaling(t, false) }
func TestTransactionJournalingNoLocals(t *testing.T) { testTransactionJournaling(t, true) }

func testTransactionJournaling(t *testing.T, nolocals bool) {
	// Create a temporary file for the journal
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create temporary journal: %v", err)
	}
	journal := file.Name()
	defer os.Remove(journal)

	// Clean up the temporary file, we only need the path for now
	file.Close()
	os.Remove(journal)

	event.Reset()
	// Create the original pool to inject transaction into the journal
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	config := testTxPoolConfig
	config.NoLocals = nolocals
	config.Journal = journal
	config.Rejournal = time.Second

	pool := New(config, params.DefaultChainconfig, blockchain)

	var (
		localName  = common.Name("localname")
		remoteName = common.Name("remotename")

		tname   = common.Name("totestname")
		assetID = uint64(0)
	)

	manager, _ := am.NewAccountManager(statedb)
	local := generateAccount(t, localName, manager, pool.pendingAccountManager)
	remote := generateAccount(t, remoteName, manager, pool.pendingAccountManager)
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	pool.curAccountManager.AddAccountBalanceByID(localName, assetID, big.NewInt(10000000000))
	pool.curAccountManager.AddAccountBalanceByID(remoteName, assetID, big.NewInt(10000000000))

	// Add three local and a remote transactions and ensure they are queued up
	if err := pool.AddLocal(pricedTransaction(0, localName, tname, 1000000, big.NewInt(1), local)); err != nil {
		t.Fatalf("failed to add local transaction: %v", err)
	}
	if err := pool.AddLocal(pricedTransaction(1, localName, tname, 1000000, big.NewInt(1), local)); err != nil {
		t.Fatalf("failed to add local transaction: %v", err)
	}
	if err := pool.AddLocal(pricedTransaction(2, localName, tname, 1000000, big.NewInt(1), local)); err != nil {
		t.Fatalf("failed to add local transaction: %v", err)
	}
	if err := pool.addRemoteSync(pricedTransaction(0, remoteName, tname, 1000000, big.NewInt(1), remote)); err != nil {
		t.Fatalf("failed to add remote transaction: %v", err)
	}
	pending, queued := pool.Stats()
	if pending != 4 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 4)
	}
	if queued != 0 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Terminate the old pool, bump the local nonce, create a new pool and ensure relevant transaction survive
	pool.Stop()

	event.Reset()
	manager.SetNonce(localName, 1)
	blockchain = &testBlockChain{statedb, 10000000, new(event.Feed)}
	pool = New(config, params.DefaultChainconfig, blockchain)
	pending, queued = pool.Stats()
	if queued != 0 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
	}
	if nolocals {
		if pending != 0 {
			t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 0)
		}
	} else {
		if pending != 2 {
			t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 2)
		}
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Bump the nonce temporarily and ensure the newly invalidated transaction is removed
	manager.SetNonce(localName, 2)
	<-pool.requestReset(nil, nil)

	time.Sleep(2 * config.Rejournal)
	pool.Stop()

	event.Reset()
	manager.SetNonce(localName, 1)
	blockchain = &testBlockChain{statedb, 10000000, new(event.Feed)}
	pool = New(config, params.DefaultChainconfig, blockchain)
	pending, queued = pool.Stats()
	if pending != 0 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 0)
	}
	if nolocals {
		if queued != 0 {
			t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
		}
	} else {
		if queued != 1 {
			t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 1)
		}
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	pool.Stop()
}

// TestTransactionStatusCheck tests that the pool can correctly retrieve the
// pending status of individual transactions.
func TestTransactionStatusCheck(t *testing.T) {
	// Create the pool to test the status retrievals with
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}
	event.Reset()
	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)
	defer pool.Stop()

	manager, _ := am.NewAccountManager(statedb)
	assetID := uint64(0)
	tname := common.Name("totestname")
	generateAccount(t, tname, manager, pool.pendingAccountManager)

	// Create the test accounts to check various transaction statuses with
	keys := make([]*ecdsa.PrivateKey, 3)
	accs := make([]common.Name, len(keys))
	for i := 0; i < len(keys); i++ {
		fname := common.Name("fromname" + strconv.Itoa(i))
		fkey := generateAccount(t, fname, manager, pool.pendingAccountManager)

		keys[i] = fkey
		accs[i] = fname
		pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))
	}

	// Generate and queue a batch of transactions, both pending and queued
	txs := []*types.Transaction{}

	txs = append(txs, pricedTransaction(0, accs[0], tname, 1000000, big.NewInt(1), keys[0])) // Pending only
	txs = append(txs, pricedTransaction(0, accs[1], tname, 1000000, big.NewInt(1), keys[1])) // Pending and queued
	txs = append(txs, pricedTransaction(2, accs[1], tname, 1000000, big.NewInt(1), keys[1]))
	txs = append(txs, pricedTransaction(2, accs[2], tname, 1000000, big.NewInt(1), keys[2])) // Queued only

	// Import the transaction and ensure they are correctly added
	pool.addRemotesSync(txs)

	pending, queued := pool.Stats()
	if pending != 2 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 2)
	}
	if queued != 2 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 2)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Retrieve the status of each transaction and validate them
	hashes := make([]common.Hash, len(txs))
	for i, tx := range txs {
		hashes[i] = tx.Hash()
	}
	hashes = append(hashes, common.Hash{})

	statuses := pool.Status(hashes)
	expect := []TxStatus{TxStatusPending, TxStatusPending, TxStatusQueued, TxStatusQueued, TxStatusUnknown}

	for i := 0; i < len(statuses); i++ {
		if statuses[i] != expect[i] {
			t.Fatalf("transaction %d: status mismatch: have %v, want %v", i, statuses[i], expect[i])
		}
	}
}

// Benchmarks the speed of validating the contents of the pending queue of the
// transaction pool.
func BenchmarkPendingDemotion100(b *testing.B)   { benchmarkPendingDemotion(b, 100) }
func BenchmarkPendingDemotion1000(b *testing.B)  { benchmarkPendingDemotion(b, 1000) }
func BenchmarkPendingDemotion10000(b *testing.B) { benchmarkPendingDemotion(b, 10000) }

func benchmarkPendingDemotion(b *testing.B, size int) {
	// Add a batch of transactions to a pool one by one
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(nil, fname, manager)
	generateAccount(nil, tname, manager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	for i := 0; i < size; i++ {
		tx := transaction(uint64(i), fname, tname, 1000000, fkey)
		pool.promoteTx(fname, tx.Hash(), tx)
	}
	// Benchmark the speed of pool validation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.demoteUnexecutables()
	}
}

// Benchmarks the speed of scheduling the contents of the future queue of the
// transaction pool.
func BenchmarkFuturePromotion100(b *testing.B)   { benchmarkFuturePromotion(b, 100) }
func BenchmarkFuturePromotion1000(b *testing.B)  { benchmarkFuturePromotion(b, 1000) }
func BenchmarkFuturePromotion10000(b *testing.B) { benchmarkFuturePromotion(b, 10000) }

func benchmarkFuturePromotion(b *testing.B, size int) {

	// Add a batch of transactions to a pool one by one
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(nil, fname, manager)
	generateAccount(nil, tname, manager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	for i := 0; i < size; i++ {
		tx := transaction(uint64(1+i), fname, tname, 100000, fkey)
		pool.enqueueTx(tx.Hash(), tx)
	}
	// Benchmark the speed of pool validation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.promoteExecutables(nil)
	}
}

// Benchmarks the speed of batched transaction insertion.
func BenchmarkPoolBatchInsert100(b *testing.B)   { benchmarkPoolBatchInsert(b, 100) }
func BenchmarkPoolBatchInsert1000(b *testing.B)  { benchmarkPoolBatchInsert(b, 1000) }
func BenchmarkPoolBatchInsert10000(b *testing.B) { benchmarkPoolBatchInsert(b, 10000) }

func benchmarkPoolBatchInsert(b *testing.B, size int) {
	// Generate a batch of transactions to enqueue into the pool
	var (
		fname   = common.Name("fromname")
		tname   = common.Name("totestname")
		assetID = uint64(0)
	)
	pool, manager := setupTxPool(fname)
	defer pool.Stop()
	fkey := generateAccount(nil, fname, manager)
	generateAccount(nil, tname, manager)

	pool.curAccountManager.AddAccountBalanceByID(fname, assetID, big.NewInt(10000000))

	batches := make([][]*types.Transaction, b.N)
	for i := 0; i < b.N; i++ {
		batches[i] = make([]*types.Transaction, size)
		for j := 0; j < size; j++ {
			batches[i][j] = transaction(uint64(size*i+j), fname, tname, 1000000, fkey)
		}
	}
	// Benchmark importing the transactions into the queue
	b.ResetTimer()
	for _, batch := range batches {
		pool.addRemotesSync(batch)
	}
}
