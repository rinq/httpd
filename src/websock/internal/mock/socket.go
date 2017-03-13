package mock

import (
	"errors"
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
func (s *Socket) Subprotocol() string {
	if s.Impl.Subprotocol != nil {
		return s.Impl.Subprotocol()
	}

	return ""
}

// SetReadDeadline forwards to s.Impl.SetReadDeadline
func (s *Socket) SetReadDeadline(t time.Time) error {
	if s.Impl.SetReadDeadline != nil {
		return s.Impl.SetReadDeadline(t)
	}

	return nil
}

// SetPongHandler forwards to s.Impl.SetPongHandler
func (s *Socket) SetPongHandler(h func(string) error) {
	if s.Impl.SetPongHandler != nil {
		s.Impl.SetPongHandler(h)
	}
}

// NextReader forwards to s.Impl.NextReader
func (s *Socket) NextReader() (int, io.Reader, error) {
	if s.Impl.NextReader != nil {
		return s.Impl.NextReader()
	}

	return 0, nil, errors.New("no NextReader implementation provided")
}

// NextWriter forwards to s.Impl.NextWriter
func (s *Socket) NextWriter(mt int) (io.WriteCloser, error) {
	if s.Impl.NextWriter != nil {
		return s.Impl.NextWriter(mt)
	}

	return nil, errors.New("no NextWriter implementation provided")
}

// WriteMessage forwards to s.Impl.WriteMessage
func (s *Socket) WriteMessage(mt int, data []byte) error {
	if s.Impl.WriteMessage != nil {
		return s.Impl.WriteMessage(mt, data)
	}

	return nil
}
