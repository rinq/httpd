package websock

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/httpd/src/websock/protocol"
)

var _ = Describe("newUpgrader", func() {
	var (
		request  *http.Request
		response *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		request = httptest.NewRequest("GET", "/", nil)
		response = httptest.NewRecorder()
	})

	It("returns an error when the request is not an upgrade", func() {
		p := protocol.NewSet()
		u := newUpgrader("*", p)

		_, code, err := u(response, request)

		Expect(code).To(Equal(http.StatusBadRequest))
		Expect(err).Should(HaveOccurred())
	})
})
