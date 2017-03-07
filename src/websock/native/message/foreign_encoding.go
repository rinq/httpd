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
	headerHandle  codec.Handle
	payloadHandle codec.Handle
}

func (e *foreignEncoding) EncodeHeader(w io.Writer, h interface{}) error {
	return codec.NewEncoder(w, e.headerHandle).Encode(h)
}

func (e *foreignEncoding) DecodeHeader(r io.Reader, n uint16, h interface{}) error {
	return codec.NewDecoder(
		&io.LimitedReader{R: r, N: int64(n)},
		e.headerHandle,
	).Decode(h)
}

func (e *foreignEncoding) EncodePayload(w io.Writer, p *rinq.Payload) error {
	return codec.NewEncoder(w, e.payloadHandle).Encode(p.Value())
}

func (e *foreignEncoding) DecodePayload(r io.Reader) (p *rinq.Payload, err error) {
	var v interface{}
	err = codec.NewDecoder(r, e.payloadHandle).Decode(&v)

	if err == nil {
		p = rinq.NewPayload(v)
	}

	return
}
