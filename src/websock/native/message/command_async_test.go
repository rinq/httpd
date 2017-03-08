package message_test

import (
	"bytes"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("AsyncCall", func() {
	Describe("Accept", func() {
		It("invokes the correct visit method", func() {
			expected := errors.New("visit error")
			v := &mockVisitor{Error: expected}
			m := &AsyncCall{}

			err := m.Accept(v)

			Expect(err).To(Equal(expected))
			Expect(v.VisitedMessage).To(Equal(m))
		})
	})

	Describe("read", func() {
		It("decodes the message", func() {
			buf := []byte{
				'A', 'C',
				0xab, 0xcd, // session index
				0, 16, // header length
			}
			buf = append(buf, `["ns","cmd",456]`...)
			buf = append(buf, `"payload"`...)

			r := bytes.NewReader(buf)
			m, err := Read(r, JSONEncoding)

			Expect(err).ShouldNot(HaveOccurred())

			expected := &AsyncCall{
				Session: 0xabcd,
				Header: AsyncCallHeader{
					Namespace: "ns",
					Command:   "cmd",
					Timeout:   456 * time.Millisecond,
				},
				Payload: rinq.NewPayload("payload"),
			}
			Expect(m).To(Equal(expected))
		})
	})
})

var _ = Describe("AsyncSuccess", func() {
	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			p := rinq.NewPayload("payload")
			m := &AsyncSuccess{
				Session: 0xabcd,
				Header: AsyncSuccessHeader{
					Namespace: "ns",
					Command:   "cmd",
				},
				Payload: p,
			}

			err := Write(&buf, JSONEncoding, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'A', 'S',
				0xab, 0xcd, // session index
				0, 12, // header size
			}
			expected = append(expected, `["ns","cmd"]`...)
			expected = append(expected, `"payload"`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})

var _ = Describe("AsyncFailure", func() {
	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			p := rinq.NewPayload("payload")
			m := &AsyncFailure{
				Session: 0xabcd,
				Header: AsyncFailureHeader{
					Namespace:      "ns",
					Command:        "cmd",
					FailureType:    "fail-type",
					FailureMessage: "message",
				},
				Payload: p,
			}

			err := Write(&buf, JSONEncoding, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'A', 'F',
				0xab, 0xcd, // session index
				0, 34, // header size
			}
			expected = append(expected, `["ns","cmd","fail-type","message"]`...)
			expected = append(expected, `"payload"`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})

var _ = Describe("AsyncError", func() {
	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			m := &AsyncError{
				Session: 0xabcd,
				Header: AsyncErrorHeader{
					Namespace: "ns",
					Command:   "cmd",
				},
			}

			err := Write(&buf, JSONEncoding, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'A', 'E',
				0xab, 0xcd, // session index
				0, 12, // header size
			}
			expected = append(expected, `["ns","cmd"]`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})
