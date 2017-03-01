package main

import (
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/statuspage"
	"github.com/rinq/httpd/src/websock"
)

func main() {
	ws := websock.NewHandler()
	ws.Upgrader.CheckOrigin = websock.NewOriginChecker(os.Getenv("RINQ_ORIGIN"))

	http.ListenAndServe(
		os.Getenv("RINQ_BIND"),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if websocket.IsWebSocketUpgrade(r) {
				ws.ServeHTTP(w, r)
			} else {
				statuspage.Write(w, r, http.StatusUpgradeRequired)
			}
		}),
	)
}
