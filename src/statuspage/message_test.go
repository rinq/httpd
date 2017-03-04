package statuspage_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/rinq/httpd/src/statuspage"
	// . "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Message", func() {
	It("returns a non-empty string for every code", func() {
		for c := 0; c < 1000; c++ {
			m := Message(c)
			Expect(m).ShouldNot(HaveLen(0))
		}
	})
})
