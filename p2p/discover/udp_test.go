// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package discover

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/p2p/enode"
	"github.com/fractalplatform/fractal/utils/rlp"
)

func init() {
	spew.Config.DisableMethods = true
}

// shared test variables
var (
	futureExp          = uint64(time.Now().Add(10 * time.Hour).Unix())
	testTarget         = encPubkey{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1}
	testRemote         = rpcEndpoint{IP: net.ParseIP("1.1.1.1").To4(), UDP: 1, TCP: 2, Rest: []rlp.RawValue{}}
	testLocalAnnounced = rpcEndpoint{IP: net.ParseIP("2.2.2.2").To4(), UDP: 3, TCP: 4, Rest: []rlp.RawValue{}}
	testLocal          = rpcEndpoint{IP: net.ParseIP("3.3.3.3").To4(), UDP: 5, TCP: 6, Rest: []rlp.RawValue{}}
)

type udpTest struct {
	t                   *testing.T
	pipe                *dgramPipe
	table               *Table
	udp                 *udp
	sent                [][]byte
	localkey, remotekey *ecdsa.PrivateKey
	remoteaddr          *net.UDPAddr
}

func newUDPTest(t *testing.T) *udpTest {
	test := &udpTest{
		t:          t,
		pipe:       newpipe(),
		localkey:   newkey(),
		remotekey:  newkey(),
		remoteaddr: &net.UDPAddr{IP: net.IP{10, 0, 1, 99}, Port: 30303},
	}
	test.table, test.udp, _ = newUDP(test.pipe, Config{PrivateKey: test.localkey})
	// Wait for initial refresh so the table doesn't send unexpected findnode.
	<-test.table.initDone
	return test
}

// handles a packet as if it had been sent to the transport.
func (test *udpTest) packetIn(wantError error, ptype byte, data packet) error {
	enc, _, err := encodePacket(test.remotekey, ptype, data)
	if err != nil {
		return test.errorf("packet (%d) encode error: %v", ptype, err)
	}
	test.sent = append(test.sent, enc)
	if err = test.udp.handlePacket(test.remoteaddr, enc); err != wantError {
		return test.errorf("error mismatch: got %q, want %q", err, wantError)
	}
	return nil
}

// waits for a packet to be sent by the transport.
// validate should have type func(*udpTest, X) error, where X is a packet type.
func (test *udpTest) waitPacketOut(validate interface{}) ([]byte, error) {
	dgram := test.pipe.waitPacketOut()
	p, _, hash, err := decodePacket(dgram)
	if err != nil {
		return hash, test.errorf("sent packet decode error: %v", err)
	}
	fn := reflect.ValueOf(validate)
	exptype := fn.Type().In(0)
	if reflect.TypeOf(p) != exptype {
		return hash, test.errorf("sent packet type mismatch, got: %v, want: %v", reflect.TypeOf(p), exptype)
	}
	fn.Call([]reflect.Value{reflect.ValueOf(p)})
	return hash, nil
}

func (test *udpTest) errorf(format string, args ...interface{}) error {
	_, file, line, ok := runtime.Caller(2) // errorf + waitPacketOut
	if ok {
		file = filepath.Base(file)
	} else {
		file = "???"
		line = 1
	}
	err := fmt.Errorf(format, args...)
	fmt.Printf("\t%s:%d: %v\n", file, line, err)
	test.t.Fail()
	return err
}

func TestUDP_packetErrors(t *testing.T) {
	test := newUDPTest(t)
	defer test.table.Close()

	test.packetIn(errExpired, pingPacket, &ping{From: testRemote, To: testLocalAnnounced, Version: 4})
	test.packetIn(errUnsolicitedReply, pongPacket, &pong{ReplyTok: []byte{}, Expiration: futureExp})
	test.packetIn(errUnknownNode, findnodePacket, &findnode{Expiration: futureExp})
	test.packetIn(errUnsolicitedReply, neighborsPacket, &neighbors{Expiration: futureExp})
}

func TestUDP_pingTimeout(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.table.Close()

	toaddr := &net.UDPAddr{IP: net.ParseIP("1.2.3.4"), Port: 2222}
	toid := enode.ID{1, 2, 3, 4}
	if err := test.udp.ping(toid, toaddr); err != errTimeout {
		t.Error("expected timeout error, got", err)
	}
}

