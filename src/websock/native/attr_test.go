package native

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("sessionAttributes", func() {
	It("includes an attribute containing the host", func() {
		request := httptest.NewRequest("GET", "/", nil)
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze("host", "example.com"),
		))
	})

	It("includes an attribute containing the remote address", func() {
		request := httptest.NewRequest("GET", "/", nil)
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze("remote-addr", "192.0.2.1"),
		))
	})

	It("supports remote addresses without ports", func() {
		request := httptest.NewRequest("GET", "/", nil)
		request.RemoteAddr = "192.0.2.2"
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze("remote-addr", "192.0.2.2"),
		))
	})

	It("uses the X-Forwarded-For header when present", func() {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Add("X-Forwarded-For", "10.1.1.1,10.2.2.2")
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze("remote-addr", "10.1.1.1"),
		))
	})
})
