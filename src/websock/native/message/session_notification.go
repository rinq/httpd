package message

import (
	"io"

	"github.com/rinq/rinq-go/src/rinq"
)

// Notification is an outoing message indicating that notification was received
// by a session.
type Notification struct {
	Session uint16
	Header  NotificationHeader
	Payload *rinq.Payload
}

// NotificationHeader is the header structure for Notification messages.
type NotificationHeader struct {
	Type string
}

func (m *Notification) write(w io.Writer, e Encoding) (err error) {
	err = writePreamble(w, sessionNotificationType, m.Session)

	if err == nil {
		err = e.EncodeHeader(w, m.Header)

		if err == nil {
			err = e.EncodePayload(w, m.Payload)
		}
	}

	return
}

// NewNotification returns an outgoing message to send a notification to the client.
func NewNotification(session uint16, n rinq.Notification) Outgoing {
	return &Notification{
		Session: session,
		Header:  NotificationHeader{Type: n.Type},
		Payload: n.Payload,
	}
}
