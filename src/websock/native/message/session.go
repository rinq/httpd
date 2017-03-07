package message

import (
	"encoding/binary"
	"io"

	"github.com/ugorji/go/codec"
)

// SessionCreate is an incoming message request that a new session is created.
type SessionCreate struct {
	Session uint16
}

// Accept calls the appropriate visit method on v.
func (m *SessionCreate) Accept(v Visitor) error {
	return v.VisitSessionCreate(m)
}

func (m *SessionCreate) Read(r io.Reader, _ *codec.Decoder) error {
	return binary.Read(r, binary.BigEndian, &m.Session)
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

func (m *SessionDestroy) Read(r io.Reader, _ *codec.Decoder) error {
	return binary.Read(r, binary.BigEndian, &m.Session)
}

func (m *SessionDestroy) Write(w io.Writer, _ *codec.Encoder) error {
	if err := binary.Write(w, binary.BigEndian, SessionDestroyType); err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, m.Session)
}
