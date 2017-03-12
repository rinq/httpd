package websock

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/internal/statuspage"
	"github.com/rinq/rinq-go/src/rinq"
)

// Handler is an interface that handles connections for one or more
// WebSocket sub-protocols.
type Handler interface {
	// Protocol returns the name of the WebSocket sub-protocol supported by this
	// handler.
	Protocol() string

	// Handle takes control of a WebSocket connection until it is closed.
	Handle(Socket, rinq.Peer, Config)
}

// httpHandler is an http.Handler that negotiates a WebSocket upgrade and
// dispatches handling to the appropriate sub-protocol.
type httpHandler struct {
	getPeer  func() (rinq.Peer, bool)
	config   Config
	logger   *log.Logger
	handlers map[string]Handler
	upgrader websocket.Upgrader
}

// NewHTTPHandler returns an HTTP handler for a set of WebSocket handlers.
func NewHTTPHandler(
	getPeer func() (rinq.Peer, bool),
	config Config,
	logger *log.Logger,
	handlers ...Handler,
) http.Handler {
	h := &httpHandler{
		getPeer:  getPeer,
		config:   config,
		logger:   logger,
		handlers: map[string]Handler{},
	}

	h.upgrader = websocket.Upgrader{
		CheckOrigin:       newOriginChecker(config.OriginPattern),
		EnableCompression: true,
		Error: func(w http.ResponseWriter, r *http.Request, c int, _ error) {
			statuspage.Write(w, r, c)
		},
	}

	for _, wsh := range handlers {
		p := wsh.Protocol()
		if _, ok := h.handlers[p]; !ok {
			h.handlers[p] = wsh
			h.upgrader.Subprotocols = append(h.upgrader.Subprotocols, p)
		}
	}

	return h
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer, ok := h.getPeer()

	if !ok {
		statuspage.Write(w, r, http.StatusServiceUnavailable)
		fmt.Println("peer not available") // TODO: log
		return
	}

	socket, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade error:", err) // TODO: log
		return
	}
	defer socket.Close()

	if wsh, ok := h.handlers[socket.Subprotocol()]; ok {
		wsh.Handle(socket, peer, h.config)
	} else {
		fmt.Println("unsupported sub-protocol") // TODO: log, pull from headers
	}
}
