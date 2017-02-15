// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// A UUIDv4 compatible byte array. Implementation is based on
// portions of github.com/satori/go.uuid.
type ID [16]byte

// Generate a new version 4 UUID. Returns errors from reading the
// entropy source.
func NewID() (ID, error) {
	id := ID{}
	if _, err := rand.Read(id[:]); err != nil {
		return ID{}, fmt.Errorf("reading random bytes for new ID: %+v", err)
	}
	id[6] = (id[6] & 0x0f) | (4 << 4)
	id[8] = (id[8] & 0xbf) | 0x80
	return id, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (id ID) MarshalBinary() ([]byte, error) {
	return id[:], nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
// and enforces the 16 byte length of an ID
func (id *ID) UnmarshalBinary(b []byte) error {
	if len(b) != 16 {
		return fmt.Errorf("id must be 16 bytes long")
	}
	copy(id[:], b)
	return nil
}

// MarshalText outputs the id as the canonical hexadecimal representation:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (id ID) MarshalText() ([]byte, error) {
	txt := make([]byte, 36)

	hex.Encode(txt[0:8], id[0:4])
	txt[8] = '-'
	hex.Encode(txt[9:13], id[4:6])
	txt[13] = '-'
	hex.Encode(txt[14:18], id[6:8])
	txt[18] = '-'
	hex.Encode(txt[19:23], id[8:10])
	txt[23] = '-'
	hex.Encode(txt[24:], id[10:])

	return txt, nil
}

// UnmarshalText populates the id with a uuid value represented by the text in
// the canonical form: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx. While any of the
// dashes in the form may be left out, any non-hexadecimal characters will result
// in an error.
func (id *ID) UnmarshalText(txt []byte) error {
	if len(txt) < 32 {
		return fmt.Errorf("UUID text must be longer than 32 characters")
	}

	buf := make([]byte, 16)

	var i int
	for _, n := range []int{8, 4, 4, 4, 12} {
		if txt[0] == '-' {
			txt = txt[1:]
		}
		if _, err := hex.Decode(buf[i:i+(n/2)], txt[0:n]); err != nil {
			return fmt.Errorf("decoding hex into uuid: %+v", err)
		}
		txt = txt[n:]
		i = i + (n / 2)
	}
	if len(txt) > 0 {
		return fmt.Errorf("too long for uuid: %+v", txt)
	}

	return id.UnmarshalBinary(buf)
}
