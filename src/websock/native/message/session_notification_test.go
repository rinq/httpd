package message

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Notification", func() {
	Describe("write", func() {
		It("encodes the message", func() {
			var buf bytes.Buffer
			p := rinq.NewPayload("payload")
			m := &Notification{
				preamble: preamble{0xabcd},
				notificationHeader: notificationHeader{
					Namespace: "ns",
					Type:      "type",
				},
				Payload: p,
			}

			err := Write(&buf, JSONEncoding, m)

			Expect(err).ShouldNot(HaveOccurred())

			expected := []byte{
				'N', 'O',
				0xab, 0xcd, // session index
				0, 13, // header size
			}
			expected = append(expected, `["ns","type"]`...)
			expected = append(expected, `"payload"`...)
			Expect(buf.Bytes()).To(Equal(expected))
		})
	})
})

var _ = Describe("NewNotification", func() {
	It("returns a notification message", func() {
		p := rinq.NewPayload(456)
		n := rinq.Notification{
			Namespace: "ns",
			Type:      "type",
			Payload:   p,
		}
		m := NewNotification(0xabcd, n)

		Expect(m).To(Equal(&Notification{
			preamble: preamble{0xabcd},
			notificationHeader: notificationHeader{
				Namespace: "ns",
				Type:      "type",
			},
			Payload: p,
		}))
	})
})
