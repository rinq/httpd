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

	if err := sess.Listen(c.notificationHandler(m.Session)); err != nil {
		return err
	}

	if err := sess.SetAsyncHandler(c.asyncHandler(m.Session)); err != nil {
		return err
	}

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

func (c *connection) VisitSyncCall(m *message.SyncCall) error {
	sess, ok := c.sessions[m.Session]
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	go c.call(sess, m)

	return nil
}

func (c *connection) VisitAsyncCall(m *message.AsyncCall) error {
	sess, ok := c.sessions[m.Session]
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	_, err := sess.CallAsync(
		context.TODO(), // needs timeout
		m.Header.Namespace,
		m.Header.Command,
		m.Payload,
	)

	return err
}

func (c *connection) VisitExecute(m *message.Execute) error {
	sess, ok := c.sessions[m.Session]
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	return sess.Execute(
		context.Background(),
		m.Header.Namespace,
		m.Header.Command,
		m.Payload,
	)
}

func (c *connection) send(msg message.Outgoing) error {
	w, err := c.socket.NextWriter(websocket.BinaryMessage)
	defer w.Close()

	if err != nil {
		return err
	}

	return message.Write(w, c.encoding, msg)
}

func (c *connection) enqueue(msg message.Outgoing) {
	select {
	case c.outgoing <- msg:
	case <-c.done:
	}
}

func (c *connection) call(sess rinq.Session, m *message.SyncCall) {
	in, err := sess.Call(
		context.TODO(), // needs timeout
		m.Header.Namespace,
		m.Header.Command,
		m.Payload,
	)

	switch e := err.(type) {
	case nil:
		c.enqueue(&message.SyncSuccess{
			Session: m.Session,
			Header:  message.SyncSuccessHeader{Seq: m.Header.Seq},
			Payload: in,
		})

	case rinq.Failure:
		c.enqueue(&message.SyncFailure{
			Session: m.Session,
			Header: message.SyncFailureHeader{
				Seq:            m.Header.Seq,
				FailureType:    e.Type,
				FailureMessage: e.Message,
			},
			Payload: in,
		})

	case rinq.CommandError:
		c.enqueue(&message.SyncError{
			Session: m.Session,
			Header:  message.SyncErrorHeader{Seq: m.Header.Seq},
		})

	default:
		c.stop(err)
	}
}

func (c *connection) notificationHandler(sessionIndex uint16) rinq.NotificationHandler {
	return func(
		ctx context.Context,
		target rinq.Session,
		n rinq.Notification,
	) {
		c.enqueue(&message.Notification{
			Session: sessionIndex,
			Header:  message.NotificationHeader{Type: n.Type},
			Payload: n.Payload,
		})
	}
}

func (c *connection) asyncHandler(sessionIndex uint16) rinq.AsyncHandler {
	return func(
		ctx context.Context,
		sess rinq.Session, msgID ident.MessageID,
		ns, cmd string,
		in *rinq.Payload, err error,
	) {
		switch e := err.(type) {
		case nil:
			c.enqueue(&message.AsyncSuccess{
				Session: sessionIndex,
				Header: message.AsyncSuccessHeader{
					Namespace: ns,
					Command:   cmd,
				},
				Payload: in,
			})

		case rinq.Failure:
			c.enqueue(&message.AsyncFailure{
				Session: sessionIndex,
				Header: message.AsyncFailureHeader{
					Namespace:      ns,
					Command:        cmd,
					FailureType:    e.Type,
					FailureMessage: e.Message,
				},
				Payload: in,
			})

		case rinq.CommandError:
			c.enqueue(&message.AsyncError{
				Session: sessionIndex,
				Header: message.AsyncErrorHeader{
					Namespace: ns,
					Command:   cmd,
				},
			})
		}
	}
}
