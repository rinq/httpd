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
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("httpHandler", func() {
	var (
		handlerA, handlerB *mockHandler
		peer               rinq.Peer
		config             Config
		logger             *log.Logger

		request  *http.Request
		response responseRecorder

		subject http.Handler
	)

	BeforeEach(func() {
		handlerA = &mockHandler{protocol: "proto-a"}
		handlerB = &mockHandler{protocol: "proto-b"}

		peer = &mockPeer{}

		request = httptest.NewRequest("GET", "/", nil)
		response = responseRecorder{httptest.NewRecorder()}

		subject = NewHTTPHandler(
			func() (rinq.Peer, bool) {
				return peer, true
			},
			config,
			logger,
			handlerA,
			handlerB,
		)
	})

	Context("when the request is a websocket", func() {
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

			Expect(handlerB.peer).To(Equal(peer))
			Expect(handlerB.config).To(Equal(&config))
		})

		It("supplies an attribute containing the host", func() {
			subject.ServeHTTP(response, request)

			Expect(handlerB.attrs).To(ContainElement(
				rinq.Freeze("rinq.httpd.host", "example.com"),
			))
		})

		It("supplies an attribute containing the remote address", func() {
			subject.ServeHTTP(response, request)

			Expect(handlerB.attrs).To(ContainElement(
				rinq.Freeze("rinq.httpd.remote-addr", "192.0.2.1"),
			))
		})

		It("supports remote addresses without ports", func() {
			request.RemoteAddr = "192.0.2.2"
			subject.ServeHTTP(response, request)

			Expect(handlerB.attrs).To(ContainElement(
				rinq.Freeze("rinq.httpd.remote-addr", "192.0.2.2"),
			))
		})

		It("uses the X-Forwarded-For header when present", func() {
			request.Header.Add("X-Forwarded-For", "10.1.1.1,10.2.2.2")
			subject.ServeHTTP(response, request)

			Expect(handlerB.attrs).To(ContainElement(
				rinq.Freeze("rinq.httpd.remote-addr", "10.1.1.1"),
			))
		})
	})

	It("renders an error page when the request is not an upgrade", func() {
		subject.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body).To(ContainSubstring("Bad Request"))
	})

	It("renders an error page when the peer is not available", func() {
		subject = NewHTTPHandler(
			func() (rinq.Peer, bool) {
				return nil, false
			},
			config,
			logger,
			handlerA,
			handlerB,
		)

		subject.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(response.Body).To(ContainSubstring("Service Unavailable"))
	})
})

type mockPeer struct {
	rinq.Peer
}

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
	protocol string
	called   bool
	config   *Config
	peer     rinq.Peer
	attrs    []rinq.Attr
}

func (h *mockHandler) Protocol() string {
	return h.protocol
}

func (h *mockHandler) Handle(s Socket, c Config, p rinq.Peer, a []rinq.Attr) {
	h.called = true
	h.config = &c
	h.peer = p
	h.attrs = a
}
