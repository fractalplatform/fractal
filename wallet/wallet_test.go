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

package wallet

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/wallet/keystore"
	"github.com/stretchr/testify/assert"
)

func TestWallet(t *testing.T) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	w := NewWallet(d, keystore.StandardScryptN, keystore.StandardScryptP)

	// test wallet.NewAccount
	a, err := w.NewAccount("password")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(a.Path, d) {
		t.Errorf("account file %s doesn't have dir prefix", a.Path)
	}

	stat, err := os.Stat(a.Path)
	if err != nil {
		t.Fatalf("account file %s doesn't exist (%v)", a.Path, err)
	}

	if runtime.GOOS != "windows" && stat.Mode() != 0600 {
		t.Fatalf("account file has wrong mode: got %o, want %o", stat.Mode(), 0600)
	}

	// test wallet.HasAddress
	if !w.HasAddress(a.Addr) {
		t.Fatalf("HasAccount(%x) should've returned true", a.Addr)
	}

	// test wallet.Update
	if err := w.Update(a, "password", "newpassword"); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// test wallet.Delete
	if err := w.Delete(a, "newpassword"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	if common.FileExist(a.Path) {
		t.Fatalf("account file %s should be gone after Delete", a.Path)
	}

	if w.HasAddress(a.Addr) {
		t.Fatalf("HasAccount(%x) should've returned true after Delete", a.Addr)
	}
}

func TestExportAndImportKey(t *testing.T) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	w := NewWallet(d, keystore.StandardScryptN, keystore.StandardScryptP)

	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), crand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	password := "password"
	newpassword := "newpassword"
	newnewpassword := "newnewpassword"

	// test wallet.ImportECDSA
	a, err := w.ImportECDSA(privateKeyECDSA, password)
	if err != nil {
		t.Fatal(err)
	}

	// test wallet.Export
	jsonbytes, err := w.Export(a, password, newpassword)
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Delete(a, password); err != nil {
		t.Fatal(err)
	}

	// test wallet.Import
	newA, err := w.Import(jsonbytes, newpassword, newnewpassword)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, a.Addr, newA.Addr)
}

func TestSignWithPassphrase(t *testing.T) {
	var hash = make([]byte, 32)

	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	w := NewWallet(d, keystore.StandardScryptN, keystore.StandardScryptP)

	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), crand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	password := "password"

	sig, err := crypto.Sign(hash, privateKeyECDSA)
	if err != nil {
		t.Fatal(err)
	}

	a, err := w.ImportECDSA(privateKeyECDSA, password)
	if err != nil {
		t.Fatal(err)
	}

	nSig, err := w.SignHashWithPassphrase(a, password, hash)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, nSig, sig)
}
