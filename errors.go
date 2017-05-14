// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus

import "github.com/pkg/errors"

// ConflictError checks to see if the error (or its Cause) is a result of conflict
// data. An error value may be a conflict data error if it implements this interface:
//
//    type conflicter interface {
//      Conflict() bool
//    }
//
// Otherwise it is not considered conflict data.
func ConflictError(e error) bool {
	type conflicter interface {
		Conflict() bool
	}
	if e, ok := e.(conflicter); ok {
		return e.Conflict()
	}
	if e, ok := errors.Cause(e).(conflicter); ok {
		return e.Conflict()
	}
	return false
}
