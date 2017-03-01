package websock_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock"
)

var _ = Describe("NewOriginChecker", func() {
	var (
		withOrigin    http.Request
		withoutOrigin http.Request
		invalidOrigin http.Request
	)

	BeforeEach(func() {
		withOrigin.Header = http.Header{}
		withOrigin.Header.Add("Origin", "https://host.domain.tld")

		invalidOrigin.Header = http.Header{}
		invalidOrigin.Header.Add("Origin", ":")
	})

	It("returns nil if the pattern is empty", func() {
		fn := NewOriginChecker("")
		Expect(fn).To(BeNil())
	})

	DescribeTable(
		"returns a checker that returns true for matches",
		func(p string, r *http.Request) {
			fn := NewOriginChecker(p)
			Expect(fn(r)).To(BeTrue())
		},
		Entry("any", "*", &withOrigin),
		Entry("suffix", "*.tld", &withOrigin),
		Entry("prefix", "host.*", &withOrigin),
		Entry("exact", "host.domain.tld", &withOrigin),

		Entry("any (no origin)", "*", &withoutOrigin),

		Entry("any (invalid origin)", "*", &invalidOrigin),
	)

	DescribeTable(
		"returns a checker that returns false for non-matches",
		func(p string, r *http.Request) {
			fn := NewOriginChecker(p)
			Expect(fn(r)).To(BeFalse())
		},
		Entry("suffix", "*.other", &withOrigin),
		Entry("prefix", "other.*", &withOrigin),
		Entry("exact", "host.other.tld", &withOrigin),

		Entry("suffix (no origin)", "*.tld", &withoutOrigin),
		Entry("prefix (no origin)", "host.*", &withoutOrigin),
		Entry("exact (no origin)", "host.domain.tld", &withoutOrigin),

		Entry("suffix (invalid origin)", "*.tld", &invalidOrigin),
		Entry("prefix (invalid origin)", "host.*", &invalidOrigin),
		Entry("exact (invalid origin)", "host.domain.tld", &invalidOrigin),
	)
})
