package native

import (
	"io"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/native/message"
)

var _ = Describe("protocol", func() {
	var (
		subject        websock.Protocol
		handle         handler
		handleCalls    uint
		handleSocket   websock.Socket
		handleEncoding message.Encoding
	)

	BeforeEach(func() {
		handleCalls = 0
		handleSocket = nil
		handleEncoding = nil

		handle = func(s websock.Socket, e message.Encoding) error {
			handleCalls++
			handleSocket = s
			handleEncoding = e
			return nil
		}

		subject = NewProtocol(handle)
	})

	Describe("Names", func() {
		It("returns cbor, then json", func() {
			Expect(subject.Names()).To(Equal(
				[]string{
					"rinq-1.0+cbor",
					"rinq-1.0+json",
				},
			))
		})
	})

	Describe("Handle", func() {
		DescribeTable(
			"invokes the handler with the correct encoding",
			func(p string, e message.Encoding) {
				socket := &mockSocket{subprotocol: p}
				subject.Handle(socket)

				Expect(handleCalls).To(BeEquivalentTo(1))
				Expect(handleSocket).To(Equal(socket))
				Expect(handleEncoding).To(Equal(e))
			},
			Entry("cbor", "rinq-1.0+cbor", message.CBOREncoding),
			Entry("json", "rinq-1.0+json", message.JSONEncoding),
		)
	})
})

type mockSocket struct {
	subprotocol string
}

func (s *mockSocket) Subprotocol() string                    { return s.subprotocol }
func (s *mockSocket) SetReadDeadline(t time.Time) error      { panic("not impl") }
func (s *mockSocket) SetPongHandler(h func(string) error)    { panic("not impl") }
func (s *mockSocket) NextReader() (int, io.Reader, error)    { panic("not impl") }
func (s *mockSocket) NextWriter(int) (io.WriteCloser, error) { panic("not impl") }
func (s *mockSocket) WriteMessage(mt int, data []byte) error { panic("not impl") }
