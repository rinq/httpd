package message

import (
	"io"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/ugorji/go/codec"
)

// foreignEncoding is an implementation of the Encoding interface that uses any codec
// handle.
//
// Application payloads are converted to/from CBOR as necessary.
type foreignEncoding struct {
	headerEncoding
	handle codec.Handle
	name   string
}

func (e *foreignEncoding) Name() string {
	return e.name
}

func (e *foreignEncoding) EncodePayload(w io.Writer, p *rinq.Payload) error {
	defer p.Close()
	return codec.NewEncoder(w, e.handle).Encode(p.Value())
}

func (e *foreignEncoding) DecodePayload(r io.Reader) (p *rinq.Payload, err error) {
	var v interface{}
	err = codec.NewDecoder(r, e.handle).Decode(&v)

	if err == nil {
		p = rinq.NewPayload(v)
	}

	return
}
