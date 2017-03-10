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

// SyncSuccess is an outgoing message containing the successful response to
// a synchronous call.
type SyncSuccess struct {
	Session uint16
	Header  SyncSuccessHeader
	Payload *rinq.Payload
}

// SyncSuccessHeader is the header structure for SyncSuccess messages.
type SyncSuccessHeader struct {
	Seq uint
}

func (m *SyncSuccess) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, commandSyncSuccessType, m.Session)

	if err == nil {
		err = e.EncodeHeader(w, m.Header)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// SyncFailure is an outgoing message containing a failure response to
// a synchronous call.
type SyncFailure struct {
	Session uint16
	Header  SyncFailureHeader
	Payload *rinq.Payload
}

// SyncFailureHeader is the header structure for SyncFailure messages.
type SyncFailureHeader struct {
	Seq            uint
	FailureType    string
	FailureMessage string
}

func (m *SyncFailure) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, commandSyncFailureType, m.Session)

	if err == nil {
		err = e.EncodeHeader(w, m.Header)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// SyncError is an outgoing message containing a failure response to
// a synchronous call.
type SyncError struct {
	Session uint16
	Header  SyncErrorHeader
}

// SyncErrorHeader is the header structure for SyncError messages.
type SyncErrorHeader struct {
	Seq uint
}

func (m *SyncError) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, commandSyncErrorType, m.Session)

	if err == nil {
		err = e.EncodeHeader(w, m.Header)
	}

	return
}

// NewSyncResponse returns an outgoing message to send a synchronous command
// response to the client.
func NewSyncResponse(
	session uint16,
	seq uint,
	p *rinq.Payload, err error,
) (Outgoing, bool) {
	switch e := err.(type) {
	case nil:
		return &SyncSuccess{
			Session: session,
			Header:  SyncSuccessHeader{Seq: seq},
			Payload: p,
		}, true

	case rinq.Failure:
		return &SyncFailure{
			Session: session,
			Header: SyncFailureHeader{
				Seq:            seq,
				FailureType:    e.Type,
				FailureMessage: e.Message,
			},
			Payload: p,
		}, true

	case rinq.CommandError:
		return &SyncError{
			Session: session,
			Header:  SyncErrorHeader{Seq: seq},
		}, true
	}

	return nil, false
}
