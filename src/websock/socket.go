package websock

import (
	"io"
	"time"
)

// Socket is an interface for Gorilla's websocket.Conn struct.
type Socket interface {
	SetReadDeadline(t time.Time) error
	SetPongHandler(h func(string) error)
	NextReader() (int, io.Reader, error)
	NextWriter(int) (io.WriteCloser, error)
	WriteMessage(mt int, data []byte) error
}
