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

func (id *ID) UnmarshalText(txt []byte) error {
	if len(txt) < 32 {
		return fmt.Errorf("UUID text must be longer than 32 characters")
	}

	buf := make([]byte, 16)

	if _, err := hex.Decode(buf[0:4], txt[0:8]); err != nil {
		return fmt.Errorf("decoding hex into uuid: %+v", err)
	}
	txt = txt[8:]
	if txt[0] == '-' {
		txt = txt[1:]
	}
	if _, err := hex.Decode(buf[4:6], txt[0:4]); err != nil {
		return fmt.Errorf("decoding hex into uuid: %+v", err)
	}
	txt = txt[4:]
	if txt[0] == '-' {
		txt = txt[1:]
	}
	if _, err := hex.Decode(buf[6:8], txt[0:4]); err != nil {
		return fmt.Errorf("decoding hex into uuid: %+v", err)
	}
	txt = txt[4:]
	if txt[0] == '-' {
		txt = txt[1:]
	}
	if _, err := hex.Decode(buf[8:10], txt[0:4]); err != nil {
		return fmt.Errorf("decoding hex into uuid: %+v", err)
	}
	txt = txt[4:]
	if txt[0] == '-' {
		txt = txt[1:]
	}
	if n, err := hex.Decode(buf[10:], txt[0:]); err != nil {
		return fmt.Errorf("decoding hex into uuid: %+v", err)
	} else if n != 6 {
		return fmt.Errorf("decoding hex into uuid: not enough bytes")
	}

	return id.UnmarshalBinary(buf)
}
