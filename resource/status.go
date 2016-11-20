// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Status represents how busy a given resource is on a scale from 0–2,
// where 0 (Free) is a completely unoccupied resource, 2 (Occupied) is
// completely occupied, and 1 (Busy) is anything between. The simplicity
// and flexibility of this scheme allows this to be used for any number
// of applications.
type Status uint8

const (
	Free     Status = iota // completely free resource
	Busy                   // resource is busy
	Occupied               // resource completely busy
)

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
		return fmt.Errorf("status out of range")
	}
	*s = tmp
	return nil
}

// For the purposes of the API, it is much cleaner to keep the
// string representation to "0,1,2" instead of the pretty text.
// Use Pretty instead for those representations. Out of range
// Status values will be returned the same as Free.
func (s Status) String() string {
	return strconv.FormatUint(uint64(s.forceRange()), 10)
}

// For those few times where the pretty version of the status
// is requested, Pretty() will return the full text representation.
// Out of range status values will be returned as "Free".
func (s Status) Pretty() string {
	switch s.forceRange() {
	case Busy:
		return "Busy"
	case Occupied:
		return "Occupied"
	case Free:
		return "Free"
	default: // this should be impossible...
		return ""
	}
}

func (s Status) inRange() bool {
	return s <= Occupied
}

// Return a valid Status in Range (only for use inside this package)
func (s Status) forceRange() Status {
	if !s.inRange() {
		return Free
	}
	return s
}

// MarshalJSON will return a numeric value in the valid range of Status values.
// A status that is higher than the defined status values will return an error
// which can be checked using the `IsOutOfRange(error)` function.
func (s Status) MarshalJSON() ([]byte, error) {
	if !s.inRange() {
		return nil, errOutOfRange
	}
	return json.Marshal(uint8(s))
}

// UnmarshalJSON will assign a valid Status value from a numeric value.
// A status that is higher than the defined status values will return an error
// which can be checked using the `IsOutOfRange(error)` function.
func (s *Status) UnmarshalJSON(raw []byte) error {
	t := new(uint8)
	if err := json.Unmarshal(raw, t); err != nil {
		return err
	}
	*s = Status(*t)
	if !s.inRange() {
		*s = Free // set to zero value by default
		return errOutOfRange
	}
	return nil
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
