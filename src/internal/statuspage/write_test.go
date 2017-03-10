package statuspage

import (
	"net/http"
	"net/http/httptest"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Write", func() {
	It("uses the default status message", func() {
		request := httptest.NewRequest("GET", "/", nil)
		response := httptest.NewRecorder()

		Write(response, request, http.StatusNotFound)

		m := Message(http.StatusNotFound)
		Expect(response.Body).To(ContainSubstring(m))
	})
})

var _ = Describe("WriteMessage", func() {
	var (
		request  *http.Request
		response *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		request = httptest.NewRequest("GET", "/", nil)
		response = httptest.NewRecorder()
	})

	It("writes the correct status code", func() {
		_, err := WriteMessage(response, request, http.StatusNotFound, "")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNotFound))
	})

	It("includes the status code in the body", func() {
		_, err := WriteMessage(response, request, http.StatusNotFound, "")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.Body).To(ContainSubstring("404"))
	})

	It("includes the status name in the body", func() {
		_, err := WriteMessage(response, request, http.StatusNotFound, "")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.Body).To(ContainSubstring("Not Found"))
	})

	It("includes the status message in the body", func() {
		_, err := WriteMessage(response, request, http.StatusNotFound, "custom-message")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.Body).To(ContainSubstring("custom-message"))
	})

	It("includes the status message in a header", func() {
		_, err := WriteMessage(response, request, http.StatusNotFound, "custom-message")

		h := response.Header().Get("X-Status-Message")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(h).To(Equal("custom-message"))
	})

	It("renders HTML if Accept header prioritizes it", func() {
		request.Header = http.Header{}
		request.Header.Add("Accept", "text/html;q=0.9,*/*;q=0.8")
		_, err := WriteMessage(response, request, http.StatusNotFound, "")

		h := response.Header().Get("Content-Type")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(h).To(ContainSubstring("text/html"))
	})

	It("renders text if Accept header prioritizes it", func() {
		request.Header = http.Header{}
		request.Header.Add("Accept", "text/html;q=0.8,text/plain;q=0.9*/*;q=0.7")
		_, err := WriteMessage(response, request, http.StatusNotFound, "")

		h := response.Header().Get("Content-Type")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(h).To(ContainSubstring("text/plain"))
	})

	It("renders text by default", func() {
		_, err := WriteMessage(response, request, http.StatusNotFound, "")

		h := response.Header().Get("Content-Type")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(h).To(ContainSubstring("text/plain"))
	})

	It("returns an error if text rendering fails", func() {
		prev := textTemplate
		textTemplate = &template.Template{}
		_, err := WriteMessage(response, request, http.StatusNotFound, "")
		textTemplate = prev

		Expect(err).Should(HaveOccurred())
	})
})
