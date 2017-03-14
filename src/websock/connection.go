package websock

import (
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection is an interface for performing IO on a WebSocket connection.
type Connection interface {
	NextReader() (io.Reader, error)
	NextWriter() (io.WriteCloser, error)
}

type connection struct {
	socket       *websocket.Conn
	pingInterval time.Duration
	closeReply   func(int, string) error

	done  chan struct{}
	mutex sync.Mutex // write mutex
}

func newConn(socket *websocket.Conn, pingInterval time.Duration) *connection {
	c := &connection{
		socket:       socket,
		pingInterval: pingInterval,
		closeReply:   socket.CloseHandler(),
		done:         make(chan struct{}),
	}

	socket.SetPongHandler(c.pong)
	socket.SetCloseHandler(c.close)

	go c.pingLoop()

	return c
}

func (c *connection) NextReader() (io.Reader, error) {
	_, r, err := c.socket.NextReader()
	return r, err
}

func (c *connection) NextWriter() (io.WriteCloser, error) {
	c.mutex.Lock()

	w, err := c.socket.NextWriter(websocket.BinaryMessage)
	if err != nil {
		c.mutex.Unlock()
		return nil, err
	}

	return writeCloser{w, &c.mutex}, nil
}

func (c *connection) close(code int, text string) error {
	close(c.done)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.closeReply(code, text)
}

func (c *connection) pingLoop() {
	ping := time.NewTicker(c.pingInterval)
	defer ping.Stop()

	for {
		select {
		case <-ping.C:
			c.ping()
		case <-c.done:
			return
		}
	}
}

func (c *connection) ping() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_ = c.socket.WriteMessage(websocket.PingMessage, nil)
}

func (c *connection) pong(string) error {
	deadline := time.Now().Add(c.pingInterval * 2)
	return c.socket.SetReadDeadline(deadline)
}

type writeCloser struct {
	io.WriteCloser
	mutex *sync.Mutex
}

func (w writeCloser) Close() error {
	defer w.mutex.Unlock()
	return w.WriteCloser.Close()
}
