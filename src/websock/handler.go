package websock

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/statuspage"
)

// Handler is a HTTP handler for WebSocket connections.
type Handler struct {
	Upgrader websocket.Upgrader
}

// NewHandler returns an HTTP handler for WebSocket connections.
func NewHandler() *Handler {
	h := &Handler{}

	h.Upgrader.Error = func(w http.ResponseWriter, r *http.Request, c int, err error) {
		statuspage.WriteMessage(w, r, c, err.Error())
	}

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	con, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer con.Close()

	for {
		mt, message, err := con.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = con.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
