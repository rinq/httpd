package native

import (
	"fmt"
	"log"

	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

const protocolPrefix = "rinq-1.0+"

// Handler is an implementation of websock.Handler that handles connections that
// use Rinq's "native" subprotocol.
type Handler struct {
	Encoding message.Encoding
	Logger   *log.Logger
}

// Protocol returns the name of the WebSocket sub-protocol supported by this
// handler.
func (h *Handler) Protocol() string {
	return protocolPrefix + h.Encoding.Name()
}

// Handle takes control of WebSocket connection until it is closed.
func (h *Handler) Handle(s websock.Socket, p rinq.Peer, c websock.Config) {
	io := newIO(s, h.Encoding, c.PingInterval)
	defer io.Stop()

	con := newConnection(p, io.Send, h.Logger)
	defer con.Close()

	for msg := range io.Messages() {
		if err := msg.Accept(con); err != nil {
			fmt.Println(err) // TODO: log
			return
		}
	}

	if err := io.Err(); err != nil {
		fmt.Println(err) // TODO: log
		return
	}
}
