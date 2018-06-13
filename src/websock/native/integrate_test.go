// +build !without_amqp

package native_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/httpd/src/websock/native"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/options"
	"github.com/rinq/rinq-go/src/rinqamqp"
	"github.com/satori/go.uuid"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

var _ = Describe("handler", func() {

	var (
		peer rinq.Peer

		req *http.Request
	)

	const (
		createSession uint16 = 'S'<<8 | 'C'

		callSync  uint16 = 'C'<<8 | 'C'
		callAsync uint16 = 'A'<<8 | 'C'
		callExec  uint16 = 'C'<<8 | 'X'

		session uint16 = 0xCAFE
	)

	BeforeEach(func() {
		req = httptest.NewRequest("GET", "/", nil)

		var err error
		peer, err = rinqamqp.DialEnv(options.Logger(rinq.NewLogger(false)))
		Expect(err).ToNot(HaveOccurred())
	})

	Context("integration testing the default settings", func() {

		const (
			nsBase      = "name-space"
			cmd         = "cmd"
			seq    uint = 1

			respBody = "pong"
		)

		var (
			ns string

			websocket *mockWebsock

			start chan struct{}
			kill  chan struct{}
		)

		BeforeEach(func() {

			ns = nsBase + uuid.NewV4().String()

			start = make(chan struct{})
			kill = make(chan struct{})

			websocket = &mockWebsock{
				start:       start,
				dead:        kill,
				serverResps: make(chan []byte),
			}

			go func() {
				defer GinkgoRecover()

				<-start
				handler := native.NewHandler(peer, message.JSONEncoding)
				err := handler.Handle(websocket, req)
				log.Println("got", err.Error(), ", handler closed")
			}()
		})

		AfterEach(func() {
			close(kill)
			peer.Stop()
			<-peer.Done()
		})

		Context("successful calls", func() {
			BeforeEach(func() {
				peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
					defer req.Payload.Close()
					res.Done(rinq.NewPayload(respBody))
				})
			})

			It("Sends a successful basic sync ping-pong message", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callSync, session, []interface{}{
					seq, ns, cmd, time.Second,
				}, "ping")

				close(start)

				resp := websocket.responseFromServer()

				expected := &message.SyncSuccess{}
				expected.Session = message.SessionIndex(session)
				expected.Seq = seq
				expected.Payload = rinq.NewPayload(respBody)

				expBytes := serializeServerResp(expected)
				Expect(resp).To(Equal(expBytes))
			})

			It("Sends a successful basic async ping-pong message", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callAsync, session, []interface{}{
					ns, cmd, time.Second,
				}, "payload!")

				close(start)

				resp := websocket.responseFromServer()

				expected := &message.AsyncSuccess{}
				expected.Session = message.SessionIndex(session)
				expected.Namespace = ns
				expected.Command = cmd
				expected.Payload = rinq.NewPayload(respBody)

				expBytes := serializeServerResp(expected)
				Expect(resp).To(Equal(expBytes))
			})

			It("Sends a successful basic exec message", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callExec, session, []interface{}{
					ns, cmd,
				}, nil)

				close(start)

				select {
				case <-websocket.serverResps:
					Fail("Received a response to an exec")
				case <-time.After(time.Second / 4):
				}
			})
		})

		Context("failure calls", func() {
			var (
				failureType = "failed"

				failureTemplate = "%s-%s"
				failureValues   = []interface{}{"namespace", "cmd"}
				failureMessage  = fmt.Sprintf(failureTemplate, failureValues...)
			)

			BeforeEach(func() {
				peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
					defer req.Payload.Close()
					res.Fail(failureType, failureTemplate, failureValues...)
				})
			})

			It("Sends a basic sync ping-pong message that fails", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callSync, session, []interface{}{
					seq, ns, cmd, time.Second,
				}, "ping")

				close(start)

				resp := websocket.responseFromServer()

				expected := &message.SyncFailure{}
				expected.Session = message.SessionIndex(session)
				expected.Seq = seq
				expected.FailureType = failureType
				expected.FailureMessage = failureMessage

				expBytes := serializeServerResp(expected)
				Expect(resp).To(Equal(expBytes))
			})

			It("Sends a basic async ping-pong message that fails", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callAsync, session, []interface{}{
					ns, cmd, time.Second,
				}, nil)

				close(start)

				resp := websocket.responseFromServer()

				expected := &message.AsyncFailure{}
				expected.Session = message.SessionIndex(session)
				expected.Namespace = ns
				expected.Command = cmd
				expected.FailureType = failureType
				expected.FailureMessage = failureMessage

				expBytes := serializeServerResp(expected)
				Expect(resp).To(Equal(expBytes))
			})

			It("Sends a basic exec message that fails", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callExec, session, []interface{}{
					ns, cmd,
				}, nil)

				close(start)

				select {
				case <-websocket.serverResps:
					Fail("Received a response to an exec")
				case <-time.After(time.Second / 4):
				}
			})
		})

		Context("error calls", func() {
			BeforeEach(func() {
				peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
					defer req.Payload.Close()
					res.Error(io.EOF)
				})
			})

			It("Sends a basic sync ping-pong message that fails", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callSync, session, []interface{}{
					seq, ns, cmd, time.Second,
				}, "ping")

				close(start)

				resp := websocket.responseFromServer()

				expected := &message.SyncError{}
				expected.Session = message.SessionIndex(session)
				expected.Seq = seq

				expBytes := serializeServerResp(expected)

				Expect(resp).To(Equal(expBytes))
			})

			It("Sends a basic async ping-pong message that fails", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callAsync, session, []interface{}{
					ns, cmd, time.Second,
				}, nil)

				close(start)

				resp := websocket.responseFromServer()

				expected := &message.AsyncError{}
				expected.Session = message.SessionIndex(session)
				expected.Namespace = ns
				expected.Command = cmd

				expBytes := serializeServerResp(expected)
				Expect(resp).To(Equal(expBytes))
			})

			It("Sends a basic exec message that fails", func() {
				websocket.requestFromClient(createSession, session, nil, nil)
				websocket.requestFromClient(callExec, session, []interface{}{
					ns, cmd,
				}, nil)

				close(start)

				select {
				case <-websocket.serverResps:
					Fail("Received a response to an exec")
				case <-time.After(time.Second / 4):
				}
			})
		})
	})

	Context("integration testing with a timeout", func() {

		var (
			subject *native.Handler
			ns      string

			websocket *mockWebsock

			start chan struct{}
			end   chan struct{}
		)

		const (
			nsBase      = "name-space"
			cmd         = "cmd"
			seq    uint = 1
		)

		BeforeEach(func() {

			ns = nsBase + uuid.NewV4().String()
			log.Println("listening on", ns)

			start = make(chan struct{})
			end = make(chan struct{})

			websocket = &mockWebsock{
				start:       start,
				dead:        end,
				serverResps: make(chan []byte),
			}

			go func() {
				defer GinkgoRecover()

				<-start

				err := subject.Handle(websocket, req)
				log.Println("got", err.Error(), ", handler closed")
			}()
		})

		AfterEach(func() {
			websocket = nil
			close(end)
		})

		It("limits a call to the server by the client timeout when the servers' timeout is longer", func() {

			deadline := make(chan time.Time)

			peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
				// never return anything

				dead, _ := ctx.Deadline()
				deadline <- dead
			})

			// client timeouts are in milliseconds
			clientTimeout := time.Duration(2000)
			serverTimeout := 10 * time.Second

			subject = native.NewHandler(peer, message.JSONEncoding, native.MaxCallTimeout(serverTimeout))

			websocket.requestFromClient(createSession, session, nil, nil)
			websocket.requestFromClient(callSync, session, []interface{}{
				seq, ns, cmd, clientTimeout,
			}, "ping")

			expectedTime := time.Now().Add(clientTimeout * time.Millisecond)

			close(start)

			Expect(<-deadline).To(BeTemporally("~", expectedTime, time.Second/2))
		})

		It("limits a call to the server by the server timeout when the clients' timeout is longer", func() {
			deadline := make(chan time.Time)

			peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
				// never return anything

				dead, _ := ctx.Deadline()
				deadline <- dead
			})

			// client timeouts are in milliseconds
			clientTimeout := time.Duration(10000)
			serverTimeout := time.Duration(2000) * time.Millisecond

			subject = native.NewHandler(peer, message.JSONEncoding, native.MaxCallTimeout(serverTimeout))

			websocket.requestFromClient(createSession, session, nil, nil)
			websocket.requestFromClient(callSync, session, []interface{}{
				seq, ns, cmd, clientTimeout,
			}, "ping")

			expectedTime := time.Now().Add(serverTimeout)

			close(start)

			Expect(<-deadline).To(BeTemporally("~", expectedTime, time.Second/2))
		})

		It("limits a call to the server by the clients' timeout when the servers' timeout is not set", func() {
			deadline := make(chan time.Time)

			peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
				// never return anything

				dead, _ := ctx.Deadline()
				deadline <- dead
			})

			// client timeouts are in milliseconds
			clientTimeout := time.Duration(10000)

			subject = native.NewHandler(peer, message.JSONEncoding)

			websocket.requestFromClient(createSession, session, nil, nil)
			websocket.requestFromClient(callSync, session, []interface{}{
				seq, ns, cmd, clientTimeout,
			}, "ping")

			expectedTime := time.Now().Add(clientTimeout * time.Millisecond)

			close(start)

			Expect(<-deadline).To(BeTemporally("~", expectedTime, time.Second/2))
		})

	})
})

