package message_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/httpd/src/websock/native/message"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Notification", func() {
	Describe("Write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			p := rinq.NewPayload("payload")
			m := &Notification{
				Session: 0xabcd,
				Header: NotificationHeader{
					Type: "type",
				},
				Payload: p,
			}

			err := m.Write(&buf, JSONEncoding)

			fmt.Println(buf.String())
			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'N', 'O',
				0xab, 0xcd, // session index
				0, 8, // header size
			}
			expected = append(expected, `["type"]`...)
			expected = append(expected, `"payload"`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})
