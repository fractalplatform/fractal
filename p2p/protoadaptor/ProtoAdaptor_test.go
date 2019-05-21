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
