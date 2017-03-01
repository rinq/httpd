package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/statuspage"
)

var upgrader websocket.Upgrader

func main() {
	if origin := os.Getenv("RINQ_ORIGIN"); origin != "" {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			if origin == "*" {
				return true
			}

			o := r.Header["Origin"]
			if len(o) == 0 {
				return true
			}

			u, err := url.Parse(o[0])
			if err != nil {
				return false
			}

			return strings.EqualFold(u.Host, origin)
		}
	}

	upgrader.Error = func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		statuspage.WriteMessage(w, r, status, reason.Error())
	}

	http.ListenAndServe(
		os.Getenv("RINQ_BIND"),
		http.HandlerFunc(handler),
	)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		statuspage.Write(w, r, http.StatusUpgradeRequired)
		return
	}

	con, err := upgrader.Upgrade(w, r, nil)
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
