package native

import (
	"bytes"
	"errors"
	"io"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/httpd/src/websock/internal/mock"
	"github.com/rinq/httpd/src/websock/native/message"
)

// // write an outgoing message to the websocket.
// func write(s websock.Socket, e message.Encoding, m message.Outgoing) error {
// 	w, err := s.NextWriter(websocket.BinaryMessage)
// 	if err != nil {
// 		return err
// 	}
// 	defer w.Close()
//
// 	return message.Write(w, e, m)
// }

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

	It("it writes a message to the reader", func() {
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
