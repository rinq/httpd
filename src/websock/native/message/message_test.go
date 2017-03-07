package message_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
)

var _ = Describe("Read", func() {
	It("returns an error if the message type is unrecognised", func() {
		r := strings.NewReader("XX")
		_, err := Read(r, nil)

		Expect(err).Should(HaveOccurred())
	})

	It("returns an error if the message type can not be read", func() {
		_, err := Read(&bytes.Buffer{}, nil)

		Expect(err).Should(HaveOccurred())
	})
})
