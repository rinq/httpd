package mock

import (
	"io"
	"time"
)

// Socket is a mock websock.Socket
type Socket struct {
	Impl struct {
		Subprotocol     func() string
		SetReadDeadline func(t time.Time) error
		SetPongHandler  func(h func(string) error)
		NextReader      func() (int, io.Reader, error)
		NextWriter      func(int) (io.WriteCloser, error)
		WriteMessage    func(mt int, data []byte) error
	}
}

// Subprotocol forwards to s.Impl.Subprotocol
func (s *Socket) Subprotocol() string { return s.Impl.Subprotocol() }

// SetReadDeadline forwards to s.Impl.SetReadDeadline
func (s *Socket) SetReadDeadline(t time.Time) error { return s.Impl.SetReadDeadline(t) }

// SetPongHandler forwards to s.Impl.SetPongHandler
func (s *Socket) SetPongHandler(h func(string) error) { s.Impl.SetPongHandler(h) }

// NextReader forwards to s.Impl.NextReader
func (s *Socket) NextReader() (int, io.Reader, error) { return s.Impl.NextReader() }

// NextWriter forwards to s.Impl.NextWriter
func (s *Socket) NextWriter(mt int) (io.WriteCloser, error) { return s.Impl.NextWriter(mt) }

// WriteMessage forwards to s.Impl.WriteMessage
func (s *Socket) WriteMessage(mt int, data []byte) error { return s.Impl.WriteMessage(mt, data) }
