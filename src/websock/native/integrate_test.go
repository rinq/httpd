// +build !without_amqp

package native_test

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/httpd/src/websock/native"
	"github.com/rinq/httpd/src/websock/native/message"
	"io"
	"bytes"
	"net/http"
	"net/http/httptest"
	"github.com/rinq/rinq-go/src/rinqamqp"
	"context"
	"log"
	"time"
	"encoding/binary"
	"fmt"
)

var _ = Describe("handler", func() {
	var (
		subject native.Handler

		peer rinq.Peer

		websocket *mockWebsock

		start func()
		kill context.CancelFunc

		req *http.Request
	)

	const (
		createSession uint16 = 'S'<<8 | 'C'

		callSync uint16 = 'C'<<8 | 'C'
		callSyncSuccess uint16 = 'C'<<8 | 'S'
		callSyncFailure uint16 = 'C'<<8 | 'F'
		callSyncError  uint16 = 'C'<<8 | 'E'

		callAsync uint16 = 'A'<<8 | 'C'
		callAsyncSuccess uint16 = 'A'<<8 | 'S'
		callAsyncFailure uint16 = 'A'<<8 | 'F'
		callAsyncError   uint16 = 'A'<<8 | 'E'

		callExec uint16 = 'C'<<8 | 'X'

		session uint16 = 0xCAFE
	)


	BeforeEach(func() {
		var err error

		subject = native.NewHandler(peer, message.JSONEncoding)

		var killCtx context.Context
		killCtx, kill = context.WithTimeout(context.Background(), 500 * time.Millisecond)

		var startCtx context.Context
		startCtx, start = context.WithCancel(context.Background())

		websocket = &mockWebsock{
			ctx: killCtx,
			start: startCtx.Done(),
			wIn: make(chan []byte),
		}

		req = httptest.NewRequest("GET", "/", nil)

		peer, err = rinqamqp.DialEnv()
		Expect(err).ToNot(HaveOccurred())

		go func() {
			defer GinkgoRecover()

			select {
			case <-peer.Done():
				Expect(peer.Err()).NotTo(HaveOccurred())
				Fail("not expected to be here during the test")

			case <-killCtx.Done():
				// normal operation
			}

		}()

	})

	AfterEach(func() {
		kill()

		peer.Stop()
		<-peer.Done()
	})

	Context("integration testing the default settings", func() {

		const (
			ns = "name-space"
			cmd = "cmd"
			seq uint = 1

			respBody = "pong"
		)

		BeforeEach(func() {
			subject = native.NewHandler(peer, message.JSONEncoding)

			go func() {
				defer GinkgoRecover()

				err := subject.Handle(websocket, req)
				log.Println("got", err.Error(), ", handler closed")
			}()
		})

		Context("successful calls", func() {
			BeforeEach(func() {
				peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
					defer req.Payload.Close()
					res.Done(rinq.NewPayload(respBody))
				})
			})

			It("Sends a successful basic sync ping-pong message", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callSync, session, []interface{}{
					seq, ns, cmd, time.Second,
				}, "ping")

				start()

				resp := websocket.getMsg()

				expected := msg(callSyncSuccess, session, []interface{}{seq}, respBody)

				Expect(resp).To(Equal(expected))
			})

			It("Sends a successful basic async ping-pong message", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callAsync, session, []interface{}{
					ns, cmd, time.Second,
				}, "payload!")

				start()

				resp := websocket.getMsg()

				expected := msg(callAsyncSuccess, session, []interface{}{ns, cmd}, respBody)

				Expect(resp).To(Equal(expected))
			})

			It("Sends a successful basic exec message", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callExec, session, []interface{}{
					ns, cmd,
				}, "payload!")

				start()

				msg := websocket.getMsg()
				Expect(msg).To(BeNil())
			})
		})

		Context("failure calls", func() {
			var (
				failureMessageStatic = "failed"

				failureMessageTemplate = "%s-%s"
				failureMessageValues = []interface{}{ "namespace", "cmd"}
				failureMessageResolved = fmt.Sprintf(failureMessageTemplate, failureMessageValues...)
			)
			BeforeEach(func() {
				peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
					defer req.Payload.Close()
					res.Fail(failureMessageStatic, failureMessageTemplate, failureMessageValues...)
				})
			})

			It("Sends a basic sync ping-pong message that fails", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callSync, session, []interface{}{
					seq, ns, cmd, time.Second,
				}, "ping")

				start()

				resp := websocket.getMsg()

				expected := msg(callSyncFailure, session, []interface{}{
					seq, failureMessageStatic, failureMessageResolved,
				}, nil)

				Expect(resp).To(Equal(expected))
			})

			It("Sends a basic async ping-pong message that fails", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callAsync, session, []interface{}{
					ns, cmd, time.Second,
				}, nil)

				start()

				resp := websocket.getMsg()

				expected := msg(callAsyncFailure, session, []interface{}{
					ns, cmd, failureMessageStatic, failureMessageResolved,
				}, nil)

				log.Println(string(resp))
				log.Println(string(expected))

				Expect(resp).To(Equal(expected))
			})

			It("Sends a basic exec message that fails", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callExec, session, []interface{}{
					ns, cmd,
				}, nil)

				start()

				msg := websocket.getMsg()
				Expect(msg).To(BeNil())
			})
		})

		XContext("error calls", func() {
			BeforeEach(func() {
				peer.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
					defer req.Payload.Close()
					res.Error(io.EOF)
				})
			})

			It("Sends a basic sync ping-pong message that fails", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callSync, session, []interface{}{
					seq, ns, cmd, time.Second,
				}, "ping")

				start()

				resp := websocket.getMsg()

				expected := msg(callSyncFailure, session, []interface{}{
					seq, failureMessageStatic, failureMessageResolved,
				}, nil)

				Expect(resp).To(Equal(expected))
			})

			It("Sends a basic async ping-pong message that fails", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callAsync, session, []interface{}{
					ns, cmd, time.Second,
				}, nil)

				start()

				resp := websocket.getMsg()

				expected := msg(callAsyncFailure, session, []interface{}{
					ns, cmd, failureMessageStatic, failureMessageResolved,
				}, nil)

				log.Println(string(resp))
				log.Println(string(expected))

				Expect(resp).To(Equal(expected))
			})

			It("Sends a basic exec message that fails", func() {
				websocket.queueMsg(createSession, session, nil, nil)
				websocket.queueMsg(callExec, session, []interface{}{
					ns, cmd,
				}, nil)

				start()

				msg := websocket.getMsg()
				Expect(msg).To(BeNil())
			})
		})

	})
})

