package message

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/ugorji/go/codec"
)

// Incoming is an interface for messages that are received from the browser.
type Incoming interface {
	// Accept calls the appropriate visit method on v.
	Accept(v Visitor) error

	// Read decodes the next message from r into this message.
	// It is assumed that the message type has already been read from r.
	Read(r io.Reader, d *codec.Decoder) error
}

// Outgoing is an interface for messages that are sent to the browser.
type Outgoing interface {
	// Write encodes this message to w, including the message type.
	Write(w io.Writer, e *codec.Encoder) error
}

// Read decodes the next message from r.
func Read(r io.Reader, d *codec.Decoder) (msg Incoming, err error) {
	var mt uint16
	err = binary.Read(r, binary.BigEndian, &mt)
	if err != nil {
		return
	}

	switch mt {
	case SessionCreateType:
		msg = &SessionCreate{}
	case SessionDestroyType:
		msg = &SessionDestroy{}
	default:
		err = errors.New("unrecognised message type")
		return
	}

	err = msg.Read(r, d)
	return
}
