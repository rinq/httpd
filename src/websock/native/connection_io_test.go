package native

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/httpd/src/websock/internal/mock"
	"github.com/rinq/httpd/src/websock/native/message"
)

var _ = Describe("connectionIO", func() {
	var (
		msg    *message.SessionDestroy
		msgBuf *bytes.Buffer

		readers chan io.Reader
		writers chan io.WriteCloser
		pings   chan int

		socket  *mock.Socket
		subject *connectionIO
	)

	BeforeEach(func() {
		msg = &message.SessionDestroy{Session: 0xabcd}
		msgBuf = &bytes.Buffer{}

		if err := message.Write(msgBuf, message.JSONEncoding, msg); err != nil {
			panic(err)
		}

		readers = make(chan io.Reader, 10)
		writers = make(chan io.WriteCloser, 10)
		pings = make(chan int, 10)

		socket = &mock.Socket{}
		socket.Impl.NextReader = func() (int, io.Reader, error) {
			if r, ok := <-readers; ok {
				return websocket.BinaryMessage, r, nil
			}

			return 0, nil, errors.New("<no more readers>")
		}
		socket.Impl.NextWriter = func(int) (io.WriteCloser, error) {
			fmt.Println("waiting for writer")
			if w, ok := <-writers; ok {
				fmt.Println("got one")
				return w, nil
			}

			fmt.Println("none left")
			return nil, errors.New("<no more writers>")
		}
		socket.Impl.WriteMessage = func(mt int, body []byte) error {
			pings <- mt
			return nil
		}

		subject = newConnectionIO(socket, message.JSONEncoding, 500*time.Millisecond)
	})

	AfterEach(func() {
		defer func() {
			recover()
		}()
		close(readers)
	})

	AfterEach(func() {
		defer func() {
			recover()
		}()
		close(writers)
	})

	Describe("Messages", func() {
		It("makes decoded messages available", func() {
			readers <- msgBuf

			select {
			case m := <-subject.Messages():
				Expect(m).To(Equal(msg))
			case <-time.After(time.Second):
				panic("timeout")
			}
		})

		It("returns a closed channel when there are no more messages", func() {
			close(readers)

			select {
			case _, ok := <-subject.Messages():
				Expect(ok).To(BeFalse())
			case <-time.After(time.Second):
				panic("timeout")
			}
		})
	})

	Describe("Send", func() {
		It("writes the message to the socket", func() {
			buf := &bytes.Buffer{}
			writers <- nopCloser{buf}
			subject.Send(msg)

			close(readers)
			subject.Wait()

			Expect(buf).To(Equal(msgBuf))
		})
	})

	Describe("Wait", func() {
		It("returns the read error", func() {
			close(readers)

			err := subject.Wait()

			Expect(err).To(MatchError("<no more readers>"))
		})

		It("returns the write error", func() {
			readers <- msgBuf
			close(writers)
			subject.Send(msg)

			err := subject.Wait()

			Expect(err).To(MatchError("<no more writers>"))
		})
	})

	It("sends a ping message if no other message is sent", func() {
		select {
		case mt := <-pings:
			Expect(mt).To(Equal(websocket.PingMessage))
		case <-time.After(time.Second):
			panic("timeout")
		}
	})
})

var _ = Describe("read", func() {
	It("it returns an error if a reader can not be obtained", func() {
		expected := errors.New("<error>")
		socket := &mock.Socket{}
		socket.Impl.NextReader = func() (int, io.Reader, error) {
			return 0, nil, expected
		}

		msg, err := read(socket, nil)

		Expect(err).To(Equal(err))
		Expect(msg).To(BeNil())
	})

	It("it reads a message from the reader", func() {
		expected := &message.SessionDestroy{Session: 0xabcd}
		buf := bytes.Buffer{}
		message.Write(&buf, message.JSONEncoding, expected)

		socket := &mock.Socket{}
		socket.Impl.NextReader = func() (int, io.Reader, error) {
			return websocket.BinaryMessage, &buf, nil
		}

		msg, err := read(socket, message.JSONEncoding)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(msg).To(Equal(expected))
	})
})

var _ = Describe("write", func() {
	It("it returns an error if a writer can not be obtained", func() {
		expected := errors.New("<error>")
		socket := &mock.Socket{}
		socket.Impl.NextWriter = func(int) (io.WriteCloser, error) {
			return nil, expected
		}

		err := write(socket, nil, nil)

		Expect(err).To(Equal(err))
	})

	It("it writes a message to the writer", func() {
		expected := &message.SessionDestroy{Session: 0xabcd}
		buf := bytes.Buffer{}

		socket := &mock.Socket{}
		socket.Impl.NextWriter = func(int) (io.WriteCloser, error) {
			return nopCloser{&buf}, nil
		}

		err := write(socket, message.JSONEncoding, expected)

		Expect(err).ShouldNot(HaveOccurred())

		msg, err := message.Read(&buf, message.JSONEncoding)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(msg).To(Equal(expected))
	})
})

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