func TestUDP_responseTimeouts(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.table.Close()

	rand.Seed(time.Now().UnixNano())
	randomDuration := func(max time.Duration) time.Duration {
		return time.Duration(rand.Int63n(int64(max)))
	}

	var (
		nReqs      = 200
		nTimeouts  = 0                       // number of requests with ptype > 128
		nilErr     = make(chan error, nReqs) // for requests that get a reply
		timeoutErr = make(chan error, nReqs) // for requests that time out
	)
	for i := 0; i < nReqs; i++ {
		// Create a matcher for a random request in udp.loop. Requests
		// with ptype <= 128 will not get a reply and should time out.
		// For all other requests, a reply is scheduled to arrive
		// within the timeout window.
		p := &pending{
			ptype:    byte(rand.Intn(255)),
			callback: func(interface{}) bool { return true },
		}
		binary.BigEndian.PutUint64(p.from[:], uint64(i))
		if p.ptype <= 128 {
			p.errc = timeoutErr
			test.udp.addpending <- p
			nTimeouts++
		} else {
			p.errc = nilErr
			test.udp.addpending <- p
			time.AfterFunc(randomDuration(60*time.Millisecond), func() {
				if !test.udp.handleReply(p.from, p.ptype, nil) {
					t.Logf("not matched: %v", p)
				}
			})
		}
		time.Sleep(randomDuration(30 * time.Millisecond))
	}

	// Check that all timeouts were delivered and that the rest got nil errors.
	// The replies must be delivered.
	var (
		recvDeadline        = time.After(20 * time.Second)
		nTimeoutsRecv, nNil = 0, 0
	)
	for i := 0; i < nReqs; i++ {
		select {
		case err := <-timeoutErr:
			if err != errTimeout {
				t.Fatalf("got non-timeout error on timeoutErr %d: %v", i, err)
			}
			nTimeoutsRecv++
		case err := <-nilErr:
			if err != nil {
				t.Fatalf("got non-nil error on nilErr %d: %v", i, err)
			}
			nNil++
		case <-recvDeadline:
			t.Fatalf("exceeded recv deadline")
		}
	}
	if nTimeoutsRecv != nTimeouts {
		t.Errorf("wrong number of timeout errors received: got %d, want %d", nTimeoutsRecv, nTimeouts)
	}
	if nNil != nReqs-nTimeouts {
		t.Errorf("wrong number of successful replies: got %d, want %d", nNil, nReqs-nTimeouts)
	}
}

func TestUDP_findnodeTimeout(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.table.Close()

	toaddr := &net.UDPAddr{IP: net.ParseIP("1.2.3.4"), Port: 2222}
	toid := enode.ID{1, 2, 3, 4}
	target := encPubkey{4, 5, 6, 7}
	result, err := test.udp.findnode(toid, toaddr, target)
	if err != errTimeout {
		t.Error("expected timeout error, got", err)
	}
	if len(result) > 0 {
		t.Error("expected empty result, got", result)
	}
}

func TestUDP_findnode(t *testing.T) {
	test := newUDPTest(t)
	defer test.table.Close()

	// put a few nodes into the table. their exact
	// distribution shouldn't matter much, although we need to
	// take care not to overflow any bucket.
	nodes := &nodesByDistance{target: testTarget.id()}
	for i := 0; i < bucketSize; i++ {
		key := newkey()
		n := wrapNode(enode.NewV4(&key.PublicKey, net.IP{10, 13, 0, 1}, 0, i))
		nodes.push(n, bucketSize)
	}
	test.table.stuff(nodes.entries)

	// ensure there's a bond with the test node,
	// findnode won't be accepted otherwise.
	remoteID := encodePubkey(&test.remotekey.PublicKey).id()
	test.table.db.UpdateLastPongReceived(remoteID, time.Now())

	// check that closest neighbors are returned.
	test.packetIn(nil, findnodePacket, &findnode{Target: testTarget, Expiration: futureExp})
	expected := test.table.closest(testTarget.id(), bucketSize)

	waitNeighbors := func(want []*node) {
		test.waitPacketOut(func(p *neighbors) {
			if len(p.Nodes) != len(want) {
				t.Errorf("wrong number of results: got %d, want %d", len(p.Nodes), bucketSize)
			}
			for i := range p.Nodes {
				if p.Nodes[i].ID.id() != want[i].ID() {
					t.Errorf("result mismatch at %d:\n  got:  %v\n  want: %v", i, p.Nodes[i], expected.entries[i])
				}
			}
		})
	}
	waitNeighbors(expected.entries[:maxNeighbors])
	waitNeighbors(expected.entries[maxNeighbors:])
}

