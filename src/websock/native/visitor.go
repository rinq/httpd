package native

import (
	"context"
	"fmt"
	"sync"

	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"time"
)

type visitor struct {
	context context.Context
	peer    rinq.Peer
	attrs   []rinq.Attr
	send    func(message.Outgoing)

	mutex   sync.RWMutex
	forward map[message.SessionIndex]rinq.Session
	reverse map[ident.SessionID]message.SessionIndex

	syncCallTimeout time.Duration
}

func newVisitor(
	context context.Context,
	peer rinq.Peer,
	attrs []rinq.Attr,
	send func(message.Outgoing),
) *visitor {
	return &visitor{
		context: context,
		peer:    peer,
		attrs:   attrs,
		send:    send,
	}
}

func (v *visitor) VisitSessionCreate(m *message.SessionCreate) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if _, ok := v.forward[m.Session]; ok {
		return fmt.Errorf("session %d already exists", m.Session)
	}

	sess, err := v.newSession()
	if err != nil {
		return err
	}

	if v.forward == nil {
		v.forward = map[message.SessionIndex]rinq.Session{}
		v.reverse = map[ident.SessionID]message.SessionIndex{}
	}

	v.forward[m.Session] = sess
	v.reverse[sess.ID()] = m.Session

	go v.monitor(sess)

	return nil
}

func (v *visitor) VisitSessionDestroy(m *message.SessionDestroy) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	sess, ok := v.forward[m.Session]
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	delete(v.forward, m.Session)
	delete(v.reverse, sess.ID())
	go sess.Destroy()

	return nil
}

func (v *visitor) VisitListen(m *message.Listen) error {
	sess, ok := v.find(m.Session)
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	for _, ns := range m.Namespaces {
		if err := sess.Listen(ns, v.notify); err != nil {
			return err
		}
	}

	return nil

}

func (v *visitor) VisitUnlisten(m *message.Unlisten) error {
	sess, ok := v.find(m.Session)
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	for _, ns := range m.Namespaces {
		if err := sess.Unlisten(ns); err != nil {
			return err
		}
	}

	return nil
}

func (v *visitor) VisitSyncCall(m *message.SyncCall) error {
	if sess, ok := v.find(m.Session); ok {
		go v.call(sess, m)
		return nil
	}

	return fmt.Errorf("session %d does not exist", m.Session)
}

func (v *visitor) VisitAsyncCall(m *message.AsyncCall) error {
	sess, ok := v.find(m.Session)
	if !ok {
		return fmt.Errorf("session %d does not exist", m.Session)
	}

	ctx, cancel := context.WithTimeout(v.context, m.Timeout)
	defer cancel()

	_, err := sess.CallAsync(ctx, m.Namespace, m.Command, m.Payload)
	return err
}

func (v *visitor) VisitExecute(m *message.Execute) error {
	if sess, ok := v.find(m.Session); ok {
		return sess.Execute(v.context, m.Namespace, m.Command, m.Payload)
	}

	return fmt.Errorf("session %d does not exist", m.Session)
}

func (v *visitor) newSession() (sess rinq.Session, err error) {
	sess = v.peer.Session()

	defer func() {
		if err != nil {
			sess.Destroy()
		}
	}()

	if err = sess.SetAsyncHandler(v.respond); err != nil {
		return
	}

	_, err = sess.CurrentRevision().Update(v.context, HttpdAttrNamespace, v.attrs...)

	return
}

func (v *visitor) find(i message.SessionIndex) (rinq.Session, bool) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	sess, ok := v.forward[i]
	return sess, ok
}

func (v *visitor) indexOf(sess rinq.Session) (message.SessionIndex, bool) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	i, ok := v.reverse[sess.ID()]
	return i, ok
}

func (v *visitor) call(sess rinq.Session, m *message.SyncCall) {

	timeout := v.capSyncCallTimeout(m.Timeout)
	ctx, cancel := context.WithTimeout(v.context, timeout)
	defer cancel()

	p, err := sess.Call(ctx, m.Namespace, m.Command, m.Payload)

	if m, ok := message.NewSyncResponse(m.Session, m.Seq, p, err); ok {
		v.send(m)
	}
}

func (v *visitor) notify(
	_ context.Context,
	sess rinq.Session,
	n rinq.Notification,
) {
	if i, ok := v.indexOf(sess); ok {
		m := message.NewNotification(i, n)
		v.send(m)
	}
}

func (v *visitor) respond(
	_ context.Context,
	sess rinq.Session,
	_ ident.MessageID,
	ns string,
	cmd string,
	p *rinq.Payload,
	err error,
) {
	if i, ok := v.indexOf(sess); ok {
		if m, ok := message.NewAsyncResponse(i, ns, cmd, p, err); ok {
			v.send(m)
		}
	}
}

// monitor waits for a session to be destroyed, then enqueues its removal from
// the session map.
func (v *visitor) monitor(sess rinq.Session) {
	select {
	case <-sess.Done():
	case <-v.context.Done():
		return
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	if i, ok := v.reverse[sess.ID()]; ok {
		delete(v.forward, i)
		delete(v.reverse, sess.ID())
		v.send(message.NewSessionDestroy(i))
	}
}

func (v *visitor) capSyncCallTimeout(t time.Duration) time.Duration {
	if v.syncCallTimeout == 0 || v.syncCallTimeout > t {
		return t
	}

	return v.syncCallTimeout
}
