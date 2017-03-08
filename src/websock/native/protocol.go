package native

import (
	"fmt"
	"time"

	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

const protocolPrefix = "rinq-1.0+"

// protocol is an implementation of websock.Protocol that provides the "native"
// protocol implementation.
type protocol struct {
	handle handler
}

type handler func(websock.Socket, message.Encoding) error

// NewProtocol returns a new websock.Protocol for the "native" Rinq protocol.
func NewProtocol(
	peer rinq.Peer,
	ping time.Duration,
) websock.Protocol {
	h := func(s websock.Socket, e message.Encoding) error {
		con := newConnection(peer, ping, s, e)
		return con.Run()
	}

	return &protocol{h}
}

func (p *protocol) Names() []string {
	return []string{
		protocolPrefix + "cbor",
		protocolPrefix + "json",
	}
}

func (p *protocol) Handle(s websock.Socket) {
	e := p.encoding(s)
	err := p.handle(s, e)
	fmt.Println(err) // TODO: log
}

func (p *protocol) encoding(s websock.Socket) message.Encoding {
	switch s.Subprotocol() {
	case protocolPrefix + "cbor":
		return message.CBOREncoding
	case protocolPrefix + "json":
		return message.JSONEncoding
	default:
		panic("selected sub-protocol is not supported")
	}
}
