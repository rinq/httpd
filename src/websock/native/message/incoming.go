package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Incoming is an interface for messages that are received from the browser.
type Incoming interface {
	// Accept calls the appropriate visit method on v.
	Accept(v Visitor) error

	// read decodes the next message from r into this message.
	// It is assumed that the message type has already been read from r.
	read(r io.Reader, e Encoding) error
}

// Read decodes the next message from r.
func Read(r io.Reader, e Encoding) (msg Incoming, err error) {
	var mt messageType

	err = binary.Read(r, binary.BigEndian, &mt)

	if err == nil {
		switch mt {
		case commandSyncCallType:
			msg = &SyncCall{}
		case commandAsyncCallType:
			msg = &AsyncCall{}
		case commandExecuteType:
			msg = &Execute{}
		case sessionCreateType:
			msg = &SessionCreate{}
		case sessionDestroyType:
			msg = &SessionDestroy{}
		default:
			err = fmt.Errorf("unrecognized incoming message type: 0x%04x", mt)
			return
		}

		err = msg.read(r, e)

		if err == nil {
			buf := make([]byte, 1)
			n, eof := r.Read(buf)
			if n > 0 || eof != io.EOF {
				err = errors.New("unconsumed frame data")
			}
		}
	}

	return
}

// Visitor is an interface that visits each of the incoming message types.
type Visitor interface {
	VisitSessionCreate(*SessionCreate) error
	VisitSessionDestroy(*SessionDestroy) error
	VisitSyncCall(*SyncCall) error
	VisitAsyncCall(*AsyncCall) error
	VisitExecute(*Execute) error
}
