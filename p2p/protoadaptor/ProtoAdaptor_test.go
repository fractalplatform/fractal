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

package protoadaptor

import (
	"crypto/ecdsa"
	"net"
	"testing"
	"time"

	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/p2p"
)

func newkey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key
}
func TestPeerPeriod(t *testing.T) {
	config := &p2p.Config{
		Name:       "test",
		MaxPeers:   10,
		ListenAddr: "127.0.0.1:0",
		PrivateKey: newkey(),
	}
	srv := NewProtoAdaptor(config)
	srv.Start()
	defer srv.Stop()
	conn, err := net.DialTimeout("tcp", srv.ListenAddr, 5*time.Second)
	if err != nil {
		t.Fatalf("could not dial: %v", err)
	}
	defer conn.Close()
}
