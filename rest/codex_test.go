package rest_test

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/lazyengineering/faststatus/rest"
	"github.com/pkg/errors"
)

func TestJSONDecoderDecodeRejectsBadMIMETypes(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
		typeError   bool
	}{
		{"empty", "", false},
		{"bad", "32908jasdfn and other. stuff-&%$#@", false},
		{"not supported", "text/plain", true},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := &rest.JSONDecoder{}
			err := d.Decode(tc.contentType, nil, nil)
			if err == nil {
				t.Fatalf("failed to reject non-json content type %q", tc.contentType)
			}
			if rest.ContentTypeError(err) != tc.typeError {
				t.Fatalf("failed to return content-type error %q", tc.contentType)
			}
		})
	}
}

func TestJSONDecoderDecodeAcceptsGoodMIMETypes(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
	}{
		{"simple", "application/json"},
		{"with encoding", "application/json; charset=utf-8"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			const input = `{"foo":"bar"}`
			d := &rest.JSONDecoder{}

			var called bool
			v := &mockJSONUnmarshaler{
				unmarshalFn: func(got []byte) error {
					called = true
					if !reflect.DeepEqual(got, []byte(input)) {
						t.Fatalf("bytes sent to unmarshal JSON %q, expected %q", got, input)
					}
					return nil
				},
			}
			err := d.Decode(tc.contentType, strings.NewReader(input), v)
			if err != nil {
				t.Fatalf("unexpected error calling Decode: %v", err)
			}
			if !called {
				t.Fatalf("failed to call unmarshal")
			}
		})
	}
}

func TestJSONDecoderDecodeLimitsReads(t *testing.T) {
	limit := 12345

	eternalReader := io.MultiReader(
		strings.NewReader(`{"foo":"`),
		&mockReader{readFn: func(b []byte) (int, error) {
			for i := range b {
				b[i] = 'a'
			}
			return len(b), nil
		}},
	)
	var readCount int
	r := &mockReader{readFn: func(b []byte) (int, error) {
		n, err := eternalReader.Read(b)
		readCount += n
		if readCount > limit {
			t.Fatalf("read %d, more than limit of %d", readCount, limit)
		}
		return n, err
	}}
	v := &mockJSONUnmarshaler{ // unlikely to be called
		unmarshalFn: func(got []byte) error {
			return nil
		},
	}

	d := &rest.JSONDecoder{Limit: int64(limit)}
	err := d.Decode("application/json", r, v)
	if err == nil {
		t.Fatalf("expect an error, always")
	}

}

func TestTextDecoderDecodeRejectsBadMIMETypes(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
		typeError   bool
	}{
		{"empty", "", false},
		{"bad", "32908jasdfn and other. stuff-&%$#@", false},
		{"not supported", "application/json", true},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := &rest.TextDecoder{}
			err := d.Decode(tc.contentType, nil, &mockTextUnmarshaler{})
			if err == nil {
				t.Fatalf("failed to reject non-text content type %q", tc.contentType)
			}
			if rest.ContentTypeError(err) != tc.typeError {
				t.Fatalf("got error: %v, failed to return content-type error %q", err, tc.contentType)
			}
		})
	}
}

func TestTextDecoderDecodeAcceptsGoodMIMETypes(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
	}{
		{"simple", "text/plain"},
		{"with encoding", "text/plain; charset=utf-8"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			const input = `foo bar`
			d := &rest.TextDecoder{}

			var called bool
			v := &mockTextUnmarshaler{
				unmarshalFn: func(got []byte) error {
					called = true
					if !reflect.DeepEqual(got, []byte(input)) {
						t.Fatalf("bytes sent to unmarshal text %q, expected %q", got, input)
					}
					return nil
				},
			}
			err := d.Decode(tc.contentType, strings.NewReader(input), v)
			if err != nil {
				t.Fatalf("unexpected error calling Decode: %v", err)
			}
			if !called {
				t.Fatalf("failed to call unmarshal")
			}
		})
	}
}

func TestTextDecoderDecodeLimitsReads(t *testing.T) {
	limit := 12345

	eternalReader := io.MultiReader(
		strings.NewReader(`foo `),
		&mockReader{readFn: func(b []byte) (int, error) {
			for i := range b {
				b[i] = 'a'
			}
			return len(b), nil
		}},
	)
	var readCount int
	r := &mockReader{readFn: func(b []byte) (int, error) {
		n, err := eternalReader.Read(b)
		readCount += n
		if readCount > limit {
			t.Fatalf("read %d, more than limit of %d", readCount, limit)
		}
		return n, err
	}}
	v := &mockTextUnmarshaler{ // unlikely to be called
		unmarshalFn: func(got []byte) error {
			return errors.New("it's an error too")
		},
	}

	d := &rest.TextDecoder{Limit: int64(limit)}
	err := d.Decode("text/plain", r, v)
	if err == nil {
		t.Fatalf("expect an error, always")
	}

}

func TestMultiDecoderDecode(t *testing.T) {
	testCases := []struct {
		name                 string
		responses            []error
		wantNegotiationError bool // assumes wantError is also true
		wantError            bool
		wantCalls            int
	}{
		{"Content negotiation errors are ignored",
			[]error{rest.ErrorContentType("foo"), nil},
			false,
			false,
			2,
		},
		{"Any other errors return immediately",
			[]error{rest.ErrorContentType("foo"), errors.New("something"), nil},
			false,
			true,
			2,
		},
		{"Success returns immediately",
			[]error{rest.ErrorContentType("foo"), nil, errors.New("something")},
			false,
			false,
			2,
		},
		{"No success or errors is a content negotiation error",
			[]error{rest.ErrorContentType("foo"), rest.ErrorContentType("bar"), rest.ErrorContentType("baz")},
			true,
			true,
			3,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var called int
			md := rest.MultiDecoder{}
			for _, d := range tc.responses {
				e := d
				md = append(md, &mockDecoder{func(string, io.Reader, interface{}) error {
					called++
					return e
				}})
			}
			err := md.Decode("testing/test", nil, &struct{ Foo string }{})
			if (err != nil) != tc.wantError {
				t.Fatalf("MultiDecoder.Decode() returned error %v, expected error? %v", err, tc.wantError)
			}
			if err != nil && rest.ContentTypeError(err) != tc.wantNegotiationError {
				t.Fatalf("MultiDecoder.Decode() error with negotiation %v, expected %v", rest.ContentTypeError(err), tc.wantNegotiationError)
			}
			if called != tc.wantCalls {
				t.Fatalf("MultiDecoder.Decode() returned after %d calls, expected %d", called, tc.wantCalls)
			}
		})
	}
}

type mockJSONUnmarshaler struct {
	unmarshalFn func([]byte) error
}

func (um *mockJSONUnmarshaler) UnmarshalJSON(b []byte) error {
	return um.unmarshalFn(b)
}

type mockTextUnmarshaler struct {
	unmarshalFn func([]byte) error
}

func (um *mockTextUnmarshaler) UnmarshalText(b []byte) error {
	return um.unmarshalFn(b)
}

type mockReader struct {
	readFn func([]byte) (int, error)
}

func (r *mockReader) Read(b []byte) (int, error) {
	return r.readFn(b)
}

type mockDecoder struct {
	decode func(string, io.Reader, interface{}) error
}

func (d *mockDecoder) Decode(ct string, r io.Reader, v interface{}) error {
	return d.decode(ct, r, v)
}
