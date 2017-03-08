package message

import (
	"io"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

// AsyncCall is an incoming message representing a synchronous command request.
type AsyncCall struct {
	Session uint16
	Header  AsyncCallHeader
	Payload *rinq.Payload
}

// AsyncCallHeader is the header structure for AsyncCall messages.
type AsyncCallHeader struct {
	Namespace string
	Command   string
	Timeout   time.Duration
}

// Accept calls the appropriate visit method on v.
func (m *AsyncCall) Accept(v Visitor) error {
	return v.VisitAsyncCall(m)
}

func (m *AsyncCall) read(r io.Reader, e Encoding) (err error) {
	m.Session, err = readPreamble(r)

	if err == nil {
		err = e.DecodeHeader(r, &m.Header)
		m.Header.Timeout *= time.Millisecond

		if err == nil {
			m.Payload, err = e.DecodePayload(r)
		}
	}

	return
}
