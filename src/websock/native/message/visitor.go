package message

// Visitor is an interface that visits each of the incoming message types.
type Visitor interface {
	VisitSessionCreate(*SessionCreate) error
	VisitSessionDestroy(*SessionDestroy) error
	VisitSyncCall(*SyncCall) error
	// VisitAsyncCall(*AsyncCall) error
	VisitExecute(*Execute) error
}