func serializeServerResp(outgoing message.Outgoing) []byte {
	expBytes := new(bytes.Buffer)
	err := message.Write(expBytes, message.JSONEncoding, outgoing)
	Expect(err).NotTo(HaveOccurred())
	return expBytes.Bytes()
}

func constructClientReq(msgType uint16, session uint16, headers interface{}, payload interface{}) []byte {

	buff := new(bytes.Buffer)

	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[:2], msgType)
	binary.BigEndian.PutUint16(b[2:], session)

	buff.Write(b)

	if headers != nil {
		if err := message.JSONEncoding.EncodeHeader(buff, headers); err != nil {
			panic(err)
		}

		p := rinq.NewPayload(payload)
		if err := message.JSONEncoding.EncodePayload(buff, p); err != nil {
			panic(err)
		}
	}

	return buff.Bytes()
}

type mockWebsock struct {
	start <-chan struct{}
	dead  <-chan struct{}

	clientReqs  []io.Reader
	serverResps chan []byte
}

func (m *mockWebsock) requestFromClient(msgType uint16, session uint16, headers interface{}, payload interface{}) {
	out := constructClientReq(msgType, session, headers, payload)
	m.clientReqs = append(m.clientReqs, bytes.NewBuffer(out))
}

func (m *mockWebsock) NextReader() (out io.Reader, err error) {
	<-m.start

	if len(m.clientReqs) == 0 {
		<-m.dead
		return nil, context.Canceled
	}

	out, m.clientReqs = m.clientReqs[0], m.clientReqs[1:]
	return out, nil
}

func (m *mockWebsock) responseFromServer() []byte {
	select {
	case <-m.dead:
		return nil
	case b := <-m.serverResps:
		return b
	}
}

func (m *mockWebsock) NextWriter() (out io.WriteCloser, err error) {
	<-m.start

	b := wcByteBuff{Buffer: new(bytes.Buffer)}

	w := make(chan struct{})
	go func() {
		<-w
		m.serverResps <- b.Buffer.Bytes()
	}()

	b.cls = func() {
		close(w)
	}

	return &b, err
}

type wcByteBuff struct {
	*bytes.Buffer
	cls func()
}

func (w *wcByteBuff) Close() error {
	w.cls()
	return nil
}
