package message

import (
	"io"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/ugorji/go/codec"
)

// Encoding is an interface for a structured data encoder/decoder used to
// serialize message headers and application payloads.
type Encoding interface {
	EncodeHeader(w io.Writer, h interface{}) error
	DecodeHeader(r io.Reader, n uint16, h interface{}) error
	EncodePayload(w io.Writer, p *rinq.Payload) error
	DecodePayload(r io.Reader) (*rinq.Payload, error)
}

var (
	// CBOREncoding is an Encoding implementation that uses CBOR, Rinq's native
	// application payload format. Application payloads do not need to be
	// marshalled for transmission over the WebSocket, they are passed directly
	// between the client and Rinq.
	CBOREncoding Encoding

	// JSONEncoding is an Encoding that uses JSON to message headers an
	// application payloads. Payloads are converted to and from Rinq's native
	// CBOR as neccessary.
	JSONEncoding Encoding
)

func init() {
	{
		headerHandle := &codec.CborHandle{}
		headerHandle.StructToArray = true
		CBOREncoding = &nativeEncoding{headerHandle}
	}

	{
		headerHandle := &codec.JsonHandle{}
		headerHandle.StructToArray = true
		JSONEncoding = &foreignEncoding{headerHandle, &codec.JsonHandle{}}
	}
}
