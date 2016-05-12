// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
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

// For the purposes of the API, it is much cleaner to keep the
// string representation to "0,1,2" instead of the pretty text.
// Use Pretty instead for those representations. Out of range
// Status values will be returned the same as Free.
func (s Status) String() string {
	return strconv.FormatUint(uint64(s.inRange()), 10)
}

// For those few times where the pretty version of the status
// is requested, Pretty() will return the full text representation.
// Out of range status values will be returned as "Free".
func (s Status) Pretty() string {
	switch s.inRange() {
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

// Return a valid Status in Range (only for use inside this package)
func (s Status) inRange() Status {
	if s > Occupied {
		return Free
	}
	return s
}
