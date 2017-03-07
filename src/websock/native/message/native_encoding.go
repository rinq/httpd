package message

import (
	"io"
	"io/ioutil"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/ugorji/go/codec"
)

// nativeEncoding is an implementation of the Encoding interface that uses CBOR.
//
// CBOR is the native format of Rinq payloads. When CBOR encoding is used, the
// HTTP server does not inspect application payloads, they are forwarded
// directly to Rinq.
type nativeEncoding struct {
	headerHandle codec.Handle
}

func (e *nativeEncoding) EncodeHeader(w io.Writer, h interface{}) error {
	enc := codec.NewEncoder(w, e.headerHandle)
	return enc.Encode(h)
}

func (e *nativeEncoding) DecodeHeader(r io.Reader, n uint16, h interface{}) error {
	dec := codec.NewDecoder(
		&io.LimitedReader{R: r, N: int64(n)},
		e.headerHandle,
	)
	return dec.Decode(h)
}

func (e *nativeEncoding) EncodePayload(w io.Writer, p *rinq.Payload) error {
	_, err := w.Write(p.Bytes())
	return err
}

func (e *nativeEncoding) DecodePayload(r io.Reader) (*rinq.Payload, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return rinq.NewPayloadFromBytes(buf), nil
}
