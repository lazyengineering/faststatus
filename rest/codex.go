package rest

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"

	"github.com/pkg/errors"
)

// ask for content type and reader
// return generic decode function
// *or*
// ask for content type, reader, and destination interface
// return error
//
// implement the same interface for single content-type support as for multi

// Decoder is used to read and decode values from an input stream claiming to be
// the provided content type (MIME type). A typical implementation may return a
// content negotiation error for unsupported content types.
type Decoder interface {
	Decode(contentType string, r io.Reader, v interface{}) error
}

// JSONDecoder is used to decode from the JSON MIME type (application/json).
// Any other content type will result in a content negotiation error.
type JSONDecoder struct {
	Limit int64
}

// Decode implements the Decoder interface for JSON (application/json) streams.
func (d *JSONDecoder) Decode(contentType string, r io.Reader, v interface{}) error {
	if t, _, err := mime.ParseMediaType(contentType); err != nil {
		return errors.Wrap(err, "parsing content type")
	} else if t != "application/json" {
		return ErrorContentType(t)
	}

	if d.Limit > 0 {
		r = io.LimitReader(r, d.Limit)
	}
	return json.NewDecoder(r).Decode(v)
}

// TextDecoder is used to decode from the plain text MIME type (text/plain).
// Any other content type will result in a content negotiation error.
type TextDecoder struct {
	Limit int64
}

// Decode implements the Decoder interface for JSON (text/plain) streams.
func (d *TextDecoder) Decode(contentType string, r io.Reader, v interface{}) error {
	um, ok := v.(encoding.TextUnmarshaler)
	if !ok {
		return errors.New("value does not unmarshal from text")
	}

	if t, _, err := mime.ParseMediaType(contentType); err != nil {
		return errors.Wrap(err, "parsing content type")
	} else if t != "text/plain" {
		return ErrorContentType(t)
	}

	if d.Limit > 0 {
		r = io.LimitReader(r, d.Limit)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "reading all")
	}

	return errors.Wrap(um.UnmarshalText(b), "unmarshaling text")
}

// ContentTypeError indicated if the error is a content negotiation error.
func ContentTypeError(err error) bool {
	switch errors.Cause(err).(type) {
	case errorContentType:
		return true
	default:
		return false
	}
}

// ErrorContentType returns a content negotiation error. Use this to indicate
// that the requested type does not match the supported type or types.
func ErrorContentType(contentType string) error {
	return errorContentType{got: contentType}
}

type errorContentType struct {
	got string
}

// Error implements the error interface.
func (err errorContentType) Error() string {
	return fmt.Sprintf("unsupported content type: %q", err.got)
}
