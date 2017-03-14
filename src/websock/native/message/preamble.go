package message

import (
	"encoding/binary"
	"io"
)

type preamble struct {
	Session SessionIndex
}

func (p *preamble) read(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, &p.Session)
}

func (p *preamble) write(w io.Writer, t messageType) (err error) {
	err = binary.Write(w, binary.BigEndian, t)

	if err == nil {
		err = binary.Write(w, binary.BigEndian, p.Session)
	}

	return
}
