package message

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SessionCreate", func() {
	Describe("Accept", func() {
		It("invokes the correct visit method", func() {
			expected := errors.New("visit error")
			v := &mockVisitor{Error: expected}
			m := &SessionCreate{}

			err := m.Accept(v)

			Expect(err).To(Equal(expected))
			Expect(v.VisitedMessage).To(Equal(m))
		})
	})

	Describe("read", func() {
		It("decodes the message", func() {
			r := strings.NewReader("SC\xab\xcd")

			m, err := Read(r, nil)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(m).To(Equal(&SessionCreate{preamble: preamble{0xabcd}}))
		})
	})
})

var _ = Describe("SessionDestroy", func() {
	Describe("Accept", func() {
		It("invokes the correct visit method", func() {
			expected := errors.New("visit error")
			v := &mockVisitor{Error: expected}
			m := &SessionDestroy{}

			err := m.Accept(v)

			Expect(err).To(Equal(expected))
			Expect(v.VisitedMessage).To(Equal(m))
		})
	})

	Describe("Accept", func() {
		It("invokes the correct visit method", func() {
			expected := errors.New("visit error")
			v := &mockVisitor{Error: expected}
			m := &SessionDestroy{}

			err := m.Accept(v)

			Expect(err).To(Equal(expected))
			Expect(v.VisitedMessage).To(Equal(m))
		})
	})

	Describe("read", func() {
		It("decodes the message", func() {
			r := strings.NewReader("SD\xab\xcd")

			m, err := Read(r, nil)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(m).To(Equal(&SessionDestroy{preamble: preamble{0xabcd}}))
		})
	})

	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			m := &SessionDestroy{preamble: preamble{0xabcd}}

			err := Write(&buf, nil, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'S', 'D',
				0xab, 0xcd, // session index
			}
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})

var _ = Describe("NewSessionDestroy", func() {
	It("returns a session destroy message", func() {
		m := NewSessionDestroy(0xabcd)

		Expect(m).To(Equal(&SessionDestroy{
			preamble: preamble{0xabcd},
		}))
	})
})
