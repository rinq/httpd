package native

import (
	"context"

	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

type session struct {
	index   uint16
	session rinq.Session

	outgoing  chan message.Outgoing
	destroyed chan uint16
	quit      chan error
	done      chan struct{}
}

func newSession(
	index uint16,
	sess rinq.Session,
	outgoing chan message.Outgoing,
	destroyed chan uint16,
	quit chan error,
	done chan struct{},
) (*session, error) {
	s := &session{
		index:   index,
		session: sess,

		outgoing:  outgoing,
		destroyed: destroyed,
		quit:      quit,
		done:      done,
	}

	if err := sess.Listen(s.notify); err != nil {
		return nil, err
	}

	if err := sess.SetAsyncHandler(s.response); err != nil {
		return nil, err
	}

	go s.monitor()

	return s, nil
}

func (s *session) Call(m *message.SyncCall) {
	go func() {
		in, err := s.session.Call(
			context.TODO(), // needs timeout
			m.Header.Namespace,
			m.Header.Command,
			m.Payload,
		)

		switch e := err.(type) {
		case nil:
			s.send(&message.SyncSuccess{
				Session: m.Session,
				Header:  message.SyncSuccessHeader{Seq: m.Header.Seq},
				Payload: in,
			})

		case rinq.Failure:
			s.send(&message.SyncFailure{
				Session: m.Session,
				Header: message.SyncFailureHeader{
					Seq:            m.Header.Seq,
					FailureType:    e.Type,
					FailureMessage: e.Message,
				},
				Payload: in,
			})

		case rinq.CommandError:
			s.send(&message.SyncError{
				Session: m.Session,
				Header:  message.SyncErrorHeader{Seq: m.Header.Seq},
			})

		default:
			s.stop(err)
		}
	}()
}

func (s *session) CallAsync(m *message.AsyncCall) error {
	_, err := s.session.CallAsync(
		context.TODO(), // needs timeout
		m.Header.Namespace,
		m.Header.Command,
		m.Payload,
	)

	return err
}

func (s *session) Execute(m *message.Execute) error {
	return s.session.Execute(
		context.Background(),
		m.Header.Namespace,
		m.Header.Command,
		m.Payload,
	)
}

func (s *session) Destroy() {
	go s.session.Destroy()
}

func (s *session) notify(
	ctx context.Context,
	target rinq.Session,
	n rinq.Notification,
) {
	s.send(&message.Notification{
		Session: s.index,
		Header:  message.NotificationHeader{Type: n.Type},
		Payload: n.Payload,
	})
}

func (s *session) response(
	ctx context.Context,
	sess rinq.Session, msgID ident.MessageID,
	ns, cmd string,
	in *rinq.Payload, err error,
) {
	switch e := err.(type) {
	case nil:
		s.send(&message.AsyncSuccess{
			Session: s.index,
			Header: message.AsyncSuccessHeader{
				Namespace: ns,
				Command:   cmd,
			},
			Payload: in,
		})

	case rinq.Failure:
		s.send(&message.AsyncFailure{
			Session: s.index,
			Header: message.AsyncFailureHeader{
				Namespace:      ns,
				Command:        cmd,
				FailureType:    e.Type,
				FailureMessage: e.Message,
			},
			Payload: in,
		})

	case rinq.CommandError:
		s.send(&message.AsyncError{
			Session: s.index,
			Header: message.AsyncErrorHeader{
				Namespace: ns,
				Command:   cmd,
			},
		})
	}
}

func (s *session) send(msg message.Outgoing) {
	select {
	case s.outgoing <- msg:
	case <-s.done:
	}
}

func (s *session) stop(err error) {
	select {
	case s.quit <- err:
	case <-s.done:
	}
}

func (s *session) monitor() {
	select {
	case <-s.session.Done():
		select {
		case s.destroyed <- s.index:
		case <-s.done:
		}
	case <-s.done:
	}
}
