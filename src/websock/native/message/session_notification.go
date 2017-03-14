package message

import (
	"io"

	"github.com/rinq/rinq-go/src/rinq"
)

// Notification is an outoing message indicating that notification was received
// by a session.
type Notification struct {
	preamble
	notificationHeader

	Payload *rinq.Payload
}

// NewNotification returns an outgoing message to send a notification to the client.
func NewNotification(session SessionIndex, n rinq.Notification) *Notification {
	return &Notification{
		preamble:           preamble{session},
		notificationHeader: notificationHeader{Type: n.Type},
		Payload:            n.Payload,
	}
}

// notificationHeader is the header structure for Notification messages.
type notificationHeader struct {
	Type string
}

func (m *Notification) write(w io.Writer, e Encoding) (err error) {
	err = m.preamble.write(w, sessionNotificationType)

	if err == nil {
		err = e.EncodeHeader(w, m.notificationHeader)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}
