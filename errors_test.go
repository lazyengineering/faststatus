// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus_test

import (
	"errors"
	"testing"

	"github.com/lazyengineering/faststatus"
)

func TestConflictError(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		wantConflict bool
	}{
		{"nil",
			nil,
			false,
		},
		{"new string",
			errors.New("an error"),
			false,
		},
		{"false conflict error",
			conflictError(false),
			false,
		},
		{"true conflict error",
			conflictError(true),
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := faststatus.ConflictError(tc.err)
			if got != tc.wantConflict {
				t.Fatalf("ConflictError(%+v) = %v, expected %v", tc.err, got, tc.wantConflict)
			}
		})
	}
}

type conflictError bool

func (e conflictError) Error() string {
	return "conflict error"
}

func (e conflictError) Conflict() bool {
	return bool(e)
}
