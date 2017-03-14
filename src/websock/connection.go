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

	mutex sync.Mutex // write mutex
	done  chan struct{}
}

func newConn(socket *websocket.Conn, pingInterval time.Duration) *connection {
	c := &connection{
		socket:       socket,
		pingInterval: pingInterval,
		done:         make(chan struct{}),
	}

	socket.SetPongHandler(c.pong)

	go c.pingLoop()

	return c
}

func (c *connection) NextReader() (io.Reader, error) {
	_, r, err := c.socket.NextReader()

	if _, ok := err.(*websocket.CloseError); ok {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		select {
		case <-c.done:
		default:
			close(c.done)
		}
	}

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
