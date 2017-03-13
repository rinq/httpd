package native

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
)

type sessionFactory func() (rinq.Session, error)

func newSessionFactory(peer rinq.Peer, attrs []rinq.Attr) sessionFactory {
	return func() (sess rinq.Session, err error) {
		sess = peer.Session()

		if err := setAttrs(sess, attrs); err != nil {
			sess.Destroy()
			return nil, err
		}

		return sess, nil
	}
}

//
// setAttrs sets the initial attributes for a session.
func setAttrs(sess rinq.Session, attrs []rinq.Attr) error {
	rev, err := sess.CurrentRevision()
	if err != nil {
		return err
	}

	_, err = rev.Update(context.Background(), attrs...)
	return err
}
