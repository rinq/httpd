package websock_test

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock"
)

var _ = Describe("httpHandler", func() {
	var (
		handlerA, handlerB *mockHandler
		logger             *log.Logger

		request  *http.Request
		response responseRecorder

		subject http.Handler
	)

	BeforeEach(func() {
		handlerA = &mockHandler{protocol: "proto-a"}
		handlerB = &mockHandler{protocol: "proto-b"}

		request = httptest.NewRequest("GET", "/", nil)
		response = responseRecorder{httptest.NewRecorder()}

		subject = NewHTTPHandler(
			"*",
			time.Second,
			logger,
			handlerA,
			handlerB,
		)
	})

	BeforeEach(func() {
		request.Header = http.Header{}
		request.Header.Add("Connection", "upgrade")
		request.Header.Add("Upgrade", "websocket")
		request.Header.Add("Sec-Websocket-Version", "13")
		request.Header.Add("Sec-Websocket-Key", "<key>")
		request.Header.Add("Sec-Websocket-Protocol", "proto-b")
	})

	It("dispatches to the correct websocket handler", func() {
		subject.ServeHTTP(response, request)

		Expect(handlerA.called).To(BeFalse())
		Expect(handlerB.called).To(BeTrue())

		Expect(handlerB.connection).ToNot(BeNil())
		Expect(handlerB.request).To(Equal(request))
	})

	It("renders an error page when the request is not an upgrade", func() {
		request.Header.Set("Upgrade", "other")
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

type mockHandler struct {
	protocol   string
	err        error
	called     bool
	connection Connection
	request    *http.Request
}

func (h *mockHandler) Protocol() string {
	return h.protocol
}

func (h *mockHandler) Handle(c Connection, r *http.Request) error {
	h.called = true
	h.connection = c
	h.request = r

	return h.err
}
