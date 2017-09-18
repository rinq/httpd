package message

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Listen", func() {
	Describe("Accept", func() {
		It("invokes the correct visit method", func() {
			expected := errors.New("visit error")
			v := &mockVisitor{Error: expected}
			m := &Listen{}

			err := m.Accept(v)

			Expect(err).To(Equal(expected))
			Expect(v.VisitedMessage).To(Equal(m))
		})
	})

	Describe("read", func() {
		It("decodes the message", func() {
			buf := []byte{
				'N', 'L',
				0xab, 0xcd, // session index
				0, 15, // header length
			}
			buf = append(buf, `[["ns1","ns2"]]`...)

			r := bytes.NewReader(buf)
			m, err := Read(r, JSONEncoding)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(m).To(Equal(&Listen{
				preamble: preamble{0xabcd},
				listenHeader: listenHeader{
					Namespaces: []string{"ns1", "ns2"},
				},
			}))
		})
	})
})

var _ = Describe("Unlisten", func() {
	Describe("Accept", func() {
		It("invokes the correct visit method", func() {
			expected := errors.New("visit error")
			v := &mockVisitor{Error: expected}
			m := &Unlisten{}

			err := m.Accept(v)

			Expect(err).To(Equal(expected))
			Expect(v.VisitedMessage).To(Equal(m))
		})
	})

	Describe("read", func() {
		It("decodes the message", func() {
			buf := []byte{
				'N', 'U',
				0xab, 0xcd, // session index
				0, 15, // header length
			}
			buf = append(buf, `[["ns1","ns2"]]`...)

			r := bytes.NewReader(buf)
			m, err := Read(r, JSONEncoding)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(m).To(Equal(&Unlisten{
				preamble: preamble{0xabcd},
				listenHeader: listenHeader{
					Namespaces: []string{"ns1", "ns2"},
				},
			}))
		})
	})
})
