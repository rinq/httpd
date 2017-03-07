package message

import (
	"encoding/binary"
	"errors"
	"io"
)

// Incoming is an interface for messages that are received from the browser.
type Incoming interface {
	// Accept calls the appropriate visit method on v.
	Accept(v Visitor) error

	// Read decodes the next message from r into this message.
	// It is assumed that the message type has already been read from r.
	Read(r io.Reader, e Encoding) error
}

// Outgoing is an interface for messages that are sent to the browser.
type Outgoing interface {
	// Write encodes this message to w, including the message type.
	Write(w io.Writer, e Encoding) error
}

// Read decodes the next message from r.
func Read(r io.Reader, e Encoding) (msg Incoming, err error) {
	var mt uint16

	err = binary.Read(r, binary.BigEndian, &mt)

	if err == nil {
		switch mt {
		case sessionCreateType:
			msg = &SessionCreate{}
		case sessionDestroyType:
			msg = &SessionDestroy{}
		default:
			err = errors.New("unrecognised incoming message type")
			return
		}

		err = msg.Read(r, e)
	}

	return
}

const (
	sessionCreateType  uint16 = 'S'<<8 | 'C'
	sessionDestroyType uint16 = 'S'<<8 | 'D'

	commandSyncRequestType uint16 = 'C'<<8 | 'C'
	commandSyncSuccessType uint16 = 'C'<<8 | 'S'
	commandSyncFailureType uint16 = 'C'<<8 | 'F'
	commandSyncErrorType   uint16 = 'C'<<8 | 'E'

	commandAsyncRequestType uint16 = 'A'<<8 | 'C'
	commandAsyncSuccessType uint16 = 'A'<<8 | 'S'
	commandAsyncFailureType uint16 = 'A'<<8 | 'F'
	commandAsyncErrorType   uint16 = 'A'<<8 | 'E'

	commandExecuteType uint16 = 'C'<<8 | 'X'

	notificationType uint16 = 'N'<<8 | 'O'
)
