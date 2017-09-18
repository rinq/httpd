package message

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

func (v *mockVisitor) VisitListen(m *Listen) error {
	v.VisitedMessage = m
	return v.Error
}

func (v *mockVisitor) VisitUnlisten(m *Unlisten) error {
	v.VisitedMessage = m
	return v.Error
}

func (v *mockVisitor) VisitSyncCall(m *SyncCall) error {
	v.VisitedMessage = m
	return v.Error
}

func (v *mockVisitor) VisitAsyncCall(m *AsyncCall) error {
	v.VisitedMessage = m
	return v.Error
}

func (v *mockVisitor) VisitExecute(m *Execute) error {
	v.VisitedMessage = m
	return v.Error
}
