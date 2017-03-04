package websock

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/websock/protocol"
)

// upgrader is a function that upgrades an HTTP request to a WebSocket
// connection.
type upgrader func(http.ResponseWriter, *http.Request) (*websocket.Conn, int, error)

// newUpgrader returns an upgrader the matches the origin pattern o, and
// supports the Rinq protocols in p.
func newUpgrader(o string, p *protocol.Set) upgrader {
	var code int
	h := func(_ http.ResponseWriter, _ *http.Request, c int, _ error) {
		code = c
	}

	u := websocket.Upgrader{
		CheckOrigin:  newOriginChecker(o),
		Error:        h,
		Subprotocols: p.Names(),
	}

	return func(w http.ResponseWriter, r *http.Request) (*websocket.Conn, int, error) {
		con, err := u.Upgrade(w, r, nil)
		return con, code, err
	}
}
