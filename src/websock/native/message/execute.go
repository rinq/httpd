package message

import (
	"io"

	"github.com/rinq/rinq-go/src/rinq"
)

// Execute is an incoming message representing a command execution request.
type Execute struct {
	Session uint16
	Header  ExecuteHeader
	Payload *rinq.Payload
}

// ExecuteHeader is the header structure for Execute messages.
type ExecuteHeader struct {
	Namespace string
	Command   string
}

// Accept calls the appropriate visit method on v.
func (m *Execute) Accept(v Visitor) error {
	return v.VisitExecute(m)
}

func (m *Execute) read(r io.Reader, e Encoding) (err error) {
	m.Session, err = readPreamble(r)

	if err == nil {
		err = e.DecodeHeader(r, &m.Header)

		if err == nil {
			m.Payload, err = e.DecodePayload(r)
		}
	}

	return
}
