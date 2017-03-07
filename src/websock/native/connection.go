package native

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

type connection struct {
	peer     rinq.Peer
	ping     time.Duration
	socket   *websocket.Conn
	encoding message.Encoding

	incoming  chan message.Incoming
	outgoing  chan message.Outgoing
	destroyed chan uint16
	quit      chan error
	done      chan struct{}

	sessions map[uint16]rinq.Session
}

func newConnection(
	peer rinq.Peer,
	ping time.Duration,
	socket *websocket.Conn,
	encoding message.Encoding,
) *connection {
	return &connection{
		peer:     peer,
		ping:     ping,
		socket:   socket,
		encoding: encoding,

		incoming:  make(chan message.Incoming),
		outgoing:  make(chan message.Outgoing),
		destroyed: make(chan uint16),
		quit:      make(chan error),
		done:      make(chan struct{}),
	}
}

func (c *connection) Run() error {
	go c.read()

	defer close(c.done)

	ping := time.NewTicker(c.ping)
	defer ping.Stop()

	for {
		select {
		case msg := <-c.incoming:
			if err := msg.Accept(c); err != nil {
				return err
			}

		case msg := <-c.outgoing:
			if err := c.send(msg); err != nil {
				return err
			}

		case <-ping.C:
			if err := c.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}

		case index := <-c.destroyed:
			if _, ok := c.sessions[index]; ok {
				delete(c.sessions, index)

				msg := &message.SessionDestroy{Session: index}
				if err := c.send(msg); err != nil {
					return err
				}
			}

		case err := <-c.quit:
			return err
		}
	}
}

func (c *connection) pong(string) error {
	deadline := time.Now().Add(c.ping * 2)
	c.socket.SetReadDeadline(deadline)
	return nil
}

func (c *connection) read() {
	c.socket.SetPongHandler(c.pong)
	c.pong("")

	for {
		_, r, err := c.socket.NextReader()
		if err != nil {
			c.stop(err)
			return
		}

		msg, err := message.Read(r, c.encoding)
		if err != nil {
			c.stop(err)
			return
		}

		select {
		case c.incoming <- msg:
		case <-c.done:
			return
		}
	}
}

func (c *connection) stop(err error) {
	select {
	case c.quit <- err:
		close(c.quit)
	case <-c.done:
	}
}

func (c *connection) monitorSession(sess rinq.Session, index uint16) {
	<-sess.Done()

	select {
	case c.destroyed <- index:
	case <-c.done:
	}
}

func (c *connection) VisitSessionCreate(m *message.SessionCreate) error {
	if _, ok := c.sessions[m.Session]; ok {
		return fmt.Errorf("session %d already exists", m.Session)
	}

	sess := c.peer.Session()
	c.sessions[m.Session] = sess

	go c.monitorSession(sess, m.Session)

	return nil
}

func (c *connection) VisitSessionDestroy(m *message.SessionDestroy) error {
	sess, ok := c.sessions[m.Session]
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	delete(c.sessions, m.Session)

	go sess.Destroy()

	return nil
}

func (c *connection) send(msg message.Outgoing) error {
	w, err := c.socket.NextWriter(websocket.BinaryMessage)
	defer w.Close()

	if err != nil {
		return err
	}

	return msg.Write(w, c.encoding)
}
