package native

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinqamqp"
)

var _ = Describe("visitor", func() {
	var (
		sent []message.Outgoing
		send func(message.Outgoing)

		subject *visitor

		peer rinq.Peer
	)

	BeforeSuite(func() {
		var err error

		peer, err = rinqamqp.DialEnv()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		peer.Stop()
		<-peer.Done()
	})

	BeforeEach(func() {
		sent = nil
		send = func(m message.Outgoing) {
			sent = append(sent, m)
		}

		subject = newVisitor(context.Background(), peer, nil, send)
	})

	Describe("VisitSessionCreate", func() {
		msg := &message.SessionCreate{}
		msg.Session = 0xabcd

		XIt("returns an error if the session index is already in use", func() {
		})
	})

	Describe("VisitSessionDestroy", func() {
		msg := &message.SessionDestroy{}
		msg.Session = 0xabcd

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitSessionDestroy(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitListen", func() {
		msg := &message.Listen{}
		msg.Session = 0xabcd

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitListen(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitUnlisten", func() {
		msg := &message.Unlisten{}
		msg.Session = 0xabcd

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitUnlisten(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitSyncCall", func() {
		msg := &message.SyncCall{}
		msg.Session = 0xabcd
		msg.Seq = 123
		msg.Namespace = "ns"
		msg.Command = "cmd"
		msg.Timeout = 456 * time.Hour
		msg.Payload = rinq.NewPayload("payload")

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitSyncCall(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitAsyncCall", func() {
		msg := &message.AsyncCall{}
		msg.Session = 0xabcd
		msg.Namespace = "ns"
		msg.Command = "cmd"
		msg.Timeout = 456 * time.Millisecond
		msg.Payload = rinq.NewPayload("payload")

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitAsyncCall(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})

	Describe("VisitExecute", func() {
		msg := &message.Execute{}
		msg.Session = 0xabcd
		msg.Namespace = "ns"
		msg.Command = "cmd"
		msg.Payload = rinq.NewPayload("payload")

		It("returns an error if the session index is not in use", func() {
			err := subject.VisitExecute(msg)
			Expect(err).To(MatchError("session 43981 does not exist"))
		})
	})
})