func msg(msgType uint16, session uint16, headers interface{}, payload interface{}) []byte {

	buff := new(bytes.Buffer)

	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[:2], msgType)
	binary.BigEndian.PutUint16(b[2:], session)

	buff.Write(b)

	if headers != nil {
		if err := message.JSONEncoding.EncodeHeader(buff, headers); err != nil {
			panic(err)
		}
	}

	if headers != nil {
		p := rinq.NewPayload(payload)
		if err := message.JSONEncoding.EncodePayload(buff, p); err != nil {
			panic(err)
		}
	}

	return buff.Bytes()
}

type mockWebsock struct {
	ctx context.Context
	start <-chan struct{}

	rOut []io.Reader
	wIn chan []byte
}

func (m *mockWebsock) queueMsg(msgType uint16, session uint16, headers interface{}, payload interface{}) {
	out := msg(msgType, session, headers, payload)
	(*m).rOut = append((*m).rOut, bytes.NewBuffer(out))
}


func (m *mockWebsock) NextReader() (out io.Reader, err error) {
	<-m.start

	if len(m.rOut) == 0 {
		<-m.ctx.Done()
		return nil, m.ctx.Err()
	}

	out, m.rOut = m.rOut[0], m.rOut[1:]
	return out, nil
}

func (m *mockWebsock) getMsg() []byte {
	select {
	case <-m.ctx.Done():
		return nil
	case b := <-m.wIn:
			return b
	}
}

func (m *mockWebsock) NextWriter() (out io.WriteCloser, err error) {
	<-m.start

	b := wcByteBuff{Buffer: new(bytes.Buffer)}

	w := make(chan struct{})
	go func() {
		<-w
		m.wIn<-b.Buffer.Bytes()
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

func (w * wcByteBuff) Close() error {
	w.cls()
	return nil
}