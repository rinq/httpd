package mock

import (
	"net/http"

	"github.com/rinq/httpd/src/websock"
)

// Handler is a mock websock.Handler
type Handler struct {
	Impl struct {
		Protocol string
		Handle   func(websock.Connection, *http.Request) error
	}
}

// Protocol returns h.Impl.Protocol
func (h *Handler) Protocol() string {
	return h.Impl.Protocol
}

// Handle forwards to h.Impl.Handle
func (h *Handler) Handle(c websock.Connection, r *http.Request) error {
	if h.Impl.Handle != nil {
		return h.Impl.Handle(c, r)
	}

	return nil
}
