package message

import (
	"io"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

// AsyncCall is an incoming message representing a synchronous command request.
type AsyncCall struct {
	preamble
	asyncCallHeader

	Payload *rinq.Payload
}

// asyncCallHeader is the header structure for AsyncCall messages.
type asyncCallHeader struct {
	Namespace string
	Command   string
	Timeout   time.Duration
}

// Accept calls the appropriate visit method on v.
func (m *AsyncCall) Accept(v Visitor) error {
	return v.VisitAsyncCall(m)
}

func (m *AsyncCall) read(r io.Reader, e Encoding) (err error) {
	err = m.preamble.read(r)

	if err == nil {
		err = e.DecodeHeader(r, &m.asyncCallHeader)
		m.asyncCallHeader.Timeout *= time.Millisecond

		if err == nil {
			m.Payload, err = e.DecodePayload(r)
		}
	}

	return
}

// AsyncSuccess is an outgoing message containing the successful response to
// a synchronous call.
type AsyncSuccess struct {
	preamble
	asyncSuccessHeader

	Payload *rinq.Payload
}

// asyncSuccessHeader is the header structure for AsyncSuccess messages.
type asyncSuccessHeader struct {
	Namespace string
	Command   string
}

func (m *AsyncSuccess) write(w io.Writer, e Encoding) (err error) {
	err = m.preamble.write(w, commandAsyncSuccessType)

	if err == nil {
		err = e.EncodeHeader(w, m.asyncSuccessHeader)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// AsyncFailure is an outgoing message containing a failure response to
// a synchronous call.
type AsyncFailure struct {
	preamble
	asyncFailureHeader

	Payload *rinq.Payload
}

// asyncFailureHeader is the header structure for AsyncFailure messages.
type asyncFailureHeader struct {
	Namespace      string
	Command        string
	FailureType    string
	FailureMessage string
}

func (m *AsyncFailure) write(w io.Writer, e Encoding) (err error) {
	err = m.preamble.write(w, commandAsyncFailureType)

	if err == nil {
		err = e.EncodeHeader(w, m.asyncFailureHeader)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// AsyncError is an outgoing message containing a failure response to
// a synchronous call.
type AsyncError struct {
	preamble
	asyncErrorHeader
}

// asyncErrorHeader is the header structure for AsyncError messages.
type asyncErrorHeader struct {
	Namespace string
	Command   string
}

func (m *AsyncError) write(w io.Writer, e Encoding) (err error) {
	err = m.preamble.write(w, commandAsyncErrorType)

	if err == nil {
		err = e.EncodeHeader(w, m.asyncErrorHeader)
	}

	return
}

// NewAsyncResponse returns an outgoing message to send an asynchronous command
// response to the client.
func NewAsyncResponse(
	session SessionIndex,
	ns, cmd string,
	p *rinq.Payload, err error,
) (Outgoing, bool) {
	switch e := err.(type) {
	case nil:
		return &AsyncSuccess{
			preamble: preamble{session},
			asyncSuccessHeader: asyncSuccessHeader{
				Namespace: ns,
				Command:   cmd,
			},
			Payload: p,
		}, true

	case rinq.Failure:
		return &AsyncFailure{
			preamble: preamble{session},
			asyncFailureHeader: asyncFailureHeader{
				Namespace:      ns,
				Command:        cmd,
				FailureType:    e.Type,
				FailureMessage: e.Message,
			},
			Payload: p,
		}, true

	case rinq.CommandError:
		return &AsyncError{
			preamble: preamble{session},
			asyncErrorHeader: asyncErrorHeader{
				Namespace: ns,
				Command:   cmd,
			},
		}, true
	}

	return nil, false
}
