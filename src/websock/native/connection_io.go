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

	recv chan message.Incoming
	send chan message.Outgoing
	stop chan struct{}
	done chan struct{}
	err  atomic.Value
}

func newIO(
	s websock.Socket,
	e message.Encoding,
	ping time.Duration,
) *connectionIO {
	i := &connectionIO{
		socket:   s,
		encoding: e,
		ping:     ping,

		recv: make(chan message.Incoming),
		send: make(chan message.Outgoing),
		stop: make(chan struct{}, 1),
		done: make(chan struct{}),
	}

	s.SetPongHandler(i.pong)

	go i.read()
	go i.write()

	return i
}

func (i *connectionIO) Stop() {
	select {
	case i.stop <- struct{}{}:
	default:
	}

	<-i.done
}

func (i *connectionIO) Err() error {
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
	case <-i.done:
	}
}

// read reads messages from the websocket and writes them to the recv channel.
func (i *connectionIO) read() {
	defer i.Stop()
	defer close(i.recv)

	err := i.pong("")

	for err == nil {
		var msg message.Incoming
		msg, err = read(i.socket, i.encoding)

		if err == nil {
			select {
			case i.recv <- msg:
				err = i.pong("")
			case <-i.stop:
				return
			}
		}
	}

	i.err.Store(err)
}

func (i *connectionIO) write() {
	defer i.Stop()

	defer close(i.done)

	ticker := time.NewTicker(i.ping)
	defer ticker.Stop()

	var err error
	for err == nil {
		select {
		case msg := <-i.send:
			err = write(i.socket, i.encoding, msg)
		case <-ticker.C:
			err = i.socket.WriteMessage(websocket.PingMessage, nil)
		case <-i.stop:
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
