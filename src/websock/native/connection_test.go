package native

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("connection", func() {
	var (
		subject *connection
	)

	BeforeEach(func() {
		subject = &connection{}
	})

	Describe("Close", func() {

	})

	Describe("VisitSessionCreate", func() {
		// msg := &message.SessionCreate{Session: 0xabcd}

		XIt("returns an error if the session index is already in use", func() {
		})
	})

	Describe("VisitSessionDestroy", func() {
		msg := &message.SessionDestroy{Session: 0xabcd}

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitSessionDestroy(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitSyncCall", func() {
		msg := &message.SyncCall{
			Session: 0xabcd,
			Header: message.SyncCallHeader{
				Seq:       123,
				Namespace: "ns",
				Command:   "cmd",
				Timeout:   456 * time.Millisecond,
			},
			Payload: rinq.NewPayload("payload"),
		}

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitSyncCall(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitAsyncCall", func() {
		msg := &message.AsyncCall{
			Session: 0xabcd,
			Header: message.AsyncCallHeader{
				Namespace: "ns",
				Command:   "cmd",
				Timeout:   456 * time.Millisecond,
			},
			Payload: rinq.NewPayload("payload"),
		}

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitAsyncCall(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitExecute", func() {
		msg := &message.Execute{
			Session: 0xabcd,
			Header: message.ExecuteHeader{
				Namespace: "ns",
				Command:   "cmd",
			},
			Payload: rinq.NewPayload("payload"),
		}

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitExecute(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})
})
