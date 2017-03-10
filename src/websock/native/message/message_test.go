package message_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
)

var _ = Describe("Read", func() {
	It("returns an error if the message type is unrecognized", func() {
		r := strings.NewReader("XX")
		_, err := Read(r, nil)

		Expect(err).Should(HaveOccurred())
	})

	It("returns an error if the message type can not be read", func() {
		_, err := Read(&bytes.Buffer{}, nil)

		Expect(err).Should(HaveOccurred())
	})

	It("returns an error if any of the data is unconsumed", func() {
		r := strings.NewReader("SC\xab\xcd\x00") // Session create message with an extra null-byte
		_, err := Read(r, nil)

		Expect(err).Should(HaveOccurred())
	})
})
