package message

import (
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/rinq/httpd/src/internal/bufferpool"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/ugorji/go/codec"
)

// Encoding is an interface for a structured data encoder/decoder used to
// serialize message headers and application payloads.
type Encoding interface {
	EncodeHeader(w io.Writer, h interface{}) error
	DecodeHeader(r io.Reader, h interface{}) error
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

// headerEncoding provides a common implementation of header encoding/decoding.
type headerEncoding struct {
	handle codec.Handle
}

func (e *headerEncoding) EncodeHeader(w io.Writer, h interface{}) (err error) {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	err = codec.NewEncoder(buf, e.handle).Encode(h)

	if err == nil {
		if buf.Len() > math.MaxUint16 {
			err = errors.New("header exceeds maximum size")
		} else {
			err = binary.Write(w, binary.BigEndian, uint16(buf.Len()))
		}
	}

	if err == nil {
		_, err = buf.WriteTo(w)
	}

	return
}

func (e *headerEncoding) DecodeHeader(r io.Reader, h interface{}) (err error) {
	var size uint16

	err = binary.Read(r, binary.BigEndian, &size)

	if err == nil && size > 0 {
		err = codec.NewDecoder(
			&io.LimitedReader{R: r, N: int64(size)},
			e.handle,
		).Decode(h)
	}

	return
}

func init() {
	{
		headerHandle := &codec.CborHandle{}
		headerHandle.StructToArray = true
		CBOREncoding = &nativeEncoding{headerEncoding{headerHandle}}
	}

	{
		headerHandle := &codec.JsonHandle{}
		headerHandle.StructToArray = true
		JSONEncoding = &foreignEncoding{
			headerEncoding{headerHandle},
			&codec.JsonHandle{},
		}
	}
}
