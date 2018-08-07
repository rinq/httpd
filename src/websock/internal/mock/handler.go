package mock

import (
	"net/http"

	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/rinq-go/src/rinq"
)

// Handler is a mock websock.Handler
type Handler struct {
	Impl struct {
		Protocol string
		IsBinary bool
		Handle   func(websock.Connection, *http.Request, map[string][]rinq.Attr) error
	}
}

// Protocol returns h.Impl.Protocol
func (h *Handler) Protocol() string {
	return h.Impl.Protocol
}

func (h *Handler) IsBinary() bool {
	return h.Impl.IsBinary
}

// Handle forwards to h.Impl.Handle
func (h *Handler) Handle(c websock.Connection, r *http.Request, attrs map[string][]rinq.Attr) error {
	if h.Impl.Handle != nil {
		return h.Impl.Handle(c, r, attrs)
	}

	return nil
}
