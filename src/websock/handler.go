package websock

import (
	"net/http"
	"time"

	"github.com/alecthomas/units"
	"github.com/gorilla/websocket"
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/httpd/src/internal/statuspage"
	"github.com/rinq/rinq-go/src/rinq"
	"golang.org/x/sync/semaphore"
)

// Handler is an interface that handles connections for one or more
// WebSocket sub-protocols.
type Handler interface {
	// Protocol returns the name of the WebSocket sub-protocol supported by this
	// handler.
	Protocol() string

	// Handle takes control of a WebSocket connection until it is closed.
	// It takes a map of namespace to rinq.Attr to apply to any sessions
	// created on this connection
	Handle(Connection, *http.Request, map[string][]rinq.Attr) error
}

// Logger defines what the HTTPHandler expects to be able to log to
type Logger interface {
	Log(fmt string, args ...interface{})
	Debug(fmt string, v ...interface{})
}

// httpHandler is an http.Handler that negotiates a WebSocket upgrade and
// dispatches handling to the appropriate sub-protocol.
type httpHandler struct {
	pingInterval       time.Duration
	maxIncomingMsgSize units.MetricBytes
	logger             Logger
	defaultHandler     Handler
	handlers           map[string]Handler
	upgrader           websocket.Upgrader

	globalLimit     *semaphore.Weighted
	maxCallsPerConn int64
}

// NewHTTPHandler returns an HTTP handler for a set of WebSocket handlers.
func NewHTTPHandler(
	handlers []Handler,
	opts ...Option,
) http.Handler {
	// set up the defaults
	h := &httpHandler{
		pingInterval: 10 * time.Second,
		logger:       twelf.DefaultLogger,
		handlers:     map[string]Handler{},
	}

	h.globalLimit = semaphore.NewWeighted(10000)
	h.maxCallsPerConn = 100

	h.upgrader = websocket.Upgrader{
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

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Log("upgrade error:", err)
		return
	}
	defer socket.Close()

	wsh, ok := h.handlers[socket.Subprotocol()]
	if !ok {
		wsh = h.defaultHandler
	}

	if wsh == nil {
		closeGracefully(socket, h)
		return
	}

	socket.SetReadLimit(int64(h.maxIncomingMsgSize))

	connLimit := semaphore.NewWeighted(h.maxCallsPerConn)
	conn := newConn(socket, h.pingInterval, h.globalLimit, connLimit)

	err = wsh.Handle(conn, r, make(map[string][]rinq.Attr))
	if err != nil {
		h.logger.Log("handler error:", err) // TODO: log
	}
}

func closeGracefully(socket *websocket.Conn, h *httpHandler) {
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
	h.logger.Log("unsupported sub-protocol") // TODO: log, pull from headers
}