func TestUDP_findnodeMultiReply(t *testing.T) {
	test := newUDPTest(t)
	defer test.table.Close()

	rid := enode.PubkeyToIDV4(&test.remotekey.PublicKey)
	test.table.db.UpdateLastPingReceived(rid, time.Now())

	// queue a pending findnode request
	resultc, errc := make(chan []*node), make(chan error)
	go func() {
		rid := encodePubkey(&test.remotekey.PublicKey).id()
		ns, err := test.udp.findnode(rid, test.remoteaddr, testTarget)
		if err != nil && len(ns) == 0 {
			errc <- err
		} else {
			resultc <- ns
		}
	}()

	// wait for the findnode to be sent.
	// after it is sent, the transport is waiting for a reply
	test.waitPacketOut(func(p *findnode) {
		if p.Target != testTarget {
			t.Errorf("wrong target: got %v, want %v", p.Target, testTarget)
		}
	})

	// send the reply as two packets.
	list := []*node{
		wrapNode(enode.MustParseV4("fnode://ba85011c70bcc5c04d8607d3a0ed29aa6179c092cbdda10d5d32684fb33ed01bd94f588ca8f91ac48318087dcb02eaf36773a7a453f0eedd6742af668097b29c@10.0.1.16:30303?discport=30304")),
		wrapNode(enode.MustParseV4("fnode://81fa361d25f157cd421c60dcc28d8dac5ef6a89476633339c5df30287474520caca09627da18543d9079b5b288698b542d56167aa5c09111e55acdbbdf2ef799@10.0.1.16:30303")),
		wrapNode(enode.MustParseV4("fnode://9bffefd833d53fac8e652415f4973bee289e8b1a5c6c4cbe70abf817ce8a64cee11b823b66a987f51aaa9fba0d6a91b3e6bf0d5a5d1042de8e9eeea057b217f8@10.0.1.36:30301?discport=17")),
		wrapNode(enode.MustParseV4("fnode://1b5b4aa662d7cb44a7221bfba67302590b643028197a7d5214790f3bac7aaa4a3241be9e83c09cf1f6c69d007c634faae3dc1b1221793e8446c0b3a09de65960@10.0.1.16:30303")),
	}
	rpclist := make([]rpcNode, len(list))
	for i := range list {
		rpclist[i] = nodeToRPC(list[i])
	}
	test.packetIn(nil, neighborsPacket, &neighbors{Expiration: futureExp, Nodes: rpclist[:2]})
	test.packetIn(nil, neighborsPacket, &neighbors{Expiration: futureExp, Nodes: rpclist[2:]})

	// check that the sent neighbors are all returned by findnode
	select {
	case result := <-resultc:
		want := append(list[:2], list[3:]...)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("neighbors mismatch:\n  got:  %v\n  want: %v", result, want)
		}
	case err := <-errc:
		t.Errorf("findnode error: %v", err)
	case <-time.After(5 * time.Second):
		t.Error("findnode did not return within 5 seconds")
	}
}

