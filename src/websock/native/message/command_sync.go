package message

import (
	"io"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

// SyncCall is an incoming message representing a synchronous command request.
type SyncCall struct {
	Session uint16
	Header  SyncCallHeader
	Payload *rinq.Payload
}

// SyncCallHeader is the header structure for SyncCall messages.
type SyncCallHeader struct {
	Seq       uint
	Namespace string
	Command   string
	Timeout   time.Duration
}

// Accept calls the appropriate visit method on v.
func (m *SyncCall) Accept(v Visitor) error {
	return v.VisitSyncCall(m)
}

func (m *SyncCall) read(r io.Reader, e Encoding) (err error) {
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
