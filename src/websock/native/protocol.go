package native

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
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
	con := newConnection(
		p.peer,
		p.ping,
		s,
		p.encoding(s),
	)

	if err := con.Run(); err != nil {
		fmt.Println("connection error", err) // TODO: log
	}
}

func (p *protocol) encoding(s *websocket.Conn) message.Encoding {
	switch s.Subprotocol() {
	case protocolPrefix + "cbor":
		return message.CBOREncoding
	case protocolPrefix + "json":
		return message.JSONEncoding
	default:
		panic("selected sub-protocol is not supported")
	}
}
