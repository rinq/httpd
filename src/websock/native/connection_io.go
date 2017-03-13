package native

import (
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
)

// connectionIO handles reading and writing messages from a websocket connection
type connectionIO struct {
	socket   websock.Socket
	encoding message.Encoding
	ping     time.Duration

	recv     chan message.Incoming
	recvDone chan struct{}
	send     chan message.Outgoing
	sendDone chan struct{}

	err atomic.Value
}

func newConnectionIO(
	s websock.Socket,
	e message.Encoding,
	ping time.Duration,
) *connectionIO {
	i := &connectionIO{
		socket:   s,
		encoding: e,
		ping:     ping,

		recv:     make(chan message.Incoming),
		recvDone: make(chan struct{}),
		send:     make(chan message.Outgoing),
		sendDone: make(chan struct{}),
	}

	s.SetPongHandler(i.pong)

	go i.read()
	go i.write()

	return i
}

func (i *connectionIO) Wait() error {
	<-i.sendDone
	<-i.recvDone

	err, _ := i.err.Load().(error)
	return err
}

// Messages returns a channel for reading messages from the websocket.
func (i *connectionIO) Messages() <-chan message.Incoming {
	return i.recv
}

// Send enqueues a message for sending.
func (i *connectionIO) Send(msg message.Outgoing) {
	select {
	case i.send <- msg:
	case <-i.sendDone:
	}
}

// read reads messages from the websocket and writes them to the recv channel.
func (i *connectionIO) read() {
	defer close(i.recv)
	defer close(i.recvDone)

	err := i.pong("")

	for err == nil {
		var msg message.Incoming
		msg, err = read(i.socket, i.encoding)

		if err == nil {
			select {
			case i.recv <- msg:
				err = i.pong("")
			case <-i.sendDone:
				return
			}
		}
	}

	i.err.Store(err)
}

func (i *connectionIO) write() {
	defer close(i.sendDone)

	ticker := time.NewTicker(i.ping)
	defer ticker.Stop()

	var err error
	for err == nil {
		select {
		case msg := <-i.send:
			err = write(i.socket, i.encoding, msg)
		case <-ticker.C:
			err = i.socket.WriteMessage(websocket.PingMessage, nil)
		case <-i.recvDone:
			return
		}
	}

	i.err.Store(err)
}

// pong is called by the socket when a pong frame is received, or by IO when
// any other message is received.
func (i *connectionIO) pong(string) error {
	deadline := time.Now().Add(i.ping * 2)
	return i.socket.SetReadDeadline(deadline)
}

// read the the next incoming message from the websocket.
func read(s websock.Socket, e message.Encoding) (message.Incoming, error) {
	_, r, err := s.NextReader()
	if err != nil {
		return nil, err
	}

	return message.Read(r, e)
}

// write an outgoing message to the websocket.
func write(s websock.Socket, e message.Encoding, m message.Outgoing) error {
	w, err := s.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	return message.Write(w, e, m)
}
