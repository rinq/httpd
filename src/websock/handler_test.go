package websock_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"context"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock"
	"github.com/rinq/httpd/src/websock/internal/mock"
)

var _ = Describe("httpHandler", func() {

	Context("when no configuration options have been set", func() {
		var (
			subject            http.Handler
			server             *httptest.Server
			handlerA, handlerB *mock.Handler
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

			con := wsClientFor(server, "proto-b")
			defer con.Close()

			select {
			case <-barrier:
			case <-time.After(time.Second):
				panic("timeout")
			}
		})

		It("closes the connection if the sub-protocol is not supported", func() {
			con := wsClientFor(server, "unsupported-protocol")
			defer con.Close()

			timeout := time.After(time.Second)

			for {
				select {
				case <-timeout:
					panic("timeout")
				default:
					_, _, err := con.ReadMessage()
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

	Context("when origin limits have been set", func() {
		var server *httptest.Server

		BeforeEach(func() {
			server, _ = startHTTPHander(LimitToOrigin("*.cats.com:80"))
		})

		AfterEach(func() {
			server.Close()
		})

		It("accepts requests from the given origin", func() {
			url := strings.Replace(server.URL, "http://", "ws://", 1)
			d := websocket.Dialer{Subprotocols: []string{"proto-b"}}
			_, _, err := d.Dial(url, http.Header{
				"Origin": []string{"http://my.cats.com:80"},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("accepts requests from the given origin", func() {
			url := strings.Replace(server.URL, "http://", "ws://", 1)
			d := websocket.Dialer{Subprotocols: []string{"proto-b"}}
			_, _, err := d.Dial(url, http.Header{
				"Origin": []string{"http://my.dogs.com:80"},
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when ping intervals have been set", func() {

		It("sends a ping request at perodically according to the interval", func() {

			timeUntilRequest := 115 * time.Millisecond
			pingTimeout := 20 * time.Millisecond

			server, handle := startHTTPHander(PingInterval(pingTimeout))
			defer server.Close()

			handle.Impl.Handle = func(cn Connection, request *http.Request) error {
				<-time.After(timeUntilRequest)

				w, err := cn.NextWriter()
				Expect(err).ToNot(HaveOccurred())
				// write a message
				w.Write([]byte("cats"))
				return nil
			}

			rawConn := wsClientFor(server, handle.Protocol())
			defer rawConn.Close()

			pingNotify := make(chan bool)
			rawConn.SetPingHandler(func(message string) error {
				pingNotify <- false
				return nil
			})

			// try and read a msg
			go rawConn.NextReader()

			// ensure we receive at least 115 / 20 times
			repeat := int(timeUntilRequest / pingTimeout)
			for i := 0; i < repeat; i++ {
				Eventually(pingNotify).Should(Receive(), "failed to receive ping %v of %v", i, repeat)
			}
		})

	})

	Context("when max message size has been set", func() {
		It("kills all connections that try to send a message that is larger than the max size", func() {
			server, handle := startHTTPHander(MaxMessageSize(10))
			defer server.Close()

			rawConn := wsClientFor(server, handle.Protocol())
			defer rawConn.Close()

			w, err := rawConn.NextWriter(websocket.TextMessage)
			Expect(err).ToNot(HaveOccurred())

			_, err = w.Write([]byte("this message is longer than ten bytes"))
			Expect(err).ToNot(HaveOccurred())

			_, _, err = rawConn.NextReader()
			Expect(websocket.IsUnexpectedCloseError(err)).To(BeTrue())
		})
	})

	Context("when the max number of concurrent calls have been set", func() {
		It("blocks requests from a single connection when they've exhausted their max concurrent calls", func() {
			perConnCap := 1
			server, handle := startHTTPHander(MaxConcurrentCalls(perConnCap, 2))
			defer server.Close()

			gotCapacityNotify, connNotify, cleanupNotify := make(chan bool, 10), make(chan struct{}), make(chan struct{})

			handle.Impl.Handle = func(cn Connection, req *http.Request) error {

				for i := 0; i < 10; i++ {
					// queue up a bunch of "requests"
					go func() {
						cn.ReserveCapacity(context.Background())
						gotCapacityNotify <- true

						// wait until we're allowed to release the capacity
						<-cleanupNotify
						cn.ReleaseCapacity()
					}()
				}

				close(connNotify)

				return nil
			}

			rawConn := wsClientFor(server, handle.Protocol())
			defer rawConn.Close()

			<-connNotify

			// should only have perConnCap items in the channel at any time
			Expect(len(gotCapacityNotify)).To(Equal(perConnCap))

			close(cleanupNotify)
		})

		It("blocks requests from all connections when the servers' max concurrent calls have been exhausted", func() {

			globalCap := 1
			server, handle := startHTTPHander(MaxConcurrentCalls(10, globalCap))
			defer server.Close()

			gotCapacityNotify, connNotify, cleanupNotify := make(chan bool, 10), make(chan bool, 2), make(chan struct{})

			handle.Impl.Handle = func(cn Connection, req *http.Request) error {

				for i := 0; i < 10; i++ {
					// queue up a bunch of "requests"
					go func() {
						cn.ReserveCapacity(context.Background())
						gotCapacityNotify <- true

						// wait until we're allowed to release the capacity
						<-cleanupNotify
						cn.ReleaseCapacity()
					}()
				}

				connNotify <- true

				return nil
			}

			//startup two conns
			defer wsClientFor(server, handle.Protocol()).Close()
			defer wsClientFor(server, handle.Protocol()).Close()

			// accept the two conns
			<-connNotify
			<-connNotify

			// should only have perConnCap items in the channel at any time
			Expect(len(gotCapacityNotify)).To(Equal(globalCap))

			close(cleanupNotify)
		})

	})
})

func startHTTPHander(option ...Option) (*httptest.Server, *mock.Handler) {
	handle := &mock.Handler{}
	handle.Impl.Protocol = "protocol"

	subject := NewHTTPHandler([]Handler{
		handle,
	}, option...)

	server := httptest.NewServer(subject)

	return server, handle
}

func wsClientFor(server *httptest.Server, protocols ...string) *websocket.Conn {
	url := strings.Replace(server.URL, "http://", "ws://", 1)
	d := websocket.Dialer{Subprotocols: protocols}
	con, _, err := d.Dial(url, nil)
	Expect(err).ShouldNot(HaveOccurred())

	return con
}
