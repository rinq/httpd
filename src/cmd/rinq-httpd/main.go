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
	"github.com/rinq/rinq-go/src/rinq/amqp"
)

func main() {
	// TODO: this env var will be handled by rinq-go
	// https://github.com/rinq/rinq-go/issues/94
	peer, err := amqp.Dial(os.Getenv("RING_AMQP_DSN"))
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	ws := websocketHandler(peer, logger)

	err = http.ListenAndServe(
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
			native.NewProtocol(
				peer,
				pingInterval(),
			),
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
