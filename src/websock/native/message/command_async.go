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

// AsyncSuccess is an outgoing message containing the successful repsonse to
// a synchronous call.
type AsyncSuccess struct {
	Session uint16
	Header  AsyncSuccessHeader
	Payload *rinq.Payload
}

// AsyncSuccessHeader is the header structure for AsyncSuccess messages.
type AsyncSuccessHeader struct {
	Namespace string
	Command   string
}

func (m *AsyncSuccess) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, commandAsyncSuccessType, m.Session)

	if err == nil {
		err = e.EncodeHeader(w, m.Header)

		if err == nil {
			e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// AsyncFailure is an outgoing message containing a failure repsonse to
// a synchronous call.
type AsyncFailure struct {
	Session uint16
	Header  AsyncFailureHeader
	Payload *rinq.Payload
}

// AsyncFailureHeader is the header structure for AsyncFailure messages.
type AsyncFailureHeader struct {
	Namespace      string
	Command        string
	FailureType    string
	FailureMessage string
}

func (m *AsyncFailure) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, commandAsyncFailureType, m.Session)

	if err == nil {
		err = e.EncodeHeader(w, m.Header)

		if err == nil {
			e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// AsyncError is an outgoing message containing a failure repsonse to
// a synchronous call.
type AsyncError struct {
	Session uint16
	Header  AsyncErrorHeader
}

// AsyncErrorHeader is the header structure for AsyncError messages.
type AsyncErrorHeader struct {
	Namespace string
	Command   string
}

func (m *AsyncError) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, commandAsyncErrorType, m.Session)

	if err == nil {
		err = e.EncodeHeader(w, m.Header)
	}

	return
}
