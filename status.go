// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus

import (
	"bytes"
	"errors"
	"fmt"
)

// Status represents how busy a given resource is on a scale from 0–2,
// where 0 (Free) is a completely unoccupied resource, 2 (Occupied) is
// completely occupied, and 1 (Busy) is anything between. The simplicity
// and flexibility of this scheme allows this to be used for any number
// of applications.
type Status uint8

// The following predefined Status values are the only valid status values
const (
	Free     Status = iota // a completely unutilized resource
	Busy                   // a resource that is being utilized, but not to capacity
	Occupied               // a resource that is being utilized to capacity
)
const statusText = "freebusyoccupied"
const statusNumbers = "012"

var statusTextIdx = [...]uint8{0, 4, 8, 16}

// MarshalBinary encodes a Status to a single byte in a slice
func (s Status) MarshalBinary() ([]byte, error) {
	return append([]byte{}, byte(s)), nil
}

// UnmarshalBinary decodes a Status from a single byte
func (s *Status) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return fmt.Errorf("status must be one byte")
	}
	tmp := Status(b[0])
	if tmp > Occupied {
		return errOutOfRange
	}
	*s = tmp
	return nil
}

// MarshalText encodes a Status to the text representation. For readable
// messages, this will be of the form "free|busy|occupied".
func (s Status) MarshalText() ([]byte, error) {
	if s < 0 || s >= Status(len(statusTextIdx)-1) {
		return nil, errOutOfRange
	}
	return []byte(statusText)[statusTextIdx[s]:statusTextIdx[s+1]], nil
}

// UnmarshalText decodes a Status from a text representation.
// This can include an integer as text or a case-insensitive name
// like "Free|BUSY|occupied"
func (s *Status) UnmarshalText(txt []byte) error {
	if len(txt) == 0 {
		return fmt.Errorf("status must be non-empty byte slice")
	}
	if len(txt) == 1 {
		for i, v := range []byte(statusNumbers) {
			if txt[0] == v {
				*s = Status(i)
				return nil
			}
		}
	}
	for i := range statusTextIdx[1:] {
		if bytes.EqualFold(txt, []byte(statusText)[statusTextIdx[i]:statusTextIdx[i+1]]) {
			*s = Status(i)
			return nil
		}
	}
	return fmt.Errorf("not a valid status value")
}

// String returns a simple text representation of the Status.
// Out of range status values will be returned as "Free".
func (s Status) String() string {
	if s < 0 || s >= Status(len(statusTextIdx)-1) {
		s = Free
	}
	txt, _ := s.MarshalText()
	return string(txt)
}

type statusError struct {
	err          error
	isOutOfRange bool
}

type outOfRanger interface {
	OutOfRange() bool
}

func (e *statusError) Error() string {
	return fmt.Sprintf("status error: %+v", e.err)
}

func (e *statusError) OutOfRange() bool {
	return e.isOutOfRange
}

// IsOutOfRange returns true for an error indicating that the `Status` is out of range
func IsOutOfRange(e error) bool {
	or, ok := e.(outOfRanger)
	return ok && or.OutOfRange()
}

var errOutOfRange = &statusError{
	errors.New("Status not in valid range"),
	true,
}
