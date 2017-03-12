package message_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("nativeEncoding / CBOREncoding", func() {
	var (
		subject Encoding

		headerBytes []byte
		headerValue testHeader

		payloadBytes []byte
		payloadValue testPayload
	)

	BeforeEach(func() {
		subject = CBOREncoding

		headerBytes = []byte{
			0, 7, // size = 7 bytes
			130, 24, 123, 99, 97, 98, 99,
		}
		headerValue = testHeader{123, "abc"}

		payloadBytes = []byte{162, 97, 65, 25, 1, 200, 97, 66, 99, 100, 101, 102}
		payloadValue = testPayload{456, "def"}
	})

	Describe("Name", func() {
		It("returns cbor", func() {
			Expect(subject.Name()).To(Equal("cbor"))
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
			Expect(buf.Bytes()).To(Equal(p.Bytes()))
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
