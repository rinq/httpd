package websock

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/internal/statuspage"
)

// Handler is an http.Handler that negotiates a WebSocket upgrade and
// dispatches handling to the appropriate sub-protocol.
type Handler struct {
	protos   *ProtocolSet
	logger   *log.Logger
	upgrader websocket.Upgrader
}

// NewHandler returns an HTTP handler for WebSocket connections.
func NewHandler(
	origin string,
	protos *ProtocolSet,
	logger *log.Logger,
) *Handler {
	h := &Handler{
		protos: protos,
		logger: logger,
		upgrader: websocket.Upgrader{
			Subprotocols:      protos.Names(),
			CheckOrigin:       newOriginChecker(origin),
			EnableCompression: true,
			Error: func(w http.ResponseWriter, r *http.Request, c int, err error) {
				if _, err := statuspage.Write(w, r, c); err != nil {
					fmt.Println(err) // TODO: log
				}
			},
		},
	}

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO: log
		fmt.Println("upgrade error:", err)
		return
	}
	defer socket.Close()

	if proto, ok := h.protos.Select(socket.Subprotocol()); ok {
		proto.Handle(socket)
	} else {
		// TODO: log
		fmt.Println("unsupported sub-protocol")
	}
}