func TestUDP_successfulPing(t *testing.T) {
	test := newUDPTest(t)
	added := make(chan *node, 1)
	test.table.nodeAddedHook = func(n *node) { added <- n }
	defer test.table.Close()

	// The remote side sends a ping packet to initiate the exchange.
	go test.packetIn(nil, pingPacket, &ping{From: testRemote, To: testLocalAnnounced, Version: 4, Expiration: futureExp})

	// the ping is replied to.
	test.waitPacketOut(func(p *pong) {
		pinghash := test.sent[0][:macSize]
		if !bytes.Equal(p.ReplyTok, pinghash) {
			t.Errorf("got pong.ReplyTok %x, want %x", p.ReplyTok, pinghash)
		}
		wantTo := rpcEndpoint{
			// The mirrored UDP address is the UDP packet sender
			IP: test.remoteaddr.IP, UDP: uint16(test.remoteaddr.Port),
			// The mirrored TCP port is the one from the ping packet
			TCP:  testRemote.TCP,
			Rest: []rlp.RawValue{},
		}
		if !reflect.DeepEqual(p.To, wantTo) {
			t.Errorf("got pong.To %v, want %v", p.To, wantTo)
		}
	})

	// remote is unknown, the table pings back.
	hash, _ := test.waitPacketOut(func(p *ping) error {
		if !reflect.DeepEqual(p.From, test.udp.ourEndpoint) {
			t.Errorf("got ping.From %v, want %v", p.From, test.udp.ourEndpoint)
		}
		wantTo := rpcEndpoint{
			// The mirrored UDP address is the UDP packet sender.
			IP: test.remoteaddr.IP, UDP: uint16(test.remoteaddr.Port),
			TCP:  0,
			Rest: []rlp.RawValue{},
		}
		if !reflect.DeepEqual(p.To, wantTo) {
			t.Errorf("got ping.To %v, want %v", p.To, wantTo)
		}
		return nil
	})
	test.packetIn(nil, pongPacket, &pong{ReplyTok: hash, Expiration: futureExp})

	// the node should be added to the table shortly after getting the
	// pong packet.
	select {
	case n := <-added:
		rid := encodePubkey(&test.remotekey.PublicKey).id()
		if n.ID() != rid {
			t.Errorf("node has wrong ID: got %v, want %v", n.ID(), rid)
		}
		if !n.IP().Equal(test.remoteaddr.IP) {
			t.Errorf("node has wrong IP: got %v, want: %v", n.IP(), test.remoteaddr.IP)
		}
		if int(n.UDP()) != test.remoteaddr.Port {
			t.Errorf("node has wrong UDP port: got %v, want: %v", n.UDP(), test.remoteaddr.Port)
		}
		if n.TCP() != int(testRemote.TCP) {
			t.Errorf("node has wrong TCP port: got %v, want: %v", n.TCP(), testRemote.TCP)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("node was not added within 2 seconds")
	}
}

var testPackets = []struct {
	input      string
	wantPacket interface{}
}{
	{
		input: "94f7c87172c7ef7f71c449543a522d5b2d3256a193b95417968959067cda1fe500d9ddac5e0a85a4e99f5c9bc3bc3fd1789dea00c0a6f72d0aa09a79cc778a345021840b6f350a472c45b07cbd5fcf8c636af42b8b7158b77fb0b5df0894e4810001ed06821234cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a355",
		wantPacket: &ping{
			Version:    udpVersion,
			From:       rpcEndpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544, []rlp.RawValue{}},
			To:         rpcEndpoint{net.ParseIP("::1"), 2222, 3333, []rlp.RawValue{}},
			Expiration: 1136239445,
			MagicNetID: 0x1234,
			Rest:       []rlp.RawValue{},
		},
	},
	{
		input: "bebe0def283229d5a8b88953134817e3ddfedd6eb148d04417dea8fdd70cdb83e9a4c8522d801a3127871e84815f6c31cf5ccdef20cb966795a00f8fbbc61a8d128e4262117b15c14342fca26f72b9602ff5fe53a6a34178e0f2fbafc26cff2e0001ef06821234cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a3550102",
		wantPacket: &ping{
			Version:    udpVersion,
			From:       rpcEndpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544, []rlp.RawValue{}},
			To:         rpcEndpoint{net.ParseIP("::1"), 2222, 3333, []rlp.RawValue{}},
			Expiration: 1136239445,
			MagicNetID: 0x1234,
			Rest:       []rlp.RawValue{{0x01}, {0x02}},
		},
	},
	{
		input: "3e96daa63e38753050cab7779aa0606c49a56a787b52854c06307b5d010e7d4e6e69275aa91e799e8a9bf18d4cd7079cc9c205b862a69e3f05a0af72c48be7a3376a966b1c6cd0999931ea4cac67593eca949d97103e30d1edaf8fa4f080a9bc0001f83f06821234d79020010db83c4d001500000000abcdef12820cfa8215a8d79020010db885a308d313198a2e037073488208ae82823a8443b9a355c50102030405",
		wantPacket: &ping{
			Version:    udpVersion,
			From:       rpcEndpoint{net.ParseIP("2001:db8:3c4d:15::abcd:ef12"), 3322, 5544, []rlp.RawValue{}},
			To:         rpcEndpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338, []rlp.RawValue{}},
			Expiration: 1136239445,
			MagicNetID: 0x1234,
			Rest:       []rlp.RawValue{{0xC5, 0x01, 0x02, 0x03, 0x04, 0x05}},
		},
	},
	{
		input: "e30a117a239a06aa29dd61baeb7fbdf2607a47295b5a26858af230a2c346cf613fe8d2a84431fe33b9830b875f1280b91eda5d62581510b6fba2a5f1915fbb3a0a53cd4d691d6ace54e4197478062dac8427c50f9a95275bf272de6d1de611be0102f84a06821234d79020010db885a308d313198a2e037073488208ae82823aa0fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c9548443b9a355c6010203c2040506",
		wantPacket: &pong{
			To:         rpcEndpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338, []rlp.RawValue{}},
			ReplyTok:   common.Hex2Bytes("fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c954"),
			Expiration: 1136239445,
			MagicNetID: 0x1234,
			Version:    udpVersion,
			Rest:       []rlp.RawValue{{0xC6, 0x01, 0x02, 0x03, 0xC2, 0x04, 0x05}, {0x06}},
		},
	},
	{
		input: "bfd7f8984f00714e74747494edac06e5485fd5a40ea57308a50eb43d1824c6a30579947cea5a109828d4599cc2781f812e7e2e9701cada94f2d6315ae6e60e4b0071ac699a5f2574d60c48594bcdbe2b2ac2a0ca3359e38639decf71ebdb636f0103f85206821234b840ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd31387574077f301b421bc84df7266c44e9e6d569fc56be00812904767bf5ccd1fc7f8443b9a35582999983999999",
		wantPacket: &findnode{
			Target:     hexEncPubkey("ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd31387574077f301b421bc84df7266c44e9e6d569fc56be00812904767bf5ccd1fc7f"),
			Expiration: 1136239445,
			MagicNetID: 0x1234,
			Version:    udpVersion,
			Rest:       []rlp.RawValue{{0x82, 0x99, 0x99}, {0x83, 0x99, 0x99, 0x99}},
		},
	},
	{
		input: "f970e78c581aa9281f6adf63d353a8c74c236c72909f08d834693af72acd1a391f97ba570203f2345f70a7bc57cc9a24691280bbdfc69af6b37ed7a6393664486a07b8eaf074a64e0af33cb07cab03500e8747823d7cfdfcf0e1369b962cec2f0104f9015f06821234f90150f84d846321163782115c82115db8403155e1427f85f10a5c9a7755877748041af1bcd8d474ec065eb33df57a97babf54bfd2103575fa829115d224c523596b401065a97f74010610fce76382c0bf32f84984010203040101b840312c55512422cf9b8a4097e9a6ad79402e87a15ae909a4bfefa22398f03d20951933beea1e4dfa6f968212385e829f04c2d314fc2d4e255e0d3bc08792b069dbf8599020010db83c4d001500000000abcdef12820d05820d05b84038643200b172dcfef857492156971f0e6aa2c538d8b74010f8e140811d53b98c765dd2d96126051913f44582e8c199ad7c6d6819e9a56483f637feaac9448aacf8599020010db885a308d313198a2e037073488203e78203e8b8408dcab8618c3253b558d459da53bd8fa68935a719aff8b811197101a4b2b47dd2d47295286fc00cc081bb542d760717d1bdd6bec2c37cd72eca367d6dd3b9df738443b9a355010203",
		wantPacket: &neighbors{
			Nodes: []rpcNode{
				{
					ID:   hexEncPubkey("3155e1427f85f10a5c9a7755877748041af1bcd8d474ec065eb33df57a97babf54bfd2103575fa829115d224c523596b401065a97f74010610fce76382c0bf32"),
					IP:   net.ParseIP("99.33.22.55").To4(),
					UDP:  4444,
					TCP:  4445,
					Rest: []rlp.RawValue{},
				},
				{
					ID:   hexEncPubkey("312c55512422cf9b8a4097e9a6ad79402e87a15ae909a4bfefa22398f03d20951933beea1e4dfa6f968212385e829f04c2d314fc2d4e255e0d3bc08792b069db"),
					IP:   net.ParseIP("1.2.3.4").To4(),
					UDP:  1,
					TCP:  1,
					Rest: []rlp.RawValue{},
				},
				{
					ID:   hexEncPubkey("38643200b172dcfef857492156971f0e6aa2c538d8b74010f8e140811d53b98c765dd2d96126051913f44582e8c199ad7c6d6819e9a56483f637feaac9448aac"),
					IP:   net.ParseIP("2001:db8:3c4d:15::abcd:ef12"),
					UDP:  3333,
					TCP:  3333,
					Rest: []rlp.RawValue{},
				},
				{
					ID:   hexEncPubkey("8dcab8618c3253b558d459da53bd8fa68935a719aff8b811197101a4b2b47dd2d47295286fc00cc081bb542d760717d1bdd6bec2c37cd72eca367d6dd3b9df73"),
					IP:   net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"),
					UDP:  999,
					TCP:  1000,
					Rest: []rlp.RawValue{},
				},
			},
			Expiration: 1136239445,
			MagicNetID: 0x1234,
			Version:    udpVersion,
			Rest:       []rlp.RawValue{{0x01}, {0x02}, {0x03}},
		},
	},
}

