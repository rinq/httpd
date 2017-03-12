package native

import (
	"context"
	"fmt"
	"log"

	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

type connection struct {
	peer   rinq.Peer
	send   func(message.Outgoing)
	logger *log.Logger

	done      chan struct{}
	destroyed chan uint16
	sessions  map[uint16]rinq.Session
}

func newConnection(
	peer rinq.Peer,
	send func(message.Outgoing),
	logger *log.Logger,
) *connection {
	return &connection{
		peer:   peer,
		send:   send,
		logger: logger,

		done:      make(chan struct{}),
		destroyed: make(chan uint16),
		sessions:  map[uint16]rinq.Session{},
	}
}

func (c *connection) Close() {
	close(c.done)

	for _, sess := range c.sessions {
		sess.Destroy()
	}
}

func (c *connection) VisitSessionCreate(m *message.SessionCreate) error {
	if _, ok := c.sessions[m.Session]; ok {
		return fmt.Errorf("session %d already exists", m.Session)
	}

	sess := c.peer.Session()

	if err := sess.Listen(newNotificationHandler(c.send, m.Session)); err != nil {
		return err
	}

	if err := sess.SetAsyncHandler(newAsyncHandler(c.send, m.Session)); err != nil {
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

// monitor waits for a session to be destroyed, then enqueues its removal from
// the session map.
func (c *connection) monitor(index uint16, sess rinq.Session) {
	select {
	case <-sess.Done():
		select {
		case c.destroyed <- index:
		case <-c.done:
		}
	case <-c.done:
	}
}

// newNotificationHandler returns a rinq.NotificationHandler that sends a
// message when a notification is received.
func newNotificationHandler(send func(message.Outgoing), index uint16) rinq.NotificationHandler {
	return func(_ context.Context, _ rinq.Session, n rinq.Notification) {
		send(message.NewNotification(index, n))
	}
}

// newAsyncHandler returns a rinq.AsyncHandler that sends a message when an
// async response is received.
func newAsyncHandler(send func(message.Outgoing), index uint16) rinq.AsyncHandler {
	return func(
		_ context.Context, _ rinq.Session, _ ident.MessageID,
		ns, cmd string,
		p *rinq.Payload, err error,
	) {
		if m, ok := message.NewAsyncResponse(index, ns, cmd, p, err); ok {
			send(m)
		}
	}
}
