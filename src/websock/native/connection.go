package native

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

type connection struct {
	session func() (rinq.Session, error)
	send    func(message.Outgoing)
	logger  *log.Logger

	parent context.Context
	cancel func()

	mutex    sync.RWMutex
	sessions map[uint16]rinq.Session
}

func newConnection(
	session func() (rinq.Session, error),
	send func(message.Outgoing),
	logger *log.Logger,
) *connection {
	c := &connection{
		session:  session,
		send:     send,
		logger:   logger,
		sessions: map[uint16]rinq.Session{},
	}

	c.parent, c.cancel = context.WithCancel(context.Background())

	return c
}

func (c *connection) Close() {
	c.cancel()

	for _, sess := range c.sessions {
		go sess.Destroy()
	}
}

func (c *connection) VisitSessionCreate(m *message.SessionCreate) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.sessions[m.Session]; ok {
		return fmt.Errorf("session %d already exists", m.Session)
	}

	sess, err := c.session()
	if err != nil {
		return err
	}

	err = registerHandlers(sess, m.Session, c.send)
	if err != nil {
		sess.Destroy()
		return err
	}

	c.sessions[m.Session] = sess
	go c.monitor(m.Session, sess)

	return nil
}

func (c *connection) VisitSessionDestroy(m *message.SessionDestroy) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if sess, ok := c.sessions[m.Session]; ok {
		delete(c.sessions, m.Session)
		go sess.Destroy()
		return nil
	}

	return fmt.Errorf("session %d does not exist", m.Session)
}

func (c *connection) VisitSyncCall(m *message.SyncCall) error {
	sess, ok := c.getSession(m.Session)
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	go c.call(sess, m)

	return nil
}

func (c *connection) VisitAsyncCall(m *message.AsyncCall) error {
	sess, ok := c.getSession(m.Session)
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	ctx, cancel := context.WithTimeout(c.parent, m.Header.Timeout)
	defer cancel()

	_, err := sess.CallAsync(
		ctx,
		m.Header.Namespace,
		m.Header.Command,
		m.Payload,
	)

	return err
}

func (c *connection) VisitExecute(m *message.Execute) error {
	sess, ok := c.getSession(m.Session)
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

func (c *connection) getSession(index uint16) (rinq.Session, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	sess, ok := c.sessions[index]
	return sess, ok
}

func (c *connection) call(sess rinq.Session, m *message.SyncCall) {
	ctx, cancel := context.WithTimeout(c.parent, m.Header.Timeout)
	defer cancel()

	p, err := sess.Call(
		ctx,
		m.Header.Namespace,
		m.Header.Command,
		m.Payload,
	)

	if m, ok := message.NewSyncResponse(m.Session, m.Header.Seq, p, err); ok {
		c.send(m)
	}
}

// monitor waits for a session to be destroyed, then enqueues its removal from
// the session map.
func (c *connection) monitor(index uint16, sess rinq.Session) {
	select {
	case <-sess.Done():
	case <-c.parent.Done():
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.sessions[index]
	if !ok {
		return
	}

	delete(c.sessions, index)
	c.send(&message.SessionDestroy{Session: index})
}

// registerHandlers sets up notification and async response handlers on a
// session.
func registerHandlers(
	sess rinq.Session,
	index uint16,
	send func(message.Outgoing),
) error {
	// register a notification listener
	if err := sess.Listen(func(
		_ context.Context,
		_ rinq.Session,
		n rinq.Notification,
	) {
		send(message.NewNotification(index, n))
	}); err != nil {
		return err
	}

	// register an async response handler
	return sess.SetAsyncHandler(func(
		_ context.Context,
		_ rinq.Session,
		_ ident.MessageID,
		ns string,
		cmd string,
		p *rinq.Payload,
		err error,
	) {
		if m, ok := message.NewAsyncResponse(index, ns, cmd, p, err); ok {
			send(m)
		}
	})
}