func TestEncodePacket(t *testing.T) {
	testkey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	var typecode byte
	for i, test := range testPackets {
		switch test.wantPacket.(type) {
		case *ping:
			typecode = pingPacket
		case *pong:
			typecode = pongPacket
		case *findnode:
			typecode = findnodePacket
		case *neighbors:
			typecode = neighborsPacket
		}
		packet, hash, err := encodePacket(testkey, typecode, test.wantPacket)
		if err != nil {
			t.Error("encodePacket failure:", err)
		}
		inhex, _ := hex.DecodeString(test.input)
		if !reflect.DeepEqual(packet, inhex) {
			t.Errorf("%d want:'%x'\ngot:'%x',\nhash:'%x'", i, inhex, packet, hash)
		}
	}
}

func TestForwardCompatibility(t *testing.T) {
	testkey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	wantNodeKey := encodePubkey(&testkey.PublicKey)

	for _, test := range testPackets {
		input, err := hex.DecodeString(test.input)
		if err != nil {
			t.Fatalf("invalid hex: %s", test.input)
		}
		packet, nodekey, _, err := decodePacket(input)
		if err != nil {
			t.Errorf("did not accept packet %s\n%v", test.input, err)
			continue
		}
		if !reflect.DeepEqual(packet, test.wantPacket) {
			t.Errorf("got %s\nwant %s", spew.Sdump(packet), spew.Sdump(test.wantPacket))
		}
		if nodekey != wantNodeKey {
			t.Errorf("got id %v\nwant id %v", nodekey, wantNodeKey)
		}
	}
}

