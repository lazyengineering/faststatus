// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// A Resource represents any resource (a person, a bathroom, a server, etc.)
// that needs to communicate how busy it is.
type Resource struct {
	ID     ID
	Status Status
	Since  time.Time
}

// NewResource creates a new Resource with a generated ID and otherwise zero-value properties.
func NewResource() Resource {
	id, _ := NewID()
	return Resource{ID: id}
}

// Equal allows quick equality comparison for two resource values.
// Use this instead of the equality operator because a Resource contains
// a `time.Time` value, which cannot be compared with confidence.
func (r Resource) Equal(other Resource) bool {
	switch {
	case r.ID != other.ID,
		r.Status != other.Status,
		!r.Since.Equal(other.Since):
		return false
	default:
		return true
	}
}

// String will return a single-line representation of a valid resource.
// In order to optimize for standard streams, the output is as follows:
//   {{ID}} {{Status}} {{Since}}
// Formatted as follows:
//   01234567-89ab-cdef-0123-456789abcdef busy 2006-01-02T15:04:05Z07:00 My Resource
func (r Resource) String() string {
	txt, err := r.MarshalText()
	if err != nil {
		return ""
	}
	return string(txt)
}

// MarshalText encodes a Resource to the text representation. In order to
// better stream text, the output is as follows:
//   {{ID}} {{Status}} {{Since}}
// Formatted as follows:
//   01234567-89ab-cdef-0123-456789abcdef busy 2006-01-02T15:04:05Z07:00 My Resource
// An invalid Status (out of range, etc.) will result in an error.
func (r Resource) MarshalText() ([]byte, error) {
	txt := make([]byte, 0, 128)

	id, err := r.ID.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("marshaling text for ID: %+v", err)
	}
	txt = append(txt, id...)

	txt = append(txt, ' ')
	status, err := r.Status.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("marshaling Status to text: %+v", err)
	}
	txt = append(txt, status...)

	txt = append(txt, ' ')
	since, err := r.Since.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("marshaling Since to text: %+v", err)
	}
	txt = append(txt, since...)

	return txt, nil
}

// UnmarshalText decodes a Resource from a line of text. This matches the
// output of the `MarshalText` method.
func (r *Resource) UnmarshalText(txt []byte) error {
	elements := bytes.Split(txt, []byte(" "))

	if len(elements) < 3 {
		return fmt.Errorf("invalid resource text")
	}

	tmp := Resource{}

	if err := (&tmp.ID).UnmarshalText(elements[0]); err != nil {
		return fmt.Errorf("parsing ID from text: %+v", err)
	}

	if err := (&tmp.Status).UnmarshalText(elements[1]); err != nil {
		return fmt.Errorf("parsing Status from text: %+v", err)
	}

	if err := (&tmp.Since).UnmarshalText(elements[2]); err != nil {
		return fmt.Errorf("parsing Since from text: %+v", err)
	}
	if tmp.Since.IsZero() {
		tmp.Since = time.Time{}
	}

	*r = tmp

	return nil
}

// MarshalJSON will return simple a simple json structure for a resource.
// Will not accept any Status that is out of range; see Status documentation
// for more information.
func (r Resource) MarshalJSON() ([]byte, error) {
	tmpResource := struct {
		ID     ID        `json:"id"`
		Status Status    `json:"status"`
		Since  time.Time `json:"since"`
	}{
		r.ID,
		r.Status,
		r.Since,
	}
	return json.Marshal(tmpResource)
}

// UnmarshalJSON will populate a Resource with data from a json struct
// according to the same format as MarshalJSON. Will overwrite any values
// already assigned to the Resource.
func (r *Resource) UnmarshalJSON(raw []byte) error {
	tmp := new(struct {
		ID     ID
		Status Status
		Since  time.Time
	})
	if err := json.Unmarshal(raw, tmp); err != nil {
		return err
	}

	r.ID = tmp.ID
	r.Status = tmp.Status
	r.Since = tmp.Since
	if r.Since.IsZero() {
		r.Since = time.Time{}
	}
	return nil
}

const binaryVersion = 0x00

// MagicBytes are the first two bytes of the portable binary representation of a Resource.
var MagicBytes = [2]byte{0x90, 0xe9}

// MarshalBinary returns a portable binary version of a Resource.
// The resulting binary must contain a header with MagicBytes (0x09 0xe9),
// a version byte, and a single empty buffer byte.
func (r Resource) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4+32)

	if n := copy(b[0:2], MagicBytes[:]); n != 2 {
		return nil, fmt.Errorf("unable to copy correct magic bytes")
	}
	b[2] = binaryVersion

	id, err := r.ID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling ID to binary: %+v", err)
	}
	copy(b[4:20], id)

	status, err := r.Status.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling Status to binary: %+v", err)
	}
	copy(b[20:21], status)

	since, err := r.Since.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling Since to binary: %+v", err)
	}
	copy(b[21:36], since)

	return b, nil
}

// UnmarshalBinary replaces a Resource with the Resource represented
// by the binary input. The input binary must match the form of the
// MarshalBinary method.
func (r *Resource) UnmarshalBinary(b []byte) error {
	switch {
	case len(b) < 36:
		return fmt.Errorf("input binary data too short")
	case len(b) > 36:
		return fmt.Errorf("input binay data too long")
	case !bytes.Equal(b[0:2], MagicBytes[:]):
		return fmt.Errorf("unexpected magic bytes")
	case b[2] > binaryVersion:
		return fmt.Errorf("unexpected version number for binary format")
	default:
	}

	tmp := Resource{}

	if err := (&tmp.ID).UnmarshalBinary(b[4:20]); err != nil {
		return fmt.Errorf("parsing ID from binary: %+v", err)
	}

	if err := (&tmp.Status).UnmarshalBinary(b[20:21]); err != nil {
		return fmt.Errorf("parsing Status from binary: %+v", err)
	}

	if err := (&tmp.Since).UnmarshalBinary(b[21:36]); err != nil {
		return fmt.Errorf("parsing Since from binary: %+v", err)
	}

	*r = tmp
	return nil
}
