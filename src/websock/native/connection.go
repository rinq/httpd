package native

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

type connection struct {
	peer     rinq.Peer
	ping     time.Duration
	socket   websock.Socket
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
	socket websock.Socket,
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

		sessions: map[uint16]rinq.Session{},
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
	case <-c.done:
	}
}

func (c *connection) send(msg message.Outgoing) error {
	w, err := c.socket.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	return message.Write(w, c.encoding, msg)
}

func (c *connection) monitor(index uint16, sess rinq.Session) {
	select {
	case <-c.done:
		return
	case <-sess.Done():
	}

	select {
	case c.destroyed <- index:
	case <-c.done:
	}
}

func (c *connection) notificationHandler(index uint16) rinq.NotificationHandler {
	return func(_ context.Context, _ rinq.Session, n rinq.Notification) {
		m := message.NewNotification(index, n)
		c.send(m)
	}
}

func (c *connection) asyncHandler(index uint16) rinq.AsyncHandler {
	return func(
		_ context.Context, _ rinq.Session, _ ident.MessageID,
		ns, cmd string,
		p *rinq.Payload, err error,
	) {
		if m, ok := message.NewAsyncResponse(index, ns, cmd, p, err); ok {
			c.send(m)
		}
	}
}

func (c *connection) VisitSessionCreate(m *message.SessionCreate) error {
	if _, ok := c.sessions[m.Session]; ok {
		return fmt.Errorf("session %d already exists", m.Session)
	}

	sess := c.peer.Session()

	if err := sess.Listen(c.notificationHandler(m.Session)); err != nil {
		return err
	}

	if err := sess.SetAsyncHandler(c.asyncHandler(m.Session)); err != nil {
		return err
	}

	c.sessions[m.Session] = sess
	go c.monitor(m.Session, sess)

	return nil
}

func (c *connection) VisitSessionDestroy(m *message.SessionDestroy) error {
	if sess, ok := c.sessions[m.Session]; ok {
		delete(c.sessions, m.Session)
		go sess.Destroy()
		return nil
	}

	return fmt.Errorf("session %d does not exist", m.Session)
}

func (c *connection) VisitSyncCall(m *message.SyncCall) error {
	if sess, ok := c.sessions[m.Session]; ok {
		go func() {
			p, err := sess.Call(
				context.TODO(), // needs timeout
				m.Header.Namespace,
				m.Header.Command,
				m.Payload,
			)

			if m, ok := message.NewSyncResponse(m.Session, m.Header.Seq, p, err); ok {
				c.send(m)
			}
		}()

		return nil
	}

	return fmt.Errorf("session %d does not exist", m.Session)
}

func (c *connection) VisitAsyncCall(m *message.AsyncCall) error {
	if sess, ok := c.sessions[m.Session]; ok {
		_, err := sess.CallAsync(
			context.TODO(), // needs timeout
			m.Header.Namespace,
			m.Header.Command,
			m.Payload,
		)

		return err
	}

	return fmt.Errorf("session %d does not exist", m.Session)
}

func (c *connection) VisitExecute(m *message.Execute) error {
	if sess, ok := c.sessions[m.Session]; ok {
		return sess.Execute(
			context.Background(),
			m.Header.Namespace,
			m.Header.Command,
			m.Payload,
		)
	}

	return fmt.Errorf("session %d does not exist", m.Session)
}
