package message

import (
	"io"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

// SyncCall is an incoming message representing a synchronous command request.
type SyncCall struct {
	preamble
	syncCallHeader

	Payload *rinq.Payload
}

// syncCallHeader is the header structure for SyncCall messages.
type syncCallHeader struct {
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
	err = m.preamble.read(r)

	if err == nil {
		err = e.DecodeHeader(r, &m.syncCallHeader)
		m.syncCallHeader.Timeout *= time.Millisecond

		if err == nil {
			m.Payload, err = e.DecodePayload(r)
		}
	}

	return
}

// SyncSuccess is an outgoing message containing the successful response to
// a synchronous call.
type SyncSuccess struct {
	preamble
	syncSuccessHeader

	Payload *rinq.Payload
}

// syncSuccessHeader is the header structure for SyncSuccess messages.
type syncSuccessHeader struct {
	Seq uint
}

func (m *SyncSuccess) write(w io.Writer, e Encoding) (err error) {
	err = m.preamble.write(w, commandSyncSuccessType)

	if err == nil {
		err = e.EncodeHeader(w, m.syncSuccessHeader)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// SyncFailure is an outgoing message containing a failure response to
// a synchronous call.
type SyncFailure struct {
	preamble
	syncFailureHeader

	Payload *rinq.Payload
}

// syncFailureHeader is the header structure for SyncFailure messages.
type syncFailureHeader struct {
	Seq            uint
	FailureType    string
	FailureMessage string
}

func (m *SyncFailure) write(w io.Writer, e Encoding) (err error) {
	err = m.preamble.write(w, commandSyncFailureType)

	if err == nil {
		err = e.EncodeHeader(w, m.syncFailureHeader)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// SyncError is an outgoing message containing a failure response to
// a synchronous call.
type SyncError struct {
	preamble
	syncErrorHeader
}

// syncErrorHeader is the header structure for SyncError messages.
type syncErrorHeader struct {
	Seq uint
}

func (m *SyncError) write(w io.Writer, e Encoding) (err error) {
	err = m.preamble.write(w, commandSyncErrorType)

	if err == nil {
		err = e.EncodeHeader(w, m.syncErrorHeader)
	}

	return
}

// NewSyncResponse returns an outgoing message to send a synchronous command
// response to the client.
//
// This method explicitly doesn't test for context-level timeouts, as the client side
// should be responsible dealing with this.
func NewSyncResponse(
	session SessionIndex,
	seq uint,
	p *rinq.Payload, err error,
) (Outgoing, bool) {
	switch e := err.(type) {
	case nil:
		return &SyncSuccess{
			preamble:          preamble{session},
			syncSuccessHeader: syncSuccessHeader{Seq: seq},
			Payload:           p,
		}, true

	case rinq.Failure:
		return &SyncFailure{
			preamble: preamble{session},
			syncFailureHeader: syncFailureHeader{
				Seq:            seq,
				FailureType:    e.Type,
				FailureMessage: e.Message,
			},
			Payload: p,
		}, true

	case rinq.CommandError:
		return &SyncError{
			preamble:        preamble{session},
			syncErrorHeader: syncErrorHeader{Seq: seq},
		}, true
	}

	return nil, false
}
