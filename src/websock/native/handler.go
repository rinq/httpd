package native

import (
	"context"
	"net/http"

	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

const protocolPrefix = "rinq-1.0+"

var _ websock.Handler = (*Handler)(nil)

// NewHandler returns a websock.Handler that implements Rinq's native websocket protocol.
// The native protocol allows for several different frame-level encodings. Each Handler
// instance supports one of these encodings, the choice of which affects the websocket
// sub-protocol that the Handler advertises itself as supporting.
func NewHandler(peer rinq.Peer, encoding message.Encoding, logger websock.Logger, options ...Option) websock.Handler {

	return &Handler{
		Peer:       peer,
		Encoding:   encoding,
		visitorOpt: options,
		Logger:     logger,
	}
}

// Handler is an implementation of websock.Handler that handles connections that
// use Rinq's "native" subprotocol.
type Handler struct {
	Peer     rinq.Peer
	Encoding message.Encoding
	Logger   websock.Logger

	visitorOpt []Option
}

// Protocol returns the name of the WebSocket sub-protocol supported by this
// Handler.
func (h *Handler) Protocol() string {
	return protocolPrefix + h.Encoding.Name()
}

// Handle takes control of WebSocket connection until it is closed.
func (h *Handler) Handle(c websock.Connection, r *http.Request, attrs map[string][]rinq.Attr) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	attrs[HttpdAttrNamespace] = sessionAttributes(r)

	v := newVisitor(
		ctx,
		h.Peer,
		attrs,
		func(m message.Outgoing) {
			if w, err := c.NextWriter(); err == nil {
				defer w.Close()
				_ = message.Write(w, h.Encoding, m)
			}
		},
		c,
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
