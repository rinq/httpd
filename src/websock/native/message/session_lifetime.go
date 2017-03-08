package message

import (
	"io"
)

// SessionCreate is an incoming message requesting that a new session be created.
type SessionCreate struct {
	Session uint16
}

// Accept calls the appropriate visit method on v.
func (m *SessionCreate) Accept(v Visitor) error {
	return v.VisitSessionCreate(m)
}

func (m *SessionCreate) read(r io.Reader, e Encoding) (err error) {
	m.Session, err = readPreamble(r)
	return
}

// SessionDestroy is a bidirectional message.
//
// When received from the browser it indicates a request that an existing
// session be destroyed.
//
// When sent to the browser it indicates that an existing session has been
// destroyed without being requested by the client.
type SessionDestroy struct {
	Session uint16
}

// Accept calls the appropriate visit method on v.
func (m *SessionDestroy) Accept(v Visitor) error {
	return v.VisitSessionDestroy(m)
}

func (m *SessionDestroy) read(r io.Reader, e Encoding) (err error) {
	m.Session, err = readPreamble(r)
	return
}

func (m *SessionDestroy) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, sessionDestroyType, m.Session)

	if err == nil {
		// empty header size
		_, err = w.Write([]byte{0, 0})
	}

	return
}
