package message

import (
	"io"

	"github.com/rinq/rinq-go/src/rinq"
)

// Execute is an incoming message representing a command execution request.
type Execute struct {
	preamble
	executeHeader

	Payload *rinq.Payload
}

// executeHeader is the header structure for Execute messages.
type executeHeader struct {
	Namespace string
	Command   string
}

// Accept calls the appropriate visit method on v.
func (m *Execute) Accept(v Visitor) error {
	return v.VisitExecute(m)
}

func (m *Execute) read(r io.Reader, e Encoding) (err error) {
	err = m.preamble.read(r)

	if err == nil {
		err = e.DecodeHeader(r, &m.executeHeader)

		if err == nil {
			m.Payload, err = e.DecodePayload(r)
		}
	}

	return
}
