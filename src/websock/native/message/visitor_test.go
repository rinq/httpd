package message_test

import . "github.com/rinq/httpd/src/websock/native/message"

type mockVisitor struct {
	VisitedMessage Incoming
	Error          error
}

var _ Visitor = &mockVisitor{} // enforce interface compliance

func (v *mockVisitor) VisitSessionCreate(m *SessionCreate) error {
	v.VisitedMessage = m
	return v.Error
}

func (v *mockVisitor) VisitSessionDestroy(m *SessionDestroy) error {
	v.VisitedMessage = m
	return v.Error
}
