package websock

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/internal/statuspage"
)

// Handler is an interface that handles connections for one or more
// WebSocket sub-protocols.
type Handler interface {
	// Protocol returns the name of the WebSocket sub-protocol supported by this
	// handler.
	Protocol() string

	// Handle takes control of a WebSocket connection until it is closed.
	Handle(Connection, *http.Request) error
}

// httpHandler is an http.Handler that negotiates a WebSocket upgrade and
// dispatches handling to the appropriate sub-protocol.
type httpHandler struct {
	pingInterval time.Duration
	logger       *log.Logger
	handlers     map[string]Handler
	upgrader     websocket.Upgrader
}

// NewHTTPHandler returns an HTTP handler for a set of WebSocket handlers.
func NewHTTPHandler(
	originPattern string,
	pingInterval time.Duration,
	logger *log.Logger,
	handlers ...Handler,
) http.Handler {
	h := &httpHandler{
		pingInterval: pingInterval,
		logger:       logger,
		handlers:     map[string]Handler{},
	}

	h.upgrader = websocket.Upgrader{
		CheckOrigin:       newOriginChecker(originPattern),
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
	socket, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade error:", err) // TODO: log
		return
	}
	defer socket.Close()

	wsh, ok := h.handlers[socket.Subprotocol()]
	if !ok {
		// Write a close message for those clients that don't automatically
		// disconnect after a failed sub-protocol negotiation.
		_ = socket.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseProtocolError,
				"unsupported sub-protocol",
			),
			time.Now().Add(time.Second),
		)
		fmt.Println("unsupported sub-protocol") // TODO: log, pull from headers
		return
	}

	conn := newConn(socket, h.pingInterval)

	err = wsh.Handle(conn, r)
	if err != nil {
		fmt.Println("handler error:", err) // TODO: log
		return
	}
}
