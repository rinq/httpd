package websock

import (
	"io"
	"sync"
	"time"

	"context"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/semaphore"
)

// Capacity represents the resources of the given server - in this case,
// maximum amount of stateful calls permissible at once per connection
// and per server
type Capacity interface {
	// ReserveCapacity either reserves capacity on the server or returns an error
	ReserveCapacity(context.Context) error
	// ReleaseCapacity releases capacity to the server
	ReleaseCapacity()
}

// Connection is an interface for performing IO on a WebSocket connection.
type Connection interface {
	NextReader() (io.Reader, error)
	NextWriter() (io.WriteCloser, error)

	Capacity
}

type connection struct {
	socket      *websocket.Conn
	messageType int

	pingInterval time.Duration
	closeReply   func(int, string) error

	globalCap *semaphore.Weighted
	localCap  *semaphore.Weighted

	mutex sync.Mutex // write mutex
	done  chan struct{}
}

func newConn(
	socket *websocket.Conn,
	isBinary bool,
	pingInterval time.Duration,
	global, local *semaphore.Weighted) *connection {
	c := &connection{
		socket:       socket,
		pingInterval: pingInterval,
		globalCap:    global,
		localCap:     local,

		done: make(chan struct{}),
	}

	if isBinary {
		c.messageType = websocket.BinaryMessage
	} else {
		c.messageType = websocket.TextMessage
	}

	socket.SetPongHandler(c.pong)

	go c.pingLoop()

	return c
}

func (c *connection) ReserveCapacity(ctx context.Context) error {
	if err := c.globalCap.Acquire(ctx, 1); err != nil {
		return err
	}

	if err := c.localCap.Acquire(ctx, 1); err != nil {
		// release the resources for the global cap
		c.globalCap.Release(1)
		return err
	}

	return nil
}

func (c *connection) ReleaseCapacity() {
	c.globalCap.Release(1)
	c.localCap.Release(1)
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

	w, err := c.socket.NextWriter(c.messageType)
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
