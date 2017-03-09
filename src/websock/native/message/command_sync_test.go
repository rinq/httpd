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

var _ = Describe("SyncCall", func() {
	Describe("Accept", func() {
		It("invokes the correct visit method", func() {
			expected := errors.New("visit error")
			v := &mockVisitor{Error: expected}
			m := &SyncCall{}

			err := m.Accept(v)

			Expect(err).To(Equal(expected))
			Expect(v.VisitedMessage).To(Equal(m))
		})
	})

	Describe("read", func() {
		It("decodes the message", func() {
			buf := []byte{
				'C', 'C',
				0xab, 0xcd, // session index
				0, 20, // header length
			}
			buf = append(buf, `[123,"ns","cmd",456]`...)
			buf = append(buf, `"payload"`...)

			r := bytes.NewReader(buf)
			m, err := Read(r, JSONEncoding)

			Expect(err).ShouldNot(HaveOccurred())

			expected := &SyncCall{
				Session: 0xabcd,
				Header: SyncCallHeader{
					Seq:       123,
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

var _ = Describe("SyncSuccess", func() {
	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			p := rinq.NewPayload("payload")
			m := &SyncSuccess{
				Session: 0xabcd,
				Header: SyncSuccessHeader{
					Seq: 123,
				},
				Payload: p,
			}

			err := Write(&buf, JSONEncoding, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'C', 'S',
				0xab, 0xcd, // session index
				0, 5, // header size
			}
			expected = append(expected, `[123]`...)
			expected = append(expected, `"payload"`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})

var _ = Describe("SyncFailure", func() {
	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			p := rinq.NewPayload("payload")
			m := &SyncFailure{
				Session: 0xabcd,
				Header: SyncFailureHeader{
					Seq:            123,
					FailureType:    "fail-type",
					FailureMessage: "message",
				},
				Payload: p,
			}

			err := Write(&buf, JSONEncoding, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'C', 'F',
				0xab, 0xcd, // session index
				0, 27, // header size
			}
			expected = append(expected, `[123,"fail-type","message"]`...)
			expected = append(expected, `"payload"`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})

var _ = Describe("SyncError", func() {
	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			m := &SyncError{
				Session: 0xabcd,
				Header: SyncErrorHeader{
					Seq: 123,
				},
			}

			err := Write(&buf, JSONEncoding, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'C', 'E',
				0xab, 0xcd, // session index
				0, 5, // header size
			}
			expected = append(expected, `[123]`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})

var _ = Describe("NewSyncResponse", func() {
	It("returns a success response if err is nil", func() {
		p := rinq.NewPayload(456)
		m, ok := NewSyncResponse(0xabcd, 123, p, nil)

		Expect(m).To(Equal(&SyncSuccess{
			Session: 0xabcd,
			Header:  SyncSuccessHeader{Seq: 123},
			Payload: p,
		}))

		Expect(ok).To(BeTrue())
	})

	It("returns a failure response if err is a failure", func() {
		p := rinq.NewPayload(456)
		err := rinq.Failure{
			Type:    "type",
			Message: "message",
			Payload: p,
		}

		m, ok := NewSyncResponse(0xabcd, 123, p, err)

		Expect(m).To(Equal(&SyncFailure{
			Session: 0xabcd,
			Header: SyncFailureHeader{
				Seq:            123,
				FailureType:    "type",
				FailureMessage: "message",
			},
			Payload: p,
		}))

		Expect(ok).To(BeTrue())
	})

	It("returns an error response if err is a command error", func() {
		err := rinq.CommandError("error")
		m, ok := NewSyncResponse(0xabcd, 123, nil, err)

		Expect(m).To(Equal(&SyncError{
			Session: 0xabcd,
			Header:  SyncErrorHeader{Seq: 123},
		}))

		Expect(ok).To(BeTrue())
	})

	It("returns false for other errors", func() {
		err := errors.New("error")
		m, ok := NewSyncResponse(0xabcd, 123, nil, err)

		Expect(m).To(BeNil())
		Expect(ok).To(BeFalse())
	})
})
