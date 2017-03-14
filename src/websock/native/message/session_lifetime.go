package message

import (
	"io"
)

// SessionCreate is an incoming message requesting that a new session be created.
type SessionCreate struct {
	preamble
}

// Accept calls the appropriate visit method on v.
func (m *SessionCreate) Accept(v Visitor) error {
	return v.VisitSessionCreate(m)
}

func (m *SessionCreate) read(r io.Reader, _ Encoding) error {
	return m.preamble.read(r)
}

// SessionDestroy is a bidirectional message.
//
// When received from the browser it indicates a request that an existing
// session be destroyed.
//
// When sent to the browser it indicates that an existing session has been
// destroyed without being requested by the client.
type SessionDestroy struct {
	preamble
}

// NewSessionDestroy returns an outgoing message to send a notification to the client.
func NewSessionDestroy(session SessionIndex) *SessionDestroy {
	return &SessionDestroy{
		preamble: preamble{session},
	}
}

// Accept calls the appropriate visit method on v.
func (m *SessionDestroy) Accept(v Visitor) error {
	return v.VisitSessionDestroy(m)
}

func (m *SessionDestroy) read(r io.Reader, _ Encoding) error {
	return m.preamble.read(r)
}

func (m *SessionDestroy) write(w io.Writer, _ Encoding) error {
	return m.preamble.write(w, sessionDestroyType)
}
