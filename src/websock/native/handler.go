package native

import (
	"context"
	"log"
	"net/http"

	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

const protocolPrefix = "rinq-1.0+"

// NewHandler returns a handler that takes control of websockets passed into the Handle
// method and interprets all frames passed through as rinq messages encoded using
// the given encoding
func NewHandler(peer rinq.Peer, encoding message.Encoding, options ...Option) Handler {

	return Handler{
		Peer:       peer,
		Encoding:   encoding,
		visitorOpt: options,
	}
}

// Handler is an implementation of websock.Handler that handles connections that
// use Rinq's "native" subprotocol.
type Handler struct {
	Peer     rinq.Peer
	Encoding message.Encoding
	Logger   *log.Logger

	visitorOpt []Option
}

// Protocol returns the name of the WebSocket sub-protocol supported by this
// handler.
func (h *Handler) Protocol() string {
	return protocolPrefix + h.Encoding.Name()
}

// Handle takes control of WebSocket connection until it is closed.
func (h *Handler) Handle(c websock.Connection, r *http.Request) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	v := newVisitor(
		ctx,
		h.Peer,
		sessionAttributes(r),
		func(m message.Outgoing) {
			if w, err := c.NextWriter(); err == nil {
				defer w.Close()
				_ = message.Write(w, h.Encoding, m)
			}
		},
	)

	for _, opt := range h.visitorOpt {
		opt.modify(v)
	}

	for {
		r, err := c.NextReader()
		if err != nil {
			return err
		}

		msg, err := message.Read(r, h.Encoding)
		if err != nil {
			return err
		}

		err = msg.Accept(v)
		if err != nil {
			return err
		}
	}
}
