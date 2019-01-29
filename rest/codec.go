package rest

import (
	"bufio"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/pkg/errors"
)

// Decoder populates a value from some other format into a go type.
type Decoder interface {
	Decode(interface{}) error
}

// DecoderFunc allows a function to be used as a Decoder.
type DecoderFunc func(interface{}) error

// Decode implements the Decoder interface with a DecoderFunc.
func (f DecoderFunc) Decode(v interface{}) error {
	return f(v)
}

// Codec creates Decoders for specified content types.
type Codec struct {
	Decoders map[string]func(io.Reader) Decoder
}

// NewDecoder creates a Decoder that best matches the contentType. If there is no
// contentType, it will attempt to sniff the content of the Reader to determine
// the content type.
//
// "application/json" will always use the JSONDecoder, "text/plain" will always use
// the TextDecoder, and "application/octet-stream" will always use the BinaryDecoder.
//
// Other content types will attempt to use a Decoder that matches the content type.
func (c *Codec) NewDecoder(contentType string, body io.Reader) (Decoder, error) {
	if contentType != "" {
		mediatype, _, err := mime.ParseMediaType(contentType)
		if err != nil && err != mime.ErrInvalidMediaParameter {
			return nil, &restError{
				err:  errors.Wrapf(err, "invalid content-type, %q", contentType),
				code: http.StatusUnsupportedMediaType,
			}
		}
		contentType = mediatype
	}
	if contentType == "" {
		reader := bufio.NewReader(body)
		b, err := reader.Peek(512) // since 512 is the most that sniffing will use
		if err != nil && err != bufio.ErrBufferFull {
			return nil, errors.Wrap(err, "peeking at the request body")
		}
		contentType = http.DetectContentType(b)
		body = reader
	}
	switch contentType {
	case "application/json":
		return JSONDecoder(body), nil
	case "text/plain":
		return TextDecoder(body), nil
	case "application/octet-stream":
		return BinaryDecoder(body), nil
	default:
		d, ok := c.Decoders[contentType]
		if !ok {
			return nil, errors.Errorf("no decoder for requested content-type, %q", contentType)
		}
		return d(body), nil
	}
}

func JSONDecoder(r io.Reader) Decoder {
	d := json.NewDecoder(r)
	return DecoderFunc(func(v interface{}) error {
		err := d.Decode(v)
		switch err.(type) {
		case *json.UnmarshalFieldError, *json.UnmarshalTypeError:
			return &restError{
				err:  fmt.Errorf("decoding resource from request: %+v", err),
				code: http.StatusBadRequest,
			}
		default:
			return errors.Wrap(err, "decoding resource from request")
		}
	})
}

func TextDecoder(r io.Reader) Decoder {
	return DecoderFunc(func(v interface{}) error {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "reading all from reader")
		}
		val, ok := v.(encoding.TextUnmarshaler)
		if !ok {
			return errors.New("value must be a encoding.TextUnmarshaler")
		}
		err = val.UnmarshalText(b)
		if err != nil {
			return &restError{
				err:  errors.Wrap(err, "unmarshaling text into value"),
				code: http.StatusBadRequest,
			}
		}
		return nil
	})
}

func BinaryDecoder(r io.Reader) Decoder {
	return DecoderFunc(func(v interface{}) error {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "reading all from reader")
		}
		val, ok := v.(encoding.BinaryUnmarshaler)
		if !ok {
			return errors.New("value must be a encoding.BinaryUnmarshaler")
		}
		err = val.UnmarshalBinary(b)
		if err != nil {
			return &restError{
				err:  errors.Wrap(err, "unmarshaling binary into value"),
				code: http.StatusBadRequest,
			}
		}
		return nil
	})
}
