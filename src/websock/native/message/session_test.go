package message_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
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
})
