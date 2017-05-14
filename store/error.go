// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package store

import (
	"strings"

	"github.com/pkg/errors"
)

type dataError struct {
	old  bool
	noID bool
}

func (e dataError) Error() string {
	reasons := []string{"bad data"}
	if e.old {
		reasons = append(reasons, "a more recent version of this resource already exists")
	}
	if e.noID {
		reasons = append(reasons, "resource ID cannot be zero-value")
	}
	return strings.Join(reasons, ", ")
}

func (e dataError) Conflict() bool {
	return e.old
}

func (e dataError) ZeroValue() bool {
	return e.noID
}

// ZeroValueError checks to see if the error (or its Cause) is a result of zero-value
// data where non-zero data is required.
//
// An error value may be a zero-value data error if it implements this interface:
//
//    type zerovaluer interface {
//      ZeroValue() bool
//    }
//
// Otherwise it is not considered a zero-value error
func ZeroValueError(e error) bool {
	type zeroValuer interface {
		ZeroValue() bool
	}
	if e, ok := e.(zeroValuer); ok {
		return e.ZeroValue()
	}
	if e, ok := errors.Cause(e).(zeroValuer); ok {
		return e.ZeroValue()
	}
	return false
}
