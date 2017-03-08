package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/internal/statuspage"
	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native"
	"github.com/rinq/rinq-go/src/rinq"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	ws := websocketHandler(nil, logger) // TODO: initialize peer

	err := http.ListenAndServe(
		os.Getenv("RINQ_BIND"),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if websocket.IsWebSocketUpgrade(r) {
				ws.ServeHTTP(w, r)
			} else {
				statuspage.Write(w, r, http.StatusUpgradeRequired)
			}
		}),
	)
	if err != nil {
		panic(err)
	}
}

func websocketHandler(peer rinq.Peer, logger *log.Logger) http.Handler {
	return websock.NewHandler(
		os.Getenv("RINQ_ORIGIN"),
		websock.NewProtocolSet(
			native.NewProtocol(nil), // TODO handler
		),
		logger,
	)
}

func pingInterval() time.Duration {
	i, err := strconv.ParseUint(os.Getenv("RINQ_PING"), 10, 64)
	if err != nil {
		return 10 * time.Second
	}

	return time.Duration(i) * time.Second
}
