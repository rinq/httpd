package websock_test

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock"
)

var _ = Describe("Handler", func() {
	var (
		protoA, protoB *mockProtocol
		protos         *ProtocolSet
		logger         *log.Logger

		request  *http.Request
		response responseRecorder

		subject *Handler
	)

	BeforeEach(func() {
		protoA = &mockProtocol{name: "proto-a"}
		protoB = &mockProtocol{name: "proto-b"}
		protos = NewProtocolSet(protoA, protoB)

		request = httptest.NewRequest("GET", "/", nil)
		response = responseRecorder{httptest.NewRecorder()}

		subject = NewHandler("*", protos, logger)
	})

	It("dispatches with the correct protocol", func() {
		request.Header = http.Header{}
		request.Header.Add("Connection", "upgrade")
		request.Header.Add("Upgrade", "websocket")
		request.Header.Add("Sec-Websocket-Version", "13")
		request.Header.Add("Sec-Websocket-Key", "<key>")
		request.Header.Add("Sec-Websocket-Protocol", "proto-b")

		subject.ServeHTTP(response, request)

		Expect(protoA.handleCalled).To(BeFalse())
		Expect(protoB.handleCalled).To(BeTrue())
	})

	It("renders an error page when the request is not an upgrade", func() {
		subject.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body).To(ContainSubstring("Bad Request"))
	})
})

type mockConnection struct {
	net.Conn
}

func (*mockConnection) Write(b []byte) (int, error) { return len(b), nil }
func (*mockConnection) SetDeadline(time.Time) error { return nil }
func (*mockConnection) Close() error                { return nil }

type responseRecorder struct {
	*httptest.ResponseRecorder
}

func (responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r := bufio.NewReader(&bytes.Buffer{})
	w := bufio.NewWriter(&bytes.Buffer{})
	return &mockConnection{}, bufio.NewReadWriter(r, w), nil
}

type mockProtocol struct {
	name         string
	handleCalled bool
}

func (p *mockProtocol) Names() []string        { return []string{p.name} }
func (p *mockProtocol) Handle(*websocket.Conn) { p.handleCalled = true }
