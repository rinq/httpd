package main

import (
	"fmt"
	"log"
	"math/rand"
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
	rand.Seed(time.Now().UnixNano())

	logger := log.New(os.Stdout, "", log.LstdFlags)
	var ws http.Handler

	server := &http.Server{
		Addr: os.Getenv("RINQ_BIND"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if websocket.IsWebSocketUpgrade(r) {
				ws.ServeHTTP(w, r)
			} else {
				statuspage.Write(w, r, http.StatusUpgradeRequired)
			}
		}),
	}

	for {
		peer := connect()
		ws = websocketHandler(peer, logger)

		done := make(chan error, 1)
		go serve(server, done)

		select {
		case <-peer.Done():
			if err := peer.Err(); err != nil {
				// TODO: log
				fmt.Println(err)
			}
			server.Close()

		case err := <-done:
			if err != nil {
				// TODO: log
				fmt.Println(err)
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func connect() rinq.Peer {
	for {
		// TODO: this env var will be handled by rinq-go
		// https://github.com/rinq/rinq-go/issues/94
		peer, err := amqp.Dial(os.Getenv("RING_AMQP_DSN"))
		if err == nil {
			return peer
		}

		fmt.Println(err) // TODO: log
		time.Sleep(3 * time.Second)
	}
}

func serve(server *http.Server, c chan<- error) {
	if err := server.ListenAndServe(); err != nil {
		c <- err
	}

	close(c)
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
