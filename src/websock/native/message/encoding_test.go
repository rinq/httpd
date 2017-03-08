package message_test

import (
	"bytes"
	"errors"
	"math"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
)

var _ = Describe("headerEncoding", func() {

	var (
		subject     Encoding
		headerBytes []byte
		headerValue testHeader
	)

	BeforeEach(func() {
		subject = JSONEncoding

		headerBytes = []byte{0, 11} // size = 11 bytes}
		headerBytes = append(headerBytes, `[123,"abc"]`...)
		headerValue = testHeader{123, "abc"}
	})

	Describe("EncodeHeader", func() {
		It("encodes the header as an array", func() {
			var buf bytes.Buffer
			err := subject.EncodeHeader(&buf, headerValue)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Bytes()).To(Equal(headerBytes))
		})

		It("returns an error if the header can not be encoded", func() {
			var buf bytes.Buffer

			err := subject.EncodeHeader(&buf, brokerHeader{})

			Expect(err).Should(HaveOccurred())
		})

		It("returns an error if the header is too long", func() {
			var buf bytes.Buffer
			longHeader := strings.Repeat("x", math.MaxUint16+1)

			err := subject.EncodeHeader(&buf, longHeader)

			Expect(err).Should(HaveOccurred())
		})

		It("returns an error if the writer is unwritable", func() {
			err := subject.EncodeHeader(brokenWriter{}, headerValue)

			Expect(err).Should(HaveOccurred())
		})
	})

	Describe("DecodeHeader", func() {
		It("decodes the header from an array", func() {
			r := bytes.NewReader(headerBytes)
			var h testHeader
			err := subject.DecodeHeader(r, &h)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(h).To(Equal(headerValue))
		})

		It("does not read past the header length", func() {
			r := bytes.NewBuffer(headerBytes)
			r.Write([]byte{10, 20, 30}) // garbage

			var h testHeader
			err := subject.DecodeHeader(r, &h)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(h).To(Equal(headerValue))
		})

		It("returns an error if the header is malformed", func() {
			r := bytes.NewBuffer([]byte{0, 1, '}'}) // invalid json
			var h testHeader
			err := subject.DecodeHeader(r, &h)

			Expect(err).Should(HaveOccurred())
		})

		It("returns an error if the size is not present", func() {
			r := bytes.NewBuffer([]byte{0}) // too short
			var h testHeader
			err := subject.DecodeHeader(r, &h)

			Expect(err).Should(HaveOccurred())
		})
	})
})

type testHeader struct {
	A int
	B string
}

type testPayload struct {
	A int
	B string
}

type brokerHeader struct{}

func (brokerHeader) MarshalJSON() ([]byte, error) { return nil, errors.New("forced error") }

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("forced error") }
