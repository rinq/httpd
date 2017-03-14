package message

import (
	"io"
)

// Outgoing is an interface for messages that are sent to the browser.
type Outgoing interface {
	write(w io.Writer, e Encoding) error
}

// Write encodes m to w.
func Write(w io.Writer, e Encoding, m Outgoing) error {
	return m.write(w, e)
}
