package message_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("foreignEncoding / JSONEncoding", func() {
	var (
		subject Encoding

		headerBytes []byte
		headerValue testHeader

		payloadBytes []byte
		payloadValue testPayload
	)

	BeforeEach(func() {
		subject = JSONEncoding

		headerBytes = []byte{0, 11} // size = 11 bytes}
		headerBytes = append(headerBytes, `[123,"abc"]`...)
		headerValue = testHeader{123, "abc"}

		payloadBytes = []byte(`{"A":456,"B":"def"}`)
		payloadValue = testPayload{456, "def"}
	})

	Describe("Name", func() {
		It("returns the encoding name", func() {
			Expect(subject.Name()).To(Equal("json"))
		})
	})

	Describe("EncodeHeader", func() {
		It("encodes the header as an array", func() {
			var buf bytes.Buffer
			err := subject.EncodeHeader(&buf, headerValue)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Bytes()).To(Equal(headerBytes))
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
	})

	Describe("EncodePayload", func() {
		It("encodes the payload", func() {
			var buf bytes.Buffer
			p := rinq.NewPayload(payloadValue)
			err := subject.EncodePayload(&buf, p)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Bytes()).To(Equal(payloadBytes))
		})
	})

	Describe("DecodePayload", func() {
		It("decodes the payload", func() {
			r := bytes.NewBuffer(payloadBytes)
			p, err := subject.DecodePayload(r)

			Expect(err).ShouldNot(HaveOccurred())

			var v testPayload
			err = p.Decode(&v)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(v).To(Equal(payloadValue))
		})
	})
})
