package native

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/ugorji/go/codec"
)

const protocolPrefix = "ring-1.0+"

// protocol is an implementation of websock.Protocol that provides the "native"
// protocol implementation.
type protocol struct {
	peer rinq.Peer
	ping time.Duration
}

// NewProtocol returns a new websock.Protocol for the "native" Rinq protocol.
func NewProtocol(peer rinq.Peer, ping time.Duration) websock.Protocol {
	return &protocol{peer, ping}
}

func (p *protocol) Names() []string {
	return []string{
		protocolPrefix + "cbor",
		protocolPrefix + "json",
	}
}

func (p *protocol) Handle(s *websocket.Conn) {
	enc, dec := p.codec(s)

	con := newConnection(
		p.peer,
		p.ping,
		s,
		enc,
		dec,
	)

	if err := con.Run(); err != nil {
		fmt.Println("connection error", err) // TODO: log
	}
}

func (p *protocol) codec(s *websocket.Conn) (enc *codec.Encoder, dec *codec.Decoder) {
	var handle codec.Handle

	switch s.Subprotocol() {
	case protocolPrefix + "cbor":
		handle = &cborHandle
	case protocolPrefix + "json":
		handle = &jsonHandle
	default:
		panic("selected sub-protocol is not supported")
	}

	enc = codec.NewEncoder(nil, handle)
	dec = codec.NewDecoder(nil, handle)
	return
}

var cborHandle codec.CborHandle
var jsonHandle codec.JsonHandle
