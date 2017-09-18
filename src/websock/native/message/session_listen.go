package message

import "io"

// Listen is an incoming message requesting that a session begin listening for
// notifications on a set of namespaces.
type Listen struct {
	preamble
	listenHeader
}

type listenHeader struct {
	Namespaces []string
}

// Accept calls the appropriate visit method on v.
func (m *Listen) Accept(v Visitor) error {
	return v.VisitListen(m)
}

func (m *Listen) read(r io.Reader, e Encoding) (err error) {
	err = m.preamble.read(r)

	if err == nil {
		err = e.DecodeHeader(r, &m.listenHeader)
	}

	return
}

// Unlisten is an incoming message requesting that a session stop listening for
// notifications on a set of namespaces.
type Unlisten struct {
	preamble
	listenHeader
}

// Accept calls the appropriate visit method on v.
func (m *Unlisten) Accept(v Visitor) error {
	return v.VisitUnlisten(m)
}

func (m *Unlisten) read(r io.Reader, e Encoding) (err error) {
	err = m.preamble.read(r)

	if err == nil {
		err = e.DecodeHeader(r, &m.listenHeader)
	}

	return
}