// dgramPipe is a fake UDP socket. It queues all sent datagrams.
type dgramPipe struct {
	mu      *sync.Mutex
	cond    *sync.Cond
	closing chan struct{}
	closed  bool
	queue   [][]byte
}

func newpipe() *dgramPipe {
	mu := new(sync.Mutex)
	return &dgramPipe{
		closing: make(chan struct{}),
		cond:    &sync.Cond{L: mu},
		mu:      mu,
	}
}

// WriteToUDP queues a datagram.
func (c *dgramPipe) WriteToUDP(b []byte, to *net.UDPAddr) (n int, err error) {
	msg := make([]byte, len(b))
	copy(msg, b)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, errors.New("closed")
	}
	c.queue = append(c.queue, msg)
	c.cond.Signal()
	return len(b), nil
}

// ReadFromUDP just hangs until the pipe is closed.
func (c *dgramPipe) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	<-c.closing
	return 0, nil, io.EOF
}

func (c *dgramPipe) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		close(c.closing)
		c.closed = true
	}
	return nil
}

func (c *dgramPipe) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: testLocal.IP, Port: int(testLocal.UDP)}
}

func (c *dgramPipe) waitPacketOut() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	for len(c.queue) == 0 {
		c.cond.Wait()
	}
	p := c.queue[0]
	copy(c.queue, c.queue[1:])
	c.queue = c.queue[:len(c.queue)-1]
	return p
}
