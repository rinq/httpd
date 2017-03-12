package native_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native"
	"github.com/rinq/httpd/src/websock/native/message"
)

var _ = Describe("handler", func() {
	var (
		subject *Handler
	)

	BeforeEach(func() {
		subject = &Handler{
			Encoding: message.JSONEncoding,
		}
	})

	Describe("Protocol", func() {
		It("returns the full protocol name", func() {
			Expect(subject.Protocol()).To(Equal("rinq-1.0+json"))
		})
	})
})
