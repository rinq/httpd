package native

import (
	"net/http/httptest"

	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"net"
	"net/http"
)

var _ = Describe("sessionAttributes", func() {
	It("includes an attribute containing the host", func() {
		request := httptest.NewRequest("GET", "/", nil)
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze(HttpdAttrHost, "example.com"),
		))
	})

	It("includes an attribute containing the client IP", func() {
		request := httptest.NewRequest("GET", "/", nil)
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze(HttpdAttrClientIP, "192.0.2.1"),
		))
	})

	It("includes an attribute containing the remote address", func() {
		request := httptest.NewRequest("GET", "/", nil)

		const addr = "192.0.2.1:9981"
		request.RemoteAddr = addr

		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze(HttpdAttrRemoteAddr, addr),
		))
	})

	It("includes an attribute containing the local address", func() {
		const addr = "192.0.2.1:9981"

		ctx, err := net.ResolveTCPAddr("tcp", addr)
		Expect(err).NotTo(HaveOccurred())

		request := httptest.NewRequest("GET", "/", nil)
		request = request.WithContext(context.WithValue(context.Background(), http.LocalAddrContextKey, ctx))

		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze(HttpdAttrLocalAddr, addr),
		))
	})

	It("supports remote addresses without ports", func() {
		request := httptest.NewRequest("GET", "/", nil)
		request.RemoteAddr = "192.0.2.2"
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze(HttpdAttrClientIP, "192.0.2.2"),
		))
	})

	It("uses the X-Forwarded-For header when present", func() {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Add("X-Forwarded-For", "10.1.1.1,10.2.2.2")
		attrs := sessionAttributes(request)

		Expect(attrs).To(ContainElement(
			rinq.Freeze(HttpdAttrClientIP, "10.1.1.1"),
		))
	})
})
