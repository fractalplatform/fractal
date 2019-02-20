package protoadaptor

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/log"
	router "github.com/fractalplatform/fractal/event"
	"github.com/fractalplatform/fractal/p2p"
	"github.com/fractalplatform/fractal/utils/rlp"
)

type pack struct {
	From     string
	To       string
	Typecode uint32
	Payload  []byte
}

type remotePeer struct {
	peer *p2p.Peer
	ws   p2p.MsgReadWriter
}

// ProtoAdaptor is subprotocol on p2p
type ProtoAdaptor struct {
	p2p.Server
	peerMangaer
	event   chan *router.Event
	station router.Station
}

// NewProtoAdaptor return new ProtoAdaptor
func NewProtoAdaptor(config *p2p.Config) *ProtoAdaptor {
	adaptor := &ProtoAdaptor{
		Server: p2p.Server{
			Config: config,
		},
		peerMangaer: peerMangaer{
			activePeers: make(map[[8]byte]*remotePeer),
			station:     nil,
		},
		event:   make(chan *router.Event),
		station: router.NewLocalStation("p2p", nil),
	}
	adaptor.peerMangaer.station = router.NewBroadcastStation("broadcast", &adaptor.peerMangaer)
	adaptor.Server.Config.Protocols = adaptor.Protocols()
	return adaptor
}

// Start start p2p protocol adaptor
func (adaptor *ProtoAdaptor) Start() error {
	router.StationRegister(adaptor.peerMangaer.station)
	router.AdaptorRegister(adaptor)
	router.Subscribe(nil, adaptor.event, router.P2pDisconectPeer, nil)
	go adaptor.adaptorEvent()
	return adaptor.Server.Start()
}

func (adaptor *ProtoAdaptor) adaptorEvent() {
	for {
		e := <-adaptor.event
		switch e.Typecode {
		case router.P2pDisconectPeer:
			peer := e.Data.(router.Station).Data().(*remotePeer)
			peer.peer.Disconnect(p2p.DiscSubprotocolError)
			//peer.Disconnect(DiscSubprotocolError)
		}
	}
}

func (adaptor *ProtoAdaptor) adaptorLoop(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	remote := remotePeer{ws: ws, peer: peer}
	station := router.NewRemoteStation(string(remote.peer.ID().Bytes()[:8]), &remote)
	adaptor.peerMangaer.addActivePeer(&remote)
	router.StationRegister(station)
	url := remote.peer.Node().String()
	router.SendTo(station, nil, router.P2pNewPeer, &url)
	defer func() {
		adaptor.peerMangaer.delActivePeer(&remote)
		router.StationUnregister(station)
		url := remote.peer.Node().String()
		router.SendTo(station, nil, router.P2pDelPeer, &url)
	}()

	for {
		msg, err := ws.ReadMsg()
		if err != nil {
			return err
		}
		pack := pack{}
		if err := msg.Decode(&pack); err != nil {
			return err
		}
		e, err := pack2event(&pack, station)
		if err != nil {
			return err
		}
		go router.SendEvent(e)
	}
}

// Protocols .
func (adaptor *ProtoAdaptor) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		p2p.Protocol{
			Name:    "FractalTest",
			Version: 1,
			Length:  1,
			Run:     adaptor.adaptorLoop,
		},
	}
}

// Stop .
func (adaptor *ProtoAdaptor) Stop() {
	adaptor.Server.Stop()
	log.Info("P2P networking stopped")
}

// SendOut .
func (adaptor *ProtoAdaptor) SendOut(e *router.Event) error {
	if e.To.IsBroadcast() {
		adaptor.msgBroadcast(e)
		return nil
	}
	return adaptor.msgSend(e)
}

func (adaptor *ProtoAdaptor) msgSend(e *router.Event) error {
	pack, err := event2pack(e)
	if err != nil {
		return err
	}
	return p2p.Send(e.To.Data().(*remotePeer).ws, 0, pack)
}

func (adaptor *ProtoAdaptor) msgBroadcast(e *router.Event) {
	te := *e
	te.To = nil
	pack, err := event2pack(&te)
	if err != nil {
		return
	}

	send := func(peer *remotePeer) {
		p2p.Send(peer.ws, 0, pack)
	}
	if e.To.Data() != nil {
		pack.To = "" // if sendto 'broadcast' station, remote will broadcast again, and dead loop (-_-)
		e.To.Data().(*peerMangaer).mapActivePeer(send)
		return
	}
	adaptor.peerMangaer.mapActivePeer(send)
}

func event2pack(e *router.Event) (*pack, error) {
	buf, err := rlp.EncodeToBytes(e.Data)
	if err != nil {
		return nil, err
	}
	from := ""
	if e.From != nil {
		from = e.From.Name()
	}
	to := ""
	if e.To != nil {
		to = e.To.Name()[8:]
	}
	return &pack{
		From:     from,
		To:       to,
		Typecode: uint32(e.Typecode),
		Payload:  buf,
	}, nil
}

func pack2event(pack *pack, station router.Station) (*router.Event, error) {
	var elem interface{}

	isPtr := false
	typ := router.GetTypeByCode(int(pack.Typecode))
	if typ == nil {
		return nil, errors.New(fmt.Sprint("unknow typecode:", pack.Typecode))
	} else {
		//for typ.Kind() == reflect.Ptr {
		if typ.Kind() == reflect.Ptr {
			isPtr = true
			typ = typ.Elem()
		}
		obj := reflect.New(typ)
		if err := rlp.DecodeBytes(pack.Payload, obj.Interface()); err != nil {
			return nil, err
		}
		if isPtr {
			elem = obj.Interface()
		} else {
			elem = obj.Elem().Interface()
		}
	}
	if pack.From != "" {
		station = router.NewRemoteStation(station.Name()+pack.From, station.Data())
	}
	return &router.Event{
		From:     station,
		To:       router.GetStationByName(pack.To),
		Typecode: int(pack.Typecode),
		Data:     elem,
	}, nil
}
