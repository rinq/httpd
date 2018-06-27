package websock_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/internal/mock"
)

var _ = Describe("httpHandler", func() {
	var (
		handlerA, handlerB *mock.Handler
		subject            http.Handler
		server             *httptest.Server
	)

	BeforeEach(func() {
		handlerA = &mock.Handler{}
		handlerA.Impl.Protocol = "proto-a"

		handlerB = &mock.Handler{}
		handlerB.Impl.Protocol = "proto-b"

		subject = NewHTTPHandler([]Handler{
			handlerA,
			handlerB,
		})

		server = httptest.NewServer(subject)
	})

	AfterEach(func() {
		server.Close()
	})

	It("dispatches based on sub-protocol", func() {
		handlerA.Impl.Handle = func(Connection, *http.Request) error {
			panic("wrong handler")
		}

		barrier := make(chan bool, 1)
		handlerB.Impl.Handle = func(Connection, *http.Request) error {
			barrier <- true
			return nil
		}

		url := strings.Replace(server.URL, "http://", "ws://", 1)
		d := websocket.Dialer{Subprotocols: []string{"proto-b"}}
		con, _, err := d.Dial(url, nil)
		if con != nil {
			defer con.Close()
		}

		Expect(err).ShouldNot(HaveOccurred())

		select {
		case <-barrier:
		case <-time.After(time.Second):
			panic("timeout")
		}
	})

	It("closes the connection if the sub-protocol is not supported", func() {
		url := strings.Replace(server.URL, "http://", "ws://", 1)
		d := websocket.Dialer{Subprotocols: []string{"unsupported-protocol"}}

		con, _, err := d.Dial(url, nil)
		if con != nil {
			defer con.Close()
		}

		Expect(err).ShouldNot(HaveOccurred())

		timeout := time.After(time.Second)

		for {
			select {
			case <-timeout:
				panic("timeout")
			default:
				_, _, err = con.ReadMessage()
				if err != nil {
					e := err.(*websocket.CloseError)
					Expect(e.Code).To(Equal(websocket.CloseProtocolError))
					return
				}
			}
		}
	})

	It("renders an error page when the request is not an upgrade", func() {
		r, err := http.Get(server.URL)
		if r != nil {
			defer r.Body.Close()
		}

		Expect(err).ShouldNot(HaveOccurred())
		body, _ := ioutil.ReadAll(r.Body)

		Expect(r.StatusCode).To(Equal(http.StatusBadRequest))
		Expect(body).To(ContainSubstring("Bad Request"))
	})
})
