package websock

import "github.com/gorilla/websocket"

// Protocol is an interface that handles one or more WebSocket sub-protocols.
type Protocol interface {
	Names() []string
	Handle(*websocket.Conn)
}

// ProtocolSet is a collection of Protocols.
type ProtocolSet struct {
	names []string
	assoc map[string]Protocol
}

// NewProtocolSet returns a new protocol set containing the protocols in p.
func NewProtocolSet(p ...Protocol) *ProtocolSet {
	s := &ProtocolSet{assoc: map[string]Protocol{}}

	for _, pr := range p {
		for _, n := range pr.Names() {
			if _, ok := s.assoc[n]; !ok {
				s.assoc[n] = pr
				s.names = append(s.names, n)
			}
		}
	}

	return s
}

// Names returns a list of protocol names for all protocols in s.
func (s *ProtocolSet) Names() []string {
	return s.names
}

// Select returns a protocol from p that can handle the protocol named n.
func (s *ProtocolSet) Select(n string) (Protocol, bool) {
	pr, ok := s.assoc[n]
	return pr, ok
}
